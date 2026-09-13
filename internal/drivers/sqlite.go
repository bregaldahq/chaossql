//go:build !js || !wasm

package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	_ "modernc.org/sqlite"
)

var memDBSeq atomic.Uint64

// SQLiteDriver implements DatabaseDriver using modernc.org/sqlite.
type SQLiteDriver struct {
	dsn string
	db  *sql.DB
	mu  sync.Mutex
}

// NewSQLiteDriver creates a new SQLite adapter.
func NewSQLiteDriver(dsn string) *SQLiteDriver {
	if dsn == "" || dsn == ":memory:" {
		seq := memDBSeq.Add(1)
		dsn = fmt.Sprintf("file:chaos_mem_%d_%d?mode=memory&cache=shared", time.Now().UnixNano(), seq)
	}
	return &SQLiteDriver{dsn: dsn}
}

func (d *SQLiteDriver) DriverName() string {
	return "sqlite"
}

func (d *SQLiteDriver) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		return nil
	}

	db, err := sql.Open("sqlite", d.dsn)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Configure WAL and busy timeout for concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		// Ignore in-memory WAL warnings if any
	}
	_, _ = db.ExecContext(ctx, "PRAGMA busy_timeout = 250;")

	d.db = db
	return nil
}

func (d *SQLiteDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		err := d.db.Close()
		d.db = nil
		return err
	}
	return nil
}

func (d *SQLiteDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	if err := d.Open(ctx); err != nil {
		return err
	}

	// Drop existing tables for clean state
	rows, err := d.db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err == nil {
		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err == nil {
				tables = append(tables, tableName)
			}
		}
		rows.Close()

		for _, t := range tables {
			_, _ = d.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s;", t))
		}
	}

	if schemaSQL != "" {
		if _, err := d.db.ExecContext(ctx, schemaSQL); err != nil {
			return fmt.Errorf("failed to execute schema SQL: %w", err)
		}
	}

	if seedSQL != "" {
		if _, err := d.db.ExecContext(ctx, seedSQL); err != nil {
			return fmt.Errorf("failed to execute seed SQL: %w", err)
		}
	}

	return nil
}

func (d *SQLiteDriver) EffectiveIsolation(requested domain.IsolationLevel) (domain.IsolationLevel, error) {
	return resolveIsolation("sqlite", requested, domain.LevelSerializable, domain.LevelSerializable, domain.LevelReadUncommitted)
}

func (d *SQLiteDriver) BeginTx(ctx context.Context, opts TransactionOptions) (Tx, error) {
	if d.db == nil {
		if err := d.Open(ctx); err != nil {
			return nil, err
		}
	}
	effective, err := d.EffectiveIsolation(opts.Isolation)
	if err != nil {
		return nil, err
	}
	readUncommitted := 0
	if effective == domain.LevelReadUncommitted {
		readUncommitted = 1
		d.db.SetMaxOpenConns(50)
		d.db.SetMaxIdleConns(50)
	}
	conn, err := d.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := conn.ExecContext(ctx, fmt.Sprintf("PRAGMA read_uncommitted = %d", readUncommitted)); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to configure sqlite isolation %q: %w", effective, err)
	}
	_, _ = conn.ExecContext(ctx, "PRAGMA busy_timeout = 250")
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &sqliteTx{Tx: tx, conn: conn}, nil
}

type sqliteTx struct {
	*sql.Tx
	conn *sql.Conn
}

func (tx *sqliteTx) Commit() error {
	commitErr := tx.Tx.Commit()
	cleanupErr := tx.resetIsolation()
	if commitErr != nil {
		return commitErr
	}
	return cleanupErr
}

func (tx *sqliteTx) Rollback() error {
	rollbackErr := tx.Tx.Rollback()
	cleanupErr := tx.resetIsolation()
	if rollbackErr != nil {
		return rollbackErr
	}
	return cleanupErr
}

func (tx *sqliteTx) resetIsolation() error {
	_, resetErr := tx.conn.ExecContext(context.Background(), "PRAGMA read_uncommitted = 0")
	closeErr := tx.conn.Close()
	if resetErr != nil {
		return fmt.Errorf("failed to reset sqlite isolation: %w", resetErr)
	}
	return closeErr
}

func (d *SQLiteDriver) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *SQLiteDriver) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *SQLiteDriver) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}
