package chaostest

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

// Tester provides a fluent Go testing SDK for writing ChaosSQL tests in standard *testing.T.
type Tester struct {
	t          testing.TB
	driver     drivers.DatabaseDriver
	schemaSQL  string
	seedSQL    string
	invariants []domain.InvariantConfig
	operations []domain.OperationConfig
	jitterMs   [2]int
	mu         sync.Mutex
}

// New initializes a fluent chaos tester bound to a testing.TB context.
func New(t testing.TB) *Tester {
	return &Tester{
		t:        t,
		jitterMs: [2]int{1, 5},
	}
}

// WithDriver sets the database driver implementation.
func (tr *Tester) WithDriver(driver drivers.DatabaseDriver) *Tester {
	tr.driver = driver
	return tr
}

// WithSchema sets the DDL schema SQL to initialize before execution.
func (tr *Tester) WithSchema(schemaSQL string) *Tester {
	tr.schemaSQL = schemaSQL
	return tr
}

// WithSeed sets the initial DML seed SQL to populate the database.
func (tr *Tester) WithSeed(seedSQL string) *Tester {
	tr.seedSQL = seedSQL
	return tr
}

// WithInvariant registers an invariant assertion against the database state.
func (tr *Tester) WithInvariant(name, query, assertExpr string) *Tester {
	tr.invariants = append(tr.invariants, domain.InvariantConfig{
		Name:   name,
		Query:  query,
		Assert: assertExpr,
	})
	return tr
}

// WithJitter sets the min and max millisecond jitter injected between operation steps.
func (tr *Tester) WithJitter(minMs, maxMs int) *Tester {
	tr.jitterMs = [2]int{minMs, maxMs}
	return tr
}

// AddOperation registers a transaction operation composed of one or more SQL steps.
// Steps can optionally specify capture variables using "-> var_name" or "=> var_name" syntax
// (e.g. "SELECT balance FROM accounts WHERE id = 1; -> current_bal").
func (tr *Tester) AddOperation(name string, steps ...string) *Tester {
	var stepConfigs []domain.StepConfig
	for _, s := range steps {
		s = strings.TrimSpace(s)
		var sqlStmt string
		var captureVar string

		if strings.Contains(s, "->") {
			parts := strings.SplitN(s, "->", 2)
			sqlStmt = strings.TrimSpace(parts[0])
			captureVar = strings.TrimSpace(parts[1])
		} else if strings.Contains(s, "=>") {
			parts := strings.SplitN(s, "=>", 2)
			sqlStmt = strings.TrimSpace(parts[0])
			captureVar = strings.TrimSpace(parts[1])
		} else {
			sqlStmt = s
		}

		stepConfigs = append(stepConfigs, domain.StepConfig{
			SQL:     sqlStmt,
			Capture: captureVar,
		})
	}

	tr.operations = append(tr.operations, domain.OperationConfig{
		Name:   name,
		Weight: 1.0,
		Steps:  stepConfigs,
	})
	return tr
}

// AddOperationWithParams registers an operation with dynamic parameter generators and steps.
func (tr *Tester) AddOperationWithParams(name string, params map[string]string, steps ...string) *Tester {
	var stepConfigs []domain.StepConfig
	for _, s := range steps {
		s = strings.TrimSpace(s)
		var sqlStmt string
		var captureVar string

		if strings.Contains(s, "->") {
			parts := strings.SplitN(s, "->", 2)
			sqlStmt = strings.TrimSpace(parts[0])
			captureVar = strings.TrimSpace(parts[1])
		} else if strings.Contains(s, "=>") {
			parts := strings.SplitN(s, "=>", 2)
			sqlStmt = strings.TrimSpace(parts[0])
			captureVar = strings.TrimSpace(parts[1])
		} else {
			sqlStmt = s
		}

		stepConfigs = append(stepConfigs, domain.StepConfig{
			SQL:     sqlStmt,
			Capture: captureVar,
		})
	}

	tr.operations = append(tr.operations, domain.OperationConfig{
		Name:   name,
		Weight: 1.0,
		Params: params,
		Steps:  stepConfigs,
	})
	return tr
}

