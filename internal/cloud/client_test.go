package cloud

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientPublishRunSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/runs" {
			t.Errorf("expected /v1/runs, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test_token_123" {
			t.Errorf("expected Bearer test_token_123, got %s", r.Header.Get("Authorization"))
		}

		var req RunIngestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(RunIngestResponse{
			Success:      true,
			RunID:        "run_test_01",
			URL:          "https://chaossql.bregalda.com/runs/run_test_01",
			IsRegression: false,
		})
	}))
	defer ts.Close()

	client := NewClient(Config{
		BaseURL: ts.URL,
		Token:   "test_token_123",
	})

	resp, err := client.PublishRun(context.Background(), &RunIngestRequest{
		Scenario: ScenarioMetadata{Name: "banking_transfer"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.Success || resp.RunID != "run_test_01" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestClientPublishRunUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := NewClient(Config{
		BaseURL: ts.URL,
		Token:   "bad_token",
	})

	_, err := client.PublishRun(context.Background(), &RunIngestRequest{})
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestClientPublishRunRetryOn503(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt32(&attempts, 1)
		if current < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(RunIngestResponse{
			Success: true,
			RunID:   "run_recovered",
		})
	}))
	defer ts.Close()

	client := NewClient(Config{
		BaseURL:    ts.URL,
		Token:      "token_123",
		MaxRetries: 3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.PublishRun(ctx, &RunIngestRequest{})
	if err != nil {
		t.Fatalf("expected successful recovery after retries, got %v", err)
	}

	if resp.RunID != "run_recovered" {
		t.Errorf("expected run_recovered, got %s", resp.RunID)
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
