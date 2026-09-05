package drivers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"sync"
)

var (
	mockDriverOnce sync.Once
)

type inmemDriver struct{}

func (d *inmemDriver) Open(name string) (driver.Conn, error) {
	return &inmemConn{}, nil
}

type inmemConn struct{}

func (c *inmemConn) Prepare(query string) (driver.Stmt, error) {
	return &inmemStmt{query: query}, nil
}

func (c *inmemConn) Close() error {
	return nil
}

func (c *inmemConn) Begin() (driver.Tx, error) {
	return &inmemTx{}, nil
}

type inmemTx struct{}

func (t *inmemTx) Commit() error {
	return nil
}

func (t *inmemTx) Rollback() error {
	return nil
}

type inmemStmt struct {
	query string
}

func (s *inmemStmt) Close() error {
	return nil
}

func (s *inmemStmt) NumInput() int {
	return -1
}

func (s *inmemStmt) Exec(args []driver.Value) (driver.Result, error) {
	return driver.RowsAffected(1), nil
}

func (s *inmemStmt) Query(args []driver.Value) (driver.Rows, error) {
	return &inmemRows{cols: []string{"val"}, pos: 0}, nil
}

type inmemRows struct {
	cols []string
	pos  int
}

func (r *inmemRows) Columns() []string {
	return r.cols
}

func (r *inmemRows) Close() error {
	return nil
}

func (r *inmemRows) Next(dest []driver.Value) error {
	if r.pos > 0 {
		return io.EOF
	}
	r.pos++
	dest[0] = int64(0)
	return nil
}

// MockDriver implements DatabaseDriver purely using standard library database/sql/driver.
type MockDriver struct {
	db *sql.DB
	mu sync.Mutex
}

// NewMockDriver returns a thread-safe in-memory mock driver.
func NewMockDriver() *MockDriver {
	mockDriverOnce.Do(func() {
		sql.Register("chaossql_mock", &inmemDriver{})
	})
	db, err := sql.Open("chaossql_mock", "")
	if err != nil {
		panic(fmt.Sprintf("failed to open chaossql_mock: %v", err))
	}
	return &MockDriver{db: db}
}

func (m *MockDriver) DriverName() string {
	return "mock"
}

func (m *MockDriver) Open(ctx context.Context) error {
	return nil
}

func (m *MockDriver) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *MockDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	return nil
}

func (m *MockDriver) BeginTx(ctx context.Context) (Tx, error) {
	return m.db.BeginTx(ctx, nil)
}

func (m *MockDriver) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return m.db.QueryRowContext(ctx, query, args...)
}

func (m *MockDriver) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return m.db.QueryContext(ctx, query, args...)
}

func (m *MockDriver) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return m.db.ExecContext(ctx, query, args...)
}
