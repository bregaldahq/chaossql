package domain_test

import (
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
