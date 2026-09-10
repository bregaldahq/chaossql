package server

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPlanConfigResolution(t *testing.T) {
	dev := GetPlan("developer")
	if dev.MaxRepositories != 1 || dev.PriceMonthlyUSD != 0 {
		t.Errorf("unexpected developer plan: %+v", dev)
	}

	team := GetPlan("team")
	if team.MaxRepositories != 10 || team.PriceMonthlyUSD != 39 || !team.HasRegressionGating {
		t.Errorf("unexpected team plan: %+v", team)
	}

	pro := GetPlan("pro")
	if pro.MaxRepositories != 30 || pro.PriceMonthlyUSD != 99 {
		t.Errorf("unexpected pro plan: %+v", pro)
	}
}

func TestOrgSubscriptionAndLimits(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	store := NewStore(db)
	if err := store.AutoMigrate(); err != nil {
		t.Fatalf("failed to automigrate: %v", err)
	}

	orgID := "org_limits_test"
	if err := store.CreateOrganization(orgID, "Limits Corp", "team"); err != nil {
		t.Fatalf("failed to create org: %v", err)
	}

	// 1. Initial subscription state (0 repos)
	sub, err := store.GetOrgSubscription(orgID)
	if err != nil {
		t.Fatalf("failed to get org subscription: %v", err)
	}

	if sub.Plan.ID != "team" {
		t.Errorf("expected plan team, got %s", sub.Plan.ID)
	}
	if sub.ActiveRepositoriesCount != 0 {
		t.Errorf("expected 0 repos, got %d", sub.ActiveRepositoriesCount)
	}
	if sub.RemainingRepositories != 10 {
		t.Errorf("expected 10 remaining repos, got %d", sub.RemainingRepositories)
	}
	if !sub.CanAddRepository {
		t.Errorf("expected canAddRepository true")
	}

	// 2. Add repos up to limit
	for i := 1; i <= 10; i++ {
		_, err := store.GetOrCreateRepo(orgID, "bregaldahq/repo-"+string(rune('0'+i)), "main")
		if err != nil {
			t.Fatalf("failed to create repo: %v", err)
		}
	}

	sub2, err := store.GetOrgSubscription(orgID)
	if err != nil {
		t.Fatalf("failed to get sub after adding repos: %v", err)
	}
	if sub2.ActiveRepositoriesCount != 10 {
		t.Errorf("expected 10 repos, got %d", sub2.ActiveRepositoriesCount)
	}
	if sub2.RemainingRepositories != 0 {
		t.Errorf("expected 0 remaining repos, got %d", sub2.RemainingRepositories)
	}
	if sub2.CanAddRepository {
		t.Errorf("expected CanAddRepository false when 10/10 reached")
	}
}
