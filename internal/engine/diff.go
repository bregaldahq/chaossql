package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

// RunDifferentialFuzzing executes the exact same schedule and seed against two database drivers,
// comparing invariant verification results to detect isolation semantic divergence.
func RunDifferentialFuzzing(ctx context.Context, spec domain.Spec, driverA, driverB drivers.DatabaseDriver, seed uint64) (*domain.DiffResult, error) {
	runnerA := NewRunner(driverA, seed)
	runnerB := NewRunner(driverB, seed)

	start := time.Now()
	resA, errA := runnerA.Run(ctx, spec)
	if errA != nil {
		return nil, fmt.Errorf("driver A (%s) execution error: %w", driverA.DriverName(), errA)
	}

	resB, errB := runnerB.Run(ctx, spec)
	if errB != nil {
		return nil, fmt.Errorf("driver B (%s) execution error: %w", driverB.DriverName(), errB)
	}

	_ = start

	divergent := false
	var diffSummary string

	if resA.ViolationDetected != resB.ViolationDetected {
		divergent = true
		diffSummary = fmt.Sprintf("Semantic divergence detected: Driver %s (violation=%v) != Driver %s (violation=%v)",
			driverA.DriverName(), resA.ViolationDetected, driverB.DriverName(), resB.ViolationDetected)
	} else if resA.ViolationDetected && resB.ViolationDetected {
		if resA.FailingInvariant != nil && resB.FailingInvariant != nil && resA.FailingInvariant.Name != resB.FailingInvariant.Name {
			divergent = true
			diffSummary = fmt.Sprintf("Different failing invariant: Driver %s (%s) vs Driver %s (%s)",
				driverA.DriverName(), resA.FailingInvariant.Name, driverB.DriverName(), resB.FailingInvariant.Name)
		} else {
			diffSummary = fmt.Sprintf("Both drivers (%s and %s) detected identical anomaly behavior.", driverA.DriverName(), driverB.DriverName())
		}
	} else {
		diffSummary = fmt.Sprintf("Both drivers (%s and %s) satisfied all invariants consistently.", driverA.DriverName(), driverB.DriverName())
	}

	return &domain.DiffResult{
		ScenarioName: spec.Name,
		DriverA:      driverA.DriverName(),
		DriverB:      driverB.DriverName(),
		ResultA:      resA,
		ResultB:      resB,
		Divergent:    divergent,
		DiffSummary:  diffSummary,
	}, nil
}
