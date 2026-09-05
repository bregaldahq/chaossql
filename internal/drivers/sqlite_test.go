//go:build !js || !wasm

package drivers_test

import (
	"context"
	"sync"
	"testing"

	"github.com/bregaldahq/chaossql/internal/drivers"
)

func TestSQLiteDriver_ResetAndConcurrency(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	schema := "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
	seed := "INSERT INTO accounts VALUES (1, 1000);"

	if err := driver.Reset(ctx, schema, seed); err != nil {
		t.Fatalf("failed to reset database: %v", err)
	}

	// Verify seed value
	var balance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1;").Scan(&balance); err != nil || balance != 1000 {
		t.Fatalf("expected balance 1000, got %d (err: %v)", balance, err)
	}

	// Test concurrent transactions
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := driver.BeginTx(ctx)
			if err != nil {
				return
			}
			_, _ = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - 10 WHERE id = 1;")
			_ = tx.Commit()
		}()
	}
	wg.Wait()
}
