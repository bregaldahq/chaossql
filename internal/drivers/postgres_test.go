//go:build !js || !wasm

package drivers_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func getPostgresDSN() string {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/chaossql_test?sslmode=disable"
	}
	return dsn
}

func TestPostgresDriver_ResetAndConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres tests in short mode")
	}

	ctx := context.Background()
	dsn := getPostgresDSN()

	opts := drivers.PostgresOptions{IsolationLevel: sql.LevelRepeatableRead}
	driver := drivers.NewPostgresDriver(dsn, opts)
	defer driver.Close()

	schema := "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
	seed := "INSERT INTO accounts VALUES (1, 1000);"

	if err := driver.Reset(ctx, schema, seed); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "dial error") ||
			strings.Contains(errStr, "failed to connect") ||
			strings.Contains(errStr, "authentication failed") ||
			strings.Contains(errStr, "SASL auth") {
			skipUnavailableDatabase(t, "postgres", err)
		}
		t.Fatalf("failed to reset database: %v", err)
	}

	var balance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1;").Scan(&balance); err != nil || balance != 1000 {
		t.Fatalf("expected balance 1000, got %d (err: %v)", balance, err)
	}

	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := driver.BeginTx(ctx, drivers.TransactionOptions{})
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - 10 WHERE id = 1;")
			if err != nil {
				tx.Rollback()
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			err = tx.Commit()
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	var finalBalance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1;").Scan(&finalBalance); err != nil {
		t.Fatalf("failed to get final balance: %v", err)
	}

	successCount := 5 - len(errs)
	expectedBalance := 1000 - 10*successCount
	if finalBalance != expectedBalance {
		t.Errorf("expected balance %d, got %d", expectedBalance, finalBalance)
	}

	for _, err := range errs {
		if !errors.Is(err, drivers.ErrSerializationFailure) && !errors.Is(err, drivers.ErrDeadlockDetected) && !errors.Is(err, drivers.ErrConnectionDropped) {
			t.Logf("transaction failed with error: %v", err)
		}
	}
}

func TestPostgresDriver_TransactionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping postgres live integration tests in short mode")
	}

	ctx := context.Background()
	driver := drivers.NewPostgresDriver(getPostgresDSN())
	t.Cleanup(func() { _ = driver.Close() })
	if err := driver.Open(ctx); err != nil {
		skipUnavailableDatabase(t, "postgres", err)
	}

	tx, err := driver.BeginTx(ctx, drivers.TransactionOptions{Isolation: domain.LevelRepeatableRead})
	if err != nil {
		t.Fatalf("begin repeatable read transaction: %v", err)
	}
	var actualIsolation string
	if err := tx.QueryRowContext(ctx, "SHOW transaction_isolation").Scan(&actualIsolation); err != nil {
		_ = tx.Rollback()
		t.Fatalf("query transaction isolation: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback isolation probe: %v", err)
	}
	if actualIsolation != "repeatable read" {
		t.Fatalf("transaction isolation = %q, want repeatable read", actualIsolation)
	}

	assertLiveRunnerTransactionSemantics(t, driver, domain.LevelRepeatableRead)
}

func assertLiveRunnerTransactionSemantics(t *testing.T, driver drivers.DatabaseDriver, isolation domain.IsolationLevel) {
	t.Helper()
	ctx := context.Background()
	spec := domain.Spec{
		Database: domain.DatabaseConfig{
			Isolation: isolation,
			Schema:    "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);",
			Seed:      "INSERT INTO accounts VALUES (1, 10);",
		},
		Engine: domain.EngineConfig{Workers: 1},
		Invariants: []domain.InvariantConfig{{
			Name: "committed_balance", Query: "SELECT balance FROM accounts WHERE id = 1", Assert: "balance == 9",
		}},
	}

	result, err := engine.NewRunner(driver, 1).RunSchedule(ctx, spec, []domain.ScheduledOp{{
		ID: 1, Name: "commit", Steps: []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"}},
	}})
	if err != nil {
		t.Fatalf("run committed operation: %v", err)
	}
	if result.Status != domain.StatusPassed || result.Isolation != isolation {
		t.Fatalf("unexpected committed result: %+v", result)
	}

	spec.Invariants[0].Assert = "balance == 10"
	result, err = engine.NewRunner(driver, 1).RunSchedule(ctx, spec, []domain.ScheduledOp{{
		ID: 2, Name: "rollback", Steps: []domain.StepConfig{
			{SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"},
			{SQL: "NOT VALID SQL"},
		},
	}})
	if err != nil {
		t.Fatalf("run failed operation: %v", err)
	}
	if result.Status != domain.StatusExecutionError || len(result.OperationErrors) != 1 {
		t.Fatalf("unexpected rollback result: %+v", result)
	}
	var balance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1").Scan(&balance); err != nil {
		t.Fatalf("query rolled back balance: %v", err)
	}
	if balance != 10 {
		t.Fatalf("balance after rollback = %d, want 10", balance)
	}
}
