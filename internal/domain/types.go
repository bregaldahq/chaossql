package domain

import (
	"fmt"
	"time"
)

// AnomalyType classifies the concurrency isolation anomaly according to Adya/Berenson.
type AnomalyType string

const (
	AnomalyLostUpdate      AnomalyType = "P4_LOST_UPDATE"
	AnomalyWriteSkew       AnomalyType = "A5B_WRITE_SKEW"
	AnomalyPhantom         AnomalyType = "A3_PHANTOM_READ"
	AnomalyA5AReadSkew     AnomalyType = "A5A_READ_SKEW"
	AnomalyG0DirtyWrite    AnomalyType = "G0_DIRTY_WRITE"
	AnomalyG1cCircularInfo AnomalyType = "G1C_CIRCULAR_INFO"
	AnomalyUnknown         AnomalyType = "UNKNOWN_INVARIANT_VIOLATION"
)

// Spec represents the complete declarative configuration for a chaos test.
type Spec struct {
	Version     string            `yaml:"version"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Database    DatabaseConfig    `yaml:"database"`
	Engine      EngineConfig      `yaml:"engine"`
	Invariants  []InvariantConfig `yaml:"invariants"`
	Operations  []OperationConfig `yaml:"operations"`
}

// DatabaseConfig holds connection and initialization paths.
type DatabaseConfig struct {
	Driver string `yaml:"driver"` // "sqlite", "postgres", or "mysql"
	DSN    string `yaml:"dsn,omitempty"`
	Schema string `yaml:"schema"`
	Seed   string `yaml:"seed"`
}

// EngineConfig holds concurrency, scheduler and PRNG settings.
type EngineConfig struct {
	Workers    int    `yaml:"workers"`
	Iterations int    `yaml:"iterations"`
	Seed       uint64 `yaml:"seed"`
	JitterMs   [2]int `yaml:"jitter_ms"`
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
	EventBegin    TraceEventType = "BEGIN"
	EventExec     TraceEventType = "EXEC"
	EventCommit   TraceEventType = "COMMIT"
	EventRollback TraceEventType = "ROLLBACK"
	EventError    TraceEventType = "ERROR"
)

// TraceEvent represents an immutable execution event in time.
type TraceEvent struct {
	Timestamp time.Duration  `json:"timestamp_us"`
	WorkerID  int            `json:"worker_id"`
	OpIndex   int            `json:"op_index"`
	OpName    string         `json:"op_name"`
	StepIndex int            `json:"step_index"`
	Type      TraceEventType `json:"type"`
	SQL       string         `json:"sql"`
	Error     string         `json:"error,omitempty"`
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

// ShrinkResult summarizes the output of the Delta-Debugging algorithm.
type ShrinkResult struct {
	OriginalSize   int           `json:"original_size"`
	ReducedSize    int           `json:"reduced_size"`
	ReductionRatio float64       `json:"reduction_ratio"`
	MinimalOps     []ScheduledOp `json:"minimal_ops"`
	Iterations     int           `json:"iterations"`
	Duration       time.Duration `json:"duration"`
}
