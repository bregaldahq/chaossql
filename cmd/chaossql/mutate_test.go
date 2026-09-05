package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMutateCmd_Execution(t *testing.T) {
	validPath := filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml")
	tmpDir := t.TempDir()

	cmd := newMutateCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{validPath, "--variants", "3", "--output-dir", tmpDir, "--seed", "123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing mutate: %v", err)
	}

	// Verify generated variants on disk
	for i := 0; i < 3; i++ {
		varFile := filepath.Join(tmpDir, "variant_"+string(rune('0'+i))+".yaml")
		info, err := os.Stat(varFile)
		if err != nil {
			t.Fatalf("variant file %s not found: %v", varFile, err)
		}
		if info.Size() == 0 {
			t.Fatalf("variant file %s is empty", varFile)
		}
	}

	// Verify schema.sql and seed.sql were copied to output dir
	schemaCopy := filepath.Join(tmpDir, "schema.sql")
	if _, err := os.Stat(schemaCopy); err != nil {
		t.Errorf("schema.sql was not copied to output dir: %v", err)
	}
	seedCopy := filepath.Join(tmpDir, "seed.sql")
	if _, err := os.Stat(seedCopy); err != nil {
		t.Errorf("seed.sql was not copied to output dir: %v", err)
	}
}

func TestMutateCmd_JSONOutput(t *testing.T) {
	validPath := filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml")
	tmpDir := t.TempDir()

	cmd := newMutateCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{validPath, "--variants", "2", "--output-dir", tmpDir, "--seed", "42", "--json"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing mutate with --json: %v", err)
	}

	var summary MutateSummary
	if err := json.Unmarshal(b.Bytes(), &summary); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v, output was: %s", err, b.String())
	}

	if summary.VariantsCount != 2 {
		t.Errorf("expected 2 variants in summary, got %d", summary.VariantsCount)
	}
	if summary.OutputDir != tmpDir {
		t.Errorf("expected output dir %s, got %s", tmpDir, summary.OutputDir)
	}
	if len(summary.Variants) != 2 {
		t.Errorf("expected 2 variant entries, got %d", len(summary.Variants))
	}
}