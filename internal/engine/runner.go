package engine

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/evaluator"
)

type RunResult = domain.ExecutionResult

// ScheduleOutcome contains the trace and structured failures from worker execution.
type ScheduleOutcome struct {
	Trace           domain.ExecutionTrace
	OperationErrors []domain.OperationError
	Canceled        bool
	Isolation       domain.IsolationLevel
	Schedule        domain.SchedulePlan
}

// Runner executes chaos schedules across database connections.
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

// GenerateSchedule creates a slice of deterministic ScheduledOps for the given spec and PRNG.
func GenerateSchedule(spec domain.Spec, prng *PRNG) []domain.ScheduledOp {
	numOps := spec.Engine.Iterations
	if numOps <= 0 {
		numOps = 10
	}
	if len(spec.Operations) == 0 {
		return nil
	}

	runPRNG := NewPRNG(prng.MasterSeed())
	masterRng := rand.New(rand.NewPCG(runPRNG.MasterSeed(), 0))
	scheduledOps := make([]domain.ScheduledOp, numOps)

	for i := 0; i < numOps; i++ {
		opTemplate := spec.Operations[masterRng.IntN(len(spec.Operations))]
		params := make(map[string]string)
		paramNames := make([]string, 0, len(opTemplate.Params))
		for name := range opTemplate.Params {
			paramNames = append(paramNames, name)
		}
		sort.Strings(paramNames)
		for _, name := range paramNames {
			params[name] = runPRNG.EvaluateParam(opTemplate.Params[name], masterRng)
		}
		scheduledOps[i] = domain.ScheduledOp{
			ID:     i + 1,
			Name:   opTemplate.Name,
			Params: params,
			Steps:  opTemplate.Steps,
		}
	}
	return scheduledOps
}

func (r *Runner) Run(ctx context.Context, spec domain.Spec) (*RunResult, error) {
	startTime := time.Now()

	if _, err := r.driver.EffectiveIsolation(spec.Database.Isolation); err != nil {
		return nil, err
	}

	// 1. Reset database
	if err := r.driver.Reset(ctx, spec.Database.Schema, spec.Database.Seed); err != nil {
		return nil, fmt.Errorf("database reset failed: %w", err)
	}

	// 2. Generate Scheduled Operations
	scheduledOps := GenerateSchedule(spec, r.prng)

	// 3. Execute Scheduled Operations with Workers
	outcome, err := r.ExecuteSchedule(ctx, spec, scheduledOps)
	if err != nil {
		if outcome.Canceled {
			return r.finalizeResult(ctx, spec, scheduledOps, outcome, startTime), err
		}
		return nil, err
	}

	return r.finalizeResult(ctx, spec, scheduledOps, outcome, startTime), nil
}

// ExecuteSchedule dispatches the given operations across workers and captures the trace.
func (r *Runner) ExecuteSchedule(ctx context.Context, spec domain.Spec, ops []domain.ScheduledOp) (ScheduleOutcome, error) {
	nWorkers := spec.Engine.Workers
	if nWorkers <= 0 {
		nWorkers = 4
	}

	plan := BuildSchedulePlan(spec, ops, r.prng)
	decisions := make(map[scheduleDecisionKey]domain.ScheduleDecision, len(plan.Decisions))
	for _, decision := range plan.Decisions {
		decisions[scheduleDecisionKey{operationID: decision.OperationID, stepIndex: decision.StepIndex}] = decision
	}

	effectiveIsolation, err := r.driver.EffectiveIsolation(spec.Database.Isolation)
	if err != nil {
		return ScheduleOutcome{}, err
	}

	var trace domain.ExecutionTrace
	var traceMu sync.Mutex
	var operationErrors []domain.OperationError
	var errorsMu sync.Mutex
	startTime := time.Now()

	addEvent := func(workerID, opIdx, stepIdx int, opName string, evType domain.TraceEventType, phase, sql string, err error) {
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
			Phase:     phase,
			SQL:       sql,
			Error:     errStr,
		})
	}
	addErrors := func(errs []domain.OperationError) {
		if len(errs) == 0 {
			return
		}
		errorsMu.Lock()
		operationErrors = append(operationErrors, errs...)
		errorsMu.Unlock()
	}

	orderedOps := append([]domain.ScheduledOp(nil), ops...)
	sort.SliceStable(orderedOps, func(i, j int) bool {
		return orderedOps[i].ID < orderedOps[j].ID
	})
	workerQueues := make([][]domain.ScheduledOp, nWorkers)
	for _, op := range orderedOps {
		workerID := (op.ID-1)%nWorkers + 1
		workerQueues[workerID-1] = append(workerQueues[workerID-1], op)
	}

	var wg sync.WaitGroup
	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go func(workerID int, queue []domain.ScheduledOp) {
			defer wg.Done()

			for _, op := range queue {
				if ctx.Err() != nil {
					break
				}
				errs := r.executeOperation(ctx, spec, op, workerID, decisions, addEvent)
				addErrors(errs)
			}
		}(w, workerQueues[w-1])
	}

	wg.Wait()
	sort.SliceStable(operationErrors, func(i, j int) bool {
		if operationErrors[i].OperationID != operationErrors[j].OperationID {
			return operationErrors[i].OperationID < operationErrors[j].OperationID
		}
		if operationErrors[i].StepIndex != operationErrors[j].StepIndex {
			return operationErrors[i].StepIndex < operationErrors[j].StepIndex
		}
		return operationErrors[i].Phase < operationErrors[j].Phase
	})
	outcome := ScheduleOutcome{
		Trace:           trace,
		OperationErrors: operationErrors,
		Canceled:        ctx.Err() != nil,
		Isolation:       effectiveIsolation,
		Schedule:        plan,
	}
	if ctx.Err() != nil {
		return outcome, ctx.Err()
	}
	return outcome, nil
}

