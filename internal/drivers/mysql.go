//go:build !js || !wasm

package drivers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/go-sql-driver/mysql"
)

// MySQLDriver implements DatabaseDriver using github.com/go-sql-driver/mysql.
type MySQLDriver struct {
	dsn            string
	db             *sql.DB
	isolationLevel sql.IsolationLevel
	mu             sync.Mutex
}

type MySQLOptions struct {
	IsolationLevel sql.IsolationLevel
}

// NewMySQLDriver creates a new MySQL adapter.
func NewMySQLDriver(dsn string, opts ...MySQLOptions) *MySQLDriver {
	level := sql.LevelRepeatableRead
	if len(opts) > 0 && opts[0].IsolationLevel != 0 {
		level = opts[0].IsolationLevel
	}
	return &MySQLDriver{
		dsn:            dsn,
		isolationLevel: level,
	}
}

func (d *MySQLDriver) DriverName() string {
	return "mysql"
}

func (d *MySQLDriver) Open(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		return nil
	}

	dsn := d.dsn
	if cfg, err := mysql.ParseDSN(dsn); err == nil {
		cfg.MultiStatements = true
		dsn = cfg.FormatDSN()
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open mysql db: %w", err)
	}

	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(50)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to ping mysql db: %w", err)
	}

	d.db = db
	return nil
}

func (d *MySQLDriver) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db != nil {
		err := d.db.Close()
		d.db = nil
		return err
	}
	return nil
}

func (d *MySQLDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	if err := d.Open(ctx); err != nil {
		return err
	}

	// Drop existing tables in current database
	rows, err := d.db.QueryContext(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE';")
	if err == nil {
		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err == nil {
				tables = append(tables, tableName)
			}
		}
		_ = rows.Close()

		if len(tables) > 0 {
			if _, err := d.db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
				return fmt.Errorf("failed to disable foreign key checks: %w", err)
			}
			for _, t := range tables {
				if _, err := d.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", t)); err != nil {
					return fmt.Errorf("failed to drop table %s: %w", t, err)
				}
			}
			if _, err := d.db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
				return fmt.Errorf("failed to enable foreign key checks: %w", err)
			}
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

// mysqlTxWrapper wraps an sql.Tx to intercept errors.
type mysqlTxWrapper struct {
	tx *sql.Tx
	d  *MySQLDriver
}

func (t *mysqlTxWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := t.tx.ExecContext(ctx, query, args...)
	return res, t.d.HandleError(err)
}

func (t *mysqlTxWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	return rows, t.d.HandleError(err)
}

func (t *mysqlTxWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *mysqlTxWrapper) Commit() error {
	return t.d.HandleError(t.tx.Commit())
}

func (t *mysqlTxWrapper) Rollback() error {
	return t.d.HandleError(t.tx.Rollback())
}

func (d *MySQLDriver) BeginTx(ctx context.Context) (Tx, error) {
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
		return nil, d.HandleError(err)
	}
	return &mysqlTxWrapper{tx: tx, d: d}, nil
}

func (d *MySQLDriver) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *MySQLDriver) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	return rows, d.HandleError(err)
}

func (d *MySQLDriver) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	res, err := d.db.ExecContext(ctx, query, args...)
	return res, d.HandleError(err)
}

func (d *MySQLDriver) HandleError(err error) error {
	if err == nil {
		return nil
	}

	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		switch myErr.Number {
		case 1213: // ER_LOCK_DEADLOCK
			return fmt.Errorf("%w: %v", domain.ErrDeadlockDetected, err)
		case 1205: // ER_LOCK_WAIT_TIMEOUT
			return fmt.Errorf("%w: %v", domain.ErrTimeout, err)
		}
	}

	// Check for connection dropped errors
	if errors.Is(err, sql.ErrConnDone) ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer") ||
		strings.Contains(err.Error(), "invalid connection") ||
		strings.Contains(err.Error(), "bad connection") ||
		strings.Contains(err.Error(), "driver: bad connection") ||
		strings.Contains(err.Error(), "closed") {
		return fmt.Errorf("%w: %v", ErrConnectionDropped, err)
	}

	return err
}
