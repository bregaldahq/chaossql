package domain

import (
	"fmt"
	"time"
)

// AnomalyType classifies the concurrency isolation anomaly according to Adya/Berenson.
type AnomalyType string

const (
	AnomalyLostUpdate          AnomalyType = "P4_LOST_UPDATE"
	AnomalyWriteSkew           AnomalyType = "A5B_WRITE_SKEW"
	AnomalyPhantom             AnomalyType = "A3_PHANTOM_READ"
	AnomalyA5AReadSkew         AnomalyType = "A5A_READ_SKEW"
	AnomalyG0DirtyWrite        AnomalyType = "G0_DIRTY_WRITE"
	AnomalyG1aDirtyRead        AnomalyType = "G1A_DIRTY_READ"
	AnomalyG1bIntermediateRead AnomalyType = "G1B_INTERMEDIATE_READ"
	AnomalyFracturedRead       AnomalyType = "FRACTURED_READ"
	AnomalyG1cCircularInfo     AnomalyType = "G1C_CIRCULAR_INFO"
	AnomalyG2AntiDependency    AnomalyType = "G2_ANTI_DEPENDENCY"
	AnomalyUnknown             AnomalyType = "UNKNOWN_INVARIANT_VIOLATION"
)

// IsolationLevel represents standard SQL transaction isolation levels.
type IsolationLevel string

const (
	LevelReadUncommitted IsolationLevel = "READ_UNCOMMITTED"
	LevelReadCommitted   IsolationLevel = "READ_COMMITTED"
	LevelRepeatableRead  IsolationLevel = "REPEATABLE_READ"
	LevelSerializable    IsolationLevel = "SERIALIZABLE"
)

// ExecutionStatus describes whether a run produced a trustworthy conclusion.
type ExecutionStatus string

const (
	StatusPassed         ExecutionStatus = "passed"
	StatusViolation      ExecutionStatus = "violation"
	StatusExecutionError ExecutionStatus = "execution_error"
	StatusInconclusive   ExecutionStatus = "inconclusive"
	StatusCanceled       ExecutionStatus = "canceled"
)

// TemporalInvariantConfig defines rules evaluated against chronological traces.
type TemporalInvariantConfig struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"` // "no_aborts", "monotonicity", "no_error_events"
	Field  string `yaml:"field,omitempty"`
	Assert string `yaml:"assert,omitempty"`
}

// Spec represents the complete declarative configuration for a chaos test.
type Spec struct {
	Version            string                    `yaml:"version"`
	Name               string                    `yaml:"name"`
	Description        string                    `yaml:"description"`
	Database           DatabaseConfig            `yaml:"database"`
	Engine             EngineConfig              `yaml:"engine"`
	Invariants         []InvariantConfig         `yaml:"invariants"`
	TemporalInvariants []TemporalInvariantConfig `yaml:"temporal_invariants,omitempty"`
	Operations         []OperationConfig         `yaml:"operations"`
}

// Validate verifies that the Spec has all mandatory fields and valid configurations.
func (s Spec) Validate() error {
	if s.Version == "" {
		return fmt.Errorf("%w: missing or empty 'version'", ErrSpecValidationFailed)
	}
	if s.Name == "" {
		return fmt.Errorf("%w: missing or empty 'name'", ErrSpecValidationFailed)
	}
	if s.Database.Driver == "" {
		return fmt.Errorf("%w: missing or empty 'database.driver'", ErrSpecValidationFailed)
	}
	switch s.Database.Isolation {
	case "", LevelReadUncommitted, LevelReadCommitted, LevelRepeatableRead, LevelSerializable:
	default:
		return fmt.Errorf("%w: unsupported database isolation %q", ErrSpecValidationFailed, s.Database.Isolation)
	}
	if len(s.Invariants) == 0 {
		return fmt.Errorf("%w: 'invariants' must have at least one entry", ErrSpecValidationFailed)
	}
	if len(s.Operations) == 0 {
		return fmt.Errorf("%w: 'operations' must have at least one entry", ErrSpecValidationFailed)
	}
	if err := s.ValidateInvariantNames(); err != nil {
		return err
	}
	for i, op := range s.Operations {
		if op.Name == "" {
			return fmt.Errorf("%w: operation[%d] missing name", ErrSpecValidationFailed, i)
		}
	}
	if s.Engine.JitterMs[0] < 0 || s.Engine.JitterMs[1] < s.Engine.JitterMs[0] {
		return fmt.Errorf("%w: invalid jitter range [%d, %d]", ErrSpecValidationFailed, s.Engine.JitterMs[0], s.Engine.JitterMs[1])
	}
	return nil
}

