package engine

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/evaluator"
)

// RunResult holds the complete outcome of a chaos execution.
type RunResult struct {
	Success           bool
	ViolationDetected bool
	FailingInvariant  *domain.InvariantResult
	Trace             domain.ExecutionTrace
	ScheduledOps      []domain.ScheduledOp
	Duration          time.Duration
	Error             error
}

// Runner orchestrates the chaos execution.
type Runner struct {
	driver    drivers.DatabaseDriver
	evaluator *evaluator.Evaluator
	prng      *PRNG
}

// NewRunner creates a new chaos runner.
func NewRunner(driver drivers.DatabaseDriver, seed uint64) *Runner {
	return &Runner{
		driver:    driver,
		evaluator: evaluator.NewEvaluator(),
		prng:      NewPRNG(seed),
	}
}

// Run executes a full chaos battery.
func (r *Runner) Run(ctx context.Context, spec domain.Spec) (*RunResult, error) {
	startTime := time.Now()

	// 1. Reset database
	if err := r.driver.Reset(ctx, spec.Database.Schema, spec.Database.Seed); err != nil {
		return nil, fmt.Errorf("database reset failed: %w", err)
	}

	// 2. Generate Scheduled Operations
	numOps := spec.Engine.Iterations
	if numOps <= 0 {
		numOps = 10
	}

	masterRng := rand.New(rand.NewPCG(r.prng.MasterSeed(), 0))
	scheduledOps := make([]domain.ScheduledOp, numOps)

	for i := 0; i < numOps; i++ {
		opTemplate := spec.Operations[masterRng.IntN(len(spec.Operations))]
		params := make(map[string]string)
		for k, v := range opTemplate.Params {
			val, err := EvaluateGenerator(v, masterRng)
			if err != nil {
				val = r.prng.EvaluateParam(v, masterRng)
			}
			params[k] = val
		}
		scheduledOps[i] = domain.ScheduledOp{
			ID:     i + 1,
			Name:   opTemplate.Name,
			Params: params,
			Steps:  opTemplate.Steps,
		}
	}

	// 3. Execute Scheduled Operations with Workers
	trace, err := r.ExecuteSchedule(ctx, spec, scheduledOps)
	if err != nil {
		return nil, err
	}

	// 4. Evaluate Invariants
	var failingInv *domain.InvariantResult
	violationFound := false

	for _, inv := range spec.Invariants {
		invRes, err := r.evaluator.Evaluate(ctx, r.driver, inv)
		if err != nil || !invRes.Passed {
			violationFound = true
			failingInv = &invRes
			break
		}
	}

	duration := time.Since(startTime)

	return &RunResult{
		Success:           !violationFound,
		ViolationDetected: violationFound,
		FailingInvariant:  failingInv,
		Trace:             trace,
		ScheduledOps:      scheduledOps,
		Duration:          duration,
	}, nil
}

// ExecuteSchedule dispatches the given operations across workers and captures the trace.
func (r *Runner) ExecuteSchedule(ctx context.Context, spec domain.Spec, ops []domain.ScheduledOp) (domain.ExecutionTrace, error) {
	nWorkers := spec.Engine.Workers
	if nWorkers <= 0 {
		nWorkers = 4
	}

	var trace domain.ExecutionTrace
	var traceMu sync.Mutex
	startTime := time.Now()

	addEvent := func(workerID, opIdx, stepIdx int, opName string, evType domain.TraceEventType, sql string, err error) {
		traceMu.Lock()
		defer traceMu.Unlock()
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		trace = append(trace, domain.TraceEvent{
			Timestamp: time.Since(startTime),
			WorkerID:  workerID,
			OpIndex:   opIdx,
			OpName:    opName,
			StepIndex: stepIdx,
			Type:      evType,
			SQL:       sql,
			Error:     errStr,
		})
	}

	opChan := make(chan domain.ScheduledOp, len(ops))
	for _, op := range ops {
		opChan <- op
	}
	close(opChan)

	var wg sync.WaitGroup
	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			workerRng := rand.New(rand.NewPCG(r.prng.WorkerSeed(workerID), uint64(workerID)))

			for op := range opChan {
				addEvent(workerID, op.ID, 0, op.Name, domain.EventBegin, "BEGIN", nil)

				localState := make(map[string]string)
				for k, v := range op.Params {
					localState[k] = v
				}

				opFailed := false
				for stepIdx, step := range op.Steps {
					// Inject jitter between steps
					jitter := r.prng.Jitter(spec.Engine.JitterMs, workerRng)
					if jitter > 0 {
						time.Sleep(jitter)
					}

					sqlStmt := SubstituteParams(step.SQL, localState)

					if step.Capture != "" {
						var capturedVal interface{}
						row := r.driver.QueryRow(ctx, sqlStmt)
						if scanErr := row.Scan(&capturedVal); scanErr != nil {
							addEvent(workerID, op.ID, stepIdx+1, op.Name, domain.EventError, sqlStmt, scanErr)
							opFailed = true
							break
						}
						localState[step.Capture] = fmt.Sprintf("%v", capturedVal)
						addEvent(workerID, op.ID, stepIdx+1, op.Name, DetectEventType(sqlStmt), sqlStmt, nil)
					} else {
						_, execErr := r.driver.Exec(ctx, sqlStmt)
						if execErr != nil {
							addEvent(workerID, op.ID, stepIdx+1, op.Name, domain.EventError, sqlStmt, execErr)
							opFailed = true
							break
						}
						addEvent(workerID, op.ID, stepIdx+1, op.Name, DetectEventType(sqlStmt), sqlStmt, nil)
					}
				}

				if opFailed {
					addEvent(workerID, op.ID, len(op.Steps)+1, op.Name, domain.EventRollback, "ROLLBACK", nil)
				} else {
					addEvent(workerID, op.ID, len(op.Steps)+1, op.Name, domain.EventCommit, "COMMIT", nil)
				}
			}
		}(w)
	}

	wg.Wait()
	return trace, nil
}

