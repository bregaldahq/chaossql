package engine_test

import (
	"context"
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
