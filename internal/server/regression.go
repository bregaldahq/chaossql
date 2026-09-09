package server

import (
	"errors"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

// RegressionEngine determines if a run introduced concurrency anomalies compared to the default branch baseline
type RegressionEngine struct {
	store *Store
}

// NewRegressionEngine creates a new engine instance
func NewRegressionEngine(store *Store) *RegressionEngine {
	return &RegressionEngine{store: store}
}

// Evaluate checks the run against baseline and updates baseline on default branch if passed
func (re *RegressionEngine) Evaluate(repo *Repository, sc *Scenario, run *RunRecord) (*cloud.BaselineComparison, bool, error) {
	if repo == nil || sc == nil || run == nil {
		return nil, false, errors.New("nil arguments passed to Evaluate")
	}

	defaultBranch := repo.DefaultBranch
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	isDefaultBranch := (run.Branch == defaultBranch) && (run.PRNumber == 0)

	// Fetch existing baseline for default branch
	baseRun, err := re.store.GetLatestBaseline(repo.ID, sc.ID, defaultBranch)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, false, fmt.Errorf("failed to query baseline: %w", err)
	}

	hasBaseline := (err == nil && baseRun != nil)

	// If this is on the default branch and passed, it becomes the new baseline!
	if isDefaultBranch && run.Status == "passed" {
		if err := re.store.SetBaseline(repo.ID, sc.ID, defaultBranch, run.ID); err != nil {
			return nil, false, fmt.Errorf("failed to update baseline: %w", err)
		}
		var comp *cloud.BaselineComparison
		if hasBaseline {
			comp = &cloud.BaselineComparison{
				RunID:     baseRun.ID,
				Status:    baseRun.Status,
				CommitSHA: baseRun.CommitSHA,
				Branch:    baseRun.Branch,
			}
		}
		return comp, false, nil
	}

	// If there is no baseline yet, we cannot declare a regression
	if !hasBaseline {
		return nil, false, nil
	}

	comp := &cloud.BaselineComparison{
		RunID:     baseRun.ID,
		Status:    baseRun.Status,
		CommitSHA: baseRun.CommitSHA,
		Branch:    baseRun.Branch,
	}

	// If baseline was passed and current run is failed -> REGRESSION!
	if baseRun.Status == "passed" && run.Status == "failed" {
		return comp, true, nil
	}

	return comp, false, nil
}
