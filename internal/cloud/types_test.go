package cloud

import (
	"encoding/json"
	"testing"
	"time"
)

func TestRunIngestRequestSerialization(t *testing.T) {
	req := RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Date(2026, 9, 8, 20, 0, 0, 0, time.UTC),
		CI: &CIContext{
			Provider:          "github-actions",
			Repository:        "bregaldahq/payments",
			CommitSHA:         "a8129c1f",
			Branch:            "feature/new-wallet",
			PullRequestNumber: 382,
		},
		Scenario: ScenarioMetadata{
			Name:       "wallet_transfer",
			Driver:     "postgres",
			Workers:    4,
			Iterations: 100,
			Seed:       184729,
		},
		Result: ExecutionSummary{
			Status:            "failed",
			Success:           false,
			ViolationDetected: true,
			AnomalyType:       "P4",
			DurationMS:        450,
			TotalSchedules:    100,
			FailedSchedules:   31,
			FailingInvariant: &InvariantSummary{
				Name:      "total_wealth_conserved",
				Query:     "SELECT sum(balance) FROM accounts;",
				Assertion: "total == 2000",
				Actual:    "1950",
			},
		},
		Reproduction: &ReproductionData{
			MinimalOperationsCount: 4,
			ShrinkDurationMS:       120,
			SanitizedMinimalTrace: []SanitizedTraceEvent{
				{Worker: "T1", OpType: "read", Table: "accounts", SQL: "SELECT balance FROM accounts WHERE id = 1"},
				{Worker: "T2", OpType: "read", Table: "accounts", SQL: "SELECT balance FROM accounts WHERE id = 1"},
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal RunIngestRequest: %v", err)
	}

	var parsed RunIngestRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal RunIngestRequest: %v", err)
	}

	if parsed.CI.Repository != "bregaldahq/payments" {
		t.Errorf("expected repository bregaldahq/payments, got %q", parsed.CI.Repository)
	}
	if parsed.Result.AnomalyType != "P4" {
		t.Errorf("expected anomaly P4, got %q", parsed.Result.AnomalyType)
	}
	if parsed.Reproduction.MinimalOperationsCount != 4 {
		t.Errorf("expected 4 minimal operations, got %d", parsed.Reproduction.MinimalOperationsCount)
	}
}

func TestRunIngestResponseDeserialization(t *testing.T) {
	raw := `{
		"success": true,
		"run_id": "run_991823a",
		"url": "https://chaossql.bregalda.com/runs/run_991823a",
		"is_regression": true,
		"baseline": {
			"run_id": "run_00129",
			"status": "passed",
			"commit_sha": "abc1234",
			"branch": "main"
		},
		"pr_comment": {
			"posted": true,
			"comment_id": "9921"
		}
	}`

	var resp RunIngestResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if !resp.IsRegression {
		t.Errorf("expected is_regression true, got false")
	}
	if resp.Baseline == nil || resp.Baseline.Branch != "main" {
		t.Errorf("expected baseline branch main, got %v", resp.Baseline)
	}
}
