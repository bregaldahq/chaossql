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
