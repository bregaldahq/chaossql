//go:build js && wasm

package drivers

import (
	"fmt"
	"strings"
)

// GetDriver returns a DatabaseDriver instance for browser WebAssembly execution.
func GetDriver(name string, dsn string) (DatabaseDriver, error) {
	switch strings.ToLower(name) {
	case "sqlite", "sqlite3", "", "mock":
		return NewMockDriver(), nil
	default:
		return nil, fmt.Errorf("database driver '%s' is not supported in browser WASM; please select sqlite", name)
	}
}
