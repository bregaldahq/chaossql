package server

import (
	"database/sql"
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
