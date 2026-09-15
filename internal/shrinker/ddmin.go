package shrinker

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

// ErrBaselineFailure means the target failure exists even when no operation runs,
// so operation-level delta debugging cannot produce a meaningful reproducer.
var ErrBaselineFailure = errors.New("failure reproduces without scheduled operations")

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func partition(c []domain.ScheduledOp, n int) [][]domain.ScheduledOp {
	subsets := make([][]domain.ScheduledOp, n)
	subsetSize := len(c) / n
	rem := len(c) % n

	idx := 0
	for i := 0; i < n; i++ {
		size := subsetSize
		if i < rem {
			size++
		}
		subsets[i] = c[idx : idx+size]
		idx += size
	}

	return subsets
}

func computeScheduleKey(ops []domain.ScheduledOp) string {
	var sb strings.Builder
	for _, op := range ops {
		sb.WriteString(strconv.Itoa(op.ID))
		sb.WriteByte(',')
	}
	return sb.String()
}

// FailureSignatureFor derives the stable identity used to compare replay and shrink outcomes.
func FailureSignatureFor(result *domain.ExecutionResult) (domain.FailureSignature, error) {
	if result == nil {
		return domain.FailureSignature{}, fmt.Errorf("cannot identify failure from a nil result")
	}
	signature := domain.FailureSignature{Status: result.Status}
	switch result.Status {
	case domain.StatusPassed:
		return signature, nil
	case domain.StatusViolation:
		if !result.ViolationDetected {
			return domain.FailureSignature{}, fmt.Errorf("violation status is missing the violation marker")
		}
		if result.FailingInvariant == nil || result.FailingInvariant.Name == "" {
			return domain.FailureSignature{}, fmt.Errorf("violation result has no failing invariant")
		}
		signature.FailingInvariant = result.FailingInvariant.Name
		return signature, nil
	default:
		return domain.FailureSignature{}, fmt.Errorf("execution status %q does not have a stable replay signature", result.Status)
	}
}

// ReproducesFailure reports whether result has the same stable failure identity.
func ReproducesFailure(result *domain.ExecutionResult, target domain.FailureSignature) bool {
	if result == nil || result.Status != target.Status {
		return false
	}
	if target.Status == domain.StatusPassed {
		return result.Status == domain.StatusPassed
	}
	if target.Status != domain.StatusViolation {
		return false
	}
	return result.ViolationDetected && result.FailingInvariant != nil &&
		result.FailingInvariant.Name == target.FailingInvariant
}

// Shrink applies the Zeller Delta-Debugging algorithm with memoization to find a minimal failing subset
func Shrink(ctx context.Context, testFn func([]domain.ScheduledOp) bool, initialOps []domain.ScheduledOp) (*domain.ShrinkResult, error) {
	start := time.Now()
	memo := make(map[string]bool)
	trials := 0
	cachedTestFn := func(ops []domain.ScheduledOp) (bool, error) {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		key := computeScheduleKey(ops)
		if res, ok := memo[key]; ok {
			return res, nil
		}
		trials++
		res := testFn(ops)
		if err := ctx.Err(); err != nil {
			return false, err
		}
		memo[key] = res
		return res, nil
	}

	// The initial set must fail to be shrinkable. In testFn, false means FAIL.
	initialPassed, err := cachedTestFn(initialOps)
	if err != nil {
		return nil, err
	}
	if initialPassed {
		return nil, fmt.Errorf("initial operations do not fail the test")
	}
	baselinePassed, err := cachedTestFn(nil)
	if err != nil {
		return nil, err
	}
	if !baselinePassed {
		return nil, ErrBaselineFailure
	}

	c := initialOps
	n := 2
	iterations := 0

	for len(c) >= 2 {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		iterations++

		subsets := partition(c, n)
		someComplementFailed := false

		for i := 0; i < n; i++ {
			complement := make([]domain.ScheduledOp, 0, len(c))
			for j := 0; j < n; j++ {
				if i != j {
					complement = append(complement, subsets[j]...)
				}
			}

			passed, err := cachedTestFn(complement)
			if err != nil {
				return nil, err
			}
			if !passed { // Failed, which means we can reduce to complement
				c = complement
				n = max(n-1, 2)
				someComplementFailed = true
				break
			}
		}

		if !someComplementFailed {
			someSubsetFailed := false
			for i := 0; i < n; i++ {
				passed, err := cachedTestFn(subsets[i])
				if err != nil {
					return nil, err
				}
				if !passed { // Failed, reduce to this subset
					c = subsets[i]
					n = 2
					someSubsetFailed = true
					break
				}
			}

			if !someSubsetFailed {
				if n == len(c) {
					break
				}
				n = min(2*n, len(c))
			}
		}
	}

	// Audit and enforce 1-minimality explicitly, independent of ddmin partition history.
	for len(c) > 0 {
		reduced := false
		for index := range c {
			candidate := make([]domain.ScheduledOp, 0, len(c)-1)
			candidate = append(candidate, c[:index]...)
			candidate = append(candidate, c[index+1:]...)
			passed, err := cachedTestFn(candidate)
			if err != nil {
				return nil, err
			}
			if !passed {
				if len(candidate) == 0 {
					return nil, ErrBaselineFailure
				}
				c = candidate
				reduced = true
				break
			}
		}
		if !reduced {
			break
		}
	}

	result := &domain.ShrinkResult{
		OriginalSize:   len(initialOps),
		ReducedSize:    len(c),
		ReductionRatio: float64(len(initialOps)-len(c)) / float64(len(initialOps)) * 100.0,
		MinimalOps:     c,
		Iterations:     iterations,
		Trials:         trials,
		Duration:       time.Since(start),
	}

	return result, nil
}

// ShrinkExecution executes Shrink using the engine's Runner
func ShrinkExecution(ctx context.Context, runner *engine.Runner, spec domain.Spec, initialOps []domain.ScheduledOp) (*domain.ShrinkResult, error) {
	initialResult, err := runner.RunSchedule(ctx, spec, initialOps)
	if err != nil {
		return nil, err
	}
	target, err := FailureSignatureFor(initialResult)
	if err != nil {
		return nil, err
	}
	testFn := func(ops []domain.ScheduledOp) bool {
		res, err := runner.RunSchedule(ctx, spec, ops)
		if err != nil {
			return true
		}
		return !ReproducesFailure(res, target)
	}
	return Shrink(ctx, testFn, initialOps)
}
