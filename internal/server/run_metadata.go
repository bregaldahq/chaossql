package server

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

// Read measurements separately from the legacy run schema: old records have
// no saved regression decision or schedule counts, so retain their nulls.
func (s *Store) populatePublicRuns(ctx context.Context, runs ...*publicRunRecord) error {
	for _, run := range runs {
		var repo, scenario string
		var driver, response sql.NullString
		var total, failed sql.NullInt64
		err := s.queryer().QueryRowContext(ctx, `SELECT repo.full_name,sc.name,COALESCE(i.driver,sc.driver),i.response_json,i.total_schedules,i.failed_schedules
   FROM runs r JOIN repositories repo ON repo.id=r.repo_id JOIN scenarios sc ON sc.id=r.scenario_id
   LEFT JOIN run_ingestions i ON i.run_id=r.id WHERE r.id=?`, run.ID).Scan(&repo, &scenario, &driver, &response, &total, &failed)
		if err != nil {
			return err
		}
		run.RepoFullName = safePublicIdentifier(repo, 255)
		run.ScenarioName = safePublicIdentifier(scenario, 128)
		run.Driver = safePublicIdentifier(driver.String, 128)
		if response.Valid {
			var saved cloud.RunIngestResponse
			if err := json.Unmarshal([]byte(response.String), &saved); err != nil {
				return err
			}
			run.IsRegression = &saved.IsRegression
		}
		if total.Valid {
			value := int(total.Int64)
			run.TotalSchedules = &value
		}
		if failed.Valid {
			value := int(failed.Int64)
			run.FailedSchedules = &value
		}
	}
	return nil
}

// Pointers preserve the distinction between an observed zero and an omitted
// measurement in legacy callers. Full validation precedes this projection.
type ingestionMeasurements struct {
	Result struct {
		Total  *int `json:"total_schedules"`
		Failed *int `json:"failed_schedules"`
	} `json:"result"`
}
