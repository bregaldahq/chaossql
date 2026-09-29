package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/examples"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

type firstRunOutput struct {
	Status      domain.ExecutionStatus `json:"status"`
	AnomalyType domain.AnomalyType     `json:"anomaly_type"`
	Isolation   domain.IsolationLevel  `json:"isolation"`
	Shrink      *struct {
		MinimalOps []json.RawMessage `json:"minimal_ops"`
	} `json:"shrink"`
}

func executeJSON(t *testing.T, cmd interface {
	SetArgs([]string)
	Execute() error
}, args ...string) firstRunOutput {
	t.Helper()
	cmd.SetArgs(append(args, "--json"))
	var runErr error
	raw := captureStdout(t, func() { runErr = cmd.Execute() })
	if runErr != nil {
		t.Fatalf("command failed: %v\n%s", runErr, raw)
	}
	var out firstRunOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, raw)
	}
	return out
}

func assertFindsLostUpdate(t *testing.T, out firstRunOutput) {
	t.Helper()
	if out.Status != domain.StatusViolation || out.AnomalyType != domain.AnomalyLostUpdate {
		t.Fatalf("expected a P4 lost update, got status %q anomaly %q at %q", out.Status, out.AnomalyType, out.Isolation)
	}
	if out.Shrink == nil || len(out.Shrink.MinimalOps) != 2 {
		t.Fatalf("expected the failure shrunk to 2 operations, got %+v", out.Shrink)
	}
}

// The README quickstart must show the bug it promises; a clean run here means
// a new user sees "all invariants satisfied" and leaves.
func TestQuickstartFindsLostUpdate(t *testing.T) {
	assertFindsLostUpdate(t, executeJSON(t, newRunCmd(), bankingSpecPath(t)))
}

// An installed binary has no examples directory: demo must use the embedded copy.
func TestDemoFindsLostUpdateOutsideTheRepository(t *testing.T) {
	t.Chdir(t.TempDir())
	assertFindsLostUpdate(t, executeJSON(t, newDemoCmd(), "banking"))
}

func TestEveryDemoAliasLoadsFromTheEmbeddedExamples(t *testing.T) {
	for _, alias := range []string{"banking", "inventory", "hospital", "financial", "auction", "crypto", "flash_crash", "ticket", "deadlock", "fk"} {
		path, err := resolveDemoPath(alias)
		if err != nil {
			t.Fatal(err)
		}
		spec, err := domain.LoadSpecFS(examples.FS, strings.TrimPrefix(path, "examples/"))
		if err != nil {
			t.Fatalf("%s: %v", alias, err)
		}
		if strings.HasSuffix(spec.Database.Schema, ".sql") || strings.HasSuffix(spec.Database.Seed, ".sql") {
			t.Fatalf("%s: schema or seed file was not resolved", alias)
		}
	}
}

func TestInitScaffoldFindsLostUpdate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "scenario")
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{dir})
	captureStdout(t, func() {
		if err := initCmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	assertFindsLostUpdate(t, executeJSON(t, newRunCmd(), filepath.Join(dir, "chaos.yaml")))
}

func TestIsolationFlagOverridesSpec(t *testing.T) {
	out := executeJSON(t, newRunCmd(), bankingSpecPath(t), "--isolation", "serializable")
	if out.Status != domain.StatusPassed || out.Isolation != domain.LevelSerializable {
		t.Fatalf("expected a serialized, passing run, got status %q at %q", out.Status, out.Isolation)
	}
}

func TestApplyDatabaseOverrides(t *testing.T) {
	defer func() { driverFlag, dsnFlag, isolationFlag = "", "", "" }()
	base := domain.Spec{Database: domain.DatabaseConfig{Driver: "sqlite", DSN: "file:x.db", Isolation: domain.LevelReadUncommitted}}

	driverFlag, dsnFlag, isolationFlag = "postgres", "", ""
	spec := base
	switched, err := applyDatabaseOverrides(&spec)
	if err != nil || !switched || spec.Database.Driver != "postgres" || spec.Database.DSN != "" {
		t.Fatalf("a new driver must drop the old DSN: switched=%v err=%v %+v", switched, err, spec.Database)
	}

	driverFlag, dsnFlag, isolationFlag = "sqlite", "file:y.db", "read-committed"
	spec = base
	switched, err = applyDatabaseOverrides(&spec)
	if err != nil || switched || spec.Database.DSN != "file:y.db" || spec.Database.Isolation != domain.LevelReadCommitted {
		t.Fatalf("unexpected overrides: switched=%v err=%v %+v", switched, err, spec.Database)
	}

	driverFlag, dsnFlag, isolationFlag = "", "", "snapshot"
	spec = base
	if _, err := applyDatabaseOverrides(&spec); err == nil {
		t.Fatal("an unknown isolation level must be rejected")
	}
}

func TestSerializedRunHint(t *testing.T) {
	serialized := &engine.RunResult{Status: domain.StatusPassed, Isolation: domain.LevelSerializable}
	if serializedRunHint("sqlite", serialized) == "" {
		t.Fatal("a clean SQLite SERIALIZABLE run must explain that nothing interleaved")
	}
	for name, tc := range map[string]struct {
		driver string
		result *engine.RunResult
	}{
		"violation":        {"sqlite", &engine.RunResult{Status: domain.StatusViolation, Isolation: domain.LevelSerializable}},
		"read uncommitted": {"sqlite", &engine.RunResult{Status: domain.StatusPassed, Isolation: domain.LevelReadUncommitted}},
		"postgres":         {"postgres", serialized},
		"no result":        {"sqlite", nil},
	} {
		if hint := serializedRunHint(tc.driver, tc.result); hint != "" {
			t.Errorf("%s: unexpected hint %q", name, hint)
		}
	}
}

func TestInitScaffoldSetsIsolationOnlyForSQLite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "pg")
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{dir, "--driver", "postgres"})
	captureStdout(t, func() {
		if err := initCmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})
	spec, err := domain.LoadSpec(filepath.Join(dir, "chaos.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if spec.Database.Driver != "postgres" || spec.Database.Isolation != "" {
		t.Fatalf("PostgreSQL shows the lost update at its default level; got %q at %q", spec.Database.Driver, spec.Database.Isolation)
	}
}

func TestRunRejectsUnknownIsolationFlag(t *testing.T) {
	cmd := newRunCmd()
	cmd.SetArgs([]string{bankingSpecPath(t), "--isolation", "snapshot"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	if err := cmd.Execute(); err == nil || !errors.Is(err, domain.ErrSpecValidationFailed) {
		t.Fatalf("expected a validation error for --isolation snapshot, got %v", err)
	}
}

// Switching the banking example (READ_UNCOMMITTED) to PostgreSQL must warn and
// fall back to PostgreSQL's default level before connecting.
func TestRunWarnsWhenSwitchedDriverLacksSpecIsolation(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = original }()

	cmd := newRunCmd()
	cmd.SetArgs([]string{bankingSpecPath(t), "--driver", "postgres", "--dsn", "postgres://chaossql:unused@127.0.0.1:1/none?sslmode=disable&connect_timeout=1"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	runErr := cmd.Execute()
	w.Close()
	os.Stderr = original
	warning, _ := io.ReadAll(r)

	if runErr == nil || !strings.Contains(runErr.Error(), "failed to initialize database driver") {
		t.Fatalf("expected the unreachable database to fail after the fallback, got %v", runErr)
	}
	if !strings.Contains(string(warning), "postgres does not support the spec isolation READ_UNCOMMITTED") {
		t.Fatalf("missing fallback warning, stderr was %q", warning)
	}
}
