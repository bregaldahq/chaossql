//go:build !js || !wasm

package drivers_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/go-sql-driver/mysql"
)

func getMySQLDSN() string {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/chaossql_test"
	}
	return dsn
}

func TestMySQLDriver_Registration(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		dsn        string
		wantErr    bool
		wantDriver string
	}{
		{
			name:       "mysql standard name",
			driverName: "mysql",
			dsn:        "root:pass@tcp(localhost:3306)/db",
			wantErr:    false,
			wantDriver: "mysql",
		},
		{
			name:       "mariadb alias",
			driverName: "mariadb",
			dsn:        "root:pass@tcp(localhost:3306)/db",
			wantErr:    false,
			wantDriver: "mysql",
		},
		{
			name:       "uppercase name",
			driverName: "MYSQL",
			dsn:        "root:pass@tcp(localhost:3306)/db",
			wantErr:    false,
			wantDriver: "mysql",
		},
		{
			name:       "sqlite standard name",
			driverName: "sqlite",
			dsn:        ":memory:",
			wantErr:    false,
			wantDriver: "sqlite",
		},
		{
			name:       "postgres standard name",
			driverName: "postgres",
			dsn:        "postgres://user:pass@localhost:5432/db",
			wantErr:    false,
			wantDriver: "postgres",
		},
		{
			name:       "unsupported driver",
			driverName: "oracle",
			dsn:        "oracle://...",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := drivers.GetDriver(tt.driverName, tt.dsn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetDriver(%q, %q) error = %v, wantErr %v", tt.driverName, tt.dsn, err, tt.wantErr)
			}
			if !tt.wantErr {
				if d == nil {
					t.Fatalf("GetDriver(%q, %q) returned nil driver", tt.driverName, tt.dsn)
				}
				if d.DriverName() != tt.wantDriver {
					t.Errorf("DriverName() = %q, want %q", d.DriverName(), tt.wantDriver)
				}
			}
		})
	}
}

func TestMySQLDriver_ErrorMapping(t *testing.T) {
	driver := drivers.NewMySQLDriver("root:pass@tcp(localhost:3306)/db")

	tests := []struct {
		name      string
		inputErr  error
		targetErr error
		wantNil   bool
	}{
		{
			name:      "nil error",
			inputErr:  nil,
			targetErr: nil,
			wantNil:   true,
		},
		{
			name: "deadlock error 1213",
			inputErr: &mysql.MySQLError{
				Number:  1213,
				Message: "Deadlock found when trying to get lock; try restarting transaction",
			},
			targetErr: domain.ErrDeadlockDetected,
		},
		{
			name: "lock wait timeout error 1205",
			inputErr: &mysql.MySQLError{
				Number:  1205,
				Message: "Lock wait timeout exceeded; try restarting transaction",
			},
			targetErr: domain.ErrTimeout,
		},
		{
			name:      "sql.ErrConnDone connection dropped",
			inputErr:  sql.ErrConnDone,
			targetErr: drivers.ErrConnectionDropped,
		},
		{
			name:      "connection refused error",
			inputErr:  errors.New("dial tcp 127.0.0.1:3306: connect: connection refused"),
			targetErr: drivers.ErrConnectionDropped,
		},
		{
			name:      "bad connection error",
			inputErr:  errors.New("driver: bad connection"),
			targetErr: drivers.ErrConnectionDropped,
		},
		{
			name: "syntax error code 1064 unmodified",
			inputErr: &mysql.MySQLError{
				Number:  1064,
				Message: "You have an error in your SQL syntax",
			},
			targetErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapped := driver.HandleError(tt.inputErr)
			if tt.wantNil {
				if mapped != nil {
					t.Fatalf("HandleError(nil) = %v, want nil", mapped)
				}
				return
			}
			if tt.targetErr != nil {
				if !errors.Is(mapped, tt.targetErr) {
					t.Errorf("HandleError(%v) = %v; want errors.Is(..., %v) to be true", tt.inputErr, mapped, tt.targetErr)
				}
			} else {
				if mapped != tt.inputErr {
					t.Errorf("HandleError(%v) = %v, want %v", tt.inputErr, mapped, tt.inputErr)
				}
			}
		})
	}
}

func TestMySQLDriver_ResetAndConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping mysql live integration tests in short mode")
	}

	ctx := context.Background()
	dsn := getMySQLDSN()

	opts := drivers.MySQLOptions{IsolationLevel: sql.LevelRepeatableRead}
	driver := drivers.NewMySQLDriver(dsn, opts)
	defer driver.Close()

	schema := "CREATE TABLE IF NOT EXISTS accounts (id INT PRIMARY KEY, balance INT);"
	seed := "INSERT INTO accounts (id, balance) VALUES (1, 1000);"

	if err := driver.Reset(ctx, schema, seed); err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "dial error") ||
			strings.Contains(errStr, "failed to connect") ||
			strings.Contains(errStr, "Access denied") ||
			strings.Contains(errStr, "invalid connection") ||
			strings.Contains(errStr, "i/o timeout") ||
			strings.Contains(errStr, "Unknown database") ||
			strings.Contains(errStr, "getsockopt") {
			t.Skipf("mysql not available or auth failed: %v", err)
		}
		t.Fatalf("failed to reset database: %v", err)
	}

	var balance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1;").Scan(&balance); err != nil || balance != 1000 {
		t.Fatalf("expected balance 1000, got %d (err: %v)", balance, err)
	}

	var wg sync.WaitGroup
	var errs []error
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := driver.BeginTx(ctx)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - 10 WHERE id = 1;")
			if err != nil {
				_ = tx.Rollback()
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			err = tx.Commit()
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	var finalBalance int
	if err := driver.QueryRow(ctx, "SELECT balance FROM accounts WHERE id = 1;").Scan(&finalBalance); err != nil {
		t.Fatalf("failed to get final balance: %v", err)
	}

	successCount := 5 - len(errs)
	expectedBalance := 1000 - 10*successCount
	if finalBalance != expectedBalance {
		t.Errorf("expected balance %d, got %d", expectedBalance, finalBalance)
	}

	for _, err := range errs {
		if !errors.Is(err, domain.ErrDeadlockDetected) && !errors.Is(err, domain.ErrTimeout) && !errors.Is(err, drivers.ErrConnectionDropped) {
			t.Logf("transaction failed with error: %v", err)
		}
	}
}
