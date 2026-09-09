package server

import (
	"strings"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("unauthorized token")
)

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository struct {
	ID            string    `json:"id"`
	OrgID         string    `json:"org_id"`
	FullName      string    `json:"full_name"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
}

type Scenario struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Name      string    `json:"name"`
	Driver    string    `json:"driver"`
	CreatedAt time.Time `json:"created_at"`
}

type RunRecord struct {
	ID          string    `json:"id"`
	RepoID      string    `json:"repo_id"`
	ScenarioID  string    `json:"scenario_id"`
	CommitSHA   string    `json:"commit_sha"`
	Branch      string    `json:"branch"`
	PRNumber    int       `json:"pr_number"`
	Status      string    `json:"status"` // "passed" or "failed"
	AnomalyType string    `json:"anomaly_type"`
	Seed        uint64    `json:"seed"`
	DurationMS  int64     `json:"duration_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

type FindingRecord struct {
	ID          string    `json:"id"`
	RunID       string    `json:"run_id"`
	AnomalyType string    `json:"anomaly_type"`
	Assertion   string    `json:"assertion"`
	MinimalOps  int       `json:"minimal_ops"`
	ReproCode   string    `json:"repro_code"`
	TraceJSON   string    `json:"trace_json"`
	CreatedAt   time.Time `json:"created_at"`
}


type WebhookRecord struct {
	ID         string    `json:"id"`
	OrgID      string    `json:"org_id"`
	TargetType string    `json:"target_type"` // "discord", "slack", "generic"
	URL        string    `json:"url"`
	Events     string    `json:"events"`      // "regression", "failure", "all"
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Store) AutoMigrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			plan TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS api_tokens (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			token_hash TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (org_id) REFERENCES organizations(id)
		);`,
		`CREATE TABLE IF NOT EXISTS repositories (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			full_name TEXT NOT NULL UNIQUE,
			default_branch TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (org_id) REFERENCES organizations(id)
		);`,
		`CREATE TABLE IF NOT EXISTS scenarios (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			name TEXT NOT NULL,
			driver TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE(repo_id, name),
			FOREIGN KEY (repo_id) REFERENCES repositories(id)
		);`,
		`CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			scenario_id TEXT NOT NULL,
			commit_sha TEXT NOT NULL,
			branch TEXT NOT NULL,
			pr_number INTEGER NOT NULL,
			status TEXT NOT NULL,
			anomaly_type TEXT,
			seed INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (repo_id) REFERENCES repositories(id),
			FOREIGN KEY (scenario_id) REFERENCES scenarios(id)
		);`,
		`CREATE TABLE IF NOT EXISTS findings (
			id TEXT PRIMARY KEY,
			run_id TEXT NOT NULL,
			anomaly_type TEXT NOT NULL,
			assertion TEXT,
			minimal_ops INTEGER NOT NULL,
			repro_code TEXT,
			trace_json TEXT,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (run_id) REFERENCES runs(id)
		);`,
				`CREATE TABLE IF NOT EXISTS webhooks (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			target_type TEXT NOT NULL,
			url TEXT NOT NULL,
			events TEXT NOT NULL,
			active INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL,
			FOREIGN KEY (org_id) REFERENCES organizations(id)
		);`,
`CREATE TABLE IF NOT EXISTS baselines (
			id TEXT PRIMARY KEY,
			repo_id TEXT NOT NULL,
			scenario_id TEXT NOT NULL,
			branch TEXT NOT NULL,
			run_id TEXT NOT NULL,
			updated_at DATETIME NOT NULL,
			UNIQUE(repo_id, scenario_id, branch),
			FOREIGN KEY (repo_id) REFERENCES repositories(id),
			FOREIGN KEY (scenario_id) REFERENCES scenarios(id),
			FOREIGN KEY (run_id) REFERENCES runs(id)
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("migration failed for query [%s]: %w", q, err)
		}
	}
	return nil
}

