package engine_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestRunDifferentialFuzzing_NonDivergent_BothPass(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_pass_a?mode=memory&cache=shared")
	defer driverA.Close()
	driverB := drivers.NewSQLiteDriver("file:diff_pass_b?mode=memory&cache=shared")
	defer driverB.Close()

	spec := domain.Spec{
		Name: "diff_test_pass",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE counters (id INT PRIMARY KEY, val INT);",
			Seed:   "INSERT INTO counters VALUES (1, 0);",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 6,
			Seed:       100,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "always_non_negative",
				Query:  "SELECT val FROM counters WHERE id = 1;",
				Assert: "val >= 0",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "inc",
				Weight: 1.0,
				Steps: []domain.StepConfig{
					{SQL: "UPDATE counters SET val = val + 1 WHERE id = 1;"},
				},
			},
		},
	}

	diffRes, err := engine.RunDifferentialFuzzing(ctx, spec, driverA, driverB, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diffRes == nil {
		t.Fatal("expected non-nil DiffResult")
	}
	if diffRes.Divergent {
		t.Errorf("expected non-divergent result, got divergent: %s", diffRes.DiffSummary)
	}
	if diffRes.ScenarioName != spec.Name {
		t.Errorf("expected scenario name %s, got %s", spec.Name, diffRes.ScenarioName)
	}
	if diffRes.DriverA != "sqlite" || diffRes.DriverB != "sqlite" {
		t.Errorf("unexpected driver names: %s, %s", diffRes.DriverA, diffRes.DriverB)
	}
	if diffRes.ResultA == nil || diffRes.ResultB == nil {
		t.Fatal("expected non-nil ResultA and ResultB")
	}
	if diffRes.ResultA.ViolationDetected || diffRes.ResultB.ViolationDetected {
		t.Errorf("expected no violations, got A: %v, B: %v", diffRes.ResultA.ViolationDetected, diffRes.ResultB.ViolationDetected)
	}
	if !strings.Contains(diffRes.DiffSummary, "satisfied all invariants") {
		t.Errorf("unexpected summary: %s", diffRes.DiffSummary)
	}
}

type wrappedDriver struct {
	drivers.DatabaseDriver
	name     string
	resetErr error
}

func (w *wrappedDriver) DriverName() string {
	if w.name != "" {
		return w.name
	}
	return w.DatabaseDriver.DriverName()
}

func (w *wrappedDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	if w.resetErr != nil {
		return w.resetErr
	}
	return w.DatabaseDriver.Reset(ctx, schemaSQL, seedSQL)
}

func TestRunDifferentialFuzzing_NilDriverError(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_nil?mode=memory&cache=shared")
	defer driverA.Close()

	spec := domain.Spec{Name: "nil_test"}

	if _, err := engine.RunDifferentialFuzzing(ctx, spec, nil, driverA, 1); err == nil {
		t.Error("expected error for nil driverA")
	}
	if _, err := engine.RunDifferentialFuzzing(ctx, spec, driverA, nil, 1); err == nil {
		t.Error("expected error for nil driverB")
	}
}

func TestRunDifferentialFuzzing_ResetError(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_err_a?mode=memory&cache=shared")
	defer driverA.Close()
	driverB := drivers.NewSQLiteDriver("file:diff_err_b?mode=memory&cache=shared")
	defer driverB.Close()

	wrapA := &wrappedDriver{
		DatabaseDriver: driverA,
		resetErr:       errors.New("failed to reset DB"),
	}
	spec := domain.Spec{Name: "reset_error_test"}

	_, err := engine.RunDifferentialFuzzing(ctx, spec, wrapA, driverB, 1)
	if err == nil {
		t.Error("expected error when driver reset fails")
	}
}

func TestRunDifferentialFuzzing_BothFailSameInvariant(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_fail_a?mode=memory&cache=shared")
	defer driverA.Close()
	driverB := drivers.NewSQLiteDriver("file:diff_fail_b?mode=memory&cache=shared")
	defer driverB.Close()

	spec := domain.Spec{
		Name: "diff_test_both_fail",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE test_data (id INT PRIMARY KEY);",
			Seed:   "INSERT INTO test_data VALUES (1);",
		},
		Engine: domain.EngineConfig{
			Workers:    1,
			Iterations: 2,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "always_fail_inv",
				Query:  "SELECT id FROM test_data WHERE id = 1;",
				Assert: "id == 9999",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "noop",
				Weight: 1.0,
				Steps: []domain.StepConfig{
					{SQL: "SELECT 1;"},
				},
			},
		},
	}

	diffRes, err := engine.RunDifferentialFuzzing(ctx, spec, driverA, driverB, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diffRes.Divergent {
		t.Errorf("expected non-divergent when both fail the same invariant, got divergent: %s", diffRes.DiffSummary)
	}
	if !diffRes.ResultA.ViolationDetected || !diffRes.ResultB.ViolationDetected {
		t.Errorf("expected both violations detected, got A: %v, B: %v", diffRes.ResultA.ViolationDetected, diffRes.ResultB.ViolationDetected)
	}
	if !strings.Contains(diffRes.DiffSummary, "always_fail_inv") {
		t.Errorf("expected summary to mention failing invariant, got: %s", diffRes.DiffSummary)
	}
}

