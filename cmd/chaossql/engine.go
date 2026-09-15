package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
	"github.com/spf13/cobra"
)

// IPCOperationStep mirrors domain.StepConfig with flexible JSON tags.
type IPCOperationStep struct {
	SQL     string `json:"sql"`
	Capture string `json:"capture,omitempty"`
}

// IPCOperation mirrors domain.OperationConfig with flexible JSON tags.
type IPCOperation struct {
	Name   string             `json:"name"`
	Weight float64            `json:"weight,omitempty"`
	Params map[string]string  `json:"params,omitempty"`
	Steps  []IPCOperationStep `json:"steps"`
}

// IPCInvariant mirrors domain.InvariantConfig with flexible JSON tags.
type IPCInvariant struct {
	Name   string `json:"name"`
	Query  string `json:"query"`
	Assert string `json:"assert"`
}

// IPCEngineConfig mirrors domain.EngineConfig with flexible JSON tags.
type IPCEngineConfig struct {
	Workers    int     `json:"workers,omitempty"`
	Iterations int     `json:"iterations,omitempty"`
	Seed       *uint64 `json:"seed,omitempty"`
	JitterMs   [2]int  `json:"jitter_ms,omitempty"`
}

// IPCDatabaseConfig mirrors domain.DatabaseConfig with flexible JSON tags.
type IPCDatabaseConfig struct {
	Driver    string                `json:"driver,omitempty"`
	DSN       string                `json:"dsn,omitempty"`
	Isolation domain.IsolationLevel `json:"isolation,omitempty"`
	Schema    string                `json:"schema,omitempty"`
	Seed      string                `json:"seed,omitempty"`
}

// IPCPayload represents an incoming execution request via JSON IPC.
type IPCPayload struct {
	// Root or nested configuration
	Version    string                `json:"version,omitempty"`
	Name       string                `json:"name,omitempty"`
	Driver     string                `json:"driver,omitempty"`
	DSN        string                `json:"dsn,omitempty"`
	Isolation  domain.IsolationLevel `json:"isolation,omitempty"`
	Schema     string                `json:"schema,omitempty"`
	Seed       string                `json:"seed,omitempty"`
	Workers    int                   `json:"workers,omitempty"`
	Iterations int                   `json:"iterations,omitempty"`
	SeedValue  *uint64               `json:"seed_value,omitempty"`
	Database   *IPCDatabaseConfig    `json:"database,omitempty"`
	Engine     *IPCEngineConfig      `json:"engine,omitempty"`
	Invariants []IPCInvariant        `json:"invariants"`
	Operations []IPCOperation        `json:"operations"`
}

// IPCResponse represents the structured JSON output returned over stdout.
type IPCResponse struct {
	Status            domain.ExecutionStatus  `json:"status"`
	Isolation         domain.IsolationLevel   `json:"isolation,omitempty"`
	Seed              uint64                  `json:"seed"`
	Schedule          domain.SchedulePlan     `json:"schedule"`
	OperationErrors   []domain.OperationError `json:"operation_errors,omitempty"`
	Success           bool                    `json:"success"`
	ViolationDetected bool                    `json:"violation_detected"`
	AnomalyType       domain.AnomalyType      `json:"anomaly_type,omitempty"`
	FailingInvariant  *domain.InvariantResult `json:"failing_invariant,omitempty"`
	DurationMs        int64                   `json:"duration_ms"`
	TraceEventsCount  int                     `json:"trace_events_count"`
	Shrink            *domain.ShrinkResult    `json:"shrink,omitempty"`
	MinimalOperations []domain.ScheduledOp    `json:"minimal_operations,omitempty"`
	Mermaid           string                  `json:"mermaid,omitempty"`
	ReproGo           string                  `json:"repro_go,omitempty"`
	ReproPython       string                  `json:"repro_python,omitempty"`
	ReproTypeScript   string                  `json:"repro_typescript,omitempty"`
	Error             string                  `json:"error,omitempty"`
}

func canceledIPCResponse(result *domain.ExecutionResult, err error) IPCResponse {
	response := IPCResponse{Status: domain.StatusCanceled, Success: false}
	if err != nil {
		response.Error = err.Error()
	}
	if result == nil {
		return response
	}
	response.Isolation = result.Isolation
	response.Seed = result.Seed
	response.Schedule = result.Schedule
	response.OperationErrors = result.OperationErrors
	response.DurationMs = result.Duration.Milliseconds()
	response.TraceEventsCount = len(result.Trace)
	return response
}

func newEngineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "engine",
		Short: "Execute a chaos test scenario received as JSON over stdin and stream results over stdout",
		Long: `The engine subcommand provides a bidirectional IPC protocol for host language SDKs
(Python chaossql-py, TypeScript @chaossql/test). It consumes a JSON scenario definition over stdin,
runs deterministic concurrency fuzzing and delta-debugging, and outputs a structured JSON report
to stdout with zero external runtime dependencies.`,
		RunE: runEngineIPC,
	}

	return cmd
}

func runEngineIPC(cmd *cobra.Command, args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dec := json.NewDecoder(cmd.InOrStdin())
	enc := json.NewEncoder(cmd.OutOrStdout())

	for {
		var payload IPCPayload
		if err := dec.Decode(&payload); err != nil {
			if err == io.EOF {
				break
			}
			resp := IPCResponse{
				Status:  domain.StatusExecutionError,
				Success: false,
				Error:   fmt.Sprintf("JSON decoding error: %v", err),
			}
			_ = enc.Encode(resp)
			return err
		}

		resp := executeIPCPayload(ctx, payload)
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}

		// If input was a one-shot JSON without stream continuation, break
		if !dec.More() {
			break
		}
	}

	return nil
}

func executeIPCPayload(ctx context.Context, p IPCPayload) IPCResponse {
	// Normalize domain.Spec
	driverName := p.Driver
	dsn := p.DSN
	isolation := p.Isolation
	schema := p.Schema
	seedSQL := p.Seed

	if p.Database != nil {
		if p.Database.Driver != "" {
			driverName = p.Database.Driver
		}
		if p.Database.DSN != "" {
			dsn = p.Database.DSN
		}
		if p.Database.Isolation != "" {
			isolation = p.Database.Isolation
		}
		if p.Database.Schema != "" {
			schema = p.Database.Schema
		}
		if p.Database.Seed != "" {
			seedSQL = p.Database.Seed
		}
	}

	if driverName == "" {
		driverName = "sqlite"
	}
	if driverName == "sqlite" && (dsn == "" || dsn == ":memory:") {
		dsn = fmt.Sprintf("file:chaossql_ipc_%d?mode=memory&cache=shared", time.Now().UnixNano())
	}

	workers := p.Workers
	iterations := p.Iterations
	jitter := [2]int{1, 5}

	var seed uint64
	if p.SeedValue != nil {
		seed = *p.SeedValue
	}

	if p.Engine != nil {
		if p.Engine.Workers > 0 {
			workers = p.Engine.Workers
		}
		if p.Engine.Iterations > 0 {
			iterations = p.Engine.Iterations
		}
		if p.Engine.Seed != nil {
			seed = *p.Engine.Seed
		}
		if p.Engine.JitterMs[1] > 0 {
			jitter = p.Engine.JitterMs
		}
	}

	if workers <= 0 {
		workers = 2
	}
	if iterations <= 0 {
		iterations = 10
	}
	specName := p.Name
	if specName == "" {
		specName = "chaossql_ipc_scenario"
	}
	version := p.Version
	if version == "" {
		version = "1.1"
	}

	var invariants []domain.InvariantConfig
	for _, inv := range p.Invariants {
		invariants = append(invariants, domain.InvariantConfig{
			Name:   inv.Name,
			Query:  inv.Query,
			Assert: inv.Assert,
		})
	}

	var operations []domain.OperationConfig
	for _, op := range p.Operations {
		var steps []domain.StepConfig
		for _, s := range op.Steps {
			sqlStmt := strings.TrimSpace(s.SQL)
			capVar := strings.TrimSpace(s.Capture)

			// Also support inline syntax "-> var" or "=> var"
			if capVar == "" {
				if strings.Contains(sqlStmt, "->") {
					parts := strings.SplitN(sqlStmt, "->", 2)
					sqlStmt = strings.TrimSpace(parts[0])
					capVar = strings.TrimSpace(parts[1])
				} else if strings.Contains(sqlStmt, "=>") {
					parts := strings.SplitN(sqlStmt, "=>", 2)
					sqlStmt = strings.TrimSpace(parts[0])
					capVar = strings.TrimSpace(parts[1])
				}
			}

			steps = append(steps, domain.StepConfig{
				SQL:     sqlStmt,
				Capture: capVar,
			})
		}

		weight := op.Weight
		if weight <= 0 {
			weight = 1.0
		}

		operations = append(operations, domain.OperationConfig{
			Name:   op.Name,
			Weight: weight,
			Params: op.Params,
			Steps:  steps,
		})
	}

	spec := domain.Spec{
		Version: version,
		Name:    specName,
		Database: domain.DatabaseConfig{
			Driver:    driverName,
			DSN:       dsn,
			Isolation: isolation,
			Schema:    schema,
			Seed:      seedSQL,
		},
		Engine: domain.EngineConfig{
			Workers:    workers,
			Iterations: iterations,
			Seed:       seed,
			JitterMs:   jitter,
		},
		Invariants: invariants,
		Operations: operations,
	}

	driver, err := drivers.GetDriver(spec.Database.Driver, spec.Database.DSN)
	if err != nil {
		return IPCResponse{Status: domain.StatusExecutionError, Success: false, Error: fmt.Sprintf("database driver error: %v", err)}
	}

	if err := driver.Open(ctx); err != nil {
		return IPCResponse{Status: domain.StatusExecutionError, Success: false, Error: fmt.Sprintf("failed to open database driver: %v", err)}
	}
	defer func() { _ = driver.Close() }()

	runner := engine.NewRunner(driver, spec.Engine.Seed)
	runResult, err := runner.Run(ctx, spec)
	if err != nil {
		if runResult != nil {
			return IPCResponse{
				Status:            runResult.Status,
				Isolation:         runResult.Isolation,
				Seed:              runResult.Seed,
				Schedule:          runResult.Schedule,
				OperationErrors:   runResult.OperationErrors,
				Success:           runResult.Success,
				ViolationDetected: runResult.ViolationDetected,
				FailingInvariant:  runResult.FailingInvariant,
				DurationMs:        runResult.Duration.Milliseconds(),
				TraceEventsCount:  len(runResult.Trace),
				Error:             err.Error(),
			}
		}
		return IPCResponse{Status: domain.StatusExecutionError, Success: false, Error: fmt.Sprintf("chaos execution failed: %v", err)}
	}

	graph := analyzer.BuildGraph(runResult.Trace)
	cycles := analyzer.FindCycles(graph)
	anomaly := domain.AnomalyUnknown
	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls != domain.AnomalyUnknown {
			anomaly = cls
			break
		}
	}
	if anomaly == domain.AnomalyUnknown && len(cycles) > 0 {
		anomaly = analyzer.ClassifyCycle(cycles[0])
	}

	var shrinkResult *domain.ShrinkResult
	minimalTrace := runResult.Trace
	minimalOps := runResult.ScheduledOps

	if runResult.ViolationDetected {
		target, _ := shrinker.FailureSignatureFor(runResult)
		testFn := func(subset []domain.ScheduledOp) bool {
			res, err := runner.RunSchedule(ctx, spec, subset)
			if err != nil {
				return true
			}
			return !shrinker.ReproducesFailure(res, target)
		}

		shrunk, err := shrinker.Shrink(ctx, testFn, runResult.ScheduledOps)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return canceledIPCResponse(runResult, err)
		}
		if err == nil && shrunk != nil {
			candidateOps := shrunk.MinimalOps
			minRunRes, err := runner.RunSchedule(ctx, spec, candidateOps)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return canceledIPCResponse(runResult, err)
			}
			if err == nil && shrinker.ReproducesFailure(minRunRes, target) {
				shrinkResult = shrunk
				minimalOps = candidateOps
				minimalTrace = minRunRes.Trace
				minGraph := analyzer.BuildGraph(minimalTrace)
				minCycles := analyzer.FindCycles(minGraph)
				for _, c := range minCycles {
					cls := analyzer.ClassifyCycle(c)
					if cls != domain.AnomalyUnknown {
						anomaly = cls
						break
					}
				}
			}
		}
	}

	mermaidCode := reporter.GenerateMermaidSequence(minimalTrace)
	reproGo := reporter.GenerateStandaloneGoRepro(spec, minimalOps, runResult.FailingInvariant)
	reproPython := reporter.GenerateStandalonePythonRepro(spec, minimalOps, runResult.FailingInvariant)
	reproTS := reporter.GenerateStandaloneTypeScriptRepro(spec, minimalOps, runResult.FailingInvariant)

	errorMessage := ""
	if runResult.Error != nil {
		errorMessage = runResult.Error.Error()
	}
	return IPCResponse{
		Status:            runResult.Status,
		Isolation:         runResult.Isolation,
		Seed:              runResult.Seed,
		Schedule:          runResult.Schedule,
		OperationErrors:   runResult.OperationErrors,
		Success:           runResult.Success,
		ViolationDetected: runResult.ViolationDetected,
		AnomalyType:       anomaly,
		FailingInvariant:  runResult.FailingInvariant,
		DurationMs:        runResult.Duration.Milliseconds(),
		TraceEventsCount:  len(runResult.Trace),
		Shrink:            shrinkResult,
		MinimalOperations: minimalOps,
		Mermaid:           mermaidCode,
		ReproGo:           reproGo,
		ReproPython:       reproPython,
		ReproTypeScript:   reproTS,
		Error:             errorMessage,
	}
}
