//go:build !js || !wasm

package drivers_test

import (
	"os"
	"testing"
)

func skipUnavailableDatabase(t *testing.T, database string, err error) {
	t.Helper()
	if os.Getenv("CHAOSSQL_REQUIRE_DATABASES") == "1" {
		t.Fatalf("required %s integration database is unavailable: %v", database, err)
	}
	t.Skipf("%s not available or authentication failed: %v", database, err)
}
