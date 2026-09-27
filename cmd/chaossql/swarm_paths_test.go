package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/swarm"
	"github.com/spf13/cobra"
)

const minimalSpecYAML = `version: "1.1"
name: %s
database:
  driver: sqlite
  schema: "CREATE TABLE t (id INT PRIMARY KEY, v INT);"
  seed: "INSERT INTO t VALUES (1, 0);"
engine:
  workers: 1
  iterations: 2
invariants:
  - name: non_negative
    query: "SELECT v FROM t WHERE id = 1"
    assert: "v >= 0"
operations:
  - name: bump
    steps:
      - sql: "UPDATE t SET v = v + 1 WHERE id = 1"
`

func writeSpec(t *testing.T, path, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Replace(minimalSpecYAML, "%s", name, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDiscoverAndLoadSpecs(t *testing.T) {
	dir := t.TempDir()

	single := filepath.Join(dir, "single", "custom.yaml")
	writeSpec(t, single, "single_file")
	specs, err := discoverAndLoadSpecs(single)
	if err != nil || len(specs) != 1 || specs[0].Name != "single_file" {
		t.Fatalf("single file: specs=%v err=%v", specs, err)
	}

	// Without chaos.yaml or variant_*.yaml, any YAML file in the tree is used.
	fallback := filepath.Join(dir, "fallback")
	writeSpec(t, filepath.Join(fallback, "nested", "scenario.yml"), "fallback_yaml")
	if err := os.WriteFile(filepath.Join(fallback, "notes.yaml"), []byte("not: [a, spec"), 0o644); err != nil {
		t.Fatal(err)
	}
	specs, err = discoverAndLoadSpecs(fallback)
	if err != nil || len(specs) != 1 || specs[0].Name != "fallback_yaml" {
		t.Fatalf("fallback: specs=%v err=%v", specs, err)
	}

	empty := filepath.Join(dir, "empty")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := discoverAndLoadSpecs(empty); err == nil || !strings.Contains(err.Error(), "no valid ChaosSQL scenario") {
		t.Fatalf("empty dir: err = %v", err)
	}
	if _, err := discoverAndLoadSpecs(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing path must fail")
	}
}

func TestSwarmCmd_PositionalDirAndErrors(t *testing.T) {
	dir := t.TempDir()
	writeSpec(t, filepath.Join(dir, "chaos.yaml"), "positional")
	run := func(args ...string) (string, error) {
		cmd := newSwarmCmd()
		out := new(bytes.Buffer)
		cmd.SetOut(out)
		cmd.SetErr(out)
		cmd.SetArgs(args)
		cmd.SilenceUsage, cmd.SilenceErrors = true, true
		err := cmd.Execute()
		return out.String(), err
	}

	out, err := run("run", dir, "--drivers", " , ", "--concurrency", "0")
	if err != nil || !strings.Contains(out, "positional") {
		t.Fatalf("positional dir with default drivers: err=%v\n%s", err, out)
	}
	if _, err := run("run", filepath.Join(dir, "does-not-exist")); err == nil {
		t.Fatal("missing scenarios directory must fail")
	}
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run("run", dir, "--markdown-summary", filepath.Join(blocker, "summary.md")); err == nil {
		t.Fatal("an unwritable markdown summary path must fail")
	}
}

func TestRenderSwarmTerminal_Outcomes(t *testing.T) {
	render := func(report *swarm.DifferentialReport) string {
		cmd := &cobra.Command{}
		out := new(bytes.Buffer)
		cmd.SetOut(out)
		renderSwarmTerminal(cmd, report, []string{"sqlite", "mock"})
		return out.String()
	}

	failed := render(&swarm.DifferentialReport{Scenarios: []swarm.ScenarioDifferential{{
		ScenarioName: "offline", Summary: "All drivers encountered execution or connection errors",
		Results: map[string]swarm.DriverExecutionResult{"sqlite": {Error: "boom"}, "mock": {Error: "boom"}},
	}}})
	for _, want := range []string{"ALL TARGET ENGINES FAILED", "ERROR", "sqlite: ERR(boom)", "All drivers encountered"} {
		if !strings.Contains(failed, want) {
			t.Errorf("failed report missing %q:\n%s", want, failed)
		}
	}

	divergent := render(&swarm.DifferentialReport{DivergentCount: 1, Scenarios: []swarm.ScenarioDifferential{
		{
			ScenarioName: strings.Repeat("very_long_scenario_name_", 3), Divergent: true, Summary: "engines disagree",
			Results: map[string]swarm.DriverExecutionResult{"sqlite": {ViolationDetected: true, FailingInvariant: "balance_preserved"}, "mock": {}},
		},
		{
			ScenarioName: "half_offline", Summary: "Single active driver mock",
			Results: map[string]swarm.DriverExecutionResult{"sqlite": {Error: "offline"}, "mock": {}},
		},
	}})
	for _, want := range []string{"1 DIVERGENCE(S) DETECTED", "DIVERGENT", "VIOLATION(balance_preserved)", "INCONCLUSIVE", "...", "engines disagree"} {
		if !strings.Contains(divergent, want) {
			t.Errorf("divergent report missing %q:\n%s", want, divergent)
		}
	}
}

func TestTruncateString(t *testing.T) {
	cases := []struct {
		in   string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"abcdefghij", 6, "abc..."},
		{"abcdef", 2, "ab"},
		{"ΔΣΘΛΞΠΦ", 5, "ΔΣ..."},
	}
	for _, tc := range cases {
		if got := truncateString(tc.in, tc.max); got != tc.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}
