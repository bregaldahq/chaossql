package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

type RouterConfig struct {
	Store         *Store
	Engine        *RegressionEngine
	PublicBaseURL string
}

type Server struct {
	cfg        RouterConfig
	dispatcher *WebhookDispatcher
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Org-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewRouter(cfg RouterConfig) http.Handler {
	if cfg.PublicBaseURL == "" {
		cfg.PublicBaseURL = "https://app.chaossql.bregalda.com"
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")

	s := &Server{
		cfg:        cfg,
		dispatcher: NewWebhookDispatcher(nil),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.handleHealth)
	mux.HandleFunc("GET /v1/runs", s.handleListAllRuns)
	mux.HandleFunc("POST /v1/runs", s.requireAuth(s.handleIngestRun))
	mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /v1/repositories/{owner}/{name}/runs", s.handleListRuns)
	mux.HandleFunc("GET /v1/organizations/{id}/subscription", s.requireAuth(s.handleGetSubscription))

	// Webhooks management API
	mux.HandleFunc("GET /v1/organizations/{id}/webhooks", s.requireAuth(s.handleListWebhooks))
	mux.HandleFunc("POST /v1/organizations/{id}/webhooks", s.requireAuth(s.handleCreateWebhook))
	mux.HandleFunc("DELETE /v1/organizations/{id}/webhooks/{wh_id}", s.requireAuth(s.handleDeleteWebhook))
	mux.HandleFunc("POST /v1/organizations/{id}/webhooks/test", s.requireAuth(s.handleTestWebhook))

	// Short-form webhook routes for the local dashboard. These are NOT
	// unauthenticated by default: see requireLocalDashboard.
	mux.HandleFunc("GET /v1/webhooks", s.requireLocalDashboard(s.handleListWebhooks))
	mux.HandleFunc("POST /v1/webhooks", s.requireLocalDashboard(s.handleCreateWebhook))
	mux.HandleFunc("DELETE /v1/webhooks/{wh_id}", s.requireLocalDashboard(s.handleDeleteWebhook))
	mux.HandleFunc("POST /v1/webhooks/test", s.requireLocalDashboard(s.handleTestWebhook))

	return corsMiddleware(mux)
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing or malformed Authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			http.Error(w, `{"error":"empty bearer token"}`, http.StatusUnauthorized)
			return
		}

		orgID, err := s.cfg.Store.ValidateToken(token)
		if err != nil || orgID == "" {
			http.Error(w, `{"error":"invalid or unauthorized token"}`, http.StatusUnauthorized)
			return
		}

		// Store orgID in request header/context if needed
		r.Header.Set("X-Org-ID", orgID)
		next(w, r)
	}
}

// localDashboardEnvVar opts in to unauthenticated webhook management. It is
// off by default: handleListWebhooks returns full WebhookRecords, and a
// webhook URL embeds the provider secret, so an unauthenticated GET on a
// reachable host hands out every stored Discord/Slack token.
const localDashboardEnvVar = "CHAOSSQL_LOCAL_DASHBOARD"

// requireLocalDashboard permits the short-form webhook routes without a token
// only when the operator explicitly opted in AND the caller is on the loopback
// interface. Anything else falls back to normal bearer-token auth.
//
// Note: if this server ever sits behind a same-host reverse proxy, every
// request arrives from loopback. Do not enable the opt-in in that topology.
func (s *Server) requireLocalDashboard(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if localDashboardEnabled() && isLoopbackRequest(r) {
			next(w, r)
			return
		}
		s.requireAuth(next)(w, r)
	}
}

func localDashboardEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(localDashboardEnvVar))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isLoopbackRequest(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	return ip != nil && ip.IsLoopback()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

func (s *Server) handleIngestRun(w http.ResponseWriter, r *http.Request) {
	orgID := r.Header.Get("X-Org-ID")
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req cloud.RunIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid json payload: " + err.Error()})
		return
	}

	repoFullName := "local/chaossql-project"
	branch := "main"
	defaultBranch := "main"
	commitSHA := "unknown"
	prNumber := 0

	if req.CI != nil {
		if req.CI.Repository != "" {
			repoFullName = req.CI.Repository
		}
		if req.CI.Branch != "" {
			branch = req.CI.Branch
		}
		if req.CI.BaseBranch != "" {
			defaultBranch = req.CI.BaseBranch
		}
		commitSHA = req.CI.CommitSHA
		prNumber = req.CI.PullRequestNumber
	}

	scenarioName := req.Scenario.Name
	if scenarioName == "" {
		scenarioName = "default"
	}
	driver := req.Scenario.Driver
	if driver == "" {
		driver = "sqlite"
	}

	repo, err := s.cfg.Store.GetOrCreateRepo(orgID, repoFullName, defaultBranch)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to resolve repo: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	sc, err := s.cfg.Store.GetOrCreateScenario(repo.ID, scenarioName, driver)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to resolve scenario: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	runID := fmt.Sprintf("run_%d", time.Now().UnixNano())
	run := &RunRecord{
		ID:          runID,
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   commitSHA,
		Branch:      branch,
		PRNumber:    prNumber,
		Status:      req.Result.Status,
		AnomalyType: req.Result.AnomalyType,
		Seed:        req.Scenario.Seed,
		DurationMS:  req.Result.DurationMS,
		CreatedAt:   time.Now().UTC(),
	}

	var finding *FindingRecord
	if req.Result.ViolationDetected || req.Reproduction != nil {
		var traceJSON string
		var reproCode string
		minimalOps := 0
		if req.Reproduction != nil {
			minimalOps = req.Reproduction.MinimalOperationsCount
			reproCode = req.Reproduction.ReproGoCode
			if len(req.Reproduction.SanitizedMinimalTrace) > 0 {
				data, _ := json.Marshal(req.Reproduction.SanitizedMinimalTrace)
				traceJSON = string(data)
			}
		}

		assertion := ""
		if req.Result.FailingInvariant != nil {
			assertion = fmt.Sprintf("%s: %s (got %s)", req.Result.FailingInvariant.Name, req.Result.FailingInvariant.Assertion, req.Result.FailingInvariant.Actual)
		}

		finding = &FindingRecord{
			ID:          fmt.Sprintf("find_%d", time.Now().UnixNano()),
			RunID:       runID,
			AnomalyType: req.Result.AnomalyType,
			Assertion:   assertion,
			MinimalOps:  minimalOps,
			ReproCode:   reproCode,
			TraceJSON:   traceJSON,
			CreatedAt:   time.Now().UTC(),
		}
	}

	if err := s.cfg.Store.SaveRun(run, finding); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to save run: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	comp, isReg, err := s.cfg.Engine.Evaluate(repo, sc, run)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to evaluate regression: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	if isReg || run.Status != "passed" {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			eventType := "regression"
			if !isReg {
				eventType = "failure"
			}

			webhooks, err := s.cfg.Store.GetActiveWebhooksForEvent(ctx, repo.OrgID, eventType)
			if err != nil || len(webhooks) == 0 {
				return
			}

			baseStatus := "PASS"
			if comp != nil && comp.Status != "" {
				baseStatus = comp.Status
			}

			anomalyName := run.AnomalyType
			if req.Result.ViolationDetected && req.Result.AnomalyType != "" {
				anomalyName = req.Result.AnomalyType
			}

			alert := &RegressionAlert{
				RepoFullName:   repo.FullName,
				Branch:         run.Branch,
				PRNumber:       run.PRNumber,
				CommitSHA:      run.CommitSHA,
				AnomalyType:    run.AnomalyType,
				AnomalyName:    anomalyName,
				Driver:         sc.Driver,
				Isolation:      "READ COMMITTED",
				Scenario:       sc.Name,
				Seed:           run.Seed,
				DurationMS:     run.DurationMS,
				BaselineStatus: baseStatus,
				RunURL:         fmt.Sprintf("%s/#/visualizer?scenario=%s&seed=%d", s.cfg.PublicBaseURL, sc.Name, run.Seed),
				IsRegression:   isReg,
			}

			for _, wh := range webhooks {
				_ = s.dispatcher.DispatchAlert(ctx, wh, alert)
			}
		}()
	}


	msg := "Execution passed successfully."
	if isReg {
		baseBranch := defaultBranch
		if comp != nil && comp.Branch != "" {
			baseBranch = comp.Branch
		}
		msg = fmt.Sprintf("Regression detected! Anomaly %s broke baseline on %s", run.AnomalyType, baseBranch)
	} else if run.Status != "passed" {
		msg = fmt.Sprintf("Execution finished with status %s.", run.Status)
	}

	resp := cloud.RunIngestResponse{
		Success:      true,
		RunID:        runID,
		URL:          fmt.Sprintf("%s/runs/%s", s.cfg.PublicBaseURL, runID),
		IsRegression: isReg,
		Baseline:     comp,
		Message:      msg,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	runID := r.PathValue("id")
	if runID == "" {
		http.Error(w, `{"error":"missing run id"}`, http.StatusBadRequest)
		return
	}

	run, finding, err := s.cfg.Store.GetRun(runID)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, `{"error":"run not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"run":     run,
		"finding": finding,
	})
}

func (s *Server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	name := r.PathValue("name")
	fullName := fmt.Sprintf("%s/%s", owner, name)

	// Lookup repo
	var repoID string
	err := s.cfg.Store.db.QueryRow(`SELECT id FROM repositories WHERE full_name = ?`, fullName).Scan(&repoID)
	if err != nil {
		http.Error(w, `{"error":"repository not found"}`, http.StatusNotFound)
		return
	}

	runs, err := s.cfg.Store.ListRuns(repoID, 50, 0)
	if err != nil {
		http.Error(w, `{"error":"failed to list runs"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"repository": fullName,
		"runs":       runs,
	})
}


