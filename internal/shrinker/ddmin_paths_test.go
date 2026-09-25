package shrinker

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestFailureSignatureFor_EdgeCases(t *testing.T) {
	if _, err := FailureSignatureFor(nil); err == nil {
		t.Fatal("nil result must be rejected")
	}
	signature, err := FailureSignatureFor(&domain.ExecutionResult{Status: domain.StatusPassed})
	if err != nil || signature.Status != domain.StatusPassed || signature.FailingInvariant != "" {
		t.Fatalf("passed result signature = %+v, err = %v", signature, err)
	}
	_, err = FailureSignatureFor(&domain.ExecutionResult{
		Status:           domain.StatusViolation,
		FailingInvariant: &domain.InvariantResult{Name: "balance_preserved"},
	})
	if err == nil || !strings.Contains(err.Error(), "violation marker") {
		t.Fatalf("violation without marker: err = %v", err)
	}
}

func TestReproducesFailure_NonViolationTargets(t *testing.T) {
	passedTarget := domain.FailureSignature{Status: domain.StatusPassed}
	if !ReproducesFailure(&domain.ExecutionResult{Status: domain.StatusPassed}, passedTarget) {
		t.Fatal("a passing result must reproduce a passing target")
	}
	errorTarget := domain.FailureSignature{Status: domain.StatusExecutionError}
	if ReproducesFailure(&domain.ExecutionResult{Status: domain.StatusExecutionError}, errorTarget) {
		t.Fatal("execution errors have no stable identity and must never match")
	}
}

// A non-monotonic oracle where every complement passes but one small subset
// fails forces ddmin through its "reduce to subset" branch.
func TestShrink_ReducesToFailingSubset(t *testing.T) {
	var ops []domain.ScheduledOp
	for i := 0; i < 8; i++ {
		ops = append(ops, domain.ScheduledOp{ID: i})
	}
	testFn := func(candidate []domain.ScheduledOp) bool {
		if len(candidate) == len(ops) {
			return false
		}
		if len(candidate) > 2 {
			return true
		}
		for _, op := range candidate {
			if op.ID == 3 {
				return false
			}
		}
		return true
	}

	result, err := Shrink(context.Background(), testFn, ops)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.MinimalOps) != 1 || result.MinimalOps[0].ID != 3 {
		t.Fatalf("MinimalOps = %#v, want only operation 3", result.MinimalOps)
	}
}

func TestShrink_StopsWhenContextIsCanceledBetweenIterations(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var ops []domain.ScheduledOp
	for i := 0; i < 16; i++ {
		ops = append(ops, domain.ScheduledOp{ID: i})
	}
	calls := 0
	testFn := func(candidate []domain.ScheduledOp) bool {
		calls++
		// Initial run and baseline complete; cancel during the first reduction.
		if calls == 3 {
			cancel()
		}
		for _, op := range candidate {
			if op.ID == 11 {
				return false
			}
		}
		return true
	}

	if _, err := Shrink(ctx, testFn, ops); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestShrinkExecution_RejectsUnidentifiableInitialFailure(t *testing.T) {
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
			Seed:   "INSERT INTO state VALUES (1, 100);",
		},
		Engine: domain.EngineConfig{Workers: 1, Seed: 7},
		Invariants: []domain.InvariantConfig{
			{Name: "unchanged", Query: "SELECT value FROM state WHERE id = 1", Assert: "value == 100"},
		},
	}
	ops := []domain.ScheduledOp{
		{ID: 1, Name: "bad_sql", Steps: []domain.StepConfig{{SQL: "UPDATE missing_table SET value = 0"}}},
	}

	result, err := ShrinkExecution(ctx, engine.NewRunner(driver, spec.Engine.Seed), spec, ops)
	if err == nil {
		t.Fatalf("result = %+v, want an error for a schedule without a stable failure", result)
	}
}
