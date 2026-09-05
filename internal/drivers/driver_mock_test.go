package drivers_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/drivers"
)

func TestMockDriver_Lifecycle(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()
	defer driver.Close()

	if driver.DriverName() != "mock" {
		t.Errorf("expected driver name 'mock', got: %s", driver.DriverName())
	}

	if err := driver.Reset(ctx, "CREATE TABLE t (x INT);", "INSERT INTO t VALUES (1);"); err != nil {
		t.Fatalf("expected nil error on Reset, got: %v", err)
	}

	tx, err := driver.BeginTx(ctx)
	if err != nil {
		t.Fatalf("expected nil error on BeginTx, got: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("expected nil error on Commit, got: %v", err)
	}

	row := driver.QueryRow(ctx, "SELECT COUNT(*) FROM t;")
	var count int64
	if err := row.Scan(&count); err != nil {
		t.Fatalf("expected successful Scan, got: %v", err)
	}
}

func TestMockDriver_Exec(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()
	defer driver.Close()

	res, err := driver.Exec(ctx, "INSERT INTO accounts (id, balance) VALUES (1, 500);")
	if err != nil {
		t.Fatalf("expected nil error on Exec, got: %v", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		t.Fatalf("expected nil error on RowsAffected, got: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 row affected, got: %d", affected)
	}
}

func TestMockDriver_Query(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()
	defer driver.Close()

	rows, err := driver.Query(ctx, "SELECT id, balance FROM accounts;")
	if err != nil {
		t.Fatalf("expected nil error on Query, got: %v", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("expected nil error on Columns, got: %v", err)
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got: %d (%v)", len(cols), cols)
	}

	count := 0
	for rows.Next() {
		var id, balance int64
		if err := rows.Scan(&id, &balance); err != nil {
			t.Fatalf("expected nil error on Scan, got: %v", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("expected nil rows.Err, got: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row scanned, got: %d", count)
	}
}

func TestMockDriver_MultiColumnScan(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()
	defer driver.Close()

	var a, b int64
	row := driver.QueryRow(ctx, "SELECT a, b FROM t;")
	if err := row.Scan(&a, &b); err != nil {
		t.Fatalf("expected successful multi-column Scan, got: %v", err)
	}
}

func TestMockDriver_TxRollback(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()
	defer driver.Close()

	tx, err := driver.BeginTx(ctx)
	if err != nil {
		t.Fatalf("expected nil error on BeginTx, got: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("expected nil error on Rollback, got: %v", err)
	}
}

func TestMockDriver_IdempotentCloseAndReopen(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewMockDriver()

	// First close
	if err := driver.Close(); err != nil {
		t.Fatalf("expected nil error on first Close, got: %v", err)
	}
	// Second close must be idempotent (no panic, nil error)
	if err := driver.Close(); err != nil {
		t.Fatalf("expected nil error on second Close, got: %v", err)
	}

	// Re-open
	if err := driver.Open(ctx); err != nil {
		t.Fatalf("expected nil error on Open, got: %v", err)
	}
	defer driver.Close()

	// Must be usable after re-open
	var val int64
	if err := driver.QueryRow(ctx, "SELECT 1;").Scan(&val); err != nil {
		t.Fatalf("expected successful Scan after re-Open, got: %v", err)
	}
}
