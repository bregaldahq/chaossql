package domain_test

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestInvariantResult_String(t *testing.T) {
	passResult := domain.InvariantResult{
		Name:   "saldo_positivo",
		Passed: true,
	}
	if passResult.String() != "PASS: Invariant 'saldo_positivo' satisfied" {
		t.Errorf("unexpected string output: %s", passResult.String())
	}

	failResult := domain.InvariantResult{
		Name:         "saldo_positivo",
		Passed:       false,
		Expression:   "balance >= 0",
		ActualValues: map[string]interface{}{"balance": -50},
	}
	if failResult.Passed {
		t.Error("expected failure result")
	}
}

func TestTraceEventTypes(t *testing.T) {
	tests := []struct {
		evType   domain.TraceEventType
		expected string
	}{
		{domain.EventBegin, "BEGIN"},
		{domain.EventExec, "EXEC"},
		{domain.EventCommit, "COMMIT"},
		{domain.EventRollback, "ROLLBACK"},
		{domain.EventError, "ERROR"},
		{domain.EventSavepoint, "SAVEPOINT"},
		{domain.EventRollbackTo, "ROLLBACK_TO"},
		{domain.EventReleaseSavepoint, "RELEASE_SAVEPOINT"},
	}

	for _, tc := range tests {
		if string(tc.evType) != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, tc.evType)
		}
	}
}
