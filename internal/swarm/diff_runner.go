package swarm

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
)

// DriverExecutionResult holds the execution outcome on a specific database driver.
type DriverExecutionResult struct {
	Driver            string `json:"driver"`
	Success           bool   `json:"success"`
	ViolationDetected bool   `json:"violation_detected"`
	FailingInvariant  string `json:"failing_invariant,omitempty"`
	DetectedAnomaly   string `json:"detected_anomaly,omitempty"`
	DurationMs        int64  `json:"duration_ms"`
	Error             string `json:"error,omitempty"`
}

// ScenarioDifferential records the cross-driver comparison outcome for a single scenario.
type ScenarioDifferential struct {
	ScenarioName string                           `json:"scenario_name"`
	Divergent    bool                             `json:"divergent"`
	Summary      string                           `json:"summary"`
	Results      map[string]DriverExecutionResult `json:"results"`
}

// DifferentialReport aggregates the complete cross-engine differential matrix execution results.
type DifferentialReport struct {
	TotalScenarios  int                    `json:"total_scenarios"`
	DivergentCount  int                    `json:"divergent_count"`
	TotalExecutions int                    `json:"total_executions"`
	Scenarios       []ScenarioDifferential `json:"scenarios"`
	DurationMs      int64                  `json:"duration_ms"`
}

type driverTask struct {
	specIndex  int
	driverName string
}

type taskOutcome struct {
	specIndex  int
	driverName string
	result     DriverExecutionResult
	err        error
}

// ExecuteDifferentialMatrix runs a matrix of scenarios across multiple database drivers concurrently,
// comparing outcomes to detect semantic divergence in transaction isolation and invariant satisfaction.
// Optional concurrency parameter defaults to 4.
func ExecuteDifferentialMatrix(ctx context.Context, specs []domain.Spec, driverNames []string, concurrencyOpt ...int) (*DifferentialReport, error) {
	concurrency := 4
	if len(concurrencyOpt) > 0 && concurrencyOpt[0] > 0 {
		concurrency = concurrencyOpt[0]
	}
	startTime := time.Now()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(specs) == 0 {
		return &DifferentialReport{
			TotalScenarios:  0,
			DivergentCount:  0,
			TotalExecutions: 0,
			Scenarios:       []ScenarioDifferential{},
			DurationMs:      0,
		}, nil
	}

	if len(driverNames) == 0 {
		driverNames = []string{"sqlite"}
	}

	if concurrency <= 0 {
		concurrency = 4
	}

	// 1. Pre-generate deterministic schedules for each scenario so all drivers receive identical operations.
	schedules := make([][]domain.ScheduledOp, len(specs))
	for i, spec := range specs {
		prng := engine.NewPRNG(spec.Engine.Seed)
		schedules[i] = engine.GenerateSchedule(spec, prng)
	}

	// 2. Build task list of (scenario, driver) pairs.
	var tasks []driverTask
	for specIdx := range specs {
		for _, dName := range driverNames {
			tasks = append(tasks, driverTask{
				specIndex:  specIdx,
				driverName: dName,
			})
		}
	}

	taskChan := make(chan driverTask, len(tasks))
	for _, t := range tasks {
		taskChan <- t
	}
	close(taskChan)

	numWorkers := concurrency
	if numWorkers > len(tasks) {
		numWorkers = len(tasks)
	}

	outcomeChan := make(chan taskOutcome, len(tasks))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskChan {
				if ctx.Err() != nil {
					outcomeChan <- taskOutcome{
						specIndex:  t.specIndex,
						driverName: t.driverName,
						err:        ctx.Err(),
					}
					return
				}

				res, err := executeDriverRun(ctx, specs[t.specIndex], schedules[t.specIndex], t.driverName)
				outcomeChan <- taskOutcome{
					specIndex:  t.specIndex,
					driverName: t.driverName,
					result:     res,
					err:        err,
				}
			}
		}()
	}

	wg.Wait()
	close(outcomeChan)

	// 3. Process outcomes.
	scenarioResults := make([]map[string]DriverExecutionResult, len(specs))
	for i := range scenarioResults {
		scenarioResults[i] = make(map[string]DriverExecutionResult)
	}

	var cancelErr error
	totalExecutions := 0

	for out := range outcomeChan {
		if out.err != nil {
			if cancelErr == nil && (errors.Is(out.err, context.Canceled) || errors.Is(out.err, context.DeadlineExceeded)) {
				cancelErr = out.err
			}
			continue
		}
		scenarioResults[out.specIndex][out.driverName] = out.result
		totalExecutions++
	}

	if cancelErr != nil {
		return nil, cancelErr
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// 4. Compute divergence per scenario.
	differentials := make([]ScenarioDifferential, len(specs))
	divergentCount := 0

	for i, spec := range specs {
		diff := EvaluateScenarioDivergence(spec.Name, scenarioResults[i], driverNames)
		if diff.Divergent {
			divergentCount++
		}
		differentials[i] = diff
	}

	return &DifferentialReport{
		TotalScenarios:  len(specs),
		DivergentCount:  divergentCount,
		TotalExecutions: totalExecutions,
		Scenarios:       differentials,
		DurationMs:      time.Since(startTime).Milliseconds(),
	}, nil
}

