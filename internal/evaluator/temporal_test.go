package evaluator_test

import (
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/evaluator"
)

func TestEvaluateTemporalInvariants(t *testing.T) {
	trace := domain.ExecutionTrace{
		{Timestamp: 100 * time.Microsecond, WorkerID: 1, Type: domain.EventBegin, SQL: "BEGIN"},
		{Timestamp: 200 * time.Microsecond, WorkerID: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 100 WHERE id = 1"},
		{Timestamp: 300 * time.Microsecond, WorkerID: 1, Type: domain.EventCommit, SQL: "COMMIT"},
	}

	invs := []domain.TemporalInvariantConfig{
		{Name: "no_abort_check", Type: "no_aborts"},
		{Name: "no_error_check", Type: "no_error_events"},
		{Name: "monotonic_check", Type: "monotonicity"},
	}

	results := evaluator.EvaluateTemporalInvariants(trace, invs)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Passed {
			t.Errorf("expected invariant %s to pass, got failed", r.Name)
		}
	}
}

func TestEvaluateTemporalInvariants_FailureCases(t *testing.T) {
	trace := domain.ExecutionTrace{
		{Timestamp: 100 * time.Microsecond, WorkerID: 1, Type: domain.EventBegin, SQL: "BEGIN"},
		{Timestamp: 200 * time.Microsecond, WorkerID: 1, Type: domain.EventError, Error: "deadlock"},
		{Timestamp: 300 * time.Microsecond, WorkerID: 1, Type: domain.EventRollback, SQL: "ROLLBACK"},
	}

	invs := []domain.TemporalInvariantConfig{
		{Name: "no_abort_check", Type: "no_aborts"},
		{Name: "no_error_check", Type: "no_error_events"},
	}

	results := evaluator.EvaluateTemporalInvariants(trace, invs)
	for _, r := range results {
		if r.Passed {
			t.Errorf("expected invariant %s to fail on aborts/errors, but passed", r.Name)
		}
	}
}
