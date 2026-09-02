package engine

import (
	"testing"
)

func TestSubstituteParams(t *testing.T) {
	state := map[string]string{
		"amount":      "100",
		"current_bal": "1000",
	}

	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		{
			name:     "direct substitution",
			sql:      "INSERT INTO t (val) VALUES ({amount})",
			expected: "INSERT INTO t (val) VALUES (100)",
		},
		{
			name:     "simple arithmetic subtraction",
			sql:      "UPDATE accounts SET balance = {current_bal - amount}",
			expected: "UPDATE accounts SET balance = 900",
		},
		{
			name:     "simple arithmetic addition",
			sql:      "UPDATE accounts SET balance = {current_bal + amount}",
			expected: "UPDATE accounts SET balance = 1100",
		},
		{
			name:     "multiple substitutions",
			sql:      "VALUES ({amount}, {current_bal - amount})",
			expected: "VALUES (100, 900)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteParams(tt.sql, state)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
