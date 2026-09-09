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

	"github.com/bregaldahq/chaossql/internal/drivers"
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
			t.Skipf("postgres not available or auth failed: %v", err)
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
			tx, err := driver.BeginTx(ctx)
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
