package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
)

func bankingSpecPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "examples", "banking_lost_update", "chaos.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunCmd_WritesEveryRequestedExport(t *testing.T) {
	specPath := bankingSpecPath(t)
	dir := t.TempDir()
	t.Chdir(dir) // repro_test.go and trace.mermaid are written to the working directory

	cmd := newRunCmd()
	cmd.SetArgs([]string{specPath, "--workers", "2", "--iterations", "20", "--seed", "42",
		"--export-repro", "--export-mermaid",
		"--export-html", "report.html", "--export-otel", "trace.otel.json",
		"--export-junit", "junit.xml", "--export-summary", "summary.md"})
	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Errorf("run failed: %v", err)
		}
	})

	for _, name := range []string{"repro_test.go", "trace.mermaid", "report.html", "trace.otel.json", "junit.xml", "summary.md"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil || info.Size() == 0 {
			t.Errorf("export %s missing or empty (err %v)", name, err)
		}
		if !strings.Contains(output, name) {
			t.Errorf("output does not announce %s", name)
		}
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "trace.otel.json")); !json.Valid(data) {
		t.Error("OpenTelemetry export is not valid JSON")
	}
}

func TestRunCmd_ReportsUnwritableExportPaths(t *testing.T) {
	specPath := bankingSpecPath(t)
	missingDir := filepath.Join(t.TempDir(), "missing")
	for _, flag := range []string{"--export-html", "--export-otel", "--export-junit", "--export-summary", "--export-sarif"} {
		t.Run(flag, func(t *testing.T) {
			cmd := newRunCmd()
			cmd.SetArgs([]string{specPath, "--workers", "1", "--iterations", "4", "--seed", "7", flag, filepath.Join(missingDir, "out")})
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			var err error
			captureStdout(t, func() { err = cmd.Execute() })
			if err == nil {
				t.Fatalf("%s into a missing directory must fail", flag)
			}
		})
	}
}

func TestRunCmd_JSONOutputDescribesTheRun(t *testing.T) {
	cmd := newRunCmd()
	cmd.SetArgs([]string{bankingSpecPath(t), "--workers", "2", "--iterations", "20", "--seed", "42", "--json"})
	var runErr error
	output := captureStdout(t, func() { runErr = cmd.Execute() })
	if runErr != nil {
		t.Fatalf("run failed: %v", runErr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, output)
	}
	for _, key := range []string{"spec", "status", "seed", "anomaly_type", "repro_go", "mermaid", "html_report", "otel_trace"} {
		if _, ok := payload[key]; !ok {
			t.Errorf("JSON output missing %q; keys: %v", key, keys(payload))
		}
	}
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestRunCmd_RejectsBadInputs(t *testing.T) {
	dir := t.TempDir()
	unsupported := filepath.Join(dir, "oracle.yaml")
	if err := os.WriteFile(unsupported, []byte(`version: "1.1"
name: oracle
database:
  driver: oracle
engine:
  workers: 1
  iterations: 1
operations:
  - name: noop
    steps:
      - sql: "SELECT 1"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, spec := range map[string]string{
		"missing spec":       filepath.Join(dir, "missing.yaml"),
		"unsupported driver": unsupported,
	} {
		t.Run(name, func(t *testing.T) {
			cmd := newRunCmd()
			cmd.SetArgs([]string{spec})
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			if err := cmd.Execute(); err == nil {
				t.Fatal("expected run to fail")
			}
		})
	}

	demo := newDemoCmd()
	demo.SetArgs([]string{"no-such-scenario"})
	demo.SilenceUsage, demo.SilenceErrors = true, true
	if err := demo.Execute(); err == nil {
		t.Fatal("unknown demo scenario must fail")
	}
}

func TestRunCmd_UIServesTraceUntilCancelled(t *testing.T) {
	if conn, err := http.Get("http://127.0.0.1:8090/health"); err == nil {
		conn.Body.Close()
		t.Skip("port 8090 is already in use")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := newRunCmd()
	cmd.SetArgs([]string{bankingSpecPath(t), "--workers", "1", "--iterations", "4", "--seed", "7", "--ui"})
	var runErr error
	captureStdout(t, func() {
		done := make(chan error, 1)
		go func() { done <- cmd.ExecuteContext(ctx) }()
		deadline := time.Now().Add(20 * time.Second)
		for {
			resp, err := http.Get("http://127.0.0.1:8090/")
			if err == nil {
				body := new(bytes.Buffer)
				_, _ = body.ReadFrom(resp.Body)
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK || !strings.Contains(body.String(), "<html") {
					t.Errorf("trace viewer returned %d", resp.StatusCode)
				}
				break
			}
			if time.Now().After(deadline) {
				t.Error("trace viewer never started")
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		cancel()
		runErr = <-done
	})
	if runErr != nil {
		t.Fatalf("run --ui returned %v after cancellation", runErr)
	}
}

func TestDominantAnomaly(t *testing.T) {
	cycle := func(types ...analyzer.DependencyType) analyzer.Cycle {
		var c analyzer.Cycle
		for _, dep := range types {
			c = append(c, analyzer.Edge{Type: dep})
		}
		return c
	}
	lostUpdate := cycle(analyzer.DepWW, analyzer.DepRW)
	writeSkew := cycle(analyzer.DepRW, analyzer.DepRW)
	unclassified := cycle(analyzer.DepWW, analyzer.DepWR)

	cases := []struct {
		name     string
		cycles   []analyzer.Cycle
		fallback domain.AnomalyType
		want     domain.AnomalyType
	}{
		{"no cycles keeps fallback", nil, domain.AnomalyWriteSkew, domain.AnomalyWriteSkew},
		{"definitive anomaly wins over lost update", []analyzer.Cycle{lostUpdate, writeSkew}, domain.AnomalyUnknown, domain.AnomalyWriteSkew},
		{"lost update replaces fallback", []analyzer.Cycle{lostUpdate}, domain.AnomalyG0DirtyWrite, domain.AnomalyLostUpdate},
		{"unknown fallback defers to first cycle", []analyzer.Cycle{unclassified}, domain.AnomalyUnknown, domain.AnomalyUnknown},
		{"known fallback survives unclassified cycles", []analyzer.Cycle{unclassified}, domain.AnomalyA5AReadSkew, domain.AnomalyA5AReadSkew},
	}
	for _, tc := range cases {
		if got := dominantAnomaly(tc.cycles, tc.fallback); got != tc.want {
			t.Errorf("%s: dominantAnomaly = %q, want %q", tc.name, got, tc.want)
		}
	}
}
