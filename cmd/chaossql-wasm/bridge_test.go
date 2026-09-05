package main

import (
	"context"
	"testing"
)

func TestValidateScenarioYAML_Valid(t *testing.T) {
	validYAML := `
version: "1.0"
name: "banking_test"
database:
  driver: "sqlite"
  schema: "CREATE TABLE accounts (id INT, balance INT);"
  seed: "INSERT INTO accounts VALUES (1, 1000);"
invariants:
  - name: "positive_balance"
    type: "sql"
    query: "SELECT COUNT(*) FROM accounts WHERE balance < 0;"
    expected: "0"
operations:
  - name: "transfer"
    steps:
      - sql: "UPDATE accounts SET balance = balance - 10 WHERE id = 1;"
`
	res := ValidateScenarioYAML(validYAML)
	if !res.Valid {
		t.Fatalf("expected valid YAML, got error: %s", res.Error)
	}
	if res.Name != "banking_test" {
		t.Errorf("expected scenario name 'banking_test', got: %s", res.Name)
	}
	if res.NumOperations != 1 || res.NumInvariants != 1 {
		t.Errorf("unexpected counts: ops=%d, invs=%d", res.NumOperations, res.NumInvariants)
	}
}

func TestValidateScenarioYAML_Invalid(t *testing.T) {
	invalidYAML := `not: valid: yaml:`
	res := ValidateScenarioYAML(invalidYAML)
	if res.Valid {
		t.Errorf("expected invalid result for corrupted YAML")
	}
	if res.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}

func TestExecuteWasmScenario_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	configJSON := `{"yamlContent":"version: '1.0'\nname: 'test'\ndatabase:\n  driver: 'sqlite'\ninvariants:\n  - name: 'i'\n    type: 'sql'\n    query: 'SELECT 1'\n    expected: '1'\noperations:\n  - name: 'o'\n    steps:\n      - sql: 'SELECT 1'","workers":2,"iterations":10}`
	_, err := ExecuteWasmScenario(ctx, configJSON, nil)
	if err == nil {
		t.Fatalf("expected error on cancelled context, got nil")
	}
}
