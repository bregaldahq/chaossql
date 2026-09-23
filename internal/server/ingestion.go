package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bregaldahq/chaossql/internal/cloud"
)

// queryer binds all store operations to the same connection during ingestion.
type sqlQueryer interface {
	Exec(string, ...any) (sql.Result, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	Query(string, ...any) (*sql.Rows, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func (s *Store) queryer() sqlQueryer {
	if s.tx != nil {
		return s.tx
	}
	return s.db
}

func (s *Store) beginOperation(ctx context.Context) (*sql.Tx, bool, error) {
	if s.tx != nil {
		return s.tx, false, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	return tx, true, err
}

func (s *Store) withTransaction(ctx context.Context, operation func(*Store) error) error {
	if s.tx != nil {
		return operation(s)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := operation(&Store{db: s.db, tx: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ensureIngestionMetadata() error {
	return s.withTransaction(context.Background(), func(bound *Store) error {
		if _, err := bound.tx.Exec(`CREATE TABLE IF NOT EXISTS run_ingestions (
 run_id TEXT PRIMARY KEY REFERENCES runs(id), request_hash TEXT NOT NULL,
 response_json TEXT NOT NULL, driver TEXT NOT NULL,
 total_schedules INTEGER, failed_schedules INTEGER)`); err != nil {
			return err
		}
		const migration = "2026-09-21-correct-commit-timestamp-provenance"
		var applied int
		if err := bound.tx.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version=?`, migration).Scan(&applied); err != nil {
			return err
		}
		if applied != 0 {
			return nil
		}
		// The old handler stored upload time in commit_timestamp. There is no
		// reliable way to recover commit time from those rows, including rows
		// referenced by baselines. Corrected ingestion has durable provenance
		// in run_ingestions, so preserve it if upgrading an interim deployment.
		if _, err := bound.tx.Exec(`UPDATE runs SET commit_timestamp=NULL
 WHERE NOT EXISTS (SELECT 1 FROM run_ingestions i WHERE i.run_id=runs.id)`); err != nil {
			return fmt.Errorf("clear legacy upload timestamps: %w", err)
		}
		_, err := bound.tx.Exec(`INSERT INTO schema_migrations(version,applied_at) VALUES(?,?)`, migration, time.Now().UTC())
		return err
	})
}

var ErrIdempotencyConflict = errors.New("idempotency key already belongs to different execution content")

func randomIdentifier(prefix string) (string, error) {
	var b [24]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(b[:]), nil
}

// ingest commits the response decision together with the evidence and alerts.
// The public response is immutable even if newer baselines arrive before a retry.
func (s *Store) ingest(ctx context.Context, orgID, key, baseURL string, req *cloud.RunIngestRequest, measurements ingestionMeasurements) (cloud.RunIngestResponse, bool, error) {
	var response cloud.RunIngestResponse
	created := false
	canonical := *req
	canonical.IdempotencyKey = ""
	encoded, err := json.Marshal(struct {
		Request      cloud.RunIngestRequest
		Measurements ingestionMeasurements
	}{canonical, measurements})
	if err != nil {
		return response, false, err
	}
	digest := sha256.Sum256(encoded)
	requestHash := hex.EncodeToString(digest[:])
	err = s.withTransaction(ctx, func(bound *Store) error {
		repoName, branch, defaultBranch, commitSHA := "local/chaossql-project", "main", "main", "unknown"
		prNumber := 0
		var commitTime *time.Time
		if req.CI != nil {
			if req.CI.Repository != "" {
				repoName = req.CI.Repository
			}
			if req.CI.Branch != "" {
				branch = req.CI.Branch
			}
			if req.CI.BaseBranch != "" {
				defaultBranch = req.CI.BaseBranch
			}
			commitSHA = req.CI.CommitSHA
			prNumber = req.CI.PullRequestNumber
			commitTime = req.CI.CommitTimestamp
		}
		repo, err := bound.GetOrCreateRepo(orgID, repoName, defaultBranch)
		if err != nil {
			return err
		}
		if key != "" {
			var storedHash, storedResponse string
			err := bound.tx.QueryRowContext(ctx, `SELECT i.request_hash,i.response_json FROM run_ingestions i JOIN runs r ON r.id=i.run_id WHERE r.repo_id=? AND r.idempotency_key=?`, repo.ID, key).Scan(&storedHash, &storedResponse)
			if err == nil {
				if storedHash != requestHash {
					return ErrIdempotencyConflict
				}
				return json.Unmarshal([]byte(storedResponse), &response)
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
			// A key from before durable response metadata cannot be replayed faithfully.
			var legacyID string
			err = bound.tx.QueryRowContext(ctx, `SELECT id FROM runs WHERE repo_id=? AND idempotency_key=?`, repo.ID, key).Scan(&legacyID)
			if err == nil {
				return ErrIdempotencyConflict
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}
		name, driver := req.Scenario.Name, req.Scenario.Driver
		if name == "" {
			name = "default"
		}
		if driver == "" {
			driver = "sqlite"
		}
		sc, err := bound.GetOrCreateScenario(repo.ID, name, driver)
		if err != nil {
			return err
		}
		runID, err := randomIdentifier("run_")
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		run := &RunRecord{ID: runID, RepoID: repo.ID, ScenarioID: sc.ID, CommitSHA: commitSHA, Branch: branch, PRNumber: prNumber, Status: req.Result.Status, AnomalyType: req.Result.AnomalyType, Seed: req.Scenario.Seed, DurationMS: req.Result.DurationMS, CreatedAt: now, IdempotencyKey: key, ScenarioFingerprint: req.Scenario.Fingerprint, CommitTimestamp: commitTime}
		var finding *FindingRecord
		if req.Result.ViolationDetected || req.Reproduction != nil {
			findingID, err := randomIdentifier("find_")
			if err != nil {
				return err
			}
			finding = &FindingRecord{ID: findingID, RunID: runID, AnomalyType: req.Result.AnomalyType, CreatedAt: now}
			if req.Result.FailingInvariant != nil {
				finding.Assertion = req.Result.FailingInvariant.Name
			}
			if req.Reproduction != nil {
				finding.MinimalOps = req.Reproduction.MinimalOperationsCount
			}
		}
		if _, _, err := bound.SaveRunTx(ctx, run, finding, nil); err != nil {
			return err
		}
		comparison, regression, err := NewRegressionEngine(bound).Evaluate(repo, sc, run)
		if err != nil {
			return err
		}
		message := "Execution passed successfully."
		if regression {
			message = fmt.Sprintf("Regression detected! Anomaly %s broke baseline on %s", run.AnomalyType, comparison.Branch)
		} else if run.Status != "passed" {
			message = fmt.Sprintf("Execution finished with status %s.", run.Status)
		}
		response = cloud.RunIngestResponse{Success: true, RunID: run.ID, URL: fmt.Sprintf("%s/dashboard?run=%s", baseURL, run.ID), IsRegression: regression, Baseline: comparison, Message: message}
		encodedResponse, err := json.Marshal(response)
		if err != nil {
			return err
		}
		if _, err := bound.tx.ExecContext(ctx, `INSERT INTO run_ingestions(run_id,request_hash,response_json,driver,total_schedules,failed_schedules) VALUES(?,?,?,?,?,?)`, run.ID, requestHash, string(encodedResponse), req.Scenario.Driver, measurements.Result.Total, measurements.Result.Failed); err != nil {
			return err
		}
		if regression || run.Status != "passed" {
			event := "failure"
			if regression {
				event = "regression"
			}
			hooks, err := bound.GetActiveWebhooksForEvent(ctx, orgID, event)
			if err != nil {
				return err
			}
			baselineStatus := ""
			if comparison != nil {
				baselineStatus = comparison.Status
			}
			alert := RegressionAlert{RepoFullName: repo.FullName, Branch: run.Branch, PRNumber: run.PRNumber, CommitSHA: run.CommitSHA, AnomalyType: run.AnomalyType, AnomalyName: run.AnomalyType, Driver: driver, Scenario: name, Seed: run.Seed, DurationMS: run.DurationMS, BaselineStatus: baselineStatus, RunURL: response.URL, IsRegression: regression}
			payload, err := json.Marshal(alert)
			if err != nil {
				return err
			}
			items := make([]*OutboxItem, 0, len(hooks))
			for _, hook := range hooks {
				id, err := randomIdentifier("outbox_")
				if err != nil {
					return err
				}
				items = append(items, &OutboxItem{ID: id, OrgID: orgID, WebhookID: hook.ID, EventType: event, PayloadJSON: string(payload), Status: "pending", NextRetryAt: now, CreatedAt: now})
			}
			if err := bound.EnqueueOutbox(ctx, items); err != nil {
				return err
			}
		}
		created = true
		return nil
	})
	return response, created, err
}

func ingestionKey(header, body string) (string, error) {
	header = strings.TrimSpace(header)
	body = strings.TrimSpace(body)
	if header != "" && body != "" && header != body {
		return "", errors.New("header and payload idempotency keys disagree")
	}
	key := body
	if header != "" {
		key = header
	}
	if key != "" && !cloud.IsSafeMetadataIdentifierWithLimit(key, 128) {
		return "", errors.New("invalid idempotency key")
	}
	return key, nil
}
