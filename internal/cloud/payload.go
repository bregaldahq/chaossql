package cloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

const MaxPayloadBytes = 64 * 1024

var ErrUnsafeMetadata = errors.New("cloud metadata contains an unsafe identifier")

var metadataIdentifierPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/+:-]*$`)

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

func validateMetadataPayload(payload *metadataPayload) error {
	fields := []struct {
		name  string
		value string
		limit int
	}{
		{"version", payload.Version, 16},
		{"scenario.name", payload.Scenario.Name, 128},
		{"scenario.fingerprint", payload.Scenario.Fingerprint, 128},
		{"scenario.driver", payload.Scenario.Driver, 32},
		{"scenario.driver_version", payload.Scenario.DriverVersion, 64},
		{"result.status", payload.Result.Status, 32},
		{"result.execution_status", payload.Result.ExecutionStatus, 32},
		{"result.anomaly_type", payload.Result.AnomalyType, 64},
	}
	if payload.CI != nil {
		fields = append(fields,
			struct {
				name  string
				value string
				limit int
			}{"ci.provider", payload.CI.Provider, 64},
			struct {
				name  string
				value string
				limit int
			}{"ci.repository", payload.CI.Repository, 255},
			struct {
				name  string
				value string
				limit int
			}{"ci.commit_sha", payload.CI.CommitSHA, 128},
			struct {
				name  string
				value string
				limit int
			}{"ci.branch", payload.CI.Branch, 255},
			struct {
				name  string
				value string
				limit int
			}{"ci.base_branch", payload.CI.BaseBranch, 255},
			struct {
				name  string
				value string
				limit int
			}{"ci.run_id", payload.CI.RunID, 128},
		)
	}
	if payload.Result.FailingInvariant != nil {
		fields = append(fields, struct {
			name  string
			value string
			limit int
		}{"result.failing_invariant.name", payload.Result.FailingInvariant.Name, 128})
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		if len(field.value) > field.limit || strings.Contains(field.value, "://") || !metadataIdentifierPattern.MatchString(field.value) {
			return fmt.Errorf("%w: %s", ErrUnsafeMetadata, field.name)
		}
	}
	return nil
}

func DecodeMetadataOnlyRequest(data []byte) (*RunIngestRequest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var payload metadataPayload
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("payload must contain one JSON object")
	}
	if err := validateMetadataPayload(&payload); err != nil {
		return nil, err
	}

	req := &RunIngestRequest{
		Version:   payload.Version,
		Timestamp: payload.Timestamp,
		Scenario: ScenarioMetadata{
			Name:          payload.Scenario.Name,
			Fingerprint:   payload.Scenario.Fingerprint,
			Driver:        payload.Scenario.Driver,
			DriverVersion: payload.Scenario.DriverVersion,
			Workers:       payload.Scenario.Workers,
			Iterations:    payload.Scenario.Iterations,
			Seed:          payload.Scenario.Seed,
		},
		Result: ExecutionSummary{
			Status:            payload.Result.Status,
			ExecutionStatus:   payload.Result.ExecutionStatus,
			Success:           payload.Result.Success,
			ViolationDetected: payload.Result.ViolationDetected,
			AnomalyType:       payload.Result.AnomalyType,
			DurationMS:        payload.Result.DurationMS,
			TotalSchedules:    payload.Result.TotalSchedules,
			FailedSchedules:   payload.Result.FailedSchedules,
		},
	}
	if payload.CI != nil {
		req.CI = &CIContext{
			Provider:          payload.CI.Provider,
			Repository:        payload.CI.Repository,
			CommitSHA:         payload.CI.CommitSHA,
			Branch:            payload.CI.Branch,
			BaseBranch:        payload.CI.BaseBranch,
			PullRequestNumber: payload.CI.PullRequestNumber,
			RunID:             payload.CI.RunID,
		}
	}
	if payload.Result.FailingInvariant != nil {
		req.Result.FailingInvariant = &InvariantSummary{Name: payload.Result.FailingInvariant.Name}
	}
	if payload.Reproduction != nil {
		req.Reproduction = &ReproductionData{
			MinimalOperationsCount: payload.Reproduction.MinimalOperationsCount,
			ShrinkDurationMS:       payload.Reproduction.ShrinkDurationMS,
		}
	}
	return req, nil
}

func IsSafeMetadataIdentifier(value string) bool {
	return value != "" && len(value) <= 128 && !strings.Contains(value, "://") && metadataIdentifierPattern.MatchString(value)
}
