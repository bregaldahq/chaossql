package cloud

import "time"

type metadataPayload struct {
	Version      string                       `json:"version"`
	Timestamp    time.Time                    `json:"timestamp"`
	CI           *metadataCI                  `json:"ci,omitempty"`
	Scenario     ScenarioMetadata             `json:"scenario"`
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
		Scenario:  req.Scenario,
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
