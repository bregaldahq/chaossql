package shrinker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestFailureSignatureMatchesOnlyOriginalInvariant(t *testing.T) {
	original := &domain.ExecutionResult{
		Status:            domain.StatusViolation,
		ViolationDetected: true,
		FailingInvariant:  &domain.InvariantResult{Name: "balance_preserved"},
	}
	signature, err := FailureSignatureFor(original)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		result *domain.ExecutionResult
		want   bool
	}{
		{
			name: "same invariant",
			result: &domain.ExecutionResult{Status: domain.StatusViolation, ViolationDetected: true,
				FailingInvariant: &domain.InvariantResult{Name: "balance_preserved"}},
			want: true,
		},
		{
			name: "different invariant",
			result: &domain.ExecutionResult{Status: domain.StatusViolation, ViolationDetected: true,
				FailingInvariant: &domain.InvariantResult{Name: "inventory_nonnegative"}},
		},
		{name: "execution error", result: &domain.ExecutionResult{Status: domain.StatusExecutionError}},
		{name: "nil result", result: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReproducesFailure(tt.result, signature); got != tt.want {
				t.Fatalf("ReproducesFailure() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestFailureSignatureRejectsUnidentifiedViolation(t *testing.T) {
	_, err := FailureSignatureFor(&domain.ExecutionResult{Status: domain.StatusViolation, ViolationDetected: true})
	if err == nil {
		t.Fatal("expected unidentified violation to be rejected")
	}
}

func TestFailureSignatureRejectsUnstableExecutionStatus(t *testing.T) {
	_, err := FailureSignatureFor(&domain.ExecutionResult{Status: domain.StatusExecutionError})
	if err == nil {
		t.Fatal("expected execution error without a stable identity to be rejected")
	}
}

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

func TestShrink_CanceledContextSkipsOracle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	_, err := Shrink(ctx, func([]domain.ScheduledOp) bool {
		calls++
		return false
	}, []domain.ScheduledOp{{ID: 1}, {ID: 2}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if calls != 0 {
		t.Fatalf("oracle called %d times after cancellation", calls)
	}
}

func TestShrink_PropagatesCancellationRaisedByOracle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	_, err := Shrink(ctx, func([]domain.ScheduledOp) bool {
		cancel()
		return false
	}, []domain.ScheduledOp{{ID: 1}, {ID: 2}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestShrink_RejectsFailureWithoutOperations(t *testing.T) {
	initial := []domain.ScheduledOp{{ID: 1}}
	result, err := Shrink(context.Background(), func([]domain.ScheduledOp) bool {
		return false
	}, initial)
	if !errors.Is(err, ErrBaselineFailure) {
		t.Fatalf("result = %#v, error = %v, want ErrBaselineFailure", result, err)
	}
}

func TestShrink_ReportsActualOracleTrials(t *testing.T) {
	initial := []domain.ScheduledOp{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}
	calls := 0
	result, err := Shrink(context.Background(), func(ops []domain.ScheduledOp) bool {
		calls++
		for _, op := range ops {
			if op.ID == 2 {
				return false
			}
		}
		return true
	}, initial)
	if err != nil {
		t.Fatal(err)
	}
	if result.Trials != calls {
		t.Fatalf("reported trials = %d, actual oracle calls = %d", result.Trials, calls)
	}
	if result.Trials == 0 {
		t.Fatal("expected at least one oracle trial")
	}
}

func TestShrink_DeterministicTieBreaking(t *testing.T) {
	initial := []domain.ScheduledOp{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}
	oracle := func(ops []domain.ScheduledOp) bool {
		for _, op := range ops {
			if op.ID == 1 || op.ID == 2 {
				return false
			}
		}
		return true
	}
	var want []domain.ScheduledOp
	for run := 0; run < 25; run++ {
		result, err := Shrink(context.Background(), oracle, initial)
		if err != nil {
			t.Fatal(err)
		}
		if run == 0 {
			want = result.MinimalOps
			continue
		}
		if !reflect.DeepEqual(result.MinimalOps, want) {
			t.Fatalf("run %d result = %#v, want %#v", run, result.MinimalOps, want)
		}
	}
}

func TestShrinkExecution_PreservesOriginalInvariantAcrossRuns(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	if err := driver.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer driver.Close()
	spec := domain.Spec{
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "CREATE TABLE state (id INT PRIMARY KEY, value INT);",
			Seed:   "INSERT INTO state VALUES (1, 100), (2, 100);",
		},
		Engine: domain.EngineConfig{Workers: 1, Seed: 33},
		Invariants: []domain.InvariantConfig{
			{Name: "first_unchanged", Query: "SELECT value FROM state WHERE id = 1", Assert: "value == 100"},
			{Name: "second_unchanged", Query: "SELECT value FROM state WHERE id = 2", Assert: "value == 100"},
		},
	}
	ops := []domain.ScheduledOp{
		{ID: 1, Name: "break_first", Steps: []domain.StepConfig{{SQL: "UPDATE state SET value = 0 WHERE id = 1"}}},
		{ID: 2, Name: "break_second", Steps: []domain.StepConfig{{SQL: "UPDATE state SET value = 0 WHERE id = 2"}}},
	}
	runner := engine.NewRunner(driver, spec.Engine.Seed)
	for run := 0; run < 20; run++ {
		result, err := ShrinkExecution(ctx, runner, spec, ops)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if len(result.MinimalOps) != 1 || result.MinimalOps[0].ID != 1 {
			t.Fatalf("run %d minimized to %#v, want only original failure operation 1", run, result.MinimalOps)
		}
	}
}
