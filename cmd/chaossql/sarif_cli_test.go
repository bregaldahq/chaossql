package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestCLI_ExportSARIFFlag_Registration(t *testing.T) {
	runCmd := newRunCmd()
	fRun := runCmd.Flags().Lookup("export-sarif")
	if fRun == nil {
		t.Fatalf("expected --export-sarif flag to be registered on runCmd")
	}

	demoCmd := newDemoCmd()
	fDemo := demoCmd.Flags().Lookup("export-sarif")
	if fDemo == nil {
		t.Fatalf("expected --export-sarif flag to be registered on demoCmd")
	}
}

func TestCLI_ExportSARIF_RunCmd(t *testing.T) {
	tmpDir := t.TempDir()
	sarifOut := filepath.Join(tmpDir, "run_sarif.json")

	specPath := filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml")

	cmd := newRunCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{specPath, "--export-sarif", sarifOut, "--workers", "2", "--iterations", "10"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("runCmd execution failed: %v", err)
	}

	data, err := os.ReadFile(sarifOut)
	if err != nil {
		t.Fatalf("expected sarif output file to exist: %v", err)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("expected valid SARIF JSON: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %q", report.Version)
	}
	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
}

func TestCLI_ExportSARIF_DemoCmd(t *testing.T) {
	tmpDir := t.TempDir()
	sarifOut := filepath.Join(tmpDir, "demo_sarif.json")

	cmd := newDemoCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{"deadlock", "--export-sarif", sarifOut})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("demoCmd execution failed: %v", err)
	}

	data, err := os.ReadFile(sarifOut)
	if err != nil {
		t.Fatalf("expected sarif output file to exist: %v", err)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("expected valid SARIF JSON: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %q", report.Version)
	}
	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
}
