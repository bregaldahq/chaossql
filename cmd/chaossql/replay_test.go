package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
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