// SubstituteParams replaces {param} or {a - b} in the SQL string.
func SubstituteParams(sql string, state map[string]string) string {
	return substituteParams(sql, state)
}

// substituteParams replaces {param} or {a - b} in the SQL string.
func substituteParams(sql string, state map[string]string) string {
	result := sql
	// First substitute direct keys
	for k, v := range state {
		placeholder := "{" + k + "}"
		result = strings.ReplaceAll(result, placeholder, v)
	}

	// Support simple arithmetic substitutions like {current_bal - amount}
	for strings.Contains(result, "{") && strings.Contains(result, "}") {
		start := strings.Index(result, "{")
		end := strings.Index(result, "}")
		if end <= start {
			break
		}
		expr := result[start+1 : end]

		// Substitute keys inside expr
		for k, v := range state {
			expr = strings.ReplaceAll(expr, k, v)
		}

		// Evaluate simple arithmetic like "1000 - 100"
		valStr := evalSimpleArithmetic(expr)
		result = result[0:start] + valStr + result[end+1:]
	}
	return result
}

func evalSimpleArithmetic(expr string) string {
	expr = strings.TrimSpace(expr)
	if strings.Contains(expr, "-") {
		parts := strings.Split(expr, "-")
		if len(parts) == 2 {
			a, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			b, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if e1 == nil && e2 == nil {
				return strconv.Itoa(a - b)
			}
		}
	}
	if strings.Contains(expr, "+") {
		parts := strings.Split(expr, "+")
		if len(parts) == 2 {
			a, e1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			b, e2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if e1 == nil && e2 == nil {
				return strconv.Itoa(a + b)
			}
		}
	}
	return expr
}

// RunSchedule executes a specific schedule, bypassing random generation.
func (r *Runner) RunSchedule(ctx context.Context, spec domain.Spec, ops []domain.ScheduledOp) (*RunResult, error) {
	startTime := time.Now()

	if err := r.driver.Reset(ctx, spec.Database.Schema, spec.Database.Seed); err != nil {
		return nil, fmt.Errorf("database reset failed: %w", err)
	}

	trace, err := r.ExecuteSchedule(ctx, spec, ops)
	if err != nil {
		return nil, err
	}

	var failingInv *domain.InvariantResult
	violationFound := false

	for _, inv := range spec.Invariants {
		invRes, err := r.evaluator.Evaluate(ctx, r.driver, inv)
		if err != nil || !invRes.Passed {
			violationFound = true
			failingInv = &invRes
			break
		}
	}

	return &RunResult{
		Success:           !violationFound,
		ViolationDetected: violationFound,
		FailingInvariant:  failingInv,
		Trace:             trace,
		ScheduledOps:      ops,
		Duration:          time.Since(startTime),
	}, nil
}

// DetectEventType inspects a SQL statement and returns the corresponding TraceEventType.
func DetectEventType(sql string) domain.TraceEventType {
	upper := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(upper, "SAVEPOINT") {
		return domain.EventSavepoint
	}
	if strings.HasPrefix(upper, "ROLLBACK TO SAVEPOINT") || strings.HasPrefix(upper, "ROLLBACK TO") {
		return domain.EventRollbackTo
	}
	if strings.HasPrefix(upper, "RELEASE SAVEPOINT") || strings.HasPrefix(upper, "RELEASE ") || upper == "RELEASE" {
		return domain.EventReleaseSavepoint
	}
	return domain.EventExec
}
