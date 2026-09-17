package cloud

import (
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// CIContext contains metadata extracted from the CI/CD environment
type CIContext struct {
	Provider          string `json:"provider"`
	Repository        string `json:"repository"`
	CommitSHA         string `json:"commit_sha"`
	Branch            string `json:"branch"`
	BaseBranch        string `json:"base_branch,omitempty"`
	PullRequestNumber int    `json:"pull_request_number,omitempty"`
	RunID             string `json:"run_id,omitempty"`
	Actor             string `json:"actor,omitempty"`
}

// ScenarioMetadata captures parameters of the chaos.yaml specification
type ScenarioMetadata struct {
	Name          string `json:"name"`
	Fingerprint   string `json:"fingerprint,omitempty"`
	Driver        string `json:"driver"`
	DriverVersion string `json:"driver_version,omitempty"`
	Workers       int    `json:"workers"`
	Iterations    int    `json:"iterations"`
	Seed          uint64 `json:"seed"`
}

// InvariantSummary describes an invariant assertion outcome
type InvariantSummary struct {
	Name      string `json:"name"`
	Query     string `json:"query,omitempty"`
	Assertion string `json:"assertion,omitempty"`
	Actual    string `json:"actual,omitempty"`
}

// ExecutionSummary describes the outcome of the chaos execution
type ExecutionSummary struct {
	Status            string            `json:"status"` // "passed" or "failed"
	ExecutionStatus   string            `json:"execution_status,omitempty"`
	Success           bool              `json:"success"`
	ViolationDetected bool              `json:"violation_detected"`
	AnomalyType       string            `json:"anomaly_type"`
	DurationMS        int64             `json:"duration_ms"`
	TotalSchedules    int               `json:"total_schedules"`
	FailedSchedules   int               `json:"failed_schedules"`
	FailingInvariant  *InvariantSummary `json:"failing_invariant,omitempty"`
}

// SanitizedTraceEvent represents structural operation metadata. Table and SQL
// remain for decoding legacy payloads and are rejected by hosted ingestion.
type SanitizedTraceEvent struct {
	Worker   string `json:"worker"`
	OpType   string `json:"op_type"`         // "read", "write", "commit", "abort"
	Table    string `json:"table,omitempty"` // Deprecated: hosted payloads must omit this field.
	SQL      string `json:"sql,omitempty"`   // Deprecated: hosted payloads must omit this field.
	Duration int64  `json:"duration_us,omitempty"`
}

// ReproductionData contains reduction metrics and legacy detail fields. Hosted
// ingestion accepts only the numeric metrics.
type ReproductionData struct {
	MinimalOperationsCount int                   `json:"minimal_operations_count"`
	ShrinkDurationMS       int64                 `json:"shrink_duration_ms"`
	ReproGoCode            string                `json:"repro_go_code,omitempty"`
	MermaidDiagram         string                `json:"mermaid_diagram,omitempty"`
	SanitizedMinimalTrace  []SanitizedTraceEvent `json:"sanitized_minimal_trace,omitempty"`
}

// RunIngestRequest is the root JSON payload sent to POST /v1/runs
type RunIngestRequest struct {
	Version      string               `json:"version"`
	Timestamp    time.Time            `json:"timestamp"`
	CI           *CIContext           `json:"ci,omitempty"`
	Scenario     ScenarioMetadata     `json:"scenario"`
	Schedule     *domain.SchedulePlan `json:"schedule,omitempty"`
	Result       ExecutionSummary     `json:"result"`
	Reproduction *ReproductionData    `json:"reproduction,omitempty"`
}

// BaselineComparison describes whether this run represents a regression
type BaselineComparison struct {
	RunID     string `json:"run_id"`
	Status    string `json:"status"`
	CommitSHA string `json:"commit_sha"`
	Branch    string `json:"branch"`
}

// PRCommentStatus notes whether an automated PR comment was created
type PRCommentStatus struct {
	Posted    bool   `json:"posted"`
	CommentID string `json:"comment_id,omitempty"`
}

// RunIngestResponse is the response returned by POST /v1/runs
type RunIngestResponse struct {
	Success      bool                `json:"success"`
	RunID        string              `json:"run_id"`
	URL          string              `json:"url"`
	IsRegression bool                `json:"is_regression"`
	Baseline     *BaselineComparison `json:"baseline,omitempty"`
	PRComment    *PRCommentStatus    `json:"pr_comment,omitempty"`
	Message      string              `json:"message,omitempty"`
}
