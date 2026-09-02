package engine_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestDifferentialFuzzing_IdenticalDrivers(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver(":memory:")
	driverB := drivers.NewSQLiteDriver(":memory:")

	spec := domain.Spec{
		Name: "banking_diff_test",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			Seed:   "INSERT INTO accounts (id, balance) VALUES (1, 1000);",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 10,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "positive_balance",
				Query:  "SELECT balance FROM accounts WHERE id = 1;",
				Assert: "balance >= 0",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name: "withdraw",
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;", Capture: "cur"},
					{SQL: "UPDATE accounts SET balance = {cur - 10} WHERE id = 1;"},
				},
			},
		},
	}

	diffRes, err := engine.RunDifferentialFuzzing(ctx, spec, driverA, driverB, 42)
	if err != nil {
		t.Fatalf("unexpected error in RunDifferentialFuzzing: %v", err)
	}

	if diffRes == nil {
		t.Fatal("expected non-nil DiffResult")
	}

	if diffRes.Divergent {
		t.Errorf("expected identical SQLite drivers with same seed to be non-divergent, got: %s", diffRes.DiffSummary)
	}

	if diffRes.ScenarioName != "banking_diff_test" {
		t.Errorf("expected scenario name 'banking_diff_test', got '%s'", diffRes.ScenarioName)
	}
}
