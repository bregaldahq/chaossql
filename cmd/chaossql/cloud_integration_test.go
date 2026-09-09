package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

func TestCloudPublishingIntegration(t *testing.T) {
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
	}()

	err := executeChaos("../../examples/banking_lost_update/chaos.yaml")
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
	if len(receivedReq.Reproduction.SanitizedMinimalTrace) == 0 {
		t.Errorf("expected sanitized minimal trace, got 0 events")
	}
}
