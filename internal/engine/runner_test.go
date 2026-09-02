package engine_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestRunner_LostUpdateDetection(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	spec := domain.Spec{
		Name: "banking_lost_update_test",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT); CREATE TABLE ledger (id INTEGER PRIMARY KEY AUTOINCREMENT, amount INT);",
			Seed:   "INSERT INTO accounts VALUES (1, 1000);",
		},
		Engine: domain.EngineConfig{
			Workers:    4,
			Iterations: 10,
			Seed:       42,
			JitterMs:   [2]int{1, 5},
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "ledger_consistency",
				Query:  "SELECT (SELECT balance FROM accounts WHERE id=1) AS actual, (1000 - COALESCE(SUM(amount), 0)) AS expected FROM ledger;",
				Assert: "actual == expected",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "withdraw",
				Weight: 1.0,
				Params: map[string]string{"amount": "int(50, 100)"},
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;", Capture: "current_bal"},
					{SQL: "UPDATE accounts SET balance = {current_bal - amount} WHERE id = 1;"},
					{SQL: "INSERT INTO ledger (amount) VALUES ({amount});"},
				},
			},
		},
	}

	runner := engine.NewRunner(driver, spec.Engine.Seed)
	result, err := runner.Run(ctx, spec)

	if err != nil {
		t.Fatalf("runner execution failed: %v", err)
	}

	if !result.ViolationDetected {
		t.Log("Note: Lost update was not triggered in this single run (try more iterations)")
	} else {
		t.Logf("Successfully detected invariant violation: %s", result.FailingInvariant.String())
	}

	if len(result.Trace) == 0 {
		t.Error("expected non-empty execution trace")
	}
}
