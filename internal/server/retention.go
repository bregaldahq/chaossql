package server

import (
	"context"
	"fmt"
	"time"
)

// RetentionReport counts the records removed by one retention pass.
type RetentionReport struct {
	Runs     int
	Findings int
	Alerts   int
}

func (r *RetentionReport) add(other RetentionReport) {
	r.Runs += other.Runs
	r.Findings += other.Findings
	r.Alerts += other.Alerts
}

// PurgeExpiredData removes run history older than each organization's plan
// retention window. Runs referenced by a current baseline are kept so
// regression comparison never loses its reference. Pending alerts are kept
// until the dispatcher finishes them. Timestamps are compared in Go because
// stored DATETIME text is not reliably ordered as a string.
func (s *Store) PurgeExpiredData(ctx context.Context, now time.Time) (RetentionReport, error) {
	var total RetentionReport
	rows, err := s.queryer().QueryContext(ctx, `SELECT id, plan FROM organizations ORDER BY id`)
	if err != nil {
		return total, err
	}
	type orgPlan struct{ id, plan string }
	var orgs []orgPlan
	for rows.Next() {
		var org orgPlan
		if err := rows.Scan(&org.id, &org.plan); err != nil {
			rows.Close()
			return total, err
		}
		orgs = append(orgs, org)
	}
	if err := rows.Close(); err != nil {
		return total, err
	}
	for _, org := range orgs {
		days := GetPlan(org.plan).RetentionDays
		if days < 0 {
			continue
		}
		cutoff := now.UTC().AddDate(0, 0, -days)
		var report RetentionReport
		err := s.withTransaction(ctx, func(bound *Store) error {
			var err error
			report, err = bound.purgeOrganization(ctx, org.id, cutoff)
			return err
		})
		if err != nil {
			return total, fmt.Errorf("purge organization %s: %w", org.id, err)
		}
		total.add(report)
	}
	return total, nil
}

func (s *Store) purgeOrganization(ctx context.Context, orgID string, cutoff time.Time) (RetentionReport, error) {
	var report RetentionReport
	runRows, err := s.tx.QueryContext(ctx, `SELECT r.id, r.created_at FROM runs r
		JOIN repositories p ON p.id = r.repo_id
		WHERE p.org_id = ? AND NOT EXISTS (SELECT 1 FROM baselines b WHERE b.run_id = r.id)`, orgID)
	if err != nil {
		return report, err
	}
	var expiredRuns []string
	for runRows.Next() {
		var id string
		var createdAt time.Time
		if err := runRows.Scan(&id, &createdAt); err != nil {
			runRows.Close()
			return report, err
		}
		if createdAt.Before(cutoff) {
			expiredRuns = append(expiredRuns, id)
		}
	}
	if err := runRows.Close(); err != nil {
		return report, err
	}
	for _, id := range expiredRuns {
		result, err := s.tx.ExecContext(ctx, `DELETE FROM findings WHERE run_id = ?`, id)
		if err != nil {
			return report, err
		}
		findings, _ := result.RowsAffected()
		report.Findings += int(findings)
		if _, err := s.tx.ExecContext(ctx, `DELETE FROM run_ingestions WHERE run_id = ?`, id); err != nil {
			return report, err
		}
		if _, err := s.tx.ExecContext(ctx, `DELETE FROM runs WHERE id = ?`, id); err != nil {
			return report, err
		}
		report.Runs++
	}

	alertRows, err := s.tx.QueryContext(ctx, `SELECT id, created_at FROM webhook_outbox WHERE org_id = ? AND status <> 'pending'`, orgID)
	if err != nil {
		return report, err
	}
	var expiredAlerts []string
	for alertRows.Next() {
		var id string
		var createdAt time.Time
		if err := alertRows.Scan(&id, &createdAt); err != nil {
			alertRows.Close()
			return report, err
		}
		if createdAt.Before(cutoff) {
			expiredAlerts = append(expiredAlerts, id)
		}
	}
	if err := alertRows.Close(); err != nil {
		return report, err
	}
	for _, id := range expiredAlerts {
		if _, err := s.tx.ExecContext(ctx, `DELETE FROM webhook_outbox WHERE id = ?`, id); err != nil {
			return report, err
		}
		report.Alerts++
	}
	return report, nil
}

// RunRetentionLoop runs a pass immediately and then every interval until ctx is
// cancelled. The clock is injected so retention stays testable and deterministic.
func RunRetentionLoop(ctx context.Context, store *Store, interval time.Duration, now func() time.Time, observe func(RetentionReport, error)) {
	pass := func() {
		report, err := store.PurgeExpiredData(ctx, now())
		if observe != nil && ctx.Err() == nil {
			observe(report, err)
		}
	}
	pass()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pass()
		}
	}
}

// RetentionInterval is how often a server with retention enforcement purges.
const RetentionInterval = time.Hour

// StartRetention launches the hourly purge used by both server entrypoints.
func StartRetention(ctx context.Context, store *Store, logf func(string, ...any)) {
	go RunRetentionLoop(ctx, store, RetentionInterval, time.Now, func(report RetentionReport, err error) {
		if err != nil {
			logf("[ChaosSQL Cloud] Retention pass failed: %v", err)
			return
		}
		if report != (RetentionReport{}) {
			logf("[ChaosSQL Cloud] Retention removed %d runs, %d findings, %d finished alerts", report.Runs, report.Findings, report.Alerts)
		}
	})
}
