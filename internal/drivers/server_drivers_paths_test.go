//go:build !js || !wasm

package drivers_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

// serverDrivers builds the server-backed drivers exercised by the shared lifecycle tests.
func serverDrivers() map[string]func() drivers.DatabaseDriver {
	return map[string]func() drivers.DatabaseDriver{
		"postgres": func() drivers.DatabaseDriver { return drivers.NewPostgresDriver(getPostgresDSN()) },
		"mysql":    func() drivers.DatabaseDriver { return drivers.NewMySQLDriver(getMySQLDSN()) },
	}
}

func TestServerDrivers_OpenFailureIsReported(t *testing.T) {
	cases := map[string]drivers.DatabaseDriver{
		"postgres": drivers.NewPostgresDriver("postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=2"),
		"mysql":    drivers.NewMySQLDriver("nobody@tcp(127.0.0.1:1)/none?timeout=2s"),
	}
	for name, d := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if err := d.Open(ctx); err == nil || !strings.Contains(err.Error(), "failed to ping") {
				t.Fatalf("Open err = %v, want a ping failure", err)
			}
			if _, err := d.BeginTx(ctx, drivers.TransactionOptions{}); err == nil {
				t.Fatal("BeginTx must surface the lazy Open failure")
			}
			if err := d.Reset(ctx, "", ""); err == nil {
				t.Fatal("Reset must surface the Open failure")
			}
			if err := d.Close(); err != nil {
				t.Fatalf("Close on never-opened driver = %v", err)
			}
		})
	}
}

func TestServerDrivers_LifecycleAndErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live database tests in short mode")
	}
	for name, newDriver := range serverDrivers() {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			d := newDriver()
			defer d.Close()

			// BeginTx opens the pool lazily.
			tx, err := d.BeginTx(ctx, drivers.TransactionOptions{})
			if err != nil {
				skipUnavailableDatabase(t, name, err)
			}
			if err := tx.Rollback(); err != nil {
				t.Fatal(err)
			}

			if err := d.Reset(ctx, "CREATE TABLE cov_items (id INT PRIMARY KEY, qty INT);", "INSERT INTO cov_items VALUES (1, 5);"); err != nil {
				t.Fatal(err)
			}
			// A second reset must drop the previous tables first.
			if err := d.Reset(ctx, "CREATE TABLE cov_items (id INT PRIMARY KEY, qty INT);", "INSERT INTO cov_items VALUES (1, 7);"); err != nil {
				t.Fatalf("second Reset: %v", err)
			}

			tx, err = d.BeginTx(ctx, drivers.TransactionOptions{Isolation: domain.LevelSerializable})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, "UPDATE cov_items SET qty = qty + 1 WHERE id = 1"); err != nil {
				t.Fatal(err)
			}
			var qty int
			if err := tx.QueryRowContext(ctx, "SELECT qty FROM cov_items WHERE id = 1").Scan(&qty); err != nil || qty != 8 {
				t.Fatalf("qty in tx = %d, err = %v; want 8", qty, err)
			}
			rows, err := tx.QueryContext(ctx, "SELECT id FROM cov_items")
			if err != nil {
				t.Fatal(err)
			}
			rows.Close()
			if _, err := tx.ExecContext(ctx, "UPDATE missing_table SET qty = 0"); err == nil {
				t.Fatal("tx exec against a missing table must fail")
			}
			_ = tx.Rollback()

			if _, err := d.Exec(ctx, "UPDATE missing_table SET qty = 0"); err == nil {
				t.Fatal("driver exec against a missing table must fail")
			}
			if _, err := d.Query(ctx, "SELECT * FROM missing_table"); err == nil {
				t.Fatal("driver query against a missing table must fail")
			}

			if err := d.Reset(ctx, "CREATE TABLE broken (", ""); err == nil || !strings.Contains(err.Error(), "schema SQL") {
				t.Fatalf("bad schema: err = %v", err)
			}
			if err := d.Reset(ctx, "", "INSERT INTO missing_table VALUES (1)"); err == nil || !strings.Contains(err.Error(), "seed SQL") {
				t.Fatalf("bad seed: err = %v", err)
			}

			if err := d.Close(); err != nil {
				t.Fatal(err)
			}
			if _, err := d.BeginTx(ctx, drivers.TransactionOptions{Isolation: "SNAPSHOT"}); err == nil {
				t.Fatal("unknown isolation must be rejected")
			}
		})
	}
}

func TestPostgresDriver_RejectsReadUncommitted(t *testing.T) {
	d := drivers.NewPostgresDriver("postgres://unused", drivers.PostgresOptions{IsolationLevel: sql.LevelReadCommitted})
	if _, err := d.EffectiveIsolation(domain.LevelReadUncommitted); err == nil {
		t.Fatal("postgres must reject READ UNCOMMITTED")
	}
	bad := drivers.NewPostgresDriver("postgres://unused", drivers.PostgresOptions{IsolationLevel: sql.LevelSnapshot})
	if _, err := bad.EffectiveIsolation(""); err == nil {
		t.Fatal("an unsupported configured default must be rejected")
	}
}

func TestPostgresDriver_ConnectionErrorsAreClassified(t *testing.T) {
	d := drivers.NewPostgresDriver(getPostgresDSN())
	ctx := context.Background()
	if err := d.Open(ctx); err != nil {
		skipUnavailableDatabase(t, "postgres", err)
	}
	d.Close()
	if err := d.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	tx, err := d.BeginTx(ctx, drivers.TransactionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); !errors.Is(err, drivers.ErrConnectionDropped) && !errors.Is(err, sql.ErrTxDone) {
		t.Fatalf("second commit err = %v, want a closed-transaction error", err)
	}
}