func executeDriverRun(ctx context.Context, spec domain.Spec, ops []domain.ScheduledOp, driverName string) (DriverExecutionResult, error) {
	if ctx.Err() != nil {
		return DriverExecutionResult{}, ctx.Err()
	}

	start := time.Now()

	dsn := ""
	if strings.EqualFold(spec.Database.Driver, driverName) && spec.Database.DSN != "" {
		dsn = spec.Database.DSN
	}
	if dsn == "" {
		switch strings.ToLower(driverName) {
		case "postgres", "postgresql":
			if env := os.Getenv("DATABASE_URL"); env != "" && strings.HasPrefix(env, "postgres") {
				dsn = env
			} else if env := os.Getenv("POSTGRES_DSN"); env != "" {
				dsn = env
			}
		case "mysql", "mariadb":
			if env := os.Getenv("MYSQL_DSN"); env != "" {
				dsn = env
			} else if env := os.Getenv("DATABASE_URL"); env != "" && strings.HasPrefix(env, "mysql") {
				dsn = env
			}
		}
	}

	driver, err := drivers.GetDriver(driverName, dsn)
	if err != nil {
		return DriverExecutionResult{
			Driver:     driverName,
			Success:    false,
			DurationMs: time.Since(start).Milliseconds(),
			Error:      fmt.Sprintf("driver initialization failed: %v", err),
		}, nil
	}
	defer func() {
		_ = driver.Close()
	}()

	runCtx, runCancel := context.WithTimeout(ctx, 15*time.Second)
	defer runCancel()

	runner := engine.NewRunner(driver, spec.Engine.Seed)
	runResult, err := runner.RunSchedule(runCtx, spec, ops)
	if err != nil {
		if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return DriverExecutionResult{}, ctx.Err()
		}
		return DriverExecutionResult{
			Driver:     driverName,
			Success:    false,
			DurationMs: time.Since(start).Milliseconds(),
			Error:      err.Error(),
		}, nil
	}

	anomaly := detectAnomalyFromTrace(runResult.Trace)
	failingInv := ""
	if runResult.FailingInvariant != nil {
		failingInv = runResult.FailingInvariant.Name
	}
	if runResult.ViolationDetected && anomaly == "" {
		anomaly = string(domain.AnomalyUnknown)
	}

	return DriverExecutionResult{
		Driver:            driverName,
		Success:           runResult.Success,
		ViolationDetected: runResult.ViolationDetected,
		FailingInvariant:  failingInv,
		DetectedAnomaly:   anomaly,
		DurationMs:        runResult.Duration.Milliseconds(),
	}, nil
}

