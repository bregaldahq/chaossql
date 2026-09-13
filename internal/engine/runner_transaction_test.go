package engine_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

var errBeginTransaction = errors.New("begin transaction failed")
var errCommitTransaction = errors.New("commit transaction failed")
var errRollbackTransaction = errors.New("rollback transaction failed")

func TestRunner_RollsBackFailedOperation(t *testing.T) {
	driver := newTransactionTestDriver(t)
	runner := engine.NewRunner(driver, 1)

	outcome, err := runner.ExecuteSchedule(context.Background(), transactionTestSpec(), []domain.ScheduledOp{{
		ID:   1,
		Name: "fail_after_write",
		Steps: []domain.StepConfig{
			{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"},
			{SQL: "NOT VALID SQL"},
		},
	}})
	if err != nil {
		t.Fatalf("execute schedule: %v", err)
	}
	if len(outcome.OperationErrors) != 1 {
		t.Fatalf("expected one operation error, got %+v", outcome.OperationErrors)
	}
	opErr := outcome.OperationErrors[0]
	if opErr.OperationID != 1 || opErr.StepIndex != 2 || opErr.Phase != "step" {
		t.Fatalf("unexpected operation error: %+v", opErr)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventBegin, domain.EventExec, domain.EventError, domain.EventRollback)
	assertBalance(t, driver, 10)
}

func TestRunner_CommitsSuccessfulOperation(t *testing.T) {
	driver := newTransactionTestDriver(t)
	runner := engine.NewRunner(driver, 1)

	outcome, err := runner.ExecuteSchedule(context.Background(), transactionTestSpec(), []domain.ScheduledOp{{
		ID:   1,
		Name: "commit_write",
		Steps: []domain.StepConfig{
			{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"},
		},
	}})
	if err != nil {
		t.Fatalf("execute schedule: %v", err)
	}
	if len(outcome.OperationErrors) != 0 {
		t.Fatalf("unexpected operation errors: %+v", outcome.OperationErrors)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventBegin, domain.EventExec, domain.EventCommit)
	assertBalance(t, driver, 9)
}

func TestRunner_RecordsBeginFailureWithoutSyntheticLifecycle(t *testing.T) {
	base := drivers.NewMockDriver()
	t.Cleanup(func() { _ = base.Close() })
	runner := engine.NewRunner(beginFailDriver{DatabaseDriver: base}, 1)

	outcome, err := runner.ExecuteSchedule(context.Background(), transactionTestSpec(), []domain.ScheduledOp{{
		ID:    3,
		Name:  "cannot_begin",
		Steps: []domain.StepConfig{{SQL: "SELECT 1"}},
	}})
	if err != nil {
		t.Fatalf("execute schedule: %v", err)
	}
	if len(outcome.OperationErrors) != 1 || outcome.OperationErrors[0].Phase != "begin" {
		t.Fatalf("unexpected operation errors: %+v", outcome.OperationErrors)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventError)
	if outcome.Trace[0].Phase != "begin" || outcome.Trace[0].Error != errBeginTransaction.Error() {
		t.Fatalf("unexpected begin trace: %+v", outcome.Trace[0])
	}
}

func TestRunner_CancellationRollsBackPromptly(t *testing.T) {
	driver := drivers.NewSQLiteDriver(filepath.Join(t.TempDir(), "cancellation.db"))
	t.Cleanup(func() { _ = driver.Close() })
	if err := driver.Reset(
		context.Background(),
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);",
		"INSERT INTO accounts VALUES (1, 10);",
	); err != nil {
		t.Fatalf("reset driver: %v", err)
	}
	runner := engine.NewRunner(driver, 1)
	spec := transactionTestSpec()
	spec.Engine.JitterMs = [2]int{500, 500}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	started := time.Now()
	outcome, err := runner.ExecuteSchedule(ctx, spec, []domain.ScheduledOp{{
		ID: 4, Name: "cancel_during_jitter", Steps: []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"}},
	}})
	elapsed := time.Since(started)

	if !errors.Is(err, context.DeadlineExceeded) || !outcome.Canceled {
		t.Fatalf("outcome=%+v error=%v, want canceled deadline", outcome, err)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("cancellation took %s, want at most 200ms", elapsed)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventBegin, domain.EventRollback)
	assertBalance(t, driver, 10)
}

func TestRunner_CommitFailureAttemptsRollback(t *testing.T) {
	driver := newTransactionTestDriver(t)
	runner := engine.NewRunner(commitFailDriver{DatabaseDriver: driver}, 1)

	outcome, err := runner.ExecuteSchedule(context.Background(), transactionTestSpec(), []domain.ScheduledOp{{
		ID: 5, Name: "commit_failure", Steps: []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"}},
	}})
	if err != nil {
		t.Fatalf("execute schedule: %v", err)
	}
	if len(outcome.OperationErrors) != 1 || outcome.OperationErrors[0].Phase != "commit" {
		t.Fatalf("unexpected operation errors: %+v", outcome.OperationErrors)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventBegin, domain.EventExec, domain.EventError, domain.EventRollback)
	assertBalance(t, driver, 10)
}

func TestRunner_RollbackFailureIsRecordedAsError(t *testing.T) {
	driver := newTransactionTestDriver(t)
	runner := engine.NewRunner(rollbackFailDriver{DatabaseDriver: driver}, 1)

	outcome, err := runner.ExecuteSchedule(context.Background(), transactionTestSpec(), []domain.ScheduledOp{{
		ID: 6, Name: "rollback_failure", Steps: []domain.StepConfig{{SQL: "NOT VALID SQL"}},
	}})
	if err != nil {
		t.Fatalf("execute schedule: %v", err)
	}
	if len(outcome.OperationErrors) != 2 || outcome.OperationErrors[0].Phase != "rollback" || outcome.OperationErrors[1].Phase != "step" {
		t.Fatalf("unexpected operation errors: %+v", outcome.OperationErrors)
	}
	assertTransactionEvents(t, outcome.Trace, domain.EventBegin, domain.EventError, domain.EventError)
	if outcome.Trace[2].Phase != "rollback" || outcome.Trace[2].Error != errRollbackTransaction.Error() {
		t.Fatalf("unexpected rollback trace: %+v", outcome.Trace[2])
	}
}

type beginFailDriver struct {
	drivers.DatabaseDriver
}

type commitFailDriver struct {
	drivers.DatabaseDriver
}

func (d commitFailDriver) BeginTx(ctx context.Context, opts drivers.TransactionOptions) (drivers.Tx, error) {
	tx, err := d.DatabaseDriver.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return commitFailTx{Tx: tx}, nil
}

type commitFailTx struct {
	drivers.Tx
}

func (commitFailTx) Commit() error {
	return errCommitTransaction
}

type rollbackFailDriver struct {
	drivers.DatabaseDriver
}

func (d rollbackFailDriver) BeginTx(ctx context.Context, opts drivers.TransactionOptions) (drivers.Tx, error) {
	tx, err := d.DatabaseDriver.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return rollbackFailTx{Tx: tx}, nil
}

type rollbackFailTx struct {
	drivers.Tx
}

func (tx rollbackFailTx) Rollback() error {
	_ = tx.Tx.Rollback()
	return errRollbackTransaction
}

func (beginFailDriver) BeginTx(context.Context, drivers.TransactionOptions) (drivers.Tx, error) {
	return nil, errBeginTransaction
}

func newTransactionTestDriver(t *testing.T) *drivers.SQLiteDriver {
	t.Helper()
	driver := drivers.NewSQLiteDriver("")
	t.Cleanup(func() { _ = driver.Close() })
	if err := driver.Reset(
		context.Background(),
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);",
		"INSERT INTO accounts VALUES (1, 10);",
	); err != nil {
		t.Fatalf("reset driver: %v", err)
	}
	return driver
}

func transactionTestSpec() domain.Spec {
	return domain.Spec{
		Database: domain.DatabaseConfig{Isolation: domain.LevelSerializable},
		Engine:   domain.EngineConfig{Workers: 1},
	}
}

func assertBalance(t *testing.T, driver drivers.DatabaseDriver, expected int) {
	t.Helper()
	var actual int
	if err := driver.QueryRow(context.Background(), "SELECT balance FROM accounts WHERE id = 1").Scan(&actual); err != nil {
		t.Fatalf("query balance: %v", err)
	}
	if actual != expected {
		t.Fatalf("balance = %d, want %d", actual, expected)
	}
}

func assertTransactionEvents(t *testing.T, trace domain.ExecutionTrace, expected ...domain.TraceEventType) {
	t.Helper()
	if len(trace) != len(expected) {
		t.Fatalf("trace length = %d, want %d: %+v", len(trace), len(expected), trace)
	}
	for i, eventType := range expected {
		if trace[i].Type != eventType {
			t.Fatalf("trace[%d] = %s, want %s: %+v", i, trace[i].Type, eventType, trace)
		}
	}
}
