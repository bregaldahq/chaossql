package drivers_test

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/drivers"
)

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "postgres://postgres:supersecret@localhost:5432/chaossql?sslmode=disable",
			expected: "postgres://postgres:******@localhost:5432/chaossql?sslmode=disable",
		},
		{
			input:    "root:mySecretP@ss@tcp(127.0.0.1:3306)/dbname",
			expected: "root:******@tcp(127.0.0.1:3306)/dbname",
		},
		{
			input:    ":memory:",
			expected: ":memory:",
		},
		{
			input:    "file:test.db?mode=memory",
			expected: "file:test.db?mode=memory",
		},
	}

	for _, tt := range tests {
		got := drivers.MaskDSN(tt.input)
		if got != tt.expected {
			t.Errorf("MaskDSN(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