// ValidateInvariantNames ensures failure signatures can identify one invariant.
// Runner entry points call it even when they accept an otherwise partial Spec.
func (s Spec) ValidateInvariantNames() error {
	invariantNames := make(map[string]struct{}, len(s.Invariants))
	for i, inv := range s.Invariants {
		if inv.Name == "" {
			return fmt.Errorf("%w: invariant[%d] missing name", ErrSpecValidationFailed, i)
		}
		if _, exists := invariantNames[inv.Name]; exists {
			return fmt.Errorf("%w: duplicate invariant name %q", ErrSpecValidationFailed, inv.Name)
		}
		invariantNames[inv.Name] = struct{}{}
	}
	return nil
}

// DatabaseConfig holds connection and initialization paths.
type DatabaseConfig struct {
	Driver    string         `yaml:"driver"` // "sqlite", "postgres", or "mysql"
	DSN       string         `yaml:"dsn,omitempty"`
	Isolation IsolationLevel `yaml:"isolation,omitempty"`
	Schema    string         `yaml:"schema"`
	Seed      string         `yaml:"seed"`
}

// FaultConfig defines stochastic fault injection parameters.
type FaultConfig struct {
	AbortProbability      float64 `yaml:"abort_probability,omitempty"`
	LatencySpikeMs        [2]int  `yaml:"latency_spike_ms,omitempty"`
	LatencyProbability    float64 `yaml:"latency_probability,omitempty"`
	DisconnectProbability float64 `yaml:"disconnect_probability,omitempty"`
}

// EngineConfig holds concurrency, scheduler, PRNG and fault injection settings.
type EngineConfig struct {
	Workers    int         `yaml:"workers"`
	Iterations int         `yaml:"iterations"`
	Seed       uint64      `yaml:"seed"`
	JitterMs   [2]int      `yaml:"jitter_ms"`
	Faults     FaultConfig `yaml:"faults,omitempty"`
}

// InvariantConfig represents a mathematical rule checked against the database state.
type InvariantConfig struct {
	Name   string `yaml:"name"`
	Query  string `yaml:"query"`
	Assert string `yaml:"assert"`
}

// OperationConfig defines a template of SQL steps executed within a transaction.
type OperationConfig struct {
	Name   string            `yaml:"name"`
	Weight float64           `yaml:"weight"`
	Params map[string]string `yaml:"params,omitempty"`
	Steps  []StepConfig      `yaml:"steps"`
}

// StepConfig defines a single SQL instruction in an operation.
type StepConfig struct {
	SQL     string `yaml:"sql"`
	Capture string `yaml:"capture,omitempty"`
}

// TraceEventType represents the lifecycle state of a transaction step.
type TraceEventType string

const (
	EventBegin            TraceEventType = "BEGIN"
	EventExec             TraceEventType = "EXEC"
	EventCommit           TraceEventType = "COMMIT"
	EventRollback         TraceEventType = "ROLLBACK"
	EventError            TraceEventType = "ERROR"
	EventSavepoint        TraceEventType = "SAVEPOINT"
	EventRollbackTo       TraceEventType = "ROLLBACK_TO"
	EventReleaseSavepoint TraceEventType = "RELEASE_SAVEPOINT"
)

// TraceEvent represents an immutable execution event in time.
type TraceEvent struct {
	Timestamp time.Duration  `json:"timestamp_us"`
	WorkerID  int            `json:"worker_id"`
	OpIndex   int            `json:"op_index"`
	OpName    string         `json:"op_name"`
	StepIndex int            `json:"step_index"`
	Type      TraceEventType `json:"type"`
	Phase     string         `json:"phase,omitempty"`
	SQL       string         `json:"sql"`
	Error     string         `json:"error,omitempty"`
}