type eventRecorder func(workerID, opIdx, stepIdx int, opName string, evType domain.TraceEventType, phase, sql string, err error)

type scheduleDecisionKey struct {
	operationID int
	stepIndex   int
}

func (r *Runner) executeOperation(
	ctx context.Context,
	spec domain.Spec,
	op domain.ScheduledOp,
	workerID int,
	decisions map[scheduleDecisionKey]domain.ScheduleDecision,
	addEvent eventRecorder,
) []domain.OperationError {
	tx, err := r.driver.BeginTx(ctx, drivers.TransactionOptions{Isolation: spec.Database.Isolation})
	if err != nil {
		addEvent(workerID, op.ID, 0, op.Name, domain.EventError, "begin", "BEGIN", err)
		return []domain.OperationError{newOperationError(op, 0, "begin", err)}
	}
	addEvent(workerID, op.ID, 0, op.Name, domain.EventBegin, "begin", "BEGIN", nil)

	rollback := func(stepIndex int, causePhase string, cause error) []domain.OperationError {
		rollbackErr := tx.Rollback()
		if ctx.Err() != nil && errors.Is(rollbackErr, sql.ErrTxDone) {
			rollbackErr = nil
		}
		addEvent(workerID, op.ID, stepIndex, op.Name, domain.EventRollback, "rollback", "ROLLBACK", rollbackErr)
		errs := make([]domain.OperationError, 0, 2)
		if cause != nil {
			errs = append(errs, newOperationError(op, stepIndex, causePhase, cause))
		}
		if rollbackErr != nil {
			errs = append(errs, newOperationError(op, stepIndex, "rollback", rollbackErr))
		}
		return errs
	}

	localState := make(map[string]string, len(op.Params))
	for key, value := range op.Params {
		localState[key] = value
	}

	for stepIdx, step := range op.Steps {
		stepNumber := stepIdx + 1
		decision := decisions[scheduleDecisionKey{operationID: op.ID, stepIndex: stepNumber}]
		if ctx.Err() != nil {
			return rollback(stepNumber, "", nil)
		}

		if !waitForContext(ctx, time.Duration(decision.JitterMs)*time.Millisecond) {
			return rollback(stepNumber, "", nil)
		}
		if !waitForContext(ctx, time.Duration(decision.LatencyMs)*time.Millisecond) {
			return rollback(stepNumber, "", nil)
		}
		if decision.Abort {
			return rollback(stepNumber, "", nil)
		}

		sqlStmt := SubstituteParams(step.SQL, localState)
		if step.Capture != "" {
			var capturedVal interface{}
			if scanErr := tx.QueryRowContext(ctx, sqlStmt).Scan(&capturedVal); scanErr != nil {
				addEvent(workerID, op.ID, stepNumber, op.Name, domain.EventError, "step", sqlStmt, scanErr)
				return rollback(stepNumber, "step", scanErr)
			}
			localState[step.Capture] = fmt.Sprintf("%v", capturedVal)
			addEvent(workerID, op.ID, stepNumber, op.Name, DetectEventType(sqlStmt), "step", sqlStmt, nil)
			continue
		}

		if _, execErr := tx.ExecContext(ctx, sqlStmt); execErr != nil {
			addEvent(workerID, op.ID, stepNumber, op.Name, domain.EventError, "step", sqlStmt, execErr)
			return rollback(stepNumber, "step", execErr)
		}
		addEvent(workerID, op.ID, stepNumber, op.Name, DetectEventType(sqlStmt), "step", sqlStmt, nil)
	}

	if commitErr := tx.Commit(); commitErr != nil {
		addEvent(workerID, op.ID, len(op.Steps)+1, op.Name, domain.EventError, "commit", "COMMIT", commitErr)
		return rollback(len(op.Steps)+1, "commit", commitErr)
	}
	addEvent(workerID, op.ID, len(op.Steps)+1, op.Name, domain.EventCommit, "commit", "COMMIT", nil)
	return nil
}

