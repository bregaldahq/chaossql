package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/bregaldahq/chaossql/internal/swarm"
)

func TestSwarmCmd_DiffTerminalOutput(t *testing.T) {
	scenarioDir := filepath.Join("..", "..", "examples", "banking_lost_update")

	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"diff", "--scenarios-dir", scenarioDir, "--drivers", "sqlite,mock", "--concurrency", "2"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing swarm diff: %v", err)
	}

	out := b.String()
	if len(out) == 0 {
		t.Fatal("expected non-empty terminal output")
	}
	if !bytes.Contains(b.Bytes(), []byte("AUTONOMOUS MULTI-ENGINE DIFFERENTIAL SWARM REPORT")) {
		t.Errorf("expected banner in output, got: %s", out)
	}
}

func TestSwarmCmd_RunJSONOutput(t *testing.T) {
	scenarioDir := filepath.Join("..", "..", "examples", "banking_lost_update")

	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"run", "--scenarios-dir", scenarioDir, "--drivers", "sqlite,mock", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing swarm run with --json: %v", err)
	}

	var report swarm.DifferentialReport
	if err := json.Unmarshal(b.Bytes(), &report); err != nil {
		t.Fatalf("failed to decode JSON output: %v, raw output: %s", err, b.String())
	}

	if report.TotalScenarios != 1 {
		t.Errorf("expected 1 scenario in report, got %d", report.TotalScenarios)
	}
	if report.TotalExecutions != 2 {
		t.Errorf("expected 2 driver executions, got %d", report.TotalExecutions)
	}
	if len(report.Scenarios) != 1 {
		t.Fatalf("expected 1 scenario differential, got %d", len(report.Scenarios))
	}
	sc := report.Scenarios[0]
	if _, ok := sc.Results["sqlite"]; !ok {
		t.Errorf("missing sqlite result")
	}
	if _, ok := sc.Results["mock"]; !ok {
		t.Errorf("missing mock result")
	}
}

func TestSwarmCmd_RootDirectInvocation(t *testing.T) {
	scenarioDir := filepath.Join("..", "..", "examples", "banking_lost_update")

	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"--scenarios-dir", scenarioDir, "--drivers", "sqlite", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing root swarm command: %v", err)
	}

	var report swarm.DifferentialReport
	if err := json.Unmarshal(b.Bytes(), &report); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if report.TotalScenarios != 1 {
		t.Errorf("expected 1 scenario, got %d", report.TotalScenarios)
	}
}

func TestSwarmCmd_InvalidScenariosDir(t *testing.T) {
	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"diff", "--scenarios-dir", "./nonexistent_directory_xyz123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent scenarios directory, got nil")
	}
}
