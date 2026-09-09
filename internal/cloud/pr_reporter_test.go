package cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatPRMarkdown_Regression(t *testing.T) {
	req := &RunIngestRequest{
		Scenario: ScenarioMetadata{
			Name:   "bank_transfer",
			Driver: "postgres",
			Seed:   42891,
		},
		Result: ExecutionSummary{
			Status:          "failed",
			AnomalyType:     "P4",
			TotalSchedules:  100,
			FailedSchedules: 28,
			DurationMS:      340,
			FailingInvariant: &InvariantSummary{
				Name:      "wealth_conservation",
				Assertion: "sum == 2000",
				Actual:    "1950",
			},
		},
		Reproduction: &ReproductionData{
			SanitizedMinimalTrace: []SanitizedTraceEvent{
				{Worker: "T1", SQL: "SELECT balance FROM accounts WHERE id = 1"},
				{Worker: "T2", SQL: "SELECT balance FROM accounts WHERE id = 1"},
				{Worker: "T1", SQL: "UPDATE accounts SET balance = balance - 50 WHERE id = 1"},
				{Worker: "T2", SQL: "UPDATE accounts SET balance = balance - 50 WHERE id = 1"},
			},
		},
	}

	resp := &RunIngestResponse{
		Success:      true,
		RunID:        "run_999",
		URL:          "https://cloud.chaossql.com/runs/run_999",
		IsRegression: true,
		Baseline: &BaselineComparison{
			RunID:  "run_888",
			Status: "passed",
			Branch: "main",
		},
	}

	md := FormatPRMarkdown(req, resp)

	if !strings.Contains(md, "ChaosSQL ❌ Concurrency Regression Detected") {
		t.Errorf("expected regression header, got: %s", md)
	}
	if !strings.Contains(md, "P4") {
		t.Errorf("expected P4 anomaly in markdown")
	}
	if !strings.Contains(md, "Baseline (`main`)") {
		t.Errorf("expected baseline branch row in markdown")
	}
	if !strings.Contains(md, "wealth_conservation") {
		t.Errorf("expected invariant details in markdown")
	}
	if !strings.Contains(md, "https://cloud.chaossql.com/runs/run_999") {
		t.Errorf("expected cloud link in markdown")
	}
}

func TestWriteStepSummary(t *testing.T) {
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "step_summary.md")

	md := "## Test Concurrency Check Passed"
	if err := WriteStepSummary(summaryFile, md); err != nil {
		t.Fatalf("failed to write step summary: %v", err)
	}

	content, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("failed to read step summary: %v", err)
	}

	if !strings.Contains(string(content), md) {
		t.Errorf("step summary file content mismatch: %s", string(content))
	}
}

func TestPostPRComment_Success(t *testing.T) {
	var receivedAuth string
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		receivedBody = payload["body"]

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id": 48210, "html_url": "https://github.com/bregaldahq/core/pull/12#issuecomment-48210"}`))
	}))
	defer ts.Close()

	// Use custom client directing to test server
	customClient := &http.Client{
		Transport: &roundTripperRewrite{targetURL: ts.URL},
	}

	url, err := PostPRComment(context.Background(), customClient, "ghp_secret_token", "bregaldahq/core", 12, "## ChaosSQL Passed")
	if err != nil {
		t.Fatalf("unexpected error posting pr comment: %v", err)
	}

	if url != "https://github.com/bregaldahq/core/pull/12#issuecomment-48210" {
		t.Errorf("expected comment url, got %s", url)
	}
	if receivedAuth != "Bearer ghp_secret_token" {
		t.Errorf("expected Bearer ghp_secret_token, got %s", receivedAuth)
	}
	if receivedBody != "## ChaosSQL Passed" {
		t.Errorf("expected body '## ChaosSQL Passed', got %s", receivedBody)
	}
}

type roundTripperRewrite struct {
	targetURL string
}

func (r *roundTripperRewrite) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(r.targetURL, "http://")
	return http.DefaultTransport.RoundTrip(req)
}
