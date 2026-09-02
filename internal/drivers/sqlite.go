package drivers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

// SQLiteDriver implements DatabaseDriver using modernc.org/sqlite.
type SQLiteDriver struct {
	dsn string
	db  *sql.DB
	mu  sync.Mutex
}

// NewSQLiteDriver creates a new SQLite adapter.
func NewSQLiteDriver(dsn string) *SQLiteDriver {
	if dsn == "" {
		dsn = "file:chaos_mem?mode=memory&cache=shared"
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

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)

	// Configure WAL and busy timeout for concurrency
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		// Ignore in-memory WAL warnings if any
	}
	_, _ = db.ExecContext(ctx, "PRAGMA busy_timeout = 5000;")

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

func (d *SQLiteDriver) BeginTx(ctx context.Context) (Tx, error) {
	if d.db == nil {
		if err := d.Open(ctx); err != nil {
			return nil, err
		}
	}
	return d.db.BeginTx(ctx, nil)
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
