//go:build !js || !wasm

package drivers_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

func TestPostgresEffectiveIsolation(t *testing.T) {
	driver := drivers.NewPostgresDriver("unused")
	assertEffectiveIsolation(t, driver, "", domain.LevelReadCommitted)
	assertEffectiveIsolation(t, driver, domain.LevelReadCommitted, domain.LevelReadCommitted)
	assertEffectiveIsolation(t, driver, domain.LevelRepeatableRead, domain.LevelRepeatableRead)
	assertEffectiveIsolation(t, driver, domain.LevelSerializable, domain.LevelSerializable)
	assertUnsupportedIsolation(t, driver, domain.LevelReadUncommitted)
}

func TestMySQLEffectiveIsolation(t *testing.T) {
	driver := drivers.NewMySQLDriver("unused")
	assertEffectiveIsolation(t, driver, "", domain.LevelRepeatableRead)
	for _, level := range allIsolationLevels() {
		assertEffectiveIsolation(t, driver, level, level)
	}
}

func TestSQLiteEffectiveIsolation(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	assertEffectiveIsolation(t, driver, "", domain.LevelSerializable)
	assertEffectiveIsolation(t, driver, domain.LevelSerializable, domain.LevelSerializable)
	assertEffectiveIsolation(t, driver, domain.LevelReadUncommitted, domain.LevelReadUncommitted)
	assertUnsupportedIsolation(t, driver, domain.LevelReadCommitted)
	assertUnsupportedIsolation(t, driver, domain.LevelRepeatableRead)
}

func TestSQLiteAppliesIsolationOnTransactionConnection(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	t.Cleanup(func() { _ = driver.Close() })
	ctx := context.Background()
	if err := driver.Open(ctx); err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	for _, test := range []struct {
		level domain.IsolationLevel
		want  int
	}{
		{level: domain.LevelReadUncommitted, want: 1},
		{level: domain.LevelSerializable, want: 0},
	} {
		tx, err := driver.BeginTx(ctx, drivers.TransactionOptions{Isolation: test.level})
		if err != nil {
			t.Fatalf("begin %s: %v", test.level, err)
		}
		var actual int
		if err := tx.QueryRowContext(ctx, "PRAGMA read_uncommitted").Scan(&actual); err != nil {
			_ = tx.Rollback()
			t.Fatalf("query %s isolation: %v", test.level, err)
		}
		if err := tx.Rollback(); err != nil {
			t.Fatalf("rollback %s: %v", test.level, err)
		}
		if actual != test.want {
			t.Fatalf("PRAGMA read_uncommitted for %s = %d, want %d", test.level, actual, test.want)
		}
	}
}

func TestSQLiteReadUncommittedUsesConcurrentConnection(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	t.Cleanup(func() { _ = driver.Close() })
	ctx := context.Background()
	if err := driver.Reset(ctx,
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
		"INSERT INTO accounts VALUES (1, 10);",
	); err != nil {
		t.Fatalf("reset sqlite: %v", err)
	}

	writer, err := driver.BeginTx(ctx, drivers.TransactionOptions{Isolation: domain.LevelSerializable})
	if err != nil {
		t.Fatalf("begin writer: %v", err)
	}
	defer writer.Rollback()
	if _, err := writer.ExecContext(ctx, "UPDATE accounts SET balance = 9 WHERE id = 1"); err != nil {
		t.Fatalf("write uncommitted value: %v", err)
	}

	reader, err := driver.BeginTx(ctx, drivers.TransactionOptions{Isolation: domain.LevelReadUncommitted})
	if err != nil {
		t.Fatalf("begin read-uncommitted reader: %v", err)
	}
	defer reader.Rollback()
	var balance int
	if err := reader.QueryRowContext(ctx, "SELECT balance FROM accounts WHERE id = 1").Scan(&balance); err != nil {
		t.Fatalf("read uncommitted value: %v", err)
	}
	if balance != 9 {
		t.Fatalf("read-uncommitted balance = %d, want 9", balance)
	}
}

func TestMockEffectiveIsolation(t *testing.T) {
	driver := drivers.NewMockDriver()
	defer driver.Close()
	assertEffectiveIsolation(t, driver, "", domain.LevelSerializable)
	for _, level := range allIsolationLevels() {
		assertEffectiveIsolation(t, driver, level, level)
	}
}

type isolationResolver interface {
	EffectiveIsolation(domain.IsolationLevel) (domain.IsolationLevel, error)
}

func assertEffectiveIsolation(t *testing.T, resolver isolationResolver, requested, expected domain.IsolationLevel) {
	t.Helper()
	actual, err := resolver.EffectiveIsolation(requested)
	if err != nil {
		t.Fatalf("resolve isolation %q: %v", requested, err)
	}
	if actual != expected {
		t.Fatalf("isolation %q resolved to %q, want %q", requested, actual, expected)
	}
}

func assertUnsupportedIsolation(t *testing.T, resolver isolationResolver, requested domain.IsolationLevel) {
	t.Helper()
	if _, err := resolver.EffectiveIsolation(requested); err == nil {
		t.Fatalf("expected isolation %q to be rejected", requested)
	}
}

func allIsolationLevels() []domain.IsolationLevel {
	return []domain.IsolationLevel{
		domain.LevelReadUncommitted,
		domain.LevelReadCommitted,
		domain.LevelRepeatableRead,
		domain.LevelSerializable,
	}
}
