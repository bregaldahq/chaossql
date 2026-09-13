package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestRunner_StatusPassed(t *testing.T) {
	result := runStatusScenario(t, []domain.StepConfig{{SQL: "SELECT 1"}}, domain.InvariantConfig{
		Name: "balance_unchanged", Query: "SELECT balance FROM accounts WHERE id = 1", Assert: "balance == 10",
	})
	assertExecutionStatus(t, result, domain.StatusPassed, true, false)
}

func TestRunner_StatusViolation(t *testing.T) {
	result := runStatusScenario(t, []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"}}, domain.InvariantConfig{
		Name: "balance_unchanged", Query: "SELECT balance FROM accounts WHERE id = 1", Assert: "balance == 10",
	})
	assertExecutionStatus(t, result, domain.StatusViolation, false, true)
	if result.FailingInvariant == nil || result.FailingInvariant.Name != "balance_unchanged" {
		t.Fatalf("missing failing invariant: %+v", result.FailingInvariant)
	}
}

func TestRunner_StatusInconclusive(t *testing.T) {
	result := runStatusScenario(t, []domain.StepConfig{{SQL: "SELECT 1"}}, domain.InvariantConfig{
		Name: "unavailable_evidence", Query: "SELECT value FROM missing_table", Assert: "value == 1",
	})
	assertExecutionStatus(t, result, domain.StatusInconclusive, false, false)
	if result.Error == nil || result.FailingInvariant == nil || result.FailingInvariant.Error == nil {
		t.Fatalf("expected invariant evaluation error: %+v", result)
	}
}

func TestRunner_StatusExecutionErrorPrecedesInvariantEvaluation(t *testing.T) {
	result := runStatusScenario(t, []domain.StepConfig{
		{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"},
		{SQL: "NOT VALID SQL"},
	}, domain.InvariantConfig{
		Name: "would_be_inconclusive", Query: "SELECT value FROM missing_table", Assert: "value == 1",
	})
	assertExecutionStatus(t, result, domain.StatusExecutionError, false, false)
	if len(result.OperationErrors) != 1 || result.FailingInvariant != nil {
		t.Fatalf("unexpected execution error result: %+v", result)
	}
}

func TestRunner_StatusCanceledPrecedesInvariantEvaluation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	driver := &cancelAfterResetDriver{
		DatabaseDriver: drivers.NewSQLiteDriver(""),
		cancel:         cancel,
	}
	t.Cleanup(func() { _ = driver.Close() })

	spec := domain.Spec{
		Database: domain.DatabaseConfig{
			Driver:    "sqlite",
			Isolation: domain.LevelSerializable,
			Schema:    "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);",
			Seed:      "INSERT INTO accounts VALUES (1, 10);",
		},
		Engine: domain.EngineConfig{Workers: 1},
		Invariants: []domain.InvariantConfig{{
			Name: "must_not_run", Query: "SELECT value FROM missing_table", Assert: "value == 1",
		}},
	}
	result, err := engine.NewRunner(driver, 1).RunSchedule(ctx, spec, []domain.ScheduledOp{{
		ID: 1, Name: "canceled_operation", Steps: []domain.StepConfig{{SQL: "SELECT 1"}},
	}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run schedule error = %v, want context canceled", err)
	}
	assertExecutionStatus(t, result, domain.StatusCanceled, false, false)
	if result.FailingInvariant != nil || !errors.Is(result.Error, context.Canceled) {
		t.Fatalf("unexpected canceled result: %+v", result)
	}
}

func TestRunner_RejectsUnsupportedIsolationBeforeReset(t *testing.T) {
	driver := &resetSpyDriver{DatabaseDriver: drivers.NewMockDriver()}
	t.Cleanup(func() { _ = driver.Close() })
	spec := domain.Spec{
		Database: domain.DatabaseConfig{Isolation: domain.IsolationLevel("INVALID")},
		Engine:   domain.EngineConfig{Workers: 1},
	}

	result, err := engine.NewRunner(driver, 1).RunSchedule(context.Background(), spec, nil)
	if err == nil || result != nil {
		t.Fatalf("result=%+v error=%v, want unsupported isolation failure", result, err)
	}
	if driver.resetCalled {
		t.Fatal("database reset occurred before isolation validation")
	}
}

type cancelAfterResetDriver struct {
	drivers.DatabaseDriver
	cancel context.CancelFunc
}

type resetSpyDriver struct {
	drivers.DatabaseDriver
	resetCalled bool
}

func (d *resetSpyDriver) Reset(context.Context, string, string) error {
	d.resetCalled = true
	return nil
}

func (d *cancelAfterResetDriver) Reset(_ context.Context, schemaSQL, seedSQL string) error {
	if err := d.DatabaseDriver.Reset(context.Background(), schemaSQL, seedSQL); err != nil {
		return err
	}
	d.cancel()
	return nil
}

func runStatusScenario(t *testing.T, steps []domain.StepConfig, invariant domain.InvariantConfig) *domain.ExecutionResult {
	t.Helper()
	driver := drivers.NewSQLiteDriver("")
	t.Cleanup(func() { _ = driver.Close() })
	spec := domain.Spec{
		Database: domain.DatabaseConfig{
			Driver:    "sqlite",
			Isolation: domain.LevelSerializable,
			Schema:    "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);",
			Seed:      "INSERT INTO accounts VALUES (1, 10);",
		},
		Engine:     domain.EngineConfig{Workers: 1},
		Invariants: []domain.InvariantConfig{invariant},
	}
	result, err := engine.NewRunner(driver, 1).RunSchedule(context.Background(), spec, []domain.ScheduledOp{{
		ID: 1, Name: "status_operation", Steps: steps,
	}})
	if err != nil {
		t.Fatalf("run schedule: %v", err)
	}
	return result
}

func assertExecutionStatus(t *testing.T, result *domain.ExecutionResult, status domain.ExecutionStatus, success, violation bool) {
	t.Helper()
	if result.Status != status || result.Success != success || result.ViolationDetected != violation {
		t.Fatalf("status=%q success=%v violation=%v, want %q/%v/%v", result.Status, result.Success, result.ViolationDetected, status, success, violation)
	}
}
