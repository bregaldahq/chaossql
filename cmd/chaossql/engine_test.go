package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestEngineCmd_PassingInvariant(t *testing.T) {
	inputPayload := `{
		"driver": "sqlite",
		"schema": "CREATE TABLE items (id INT PRIMARY KEY, qty INT);",
		"seed": "INSERT INTO items VALUES (1, 100);",
		"invariants": [
			{
				"name": "positive_qty",
				"query": "SELECT qty FROM items WHERE id = 1;",
				"assert": "qty > 0"
			}
		],
		"operations": [
			{
				"name": "noop",
				"steps": [
					{"sql": "SELECT qty FROM items WHERE id = 1"}
				]
			}
		],
		"workers": 2,
		"iterations": 5,
		"seed_value": 42
	}`

	cmd := newRootCmd()
	inBuf := bytes.NewBufferString(inputPayload)
	outBuf := new(bytes.Buffer)
	cmd.SetIn(inBuf)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"engine"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("engine command failed: %v", err)
	}

	var resp IPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nOutput:\n%s", err, outBuf.String())
	}

	if !resp.Success {
		t.Fatalf("expected success: true, got: false, err: %s", resp.Error)
	}
	if resp.ViolationDetected {
		t.Fatalf("expected no violation detected")
	}
	if resp.Seed != 42 || resp.Schedule.Seed != 42 || len(resp.Schedule.Decisions) != 5 {
		t.Fatalf("expected effective seed and logical schedule in response: %+v", resp)
	}
	if resp.ReproPython == "" || resp.ReproTypeScript == "" || resp.ReproGo == "" {
		t.Errorf("expected repro code templates in response")
	}
}

func TestEngineCmd_ExecutionErrorStatus(t *testing.T) {
	inputPayload := `{
		"driver": "sqlite",
		"isolation": "SERIALIZABLE",
		"schema": "CREATE TABLE items (id INT PRIMARY KEY, qty INT);",
		"seed": "INSERT INTO items VALUES (1, 100);",
		"invariants": [{"name": "must_not_run", "query": "SELECT value FROM missing_table", "assert": "value == 1"}],
		"operations": [{"name": "invalid", "steps": [{"sql": "NOT VALID SQL"}]}],
		"workers": 1,
		"iterations": 1,
		"seed_value": 42
	}`

	cmd := newRootCmd()
	outBuf := new(bytes.Buffer)
	cmd.SetIn(bytes.NewBufferString(inputPayload))
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"engine"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("engine command failed: %v", err)
	}

	var resp IPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v\n%s", err, outBuf.String())
	}
	if resp.Status != domain.StatusExecutionError || resp.Success || resp.ViolationDetected {
		t.Fatalf("unexpected status response: %+v", resp)
	}
	if resp.Isolation != domain.LevelSerializable || len(resp.OperationErrors) != 1 || resp.FailingInvariant != nil {
		t.Fatalf("unexpected execution diagnostics: %+v", resp)
	}
}

func TestUnreliableRunError(t *testing.T) {
	if err := unreliableRunError(&domain.ExecutionResult{Status: domain.StatusPassed}); err != nil {
		t.Fatalf("passed result returned error: %v", err)
	}
	if err := unreliableRunError(&domain.ExecutionResult{Status: domain.StatusViolation}); err != nil {
		t.Fatalf("violation result returned execution error: %v", err)
	}
	if err := unreliableRunError(&domain.ExecutionResult{Status: domain.StatusInconclusive}); err == nil {
		t.Fatal("inconclusive result returned nil error")
	}
}

func TestPreserveRunResult_KeepsCanceledResultForReporting(t *testing.T) {
	result := &domain.ExecutionResult{Status: domain.StatusCanceled, Error: context.Canceled}
	got, err := preserveRunResult(result, context.Canceled)
	if err != nil || got != result {
		t.Fatalf("got result=%p err=%v, want preserved result=%p", got, err, result)
	}
}

func TestCanceledIPCResponsePreservesRunMetadata(t *testing.T) {
	result := &domain.ExecutionResult{
		Status:    domain.StatusViolation,
		Isolation: domain.LevelSerializable,
		Seed:      42,
		Schedule:  domain.SchedulePlan{Version: 1, Seed: 42, Workers: 2},
		Trace:     domain.ExecutionTrace{{Type: domain.EventBegin}},
	}
	response := canceledIPCResponse(result, context.DeadlineExceeded)
	if response.Status != domain.StatusCanceled || response.Success || response.ViolationDetected {
		t.Fatalf("unexpected cancellation status: %+v", response)
	}
	if response.Isolation != result.Isolation || response.Seed != 42 || response.Schedule.Version != 1 || response.TraceEventsCount != 1 {
		t.Fatalf("cancellation lost run metadata: %+v", response)
	}
}

func TestResolveEffectiveSeed_ExplicitZeroOverridesSpec(t *testing.T) {
	if got := resolveEffectiveSeed(99, 0, true); got != 0 {
		t.Fatalf("explicit seed zero resolved to %d", got)
	}
	if got := resolveEffectiveSeed(99, 0, false); got != 99 {
		t.Fatalf("omitted seed changed spec seed to %d", got)
	}
}

func TestEngineCmd_FailingInvariant(t *testing.T) {
	inputPayload := `{
		"driver": "sqlite",
		"schema": "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
		"seed": "INSERT INTO accounts VALUES (1, 1000);",
		"invariants": [
			{
				"name": "balance_preserved",
				"query": "SELECT balance FROM accounts WHERE id = 1;",
				"assert": "balance == 700"
			}
		],
		"operations": [
			{
				"name": "withdraw",
				"steps": [
					{"sql": "SELECT balance FROM accounts WHERE id = 1 -> cur"},
					{"sql": "UPDATE accounts SET balance = {cur - 100} WHERE id = 1"}
				]
			}
		],
		"workers": 1,
		"iterations": 2,
		"seed_value": 42
	}`

	cmd := newRootCmd()
	inBuf := bytes.NewBufferString(inputPayload)
	outBuf := new(bytes.Buffer)
	cmd.SetIn(inBuf)
	cmd.SetOut(outBuf)
	cmd.SetArgs([]string{"engine"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("engine command failed: %v", err)
	}

	var resp IPCResponse
	if err := json.Unmarshal(outBuf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode JSON response: %v\nOutput:\n%s", err, outBuf.String())
	}

	if resp.Success {
		t.Fatalf("expected success: false on anomaly detection")
	}
	if !resp.ViolationDetected {
		t.Fatalf("expected violation_detected: true")
	}
	if resp.FailingInvariant == nil || resp.FailingInvariant.Name != "balance_preserved" {
		t.Fatalf("expected failing invariant 'balance_preserved'")
	}
	if len(resp.MinimalOperations) == 0 {
		t.Fatalf("expected minimal_operations after shrinking")
	}
	if !strings.Contains(resp.ReproPython, "balance_preserved") {
		t.Fatalf("expected python repro to reference failing invariant")
	}
	if !strings.Contains(resp.ReproTypeScript, "balance_preserved") {
		t.Fatalf("expected typescript repro to reference failing invariant")
	}
}
