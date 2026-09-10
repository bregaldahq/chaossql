package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
	"github.com/bregaldahq/chaossql/internal/server"
	_ "modernc.org/sqlite"
)

func TestCloudControlPlaneE2E(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	store := server.NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	orgID := "org_bregalda_cloud"
	if err := store.CreateOrganization(orgID, "Bregalda Cloud Testing", "enterprise"); err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	token := "chaossql_e2e_secret_token"
	if err := store.CreateAPIToken("tok_e2e", orgID, token, "E2E Token"); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	engine := server.NewRegressionEngine(store)
	router := server.NewRouter(server.RouterConfig{
		Store:         store,
		Engine:        engine,
		PublicBaseURL: "https://cloud.chaossql.com",
	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	client := cloud.NewClient(cloud.Config{
		BaseURL: ts.URL,
		Token:   token,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Run 1 on main branch (PASSED -> establishes baseline)
	run1Req := &cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:   "github-actions",
			Repository: "acme/fintech-ledger",
			CommitSHA:  "commit_base_100",
			Branch:     "main",
			Actor:      "octocat",
		},
		Scenario: cloud.ScenarioMetadata{
			Name:       "double_spend_check",
			Driver:     "postgres",
			Workers:    4,
			Iterations: 100,
			Seed:       12345,
		},
		Result: cloud.ExecutionSummary{
			Status:            "passed",
			Success:           true,
			ViolationDetected: false,
			AnomalyType:       "NONE",
			DurationMS:        410,
			TotalSchedules:    100,
			FailedSchedules:   0,
		},
	}

	resp1, err := client.PublishRun(ctx, run1Req)
	if err != nil {
		t.Fatalf("failed to publish baseline run: %v", err)
	}
	if !resp1.Success {
		t.Errorf("expected resp1 success true, got false")
	}
	if resp1.IsRegression {
		t.Errorf("expected resp1 is_regression false for first run, got true")
	}
	t.Logf("Baseline run established: %s (URL: %s)", resp1.RunID, resp1.URL)

	// Step 2: Run 2 on Pull Request (FAILED -> must trigger REGRESSION against baseline)
	run2Req := &cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:          "github-actions",
			Repository:        "acme/fintech-ledger",
			CommitSHA:         "commit_pr_200",
			Branch:            "feat/instant-settlement",
			PullRequestNumber: 42,
			Actor:             "contributor",
		},
		Scenario: cloud.ScenarioMetadata{
			Name:       "double_spend_check",
			Driver:     "postgres",
			Workers:    4,
			Iterations: 100,
			Seed:       99999,
		},
		Result: cloud.ExecutionSummary{
			Status:            "failed",
			Success:           false,
			ViolationDetected: true,
			AnomalyType:       "P4",
			DurationMS:        550,
			TotalSchedules:    100,
			FailedSchedules:   18,
			FailingInvariant: &cloud.InvariantSummary{
				Name:      "no_negative_balance",
				Query:     "SELECT count(*) FROM accounts WHERE balance < 0",
				Assertion: "count == 0",
				Actual:    "1",
			},
		},
		Reproduction: &cloud.ReproductionData{
			MinimalOperationsCount: 4,
			ShrinkDurationMS:       80,
			ReproGoCode:            "package main\nfunc main() { /* repro */ }",
			SanitizedMinimalTrace: []cloud.SanitizedTraceEvent{
				{Worker: "T1", OpType: "read", Table: "accounts", SQL: "SELECT balance FROM accounts WHERE id = 1"},
				{Worker: "T2", OpType: "read", Table: "accounts", SQL: "SELECT balance FROM accounts WHERE id = 1"},
				{Worker: "T1", OpType: "write", Table: "accounts", SQL: "UPDATE accounts SET balance = balance - 100 WHERE id = 1"},
				{Worker: "T2", OpType: "write", Table: "accounts", SQL: "UPDATE accounts SET balance = balance - 100 WHERE id = 1"},
			},
		},
	}

	resp2, err := client.PublishRun(ctx, run2Req)
	if err != nil {
		t.Fatalf("failed to publish PR run: %v", err)
	}
	if !resp2.Success {
		t.Errorf("expected resp2 success true, got false")
	}
	if !resp2.IsRegression {
		t.Errorf("expected resp2 IS_REGRESSION = TRUE! PR broke baseline on main")
	}
	if resp2.Baseline == nil || resp2.Baseline.RunID != resp1.RunID {
		t.Errorf("expected resp2 baseline to point to run1 (%s), got %+v", resp1.RunID, resp2.Baseline)
	}
	t.Logf("Regression successfully flagged: %s against baseline %s", resp2.RunID, resp1.RunID)

	// Step 3: Verify GET /v1/runs/{id} returns full run and finding
	getURL := ts.URL + "/v1/runs/" + resp2.RunID
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, getURL, nil)
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("failed to fetch run details: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from GET /v1/runs/{id}, got %d", httpResp.StatusCode)
	}

	var runPayload struct {
		Run     server.RunRecord     `json:"run"`
		Finding server.FindingRecord `json:"finding"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("failed to parse run payload: %v", err)
	}

	if runPayload.Run.AnomalyType != "P4" {
		t.Errorf("expected anomaly P4 in stored run, got %s", runPayload.Run.AnomalyType)
	}
	if runPayload.Finding.MinimalOps != 4 {
		t.Errorf("expected 4 minimal ops in finding, got %d", runPayload.Finding.MinimalOps)
	}
	t.Logf("E2E Test Verified: finding has minimal_ops=%d and repro_code=%t",
		runPayload.Finding.MinimalOps, len(runPayload.Finding.ReproCode) > 0)
}
