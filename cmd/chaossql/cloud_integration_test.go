package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/cloud"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

func TestCloudPublishingIntegration(t *testing.T) {
	isolateCloudTest(t)
	workersFlag, iterationsFlag, seedFlag = 1, 2, 42
	var receivedReq *cloud.RunIngestRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/runs" {
			t.Errorf("expected /v1/runs, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer mock_secret_token" {
			t.Errorf("expected Bearer mock_secret_token, got %s", r.Header.Get("Authorization"))
		}

		_ = json.NewDecoder(r.Body).Decode(&receivedReq)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cloud.RunIngestResponse{
			Success:      true,
			RunID:        "run_mock_integration_123",
			URL:          "https://chaossql.bregalda.com/runs/run_mock_integration_123",
			IsRegression: false,
		})
	}))
	defer ts.Close()

	cloudTokenFlag = "mock_secret_token"
	cloudURLFlag = ts.URL
	cloudFailFastFlag = true
	defer func() {
		cloudTokenFlag = ""
		cloudURLFlag = defaultCloudURL()
		cloudFailFastFlag = false
		workersFlag, iterationsFlag, seedFlag = 0, 0, 0
	}()

	err := executeChaos(context.Background(), "../../examples/banking_lost_update/chaos.yaml")
	if err != nil {
		t.Fatalf("unexpected error executing chaos with cloud integration: %v", err)
	}

	if receivedReq == nil {
		t.Fatal("expected mock server to receive run ingest request, but got nil")
	}

	if receivedReq.Scenario.Name != "banking_lost_update" {
		t.Errorf("expected scenario name banking_lost_update, got %q", receivedReq.Scenario.Name)
	}
	if receivedReq.Result.AnomalyType == "" {
		t.Errorf("expected P4 anomaly in received request, got %q", receivedReq.Result.AnomalyType)
	}
	if receivedReq.Scenario.Seed != 42 {
		t.Errorf("expected effective seed in cloud request, got scenario=%+v", receivedReq.Scenario)
	}
	if receivedReq.Schedule != nil {
		t.Errorf("expected replay schedule to remain local, got %+v", receivedReq.Schedule)
	}
	if len(receivedReq.Scenario.Fingerprint) != 64 {
		t.Errorf("missing semantic fingerprint: %q", receivedReq.Scenario.Fingerprint)
	}
	if receivedReq.Result.TotalSchedules != 1 {
		t.Errorf("got %d schedules, one execution runs one schedule", receivedReq.Result.TotalSchedules)
	}
	if receivedReq.Reproduction != nil && len(receivedReq.Reproduction.SanitizedMinimalTrace) != 0 {
		t.Errorf("expected detailed trace to remain local, got %d events", len(receivedReq.Reproduction.SanitizedMinimalTrace))
	}
}

func TestCloudPRReporterStepSummaryIntegration(t *testing.T) {
	isolateCloudTest(t)
	workersFlag, iterationsFlag, seedFlag = 1, 2, 42
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cloud.RunIngestResponse{
			Success:      true,
			RunID:        "run_summary_test",
			URL:          "https://chaossql.bregalda.com/runs/run_summary_test",
			IsRegression: false,
		})
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	summaryFile := tmpDir + "/step_summary.md"
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	cloudTokenFlag = "mock_secret_token"
	cloudURLFlag = ts.URL
	cloudFailFastFlag = false
	defer func() {
		cloudTokenFlag = ""
		cloudURLFlag = defaultCloudURL()
		workersFlag, iterationsFlag, seedFlag = 0, 0, 0
	}()

	err := executeChaos(context.Background(), "../../examples/banking_lost_update/chaos.yaml")
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	data, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("expected step summary file to be written, err: %v", err)
	}

	summaryContent := string(data)
	if !strings.Contains(summaryContent, "ChaosSQL") {
		t.Errorf("expected summary to contain ChaosSQL, got: %s", summaryContent)
	}
	if !strings.Contains(summaryContent, "banking_lost_update") {
		t.Errorf("expected summary to contain scenario name banking_lost_update")
	}
}

func isolateCloudTest(t *testing.T) {
	t.Helper()
	oldToken, oldPR, oldJSON := githubTokenFlag, prCommentFlag, jsonFlag
	githubTokenFlag = ""
	prCommentFlag = false
	jsonFlag = false
	t.Setenv("GITHUB_OUTPUT", "")
	t.Setenv("GITHUB_STEP_SUMMARY", "")
	t.Cleanup(func() { githubTokenFlag = oldToken; prCommentFlag = oldPR; jsonFlag = oldJSON })
}

func TestCloudFailFastInJSON(t *testing.T) {
	isolateCloudTest(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer ts.Close()
	oldToken, oldURL, oldFast := cloudTokenFlag, cloudURLFlag, cloudFailFastFlag
	oldWorkers, oldIterations := workersFlag, iterationsFlag
	defer func() {
		cloudTokenFlag, cloudURLFlag, cloudFailFastFlag = oldToken, oldURL, oldFast
		workersFlag, iterationsFlag = oldWorkers, oldIterations
	}()
	cloudTokenFlag = "test"
	cloudURLFlag = ts.URL
	cloudFailFastFlag = true
	jsonFlag = true
	workersFlag = 1
	iterationsFlag = 2
	err := executeChaos(context.Background(), "../../examples/banking_lost_update/chaos.yaml")
	if err == nil || !strings.Contains(err.Error(), "failed to publish") {
		t.Fatalf("JSON suppressed cloud-fail-fast: %v", err)
	}
}

func TestCloudPublishesActionOutputsAndFailureCount(t *testing.T) {
	isolateCloudTest(t)
	var received cloud.RunIngestRequest
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		_ = json.NewEncoder(w).Encode(cloud.RunIngestResponse{Success: true, RunID: "run_123", URL: "https://cloud.example/runs/run_123", IsRegression: true})
	}))
	defer ts.Close()
	oldToken, oldURL := cloudTokenFlag, cloudURLFlag
	defer func() { cloudTokenFlag, cloudURLFlag = oldToken, oldURL }()
	cloudTokenFlag = "test"
	cloudURLFlag = ts.URL
	output := filepath.Join(t.TempDir(), "outputs")
	t.Setenv("GITHUB_OUTPUT", output)
	_, err := publishToCloud(context.Background(), domain.Spec{Name: "example", Database: domain.DatabaseConfig{Driver: "mock"}}, &engine.RunResult{Status: domain.StatusViolation, ViolationDetected: true}, nil, nil, domain.AnomalyLostUpdate, "", "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{"cloud-run-id=run_123\n", "cloud-run-url=https://cloud.example/runs/run_123\n", "is-regression=true\n"} {
		if !strings.Contains(string(data), entry) {
			t.Errorf("missing %q in %s", entry, data)
		}
	}
	if received.Result.TotalSchedules != 1 || received.Result.FailedSchedules != 1 {
		t.Fatalf("unfaithful schedule counts: %+v", received.Result)
	}
}
