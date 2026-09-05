package drivers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode"
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
	cols := extractColumns(s.query)
	return &inmemRows{cols: cols, pos: 0}, nil
}

func extractColumns(query string) []string {
	upper := strings.ToUpper(query)
	selIdx := strings.Index(upper, "SELECT")
	if selIdx == -1 {
		return []string{"val"}
	}

	afterSelect := strings.TrimSpace(query[selIdx+len("SELECT"):])
	if afterSelect == "" {
		return []string{"val"}
	}

	// Find top-level FROM (parenDepth == 0)
	parenDepth := 0
	fromIdx := -1
	upperAfter := strings.ToUpper(afterSelect)
	for i := 0; i < len(upperAfter); i++ {
		switch upperAfter[i] {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		default:
			if parenDepth == 0 && strings.HasPrefix(upperAfter[i:], "FROM") {
				beforeOk := (i == 0 || unicode.IsSpace(rune(upperAfter[i-1])))
				afterEnd := i + 4
				afterOk := (afterEnd < len(upperAfter) && unicode.IsSpace(rune(upperAfter[afterEnd])))
				if beforeOk && afterOk {
					fromIdx = i
					break
				}
			}
		}
		if fromIdx != -1 {
			break
		}
	}

	var colStr string
	if fromIdx != -1 {
		colStr = afterSelect[:fromIdx]
	} else {
		colStr = afterSelect
	}

	colStr = strings.TrimRight(strings.TrimSpace(colStr), ";")
	if colStr == "" {
		return []string{"val"}
	}

	var cols []string
	parenDepth = 0
	start := 0
	for i, r := range colStr {
		switch r {
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case ',':
			if parenDepth == 0 {
				col := cleanColumn(colStr[start:i])
				if col != "" {
					cols = append(cols, col)
				}
				start = i + 1
			}
		}
	}
	if start < len(colStr) {
		col := cleanColumn(colStr[start:])
		if col != "" {
			cols = append(cols, col)
		}
	}

	if len(cols) == 0 {
		return []string{"val"}
	}
	return cols
}

func cleanColumn(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return ""
	}
	upper := strings.ToUpper(expr)
	if asIdx := strings.LastIndex(upper, " AS "); asIdx != -1 {
		alias := strings.TrimSpace(expr[asIdx+4:])
		alias = strings.Trim(alias, "`\"'")
		if alias != "" {
			return alias
		}
	}
	return strings.Trim(expr, "`\"'")
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
	for i := range dest {
		dest[i] = int64(0)
	}
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
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.db == nil {
		mockDriverOnce.Do(func() {
			sql.Register("chaossql_mock", &inmemDriver{})
		})
		db, err := sql.Open("chaossql_mock", "")
		if err != nil {
			return fmt.Errorf("failed to open chaossql_mock: %w", err)
		}
		m.db = db
	}
	return nil
}

func (m *MockDriver) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.db != nil {
		err := m.db.Close()
		m.db = nil
		return err
	}
	return nil
}

func (m *MockDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	return nil
}

func (m *MockDriver) BeginTx(ctx context.Context) (Tx, error) {
	m.mu.Lock()
	db := m.db
	m.mu.Unlock()
	if db == nil {
		return nil, sql.ErrConnDone
	}
	return db.BeginTx(ctx, nil)
}

func (m *MockDriver) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	m.mu.Lock()
	db := m.db
	m.mu.Unlock()
	if db == nil {
		closedDB, _ := sql.Open("chaossql_mock", "")
		closedDB.Close()
		return closedDB.QueryRowContext(ctx, query, args...)
	}
	return db.QueryRowContext(ctx, query, args...)
}

func (m *MockDriver) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	m.mu.Lock()
	db := m.db
	m.mu.Unlock()
	if db == nil {
		return nil, sql.ErrConnDone
	}
	return db.QueryContext(ctx, query, args...)
}

func (m *MockDriver) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	m.mu.Lock()
	db := m.db
	m.mu.Unlock()
	if db == nil {
		return nil, sql.ErrConnDone
	}
	return db.ExecContext(ctx, query, args...)
}

