package server

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestStore(t *testing.T) *Store {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}
	db.SetMaxOpenConns(1)
	s := NewStore(db)
	if err := s.AutoMigrate(); err != nil {
		t.Fatalf("failed to run automigrate: %v", err)
	}
	return s
}

func TestStoreTokenAndOrganization(t *testing.T) {
	s := newTestStore(t)

	orgID := "org_bregalda_01"
	if err := s.CreateOrganization(orgID, "Bregalda", "pro"); err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	token := "chaossql_secret_token_abc"
	if err := s.CreateAPIToken("tok_01", orgID, token, "CI token"); err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	gotOrg, err := s.ValidateToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if gotOrg != orgID {
		t.Errorf("expected org %s, got %s", orgID, gotOrg)
	}

	_, err = s.ValidateToken("invalid_token")
	if err != ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestStoreAuthenticateTokenReturnsPrincipalRole(t *testing.T) {
	s := newTestStore(t)
	if err := s.CreateOrganization("org_auth", "Auth Org", "team"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAPITokenWithRole("tok_admin", "org_auth", "secret", "Admin", RoleAdmin); err != nil {
		t.Fatal(err)
	}

	principal, err := s.AuthenticateToken("secret")
	if err != nil {
		t.Fatal(err)
	}
	if principal.TokenID != "tok_admin" || principal.OrgID != "org_auth" || principal.Role != RoleAdmin {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestStoreRejectsInvalidTokenRole(t *testing.T) {
	s := newTestStore(t)
	if err := s.CreateOrganization("org_auth", "Auth Org", "team"); err != nil {
		t.Fatal(err)
	}
	err := s.CreateAPITokenWithRole("tok_bad", "org_auth", "secret", "Bad", Role("superuser"))
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("error = %v, want ErrInvalidRole", err)
	}
}

func TestStoreUpsertsAPITokenRoleAndCredential(t *testing.T) {
	s := newTestStore(t)
	if err := s.CreateOrganization("org_auth", "Auth Org", "team"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateAPIToken("tok_admin", "org_auth", "old-secret", "Legacy"); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertAPITokenWithRole("tok_admin", "org_auth", "new-secret", "Bootstrap", RoleOwner); err != nil {
		t.Fatal(err)
	}
	if _, err := s.AuthenticateToken("old-secret"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("old credential remains active: %v", err)
	}
	principal, err := s.AuthenticateToken("new-secret")
	if err != nil || principal.Role != RoleOwner {
		t.Fatalf("upserted principal=%+v error=%v", principal, err)
	}
}

func TestAutoMigrateAddsMemberRoleToLegacyTokens(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE organizations (id TEXT PRIMARY KEY, name TEXT NOT NULL, plan TEXT NOT NULL, created_at DATETIME NOT NULL);
		CREATE TABLE api_tokens (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, token_hash TEXT NOT NULL UNIQUE, name TEXT NOT NULL, created_at DATETIME NOT NULL);
		INSERT INTO organizations VALUES ('org_legacy', 'Legacy', 'team', CURRENT_TIMESTAMP);
		INSERT INTO api_tokens VALUES ('tok_legacy', 'org_legacy', ?, 'Legacy token', CURRENT_TIMESTAMP);
	`, hashToken("legacy-secret"))
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(db)
	if err := s.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	principal, err := s.AuthenticateToken("legacy-secret")
	if err != nil {
		t.Fatal(err)
	}
	if principal.Role != RoleMember {
		t.Fatalf("legacy role = %q, want %q", principal.Role, RoleMember)
	}
}

func TestStoreRunsAndBaselines(t *testing.T) {
	s := newTestStore(t)

	orgID := "org_test"
	_ = s.CreateOrganization(orgID, "Test Org", "free")

	repo, err := s.GetOrCreateRepo(orgID, "bregaldahq/payments", "main")
	if err != nil {
		t.Fatalf("failed to get/create repo: %v", err)
	}

	sc, err := s.GetOrCreateScenario(repo.ID, "wallet_transfer", "postgres")
	if err != nil {
		t.Fatalf("failed to get/create scenario: %v", err)
	}

	run1 := &RunRecord{
		ID:          "run_01",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "f4b18c0",
		Branch:      "main",
		PRNumber:    0,
		Status:      "passed",
		AnomalyType: "NONE",
		Seed:        12345,
		DurationMS:  320,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.SaveRun(run1, nil); err != nil {
		t.Fatalf("failed to save run1: %v", err)
	}

	if err := s.SetBaseline(repo.ID, sc.ID, "main", run1.ID); err != nil {
		t.Fatalf("failed to set baseline: %v", err)
	}

	base, err := s.GetLatestBaseline(repo.ID, sc.ID, "main")
	if err != nil {
		t.Fatalf("failed to get baseline: %v", err)
	}
	if base.ID != run1.ID {
		t.Errorf("expected baseline run %s, got %s", run1.ID, base.ID)
	}
}

func TestStoreScopesRepositoriesAndRunsByOrganization(t *testing.T) {
	s := newTestStore(t)
	for _, orgID := range []string{"org_a", "org_b"} {
		if err := s.CreateOrganization(orgID, orgID, "team"); err != nil {
			t.Fatal(err)
		}
	}
	repoA, err := s.GetOrCreateRepo("org_a", "shared/payments", "main")
	if err != nil {
		t.Fatal(err)
	}
	repoB, err := s.GetOrCreateRepo("org_b", "shared/payments", "main")
	if err != nil {
		t.Fatal(err)
	}
	if repoA.ID == repoB.ID || repoA.OrgID == repoB.OrgID {
		t.Fatalf("repositories crossed tenant boundary: A=%+v B=%+v", repoA, repoB)
	}
	scenarioA, err := s.GetOrCreateScenario(repoA.ID, "transfer", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	runA := &RunRecord{
		ID: "run_a", RepoID: repoA.ID, ScenarioID: scenarioA.ID, CommitSHA: "abc",
		Branch: "main", Status: "passed", Seed: 1, CreatedAt: time.Now().UTC(),
	}
	if err := s.SaveRun(runA, nil); err != nil {
		t.Fatal(err)
	}

	if _, _, err := s.GetRunForOrg("org_b", runA.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant GetRunForOrg error = %v, want ErrNotFound", err)
	}
	owned, _, err := s.GetRunForOrg("org_a", runA.ID)
	if err != nil || owned.ID != runA.ID {
		t.Fatalf("owner could not read run: run=%+v error=%v", owned, err)
	}
	runs, err := s.ListRecentRunsForOrg("org_b", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 0 {
		t.Fatalf("organization B listed A runs: %+v", runs)
	}
	if _, err := s.GetRepositoryByFullName("org_b", "shared/payments"); err != nil {
		t.Fatalf("organization B could not resolve its repository: %v", err)
	}
	if err := s.CreateWebhook(context.Background(), &WebhookRecord{ID: "wh_b", OrgID: "org_b", TargetType: "generic", URL: "https://example.com/hook", Events: "all", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteWebhook(context.Background(), "org_a", "wh_b"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-tenant webhook deletion error = %v, want ErrNotFound", err)
	}
	webhooksB, err := s.ListWebhooks(context.Background(), "org_b")
	if err != nil || len(webhooksB) != 1 {
		t.Fatalf("organization B webhook changed: webhooks=%+v error=%v", webhooksB, err)
	}
}

func TestAutoMigrateReplacesLegacyGlobalRepositoryUniqueness(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/legacy.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		PRAGMA foreign_keys = ON;
		CREATE TABLE organizations (id TEXT PRIMARY KEY, name TEXT NOT NULL, plan TEXT NOT NULL, created_at DATETIME NOT NULL);
		CREATE TABLE repositories (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, full_name TEXT NOT NULL UNIQUE, default_branch TEXT NOT NULL, created_at DATETIME NOT NULL, FOREIGN KEY (org_id) REFERENCES organizations(id));
		CREATE TABLE scenarios (id TEXT PRIMARY KEY, repo_id TEXT NOT NULL, name TEXT NOT NULL, driver TEXT NOT NULL, created_at DATETIME NOT NULL, UNIQUE(repo_id, name), FOREIGN KEY (repo_id) REFERENCES repositories(id));
		CREATE TABLE runs (id TEXT PRIMARY KEY, repo_id TEXT NOT NULL, scenario_id TEXT NOT NULL, commit_sha TEXT NOT NULL, branch TEXT NOT NULL, pr_number INTEGER NOT NULL, status TEXT NOT NULL, anomaly_type TEXT, seed INTEGER NOT NULL, duration_ms INTEGER NOT NULL, created_at DATETIME NOT NULL, FOREIGN KEY (repo_id) REFERENCES repositories(id), FOREIGN KEY (scenario_id) REFERENCES scenarios(id));
		INSERT INTO organizations VALUES ('org_a', 'A', 'team', CURRENT_TIMESTAMP), ('org_b', 'B', 'team', CURRENT_TIMESTAMP);
		INSERT INTO repositories VALUES ('repo_a', 'org_a', 'shared/api', 'main', CURRENT_TIMESTAMP);
		INSERT INTO scenarios VALUES ('scenario_a', 'repo_a', 'transfer', 'sqlite', CURRENT_TIMESTAMP);
		INSERT INTO runs VALUES ('run_a', 'repo_a', 'scenario_a', 'abc', 'main', 0, 'passed', '', 1, 1, CURRENT_TIMESTAMP);
	`)
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(db)
	if err := s.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	repoB, err := s.GetOrCreateRepo("org_b", "shared/api", "main")
	if err != nil {
		t.Fatalf("tenant-scoped repository uniqueness was not migrated: %v", err)
	}
	if repoB.OrgID != "org_b" {
		t.Fatalf("unexpected migrated repository: %+v", repoB)
	}
	if run, _, err := s.GetRunForOrg("org_a", "run_a"); err != nil || run.RepoID != "repo_a" {
		t.Fatalf("dependent run was not preserved: run=%+v error=%v", run, err)
	}
	var foreignKeys int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&foreignKeys); err != nil || foreignKeys != 1 {
		t.Fatalf("foreign key mode was not restored: enabled=%d error=%v", foreignKeys, err)
	}
	var invalidTable string
	if err := db.QueryRow(`PRAGMA foreign_key_check`).Scan(&invalidTable); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("repository migration left an invalid foreign key in table %q: %v", invalidTable, err)
	}
}

func TestAutoMigratePurgesLegacyHostedFindingDetails(t *testing.T) {
	db, err := sql.Open("sqlite", t.TempDir()+"/legacy-findings.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(`
		CREATE TABLE organizations (id TEXT PRIMARY KEY, name TEXT NOT NULL, plan TEXT NOT NULL, created_at DATETIME NOT NULL);
		CREATE TABLE repositories (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, full_name TEXT NOT NULL, default_branch TEXT NOT NULL, created_at DATETIME NOT NULL, UNIQUE(org_id, full_name));
		CREATE TABLE scenarios (id TEXT PRIMARY KEY, repo_id TEXT NOT NULL, name TEXT NOT NULL, driver TEXT NOT NULL, created_at DATETIME NOT NULL, UNIQUE(repo_id, name));
		CREATE TABLE runs (id TEXT PRIMARY KEY, repo_id TEXT NOT NULL, scenario_id TEXT NOT NULL, commit_sha TEXT NOT NULL, branch TEXT NOT NULL, pr_number INTEGER NOT NULL, status TEXT NOT NULL, anomaly_type TEXT, seed INTEGER NOT NULL, duration_ms INTEGER NOT NULL, created_at DATETIME NOT NULL);
		CREATE TABLE findings (id TEXT PRIMARY KEY, run_id TEXT NOT NULL, anomaly_type TEXT NOT NULL, assertion TEXT, minimal_ops INTEGER NOT NULL, repro_code TEXT, trace_json TEXT, created_at DATETIME NOT NULL);
		INSERT INTO organizations VALUES ('org_a', 'A', 'team', CURRENT_TIMESTAMP);
		INSERT INTO repositories VALUES ('repo_a', 'org_a', 'acme/private', 'main', CURRENT_TIMESTAMP);
		INSERT INTO scenarios VALUES ('scenario_a', 'repo_a', 'transfer', 'sqlite', CURRENT_TIMESTAMP);
		INSERT INTO runs VALUES ('run_a', 'repo_a', 'scenario_a', 'abc', 'main', 0, 'failed', 'P4', 1, 1, CURRENT_TIMESTAMP);
		INSERT INTO findings VALUES ('finding_a', 'run_a', 'P4', 'email = alice@example.com', 2, 'postgres://alice:secret@db/private', '[{"sql":"SELECT secret"}]', CURRENT_TIMESTAMP);
	`)
	if err != nil {
		t.Fatal(err)
	}

	store := NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	_, finding, err := store.GetRunForOrg("org_a", "run_a")
	if err != nil {
		t.Fatal(err)
	}
	if finding == nil || finding.Assertion != "" || finding.ReproCode != "" || finding.TraceJSON != "" {
		t.Fatalf("legacy finding details were not purged: %+v", finding)
	}
	var migrations int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = '2026-09-16-purge-hosted-finding-details'`).Scan(&migrations); err != nil || migrations != 1 {
		t.Fatalf("privacy migration marker count=%d error=%v", migrations, err)
	}
}

func TestStoreListRecentRuns(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_list_test"
	_ = s.CreateOrganization(orgID, "List Org", "free")
	repo, _ := s.GetOrCreateRepo(orgID, "bregaldahq/auth", "main")
	sc, _ := s.GetOrCreateScenario(repo.ID, "session_race", "sqlite")

	for i := 0; i < 5; i++ {
		_ = s.SaveRun(&RunRecord{
			ID:          "run_test_" + string(rune('a'+i)),
			RepoID:      repo.ID,
			ScenarioID:  sc.ID,
			CommitSHA:   "sha_" + string(rune('a'+i)),
			Branch:      "main",
			Status:      "passed",
			AnomalyType: "NONE",
			Seed:        uint64(100 + i),
			DurationMS:  int64(10 + i),
			CreatedAt:   time.Now().UTC().Add(time.Duration(i) * time.Minute),
		}, nil)
	}

	recent, err := s.ListRecentRuns(10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recent) != 5 {
		t.Errorf("expected 5 recent runs, got %d", len(recent))
	}
	// Verify descending order
	if recent[0].ID != "run_test_e" {
		t.Errorf("expected newest run_test_e first, got %s", recent[0].ID)
	}
}

func TestStore_IdempotentRunsAndOutboxMigration(t *testing.T) {
	s := newTestStore(t)

	// Verify columns on runs table
	rows, err := s.db.Query(`PRAGMA table_info(runs)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	cols := make(map[string]bool)
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}

	if !cols["idempotency_key"] {
		t.Errorf("expected runs to have idempotency_key column")
	}
	if !cols["scenario_fingerprint"] {
		t.Errorf("expected runs to have scenario_fingerprint column")
	}
	if !cols["commit_timestamp"] {
		t.Errorf("expected runs to have commit_timestamp column")
	}

	// Verify webhook_outbox table
	var outboxCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'webhook_outbox'`).Scan(&outboxCount)
	if err != nil || outboxCount != 1 {
		t.Fatalf("expected webhook_outbox table to exist, count=%d, err=%v", outboxCount, err)
	}

	// Verify migration recorded
	var applied int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = '2026-09-17-idempotent-runs-and-outbox'`).Scan(&applied)
	if err != nil || applied != 1 {
		t.Fatalf("expected migration '2026-09-17-idempotent-runs-and-outbox' recorded, applied=%d, err=%v", applied, err)
	}
}
func TestStore_SaveRunTx_Atomicity(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_atomicity"
	_ = s.CreateOrganization(orgID, "Atomicity Org", "free")
	repo, _ := s.GetOrCreateRepo(orgID, "bregaldahq/atomic", "main")
	sc, _ := s.GetOrCreateScenario(repo.ID, "atomic_sc", "sqlite")

	// 1. Successful atomic save of run, finding, and outbox item
	runID := "run_atomic_success"
	now := time.Now().UTC()
	run := &RunRecord{
		ID:                  runID,
		RepoID:              repo.ID,
		ScenarioID:          sc.ID,
		CommitSHA:           "sha_atomic",
		Branch:              "main",
		Status:              "failed",
		AnomalyType:         "G1a",
		Seed:                42,
		DurationMS:          120,
		CreatedAt:           now,
		IdempotencyKey:      "key_atomic_1",
		ScenarioFingerprint: "fp_atomic_1",
		CommitTimestamp:     &now,
	}
	finding := &FindingRecord{
		ID:          "find_atomic_success",
		RunID:       runID,
		AnomalyType: "G1a",
		Assertion:   "no dirty read",
		MinimalOps:  2,
		CreatedAt:   now,
	}
	outbox := []*OutboxItem{
		{
			ID:          "outbox_1",
			OrgID:       orgID,
			WebhookID:   "wh_fake", // Note: won't enforce FK if webhook isn't created, or create webhook first
			EventType:   "failure",
			PayloadJSON: `{"status":"failed"}`,
			Status:      "pending",
			NextRetryAt: now,
			CreatedAt:   now,
		},
	}
	// Create a webhook for FK
	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID:         "wh_fake",
		OrgID:      orgID,
		TargetType: "slack",
		URL:        "https://example.com/slack",
		Events:     "failure,regression",
		Active:     true,
		CreatedAt:  now,
	})

	savedRun, created, err := s.SaveRunTx(context.Background(), run, finding, outbox)
	if err != nil {
		t.Fatalf("expected successful SaveRunTx, got: %v", err)
	}
	if !created {
		t.Fatalf("expected created=true for new run")
	}
	if savedRun.ID != runID {
		t.Errorf("expected savedRun.ID=%s, got %s", runID, savedRun.ID)
	}

	// Verify all persisted
	r, f, err := s.GetRun(runID)
	if err != nil || r == nil || f == nil {
		t.Fatalf("expected run and finding persisted, err=%v", err)
	}
	if r.IdempotencyKey != "key_atomic_1" || r.ScenarioFingerprint != "fp_atomic_1" {
		t.Errorf("unexpected run fields: key=%s fp=%s", r.IdempotencyKey, r.ScenarioFingerprint)
	}
	var outboxCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM webhook_outbox WHERE id = 'outbox_1'`).Scan(&outboxCount); err != nil || outboxCount != 1 {
		t.Fatalf("expected outbox item persisted, count=%d, err=%v", outboxCount, err)
	}

	// 2. Failed atomic save: simulate failure (e.g. invalid duplicate finding ID)
	failedRunID := "run_atomic_fail"
	failedRun := &RunRecord{
		ID:             failedRunID,
		RepoID:         repo.ID,
		ScenarioID:     sc.ID,
		CommitSHA:      "sha_fail",
		Branch:         "main",
		Status:         "failed",
		CreatedAt:      now,
		IdempotencyKey: "key_atomic_2",
	}
	// duplicate finding ID "find_atomic_success" should violate PRIMARY KEY
	duplicateFinding := &FindingRecord{
		ID:          "find_atomic_success",
		RunID:       failedRunID,
		AnomalyType: "G1a",
		CreatedAt:   now,
	}
	_, _, err = s.SaveRunTx(context.Background(), failedRun, duplicateFinding, nil)
	if err == nil {
		t.Fatalf("expected error on duplicate finding primary key")
	}

	// Verify rollback: failedRun MUST NOT exist in runs table
	var exists int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE id = ?`, failedRunID).Scan(&exists)
	if exists != 0 {
		t.Fatalf("expected failedRun to be rolled back, but found in runs table")
	}
}

func TestStore_SaveRunTx_Idempotency(t *testing.T) {
	s := newTestStore(t)
	orgID := "org_idem"
	_ = s.CreateOrganization(orgID, "Idem Org", "free")
	repo, _ := s.GetOrCreateRepo(orgID, "bregaldahq/idem", "main")
	sc, _ := s.GetOrCreateScenario(repo.ID, "idem_sc", "sqlite")

	now := time.Now().UTC()
	_ = s.CreateWebhook(context.Background(), &WebhookRecord{
		ID:         "wh_idem",
		OrgID:      orgID,
		TargetType: "discord",
		URL:        "https://example.com/discord",
		Events:     "failure",
		Active:     true,
		CreatedAt:  now,
	})

	run1 := &RunRecord{
		ID:                  "run_idem_first",
		RepoID:              repo.ID,
		ScenarioID:          sc.ID,
		CommitSHA:           "sha_idem",
		Branch:              "main",
		Status:              "failed",
		AnomalyType:         "G2",
		Seed:                999,
		DurationMS:          300,
		CreatedAt:           now,
		IdempotencyKey:      "ci_workflow_run_42",
		ScenarioFingerprint: "fp_idem_1",
		CommitTimestamp:     &now,
	}
	finding1 := &FindingRecord{
		ID:          "find_idem_1",
		RunID:       run1.ID,
		AnomalyType: "G2",
		MinimalOps:  3,
		CreatedAt:   now,
	}
	outbox1 := []*OutboxItem{
		{
			ID:          "outbox_idem_1",
			OrgID:       orgID,
			WebhookID:   "wh_idem",
			EventType:   "failure",
			PayloadJSON: `{"event":"failure"}`,
			Status:      "pending",
			NextRetryAt: now,
			CreatedAt:   now,
		},
	}

	saved1, created1, err := s.SaveRunTx(context.Background(), run1, finding1, outbox1)
	if err != nil || !created1 {
		t.Fatalf("first save failed: created=%v err=%v", created1, err)
	}
	if saved1.ID != run1.ID {
		t.Errorf("expected saved1.ID=%s, got %s", run1.ID, saved1.ID)
	}

	// Second attempt with same idempotency key and repo, but different run ID (simulating retry)
	run2 := &RunRecord{
		ID:                  "run_idem_retry",
		RepoID:              repo.ID,
		ScenarioID:          sc.ID,
		CommitSHA:           "sha_idem",
		Branch:              "main",
		Status:              "failed",
		AnomalyType:         "G2",
		Seed:                999,
		DurationMS:          300,
		CreatedAt:           now.Add(10 * time.Second),
		IdempotencyKey:      "ci_workflow_run_42",
		ScenarioFingerprint: "fp_idem_1",
	}
	finding2 := &FindingRecord{
		ID:          "find_idem_2",
		RunID:       run2.ID,
		AnomalyType: "G2",
		CreatedAt:   now.Add(10 * time.Second),
	}
	outbox2 := []*OutboxItem{
		{
			ID:          "outbox_idem_2",
			OrgID:       orgID,
			WebhookID:   "wh_idem",
			EventType:   "failure",
			PayloadJSON: `{"event":"failure"}`,
			Status:      "pending",
			NextRetryAt: now,
			CreatedAt:   now,
		},
	}

	saved2, created2, err := s.SaveRunTx(context.Background(), run2, finding2, outbox2)
	if err != nil {
		t.Fatalf("second save returned unexpected error: %v", err)
	}
	if created2 {
		t.Fatalf("expected created=false for duplicate idempotency key")
	}
	if saved2.ID != run1.ID {
		t.Errorf("expected duplicate save to return original run ID %s, got %s", run1.ID, saved2.ID)
	}

	// Verify database only has 1 run, 1 finding, and 1 outbox item
	var totalRuns int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM runs WHERE repo_id = ?`, repo.ID).Scan(&totalRuns)
	if totalRuns != 1 {
		t.Errorf("expected 1 run in db, got %d", totalRuns)
	}
	var totalFindings int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM findings WHERE run_id = ?`, run1.ID).Scan(&totalFindings)
	if totalFindings != 1 {
		t.Errorf("expected 1 finding in db, got %d", totalFindings)
	}
	var totalOutbox int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM webhook_outbox WHERE org_id = ?`, orgID).Scan(&totalOutbox)
	if totalOutbox != 1 {
		t.Errorf("expected 1 outbox item in db, got %d", totalOutbox)
	}
}