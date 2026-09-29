package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/cloud"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestExitCode(t *testing.T) {
	finding := &findingError{msg: "violated"}
	for name, tc := range map[string]struct {
		err  error
		want int
	}{
		"success":                     {nil, exitOK},
		"finding":                     {finding, exitFinding},
		"wrapped finding":             {fmt.Errorf("run: %w", finding), exitFinding},
		"setup error":                 {errors.New("no such file"), exitTrouble},
		"finding joined with nil":     {errors.Join(finding, nil), exitFinding},
		"finding and a Cloud failure": {errors.Join(finding, errors.New("cloud 503")), exitTrouble},
	} {
		if got := exitCode(tc.err); got != tc.want {
			t.Errorf("%s: exitCode = %d, want %d", name, got, tc.want)
		}
	}
}

func TestCheckFailOn(t *testing.T) {
	for _, value := range []string{failOnViolation, failOnNever} {
		if err := checkFailOn(value, ""); err != nil {
			t.Errorf("%s: %v", value, err)
		}
	}
	if err := checkFailOn(failOnRegression, "token"); err != nil {
		t.Errorf("regression with a token: %v", err)
	}
	if err := checkFailOn(failOnRegression, ""); err == nil || !strings.Contains(err.Error(), "--cloud-token") {
		t.Errorf("regression without a token must ask for one, got %v", err)
	}
	if err := checkFailOn("always", ""); err == nil {
		t.Error("an unknown value must be rejected")
	}
}

func TestRunOutcomeError(t *testing.T) {
	violation := &engine.RunResult{Status: domain.StatusViolation, FailingInvariant: &domain.InvariantResult{Name: "ledger"}}
	passed := &engine.RunResult{Status: domain.StatusPassed}
	broken := &engine.RunResult{Status: domain.StatusExecutionError, Error: errors.New("deadlock")}
	regression := &cloud.RunIngestResponse{IsRegression: true}
	clean := &cloud.RunIngestResponse{}

	for name, tc := range map[string]struct {
		result *engine.RunResult
		failOn string
		cloud  *cloud.RunIngestResponse
		want   int
		text   string
	}{
		"violation fails":                {violation, failOnViolation, nil, exitFinding, `invariant "ledger" violated (P4_LOST_UPDATE)`},
		"pass succeeds":                  {passed, failOnViolation, nil, exitOK, ""},
		"never ignores a violation":      {violation, failOnNever, nil, exitOK, ""},
		"execution error always 2":       {broken, failOnNever, nil, exitTrouble, "execution_error"},
		"regression fails":               {violation, failOnRegression, regression, exitFinding, "regression"},
		"known violation, no regression": {violation, failOnRegression, clean, exitOK, ""},
		"regression without a verdict":   {passed, failOnRegression, nil, exitTrouble, "could not get a verdict"},
	} {
		err := runOutcomeError(tc.result, domain.AnomalyLostUpdate, tc.failOn, tc.cloud)
		if got := exitCode(err); got != tc.want {
			t.Errorf("%s: exit %d, want %d (err %v)", name, got, tc.want, err)
		}
		if tc.text != "" && (err == nil || !strings.Contains(err.Error(), tc.text)) {
			t.Errorf("%s: error %v does not mention %q", name, err, tc.text)
		}
	}

	unnamed := runOutcomeError(&engine.RunResult{Status: domain.StatusViolation}, domain.AnomalyUnknown, failOnViolation, nil)
	if unnamed == nil || !strings.Contains(unnamed.Error(), "an invariant violated") {
		t.Errorf("a violation without an invariant name still fails, got %v", unnamed)
	}
}

func executeRun(t *testing.T, args ...string) error {
	t.Helper()
	cmd := newRunCmd()
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	var err error
	captureStdout(t, func() { err = cmd.Execute() })
	return err
}

func TestRunRejectsInvalidFailOn(t *testing.T) {
	if err := executeRun(t, bankingSpecPath(t), "--fail-on", "always"); exitCode(err) != exitTrouble {
		t.Fatalf("expected exit 2 for --fail-on always, got %v", err)
	}
	t.Setenv("CHAOSSQL_CLOUD_TOKEN", "")
	if err := executeRun(t, bankingSpecPath(t), "--fail-on", "regression"); exitCode(err) != exitTrouble || !strings.Contains(err.Error(), "--cloud-token") {
		t.Fatalf("expected exit 2 asking for a Cloud token, got %v", err)
	}
}

