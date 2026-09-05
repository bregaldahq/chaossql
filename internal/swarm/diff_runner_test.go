package swarm_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/swarm"
)

func makeTestSpec(name string, schema string, seed string, query string, assert string) domain.Spec {
	return domain.Spec{
		Version: "1.0",
		Name:    name,
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: schema,
			Seed:   seed,
		},
		Engine: domain.EngineConfig{
			Workers:    1,
			Iterations: 2,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "test_inv",
				Query:  query,
				Assert: assert,
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name: "op1",
				Steps: []domain.StepConfig{
					{SQL: "SELECT 1;"},
				},
			},
		},
	}
}

func TestExecuteDifferentialMatrix_EmptySpecs(t *testing.T) {
	ctx := context.Background()
	report, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{}, []string{"sqlite", "mock"}, 4)
	if err != nil {
		t.Fatalf("unexpected error on empty specs: %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report on empty specs")
	}
	if report.TotalScenarios != 0 {
		t.Errorf("expected 0 total scenarios, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 0 {
		t.Errorf("expected 0 total executions, got %d", report.TotalExecutions)
	}
	if report.DivergentCount != 0 {
		t.Errorf("expected 0 divergent count, got %d", report.DivergentCount)
	}
	if len(report.Scenarios) != 0 {
		t.Errorf("expected empty scenarios slice, got len %d", len(report.Scenarios))
	}
}

