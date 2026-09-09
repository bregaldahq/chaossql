package server

import (
	"database/sql"
	"errors"
	"fmt"
)

// PlanConfig defines the feature gates and usage allowances per tier
type PlanConfig struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	PriceMonthlyUSD     int    `json:"price_monthly_usd"`
	PriceAnnualUSD      int    `json:"price_annual_usd"`
	MaxRepositories     int    `json:"max_repositories"` // -1 for unlimited
	RetentionDays       int    `json:"retention_days"`   // -1 for unlimited
	HasRegressionGating bool   `json:"has_regression_gating"`
	HasPRComments       bool   `json:"has_pr_comments"`
	HasPrioritySupport  bool   `json:"has_priority_support"`
}

var (
	PlanDeveloper = PlanConfig{
		ID:                  "developer",
		Name:                "Cloud Developer",
		PriceMonthlyUSD:     0,
		PriceAnnualUSD:      0,
		MaxRepositories:     1,
		RetentionDays:       7,
		HasRegressionGating: false,
		HasPRComments:       true,
		HasPrioritySupport:  false,
	}

	PlanTeam = PlanConfig{
		ID:                  "team",
		Name:                "Cloud Team",
		PriceMonthlyUSD:     39,
		PriceAnnualUSD:      31,
		MaxRepositories:     10,
		RetentionDays:       90,
		HasRegressionGating: true,
		HasPRComments:       true,
		HasPrioritySupport:  false,
	}

	PlanPro = PlanConfig{
		ID:                  "pro",
		Name:                "Cloud Pro",
		PriceMonthlyUSD:     99,
		PriceAnnualUSD:      79,
		MaxRepositories:     30,
		RetentionDays:       365,
		HasRegressionGating: true,
		HasPRComments:       true,
		HasPrioritySupport:  true,
	}

	PlanEnterprise = PlanConfig{
		ID:                  "enterprise",
		Name:                "Enterprise Custom",
		PriceMonthlyUSD:     499,
		PriceAnnualUSD:      399,
		MaxRepositories:     -1,
		RetentionDays:       -1,
		HasRegressionGating: true,
		HasPRComments:       true,
		HasPrioritySupport:  true,
	}
)

func GetPlan(id string) PlanConfig {
	switch id {
	case "team":
		return PlanTeam
	case "pro":
		return PlanPro
	case "enterprise":
		return PlanEnterprise
	default:
		return PlanDeveloper
	}
}

type OrgSubscription struct {
	OrgID                   string     `json:"org_id"`
	OrgName                 string     `json:"org_name"`
	Plan                    PlanConfig `json:"plan"`
	ActiveRepositoriesCount int        `json:"active_repositories_count"`
	RemainingRepositories   int        `json:"remaining_repositories"`
	CanAddRepository        bool       `json:"can_add_repository"`
}

var ErrPlanLimitReached = errors.New("repository plan limit reached")

func (s *Store) GetOrgSubscription(orgID string) (*OrgSubscription, error) {
	var orgName, planID string
	err := s.db.QueryRow(`SELECT name, plan FROM organizations WHERE id = ?`, orgID).Scan(&orgName, &planID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}

	var count int
	err = s.db.QueryRow(`SELECT count(*) FROM repositories WHERE org_id = ?`, orgID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count repositories: %w", err)
	}

	plan := GetPlan(planID)
	remaining := -1
	canAdd := true

	if plan.MaxRepositories >= 0 {
		remaining = plan.MaxRepositories - count
		if remaining < 0 {
			remaining = 0
		}
		canAdd = count < plan.MaxRepositories
	}

	return &OrgSubscription{
		OrgID:                   orgID,
		OrgName:                 orgName,
		Plan:                    plan,
		ActiveRepositoriesCount: count,
		RemainingRepositories:   remaining,
		CanAddRepository:        canAdd,
	}, nil
}
