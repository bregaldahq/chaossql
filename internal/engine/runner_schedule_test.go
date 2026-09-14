package engine_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestExecuteSchedule_UsesDeterministicWorkerQueues(t *testing.T) {
	const workers = 3
	spec := domain.Spec{Engine: domain.EngineConfig{Workers: workers}}
	ops := make([]domain.ScheduledOp, 30)
	for i := range ops {
		ops[i] = domain.ScheduledOp{
			ID:    i + 1,
			Name:  "write",
			Steps: []domain.StepConfig{{SQL: "UPDATE records SET value = 1"}},
		}
	}

	for run := 0; run < 20; run++ {
		driver := drivers.NewMockDriver()
		outcome, err := engine.NewRunner(driver, 44).ExecuteSchedule(context.Background(), spec, ops)
		_ = driver.Close()
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}

		seenBegin := make(map[int]bool, len(ops))
		for _, event := range outcome.Trace {
			if event.Type != domain.EventBegin {
				continue
			}
			seenBegin[event.OpIndex] = true
			wantWorker := (event.OpIndex-1)%workers + 1
			if event.WorkerID != wantWorker {
				t.Fatalf("run %d operation %d worker = %d, want %d", run, event.OpIndex, event.WorkerID, wantWorker)
			}
		}
		if len(seenBegin) != len(ops) {
			t.Fatalf("run %d saw %d begins, want %d", run, len(seenBegin), len(ops))
		}
	}
}

func TestExecuteSchedule_RejectsInvalidOperationIDs(t *testing.T) {
	tests := []struct {
		name string
		ops  []domain.ScheduledOp
		want string
	}{
		{name: "zero", ops: []domain.ScheduledOp{{ID: 0}}, want: "must be positive"},
		{name: "negative", ops: []domain.ScheduledOp{{ID: -1}}, want: "must be positive"},
		{name: "duplicate", ops: []domain.ScheduledOp{{ID: 1}, {ID: 1}}, want: "must be unique"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver := drivers.NewMockDriver()
			defer driver.Close()
			_, err := engine.NewRunner(driver, 1).ExecuteSchedule(context.Background(), domain.Spec{}, tt.ops)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want text %q", err, tt.want)
			}
		})
	}
}

func TestExecuteSchedule_UsesPrecomputedAbortDecisions(t *testing.T) {
	spec := domain.Spec{Engine: domain.EngineConfig{
		Workers: 4,
		Faults:  domain.FaultConfig{AbortProbability: 0.5},
	}}
	ops := make([]domain.ScheduledOp, 40)
	for i := range ops {
		ops[i] = domain.ScheduledOp{
			ID:    i + 1,
			Name:  "write",
			Steps: []domain.StepConfig{{SQL: "UPDATE records SET value = 1"}},
		}
	}
	wantPlan := engine.BuildSchedulePlan(spec, ops, engine.NewPRNG(55))
	wantAbort := make(map[int]bool, len(wantPlan.Decisions))
	for _, decision := range wantPlan.Decisions {
		wantAbort[decision.OperationID] = decision.Abort
	}

	driver := drivers.NewMockDriver()
	defer driver.Close()
	outcome, err := engine.NewRunner(driver, 55).ExecuteSchedule(context.Background(), spec, ops)
	if err != nil {
		t.Fatal(err)
	}
	gotAbort := make(map[int]bool, len(ops))
	for _, event := range outcome.Trace {
		if event.Type == domain.EventRollback {
			gotAbort[event.OpIndex] = true
		}
	}
	for operationID, want := range wantAbort {
		if got := gotAbort[operationID]; got != want {
			t.Fatalf("operation %d abort = %t, want planned %t", operationID, got, want)
		}
	}
}

func TestRunner_PersistsEffectiveSeedAndSchedule(t *testing.T) {
	driver := drivers.NewMockDriver()
	defer driver.Close()
	spec := domain.Spec{
		Engine: domain.EngineConfig{Workers: 2, Iterations: 2, Seed: 0},
		Invariants: []domain.InvariantConfig{{
			Name:   "ok",
			Query:  "SELECT 1 AS value",
			Assert: "value == 1",
		}},
		Operations: []domain.OperationConfig{{
			Name:  "read",
			Steps: []domain.StepConfig{{SQL: "SELECT 1"}},
		}},
	}

	result, err := engine.NewRunner(driver, 0).Run(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if result.Seed != 0 || result.Schedule.Seed != 0 {
		t.Fatalf("expected effective seed 0 in result and schedule: %#v", result)
	}
	if result.Schedule.Version != 1 || result.Schedule.Workers != 2 || len(result.Schedule.Decisions) != 2 {
		t.Fatalf("unexpected persisted schedule: %#v", result.Schedule)
	}
}
