package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
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

func TestExecuteWasmScenario_Success(t *testing.T) {
	ctx := context.Background()
	validYAML := `
version: "1.0"
name: "success_test"
database:
  driver: "sqlite"
invariants:
  - name: "always_zero"
    type: "sql"
    query: "SELECT count FROM accounts;"
    assert: "count == 0"
operations:
  - name: "noop"
    steps:
      - sql: "SELECT 1;"
`
	configJSON := fmt.Sprintf(`{"yamlContent":%q,"workers":2,"iterations":5}`, validYAML)
	report, err := ExecuteWasmScenario(ctx, configJSON, nil)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report")
	}
	if !report.Success {
		t.Errorf("expected report.Success == true, got false")
	}
	if report.ViolationFound {
		t.Errorf("expected report.ViolationFound == false, got true")
	}
	if report.TotalOps <= 0 {
		t.Errorf("expected report.TotalOps > 0, got %d", report.TotalOps)
	}
	if report.DurationMs < 0 {
		t.Errorf("expected report.DurationMs >= 0, got %d", report.DurationMs)
	}
}

func TestExecuteWasmScenario_ViolationAndShrink(t *testing.T) {
	ctx := context.Background()
	violationYAML := `
version: "1.0"
name: "violation_shrink_test"
database:
  driver: "sqlite"
invariants:
  - name: "fail_condition"
    type: "sql"
    query: "SELECT count FROM accounts;"
    assert: "count > 0"
operations:
  - name: "op_step"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
      - sql: "UPDATE accounts SET balance = balance - 10 WHERE id = 1;"
`
	configJSON := fmt.Sprintf(`{"yamlContent":%q,"workers":2,"iterations":10}`, violationYAML)
	report, err := ExecuteWasmScenario(ctx, configJSON, nil)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report")
	}
	if !report.ViolationFound {
		t.Errorf("expected report.ViolationFound == true, got false")
	}
	if report.Success {
		t.Errorf("expected report.Success == false, got true")
	}
	if report.FailingInvariant == "" {
		t.Errorf("expected report.FailingInvariant to be non-empty")
	}
	if report.ReducedOps > report.TotalOps {
		t.Errorf("expected report.ReducedOps (%d) <= report.TotalOps (%d)", report.ReducedOps, report.TotalOps)
	}
}

func TestExecuteWasmScenario_MidFlightCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	violationYAML := `
version: "1.0"
name: "mid_cancellation"
database:
  driver: "sqlite"
invariants:
  - name: "always_fail"
    type: "sql"
    query: "SELECT count FROM accounts;"
    assert: "count == 100"
operations:
  - name: "op1"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
      - sql: "UPDATE accounts SET balance = balance - 10 WHERE id = 1;"
`
	configJSON := fmt.Sprintf(`{"yamlContent":%q,"workers":2,"iterations":10,"jitterMs":5}`, violationYAML)

	var cancelOnce sync.Once
	onProgress := func(ev ProgressEvent) {
		cancelOnce.Do(func() {
			cancel()
		})
	}

	timer := time.AfterFunc(2*time.Millisecond, func() {
		cancelOnce.Do(func() {
			cancel()
		})
	})
	defer timer.Stop()

	report, err := ExecuteWasmScenario(ctx, configJSON, onProgress)
	if report != nil {
		t.Errorf("expected nil report on cancellation, got %+v", report)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected err == context.Canceled, got: %v", err)
	}
}

func TestExecuteWasmScenario_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	_, err := ExecuteWasmScenario(ctx, "not-valid-json", nil)
	if err == nil {
		t.Fatalf("expected non-nil error for invalid JSON, got nil")
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
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected err == context.Canceled, got: %v", err)
	}
}

func TestExecuteWasmScenario_CycleDetectionEdges(t *testing.T) {
	ctx := context.Background()
	scenarioYAML := `
version: "1.0"
name: "cycle_edges_test"
database:
  driver: "sqlite"
invariants:
  - name: "always_ok"
    type: "sql"
    query: "SELECT count FROM accounts;"
    assert: "count == 0"
operations:
  - name: "tx_op"
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
      - sql: "UPDATE accounts SET balance = balance + 10 WHERE id = 1;"
`
	configJSON := fmt.Sprintf(`{"yamlContent":%q,"workers":4,"iterations":30,"seed":42}`, scenarioYAML)

	var cycleEvents []ProgressEvent
	var mu sync.Mutex
	onProgress := func(ev ProgressEvent) {
		mu.Lock()
		defer mu.Unlock()
		if ev.Type == "CYCLE_DETECTED" {
			cycleEvents = append(cycleEvents, ev)
		}
	}

	report, err := ExecuteWasmScenario(ctx, configJSON, onProgress)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(cycleEvents) > 0 {
		for _, ev := range cycleEvents {
			if len(ev.Edges) == 0 {
				t.Fatalf("expected cycle event to have non-empty edges")
			}
			for _, edge := range ev.Edges {
				if !strings.Contains(edge, "->") {
					t.Errorf("expected formatted edge containing '->', got: %s", edge)
				}
			}
		}
	}
}

func TestExecuteWasmScenario_ExplicitZeroJitter(t *testing.T) {
	ctx := context.Background()
	scenarioYAML := `
version: "1.0"
name: "zero_jitter_test"
database:
  driver: "sqlite"
invariants:
  - name: "always_ok"
    type: "sql"
    query: "SELECT count FROM accounts;"
    assert: "count == 0"
operations:
  - name: "noop"
    steps:
      - sql: "SELECT 1;"
`
	configJSON := fmt.Sprintf(`{"yamlContent":%q,"workers":1,"iterations":1,"jitterMs":0}`, scenarioYAML)
	report, err := ExecuteWasmScenario(ctx, configJSON, nil)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}
	if report == nil {
		t.Fatalf("expected non-nil report")
	}
	if !report.Success {
		t.Errorf("expected report.Success == true, got false")
	}
}

