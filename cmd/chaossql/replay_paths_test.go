package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestValidateReplayArtifact_RejectsInconsistentArtifacts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*ReplayPayload)
		want   string
	}{
		{"missing spec", func(p *ReplayPayload) { p.Spec = nil }, "missing the complete specification"},
		{"invalid spec", func(p *ReplayPayload) { spec := *p.Spec; spec.Operations = nil; p.Spec = &spec }, "invalid replay specification"},
		{"no operations", func(p *ReplayPayload) { p.ScheduledOps = nil }, "no scheduled operations"},
		{"schedule version", func(p *ReplayPayload) { p.Schedule.Version = 7 }, "logical schedule version"},
		{"status mismatch", func(p *ReplayPayload) { p.Status = domain.StatusPassed }, "status does not match"},
		{"missing failing invariant", func(p *ReplayPayload) { p.FailingInvariant = nil }, "stored failing invariant"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifact := replayViolationFixture(t)
			tt.mutate(&artifact)
			err := validateReplayArtifact(artifact)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestVerifyReplayArtifact_ReportsDriverFailures(t *testing.T) {
	artifact := replayViolationFixture(t)
	spec := *artifact.Spec
	spec.Database.Driver = "postgres"
	spec.Database.DSN = "postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=2"
	artifact.Spec = &spec
	if _, err := verifyReplayArtifact(context.Background(), artifact); err == nil || !strings.Contains(err.Error(), "open replay database driver") {
		t.Fatalf("err = %v, want an open failure", err)
	}
}

func TestReplayCmd_InputFormatsAndErrors(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) (string, error) {
		cmd := newReplayCmd()
		out := new(bytes.Buffer)
		cmd.SetOut(out)
		cmd.SetErr(out)
		cmd.SetArgs(args)
		cmd.SilenceUsage, cmd.SilenceErrors = true, true
		err := cmd.Execute()
		return out.String(), err
	}

	traceArray := filepath.Join(dir, "trace.json")
	if err := os.WriteFile(traceArray, []byte(`[{"timestamp_us":10,"worker_id":1,"op_index":1,"op_name":"bump","type":"EXEC","sql":"UPDATE t SET v = 1"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := run(traceArray, "--max-events", "1"); err != nil || !strings.Contains(out, "UPDATE t SET v = 1") {
		t.Fatalf("trace array: err=%v\n%s", err, out)
	}

	garbage := filepath.Join(dir, "garbage.json")
	if err := os.WriteFile(garbage, []byte(`not json`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(garbage); err == nil || !strings.Contains(err.Error(), "invalid json trace format") {
		t.Fatalf("garbage: err = %v", err)
	}
	if _, err := run(filepath.Join(dir, "missing.json")); err == nil {
		t.Fatal("missing file must fail")
	}
	if _, err := run(traceArray, "--verify"); err == nil || !strings.Contains(err.Error(), "replay verification failed") {
		t.Fatalf("verifying a bare trace must fail, got %v", err)
	}
}

func TestWriteRestrictedAtomic_ReportsUnwritableDirectory(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeRestrictedAtomic(filepath.Join(blocker, "artifact.json"), []byte("{}")); err == nil {
		t.Fatal("writing below a regular file must fail")
	}
}
