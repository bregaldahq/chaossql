package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
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

func TestValidateScenarioSpec_ReportsEveryIssue(t *testing.T) {
	spec := &domain.Spec{
		Database: domain.DatabaseConfig{Driver: "oracle"},
		Invariants: []domain.InvariantConfig{
			{},
			{Name: "bad_expr", Query: "SELECT 1", Assert: "balance >"},
		},
		Operations: []domain.OperationConfig{{}},
	}
	got := map[string]string{}
	for _, issue := range validateScenarioSpec(spec) {
		got[issue.Field+"|"+issue.Level] = issue.Message
	}
	for _, want := range []string{
		"database.driver|WARNING",
		"database.schema|ERROR",
		"database.seed|WARNING",
		"invariants[0].name|ERROR",
		"invariants[0].query|ERROR",
		"invariants[0].assert|ERROR",
		"invariants[1].assert|ERROR",
		"operations[0].name|ERROR",
		"operations[0].steps|ERROR",
	} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing issue %s; got %v", want, got)
		}
	}
}

func TestValidateCmd_FailsOnErrorsAndMissingFiles(t *testing.T) {
	dir := t.TempDir()
	invalid := filepath.Join(dir, "chaos.yaml")
	if err := os.WriteFile(invalid, []byte(`version: "1.1"
name: broken
database:
  driver: sqlite
engine:
  workers: 1
  iterations: 1
invariants:
  - name: bad
    query: "SELECT 1 AS one"
    assert: "one >"
operations:
  - name: noop
    steps:
      - sql: "SELECT 1"
`), 0o644); err != nil {
		t.Fatal(err)
	}

	for name, args := range map[string][]string{
		"invalid spec": {invalid},
		"missing file": {filepath.Join(dir, "missing.yaml")},
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newValidateCmd()
			out := new(bytes.Buffer)
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("expected validation to fail; output:\n%s", out)
			}
		})
	}
}
