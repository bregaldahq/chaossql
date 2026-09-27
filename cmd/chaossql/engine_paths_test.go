package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func runEngine(t *testing.T, input string) ([]IPCResponse, error) {
	t.Helper()
	cmd := newRootCmd()
	out := new(bytes.Buffer)
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(out)
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"engine"})
	cmd.SilenceUsage, cmd.SilenceErrors = true, true
	err := cmd.Execute()
	var responses []IPCResponse
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<24)
	for scanner.Scan() {
		var resp IPCResponse
		if jsonErr := json.Unmarshal(scanner.Bytes(), &resp); jsonErr != nil {
			t.Fatalf("invalid response line %q: %v", scanner.Text(), jsonErr)
		}
		responses = append(responses, resp)
	}
	return responses, err
}

// Nested database/engine blocks override root fields, inline "->" and "=>"
// captures are honoured, and the lost update is detected and shrunk.
func TestEngineCmd_NestedConfigAndInlineCaptures(t *testing.T) {
	payload := `{
		"name": "nested_lost_update",
		"version": "1.1",
		"driver": "mock",
		"database": {
			"driver": "sqlite",
			"dsn": ":memory:",
			"isolation": "READ_UNCOMMITTED",
			"schema": "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			"seed": "INSERT INTO accounts VALUES (1, 100);"
		},
		"engine": {"workers": 1, "iterations": 6, "seed": 11, "jitter_ms": [0, 1]},
		"invariants": [{"name": "balance_preserved", "query": "SELECT balance FROM accounts WHERE id = 1", "assert": "balance == 100"}],
		"operations": [
			{"name": "withdraw", "weight": 0, "steps": [
				{"sql": "SELECT balance FROM accounts WHERE id = 1 -> bal"},
				{"sql": "UPDATE accounts SET balance = {bal} - 1 WHERE id = 1"}
			]},
			{"name": "audit", "weight": 2, "steps": [
				{"sql": "SELECT balance FROM accounts WHERE id = 1 => seen"}
			]}
		]
	}`
	responses, err := runEngine(t, payload)
	if err != nil {
		t.Fatalf("engine failed: %v", err)
	}
	if len(responses) != 1 {
		t.Fatalf("got %d responses, want 1", len(responses))
	}
	resp := responses[0]
	if resp.Seed != 11 || resp.Isolation != domain.LevelReadUncommitted {
		t.Fatalf("seed/isolation = %d/%q, want nested engine and database settings", resp.Seed, resp.Isolation)
	}
	if !resp.ViolationDetected || resp.FailingInvariant == nil || resp.FailingInvariant.Name != "balance_preserved" {
		t.Fatalf("response = %+v, want a balance_preserved violation", resp)
	}
	if resp.Shrink == nil || len(resp.MinimalOperations) == 0 || resp.ReproGo == "" {
		t.Fatalf("expected a shrunk reproduction, got shrink=%+v ops=%d", resp.Shrink, len(resp.MinimalOperations))
	}
	for _, op := range resp.MinimalOperations {
		for _, step := range op.Steps {
			if strings.Contains(step.SQL, "->") || strings.Contains(step.SQL, "=>") {
				t.Fatalf("inline capture was not stripped from %q", step.SQL)
			}
		}
	}
}

func TestEngineCmd_StreamsOneResponsePerPayload(t *testing.T) {
	one := `{"schema":"CREATE TABLE t (id INT);","invariants":[],"operations":[{"name":"noop","steps":[{"sql":"SELECT 1"}]}],"workers":1,"iterations":1}`
	responses, err := runEngine(t, one+"\n"+one)
	if err != nil {
		t.Fatal(err)
	}
	if len(responses) != 2 || !responses[0].Success || !responses[1].Success {
		t.Fatalf("responses = %+v, want two successful runs", responses)
	}
}

func TestEngineCmd_ReportsInvalidInputs(t *testing.T) {
	responses, err := runEngine(t, `{"driver": `)
	if err == nil || len(responses) != 1 || !strings.Contains(responses[0].Error, "JSON decoding error") {
		t.Fatalf("malformed JSON: err=%v responses=%+v", err, responses)
	}

	responses, err = runEngine(t, `{"driver":"oracle","invariants":[],"operations":[]}`)
	if err != nil || len(responses) != 1 || responses[0].Status != domain.StatusExecutionError || !strings.Contains(responses[0].Error, "driver") {
		t.Fatalf("unsupported driver: err=%v responses=%+v", err, responses)
	}

	responses, err = runEngine(t, `{"driver":"postgres","dsn":"postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=2","invariants":[],"operations":[]}`)
	if err != nil || len(responses) != 1 || !strings.Contains(responses[0].Error, "failed to open database driver") {
		t.Fatalf("unreachable database: err=%v responses=%+v", err, responses)
	}
}

func TestExecuteIPCPayload_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resp := executeIPCPayload(ctx, IPCPayload{
		Schema:     "CREATE TABLE t (id INT);",
		Operations: []IPCOperation{{Name: "noop", Steps: []IPCOperationStep{{SQL: "SELECT 1"}}}},
	})
	if resp.Success {
		t.Fatalf("a cancelled run must not report success: %+v", resp)
	}
}