func (s *Store) CreateOrganization(id, name, plan string) error {
	_, err := s.db.Exec(`INSERT INTO organizations (id, name, plan, created_at) VALUES (?, ?, ?, ?)`,
		id, name, plan, time.Now().UTC())
	return err
}

func (s *Store) CreateAPIToken(id, orgID, token, name string) error {
	th := hashToken(token)
	_, err := s.db.Exec(`INSERT INTO api_tokens (id, org_id, token_hash, name, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, orgID, th, name, time.Now().UTC())
	return err
}

func (s *Store) ValidateToken(token string) (string, error) {
	th := hashToken(token)
	var orgID string
	err := s.db.QueryRow(`SELECT org_id FROM api_tokens WHERE token_hash = ?`, th).Scan(&orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUnauthorized
	}
	return orgID, err
}

func (s *Store) GetOrCreateRepo(orgID, fullName, defaultBranch string) (*Repository, error) {
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	var repo Repository
	err := s.db.QueryRow(`SELECT id, org_id, full_name, default_branch, created_at FROM repositories WHERE full_name = ?`, fullName).
		Scan(&repo.ID, &repo.OrgID, &repo.FullName, &repo.DefaultBranch, &repo.CreatedAt)
	if err == nil {
		return &repo, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	id := fmt.Sprintf("repo_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	_, err = s.db.Exec(`INSERT INTO repositories (id, org_id, full_name, default_branch, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, orgID, fullName, defaultBranch, now)
	if err != nil {
		return nil, err
	}

	return &Repository{
		ID:            id,
		OrgID:         orgID,
		FullName:      fullName,
		DefaultBranch: defaultBranch,
		CreatedAt:     now,
	}, nil
}

func (s *Store) GetOrCreateScenario(repoID, name, driver string) (*Scenario, error) {
	var sc Scenario
	err := s.db.QueryRow(`SELECT id, repo_id, name, driver, created_at FROM scenarios WHERE repo_id = ? AND name = ?`, repoID, name).
		Scan(&sc.ID, &sc.RepoID, &sc.Name, &sc.Driver, &sc.CreatedAt)
	if err == nil {
		return &sc, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	id := fmt.Sprintf("sc_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	_, err = s.db.Exec(`INSERT INTO scenarios (id, repo_id, name, driver, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, repoID, name, driver, now)
	if err != nil {
		return nil, err
	}

	return &Scenario{
		ID:        id,
		RepoID:    repoID,
		Name:      name,
		Driver:    driver,
		CreatedAt: now,
	}, nil
}

func (s *Store) SaveRun(run *RunRecord, finding *FindingRecord) error {
	_, err := s.db.Exec(`INSERT INTO runs (id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.RepoID, run.ScenarioID, run.CommitSHA, run.Branch, run.PRNumber, run.Status, run.AnomalyType, run.Seed, run.DurationMS, run.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert run: %w", err)
	}

	if finding != nil {
		_, err = s.db.Exec(`INSERT INTO findings (id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			finding.ID, finding.RunID, finding.AnomalyType, finding.Assertion, finding.MinimalOps, finding.ReproCode, finding.TraceJSON, finding.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert finding: %w", err)
		}
	}
	return nil
}

func (s *Store) GetRun(runID string) (*RunRecord, *FindingRecord, error) {
	var run RunRecord
	err := s.db.QueryRow(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at
		FROM runs WHERE id = ?`, runID).
		Scan(&run.ID, &run.RepoID, &run.ScenarioID, &run.CommitSHA, &run.Branch, &run.PRNumber, &run.Status, &run.AnomalyType, &run.Seed, &run.DurationMS, &run.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	var finding FindingRecord
	fErr := s.db.QueryRow(`SELECT id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at
		FROM findings WHERE run_id = ?`, runID).
		Scan(&finding.ID, &finding.RunID, &finding.AnomalyType, &finding.Assertion, &finding.MinimalOps, &finding.ReproCode, &finding.TraceJSON, &finding.CreatedAt)
	if errors.Is(fErr, sql.ErrNoRows) {
		return &run, nil, nil
	}
	if fErr != nil {
		return &run, nil, fErr
	}

	return &run, &finding, nil
}

func (s *Store) GetLatestBaseline(repoID, scenarioID, branch string) (*RunRecord, error) {
	var runID string
	err := s.db.QueryRow(`SELECT run_id FROM baselines WHERE repo_id = ? AND scenario_id = ? AND branch = ?`,
		repoID, scenarioID, branch).Scan(&runID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	run, _, err := s.GetRun(runID)
	return run, err
}

func (s *Store) SetBaseline(repoID, scenarioID, branch, runID string) error {
	id := fmt.Sprintf("base_%s_%s_%s", repoID, scenarioID, branch)
	now := time.Now().UTC()

	_, err := s.db.Exec(`INSERT INTO baselines (id, repo_id, scenario_id, branch, run_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, scenario_id, branch) DO UPDATE SET run_id = excluded.run_id, updated_at = excluded.updated_at`,
		id, repoID, scenarioID, branch, runID, now)
	return err
}

func (s *Store) ListRuns(repoID string, limit, offset int) ([]*RunRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at
		FROM runs WHERE repo_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, repoID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*RunRecord
	for rows.Next() {
		var r RunRecord
		if err := rows.Scan(&r.ID, &r.RepoID, &r.ScenarioID, &r.CommitSHA, &r.Branch, &r.PRNumber, &r.Status, &r.AnomalyType, &r.Seed, &r.DurationMS, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}

func (s *Store) ListRecentRuns(limit, offset int) ([]*RunRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at
		FROM runs ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*RunRecord
	for rows.Next() {
		var r RunRecord
		if err := rows.Scan(&r.ID, &r.RepoID, &r.ScenarioID, &r.CommitSHA, &r.Branch, &r.PRNumber, &r.Status, &r.AnomalyType, &r.Seed, &r.DurationMS, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &r)
	}
	return results, rows.Err()
}

func (s *Store) CreateWebhook(ctx context.Context, w *WebhookRecord) error {
	if w.CreatedAt.IsZero() {
		w.CreatedAt = time.Now().UTC()
	}
	activeInt := 1
	if !w.Active {
		activeInt = 0
	}
	query := `INSERT INTO webhooks (id, org_id, target_type, url, events, active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, w.ID, w.OrgID, w.TargetType, w.URL, w.Events, activeInt, w.CreatedAt)
	return err
}

func (s *Store) ListWebhooks(ctx context.Context, orgID string) ([]WebhookRecord, error) {
	query := `SELECT id, org_id, target_type, url, events, active, created_at
		FROM webhooks WHERE org_id = ? ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var list []WebhookRecord
	for rows.Next() {
		var w WebhookRecord
		var activeInt int
		if err := rows.Scan(&w.ID, &w.OrgID, &w.TargetType, &w.URL, &w.Events, &activeInt, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.Active = activeInt == 1
		list = append(list, w)
	}
	return list, rows.Err()
}

func (s *Store) DeleteWebhook(ctx context.Context, orgID, webhookID string) error {
	query := `DELETE FROM webhooks WHERE id = ? AND org_id = ?`
	_, err := s.db.ExecContext(ctx, query, webhookID, orgID)
	return err
}

func (s *Store) GetActiveWebhooksForEvent(ctx context.Context, orgID, eventType string) ([]WebhookRecord, error) {
	all, err := s.ListWebhooks(ctx, orgID)
	if err != nil {
		return nil, err
	}
	var matched []WebhookRecord
	for _, w := range all {
		if !w.Active {
			continue
		}
		eventsList := strings.Split(w.Events, ",")
		for _, e := range eventsList {
			clean := strings.TrimSpace(e)
			if clean == "all" || clean == eventType {
				matched = append(matched, w)
				break
			}
		}
	}
	return matched, nil
}