func TestExecuteDifferentialMatrix_SingleDriver(t *testing.T) {
	ctx := context.Background()
	spec := makeTestSpec(
		"single_driver_spec",
		"CREATE TABLE items (id INT PRIMARY KEY, qty INT);",
		"INSERT INTO items (id, qty) VALUES (1, 10);",
		"SELECT qty FROM items WHERE id = 1;",
		"qty == 10",
	)

	report, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{spec}, []string{"sqlite"}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalScenarios != 1 {
		t.Errorf("expected 1 scenario, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 1 {
		t.Errorf("expected 1 execution, got %d", report.TotalExecutions)
	}
	if report.DivergentCount != 0 {
		t.Errorf("single driver cannot diverge, got divergent count %d", report.DivergentCount)
	}
	sc := report.Scenarios[0]
	if sc.Divergent {
		t.Errorf("expected non-divergent for single driver, got true")
	}
	res, ok := sc.Results["sqlite"]
	if !ok {
		t.Fatalf("expected result for sqlite driver")
	}
	if !res.Success {
		t.Errorf("expected success for sqlite run, got error: %s", res.Error)
	}
	if res.ViolationDetected {
		t.Errorf("expected no violation for sqlite run")
	}
}

func TestExecuteDifferentialMatrix_DivergenceDetection(t *testing.T) {
	ctx := context.Background()
	// Invariant: balance == 0.
	// In SQLite: initial balance is 100, so balance == 0 is FALSE -> ViolationDetected = true.
	// In Mock: QueryRow returns 0, so balance == 0 is TRUE -> ViolationDetected = false.
	// Result: divergence between sqlite and mock!
	spec := makeTestSpec(
		"divergence_spec",
		"CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
		"INSERT INTO accounts (id, balance) VALUES (1, 100);",
		"SELECT balance FROM accounts WHERE id = 1;",
		"balance == 0",
	)

	report, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{spec}, []string{"sqlite", "mock"}, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalScenarios != 1 {
		t.Fatalf("expected 1 scenario, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 2 {
		t.Errorf("expected 2 executions, got %d", report.TotalExecutions)
	}
	if report.DivergentCount != 1 {
		t.Errorf("expected 1 divergent scenario, got %d", report.DivergentCount)
	}

	sc := report.Scenarios[0]
	if !sc.Divergent {
		t.Errorf("expected scenario to be divergent, got false (summary: %s)", sc.Summary)
	}
	sqliteRes := sc.Results["sqlite"]
	mockRes := sc.Results["mock"]

	if !sqliteRes.ViolationDetected {
		t.Errorf("expected sqlite violation detected, got false")
	}
	if mockRes.ViolationDetected {
		t.Errorf("expected mock violation not detected, got true")
	}
	if sqliteRes.FailingInvariant != "test_inv" {
		t.Errorf("expected failing invariant 'test_inv', got %q", sqliteRes.FailingInvariant)
	}
}

func TestExecuteDifferentialMatrix_DriverErrorResilience(t *testing.T) {
	ctx := context.Background()
	spec := makeTestSpec(
		"resilience_spec",
		"CREATE TABLE test (id INT PRIMARY KEY);",
		"",
		"SELECT 1;",
		"1 == 1",
	)

	// One valid driver (sqlite) and one invalid driver name ("unknown_invalid_driver")
	report, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{spec}, []string{"sqlite", "unknown_invalid_driver"}, 2)
	if err != nil {
		t.Fatalf("matrix should not abort on driver error, got: %v", err)
	}
	if report.TotalScenarios != 1 {
		t.Fatalf("expected 1 scenario, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 2 {
		t.Errorf("expected 2 executions recorded, got %d", report.TotalExecutions)
	}

	sc := report.Scenarios[0]
	invRes, ok := sc.Results["unknown_invalid_driver"]
	if !ok {
		t.Fatalf("expected entry for invalid driver")
	}
	if invRes.Error == "" {
		t.Errorf("expected error message recorded for invalid driver, got empty")
	}
	if invRes.Success {
		t.Errorf("expected Success=false for invalid driver")
	}
}

func TestExecuteDifferentialMatrix_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	spec := makeTestSpec(
		"cancel_spec",
		"CREATE TABLE test (id INT PRIMARY KEY);",
		"",
		"SELECT 1;",
		"1 == 1",
	)

	_, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{spec}, []string{"sqlite", "mock"}, 2)
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestExecuteDifferentialMatrix_MultipleScenariosConcurrency(t *testing.T) {
	ctx := context.Background()
	specs := []domain.Spec{
		makeTestSpec("spec_1", "CREATE TABLE t1 (id INT);", "", "SELECT 1;", "1 == 1"),
		makeTestSpec("spec_2", "CREATE TABLE t2 (id INT);", "", "SELECT 1;", "1 == 1"),
		makeTestSpec("spec_3", "CREATE TABLE t3 (id INT);", "", "SELECT 1;", "1 == 1"),
	}

	start := time.Now()
	report, err := swarm.ExecuteDifferentialMatrix(ctx, specs, []string{"sqlite", "mock"}, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	duration := time.Since(start)

	if report.TotalScenarios != 3 {
		t.Errorf("expected 3 scenarios, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 6 {
		t.Errorf("expected 6 total executions, got %d", report.TotalExecutions)
	}
	if report.DurationMs < 0 {
		t.Errorf("expected non-negative DurationMs, got %d", report.DurationMs)
	}
	_ = duration
}

func TestExecuteDifferentialMatrix_DefaultDriver(t *testing.T) {
	ctx := context.Background()
	spec := makeTestSpec("default_spec", "CREATE TABLE def (id INT);", "", "SELECT 1;", "1 == 1")

	report, err := swarm.ExecuteDifferentialMatrix(ctx, []domain.Spec{spec}, nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.TotalScenarios != 1 {
		t.Errorf("expected 1 scenario, got %d", report.TotalScenarios)
	}
	if _, ok := report.Scenarios[0].Results["sqlite"]; !ok {
		t.Errorf("expected default driver to be sqlite")
	}
}

func TestEvaluateScenarioDivergence_DifferingFailingInvariants(t *testing.T) {
	results := map[string]swarm.DriverExecutionResult{
		"sqlite": {
			Driver:            "sqlite",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_positive_balance",
			DetectedAnomaly:   "P4_LOST_UPDATE",
		},
		"postgres": {
			Driver:            "postgres",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_non_negative_balance",
			DetectedAnomaly:   "P4_LOST_UPDATE",
		},
	}

	diff := swarm.EvaluateScenarioDivergence("test_scenario", results, []string{"sqlite", "postgres"})
	if !diff.Divergent {
		t.Errorf("expected divergence due to different failing invariants")
	}
}

func TestEvaluateScenarioDivergence_DifferingAnomalies(t *testing.T) {
	results := map[string]swarm.DriverExecutionResult{
		"sqlite": {
			Driver:            "sqlite",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_common",
			DetectedAnomaly:   "P4_LOST_UPDATE",
		},
		"postgres": {
			Driver:            "postgres",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_common",
			DetectedAnomaly:   "A5B_WRITE_SKEW",
		},
	}

	diff := swarm.EvaluateScenarioDivergence("test_scenario", results, []string{"sqlite", "postgres"})
	if !diff.Divergent {
		t.Errorf("expected divergence due to different detected anomalies")
	}
}

func TestEvaluateScenarioDivergence_ConsistentViolations(t *testing.T) {
	results := map[string]swarm.DriverExecutionResult{
		"sqlite": {
			Driver:            "sqlite",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_common",
			DetectedAnomaly:   "P4_LOST_UPDATE",
		},
		"postgres": {
			Driver:            "postgres",
			Success:           false,
			ViolationDetected: true,
			FailingInvariant:  "inv_common",
			DetectedAnomaly:   "P4_LOST_UPDATE",
		},
	}

	diff := swarm.EvaluateScenarioDivergence("test_scenario", results, []string{"sqlite", "postgres"})
	if diff.Divergent {
		t.Errorf("expected non-divergent when both detect same invariant and anomaly")
	}
}

func TestEvaluateScenarioDivergence_AllErrors(t *testing.T) {
	results := map[string]swarm.DriverExecutionResult{
		"postgres": {
			Driver:  "postgres",
			Success: false,
			Error:   "connection refused",
		},
		"mysql": {
			Driver:  "mysql",
			Success: false,
			Error:   "connection refused",
		},
	}

	diff := swarm.EvaluateScenarioDivergence("error_scenario", results, []string{"postgres", "mysql"})
	if diff.Divergent {
		t.Errorf("expected non-divergent when all drivers error")
	}
}