func waitForContext(ctx context.Context, duration time.Duration) bool {
	if duration <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func newOperationError(op domain.ScheduledOp, stepIndex int, phase string, err error) domain.OperationError {
	return domain.OperationError{
		OperationID: op.ID,
		Operation:   op.Name,
		StepIndex:   stepIndex,
		Phase:       phase,
		Message:     err.Error(),
	}
}

// SubstituteParams replaces {param} or {a - b} in the SQL string.
func SubstituteParams(sql string, state map[string]string) string {
	return substituteParams(sql, state)
}

func substituteParams(sql string, state map[string]string) string {
	result := sql
	for k, v := range state {
		placeholder := "{" + k + "}"
		result = strings.ReplaceAll(result, placeholder, v)
	}

	for strings.Contains(result, "{") && strings.Contains(result, "}") {
		start := strings.Index(result, "{")
		end := strings.Index(result, "}")
		if end <= start {
			break
		}
		expr := result[start+1 : end]

		for k, v := range state {
			expr = strings.ReplaceAll(expr, k, v)
		}

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

	if _, err := r.driver.EffectiveIsolation(spec.Database.Isolation); err != nil {
		return nil, err
	}

	if err := r.driver.Reset(ctx, spec.Database.Schema, spec.Database.Seed); err != nil {
		return nil, fmt.Errorf("database reset failed: %w", err)
	}

	outcome, err := r.ExecuteSchedule(ctx, spec, ops)
	if err != nil {
		if outcome.Canceled {
			return r.finalizeResult(ctx, spec, ops, outcome, startTime), err
		}
		return nil, err
	}

	return r.finalizeResult(ctx, spec, ops, outcome, startTime), nil
}

func (r *Runner) finalizeResult(
	ctx context.Context,
	spec domain.Spec,
	ops []domain.ScheduledOp,
	outcome ScheduleOutcome,
	started time.Time,
) *RunResult {
	result := &RunResult{
		Trace:           outcome.Trace,
		OperationErrors: outcome.OperationErrors,
		Isolation:       outcome.Isolation,
		Seed:            r.prng.MasterSeed(),
		Schedule:        outcome.Schedule,
		ScheduledOps:    ops,
		Duration:        time.Since(started),
	}

	setStatus := func(status domain.ExecutionStatus) {
		result.Status = status
		result.Success = status == domain.StatusPassed
		result.ViolationDetected = status == domain.StatusViolation
	}

	if outcome.Canceled || ctx.Err() != nil {
		setStatus(domain.StatusCanceled)
		result.Error = ctx.Err()
		return result
	}

	if len(outcome.OperationErrors) > 0 {
		setStatus(domain.StatusExecutionError)
		errs := make([]error, 0, len(outcome.OperationErrors))
		for _, opErr := range outcome.OperationErrors {
			errs = append(errs, fmt.Errorf("operation %d (%s), %s: %s", opErr.OperationID, opErr.Operation, opErr.Phase, opErr.Message))
		}
		result.Error = errors.Join(errs...)
		return result
	}

	for _, inv := range spec.Invariants {
		invResult, err := r.evaluator.Evaluate(ctx, r.driver, inv)
		if err != nil {
			setStatus(domain.StatusInconclusive)
			result.FailingInvariant = &invResult
			result.Error = err
			return result
		}
		if !invResult.Passed {
			setStatus(domain.StatusViolation)
			result.FailingInvariant = &invResult
			return result
		}
	}

	setStatus(domain.StatusPassed)
	return result
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
