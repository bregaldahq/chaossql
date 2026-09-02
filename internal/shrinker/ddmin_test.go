package shrinker

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestShrink_SyntheticOracle(t *testing.T) {
	ctx := context.Background()

	// 100 operations
	var initialOps []domain.ScheduledOp
	for i := 0; i < 100; i++ {
		initialOps = append(initialOps, domain.ScheduledOp{
			ID:   i,
			Name: "OP",
		})
	}

	// We define a bug that is reproduced ONLY if operation with ID 42 AND ID 77 are present.
	testFn := func(ops []domain.ScheduledOp) bool {
		has42 := false
		has77 := false
		for _, op := range ops {
			if op.ID == 42 {
				has42 = true
			}
			if op.ID == 77 {
				has77 = true
			}
		}
		if has42 && has77 {
			return false // FAILS (reproduces bug)
		}
		return true // PASSES
	}

	res, err := Shrink(ctx, testFn, initialOps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ReducedSize != 2 {
		t.Errorf("expected 2 ops, got %d", res.ReducedSize)
	}

	if res.ReductionRatio < 95.0 {
		t.Errorf("expected > 95%% reduction, got %.2f%%", res.ReductionRatio)
	}

	// 1-minimality verification test
	// Asserting that removing any single remaining operation causes the oracle to pass.
	for i := 0; i < len(res.MinimalOps); i++ {
		complement := make([]domain.ScheduledOp, 0)
		for j, op := range res.MinimalOps {
			if i != j {
				complement = append(complement, op)
			}
		}
		
		if !testFn(complement) {
			t.Errorf("not 1-minimal! Removing element at index %d still reproduces the bug", i)
		}
	}
}

func TestShrink_NoBug(t *testing.T) {
	ctx := context.Background()
	initialOps := []domain.ScheduledOp{{ID: 1}}
	testFn := func(ops []domain.ScheduledOp) bool {
		return true // ALWAYS PASSES
	}
	
	_, err := Shrink(ctx, testFn, initialOps)
	if err == nil {
		t.Errorf("expected error when initial ops don't fail")
	}
}
