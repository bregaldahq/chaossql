package shrinker

import (
	"context"
	"fmt"
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

// Shrink applies the Zeller Delta-Debugging algorithm to find a minimal failing subset
func Shrink(ctx context.Context, testFn func([]domain.ScheduledOp) bool, initialOps []domain.ScheduledOp) (*domain.ShrinkResult, error) {
	start := time.Now()

	// initial set must fail to be shrinkable. In testFn, false means FAIL (reproduces bug), true means PASS.
	if testFn(initialOps) {
		return nil, fmt.Errorf("initial operations do not fail the test")
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
			complement := make([]domain.ScheduledOp, 0)
			for j := 0; j < n; j++ {
				if i != j {
					complement = append(complement, subsets[j]...)
				}
			}

			if !testFn(complement) { // Failed, which means we can reduce to complement
				c = complement
				n = max(n-1, 2)
				someComplementFailed = true
				break
			}
		}

		if !someComplementFailed {
			someSubsetFailed := false
			for i := 0; i < n; i++ {
				if !testFn(subsets[i]) { // Failed, reduce to this subset
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
			// If infrastructure fails, we consider it a pass (it doesn't reproduce the invariant violation)
			return true
		}
		// res.Success is true if NO violation detected.
		// If Success == true, it PASSES. So we return true.
		// If Success == false, it FAILS (reproduces bug). So we return false.
		return res.Success
	}
	return Shrink(ctx, testFn, initialOps)
}