func (s *Server) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	callerOrgID := r.Header.Get("X-Org-ID")
	if orgID == "me" || orgID == "" {
		orgID = callerOrgID
	}

	sub, err := s.cfg.Store.GetOrgSubscription(orgID)
	if errors.Is(err, ErrNotFound) {
		http.Error(w, `{"error":"organization not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sub)
}

func (s *Server) handleListAllRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.cfg.Store.ListRecentRuns(50, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list runs: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"runs":  runs,
		"count": len(runs),
	})
}

type CreateWebhookRequest struct {
	TargetType string `json:"target_type"` // "discord", "slack", "generic"
	URL        string `json:"url"`
	Events     string `json:"events"`      // "regression", "failure", "all"
	Active     *bool  `json:"active,omitempty"`
}

type TestWebhookRequest struct {
	TargetType string `json:"target_type"`
	URL        string `json:"url"`
}

func (s *Server) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = r.Header.Get("X-Org-ID")
	}
	if orgID == "" {
		orgID = "org_default"
	}

	webhooks, err := s.cfg.Store.ListWebhooks(r.Context(), orgID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to list webhooks: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if webhooks == nil {
		webhooks = []WebhookRecord{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(webhooks)
}

func (s *Server) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = r.Header.Get("X-Org-ID")
	}
	if orgID == "" {
		orgID = "org_default"
	}

	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := ValidateWebhookTarget(req.URL); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}

	targetType := strings.ToLower(req.TargetType)
	if targetType == "" {
		targetType = "generic"
	}
	events := req.Events
	if events == "" {
		events = "regression,all"
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	wh := &WebhookRecord{
		ID:         fmt.Sprintf("wh_%d", time.Now().UnixNano()),
		OrgID:      orgID,
		TargetType: targetType,
		URL:        req.URL,
		Events:     events,
		Active:     active,
		CreatedAt:  time.Now().UTC(),
	}

	if err := s.cfg.Store.CreateWebhook(r.Context(), wh); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create webhook: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(wh)
}

func (s *Server) handleDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	orgID := r.PathValue("id")
	if orgID == "" {
		orgID = r.Header.Get("X-Org-ID")
	}
	if orgID == "" {
		orgID = "org_default"
	}

	whID := r.PathValue("wh_id")
	if whID == "" {
		whID = r.PathValue("id")
	}

	if err := s.cfg.Store.DeleteWebhook(r.Context(), orgID, whID); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to delete webhook: %s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleTestWebhook(w http.ResponseWriter, r *http.Request) {
	var req TestWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := ValidateWebhookTarget(req.URL); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		return
	}

	testAlert := &RegressionAlert{
		RepoFullName:   "acme/payments",
		Branch:         "feat/instant-settlement",
		PRNumber:       42,
		CommitSHA:      "a1b2c3d",
		AnomalyType:    "P4",
		AnomalyName:    "Lost Update (Test Simulation)",
		Driver:         "PostgreSQL 16",
		Isolation:      "READ COMMITTED",
		Scenario:       "wallet_transfer",
		Seed:           99999,
		DurationMS:     280,
		BaselineStatus: "PASS",
		RunURL:         s.cfg.PublicBaseURL,
		IsRegression:   true,
	}

	wh := WebhookRecord{
		ID:         "test",
		TargetType: req.TargetType,
		URL:        req.URL,
	}

	if err := s.dispatcher.DispatchAlert(r.Context(), wh, testAlert); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"webhook dispatch failed: %s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Webhook test alert dispatched successfully!",
	})
}
