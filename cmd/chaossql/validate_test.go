package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestValidateCmd_ValidScenario(t *testing.T) {
	validPath := filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml")

	cmd := newValidateCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{validPath})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error validating valid scenario: %v", err)
	}

	out := b.String()
	if len(out) == 0 {
		t.Errorf("expected non-empty output from validate command")
	}
}
