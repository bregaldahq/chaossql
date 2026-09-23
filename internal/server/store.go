package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("unauthorized token")
	ErrInvalidRole  = errors.New("invalid organization role")
)

type Role string

const (
	RoleMember Role = "member"
	RoleAdmin  Role = "admin"
	RoleOwner  Role = "owner"
)

func (r Role) valid() bool {
	return r == RoleMember || r == RoleAdmin || r == RoleOwner
}

func (r Role) Allows(required Role) bool {
	rank := map[Role]int{RoleMember: 1, RoleAdmin: 2, RoleOwner: 3}
	return r.valid() && required.valid() && rank[r] >= rank[required]
}

type Principal struct {
	TokenID string
	OrgID   string
	Role    Role
}

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
	ID                  string     `json:"id"`
	RepoID              string     `json:"repo_id"`
	ScenarioID          string     `json:"scenario_id"`
	CommitSHA           string     `json:"commit_sha"`
	Branch              string     `json:"branch"`
	PRNumber            int        `json:"pr_number"`
	Status              string     `json:"status"` // "passed" or "failed"
	AnomalyType         string     `json:"anomaly_type"`
	Seed                uint64     `json:"seed"`
	DurationMS          int64      `json:"duration_ms"`
	CreatedAt           time.Time  `json:"created_at"`
	IdempotencyKey      string     `json:"idempotency_key,omitempty"`
	ScenarioFingerprint string     `json:"scenario_fingerprint,omitempty"`
	CommitTimestamp     *time.Time `json:"commit_timestamp,omitempty"`
}

type OutboxItem struct {
	ID          string     `json:"id"`
	OrgID       string     `json:"org_id"`
	WebhookID   string     `json:"webhook_id"`
	EventType   string     `json:"event_type"`
	PayloadJSON string     `json:"payload_json"`
	Status      string     `json:"status"` // pending, delivered, failed
	Attempts    int        `json:"attempts"`
	NextRetryAt time.Time  `json:"next_retry_at"`
	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	LastError   string     `json:"last_error,omitempty"`
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
	Events     string    `json:"events"` // "regression", "failure", "all"
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
}

type Store struct {
	db *sql.DB
	tx *sql.Tx
}

func NewStore(db *sql.DB) *Store {
	// SQLite foreign-key pragmas are connection-local. Keep the control-plane
	// store on one connection so migrations and every subsequent query share
	// the same enforcement mode.
	db.SetMaxOpenConns(1)
	return &Store{db: db}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (s *Store) AutoMigrate() error {
	queries := []string{
		`PRAGMA foreign_keys = ON;`,
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
			role TEXT NOT NULL DEFAULT 'member',
			created_at DATETIME NOT NULL,
			FOREIGN KEY (org_id) REFERENCES organizations(id)
		);`,
		`CREATE TABLE IF NOT EXISTS repositories (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			full_name TEXT NOT NULL,
			default_branch TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE(org_id, full_name),
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
		`CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL
		);`,
	}

	for _, q := range queries {
		if _, err := s.queryer().Exec(q); err != nil {
			return fmt.Errorf("migration failed for query [%s]: %w", q, err)
		}
	}
	if err := s.ensureAPITokenRoleColumn(); err != nil {
		return err
	}
	if err := s.ensureTenantRepositoryUniqueness(); err != nil {
		return err
	}
	if err := s.purgeLegacyFindingDetails(); err != nil {
		return err
	}
	if err := s.ensureIdempotentRunsAndOutbox(); err != nil {
		return err
	}
	return s.ensureIngestionMetadata()
}

func (s *Store) purgeLegacyFindingDetails() error {
	const migration = "2026-09-16-purge-hosted-finding-details"
	var applied int
	if err := s.queryer().QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, migration).Scan(&applied); err != nil {
		return fmt.Errorf("inspect finding privacy migration: %w", err)
	}
	if applied != 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin finding privacy migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`UPDATE findings SET assertion = '', repro_code = '', trace_json = ''`); err != nil {
		return fmt.Errorf("purge legacy finding details: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, migration, time.Now().UTC()); err != nil {
		return fmt.Errorf("record finding privacy migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit finding privacy migration: %w", err)
	}
	return nil
}