// With --fail-on regression the Cloud verdict decides, not the local violation.
func TestRunFailOnRegressionFollowsTheCloudVerdict(t *testing.T) {
	isolateCloudTest(t)
	defer func() { cloudTokenFlag, cloudURLFlag = "", defaultCloudURL() }()

	for name, tc := range map[string]struct {
		status       int
		isRegression bool
		want         int
	}{
		"regression":       {http.StatusOK, true, exitFinding},
		"no regression":    {http.StatusOK, false, exitOK},
		"Cloud is failing": {http.StatusServiceUnavailable, false, exitTrouble},
	} {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.status != http.StatusOK {
					w.WriteHeader(tc.status)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(cloud.RunIngestResponse{Success: true, RunID: "run_1", URL: "https://example.test/runs/run_1", IsRegression: tc.isRegression})
			}))
			defer ts.Close()

			err := executeRun(t, bankingSpecPath(t), "--fail-on", "regression", "--cloud-token", "token", "--cloud-url", ts.URL,
				"--github-token", "", "--pr-comment=false")
			if got := exitCode(err); got != tc.want {
				t.Fatalf("exit %d, want %d (err %v)", got, tc.want, err)
			}
		})
	}
}

func TestFailOnDivergence(t *testing.T) {
	// The banking example violates on SQLite (READ_UNCOMMITTED) and passes on
	// the mock driver, whose queries all return zeros.
	bankingDir := filepath.Dir(bankingSpecPath(t))

	for name, args := range map[string][]string{
		"diff":  {"--driver-a", "sqlite", "--driver-b", "mock", bankingSpecPath(t)},
		"swarm": {"--scenarios-dir", bankingDir, "--drivers", "sqlite,mock"},
	} {
		t.Run(name, func(t *testing.T) {
			newCmd := newDiffCmd
			if name == "swarm" {
				newCmd = newSwarmCmd
			}
			for _, tc := range []struct {
				failOn, json bool
			}{{false, true}, {true, true}, {true, false}} {
				cmd := newCmd()
				flags := []string{fmt.Sprintf("--fail-on-divergence=%v", tc.failOn), fmt.Sprintf("--json=%v", tc.json)}
				cmd.SetArgs(append(flags, args...))
				cmd.SetOut(io.Discard)
				cmd.SetErr(io.Discard)
				err := cmd.Execute()
				want := exitOK
				if tc.failOn {
					want = exitFinding
				}
				if got := exitCode(err); got != want {
					t.Fatalf("%v: exit %d, want %d (err %v)", flags, got, want, err)
				}
			}

			// A report that cannot be written is an error, not a finding.
			cmd := newCmd()
			cmd.SetArgs(append([]string{"--json", "--fail-on-divergence"}, args...))
			cmd.SetOut(failingWriter{})
			cmd.SetErr(io.Discard)
			if err := cmd.Execute(); exitCode(err) != exitTrouble {
				t.Fatalf("an unwritable report must exit 2, got %v", err)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }

// Outside --json the Cloud failure under --cloud-fail-fast also wins over the
// violation: the run is reported, but the command exits 2.
func TestRunCloudFailFastInTerminalMode(t *testing.T) {
	isolateCloudTest(t)
	defer func() { cloudTokenFlag, cloudURLFlag, cloudFailFastFlag = "", defaultCloudURL(), false }()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer ts.Close()

	err := executeRun(t, bankingSpecPath(t), "--cloud-token", "token", "--cloud-url", ts.URL, "--cloud-fail-fast",
		"--github-token", "", "--pr-comment=false")
	if exitCode(err) != exitTrouble || !strings.Contains(err.Error(), "failed to publish") || !strings.Contains(err.Error(), "violated") {
		t.Fatalf("expected exit 2 naming both the violation and the Cloud failure, got %v", err)
	}
}

func TestRunUIFailureIsNotMaskedByTheOutcome(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:8090")
	if err != nil {
		t.Skip("port 8090 is not available to occupy")
	}
	defer ln.Close()

	err = executeRun(t, bankingSpecPath(t), "--isolation", "serializable", "--ui")
	if exitCode(err) != exitTrouble {
		t.Fatalf("a trace viewer that cannot listen must exit 2, got %v", err)
	}
}
