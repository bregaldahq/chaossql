package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresDriver implements DatabaseDriver using pgx.
type PostgresDriver struct {
	dsn            string
	db             *sql.DB
	isolationLevel sql.IsolationLevel
	mu             sync.Mutex
}

type PostgresOptions struct {
	IsolationLevel sql.IsolationLevel
}

// NewPostgresDriver creates a new Postgres adapter.
func NewPostgresDriver(dsn string, opts ...PostgresOptions) *PostgresDriver {
	level := sql.LevelReadCommitted
	if len(opts) > 0 {
		level = opts[0].IsolationLevel
	}
	return &PostgresDriver{
		dsn:            dsn,
		isolationLevel: level,
	}
}

func (d *PostgresDriver) DriverName() string {
	return "postgres"
}

func (d *PostgresDriver) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		return nil
	}

	db, err := sql.Open("pgx", d.dsn)
	if err != nil {
		return fmt.Errorf("failed to open postgres db: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping postgres db: %w", err)
	}

	d.db = db
	return nil
}

func (d *PostgresDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		err := d.db.Close()
		d.db = nil
		return err
	}
	return nil
}

func (d *PostgresDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	if err := d.Open(ctx); err != nil {
		return err
	}

	// Drop the public schema and recreate it for a clean slate
	if _, err := d.db.ExecContext(ctx, "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"); err != nil {
		return fmt.Errorf("failed to reset public schema: %w", err)
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

// txWrapper wraps an sql.Tx to intercept errors.
type txWrapper struct {
	tx *sql.Tx
	d  *PostgresDriver
}

func (t *txWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	return res, t.d.handleError(err)
}

func (t *txWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	return rows, t.d.handleError(err)
}

func (t *txWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *txWrapper) Commit() error {
	return t.d.handleError(t.tx.Commit())
}

func (t *txWrapper) Rollback() error {
	return t.d.handleError(t.tx.Rollback())
}

func (d *PostgresDriver) BeginTx(ctx context.Context) (Tx, error) {
	if d.db == nil {
		if err := d.Open(ctx); err != nil {
			return nil, err
		}
	}

	opts := &sql.TxOptions{
		Isolation: d.isolationLevel,
	}
	tx, err := d.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, d.handleError(err)
	}
	return &txWrapper{tx: tx, d: d}, nil
}

func (d *PostgresDriver) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *PostgresDriver) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	return rows, d.handleError(err)
}

func (d *PostgresDriver) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := d.db.ExecContext(ctx, query, args...)
	return res, d.handleError(err)
}

func (d *PostgresDriver) handleError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "40P01":
			return fmt.Errorf("%w: %v", ErrDeadlockDetected, err)
		case "40001":
			return fmt.Errorf("%w: %v", ErrSerializationFailure, err)
		}
	}

	// Check for connection dropped errors
	if errors.Is(err, sql.ErrConnDone) || strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "broken pipe") || strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "closed") {
		return fmt.Errorf("%w: %v", ErrConnectionDropped, err)
	}

	return err
}
