package cloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestDetectCIContextGitLab(t *testing.T) {
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "true")
	t.Setenv("CI_PROJECT_PATH", "group/service")
	t.Setenv("CI_COMMIT_SHA", "not-a-sha")
	t.Setenv("CI_COMMIT_REF_NAME", "feature/x")
	t.Setenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "main")
	t.Setenv("CI_MERGE_REQUEST_IID", "42")
	t.Setenv("CI_PIPELINE_ID", "9001")
	t.Setenv("GITLAB_USER_LOGIN", "dev")

	ci := DetectCIContext()
	want := CIContext{Provider: "gitlab-ci", Repository: "group/service", CommitSHA: "not-a-sha", Branch: "feature/x",
		BaseBranch: "main", PullRequestNumber: 42, RunID: "9001", Actor: "dev"}
	if *ci != want {
		t.Fatalf("DetectCIContext = %+v, want %+v", *ci, want)
	}
}

func TestDetectCIContextLocalGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q", "-b", "trunk")
	git("remote", "add", "origin", "https://user:secret@github.com/acme/payments.git")
	git("commit", "-q", "--allow-empty", "-m", "init")

	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	t.Setenv("USER", "local-dev")
	t.Chdir(dir)

	ci := DetectCIContext()
	if ci.Provider != "local" || ci.Repository != "acme/payments" || ci.Branch != "trunk" || ci.Actor != "local-dev" {
		t.Fatalf("DetectCIContext = %+v", ci)
	}
	if !commitSHAPattern.MatchString(ci.CommitSHA) || ci.CommitTimestamp == nil {
		t.Fatalf("expected HEAD sha and commit timestamp, got %+v", ci)
	}
}

func TestDetectLocalRepoWithoutRemote(t *testing.T) {
	t.Chdir(t.TempDir())
	if got := detectLocalRepo(); got != "local/repository" {
		t.Fatalf("detectLocalRepo = %q, want local/repository", got)
	}
}

func TestClassifyOpType(t *testing.T) {
	cases := map[string]domain.TraceEvent{
		"begin":     {Type: domain.EventBegin},
		"commit":    {Type: domain.EventCommit},
		"rollback":  {Type: domain.EventRollbackTo},
		"savepoint": {Type: domain.EventSavepoint},
		"error":     {Type: domain.EventError},
		"read":      {Type: domain.EventExec, SQL: "select 1"},
		"write":     {Type: domain.EventExec, SQL: "DELETE FROM t"},
		"exec":      {Type: domain.EventExec, SQL: "PRAGMA foreign_keys = ON"},
	}
	for want, event := range cases {
		if got := ClassifyOpType(event); got != want {
			t.Errorf("ClassifyOpType(%+v) = %q, want %q", event, got, want)
		}
	}
}

func TestWriteActionOutputsUsesGitHubOutputEnv(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "")
	if err := WriteActionOutputs("", &RunIngestResponse{RunID: "r"}); err != nil {
		t.Fatalf("no output file configured should be a no-op, got %v", err)
	}
	path := filepath.Join(t.TempDir(), "out")
	t.Setenv("GITHUB_OUTPUT", path)
	if err := WriteActionOutputs("", &RunIngestResponse{RunID: "run-1", URL: "https://x/dashboard?run=run-1", IsRegression: true}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "cloud-run-id=run-1\ncloud-run-url=https://x/dashboard?run=run-1\nis-regression=true\n" {
		t.Fatalf("outputs = %q", data)
	}
	if err := WriteActionOutputs(filepath.Join(path, "not-a-dir", "out"), &RunIngestResponse{}); err == nil {
		t.Fatal("expected an error for an unwritable output path")
	}
}

func TestDecodeMetadataOnlyRequestRejectsUnsafeEnums(t *testing.T) {
	base := `{"version":"1","timestamp":"2026-09-01T00:00:00Z","scenario":{"name":"s","driver":"%s","workers":1,"iterations":1,"seed":1},"result":{"status":"%s","execution_status":"%s","anomaly_type":"%s"}%s}`
	cases := []struct {
		name, driver, status, execStatus, anomaly, extra, wantField string
	}{
		{"driver", "oracle", "passed", "passed", "NONE", "", "scenario.driver"},
		{"status", "sqlite", "flaky", "passed", "NONE", "", "result.status"},
		{"execution status", "sqlite", "passed", "exploded", "NONE", "", "result.execution_status"},
		{"anomaly", "sqlite", "failed", "violation", "G9", "", "result.anomaly_type"},
		{"ci provider", "sqlite", "passed", "passed", "NONE", `,"ci":{"provider":"jenkins","repository":"a/b","commit_sha":"abc1234","branch":"main"}`, "ci.provider"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(base, tc.driver, tc.status, tc.execStatus, tc.anomaly, tc.extra)
			_, err := DecodeMetadataOnlyRequest([]byte(body))
			if !errors.Is(err, ErrUnsafeMetadata) || !strings.Contains(err.Error(), tc.wantField) {
				t.Fatalf("err = %v, want ErrUnsafeMetadata for %s", err, tc.wantField)
			}
		})
	}
}

