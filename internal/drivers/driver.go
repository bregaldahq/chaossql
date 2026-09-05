package drivers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bregaldahq/chaossql/internal/domain"
)

var (
	ErrSerializationFailure = domain.ErrSerializationFailure
	ErrDeadlockDetected     = domain.ErrDeadlockDetected
	ErrTimeout              = domain.ErrTimeout
	ErrConnectionDropped    = errors.New("connection dropped")
)

// Tx represents an isolated database transaction.
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Commit() error
	Rollback() error
}

// DatabaseDriver defines the port for database operations.
type DatabaseDriver interface {
	DriverName() string
	Open(ctx context.Context) error
	Close() error
	Reset(ctx context.Context, schemaSQL, seedSQL string) error
	BeginTx(ctx context.Context) (Tx, error)
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}
