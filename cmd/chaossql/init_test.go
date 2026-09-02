package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInitCmd_Execution(t *testing.T) {
	tmpDir := t.TempDir()
	targetScenario := filepath.Join(tmpDir, "test_custom_scenario")

	cmd := newInitCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)
	cmd.SetArgs([]string{targetScenario, "--name", "my_custom_test"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing initCmd: %v", err)
	}

	for _, f := range []string{"schema.sql", "seed.sql", "chaos.yaml", "README.md"} {
		p := filepath.Join(targetScenario, f)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created by init command, but not found", f)
		}
	}

	// Test force error when file exists without --force
	cmd2 := newInitCmd()
	b2 := bytes.NewBufferString("")
	cmd2.SetOut(b2)
	cmd2.SetErr(b2)
	cmd2.SetArgs([]string{targetScenario})
	if err := cmd2.Execute(); err == nil {
		t.Errorf("expected error when running init on existing directory without --force")
	}
}
