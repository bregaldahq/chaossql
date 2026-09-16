package cloud

import (
	"errors"
	"fmt"
	"time"
)

const MaxPayloadBytes = 64 * 1024

var ErrForbiddenPayloadDetail = errors.New("cloud payload contains details that must remain local")

type metadataPayload struct {
	Version      string                       `json:"version"`
	Timestamp    time.Time                    `json:"timestamp"`
	CI           *metadataCI                  `json:"ci,omitempty"`
	Scenario     metadataScenario             `json:"scenario"`
	Result       metadataExecutionSummary     `json:"result"`
	Reproduction *metadataReproductionSummary `json:"reproduction,omitempty"`
}

type metadataCI struct {
	Provider          string `json:"provider"`
	Repository        string `json:"repository"`
	CommitSHA         string `json:"commit_sha"`
	Branch            string `json:"branch"`
	BaseBranch        string `json:"base_branch,omitempty"`
	PullRequestNumber int    `json:"pull_request_number,omitempty"`
	RunID             string `json:"run_id,omitempty"`
}

type metadataScenario struct {
	Name          string `json:"name"`
	Fingerprint   string `json:"fingerprint,omitempty"`
	Driver        string `json:"driver"`
	DriverVersion string `json:"driver_version,omitempty"`
	Workers       int    `json:"workers"`
	Iterations    int    `json:"iterations"`
	Seed          uint64 `json:"seed"`
}

type metadataExecutionSummary struct {
	Status            string                    `json:"status"`
	ExecutionStatus   string                    `json:"execution_status,omitempty"`
	Success           bool                      `json:"success"`
	ViolationDetected bool                      `json:"violation_detected"`
	AnomalyType       string                    `json:"anomaly_type"`
	DurationMS        int64                     `json:"duration_ms"`
	TotalSchedules    int                       `json:"total_schedules"`
	FailedSchedules   int                       `json:"failed_schedules"`
	FailingInvariant  *metadataInvariantSummary `json:"failing_invariant,omitempty"`
}

type metadataInvariantSummary struct {
	Name string `json:"name"`
}

type metadataReproductionSummary struct {
	MinimalOperationsCount int   `json:"minimal_operations_count"`
	ShrinkDurationMS       int64 `json:"shrink_duration_ms"`
}

func projectMetadataPayload(req *RunIngestRequest, now time.Time) metadataPayload {
	version := req.Version
	if version == "" {
		version = "1.0"
	}
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = now.UTC()
	}

	payload := metadataPayload{
		Version:   version,
		Timestamp: timestamp,
		Scenario: metadataScenario{
			Name:          req.Scenario.Name,
			Fingerprint:   req.Scenario.Fingerprint,
			Driver:        req.Scenario.Driver,
			DriverVersion: req.Scenario.DriverVersion,
			Workers:       req.Scenario.Workers,
			Iterations:    req.Scenario.Iterations,
			Seed:          req.Scenario.Seed,
		},
		Result: metadataExecutionSummary{
			Status:            req.Result.Status,
			ExecutionStatus:   req.Result.ExecutionStatus,
			Success:           req.Result.Success,
			ViolationDetected: req.Result.ViolationDetected,
			AnomalyType:       req.Result.AnomalyType,
			DurationMS:        req.Result.DurationMS,
			TotalSchedules:    req.Result.TotalSchedules,
			FailedSchedules:   req.Result.FailedSchedules,
		},
	}
	if req.CI != nil {
		payload.CI = &metadataCI{
			Provider:          req.CI.Provider,
			Repository:        req.CI.Repository,
			CommitSHA:         req.CI.CommitSHA,
			Branch:            req.CI.Branch,
			BaseBranch:        req.CI.BaseBranch,
			PullRequestNumber: req.CI.PullRequestNumber,
			RunID:             req.CI.RunID,
		}
	}
	if req.Result.FailingInvariant != nil {
		payload.Result.FailingInvariant = &metadataInvariantSummary{Name: req.Result.FailingInvariant.Name}
	}
	if req.Reproduction != nil {
		payload.Reproduction = &metadataReproductionSummary{
			MinimalOperationsCount: req.Reproduction.MinimalOperationsCount,
			ShrinkDurationMS:       req.Reproduction.ShrinkDurationMS,
		}
	}
	return payload
}

func ValidateMetadataOnlyRequest(req *RunIngestRequest) error {
	if req.CI != nil && req.CI.Actor != "" {
		return fmt.Errorf("%w: ci.actor", ErrForbiddenPayloadDetail)
	}
	if req.Schedule != nil {
		return fmt.Errorf("%w: schedule", ErrForbiddenPayloadDetail)
	}
	if invariant := req.Result.FailingInvariant; invariant != nil {
		if invariant.Query != "" || invariant.Assertion != "" || invariant.Actual != "" {
			return fmt.Errorf("%w: invariant query, assertion, and actual values", ErrForbiddenPayloadDetail)
		}
	}
	if reproduction := req.Reproduction; reproduction != nil {
		if reproduction.ReproGoCode != "" || reproduction.MermaidDiagram != "" || len(reproduction.SanitizedMinimalTrace) > 0 {
			return fmt.Errorf("%w: traces, diagrams, and reproduction code", ErrForbiddenPayloadDetail)
		}
	}
	return nil
}
