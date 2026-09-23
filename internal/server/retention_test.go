package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

var retentionNow = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func newRetentionStore(t *testing.T) *Store {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	store := NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatal(err)
	}
	return store
}

type retentionTenant struct {
	repo     *Repository
	scenario *Scenario
}

func seedRetentionTenant(t *testing.T, store *Store, orgID, plan string) retentionTenant {
	t.Helper()
	if err := store.CreateOrganization(orgID, orgID, plan); err != nil {
		t.Fatal(err)
	}
	repo, err := store.GetOrCreateRepo(orgID, "acme/"+orgID, "main")
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := store.GetOrCreateScenario(repo.ID, "transfer", "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	return retentionTenant{repo: repo, scenario: scenario}
}

func saveAgedRun(t *testing.T, store *Store, tenant retentionTenant, id string, age time.Duration) {
	t.Helper()
	run := &RunRecord{ID: id, RepoID: tenant.repo.ID, ScenarioID: tenant.scenario.ID, CommitSHA: "sha", Branch: "main", Status: "failed", CreatedAt: retentionNow.Add(-age)}
	finding := &FindingRecord{ID: "finding_" + id, RunID: id, AnomalyType: "LOST_UPDATE", MinimalOps: 2, CreatedAt: run.CreatedAt}
	if err := store.SaveRun(run, finding); err != nil {
		t.Fatal(err)
	}
	if _, err := store.db.Exec(`INSERT INTO run_ingestions(run_id,request_hash,response_json,driver) VALUES(?,?,?,?)`, id, "hash", "{}", "sqlite"); err != nil {
		t.Fatal(err)
	}
}

func rowCount(t *testing.T, store *Store, query string, args ...any) int {
	t.Helper()
	var count int
	if err := store.db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func runExists(t *testing.T, store *Store, id string) bool {
	return rowCount(t, store, `SELECT COUNT(*) FROM runs WHERE id = ?`, id) == 1
}

func TestPurgeExpiredDataAppliesEachPlanWindow(t *testing.T) {
	store := newRetentionStore(t)
	developer := seedRetentionTenant(t, store, "org_dev", "developer") // 7 days
	team := seedRetentionTenant(t, store, "org_team", "team")          // 90 days
	enterprise := seedRetentionTenant(t, store, "org_ent", "enterprise")

	day := 24 * time.Hour
	saveAgedRun(t, store, developer, "dev_old", 8*day)
	saveAgedRun(t, store, developer, "dev_recent", 6*day)
	saveAgedRun(t, store, team, "team_mid", 30*day)
	saveAgedRun(t, store, team, "team_old", 91*day)
	saveAgedRun(t, store, enterprise, "ent_ancient", 900*day)

	report, err := store.PurgeExpiredData(context.Background(), retentionNow)
	if err != nil {
		t.Fatal(err)
	}
	if report.Runs != 2 || report.Findings != 2 {
		t.Fatalf("unexpected purge report: %+v", report)
	}
	for id, want := range map[string]bool{"dev_old": false, "dev_recent": true, "team_mid": true, "team_old": false, "ent_ancient": true} {
		if got := runExists(t, store, id); got != want {
			t.Fatalf("run %s exists=%v, want %v", id, got, want)
		}
	}
	for _, id := range []string{"dev_old", "team_old"} {
		if n := rowCount(t, store, `SELECT COUNT(*) FROM findings WHERE run_id = ?`, id); n != 0 {
			t.Fatalf("finding for %s survived", id)
		}
		if n := rowCount(t, store, `SELECT COUNT(*) FROM run_ingestions WHERE run_id = ?`, id); n != 0 {
			t.Fatalf("ingestion record for %s survived", id)
		}
	}
	// Repositories and scenarios are configuration, not history.
	if n := rowCount(t, store, `SELECT COUNT(*) FROM repositories`); n != 3 {
		t.Fatalf("retention removed repositories: %d", n)
	}
}

func TestPurgeExpiredDataKeepsCurrentBaseline(t *testing.T) {
	store := newRetentionStore(t)
	tenant := seedRetentionTenant(t, store, "org_dev", "developer")
	saveAgedRun(t, store, tenant, "old_baseline", 30*24*time.Hour)
	saveAgedRun(t, store, tenant, "old_superseded", 40*24*time.Hour)
	if err := store.SetBaseline(tenant.repo.ID, tenant.scenario.ID, "main", "old_baseline"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PurgeExpiredData(context.Background(), retentionNow); err != nil {
		t.Fatal(err)
	}
	if !runExists(t, store, "old_baseline") {
		t.Fatal("purge removed the run the current baseline depends on")
	}
	if n := rowCount(t, store, `SELECT COUNT(*) FROM findings WHERE run_id = 'old_baseline'`); n != 1 {
		t.Fatal("purge removed the baseline run's finding")
	}
	if runExists(t, store, "old_superseded") {
		t.Fatal("superseded expired run survived")
	}
	baseline, err := store.GetLatestBaseline(tenant.repo.ID, tenant.scenario.ID, "main")
	if err != nil || baseline.ID != "old_baseline" {
		t.Fatalf("baseline changed: %+v err=%v", baseline, err)
	}
}

func TestPurgeExpiredDataRemovesFinishedAlertsOnly(t *testing.T) {
	store := newRetentionStore(t)
	seedRetentionTenant(t, store, "org_dev", "developer")
	if err := store.CreateWebhook(context.Background(), &WebhookRecord{ID: "wh", OrgID: "org_dev", TargetType: "generic", URL: "https://hooks.example.com/x", Events: "all", Active: true}); err != nil {
		t.Fatal(err)
	}
	old := retentionNow.Add(-30 * 24 * time.Hour)
	items := []*OutboxItem{
		{ID: "delivered_old", OrgID: "org_dev", WebhookID: "wh", EventType: "regression", PayloadJSON: "{}", Status: "delivered", CreatedAt: old},
		{ID: "failed_old", OrgID: "org_dev", WebhookID: "wh", EventType: "regression", PayloadJSON: "{}", Status: "failed", CreatedAt: old},
		{ID: "pending_old", OrgID: "org_dev", WebhookID: "wh", EventType: "regression", PayloadJSON: "{}", Status: "pending", CreatedAt: old},
		{ID: "delivered_recent", OrgID: "org_dev", WebhookID: "wh", EventType: "regression", PayloadJSON: "{}", Status: "delivered", CreatedAt: retentionNow.Add(-time.Hour)},
	}
	if err := store.EnqueueOutbox(context.Background(), items); err != nil {
		t.Fatal(err)
	}
	report, err := store.PurgeExpiredData(context.Background(), retentionNow)
	if err != nil {
		t.Fatal(err)
	}
	if report.Alerts != 2 {
		t.Fatalf("unexpected alert purge count: %+v", report)
	}
	for id, want := range map[string]int{"delivered_old": 0, "failed_old": 0, "pending_old": 1, "delivered_recent": 1} {
		if got := rowCount(t, store, `SELECT COUNT(*) FROM webhook_outbox WHERE id = ?`, id); got != want {
			t.Fatalf("outbox %s count=%d, want %d", id, got, want)
		}
	}
}

func TestPurgeExpiredDataIsIdempotent(t *testing.T) {
	store := newRetentionStore(t)
	tenant := seedRetentionTenant(t, store, "org_dev", "developer")
	saveAgedRun(t, store, tenant, "dev_old", 8*24*time.Hour)
	if _, err := store.PurgeExpiredData(context.Background(), retentionNow); err != nil {
		t.Fatal(err)
	}
	report, err := store.PurgeExpiredData(context.Background(), retentionNow)
	if err != nil {
		t.Fatal(err)
	}
	if report != (RetentionReport{}) {
		t.Fatalf("second purge removed data: %+v", report)
	}
}

func TestRetentionLoopRunsImmediatelyAndStops(t *testing.T) {
	store := newRetentionStore(t)
	tenant := seedRetentionTenant(t, store, "org_dev", "developer")
	saveAgedRun(t, store, tenant, "dev_old", 8*24*time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	reports := make(chan RetentionReport, 1)
	go func() {
		defer close(done)
		RunRetentionLoop(ctx, store, time.Hour, func() time.Time { return retentionNow }, func(report RetentionReport, err error) {
			if err != nil {
				t.Errorf("retention pass failed: %v", err)
			}
			select {
			case reports <- report:
			default:
			}
		})
	}()
	select {
	case report := <-reports:
		if report.Runs != 1 {
			t.Fatalf("first pass report: %+v", report)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("retention loop did not run at startup")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("retention loop did not stop on cancellation")
	}
}
