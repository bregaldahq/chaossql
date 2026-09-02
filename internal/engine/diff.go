package engine

import (
	"context"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
)

// RunDifferentialFuzzing executes an identical chaos schedule against two database drivers
// and evaluates whether their isolation semantics diverge (e.g., one driver violates an invariant
// while the other satisfies it, or they produce divergent invariant violation outcomes).
func RunDifferentialFuzzing(ctx context.Context, spec domain.Spec, driverA, driverB drivers.DatabaseDriver, seed int64) (*domain.DiffResult, error) {
	if driverA == nil || driverB == nil {
		return nil, fmt.Errorf("both driverA and driverB must be non-nil")
	}

	effectiveSeed := spec.Engine.Seed
	if seed != 0 {
		effectiveSeed = uint64(seed)
	}
	if effectiveSeed == 0 {
		effectiveSeed = 42
	}
	spec.Engine.Seed = effectiveSeed

	// Generate deterministic schedule for both drivers
	prng := NewPRNG(effectiveSeed)
	scheduledOps := GenerateSchedule(spec, prng)

	// Execute schedule on Driver A
	runnerA := NewRunner(driverA, effectiveSeed)
	resultA, errA := runnerA.RunSchedule(ctx, spec, scheduledOps)
	if errA != nil {
		return nil, fmt.Errorf("driver A (%s) execution failed: %w", driverA.DriverName(), errA)
	}

	// Execute schedule on Driver B
	runnerB := NewRunner(driverB, effectiveSeed)
	resultB, errB := runnerB.RunSchedule(ctx, spec, scheduledOps)
	if errB != nil {
		return nil, fmt.Errorf("driver B (%s) execution failed: %w", driverB.DriverName(), errB)
	}

	nameA := driverA.DriverName()
	if nameA == "" {
		nameA = "driver_a"
	}
	nameB := driverB.DriverName()
	if nameB == "" {
		nameB = "driver_b"
	}

	divergent := false
	var diffSummary string

	if resultA.ViolationDetected != resultB.ViolationDetected {
		divergent = true
		if resultA.ViolationDetected {
			invName := ""
			if resultA.FailingInvariant != nil {
				invName = resultA.FailingInvariant.Name
			}
			diffSummary = fmt.Sprintf("Isolation divergence detected: %s violated invariant '%s' while %s satisfied all invariants", nameA, invName, nameB)
		} else {
			invName := ""
			if resultB.FailingInvariant != nil {
				invName = resultB.FailingInvariant.Name
			}
			diffSummary = fmt.Sprintf("Isolation divergence detected: %s satisfied all invariants while %s violated invariant '%s'", nameA, nameB, invName)
		}
	} else if resultA.ViolationDetected && resultB.ViolationDetected {
		invA := ""
		if resultA.FailingInvariant != nil {
			invA = resultA.FailingInvariant.Name
		}
		invB := ""
		if resultB.FailingInvariant != nil {
			invB = resultB.FailingInvariant.Name
		}
		if invA != invB {
			divergent = true
			diffSummary = fmt.Sprintf("Isolation divergence detected: %s violated invariant '%s' while %s violated invariant '%s'", nameA, invA, nameB, invB)
		} else {
			divergent = false
			diffSummary = fmt.Sprintf("No divergence: both %s and %s violated invariant '%s'", nameA, nameB, invA)
		}
	} else {
		divergent = false
		diffSummary = fmt.Sprintf("No divergence: both %s and %s satisfied all invariants", nameA, nameB)
	}

	return &domain.DiffResult{
		ScenarioName: spec.Name,
		DriverA:      nameA,
		DriverB:      nameB,
		ResultA:      resultA,
		ResultB:      resultB,
		Divergent:    divergent,
		DiffSummary:  diffSummary,
	}, nil
}