// Run executes the configured chaos workload and returns the execution result,
// performing causal delta-debugging shrinking if an invariant violation is detected.
func (tr *Tester) Run(ctx context.Context, workers, iterations int, seed uint64) (*domain.ExecutionResult, *domain.ShrinkResult, error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	driver := tr.driver
	if driver == nil {
		dsn := fmt.Sprintf("file:chaostest_mem_%d?mode=memory&cache=shared", time.Now().UnixNano())
		driver = drivers.NewSQLiteDriver(dsn)
		tr.driver = driver
	}

	if err := driver.Open(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to open database driver: %w", err)
	}

	spec := domain.Spec{
		Version: "1.1",
		Name:    "chaostest",
		Database: domain.DatabaseConfig{
			Driver: driver.DriverName(),
			Schema: tr.schemaSQL,
			Seed:   tr.seedSQL,
		},
		Engine: domain.EngineConfig{
			Workers:    workers,
			Iterations: iterations,
			Seed:       seed,
			JitterMs:   tr.jitterMs,
		},
		Invariants: tr.invariants,
		Operations: tr.operations,
	}

	runner := engine.NewRunner(driver, seed)
	execRes, err := runner.Run(ctx, spec)
	if err != nil {
		return nil, nil, fmt.Errorf("chaos execution failed: %w", err)
	}

	if !execRes.ViolationDetected {
		return execRes, nil, nil
	}

	// Invariant violation detected: perform Causal Delta-Debugging (ddmin)
	testFn := func(subset []domain.ScheduledOp) bool {
		res, err := runner.RunSchedule(ctx, spec, subset)
		if err != nil {
			return true
		}
		return !res.ViolationDetected
	}

	shrinkRes, shrinkErr := shrinker.Shrink(ctx, testFn, execRes.ScheduledOps)
	if shrinkErr != nil || shrinkRes == nil {
		fallbackShrink := &domain.ShrinkResult{
			OriginalSize:   len(execRes.ScheduledOps),
			ReducedSize:    len(execRes.ScheduledOps),
			ReductionRatio: 0.0,
			MinimalOps:     execRes.ScheduledOps,
			Iterations:     0,
		}
		return execRes, fallbackShrink, nil
	}

	return execRes, shrinkRes, nil
}

// AssertNoAnomalies runs the chaos test and calls t.Fatalf if an invariant violation is detected,
// outputting the classified anomaly type, failing invariant, and minimal reproducing schedule.
func (tr *Tester) AssertNoAnomalies(ctx context.Context, workers, iterations int, seed uint64) {
	if tr.t != nil {
		tr.t.Helper()
	}

	execRes, shrinkRes, err := tr.Run(ctx, workers, iterations, seed)
	if err != nil {
		if tr.t != nil {
			tr.t.Fatalf("ChaosSQL execution failed with error: %v", err)
		}
		return
	}

	if !execRes.ViolationDetected {
		return
	}

	// Classify anomaly type using Adya dependency graph analysis
	trace := execRes.Trace
	if shrinkRes != nil && len(shrinkRes.MinimalOps) > 0 {
		spec := domain.Spec{
			Version: "1.1",
			Name:    "chaostest_min",
			Database: domain.DatabaseConfig{
				Driver: tr.driver.DriverName(),
				Schema: tr.schemaSQL,
				Seed:   tr.seedSQL,
			},
			Engine: domain.EngineConfig{
				Workers:    workers,
				Iterations: len(shrinkRes.MinimalOps),
				Seed:       seed,
				JitterMs:   tr.jitterMs,
			},
			Invariants: tr.invariants,
			Operations: tr.operations,
		}
		runner := engine.NewRunner(tr.driver, seed)
		minRes, minErr := runner.RunSchedule(ctx, spec, shrinkRes.MinimalOps)
		if minErr == nil && minRes != nil && len(minRes.Trace) > 0 {
			trace = minRes.Trace
		}
	}

	graph := analyzer.BuildGraph(trace)
	cycles := analyzer.FindCycles(graph)
	anomaly := domain.AnomalyUnknown

	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls == domain.AnomalyG1aDirtyRead || cls == domain.AnomalyG0DirtyWrite ||
			cls == domain.AnomalyG1cCircularInfo || cls == domain.AnomalyG2AntiDependency ||
			cls == domain.AnomalyA5AReadSkew || cls == domain.AnomalyWriteSkew ||
			cls == domain.AnomalyLostUpdate {
			anomaly = cls
			break
		}
	}
	if anomaly == domain.AnomalyUnknown && len(cycles) > 0 {
		anomaly = analyzer.ClassifyCycle(cycles[0])
	}

	invInfo := "Unknown invariant"
	if execRes.FailingInvariant != nil {
		invInfo = execRes.FailingInvariant.String()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n🚨 ChaosSQL Isolation Anomaly Detected: %s\n", anomaly))
	sb.WriteString(fmt.Sprintf("   Failing Invariant: %s\n", invInfo))

	if shrinkRes != nil {
		sb.WriteString(fmt.Sprintf("   Delta-Debugging Reduction: %d -> %d ops (%.1f%% reduction in %d iterations)\n",
			shrinkRes.OriginalSize, shrinkRes.ReducedSize, shrinkRes.ReductionRatio, shrinkRes.Iterations))
		sb.WriteString("   Minimal Reproducing Operations:\n")
		for _, op := range shrinkRes.MinimalOps {
			sb.WriteString(fmt.Sprintf("     - Op #%d [%s]\n", op.ID, op.Name))
			for _, step := range op.Steps {
				capStr := ""
				if step.Capture != "" {
					capStr = fmt.Sprintf(" (capture: %s)", step.Capture)
				}
				sb.WriteString(fmt.Sprintf("         SQL: %s%s\n", step.SQL, capStr))
			}
		}
	}

	mermaidSeq := reporter.GenerateMermaidSequence(trace)
	if mermaidSeq != "" {
		sb.WriteString("\n   Mermaid Sequence Diagram:\n")
		sb.WriteString("   ```mermaid\n")
		for _, line := range strings.Split(mermaidSeq, "\n") {
			sb.WriteString("   " + line + "\n")
		}
		sb.WriteString("   ```\n")
	}

	if tr.t != nil {
		tr.t.Fatalf("%s", sb.String())
	}
}
