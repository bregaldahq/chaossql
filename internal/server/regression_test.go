package server

import (
	"testing"
	"time"
)

func setupTestRepoAndScenario(t *testing.T, s *Store) (*Repository, *Scenario) {
	orgID := "org_test"
	_ = s.CreateOrganization(orgID, "Test Org", "team")

	repo, err := s.GetOrCreateRepo(orgID, "bregaldahq/core-banking", "main")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	sc, err := s.GetOrCreateScenario(repo.ID, "account_balance_check", "postgres")
	if err != nil {
		t.Fatalf("failed to create scenario: %v", err)
	}

	return repo, sc
}

func TestRegressionEngine_NoBaseline(t *testing.T) {
	s := newTestStore(t)
	repo, sc := setupTestRepoAndScenario(t, s)
	engine := NewRegressionEngine(s)

	runPR := &RunRecord{
		ID:          "run_pr_01",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "abc111",
		Branch:      "feat/fast-transfer",
		PRNumber:    42,
		Status:      "failed",
		AnomalyType: "P4",
		Seed:        999,
		DurationMS:  150,
		CreatedAt:   time.Now().UTC(),
	}

	comp, isReg, err := engine.Evaluate(repo, sc, runPR)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isReg {
		t.Errorf("expected isReg false without baseline, got true")
	}
	if comp != nil {
		t.Errorf("expected comp nil without baseline, got %v", comp)
	}
}

func TestRegressionEngine_BaselineEstablishedAndPRRegresses(t *testing.T) {
	s := newTestStore(t)
	repo, sc := setupTestRepoAndScenario(t, s)
	engine := NewRegressionEngine(s)

	// 1. Run on main passes -> establishes baseline
	mainRun := &RunRecord{
		ID:          "run_main_01",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "main001",
		Branch:      "main",
		PRNumber:    0,
		Status:      "passed",
		AnomalyType: "NONE",
		Seed:        100,
		DurationMS:  200,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.SaveRun(mainRun, nil); err != nil {
		t.Fatalf("failed to save main run: %v", err)
	}

	comp, isReg, err := engine.Evaluate(repo, sc, mainRun)
	if err != nil {
		t.Fatalf("unexpected error on main: %v", err)
	}
	if isReg {
		t.Errorf("expected isReg false on initial baseline establishment")
	}

	// Check baseline stored
	base, err := s.GetLatestBaseline(repo.ID, sc.ID, "main")
	if err != nil || base == nil {
		t.Fatalf("expected baseline stored, err: %v", err)
	}
	if base.ID != "run_main_01" {
		t.Errorf("expected baseline ID run_main_01, got %s", base.ID)
	}

	// 2. PR run fails -> must detect REGRESSION!
	prRun := &RunRecord{
		ID:          "run_pr_99",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "pr999",
		Branch:      "feat/withdrawal-cache",
		PRNumber:    12,
		Status:      "failed",
		AnomalyType: "P4",
		Seed:        200,
		DurationMS:  180,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.SaveRun(prRun, nil); err != nil {
		t.Fatalf("failed to save pr run: %v", err)
	}

	comp, isReg, err = engine.Evaluate(repo, sc, prRun)
	if err != nil {
		t.Fatalf("unexpected error on pr: %v", err)
	}
	if !isReg {
		t.Errorf("expected isReg TRUE for failing PR when baseline passed")
	}
	if comp == nil || comp.RunID != "run_main_01" {
		t.Errorf("expected comparison against run_main_01, got %v", comp)
	}

	// 3. Another PR run passes -> NO regression
	prPassing := &RunRecord{
		ID:          "run_pr_100",
		RepoID:      repo.ID,
		ScenarioID:  sc.ID,
		CommitSHA:   "pr1000",
		Branch:      "fix/locking",
		PRNumber:    13,
		Status:      "passed",
		AnomalyType: "NONE",
		Seed:        300,
		DurationMS:  210,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.SaveRun(prPassing, nil); err != nil {
		t.Fatalf("failed to save pr passing run: %v", err)
	}

	comp, isReg, err = engine.Evaluate(repo, sc, prPassing)
	if err != nil {
		t.Fatalf("unexpected error on pr passing: %v", err)
	}
	if isReg {
		t.Errorf("expected isReg FALSE for passing PR")
	}
	if comp == nil || comp.RunID != "run_main_01" {
		t.Errorf("expected comparison against run_main_01, got %v", comp)
	}
}
