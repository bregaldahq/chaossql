package server

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newProvisioningStore(t *testing.T) *Store {
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

func TestProvisionOrganizationIssuesOwnerTokenAtomically(t *testing.T) {
	store := newProvisioningStore(t)
	org, token, err := store.ProvisionOrganization("Acme Payments", "team")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(org.ID, "org_") || org.Name != "Acme Payments" || org.Plan != "team" {
		t.Fatalf("unexpected organization: %+v", org)
	}
	if !strings.HasPrefix(token, "csql_") {
		t.Fatalf("unexpected token format %q", token)
	}
	principal, err := store.AuthenticateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if principal.OrgID != org.ID || principal.Role != RoleOwner {
		t.Fatalf("owner token bound to %+v", principal)
	}
	var stored int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM api_tokens WHERE token_hash = ?`, token).Scan(&stored); err != nil || stored != 0 {
		t.Fatalf("raw token persisted: count=%d err=%v", stored, err)
	}

	second, secondToken, err := store.ProvisionOrganization("Acme Payments", "team")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == org.ID || secondToken == token {
		t.Fatal("provisioning reused organization identity or credential")
	}
}

func TestProvisionOrganizationRejectsInvalidInput(t *testing.T) {
	store := newProvisioningStore(t)
	for _, test := range []struct{ name, plan string }{
		{"", "team"},
		{"   ", "team"},
		{"Acme", ""},
		{"Acme", "gold"},
		{strings.Repeat("a", 121), "team"},
	} {
		if _, _, err := store.ProvisionOrganization(test.name, test.plan); err == nil {
			t.Fatalf("accepted name=%q plan=%q", test.name, test.plan)
		}
	}
	orgs, err := store.ListOrganizations()
	if err != nil {
		t.Fatal(err)
	}
	if len(orgs) != 0 {
		t.Fatalf("rejected input left organizations behind: %+v", orgs)
	}
}

func TestProvisionOrganizationRollsBackWhenTokenFails(t *testing.T) {
	store := newProvisioningStore(t)
	if _, err := store.db.Exec(`CREATE TRIGGER reject_tokens BEFORE INSERT ON api_tokens BEGIN SELECT RAISE(ABORT, 'injected'); END`); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.ProvisionOrganization("Acme", "team"); err == nil {
		t.Fatal("expected injected token failure")
	}
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM organizations`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("organization persisted without owner token: count=%d err=%v", count, err)
	}
}

func TestIssueTokenRequiresExistingOrganization(t *testing.T) {
	store := newProvisioningStore(t)
	if _, err := store.IssueToken("org_typo", "CI", RoleMember); !errors.Is(err, ErrNotFound) {
		t.Fatalf("issued token for unknown organization: %v", err)
	}
	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM organizations`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("token issuance created an organization: count=%d err=%v", count, err)
	}

	org, _, err := store.ProvisionOrganization("Acme", "team")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueToken(org.ID, "CI", Role("root")); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("accepted invalid role: %v", err)
	}
	token, err := store.IssueToken(org.ID, "CI", RoleMember)
	if err != nil {
		t.Fatal(err)
	}
	principal, err := store.AuthenticateToken(token)
	if err != nil || principal.OrgID != org.ID || principal.Role != RoleMember {
		t.Fatalf("issued token principal=%+v err=%v", principal, err)
	}
}

func TestSetOrganizationPlan(t *testing.T) {
	store := newProvisioningStore(t)
	org, _, err := store.ProvisionOrganization("Acme", "developer")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetOrganizationPlan(org.ID, "pro"); err != nil {
		t.Fatal(err)
	}
	subscription, err := store.GetOrgSubscription(org.ID)
	if err != nil || subscription.Plan.ID != "pro" {
		t.Fatalf("plan not updated: %+v err=%v", subscription, err)
	}
	if err := store.SetOrganizationPlan(org.ID, "gold"); !errors.Is(err, ErrInvalidPlan) {
		t.Fatalf("accepted unknown plan: %v", err)
	}
	if err := store.SetOrganizationPlan("org_missing", "pro"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("updated missing organization: %v", err)
	}
}

func TestListOrganizationsReportsUsage(t *testing.T) {
	store := newProvisioningStore(t)
	first, _, err := store.ProvisionOrganization("First", "team")
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := store.ProvisionOrganization("Second", "developer")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.IssueToken(first.ID, "CI", RoleMember); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetOrCreateRepo(first.ID, "acme/one", "main"); err != nil {
		t.Fatal(err)
	}
	orgs, err := store.ListOrganizations()
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]OrganizationSummary{}
	for _, org := range orgs {
		byID[org.ID] = org
	}
	if got := byID[first.ID]; got.Repositories != 1 || got.Tokens != 2 || got.Plan != "team" {
		t.Fatalf("first organization summary: %+v", got)
	}
	if got := byID[second.ID]; got.Repositories != 0 || got.Tokens != 1 || got.Plan != "developer" {
		t.Fatalf("second organization summary: %+v", got)
	}
}

// Two provisioned tenants on one instance must never observe each other's data.
func TestProvisionedOrganizationsAreIsolated(t *testing.T) {
	store := newProvisioningStore(t)
	handler := NewRouter(RouterConfig{Store: store, Engine: NewRegressionEngine(store), PublicBaseURL: "https://cloud.example"})
	orgA, ownerA, err := store.ProvisionOrganization("Tenant A", "team")
	if err != nil {
		t.Fatal(err)
	}
	orgB, ownerB, err := store.ProvisionOrganization("Tenant B", "team")
	if err != nil {
		t.Fatal(err)
	}
	for _, tenant := range []struct{ org, repo string }{{orgA.ID, "acme/a"}, {orgB.ID, "acme/b"}} {
		repo, err := store.GetOrCreateRepo(tenant.org, tenant.repo, "main")
		if err != nil {
			t.Fatal(err)
		}
		scenario, err := store.GetOrCreateScenario(repo.ID, "transfer", "sqlite")
		if err != nil {
			t.Fatal(err)
		}
		run := &RunRecord{ID: "run_" + tenant.org, RepoID: repo.ID, ScenarioID: scenario.ID, CommitSHA: "sha", Branch: "main", Status: "passed", CreatedAt: time.Now().UTC()}
		if err := store.SaveRun(run, nil); err != nil {
			t.Fatal(err)
		}
	}

	get := func(path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	list := get("/v1/runs", ownerA)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "run_"+orgA.ID) || strings.Contains(list.Body.String(), orgB.ID) {
		t.Fatalf("tenant A run list leaked or missed data: %d %s", list.Code, list.Body.String())
	}
	for _, path := range []string{"/v1/runs/run_" + orgB.ID, "/v1/repositories/acme/b/runs", "/v1/organizations/" + orgB.ID + "/subscription"} {
		if rec := get(path, ownerA); rec.Code != http.StatusNotFound {
			t.Fatalf("tenant A reached %s: %d %s", path, rec.Code, rec.Body.String())
		}
	}
	if rec := get("/v1/organizations/me/subscription", ownerB); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), orgB.ID) {
		t.Fatalf("tenant B subscription: %d %s", rec.Code, rec.Body.String())
	}
}