func (s *Store) ensureIdempotentRunsAndOutbox() error {
	const migration = "2026-09-17-idempotent-runs-and-outbox"
	var applied int
	if err := s.queryer().QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, migration).Scan(&applied); err != nil {
		return fmt.Errorf("inspect outbox migration: %w", err)
	}
	if applied != 0 {
		return nil
	}

	rows, err := s.queryer().Query(`PRAGMA table_info(runs)`)
	if err != nil {
		return fmt.Errorf("inspect runs table info: %w", err)
	}
	cols := make(map[string]bool)
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan runs column: %w", err)
		}
		cols[name] = true
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close runs columns query: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin outbox migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if !cols["idempotency_key"] {
		if _, err := tx.Exec(`ALTER TABLE runs ADD COLUMN idempotency_key TEXT`); err != nil {
			return fmt.Errorf("add idempotency_key: %w", err)
		}
	}
	if !cols["scenario_fingerprint"] {
		if _, err := tx.Exec(`ALTER TABLE runs ADD COLUMN scenario_fingerprint TEXT NOT NULL DEFAULT ''`); err != nil {
			return fmt.Errorf("add scenario_fingerprint: %w", err)
		}
	}
	if !cols["commit_timestamp"] {
		if _, err := tx.Exec(`ALTER TABLE runs ADD COLUMN commit_timestamp DATETIME`); err != nil {
			return fmt.Errorf("add commit_timestamp: %w", err)
		}
	}

	if _, err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_runs_repo_idempotency ON runs(repo_id, idempotency_key) WHERE idempotency_key IS NOT NULL`); err != nil {
		return fmt.Errorf("create idempotency index: %w", err)
	}

	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS webhook_outbox (
		id TEXT PRIMARY KEY,
		org_id TEXT NOT NULL,
		webhook_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		payload_json TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		next_retry_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		delivered_at DATETIME,
		last_error TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (org_id) REFERENCES organizations(id),
		FOREIGN KEY (webhook_id) REFERENCES webhooks(id)
	)`); err != nil {
		return fmt.Errorf("create webhook_outbox table: %w", err)
	}

	if _, err := tx.Exec(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, migration, time.Now().UTC()); err != nil {
		return fmt.Errorf("record outbox migration: %w", err)
	}

	return tx.Commit()
}

func (s *Store) ensureAPITokenRoleColumn() error {
	rows, err := s.queryer().Query(`PRAGMA table_info(api_tokens)`)
	if err != nil {
		return fmt.Errorf("inspect api token schema: %w", err)
	}
	hasRole := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			_ = rows.Close()
			return fmt.Errorf("inspect api token column: %w", err)
		}
		if name == "role" {
			hasRole = true
		}
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close api token schema inspection: %w", err)
	}
	if hasRole {
		return nil
	}
	if _, err := s.queryer().Exec(`ALTER TABLE api_tokens ADD COLUMN role TEXT NOT NULL DEFAULT 'member'`); err != nil {
		return fmt.Errorf("add api token role: %w", err)
	}
	return nil
}

