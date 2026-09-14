package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestReplayCmd_Execution(t *testing.T) {
	tmpDir := t.TempDir()
	traceFile := filepath.Join(tmpDir, "trace.json")

	dummyJSON := `{
		"trace": [
			{"timestamp_us": 1000, "worker_id": 1, "op_index": 1, "op_name": "transfer", "type": "BEGIN", "sql": "BEGIN"},
			{"timestamp_us": 2000, "worker_id": 1, "op_index": 1, "op_name": "transfer", "type": "EXEC", "sql": "UPDATE accounts SET balance = 500 WHERE id = 1"},
			{"timestamp_us": 3000, "worker_id": 1, "op_index": 1, "op_name": "transfer", "type": "COMMIT", "sql": "COMMIT"}
		]
	}`
	if err := os.WriteFile(traceFile, []byte(dummyJSON), 0644); err != nil {
		t.Fatalf("failed to write dummy trace file: %v", err)
	}

	cmd := newReplayCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{traceFile})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing replayCmd: %v", err)
	}

	out := b.String()
	if len(out) == 0 {
		t.Errorf("expected non-empty output from replay command")
	}
}

func TestReplayArtifactV1_RoundTrip(t *testing.T) {
	spec := domain.Spec{
		Version: "1.0",
		Name:    "replay_round_trip",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			Seed:   "INSERT INTO accounts VALUES (1, 100);",
		},
		Engine: domain.EngineConfig{Workers: 1, Iterations: 2, Seed: 77},
		Invariants: []domain.InvariantConfig{{
			Name: "balance_preserved", Query: "SELECT balance FROM accounts WHERE id = 1", Assert: "balance == 100",
		}},
		Operations: []domain.OperationConfig{{Name: "write", Steps: []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 0 WHERE id = 1"}}}},
	}
	ops := []domain.ScheduledOp{{ID: 1, Name: "write", Steps: spec.Operations[0].Steps}}
	result := &domain.ExecutionResult{
		Status: domain.StatusViolation, Seed: 77, ViolationDetected: true,
		FailingInvariant: &domain.InvariantResult{Name: "balance_preserved"},
	}

	payload, err := buildReplayArtifact(spec, result, ops, nil, domain.AnomalyLostUpdate)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "finding.json")
	if err := writeReplayArtifact(path, payload); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ReplayPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != 1 || decoded.Spec == nil || decoded.Spec.Database.Schema != spec.Database.Schema {
		t.Fatalf("incomplete replay artifact: %#v", decoded)
	}
	if decoded.FailureSignature.FailingInvariant != "balance_preserved" {
		t.Fatalf("unexpected failure signature: %#v", decoded.FailureSignature)
	}
	if !reflect.DeepEqual(decoded.ScheduledOps, ops) {
		t.Fatalf("scheduled operations changed: %#v", decoded.ScheduledOps)
	}
	if decoded.Schedule.Seed != 77 || len(decoded.Schedule.Decisions) != 1 {
		t.Fatalf("unexpected minimal schedule: %#v", decoded.Schedule)
	}
}

func TestRunCmd_HasExportResultFlag(t *testing.T) {
	flag := newRunCmd().Flags().Lookup("export-result")
	if flag == nil {
		t.Fatal("run command has no --export-result flag")
	}
}

func TestRunCmd_ExportsMinimalReplayArtifact(t *testing.T) {
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "chaos.yaml")
	artifactPath := filepath.Join(tmpDir, "finding.json")
	specYAML := `version: "1.0"
name: export_replay
database:
  driver: sqlite
  dsn: ":memory:"
  schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
  seed: "INSERT INTO accounts VALUES (1, 100);"
engine:
  workers: 1
  iterations: 1
  seed: 9
invariants:
  - name: balance_preserved
    type: sql
    query: "SELECT balance FROM accounts WHERE id = 1"
    assert: "balance == 100"
operations:
  - name: break_balance
    steps:
      - sql: "UPDATE accounts SET balance = 0 WHERE id = 1"
`
	if err := os.WriteFile(specPath, []byte(specYAML), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newRunCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{specPath, "--export-result", artifactPath})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatal(err)
	}
	var artifact ReplayPayload
	if err := json.Unmarshal(data, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.Version != 1 || len(artifact.ScheduledOps) != 1 || artifact.FailureSignature.FailingInvariant != "balance_preserved" {
		t.Fatalf("unexpected exported artifact: %#v", artifact)
	}
}

func TestVerifyReplayArtifact_ReexecutesSameFailure(t *testing.T) {
	artifact := replayViolationFixture(t)
	for run := 0; run < 20; run++ {
		result, err := verifyReplayArtifact(context.Background(), artifact)
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		if result.FailingInvariant == nil || result.FailingInvariant.Name != "balance_preserved" {
			t.Fatalf("run %d wrong replay failure: %#v", run, result.FailingInvariant)
		}
	}
}

func TestVerifyReplayArtifact_RejectsTampering(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ReplayPayload)
		want   string
	}{
		{name: "seed", mutate: func(p *ReplayPayload) { p.Seed++ }, want: "seed"},
		{name: "schedule", mutate: func(p *ReplayPayload) { p.Schedule.Decisions[0].WorkerID++ }, want: "schedule"},
		{name: "failure", mutate: func(p *ReplayPayload) { p.FailureSignature.FailingInvariant = "other" }, want: "failure"},
		{name: "version", mutate: func(p *ReplayPayload) { p.Version = 99 }, want: "version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifact := replayViolationFixture(t)
			tt.mutate(&artifact)
			_, err := verifyReplayArtifact(context.Background(), artifact)
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tt.want) {
				t.Fatalf("error = %v, want text %q", err, tt.want)
			}
		})
	}
}

func TestReplayCmd_Verify(t *testing.T) {
	artifact := replayViolationFixture(t)
	path := filepath.Join(t.TempDir(), "finding.json")
	if err := writeReplayArtifact(path, artifact); err != nil {
		t.Fatal(err)
	}
	cmd := newReplayCmd()
	output := new(bytes.Buffer)
	cmd.SetOut(output)
	cmd.SetErr(output)
	cmd.SetArgs([]string{path, "--verify"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "REPLAY VERIFIED") {
		t.Fatalf("missing verification output: %s", output.String())
	}
}

func replayViolationFixture(t *testing.T) ReplayPayload {
	t.Helper()
	spec := domain.Spec{
		Version: "1.0",
		Name:    "replay_violation",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			DSN:    ":memory:",
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			Seed:   "INSERT INTO accounts VALUES (1, 100);",
		},
		Engine: domain.EngineConfig{Workers: 1, Iterations: 1, Seed: 77},
		Invariants: []domain.InvariantConfig{{
			Name: "balance_preserved", Query: "SELECT balance FROM accounts WHERE id = 1", Assert: "balance == 100",
		}},
		Operations: []domain.OperationConfig{{Name: "break_balance", Steps: []domain.StepConfig{{SQL: "UPDATE accounts SET balance = 0 WHERE id = 1"}}}},
	}
	ops := []domain.ScheduledOp{{ID: 1, Name: "break_balance", Steps: spec.Operations[0].Steps}}
	result := &domain.ExecutionResult{
		Status: domain.StatusViolation, Seed: 77, ViolationDetected: true,
		FailingInvariant: &domain.InvariantResult{Name: "balance_preserved"},
	}
	artifact, err := buildReplayArtifact(spec, result, ops, nil, domain.AnomalyLostUpdate)
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}
