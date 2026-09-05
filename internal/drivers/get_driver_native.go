//go:build !js || !wasm

package drivers

import (
	"fmt"
	"strings"
)

// GetDriver returns a DatabaseDriver instance based on driver name and DSN for native environments.
func GetDriver(name string, dsn string) (DatabaseDriver, error) {
	switch strings.ToLower(name) {
	case "sqlite", "sqlite3", "":
		return NewSQLiteDriver(dsn), nil
	case "postgres", "postgresql":
		return NewPostgresDriver(dsn), nil
	case "mysql", "mariadb":
		return NewMySQLDriver(dsn), nil
	case "mock":
		return NewMockDriver(), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", name)
	}
}
