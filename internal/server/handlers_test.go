package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
	_ "modernc.org/sqlite"
)

func newTestServer(t *testing.T) (http.Handler, *Store, string) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	store := NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	orgID := "org_cloud_test"
	if err := store.CreateOrganization(orgID, "Cloud Test Corp", "enterprise"); err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	token := "chaossql_live_token_777"
	if err := store.CreateAPIToken("tok_777", orgID, token, "Production CI"); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	engine := NewRegressionEngine(store)
	handler := NewRouter(RouterConfig{
		Store:         store,
		Engine:        engine,
		PublicBaseURL: "https://cloud.chaossql.com",
	})

	return handler, store, token
}

func TestHealthEndpoint(t *testing.T) {
	handler, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if res["status"] != "ok" {
		t.Errorf("expected status ok, got %v", res["status"])
	}
}

func TestIngestRunAuthFailure(t *testing.T) {
	handler, _, _ := newTestServer(t)

	body, _ := json.Marshal(cloud.RunIngestRequest{})
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	// Without Authorization header
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", w.Code)
	}

	// With invalid token
	req = httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer invalid_secret")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with invalid token, got %d", w.Code)
	}
}

func TestIngestRunFlowAndRegression(t *testing.T) {
	handler, _, token := newTestServer(t)

	// 1. Submit passing run on main branch
	passReq := cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:   "github-actions",
			Repository: "bregaldahq/core-banking",
			CommitSHA:  "main_sha_001",
			Branch:     "main",
		},
		Scenario: cloud.ScenarioMetadata{
			Name:   "p4_lost_update_test",
			Driver: "postgres",
			Seed:   42,
		},
		Result: cloud.ExecutionSummary{
			Status:            "passed",
			Success:           true,
			ViolationDetected: false,
			AnomalyType:       "NONE",
			DurationMS:        350,
		},
	}

	body, _ := json.Marshal(passReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on main run, got %d: %s", w.Code, w.Body.String())
	}

	var mainResp cloud.RunIngestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &mainResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !mainResp.Success || mainResp.IsRegression {
		t.Fatalf("unexpected mainResp: %+v", mainResp)
	}

	// 2. Submit failing run on Pull Request -> REGRESSION!
	prReq := cloud.RunIngestRequest{
		Version:   "1.0",
		Timestamp: time.Now().UTC(),
		CI: &cloud.CIContext{
			Provider:          "github-actions",
			Repository:        "bregaldahq/core-banking",
			CommitSHA:         "pr_sha_002",
			Branch:            "feat/unlocked-tx",
			PullRequestNumber: 15,
		},
		Scenario: cloud.ScenarioMetadata{
			Name:   "p4_lost_update_test",
			Driver: "postgres",
			Seed:   99,
		},
		Result: cloud.ExecutionSummary{
			Status:            "failed",
			Success:           false,
			ViolationDetected: true,
			AnomalyType:       "P4",
			DurationMS:        410,
			FailingInvariant: &cloud.InvariantSummary{
				Name:      "balance_sum",
				Assertion: "total == 2000",
				Actual:    "1950",
			},
		},
		Reproduction: &cloud.ReproductionData{
			MinimalOperationsCount: 4,
			SanitizedMinimalTrace: []cloud.SanitizedTraceEvent{
				{Worker: "w1", OpType: "write", Table: "accounts", SQL: "UPDATE accounts SET balance = balance - 50"},
			},
		},
	}

	body, _ = json.Marshal(prReq)
	req = httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created on PR run, got %d: %s", w.Code, w.Body.String())
	}

	var prResp cloud.RunIngestResponse
	if err := json.Unmarshal(w.Body.Bytes(), &prResp); err != nil {
		t.Fatalf("failed to decode pr response: %v", err)
	}
	if !prResp.IsRegression {
		t.Errorf("expected is_regression TRUE for PR run, got false")
	}
	if prResp.Baseline == nil || prResp.Baseline.RunID != mainResp.RunID {
		t.Errorf("expected baseline to point to %s, got %+v", mainResp.RunID, prResp.Baseline)
	}

	// 3. Test GET /v1/runs/{id}
	getReq := httptest.NewRequest(http.MethodGet, "/v1/runs/"+prResp.RunID, nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, getReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on GET /v1/runs/{id}, got %d", w.Code)
	}
	var getResult map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &getResult)
	if getResult["run"] == nil || getResult["finding"] == nil {
		t.Errorf("expected run and finding in payload, got %v", getResult)
	}
}

func TestCORSHeaders(t *testing.T) {
	handler, _, _ := newTestServer(t)

	// OPTIONS preflight
	req := httptest.NewRequest(http.MethodOptions, "/v1/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("expected Access-Control-Allow-Methods header")
	}

	// GET request carries CORS header
	getReq := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	gw := httptest.NewRecorder()
	handler.ServeHTTP(gw, getReq)
	if gw.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected Access-Control-Allow-Origin on GET, got %s", gw.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestListAllRuns(t *testing.T) {
	handler, store, _ := newTestServer(t)

	orgID := "org_cloud_test"
	repo, _ := store.GetOrCreateRepo(orgID, "bregaldahq/core-banking", "main")
	sc, _ := store.GetOrCreateScenario(repo.ID, "p4_lost_update_test", "postgres")

	_ = store.SaveRun(&RunRecord{
		ID:          "run_all_01",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "sha_01",
		Branch:      "main",
		Status:      "passed",
		AnomalyType: "NONE",
		Seed:        42,
		DurationMS:  120,
		CreatedAt:   time.Now().UTC(),
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/v1/runs", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GET /v1/runs, got %d", w.Code)
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	runs, ok := res["runs"].([]interface{})
	if !ok || len(runs) == 0 {
		t.Errorf("expected at least 1 run, got %v", res["runs"])
	}
}
