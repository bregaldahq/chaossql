package server

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newMigratedStore(t *testing.T) *Store {
	t.Helper()
	store := newTestStore(t)
	if err := store.CreateOrganization("org_a", "Org A", "pro"); err != nil {
		t.Fatal(err)
	}
	return store
}

// Storage failures must propagate to callers instead of being reported as
// empty results, so the API can answer 5xx rather than silently hiding data.
func TestStoreMethodsPropagateStorageFailures(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name  string
		table string
		call  func(*Store) error
	}{
		{"authenticate", "api_tokens", func(s *Store) error { _, err := s.AuthenticateToken("secret"); return err }},
		{"validate token", "api_tokens", func(s *Store) error { _, err := s.ValidateToken("secret"); return err }},
		{"get or create repo", "repositories", func(s *Store) error { _, err := s.GetOrCreateRepo("org_a", "acme/api", "main"); return err }},
		{"get repo by name", "repositories", func(s *Store) error { _, err := s.GetRepositoryByFullName("org_a", "acme/api"); return err }},
		{"get or create scenario", "scenarios", func(s *Store) error { _, err := s.GetOrCreateScenario("repo_1", "banking", "sqlite"); return err }},
		{"save run", "runs", func(s *Store) error {
			_, _, err := s.SaveRunTx(ctx, &RunRecord{ID: "run_1", RepoID: "repo_1", ScenarioID: "sc_1", Status: "passed"}, nil, nil)
			return err
		}},
		{"enqueue outbox", "webhook_outbox", func(s *Store) error {
			return s.EnqueueOutbox(ctx, []*OutboxItem{{ID: "ob_1", OrgID: "org_a", WebhookID: "wh_1", EventType: "regression", PayloadJSON: "{}"}})
		}},
		{"pending outbox", "webhook_outbox", func(s *Store) error { _, err := s.GetPendingOutboxItems(ctx, 10); return err }},
		{"mark delivered", "webhook_outbox", func(s *Store) error { return s.MarkOutboxDelivered(ctx, "ob_1") }},
		{"mark failed", "webhook_outbox", func(s *Store) error { return s.MarkOutboxAttemptFailed(ctx, "ob_1", "boom", 3) }},
		{"get run", "runs", func(s *Store) error { _, _, err := s.GetRun("run_1"); return err }},
		{"get run for org", "runs", func(s *Store) error { _, _, err := s.GetRunForOrg("org_a", "run_1"); return err }},
		{"latest baseline", "baselines", func(s *Store) error { _, err := s.GetLatestBaseline("repo_1", "sc_1", "main"); return err }},
		{"set baseline", "baselines", func(s *Store) error { return s.SetBaseline("repo_1", "sc_1", "main", "run_1") }},
		{"list runs", "runs", func(s *Store) error { _, err := s.ListRuns("repo_1", 10, 0); return err }},
		{"list recent runs", "runs", func(s *Store) error { _, err := s.ListRecentRuns(10, 0); return err }},
		{"list org runs", "runs", func(s *Store) error { _, err := s.ListRecentRunsForOrg("org_a", 10, 0); return err }},
		{"create webhook", "webhooks", func(s *Store) error {
			return s.CreateWebhook(ctx, &WebhookRecord{ID: "wh_1", OrgID: "org_a", URL: "https://hooks.slack.com/services/T/B/C"})
		}},
		{"list webhooks", "webhooks", func(s *Store) error { _, err := s.ListWebhooks(ctx, "org_a"); return err }},
		{"delete webhook", "webhooks", func(s *Store) error { return s.DeleteWebhook(ctx, "org_a", "wh_1") }},
		{"get webhook", "webhooks", func(s *Store) error { _, err := s.GetWebhookByID(ctx, "org_a", "wh_1"); return err }},
		{"active webhooks", "webhooks", func(s *Store) error { _, err := s.GetActiveWebhooksForEvent(ctx, "org_a", "regression"); return err }},
		{"subscription", "organizations", func(s *Store) error { _, err := s.GetOrgSubscription("org_a"); return err }},
		{"list organizations", "organizations", func(s *Store) error { _, err := s.ListOrganizations(); return err }},
		{"set plan", "organizations", func(s *Store) error { return s.SetOrganizationPlan("org_a", "enterprise") }},
		{"issue token", "api_tokens", func(s *Store) error { _, err := s.IssueToken("org_a", "ci", RoleMember); return err }},
		{"provision organization", "organizations", func(s *Store) error { _, _, err := s.ProvisionOrganization("New Org", "pro"); return err }},
		{"purge expired data", "runs", func(s *Store) error { _, err := s.PurgeExpiredData(ctx, time.Now()); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := newMigratedStore(t)
			execWithoutForeignKeys(t, store, `DROP TABLE `+tc.table)
			err := tc.call(store)
			if err == nil {
				t.Fatalf("expected an error after dropping %s", tc.table)
			}
			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUnauthorized) {
				t.Fatalf("storage failure was misreported as %v", err)
			}
		})
	}
}

func TestAutoMigrateIsIdempotentAndReportsClosedDatabase(t *testing.T) {
	store := newMigratedStore(t)
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("second AutoMigrate: %v", err)
	}
	store.db.Close()
	if err := store.AutoMigrate(); err == nil {
		t.Fatal("AutoMigrate on a closed database must fail")
	}
}

func TestStoreRejectsInvalidRoles(t *testing.T) {
	store := newMigratedStore(t)
	if err := store.UpsertAPITokenWithRole("tok", "org_a", "secret", "x", Role("superuser")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("upsert err = %v, want ErrInvalidRole", err)
	}
	if _, err := store.IssueToken("org_a", "ci", Role("superuser")); err == nil {
		t.Fatal("IssueToken must reject an invalid role")
	}
	if err := store.EnsureOrganization("org_a", "Renamed", "free"); err != nil {
		t.Fatalf("EnsureOrganization on an existing org must be a no-op, got %v", err)
	}
	if sub, err := store.GetOrgSubscription("org_a"); err != nil || sub == nil {
		t.Fatalf("subscription = %+v, err = %v", sub, err)
	}
}