func (s *Store) ensureTenantRepositoryUniqueness() error {
	ctx := context.Background()
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire repository migration connection: %w", err)
	}
	defer func() { _ = conn.Close() }()

	var schema string
	if err := conn.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'repositories'`).Scan(&schema); err != nil {
		return fmt.Errorf("inspect repository schema: %w", err)
	}
	normalized := strings.ToLower(strings.Join(strings.Fields(schema), " "))
	if strings.Contains(normalized, "unique(org_id, full_name)") || strings.Contains(normalized, "unique (org_id, full_name)") {
		return nil
	}

	var foreignKeys int
	if err := conn.QueryRowContext(ctx, `PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil {
		return fmt.Errorf("inspect foreign key mode: %w", err)
	}
	if foreignKeys != 0 {
		if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
			return fmt.Errorf("disable foreign keys for repository migration: %w", err)
		}
		defer func() { _, _ = conn.ExecContext(ctx, `PRAGMA foreign_keys = ON`) }()
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin repository migration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	statements := []string{
		`DROP TABLE IF EXISTS repositories_tenant_migration`,
		`CREATE TABLE repositories_tenant_migration (
			id TEXT PRIMARY KEY,
			org_id TEXT NOT NULL,
			full_name TEXT NOT NULL,
			default_branch TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			UNIQUE(org_id, full_name),
			FOREIGN KEY (org_id) REFERENCES organizations(id)
		)`,
		`INSERT INTO repositories_tenant_migration (id, org_id, full_name, default_branch, created_at)
			SELECT id, org_id, full_name, default_branch, created_at FROM repositories`,
		`DROP TABLE repositories`,
		`ALTER TABLE repositories_tenant_migration RENAME TO repositories`,
	}
	for _, statement := range statements {
		if _, err := tx.Exec(statement); err != nil {
			return fmt.Errorf("migrate repository uniqueness: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit repository uniqueness migration: %w", err)
	}
	return nil
}

func (s *Store) CreateOrganization(id, name, plan string) error {
	_, err := s.queryer().Exec(`INSERT INTO organizations (id, name, plan, created_at) VALUES (?, ?, ?, ?)`,
		id, name, plan, time.Now().UTC())
	return err
}

func (s *Store) EnsureOrganization(id, name, plan string) error {
	_, err := s.queryer().Exec(`INSERT INTO organizations (id, name, plan, created_at) VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING`, id, name, plan, time.Now().UTC())
	return err
}

func (s *Store) CreateAPIToken(id, orgID, token, name string) error {
	return s.CreateAPITokenWithRole(id, orgID, token, name, RoleMember)
}

