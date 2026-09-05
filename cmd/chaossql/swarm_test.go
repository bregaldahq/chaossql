package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestSwarmCmd_DiffMarkdownSummaryFlag(t *testing.T) {
	scenarioDir := filepath.Join("..", "..", "examples", "banking_lost_update")
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "swarm_summary.md")

	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{
		"diff",
		"--scenarios-dir", scenarioDir,
		"--drivers", "sqlite,mock",
		"--markdown-summary", summaryFile,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing swarm diff with markdown summary: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read generated markdown summary: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit") {
		t.Errorf("markdown summary missing expected header, got:\n%s", content)
	}
	if !strings.Contains(content, "Total Scenarios") {
		t.Errorf("markdown summary missing Total Scenarios, got:\n%s", content)
	}
}

func TestSwarmCmd_RunMarkdownSummaryFlag(t *testing.T) {
	scenarioDir := filepath.Join("..", "..", "examples", "banking_lost_update")
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "subdir", "summary.md")

	cmd := newSwarmCmd()
	b := new(bytes.Buffer)
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{
		"run",
		"--scenarios-dir", scenarioDir,
		"--drivers", "sqlite",
		"--json",
		"--markdown-summary", summaryFile,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing swarm run with markdown summary: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read generated markdown summary: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit") {
		t.Errorf("markdown summary missing expected header, got:\n%s", content)
	}
}