func TestDecodeMetadataOnlyRequestKeepsOptionalSections(t *testing.T) {
	body := `{"version":"1","timestamp":"2026-09-01T00:00:00Z","scenario":{"name":"banking","driver":"postgres","workers":2,"iterations":5,"seed":3},` +
		`"result":{"status":"failed","execution_status":"violation","anomaly_type":"P4","failing_invariant":{"name":"balance_preserved"}},` +
		`"reproduction":{"minimal_operations_count":2,"shrink_duration_ms":15}}`
	req, err := DecodeMetadataOnlyRequest([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if req.Result.FailingInvariant == nil || req.Result.FailingInvariant.Name != "balance_preserved" {
		t.Fatalf("failing invariant = %+v", req.Result.FailingInvariant)
	}
	if req.Reproduction == nil || req.Reproduction.MinimalOperationsCount != 2 || req.Reproduction.ShrinkDurationMS != 15 {
		t.Fatalf("reproduction = %+v", req.Reproduction)
	}

	if _, err := DecodeMetadataOnlyRequest([]byte(body + body)); err == nil {
		t.Fatal("two concatenated objects must be rejected")
	}
}

func TestFormatPRMarkdownStatesAndUnsafeNames(t *testing.T) {
	passed := FormatPRMarkdown(&RunIngestRequest{Scenario: ScenarioMetadata{Name: "https://evil.example/x"}, Result: ExecutionSummary{Status: "passed", Success: true}}, nil)
	if !strings.Contains(passed, "Verification Passed") || !strings.Contains(passed, "`default`") || strings.Contains(passed, "evil") {
		t.Fatalf("passed report:\n%s", passed)
	}
	failed := FormatPRMarkdown(&RunIngestRequest{Scenario: ScenarioMetadata{Name: "banking"}, Result: ExecutionSummary{Status: "failed"}}, &RunIngestResponse{})
	if !strings.Contains(failed, "Anomaly Detected") {
		t.Fatalf("failed report:\n%s", failed)
	}
}

func TestWriteStepSummaryEnvAndErrors(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "")
	if err := WriteStepSummary("", "# report"); err != nil {
		t.Fatalf("unset summary path should be a no-op, got %v", err)
	}
	path := filepath.Join(t.TempDir(), "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", path)
	if err := WriteStepSummary("", "# report"); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); string(data) != "# report\n\n" {
		t.Fatalf("summary = %q", data)
	}
	if err := WriteStepSummary(filepath.Join(path, "nested"), "x"); err == nil {
		t.Fatal("expected an error when the summary path cannot be opened")
	}
}

func TestPostPRCommentSkipsAndReportsFailures(t *testing.T) {
	if id, err := PostPRComment(context.Background(), nil, "", "a/b", 1, "x"); id != "" || err != nil {
		t.Fatalf("missing token should skip, got %q %v", id, err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/issues/2/") {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
			return
		}
		http.Error(w, `{"message":"Resource not accessible"}`, http.StatusForbidden)
	}))
	defer ts.Close()
	client := &http.Client{Transport: &roundTripperRewrite{targetURL: ts.URL}}

	if _, err := PostPRComment(context.Background(), client, "tok", "a/b", 1, "x"); err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Fatalf("err = %v, want HTTP 403", err)
	}
	if id, err := PostPRComment(context.Background(), client, "tok", "a/b", 2, "x"); err != nil || !strings.HasPrefix(id, "comment_") {
		t.Fatalf("id = %q err = %v, want a synthetic comment id when GitHub omits html_url", id, err)
	}

	unreachable := &http.Client{Transport: &roundTripperRewrite{targetURL: "http://127.0.0.1:1"}}
	if _, err := PostPRComment(context.Background(), unreachable, "tok", "a/b", 1, "x"); err == nil || !strings.Contains(err.Error(), "dispatch") {
		t.Fatalf("err = %v, want a dispatch failure", err)
	}
}

func TestClientPublishRunRejectsMissingInputs(t *testing.T) {
	client := NewClient(Config{BaseURL: "http://example.invalid/", Token: "tok"})
	if client.BaseURL() != "http://example.invalid" {
		t.Fatalf("BaseURL = %q, want trailing slash trimmed", client.BaseURL())
	}
	if _, err := client.PublishRun(context.Background(), nil); err == nil {
		t.Fatal("nil payload must be rejected")
	}
	if _, err := NewClient(Config{}).PublishRun(context.Background(), &RunIngestRequest{}); err == nil || !strings.Contains(err.Error(), "token") {
		t.Fatalf("err = %v, want missing token", err)
	}
}

func TestClientPublishRunRejectsMalformedAcknowledgement(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not json`))
	}))
	defer ts.Close()
	_, err := NewClient(Config{BaseURL: ts.URL, Token: "tok"}).PublishRun(context.Background(), &RunIngestRequest{Scenario: ScenarioMetadata{Name: "s"}})
	if err == nil || !strings.Contains(err.Error(), "parse cloud response") {
		t.Fatalf("err = %v, want a parse failure", err)
	}
}
