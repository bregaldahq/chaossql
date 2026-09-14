package domain_test

import (
	"encoding/json"
	"errors"
	"strings"
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

func TestSpecValidateIsolation(t *testing.T) {
	validLevels := []domain.IsolationLevel{
		"",
		domain.LevelReadUncommitted,
		domain.LevelReadCommitted,
		domain.LevelRepeatableRead,
		domain.LevelSerializable,
	}

	for _, level := range validLevels {
		spec := validDomainSpec()
		spec.Database.Isolation = level
		if err := spec.Validate(); err != nil {
			t.Fatalf("isolation %q should be valid: %v", level, err)
		}
	}

	spec := validDomainSpec()
	spec.Database.Isolation = domain.IsolationLevel("SNAPSHOT")
	err := spec.Validate()
	if !errors.Is(err, domain.ErrSpecValidationFailed) {
		t.Fatalf("expected ErrSpecValidationFailed, got %v", err)
	}
	if !strings.Contains(err.Error(), "unsupported database isolation") {
		t.Fatalf("expected isolation detail, got %v", err)
	}
}

func TestExecutionStatusJSON(t *testing.T) {
	statuses := []domain.ExecutionStatus{
		domain.StatusPassed,
		domain.StatusViolation,
		domain.StatusExecutionError,
		domain.StatusInconclusive,
		domain.StatusCanceled,
	}
	want := []string{"passed", "violation", "execution_error", "inconclusive", "canceled"}

	for i, status := range statuses {
		if string(status) != want[i] {
			t.Fatalf("status %d = %q, want %q", i, status, want[i])
		}
	}

	result := domain.ExecutionResult{
		Status:    domain.StatusExecutionError,
		Isolation: domain.LevelSerializable,
		OperationErrors: []domain.OperationError{{
			OperationID: 7,
			Operation:   "withdraw",
			StepIndex:   2,
			Phase:       "step",
			Message:     "syntax error",
		}},
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(encoded)
	for _, fragment := range []string{
		`"status":"execution_error"`,
		`"isolation":"SERIALIZABLE"`,
		`"operation_errors":[`,
		`"operation_id":7`,
		`"phase":"step"`,
	} {
		if !strings.Contains(jsonText, fragment) {
			t.Fatalf("expected %s in %s", fragment, jsonText)
		}
	}
}

func validDomainSpec() domain.Spec {
	return domain.Spec{
		Version:  "1.0",
		Name:     "transaction_contract",
		Database: domain.DatabaseConfig{Driver: "sqlite"},
		Invariants: []domain.InvariantConfig{{
			Name: "balance",
		}},
		Operations: []domain.OperationConfig{{
			Name: "read",
		}},
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
