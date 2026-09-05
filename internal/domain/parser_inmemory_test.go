package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestParseSpecBytes_Valid(t *testing.T) {
	yamlData := []byte(`
version: "1.0"
name: "in_memory_lost_update"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "balance_positive"
    type: "sql"
    query: "SELECT COUNT(*) FROM accounts WHERE balance < 0;"
    expected: "0"
operations:
  - name: "withdraw"
    steps:
      - sql: "UPDATE accounts SET balance = balance - 100 WHERE id = 1;"
`)

	spec, err := domain.ParseSpecBytes(yamlData)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if spec.Name != "in_memory_lost_update" {
		t.Errorf("expected name 'in_memory_lost_update', got: %s", spec.Name)
	}
	if spec.Database.Schema == "" || spec.Database.Seed == "" {
		t.Errorf("expected inline schema and seed to be preserved")
	}
}

func TestParseSpecString_WithOverrides(t *testing.T) {
	yamlStr := `
version: "1.0"
name: "override_test"
database:
  driver: "sqlite"
invariants:
  - name: "inv1"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
operations:
  - name: "op1"
    steps:
      - sql: "SELECT 1"
`
	schema := "CREATE TABLE t (x INT);"
	seed := "INSERT INTO t VALUES (42);"

	spec, err := domain.ParseSpecString(yamlStr, schema, seed)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if spec.Database.Schema != schema || spec.Database.Seed != seed {
		t.Errorf("expected schema and seed overrides to be set")
	}
}

func TestParseSpecBytes_Errors(t *testing.T) {
	tests := []struct {
		name          string
		yamlData      string
		expectedErr   error
		errorContains string
	}{
		{
			name:          "Malformed YAML",
			yamlData:      "version: '1.0'\n  invalid: \n- unindented",
			errorContains: "failed to parse yaml",
		},
		{
			name: "Missing Version",
			yamlData: `
name: "test"
database:
  driver: "sqlite"
invariants:
  - name: "inv"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
operations:
  - name: "op"
    steps:
      - sql: "SELECT 1"
`,
			expectedErr:   domain.ErrSpecValidationFailed,
			errorContains: "missing or empty 'version'",
		},
		{
			name: "Missing Name",
			yamlData: `
version: "1.0"
database:
  driver: "sqlite"
invariants:
  - name: "inv"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
operations:
  - name: "op"
    steps:
      - sql: "SELECT 1"
`,
			expectedErr:   domain.ErrSpecValidationFailed,
			errorContains: "missing or empty 'name'",
		},
		{
			name: "Missing Driver",
			yamlData: `
version: "1.0"
name: "test"
invariants:
  - name: "inv"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
operations:
  - name: "op"
    steps:
      - sql: "SELECT 1"
`,
			expectedErr:   domain.ErrSpecValidationFailed,
			errorContains: "missing or empty 'database.driver'",
		},
		{
			name: "Missing Invariants",
			yamlData: `
version: "1.0"
name: "test"
database:
  driver: "sqlite"
operations:
  - name: "op"
    steps:
      - sql: "SELECT 1"
`,
			expectedErr:   domain.ErrSpecValidationFailed,
			errorContains: "'invariants' must have at least one entry",
		},
		{
			name: "Missing Operations",
			yamlData: `
version: "1.0"
name: "test"
database:
  driver: "sqlite"
invariants:
  - name: "inv"
    type: "sql"
    query: "SELECT 1"
    expected: "1"
`,
			expectedErr:   domain.ErrSpecValidationFailed,
			errorContains: "'operations' must have at least one entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ParseSpecBytes([]byte(tt.yamlData))
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error wrapping %v, got %v", tt.expectedErr, err)
			}
			if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
				t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
			}
		})
	}
}

