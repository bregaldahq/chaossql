package engine_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestBuildSchedulePlan_IsDeterministic(t *testing.T) {
	spec := domain.Spec{
		Engine: domain.EngineConfig{
			Workers:  3,
			JitterMs: [2]int{1, 20},
			Faults: domain.FaultConfig{
				AbortProbability:   0.4,
				LatencyProbability: 0.8,
				LatencySpikeMs:     [2]int{5, 25},
			},
		},
	}
	ops := []domain.ScheduledOp{
		{ID: 3, Name: "third", Steps: []domain.StepConfig{{SQL: "SELECT 3"}}},
		{ID: 1, Name: "first", Steps: []domain.StepConfig{{SQL: "SELECT 1"}, {SQL: "UPDATE t SET v=1"}}},
		{ID: 2, Name: "second", Steps: []domain.StepConfig{{SQL: "SELECT 2"}}},
	}

	want := engine.BuildSchedulePlan(spec, ops, engine.NewPRNG(77))
	for i := 0; i < 100; i++ {
		got := engine.BuildSchedulePlan(spec, ops, engine.NewPRNG(77))
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("plan %d differs\nwant: %#v\n got: %#v", i, want, got)
		}
	}

	if want.Version != 1 || want.Seed != 77 || want.Workers != 3 {
		t.Fatalf("unexpected plan metadata: %#v", want)
	}
	if len(want.Decisions) != 4 {
		t.Fatalf("expected four step decisions, got %d", len(want.Decisions))
	}
	wantIdentity := [][3]int{{1, 1, 1}, {2, 1, 2}, {3, 2, 1}, {4, 3, 1}}
	for stepIndex, decision := range want.Decisions {
		got := [3]int{decision.Sequence, decision.OperationID, decision.StepIndex}
		if got != wantIdentity[stepIndex] {
			t.Fatalf("step decision %d identity = %v, want %v", stepIndex, got, wantIdentity[stepIndex])
		}
		wantWorker := (decision.OperationID-1)%3 + 1
		if decision.WorkerID != wantWorker {
			t.Fatalf("operation %d worker = %d, want %d", decision.OperationID, decision.WorkerID, wantWorker)
		}
	}
}

func TestBuildSchedulePlan_DecisionsDependOnIdentityNotInputOrder(t *testing.T) {
	spec := domain.Spec{Engine: domain.EngineConfig{
		Workers:  2,
		JitterMs: [2]int{1, 100},
		Faults: domain.FaultConfig{
			AbortProbability:   0.5,
			LatencyProbability: 1,
			LatencySpikeMs:     [2]int{10, 50},
		},
	}}
	first := domain.ScheduledOp{ID: 1, Steps: []domain.StepConfig{{SQL: "SELECT 1"}}}
	second := domain.ScheduledOp{ID: 2, Steps: []domain.StepConfig{{SQL: "SELECT 2"}}}

	a := engine.BuildSchedulePlan(spec, []domain.ScheduledOp{first, second}, engine.NewPRNG(91))
	b := engine.BuildSchedulePlan(spec, []domain.ScheduledOp{second, first}, engine.NewPRNG(91))
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("input completion/order changed decisions\na: %#v\nb: %#v", a, b)
	}

	c := engine.BuildSchedulePlan(spec, []domain.ScheduledOp{first, second}, engine.NewPRNG(92))
	if reflect.DeepEqual(a.Decisions, c.Decisions) {
		t.Fatal("different seeds produced identical decisions")
	}
}

func TestBuildSchedulePlan_EmptyDecisionsEncodeAsArray(t *testing.T) {
	plan := engine.BuildSchedulePlan(domain.Spec{}, nil, engine.NewPRNG(0))
	encoded, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(encoded, []byte(`"decisions":[]`)) {
		t.Fatalf("empty decisions must encode as an array: %s", encoded)
	}
}

func TestDeterministicScheduleV1GoldenCorpus(t *testing.T) {
	spec := domain.Spec{
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 4,
			Seed:       0,
			JitterMs:   [2]int{2, 9},
			Faults: domain.FaultConfig{
				AbortProbability:   0.35,
				LatencyProbability: 0.75,
				LatencySpikeMs:     [2]int{4, 12},
			},
		},
		Operations: []domain.OperationConfig{{
			Name: "corpus_write",
			Params: map[string]string{
				"amount":   "$random_int(10, 99)",
				"category": "$random_choice('alpha', 'beta', 'gamma')",
				"sequence": "$monotonic_counter(100, 5)",
			},
			Steps: []domain.StepConfig{
				{SQL: "SELECT value FROM records WHERE id = {sequence}", Capture: "current"},
				{SQL: "UPDATE records SET value = {amount} WHERE id = {sequence}"},
			},
		}},
	}
	prng := engine.NewPRNG(spec.Engine.Seed)
	ops := engine.GenerateSchedule(spec, prng)
	artifact := struct {
		Operations []domain.ScheduledOp `json:"operations"`
		Schedule   domain.SchedulePlan  `json:"schedule"`
	}{
		Operations: ops,
		Schedule:   engine.BuildSchedulePlan(spec, ops, prng),
	}
	got, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')

	goldenPath := filepath.Join("testdata", "deterministic_schedule_v1.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("deterministic schedule corpus changed; bump the schedule version or run an intentional golden update\nwant:\n%s\ngot:\n%s", want, got)
	}
}
