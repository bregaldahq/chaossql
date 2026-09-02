package shrinker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

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

// Shrink applies the Zeller Delta-Debugging algorithm with memoization to find a minimal failing subset
func Shrink(ctx context.Context, testFn func([]domain.ScheduledOp) bool, initialOps []domain.ScheduledOp) (*domain.ShrinkResult, error) {
	start := time.Now()

	// initial set must fail to be shrinkable. In testFn, false means FAIL (reproduces bug), true means PASS.
	if testFn(initialOps) {
		return nil, fmt.Errorf("initial operations do not fail the test")
	}

	memo := make(map[string]bool)
	cachedTestFn := func(ops []domain.ScheduledOp) bool {
		if len(ops) == 0 {
			return true
		}
		key := computeScheduleKey(ops)
		if res, ok := memo[key]; ok {
			return res
		}
		res := testFn(ops)
		memo[key] = res
		return res
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

			if !cachedTestFn(complement) { // Failed, which means we can reduce to complement
				c = complement
				n = max(n-1, 2)
				someComplementFailed = true
				break
			}
		}

		if !someComplementFailed {
			someSubsetFailed := false
			for i := 0; i < n; i++ {
				if !cachedTestFn(subsets[i]) { // Failed, reduce to this subset
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

	result := &domain.ShrinkResult{
		OriginalSize:   len(initialOps),
		ReducedSize:    len(c),
		ReductionRatio: float64(len(initialOps)-len(c)) / float64(len(initialOps)) * 100.0,
		MinimalOps:     c,
		Iterations:     iterations,
		Duration:       time.Since(start),
	}

	return result, nil
}

// ShrinkExecution executes Shrink using the engine's Runner
func ShrinkExecution(ctx context.Context, runner *engine.Runner, spec domain.Spec, initialOps []domain.ScheduledOp) (*domain.ShrinkResult, error) {
	testFn := func(ops []domain.ScheduledOp) bool {
		res, err := runner.RunSchedule(ctx, spec, ops)
		if err != nil {
			return true
		}
		return res.Success
	}
	return Shrink(ctx, testFn, initialOps)
}