type customSeedDriver struct {
	drivers.DatabaseDriver
	name       string
	customSeed string
}

func (c *customSeedDriver) DriverName() string {
	if c.name != "" {
		return c.name
	}
	return c.DatabaseDriver.DriverName()
}

func (c *customSeedDriver) Reset(ctx context.Context, schemaSQL, seedSQL string) error {
	seed := seedSQL
	if c.customSeed != "" {
		seed = c.customSeed
	}
	return c.DatabaseDriver.Reset(ctx, schemaSQL, seed)
}

func TestRunDifferentialFuzzing_Divergent_OneViolatesOneSatisfies(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_div_inv_a?mode=memory&cache=shared")
	defer driverA.Close()
	driverB := drivers.NewSQLiteDriver("file:diff_div_inv_b?mode=memory&cache=shared")
	defer driverB.Close()

	specA := domain.Spec{
		Name: "diff_divergence_test",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE vals (id INT PRIMARY KEY, val INT);",
			Seed:   "INSERT INTO vals VALUES (1, 0);",
		},
		Engine: domain.EngineConfig{
			Workers:    1,
			Iterations: 1,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "val_is_positive",
				Query:  "SELECT val FROM vals WHERE id = 1;",
				Assert: "val > 0",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "noop",
				Weight: 1.0,
				Steps:  []domain.StepConfig{{SQL: "SELECT 1;"}},
			},
		},
	}

	customWrapB := &customSeedDriver{
		DatabaseDriver: driverB,
		name:           "passing_engine",
		customSeed:     "INSERT INTO vals VALUES (1, 10);",
	}

	diffRes, err := engine.RunDifferentialFuzzing(ctx, specA, driverA, customWrapB, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !diffRes.Divergent {
		t.Errorf("expected divergent result, got non-divergent: %s", diffRes.DiffSummary)
	}
	if !diffRes.ResultA.ViolationDetected || diffRes.ResultB.ViolationDetected {
		t.Errorf("expected DriverA violation and DriverB pass, got A=%v B=%v", diffRes.ResultA.ViolationDetected, diffRes.ResultB.ViolationDetected)
	}
	if !strings.Contains(diffRes.DiffSummary, "divergence") && !strings.Contains(diffRes.DiffSummary, "Divergence") {
		t.Errorf("expected summary to mention divergence, got: %s", diffRes.DiffSummary)
	}
}

func TestRunDifferentialFuzzing_Divergent_DifferentInvariants(t *testing.T) {
	ctx := context.Background()
	driverA := drivers.NewSQLiteDriver("file:diff_diffinv_a?mode=memory&cache=shared")
	defer driverA.Close()
	driverB := drivers.NewSQLiteDriver("file:diff_diffinv_b?mode=memory&cache=shared")
	defer driverB.Close()

	spec := domain.Spec{
		Name: "diff_two_invariants",
		Database: domain.DatabaseConfig{
			Schema: "CREATE TABLE t (id INT PRIMARY KEY, val1 INT, val2 INT);",
			Seed:   "INSERT INTO t VALUES (1, 0, 10);",
		},
		Engine: domain.EngineConfig{
			Workers:    1,
			Iterations: 1,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "inv_val1_positive",
				Query:  "SELECT val1 FROM t WHERE id = 1;",
				Assert: "val1 > 0",
			},
			{
				Name:   "inv_val2_positive",
				Query:  "SELECT val2 FROM t WHERE id = 1;",
				Assert: "val2 > 0",
			},
		},
		Operations: []domain.OperationConfig{
			{
				Name:   "noop",
				Weight: 1.0,
				Steps:  []domain.StepConfig{{SQL: "SELECT 1;"}},
			},
		},
	}

	driverBWrap := &customSeedDriver{
		DatabaseDriver: driverB,
		name:           "driver_b_inverted",
		customSeed:     "INSERT INTO t VALUES (1, 10, 0);",
	}

	diffRes, err := engine.RunDifferentialFuzzing(ctx, spec, driverA, driverBWrap, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !diffRes.Divergent {
		t.Errorf("expected divergent when different invariants fail, got summary: %s", diffRes.DiffSummary)
	}
	if !strings.Contains(diffRes.DiffSummary, "inv_val1_positive") || !strings.Contains(diffRes.DiffSummary, "inv_val2_positive") {
		t.Errorf("expected summary to mention both invariants, got: %s", diffRes.DiffSummary)
	}
}