func detectAnomalyFromTrace(trace domain.ExecutionTrace) string {
	if len(trace) == 0 {
		return ""
	}
	graph := analyzer.BuildGraph(trace)
	cycles := analyzer.FindCycles(graph)
	anomaly := domain.AnomalyUnknown
	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls == domain.AnomalyG1aDirtyRead {
			return string(domain.AnomalyG1aDirtyRead)
		}
		if cls == domain.AnomalyG0DirtyWrite {
			return string(domain.AnomalyG0DirtyWrite)
		}
		if cls == domain.AnomalyG1cCircularInfo {
			return string(domain.AnomalyG1cCircularInfo)
		}
		if cls == domain.AnomalyG2AntiDependency {
			return string(domain.AnomalyG2AntiDependency)
		}
		if cls == domain.AnomalyWriteSkew {
			return string(domain.AnomalyWriteSkew)
		}
		if cls == domain.AnomalyA5AReadSkew {
			return string(domain.AnomalyA5AReadSkew)
		}
		if cls == domain.AnomalyLostUpdate {
			anomaly = domain.AnomalyLostUpdate
		}
	}
	if anomaly != domain.AnomalyUnknown {
		return string(anomaly)
	}
	if len(cycles) > 0 {
		return string(analyzer.ClassifyCycle(cycles[0]))
	}
	return ""
}

func EvaluateScenarioDivergence(scenarioName string, results map[string]DriverExecutionResult, driverOrder []string) ScenarioDifferential {
	var validDrivers []string
	for _, d := range driverOrder {
		if r, ok := results[d]; ok && r.Error == "" {
			validDrivers = append(validDrivers, d)
		}
	}

	if len(validDrivers) == 0 {
		return ScenarioDifferential{
			ScenarioName: scenarioName,
			Divergent:    false,
			Summary:      "All drivers encountered execution or connection errors",
			Results:      results,
		}
	}

	if len(validDrivers) == 1 {
		r := results[validDrivers[0]]
		status := "satisfied invariants"
		if r.ViolationDetected {
			status = fmt.Sprintf("violation detected (%s)", r.FailingInvariant)
		}
		return ScenarioDifferential{
			ScenarioName: scenarioName,
			Divergent:    false,
			Summary:      fmt.Sprintf("Single active driver %s: %s", validDrivers[0], status),
			Results:      results,
		}
	}

	divergent := false
	var divergenceReasons []string

	for i := 0; i < len(validDrivers); i++ {
		for j := i + 1; j < len(validDrivers); j++ {
			dA := validDrivers[i]
			dB := validDrivers[j]
			rA := results[dA]
			rB := results[dB]

			if rA.ViolationDetected != rB.ViolationDetected {
				divergent = true
				divergenceReasons = append(divergenceReasons,
					fmt.Sprintf("Driver %s (violation=%v) != Driver %s (violation=%v)", dA, rA.ViolationDetected, dB, rB.ViolationDetected))
			} else if rA.ViolationDetected && rB.ViolationDetected {
				if rA.FailingInvariant != "" && rB.FailingInvariant != "" && rA.FailingInvariant != rB.FailingInvariant {
					divergent = true
					divergenceReasons = append(divergenceReasons,
						fmt.Sprintf("Differing failing invariants: %s (%s) vs %s (%s)", dA, rA.FailingInvariant, dB, rB.FailingInvariant))
				} else if rA.DetectedAnomaly != "" && rB.DetectedAnomaly != "" && rA.DetectedAnomaly != rB.DetectedAnomaly {
					divergent = true
					divergenceReasons = append(divergenceReasons,
						fmt.Sprintf("Differing anomaly classifications: %s (%s) vs %s (%s)", dA, rA.DetectedAnomaly, dB, rB.DetectedAnomaly))
				}
			}
		}
	}

	if divergent {
		return ScenarioDifferential{
			ScenarioName: scenarioName,
			Divergent:    true,
			Summary:      fmt.Sprintf("Semantic divergence detected: %s", strings.Join(divergenceReasons, "; ")),
			Results:      results,
		}
	}

	firstValid := results[validDrivers[0]]
	if firstValid.ViolationDetected {
		summary := fmt.Sprintf("All engines consistently detected invariant violation (%s)", firstValid.FailingInvariant)
		if firstValid.DetectedAnomaly != "" {
			summary = fmt.Sprintf("All engines consistently detected invariant violation (%s, %s)", firstValid.FailingInvariant, firstValid.DetectedAnomaly)
		}
		return ScenarioDifferential{
			ScenarioName: scenarioName,
			Divergent:    false,
			Summary:      summary,
			Results:      results,
		}
	}

	return ScenarioDifferential{
		ScenarioName: scenarioName,
		Divergent:    false,
		Summary:      "All engines satisfied invariants consistently",
		Results:      results,
	}
}
