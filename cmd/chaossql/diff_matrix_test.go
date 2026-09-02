package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestDiffCmd_Execution(t *testing.T) {
	cmd := newDiffCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)

	specPath := filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml")
	cmd.SetArgs([]string{specPath, "--driver-a", "sqlite", "--driver-b", "sqlite", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing diffCmd: %v", err)
	}

	out := b.String()
	if len(out) == 0 {
		t.Errorf("expected non-empty JSON output from diff command")
	}
}

func TestMatrixCmd_Execution(t *testing.T) {
	cmd := newMatrixCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)

	cmd.SetArgs([]string{"--driver", "sqlite", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing matrixCmd: %v", err)
	}

	out := b.String()
	if len(out) == 0 {
		t.Errorf("expected non-empty JSON output from matrix command")
	}
}
