//go:build !js || !wasm

package drivers

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestExtractColumns(t *testing.T) {
	cases := map[string][]string{
		"UPDATE t SET v = 1": {"val"},
		"SELECT":             {"val"},
		"SELECT FROM t":      {"val"},
		"SELECT ,, FROM t":   {"val"},
		"SELECT balance AS total, COUNT(*) FROM accounts":    {"total", "COUNT(*)"},
		"select coalesce(sum(a), 0) as s, b from (select 1)": {"s", "b"},
		"SELECT 1;": {"1"},
	}
	for query, want := range cases {
		if got := extractColumns(query); !reflect.DeepEqual(got, want) {
			t.Errorf("extractColumns(%q) = %q, want %q", query, got, want)
		}
	}
}

func TestMockDriverClosedAndTransactionStats(t *testing.T) {
	ctx := context.Background()
	m := NewMockDriver()
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := m.BeginTx(ctx, TransactionOptions{}); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("BeginTx on closed driver: err = %v", err)
	}
	if _, err := m.Query(ctx, "SELECT 1"); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("Query on closed driver: err = %v", err)
	}
	if _, err := m.Exec(ctx, "UPDATE t SET v = 1"); !errors.Is(err, sql.ErrConnDone) {
		t.Fatalf("Exec on closed driver: err = %v", err)
	}
	if err := m.QueryRow(ctx, "SELECT 1").Scan(new(int64)); err == nil {
		t.Fatal("QueryRow on closed driver must fail on Scan")
	}

	if err := m.Open(ctx); err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if _, err := m.BeginTx(ctx, TransactionOptions{Isolation: "SNAPSHOT"}); err == nil {
		t.Fatal("unsupported isolation must be rejected")
	}
	for _, commit := range []bool{true, false, false} {
		tx, err := m.BeginTx(ctx, TransactionOptions{Isolation: domain.LevelReadCommitted})
		if err != nil {
			t.Fatal(err)
		}
		if commit {
			err = tx.Commit()
		} else {
			err = tx.Rollback()
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if got, want := m.TransactionStats(), (MockTransactionStats{Opened: 3, Committed: 1, RolledBack: 2}); got != want {
		t.Fatalf("TransactionStats = %+v, want %+v", got, want)
	}
	m.ResetTransactionStats()
	if got := m.TransactionStats(); got != (MockTransactionStats{}) {
		t.Fatalf("after reset TransactionStats = %+v", got)
	}
}

func TestSQLiteDriverLifecycle(t *testing.T) {
	ctx := context.Background()
	d := NewSQLiteDriver("")
	defer d.Close()

	// BeginTx opens the database lazily.
	tx, err := d.BeginTx(ctx, TransactionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := d.Open(ctx); err != nil {
		t.Fatalf("second Open must be a no-op, got %v", err)
	}

	if err := d.Reset(ctx, "CREATE TABLE a (id INT); CREATE TABLE b (id INT);", "INSERT INTO a VALUES (1);"); err != nil {
		t.Fatal(err)
	}
	// A second Reset drops the previous tables before recreating the schema.
	if err := d.Reset(ctx, "CREATE TABLE a (id INT, v INT);", "INSERT INTO a VALUES (1, 2);"); err != nil {
		t.Fatal(err)
	}
	rows, err := d.Query(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' ORDER BY name")
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	rows.Close()
	if !reflect.DeepEqual(tables, []string{"a"}) {
		t.Fatalf("tables after reset = %v, want [a]", tables)
	}

	result, err := d.Exec(ctx, "UPDATE a SET v = 3 WHERE id = 1")
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := result.RowsAffected(); n != 1 {
		t.Fatalf("RowsAffected = %d, want 1", n)
	}

	if err := d.Reset(ctx, "CREATE TABLE broken (", ""); err == nil || !strings.Contains(err.Error(), "schema SQL") {
		t.Fatalf("bad schema: err = %v", err)
	}
	if err := d.Reset(ctx, "CREATE TABLE c (id INT);", "INSERT INTO missing VALUES (1);"); err == nil || !strings.Contains(err.Error(), "seed SQL") {
		t.Fatalf("bad seed: err = %v", err)
	}
	if _, err := d.BeginTx(ctx, TransactionOptions{Isolation: domain.LevelRepeatableRead}); err == nil {
		t.Fatal("sqlite must reject REPEATABLE READ")
	}
}

func TestSQLIsolationConversionsRoundTrip(t *testing.T) {
	for _, level := range []domain.IsolationLevel{domain.LevelReadUncommitted, domain.LevelReadCommitted, domain.LevelRepeatableRead, domain.LevelSerializable} {
		sqlLevel, err := toSQLIsolation(level)
		if err != nil {
			t.Fatal(err)
		}
		back, err := fromSQLIsolation(sqlLevel, "")
		if err != nil || back != level {
			t.Fatalf("round trip %q -> %v -> %q (err %v)", level, sqlLevel, back, err)
		}
	}
	if _, err := toSQLIsolation("SNAPSHOT"); err == nil {
		t.Fatal("unknown level must be rejected")
	}
	if got, err := fromSQLIsolation(sql.LevelDefault, domain.LevelReadCommitted); err != nil || got != domain.LevelReadCommitted {
		t.Fatalf("default level = %q, %v; want the fallback", got, err)
	}
	if _, err := fromSQLIsolation(sql.LevelSnapshot, ""); err == nil {
		t.Fatal("snapshot has no domain equivalent and must be rejected")
	}
}