func (s *Store) CreateAPITokenWithRole(id, orgID, token, name string, role Role) error {
	if !role.valid() {
		return fmt.Errorf("%w: %q", ErrInvalidRole, role)
	}
	th := hashToken(token)
	_, err := s.queryer().Exec(`INSERT INTO api_tokens (id, org_id, token_hash, name, role, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id, orgID, th, name, role, time.Now().UTC())
	return err
}

func (s *Store) UpsertAPITokenWithRole(id, orgID, token, name string, role Role) error {
	if !role.valid() {
		return fmt.Errorf("%w: %q", ErrInvalidRole, role)
	}
	_, err := s.queryer().Exec(`INSERT INTO api_tokens (id, org_id, token_hash, name, role, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			org_id = excluded.org_id,
			token_hash = excluded.token_hash,
			name = excluded.name,
			role = excluded.role`,
		id, orgID, hashToken(token), name, role, time.Now().UTC())
	return err
}

func (s *Store) ValidateToken(token string) (string, error) {
	principal, err := s.AuthenticateToken(token)
	if err != nil {
		return "", err
	}
	return principal.OrgID, nil
}

func (s *Store) AuthenticateToken(token string) (Principal, error) {
	th := hashToken(token)
	var principal Principal
	err := s.queryer().QueryRow(`SELECT id, org_id, role FROM api_tokens WHERE token_hash = ?`, th).
		Scan(&principal.TokenID, &principal.OrgID, &principal.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return Principal{}, ErrUnauthorized
	}
	if err != nil {
		return Principal{}, err
	}
	if !principal.Role.valid() {
		return Principal{}, fmt.Errorf("%w: stored role %q", ErrInvalidRole, principal.Role)
	}
	return principal, nil
}

func (s *Store) GetOrCreateRepo(orgID, fullName, defaultBranch string) (*Repository, error) {
	if s.tx == nil {
		var repo *Repository
		err := s.withTransaction(context.Background(), func(bound *Store) error {
			var err error
			repo, err = bound.GetOrCreateRepo(orgID, fullName, defaultBranch)
			return err
		})
		return repo, err
	}

	if defaultBranch == "" {
		defaultBranch = "main"
	}

	var repo Repository
	err := s.queryer().QueryRow(`SELECT id, org_id, full_name, default_branch, created_at FROM repositories WHERE org_id = ? AND full_name = ?`, orgID, fullName).
		Scan(&repo.ID, &repo.OrgID, &repo.FullName, &repo.DefaultBranch, &repo.CreatedAt)
	if err == nil {
		return &repo, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	subscription, err := s.GetOrgSubscription(orgID)
	if err != nil {
		return nil, err
	}
	if !subscription.CanAddRepository {
		return nil, ErrPlanLimitReached
	}
	id := fmt.Sprintf("repo_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	_, err = s.queryer().Exec(`INSERT INTO repositories (id, org_id, full_name, default_branch, created_at) VALUES (?, ?, ?, ?, ?)`,
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

func (s *Store) GetRepositoryByFullName(orgID, fullName string) (*Repository, error) {
	var repo Repository
	err := s.queryer().QueryRow(`SELECT id, org_id, full_name, default_branch, created_at
		FROM repositories WHERE org_id = ? AND full_name = ?`, orgID, fullName).
		Scan(&repo.ID, &repo.OrgID, &repo.FullName, &repo.DefaultBranch, &repo.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *Store) GetOrCreateScenario(repoID, name, driver string) (*Scenario, error) {
	var sc Scenario
	err := s.queryer().QueryRow(`SELECT id, repo_id, name, driver, created_at FROM scenarios WHERE repo_id = ? AND name = ?`, repoID, name).
		Scan(&sc.ID, &sc.RepoID, &sc.Name, &sc.Driver, &sc.CreatedAt)
	if err == nil {
		return &sc, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	id := fmt.Sprintf("sc_%d", time.Now().UnixNano())
	now := time.Now().UTC()
	_, err = s.queryer().Exec(`INSERT INTO scenarios (id, repo_id, name, driver, created_at) VALUES (?, ?, ?, ?, ?)`,
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
	_, _, err := s.SaveRunTx(context.Background(), run, finding, nil)
	return err
}

func (s *Store) SaveRunTx(ctx context.Context, run *RunRecord, finding *FindingRecord, outbox []*OutboxItem) (*RunRecord, bool, error) {
	if run == nil {
		return nil, false, errors.New("cannot save nil run")
	}

	tx, finish, err := s.beginOperation(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("begin save run tx: %w", err)
	}
	defer func() {
		if finish {
			_ = tx.Rollback()
		}
	}()

	if run.IdempotencyKey != "" {
		var existing RunRecord
		var idempotencyKey, scenarioFingerprint sql.NullString
		var commitTimestamp sql.NullTime
		err := tx.QueryRowContext(ctx, `SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number,
			status, anomaly_type, seed, duration_ms, created_at, idempotency_key, scenario_fingerprint, commit_timestamp
			FROM runs WHERE repo_id = ? AND idempotency_key = ?`, run.RepoID, run.IdempotencyKey).
			Scan(&existing.ID, &existing.RepoID, &existing.ScenarioID, &existing.CommitSHA, &existing.Branch,
				&existing.PRNumber, &existing.Status, &existing.AnomalyType, &existing.Seed, &existing.DurationMS,
				&existing.CreatedAt, &idempotencyKey, &scenarioFingerprint, &commitTimestamp)
		if err == nil {
			if idempotencyKey.Valid {
				existing.IdempotencyKey = idempotencyKey.String
			}
			if scenarioFingerprint.Valid {
				existing.ScenarioFingerprint = scenarioFingerprint.String
			}
			if commitTimestamp.Valid {
				t := commitTimestamp.Time.UTC()
				existing.CommitTimestamp = &t
			}
			return &existing, false, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, false, fmt.Errorf("query existing idempotency run: %w", err)
		}
	}

	var nullKey any
	if run.IdempotencyKey != "" {
		nullKey = run.IdempotencyKey
	}
	var commitTime any
	if run.CommitTimestamp != nil {
		commitTime = run.CommitTimestamp.UTC()
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO runs (id, repo_id, scenario_id, commit_sha, branch, pr_number,
		status, anomaly_type, seed, duration_ms, created_at, idempotency_key, scenario_fingerprint, commit_timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.RepoID, run.ScenarioID, run.CommitSHA, run.Branch, run.PRNumber,
		run.Status, run.AnomalyType, run.Seed, run.DurationMS, run.CreatedAt.UTC(),
		nullKey, run.ScenarioFingerprint, commitTime)
	if err != nil {
		return nil, false, fmt.Errorf("insert run: %w", err)
	}

	if finding != nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO findings (id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			finding.ID, finding.RunID, finding.AnomalyType, finding.Assertion, finding.MinimalOps, finding.ReproCode, finding.TraceJSON, finding.CreatedAt.UTC())
		if err != nil {
			return nil, false, fmt.Errorf("insert finding: %w", err)
		}
	}

	for _, item := range outbox {
		if item == nil {
			continue
		}
		status := item.Status
		if status == "" {
			status = "pending"
		}
		nextRetry := item.NextRetryAt
		if nextRetry.IsZero() {
			nextRetry = time.Now().UTC()
		}
		createdAt := item.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		var delTime any
		if item.DeliveredAt != nil {
			delTime = item.DeliveredAt.UTC()
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO webhook_outbox (id, org_id, webhook_id, event_type, payload_json, status, attempts, next_retry_at, created_at, delivered_at, last_error)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.OrgID, item.WebhookID, item.EventType, item.PayloadJSON, status, item.Attempts, nextRetry.UTC(), createdAt.UTC(), delTime, item.LastError)
		if err != nil {
			return nil, false, fmt.Errorf("insert outbox item %s: %w", item.ID, err)
		}
	}

	if finish {
		if err := tx.Commit(); err != nil {
			return nil, false, fmt.Errorf("commit save run tx: %w", err)
		}
	}

	return run, true, nil
}

func (s *Store) EnqueueOutbox(ctx context.Context, items []*OutboxItem) error {
	if len(items) == 0 {
		return nil
	}
	tx, finish, err := s.beginOperation(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if finish {
			_ = tx.Rollback()
		}
	}()

	for _, item := range items {
		if item == nil {
			continue
		}
		status := item.Status
		if status == "" {
			status = "pending"
		}
		nextRetry := item.NextRetryAt
		if nextRetry.IsZero() {
			nextRetry = time.Now().UTC()
		}
		createdAt := item.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}
		var delTime any
		if item.DeliveredAt != nil {
			delTime = item.DeliveredAt.UTC()
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO webhook_outbox (id, org_id, webhook_id, event_type, payload_json, status, attempts, next_retry_at, created_at, delivered_at, last_error)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID, item.OrgID, item.WebhookID, item.EventType, item.PayloadJSON, status, item.Attempts, nextRetry.UTC(), createdAt.UTC(), delTime, item.LastError)
		if err != nil {
			return fmt.Errorf("insert outbox item %s: %w", item.ID, err)
		}
	}
	if finish {
		return tx.Commit()
	}
	return nil
}

func (s *Store) GetPendingOutboxItems(ctx context.Context, limit int) ([]*OutboxItem, error) {
	if limit <= 0 {
		limit = 20
	}
	now := time.Now().UTC()
	rows, err := s.queryer().QueryContext(ctx, `SELECT id, org_id, webhook_id, event_type, payload_json, status, attempts, next_retry_at, created_at, delivered_at, last_error
		FROM webhook_outbox
		WHERE status = 'pending' AND next_retry_at <= ?
		ORDER BY created_at ASC LIMIT ?`, now, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []*OutboxItem
	for rows.Next() {
		var it OutboxItem
		var deliveredAt sql.NullTime
		if err := rows.Scan(&it.ID, &it.OrgID, &it.WebhookID, &it.EventType, &it.PayloadJSON, &it.Status, &it.Attempts, &it.NextRetryAt, &it.CreatedAt, &deliveredAt, &it.LastError); err != nil {
			return nil, err
		}
		if deliveredAt.Valid {
			t := deliveredAt.Time.UTC()
			it.DeliveredAt = &t
		}
		items = append(items, &it)
	}
	return items, rows.Err()
}

func (s *Store) MarkOutboxDelivered(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := s.queryer().ExecContext(ctx, `UPDATE webhook_outbox SET status = 'delivered', delivered_at = ?, last_error = '' WHERE id = ?`, now, id)
	return err
}

func (s *Store) MarkOutboxAttemptFailed(ctx context.Context, id string, lastErr string, maxAttempts int) error {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	now := time.Now().UTC()
	var currentAttempts int
	err := s.queryer().QueryRowContext(ctx, `SELECT attempts FROM webhook_outbox WHERE id = ?`, id).Scan(&currentAttempts)
	if err != nil {
		return err
	}
	newAttempts := currentAttempts + 1
	status := "pending"
	backoffSec := 2 * (1 << (newAttempts - 1))
	if newAttempts >= maxAttempts {
		status = "failed"
	}
	nextRetry := now.Add(time.Duration(backoffSec) * time.Second)
	_, err = s.queryer().ExecContext(ctx, `UPDATE webhook_outbox SET attempts = ?, status = ?, next_retry_at = ?, last_error = ? WHERE id = ?`,
		newAttempts, status, nextRetry, lastErr, id)
	return err
}

func (s *Store) GetWebhookByID(ctx context.Context, orgID, webhookID string) (*WebhookRecord, error) {
	var w WebhookRecord
	var activeInt int
	err := s.queryer().QueryRowContext(ctx, `SELECT id, org_id, target_type, url, events, active, created_at
		FROM webhooks WHERE org_id = ? AND id = ?`, orgID, webhookID).
		Scan(&w.ID, &w.OrgID, &w.TargetType, &w.URL, &w.Events, &activeInt, &w.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	w.Active = activeInt == 1
	return &w, nil
}

func (s *Store) GetRun(runID string) (*RunRecord, *FindingRecord, error) {
	var run RunRecord
	var idempotencyKey, scenarioFingerprint sql.NullString
	var commitTimestamp sql.NullTime
	err := s.queryer().QueryRow(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at,
		idempotency_key, scenario_fingerprint, commit_timestamp
		FROM runs WHERE id = ?`, runID).
		Scan(&run.ID, &run.RepoID, &run.ScenarioID, &run.CommitSHA, &run.Branch, &run.PRNumber, &run.Status, &run.AnomalyType, &run.Seed, &run.DurationMS, &run.CreatedAt,
			&idempotencyKey, &scenarioFingerprint, &commitTimestamp)
	if err == nil {
		if idempotencyKey.Valid {
			run.IdempotencyKey = idempotencyKey.String
		}
		if scenarioFingerprint.Valid {
			run.ScenarioFingerprint = scenarioFingerprint.String
		}
		if commitTimestamp.Valid {
			t := commitTimestamp.Time.UTC()
			run.CommitTimestamp = &t
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	var finding FindingRecord
	fErr := s.queryer().QueryRow(`SELECT id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at
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

func (s *Store) GetRunForOrg(orgID, runID string) (*RunRecord, *FindingRecord, error) {
	var run RunRecord
	var idempotencyKey, scenarioFingerprint sql.NullString
	var commitTimestamp sql.NullTime
	err := s.queryer().QueryRow(`SELECT r.id, r.repo_id, r.scenario_id, r.commit_sha, r.branch, r.pr_number,
		r.status, r.anomaly_type, r.seed, r.duration_ms, r.created_at,
		r.idempotency_key, r.scenario_fingerprint, r.commit_timestamp
		FROM runs r
		JOIN repositories repo ON repo.id = r.repo_id
		WHERE r.id = ? AND repo.org_id = ?`, runID, orgID).
		Scan(&run.ID, &run.RepoID, &run.ScenarioID, &run.CommitSHA, &run.Branch, &run.PRNumber,
			&run.Status, &run.AnomalyType, &run.Seed, &run.DurationMS, &run.CreatedAt,
			&idempotencyKey, &scenarioFingerprint, &commitTimestamp)
	if err == nil {
		if idempotencyKey.Valid {
			run.IdempotencyKey = idempotencyKey.String
		}
		if scenarioFingerprint.Valid {
			run.ScenarioFingerprint = scenarioFingerprint.String
		}
		if commitTimestamp.Valid {
			t := commitTimestamp.Time.UTC()
			run.CommitTimestamp = &t
		}
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	return s.findingForRun(&run)
}

func (s *Store) findingForRun(run *RunRecord) (*RunRecord, *FindingRecord, error) {
	var finding FindingRecord
	err := s.queryer().QueryRow(`SELECT id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at
		FROM findings WHERE run_id = ?`, run.ID).
		Scan(&finding.ID, &finding.RunID, &finding.AnomalyType, &finding.Assertion, &finding.MinimalOps, &finding.ReproCode, &finding.TraceJSON, &finding.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return run, nil, nil
	}
	if err != nil {
		return run, nil, err
	}
	return run, &finding, nil
}

func (s *Store) GetLatestBaseline(repoID, scenarioID, branch string) (*RunRecord, error) {
	var runID string
	err := s.queryer().QueryRow(`SELECT run_id FROM baselines WHERE repo_id = ? AND scenario_id = ? AND branch = ?`,
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

	_, err := s.queryer().Exec(`INSERT INTO baselines (id, repo_id, scenario_id, branch, run_id, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo_id, scenario_id, branch) DO UPDATE SET run_id = excluded.run_id, updated_at = excluded.updated_at`,
		id, repoID, scenarioID, branch, runID, now)
	return err
}

func (s *Store) ListRuns(repoID string, limit, offset int) ([]*RunRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.queryer().Query(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at
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
	rows, err := s.queryer().Query(`SELECT id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at
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

func (s *Store) ListRecentRunsForOrg(orgID string, limit, offset int) ([]*RunRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.queryer().Query(`SELECT r.id, r.repo_id, r.scenario_id, r.commit_sha, r.branch, r.pr_number,
		r.status, r.anomaly_type, r.seed, r.duration_ms, r.created_at
		FROM runs r
		JOIN repositories repo ON repo.id = r.repo_id
		WHERE repo.org_id = ?
		ORDER BY r.created_at DESC LIMIT ? OFFSET ?`, orgID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*RunRecord
	for rows.Next() {
		var run RunRecord
		if err := rows.Scan(&run.ID, &run.RepoID, &run.ScenarioID, &run.CommitSHA, &run.Branch,
			&run.PRNumber, &run.Status, &run.AnomalyType, &run.Seed, &run.DurationMS, &run.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &run)
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
	_, err := s.queryer().ExecContext(ctx, query, w.ID, w.OrgID, w.TargetType, w.URL, w.Events, activeInt, w.CreatedAt)
	return err
}

func (s *Store) ListWebhooks(ctx context.Context, orgID string) ([]WebhookRecord, error) {
	query := `SELECT id, org_id, target_type, url, events, active, created_at
		FROM webhooks WHERE org_id = ? ORDER BY created_at DESC`
	rows, err := s.queryer().QueryContext(ctx, query, orgID)
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
	if s.tx == nil {
		return s.withTransaction(ctx, func(bound *Store) error { return bound.DeleteWebhook(ctx, orgID, webhookID) })
	}
	if _, err := s.GetWebhookByID(ctx, orgID, webhookID); err != nil {
		return err
	}
	// A removed destination must not retain queued deliveries. Delete its
	// dependent records in the same tenant-scoped transaction.
	if _, err := s.tx.ExecContext(ctx, `DELETE FROM webhook_outbox WHERE webhook_id=? AND org_id=?`, webhookID, orgID); err != nil {
		return err
	}
	query := `DELETE FROM webhooks WHERE id = ? AND org_id = ?`
	result, err := s.queryer().ExecContext(ctx, query, webhookID, orgID)
	if err != nil {
		return err
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return ErrNotFound
	}
	return nil
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
