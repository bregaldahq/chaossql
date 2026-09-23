package server

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

// This fixture has the run/baseline schema from before corrected ingestion.
// In that release the handler stored the upload time as commit_timestamp.
// It deliberately does not create the new run_ingestions provenance table.
func legacyIngestionStore(t *testing.T, path string) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	s := NewStore(db)
	statements := []string{
		`CREATE TABLE organizations(id TEXT PRIMARY KEY,name TEXT NOT NULL,plan TEXT NOT NULL,created_at DATETIME NOT NULL)`,
		`CREATE TABLE repositories(id TEXT PRIMARY KEY,org_id TEXT NOT NULL,full_name TEXT NOT NULL,default_branch TEXT NOT NULL,created_at DATETIME NOT NULL,UNIQUE(org_id, full_name),FOREIGN KEY(org_id) REFERENCES organizations(id))`,
		`CREATE TABLE scenarios(id TEXT PRIMARY KEY,repo_id TEXT NOT NULL,name TEXT NOT NULL,driver TEXT NOT NULL,created_at DATETIME NOT NULL,UNIQUE(repo_id,name),FOREIGN KEY(repo_id) REFERENCES repositories(id))`,
		`CREATE TABLE runs(id TEXT PRIMARY KEY,repo_id TEXT NOT NULL,scenario_id TEXT NOT NULL,commit_sha TEXT NOT NULL,branch TEXT NOT NULL,pr_number INTEGER NOT NULL,status TEXT NOT NULL,anomaly_type TEXT,seed INTEGER NOT NULL,duration_ms INTEGER NOT NULL,created_at DATETIME NOT NULL,idempotency_key TEXT,scenario_fingerprint TEXT NOT NULL DEFAULT '',commit_timestamp DATETIME,FOREIGN KEY(repo_id) REFERENCES repositories(id),FOREIGN KEY(scenario_id) REFERENCES scenarios(id))`,
		`CREATE TABLE baselines(id TEXT PRIMARY KEY,repo_id TEXT NOT NULL,scenario_id TEXT NOT NULL,branch TEXT NOT NULL,run_id TEXT NOT NULL,updated_at DATETIME NOT NULL,UNIQUE(repo_id,scenario_id,branch),FOREIGN KEY(repo_id) REFERENCES repositories(id),FOREIGN KEY(scenario_id) REFERENCES scenarios(id),FOREIGN KEY(run_id) REFERENCES runs(id))`,
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	uploaded := time.Date(2026, 9, 17, 16, 0, 0, 0, time.UTC)
	for _, statement := range []string{
		`INSERT INTO organizations VALUES('org_cloud_test','Legacy Org','enterprise',?)`,
		`INSERT INTO repositories VALUES('legacy_repo','org_cloud_test','acme/payments','main',?)`,
		`INSERT INTO scenarios VALUES('legacy_scenario','legacy_repo','transfer','sqlite',?)`,
	} {
		if _, err := db.Exec(statement, uploaded); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO runs VALUES('legacy_run','legacy_repo','legacy_scenario','oldsha','main',0,'passed','',42,10,?,NULL,'',?)`, uploaded, uploaded); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO baselines VALUES('legacy_base','legacy_repo','legacy_scenario','main','legacy_run',?)`, uploaded); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestUpgradePreservesCorrectedIngestionProvenance(t *testing.T) {
	s := legacyIngestionStore(t, filepath.Join(t.TempDir(), "interim.sqlite"))
	defer s.db.Close()
	// An interim corrected deployment already wrote commit time and durable
	// responses, but did not yet have the timestamp-provenance migration.
	if _, err := s.db.Exec(`CREATE TABLE run_ingestions(run_id TEXT PRIMARY KEY REFERENCES runs(id),request_hash TEXT NOT NULL,response_json TEXT NOT NULL,driver TEXT NOT NULL,total_schedules INTEGER,failed_schedules INTEGER)`); err != nil {
		t.Fatal(err)
	}
	committed := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	if _, err := s.db.Exec(`INSERT INTO runs VALUES('corrected_run','legacy_repo','legacy_scenario','newsha','feature',2,'passed','',42,10,?,NULL,'fp1',?)`, committed, committed); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(`INSERT INTO run_ingestions VALUES('corrected_run','digest','{}','sqlite',NULL,NULL)`); err != nil {
		t.Fatal(err)
	}
	if err := s.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	corrected, _, err := s.GetRun("corrected_run")
	if err != nil {
		t.Fatal(err)
	}
	if corrected.CommitTimestamp == nil || !corrected.CommitTimestamp.Equal(committed) {
		t.Fatalf("erased corrected ingestion timestamp: %+v", corrected)
	}
	legacy, _, err := s.GetRun("legacy_run")
	if err != nil {
		t.Fatal(err)
	}
	if legacy.CommitTimestamp != nil {
		t.Fatal("legacy timestamp remained trusted")
	}
}

func TestUpgradeClearsLegacyUploadTimeAndPreservesNewCommitTime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upgrade.sqlite")
	s := legacyIngestionStore(t, path)
	defer func() { _ = s.db.Close() }()
	if err := s.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	legacy, err := s.GetLatestBaseline("legacy_repo", "legacy_scenario", "main")
	if err != nil {
		t.Fatal(err)
	}
	if legacy.CommitTimestamp != nil {
		t.Fatalf("upgrade trusts legacy upload time as commit time: %v", legacy.CommitTimestamp)
	}

	req := correctionPayload()
	actualCommit := time.Date(2026, 9, 17, 14, 0, 0, 0, time.UTC)
	req.CI.CommitTimestamp = &actualCommit
	trusted, created, err := s.ingest(context.Background(), "org_cloud_test", "trusted-upgrade", "https://cloud.example", &req, ingestionMeasurements{})
	if err != nil || !created {
		t.Fatalf("ingest after upgrade: created=%v err=%v", created, err)
	}
	baseline, err := s.GetLatestBaseline("legacy_repo", "legacy_scenario", "main")
	if err != nil {
		t.Fatal(err)
	}
	if baseline.ID != trusted.RunID || baseline.CommitTimestamp == nil || !baseline.CommitTimestamp.Equal(actualCommit) {
		t.Fatalf("trusted baseline blocked by legacy chronology: %+v", baseline)
	}
	if trusted.Baseline != nil {
		t.Fatal("legacy identity was treated as compatible evidence")
	}

	req.CI.Branch = "feature"
	req.CI.PullRequestNumber = 7
	req.Result.Status = "failed"
	regression, _, err := s.ingest(context.Background(), "org_cloud_test", "upgraded-pr", "https://cloud.example", &req, ingestionMeasurements{})
	if err != nil || !regression.IsRegression || regression.Baseline == nil || regression.Baseline.RunID != trusted.RunID {
		t.Fatalf("upgraded PR failed to compare: %+v err=%v", regression, err)
	}

	// Rows written through other store callers after migration must also survive
	// subsequent starts: this correction is a one-time provenance migration.
	manual := &RunRecord{ID: "after_migration", RepoID: "legacy_repo", ScenarioID: "legacy_scenario", Branch: "feature", Status: "passed", CreatedAt: time.Now().UTC(), CommitTimestamp: &actualCommit}
	if err := s.SaveRun(manual, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	s = NewStore(db)
	if err := s.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{trusted.RunID, "after_migration"} {
		run, _, err := s.GetRun(id)
		if err != nil {
			t.Fatal(err)
		}
		if run.CommitTimestamp == nil || !run.CommitTimestamp.Equal(actualCommit) {
			t.Fatalf("restart erased corrected timestamp for %s: %+v", id, run)
		}
	}
}