// OperationError identifies the operation lifecycle action that failed.
type OperationError struct {
	OperationID int    `json:"operation_id"`
	Operation   string `json:"operation"`
	StepIndex   int    `json:"step_index,omitempty"`
	Phase       string `json:"phase"`
	Message     string `json:"message"`
}

// ExecutionTrace is an ordered log of concurrency events.
type ExecutionTrace []TraceEvent

// InvariantResult holds the outcome of evaluating an invariant.
type InvariantResult struct {
	Name         string                 `json:"name"`
	Passed       bool                   `json:"passed"`
	Expression   string                 `json:"expression"`
	ActualValues map[string]interface{} `json:"actual_values"`
	Error        error                  `json:"error,omitempty"`
}

func (r InvariantResult) String() string {
	if r.Passed {
		return fmt.Sprintf("PASS: Invariant '%s' satisfied", r.Name)
	}
	return fmt.Sprintf("FAIL: Invariant '%s' violated! Expr: (%s), State: %v", r.Name, r.Expression, r.ActualValues)
}

// ScheduledOp represents a concrete, parameter-bound operation instance.
type ScheduledOp struct {
	ID     int               `json:"id"`
	Name   string            `json:"name"`
	Params map[string]string `json:"params"`
	Steps  []StepConfig      `json:"steps"`
}

// SchedulePlan records the deterministic decisions controlled by the harness.
type SchedulePlan struct {
	Version   int                `json:"version"`
	Seed      uint64             `json:"seed"`
	Workers   int                `json:"workers"`
	Decisions []ScheduleDecision `json:"decisions"`
}

// ScheduleDecision records the worker and injected behavior for one SQL step.
type ScheduleDecision struct {
	Sequence    int  `json:"sequence"`
	OperationID int  `json:"operation_id"`
	WorkerID    int  `json:"worker_id"`
	StepIndex   int  `json:"step_index"`
	JitterMs    int  `json:"jitter_ms"`
	LatencyMs   int  `json:"latency_ms"`
	Abort       bool `json:"abort"`
}

// FailureSignature identifies the stable outcome that a replay or shrink candidate must reproduce.
type FailureSignature struct {
	Status           ExecutionStatus `json:"status"`
	FailingInvariant string          `json:"failing_invariant,omitempty"`
}

// ShrinkResult summarizes the output of the Delta-Debugging algorithm.
type ShrinkResult struct {
	OriginalSize   int           `json:"original_size"`
	ReducedSize    int           `json:"reduced_size"`
	ReductionRatio float64       `json:"reduction_ratio"`
	MinimalOps     []ScheduledOp `json:"minimal_ops"`
	Iterations     int           `json:"iterations"`
	Trials         int           `json:"trials"`
	Duration       time.Duration `json:"duration"`
}

// ExecutionResult holds the complete outcome of a chaos execution.
type ExecutionResult struct {
	Status            ExecutionStatus  `json:"status"`
	Isolation         IsolationLevel   `json:"isolation,omitempty"`
	Seed              uint64           `json:"seed"`
	Schedule          SchedulePlan     `json:"schedule"`
	OperationErrors   []OperationError `json:"operation_errors,omitempty"`
	Success           bool             `json:"success"`
	ViolationDetected bool             `json:"violation_detected"`
	FailingInvariant  *InvariantResult `json:"failing_invariant,omitempty"`
	Trace             ExecutionTrace   `json:"trace,omitempty"`
	ScheduledOps      []ScheduledOp    `json:"scheduled_ops,omitempty"`
	Duration          time.Duration    `json:"duration"`
	Error             error            `json:"error,omitempty"`
}

// DiffResult represents the outcome of cross-engine differential fuzzing.
type DiffResult struct {
	ScenarioName string           `json:"scenario_name"`
	DriverA      string           `json:"driver_a"`
	DriverB      string           `json:"driver_b"`
	ResultA      *ExecutionResult `json:"result_a"`
	ResultB      *ExecutionResult `json:"result_b"`
	Divergent    bool             `json:"divergent"`
	DiffSummary  string           `json:"diff_summary"`
}
