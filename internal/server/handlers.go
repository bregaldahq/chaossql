package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	cfg RouterConfig
}

func NewRouter(cfg RouterConfig) http.Handler {
	if cfg.PublicBaseURL == "" {
		cfg.PublicBaseURL = "https://app.chaossql.bregalda.com"
	}
	cfg.PublicBaseURL = strings.TrimRight(cfg.PublicBaseURL, "/")

	s := &Server{cfg: cfg}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", s.handleHealth)
	mux.HandleFunc("POST /v1/runs", s.requireAuth(s.handleIngestRun))
	mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /v1/repositories/{owner}/{name}/runs", s.handleListRuns)

	return mux
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
