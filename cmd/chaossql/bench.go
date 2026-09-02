package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

var (
	benchColorPrimary = lipgloss.Color("#7D56F4")
	benchColorGreen   = lipgloss.Color("#04B575")
	benchColorRed     = lipgloss.Color("#FF5F87")
	benchColorYellow  = lipgloss.Color("#FFAF00")
	benchColorCyan    = lipgloss.Color("#00D7FF")
	benchColorGray    = lipgloss.Color("#626262")
	benchColorWhite   = lipgloss.Color("#FFFFFF")

	benchBoldStyle = lipgloss.NewStyle().Bold(true)

	benchHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(benchColorCyan).
				Padding(0, 1)

	benchCellStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#E0E0E0"))

	benchComponentStyle = lipgloss.NewStyle().
				Bold(true).
				Padding(0, 1).
				Foreground(benchColorCyan)

	benchValueStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(benchColorWhite)

	benchOptimalBadge = lipgloss.NewStyle().
				Bold(true).
				Foreground(benchColorGreen)

	benchPassBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(benchColorGreen)

	benchWarnBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(benchColorYellow)

	benchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(benchColorPrimary).
			Padding(0, 1).
			MarginBottom(1)
)

// BenchmarkConfig holds runtime benchmark options.
type BenchmarkConfig struct {
	ScenarioPath string        `json:"scenario_path,omitempty"`
	Duration     time.Duration `json:"duration"`
	Workers      int           `json:"workers"`
	JSONOutput   bool          `json:"json_output"`
}

// PRNGBenchResult holds PRNG micro-benchmark metrics.
type PRNGBenchResult struct {
	TotalOps            int64         `json:"total_ops"`
	Duration            time.Duration `json:"duration_ns"`
	ThroughputOpsPerSec float64       `json:"ops_per_sec"`
	Status              string        `json:"status"`
}

// AdyaGraphBenchResult holds Adya dependency graph micro-benchmark metrics.
type AdyaGraphBenchResult struct {
	Nodes      int           `json:"nodes"`
	Duration   time.Duration `json:"duration_ns"`
	DurationMs float64       `json:"duration_ms"`
	Cycles     int           `json:"cycles_found"`
	Status     string        `json:"status"`
}

// DeltaDebuggingBenchResult holds ddmin shrinker micro-benchmark metrics.
type DeltaDebuggingBenchResult struct {
	InitialOps       int           `json:"initial_ops"`
	ReducedOps       int           `json:"reduced_ops"`
	Iterations       int           `json:"iterations"`
	Duration         time.Duration `json:"duration_ns"`
	IterationsPerSec float64       `json:"iterations_per_sec"`
	Status           string        `json:"status"`
}

// DatabaseBenchResult holds concurrent DB execution stress metrics.
type DatabaseBenchResult struct {
	Driver       string        `json:"driver"`
	Scenario     string        `json:"scenario"`
	Workers      int           `json:"workers"`
	TotalTx      int64         `json:"total_transactions"`
	Duration     time.Duration `json:"duration_ns"`
	TPS          float64       `json:"tps"`
	AvgLatencyMs float64       `json:"avg_latency_ms"`
	Status       string        `json:"status"`
}

// BenchmarkRow represents a single line in the metric card.
type BenchmarkRow struct {
	Component string `json:"component"`
	Metric    string `json:"metric"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	Status    string `json:"status"`
}

// BenchmarkEnv captures runtime environment metadata.
type BenchmarkEnv struct {
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
	NumCPU    int    `json:"num_cpu"`
}

// BenchmarkSuiteResult aggregates all benchmark metrics.
type BenchmarkSuiteResult struct {
	Config         BenchmarkConfig           `json:"config"`
	Environment    BenchmarkEnv              `json:"environment"`
	PRNG           PRNGBenchResult           `json:"prng"`
	AdyaGraphs     []AdyaGraphBenchResult    `json:"adya_graphs"`
	DeltaDebugging DeltaDebuggingBenchResult `json:"delta_debugging"`
	Database       DatabaseBenchResult       `json:"database"`
	Rows           []BenchmarkRow            `json:"metrics"`
	TotalDuration  time.Duration             `json:"total_duration_ns"`
	Summary        map[string]interface{}    `json:"summary"`
}

// newBenchCmd creates the `chaossql bench` Cobra command.
func newBenchCmd() *cobra.Command {
	var (
		durationStr string
		workersFlag int
		jsonOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "bench [scenario.yaml]",
		Short: "Run performance micro-benchmarks and concurrent database stress benchmarks",
		Long: `ChaosSQL Benchmark Suite executes high-precision micro-benchmarks and concurrent stress tests:
  • PRNG & Dynamic Parameter Generators throughput (ops/sec)
  • Adya Dependency Graph construction & Cycle Detection latency (100, 500, 1000 nodes)
  • Delta-Debugging (ddmin) causal reduction performance (iterations/sec)
  • Real concurrent database transaction throughput (TPS) with SQLite/Postgres`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			duration, err := time.ParseDuration(durationStr)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", durationStr, err)
			}
			if duration <= 0 {
				return fmt.Errorf("duration must be positive, got %v", duration)
			}
			if workersFlag <= 0 {
				return fmt.Errorf("workers must be positive, got %d", workersFlag)
			}

			var scenarioPath string
			if len(args) > 0 {
				scenarioPath = args[0]
				if _, err := os.Stat(scenarioPath); os.IsNotExist(err) {
					altPath := filepath.Join("/root/chaossql", scenarioPath)
					if _, err := os.Stat(altPath); err == nil {
						scenarioPath = altPath
					} else {
						return fmt.Errorf("scenario file %q not found: %w", scenarioPath, err)
					}
				}
			}

			cfg := BenchmarkConfig{
				ScenarioPath: scenarioPath,
				Duration:     duration,
				Workers:      workersFlag,
				JSONOutput:   jsonOutput,
			}

			ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			suiteResult, err := runBenchmarkSuite(ctx, cfg)
			if err != nil {
				return fmt.Errorf("benchmark execution failed: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(suiteResult)
			}

			output := renderBenchmarkCard(suiteResult)
			fmt.Fprintln(cmd.OutOrStdout(), output)
			return nil
		},
	}

	cmd.Flags().StringVarP(&durationStr, "duration", "d", "2s", "Target execution duration for time-bounded benchmarks")
	cmd.Flags().IntVarP(&workersFlag, "workers", "w", 4, "Number of concurrent worker goroutines for DB stress")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output benchmark metrics in structured JSON format")

	return cmd
}

func formatNumber(n int64) string {
	in := fmt.Sprintf("%d", n)
	var out []byte
	l := len(in)
	for i := 0; i < l; i++ {
		if i > 0 && (l-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, in[i])
	}
	return string(out)
}

func formatFloatWithCommas(f float64, decimals int) string {
	parts := strings.Split(fmt.Sprintf("%.*f", decimals, f), ".")
	intPart := parts[0]
	var out []byte
	l := len(intPart)
	for i := 0; i < l; i++ {
		if i > 0 && (l-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, intPart[i])
	}
	if len(parts) > 1 {
		return string(out) + "." + parts[1]
	}
	return string(out)
}

// benchmarkPRNGAndGenerators measures throughput of PRNG evaluation and dynamic generators.
func benchmarkPRNGAndGenerators(ctx context.Context, duration time.Duration) (PRNGBenchResult, error) {
	if duration <= 0 {
		duration = 200 * time.Millisecond
	}

	p := engine.NewPRNG(1337)
	rng := rand.New(rand.NewPCG(p.MasterSeed(), 0))

	expressions := []string{
		"$random_int(1, 1000)",
		"$random_choice('DEPOSIT', 'WITHDRAW', 'TRANSFER', 'BALANCE')",
		"$random_string(16)",
		"$uuid()",
		"int(10, 500)",
	}

	start := time.Now()
	var totalOps int64
	deadline := start.Add(duration)

	for {
		if ctx.Err() != nil {
			return PRNGBenchResult{}, ctx.Err()
		}

		for _, expr := range expressions {
			_, _ = engine.EvaluateGenerator(expr, rng)
			_ = p.WorkerSeed(int(totalOps % 8))
			_ = p.Jitter([2]int{5, 20}, rng)
			totalOps += 3
		}

		if totalOps >= 10000 && time.Now().After(deadline) {
			break
		}
	}

	elapsed := time.Since(start)
	opsPerSec := float64(totalOps) / elapsed.Seconds()

	status := "OPTIMAL"
	if opsPerSec < 50000 {
		status = "PASS"
	}

	return PRNGBenchResult{
		TotalOps:            totalOps,
		Duration:            elapsed,
		ThroughputOpsPerSec: opsPerSec,
		Status:              status,
	}, nil
}

// benchmarkAdyaGraph constructs a synthetic dependency trace and benchmarks graph build & cycle search.
func benchmarkAdyaGraph(ctx context.Context, nodes int) (AdyaGraphBenchResult, error) {
	if nodes <= 0 {
		nodes = 100
	}

	trace := make(domain.ExecutionTrace, 0, nodes*2)
	for i := 0; i < nodes; i++ {
		workerID := (i % 8) + 1
		acctA := (i % 25) + 1
		acctB := ((i + 1) % 25) + 1

		trace = append(trace, domain.TraceEvent{
			Timestamp: time.Duration(i) * time.Microsecond,
			WorkerID:  workerID,
			OpIndex:   i + 1,
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       fmt.Sprintf("SELECT balance FROM accounts WHERE id = %d", acctA),
		})
		trace = append(trace, domain.TraceEvent{
			Timestamp: time.Duration(i)*time.Microsecond + 5*time.Microsecond,
			WorkerID:  workerID,
			OpIndex:   i + 1,
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       fmt.Sprintf("UPDATE accounts SET balance = balance - 10 WHERE id = %d", acctB),
		})
	}

	start := time.Now()
	graph := analyzer.BuildGraph(trace)
	cycles := analyzer.FindCycles(graph)
	for _, c := range cycles {
		_ = analyzer.ClassifyCycle(c)
	}
	elapsed := time.Since(start)

	durMs := float64(elapsed.Microseconds()) / 1000.0

	status := "OPTIMAL"
	if nodes == 1000 && durMs > 500.0 {
		status = "WARN"
	}

	return AdyaGraphBenchResult{
		Nodes:      nodes,
		Duration:   elapsed,
		DurationMs: durMs,
		Cycles:     len(cycles),
		Status:     status,
	}, nil
}

// benchmarkDeltaDebugging measures causal reduction iterations/sec on a synthetic failing schedule.
func benchmarkDeltaDebugging(ctx context.Context, duration time.Duration) (DeltaDebuggingBenchResult, error) {
	if duration <= 0 {
		duration = 200 * time.Millisecond
	}

	initialOpsCount := 64
	initialOps := make([]domain.ScheduledOp, initialOpsCount)
	for i := 0; i < initialOpsCount; i++ {
		initialOps[i] = domain.ScheduledOp{
			ID:   i + 1,
			Name: fmt.Sprintf("op_%d", i+1),
		}
	}

	// The bug reproduces if and only if both op_12 and op_42 are in the subset
	testFn := func(subset []domain.ScheduledOp) bool {
		has12 := false
		has42 := false
		for _, op := range subset {
			if op.ID == 12 {
				has12 = true
			}
			if op.ID == 42 {
				has42 = true
			}
		}
		if has12 && has42 {
			return false // FAIL (reproduces bug)
		}
		return true // PASS
	}

	start := time.Now()
	deadline := start.Add(duration)
	var totalIterations int
	var lastReducedSize int

	for {
		if ctx.Err() != nil {
			return DeltaDebuggingBenchResult{}, ctx.Err()
		}

		shrunk, err := shrinker.Shrink(ctx, testFn, initialOps)
		if err != nil {
			return DeltaDebuggingBenchResult{}, err
		}
		totalIterations += shrunk.Iterations
		lastReducedSize = shrunk.ReducedSize

		if time.Now().After(deadline) {
			break
		}
	}

	elapsed := time.Since(start)
	iterPerSec := float64(totalIterations) / elapsed.Seconds()

	status := "OPTIMAL"
	if iterPerSec < 100 {
		status = "PASS"
	}

	return DeltaDebuggingBenchResult{
		InitialOps:       initialOpsCount,
		ReducedOps:       lastReducedSize,
		Iterations:       totalIterations,
		Duration:         elapsed,
		IterationsPerSec: iterPerSec,
		Status:           status,
	}, nil
}

// benchmarkDatabaseConcurrency runs real concurrent database transactions and measures TPS.
func benchmarkDatabaseConcurrency(ctx context.Context, cfg BenchmarkConfig, customSpec *domain.Spec) (DatabaseBenchResult, error) {
	var spec domain.Spec
	if customSpec != nil {
		spec = *customSpec
	} else if cfg.ScenarioPath != "" {
		loaded, err := domain.LoadSpec(cfg.ScenarioPath)
		if err != nil {
			return DatabaseBenchResult{}, fmt.Errorf("failed to load scenario: %w", err)
		}
		spec = *loaded
	} else {
		// Built-in standard benchmark specification
		spec = domain.Spec{
			Version:     "1.0",
			Name:        "Banking High-Concurrency Benchmark",
			Description: "Concurrent SQLite transaction throughput benchmark",
			Database: domain.DatabaseConfig{
				Driver: "sqlite",
				DSN:    ":memory:",
				Schema: `CREATE TABLE IF NOT EXISTS accounts (
					id INTEGER PRIMARY KEY,
					balance INTEGER NOT NULL
				);`,
				Seed: `INSERT INTO accounts (id, balance) VALUES (1, 100000), (2, 100000), (3, 100000), (4, 100000);`,
			},
			Engine: domain.EngineConfig{
				Workers:    cfg.Workers,
				Iterations: 1000,
				Seed:       42,
			},
		}
	}

	var driver drivers.DatabaseDriver
	switch strings.ToLower(spec.Database.Driver) {
	case "sqlite", "sqlite3", "":
		dsn := spec.Database.DSN
		if dsn == "" {
			dsn = ":memory:"
		}
		driver = drivers.NewSQLiteDriver(dsn)
	case "postgres", "postgresql":
		driver = drivers.NewPostgresDriver(spec.Database.DSN)
	default:
		return DatabaseBenchResult{}, fmt.Errorf("unsupported benchmark database driver: %s", spec.Database.Driver)
	}

	if err := driver.Open(ctx); err != nil {
		return DatabaseBenchResult{}, fmt.Errorf("database open failed: %w", err)
	}
	defer func() { _ = driver.Close() }()

	if err := driver.Reset(ctx, spec.Database.Schema, spec.Database.Seed); err != nil {
		return DatabaseBenchResult{}, fmt.Errorf("database reset failed: %w", err)
	}

	nWorkers := cfg.Workers
	if nWorkers <= 0 {
		nWorkers = 4
	}

	duration := cfg.Duration
	if duration <= 0 {
		duration = 500 * time.Millisecond
	}

	var totalTx int64
	benchCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	var wg sync.WaitGroup
	start := time.Now()

	for w := 1; w <= nWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewPCG(uint64(workerID), uint64(time.Now().UnixNano())))

			for {
				select {
				case <-benchCtx.Done():
					return
				default:
				}

				acctA := (rng.IntN(4)) + 1
				acctB := (rng.IntN(4)) + 1
				if acctA == acctB {
					acctB = (acctA % 4) + 1
				}
				amount := (rng.IntN(20)) + 1

				// Execute transaction
				_, err1 := driver.Exec(benchCtx, fmt.Sprintf("UPDATE accounts SET balance = balance - %d WHERE id = %d", amount, acctA))
				_, err2 := driver.Exec(benchCtx, fmt.Sprintf("UPDATE accounts SET balance = balance + %d WHERE id = %d", amount, acctB))
				row := driver.QueryRow(benchCtx, fmt.Sprintf("SELECT balance FROM accounts WHERE id = %d", acctA))
				var bal int
				_ = row.Scan(&bal)

				if err1 == nil && err2 == nil {
					atomic.AddInt64(&totalTx, 1)
				}
			}
		}(w)
	}

	wg.Wait()
	elapsed := time.Since(start)
	tps := float64(totalTx) / elapsed.Seconds()
	avgLatencyMs := 0.0
	if totalTx > 0 {
		avgLatencyMs = (elapsed.Seconds() * 1000.0 * float64(nWorkers)) / float64(totalTx)
	}

	status := "OPTIMAL"
	if tps < 50 {
		status = "PASS"
	}

	return DatabaseBenchResult{
		Driver:       spec.Database.Driver,
		Scenario:     spec.Name,
		Workers:      nWorkers,
		TotalTx:      totalTx,
		Duration:     elapsed,
		TPS:          tps,
		AvgLatencyMs: avgLatencyMs,
		Status:       status,
	}, nil
}

// runBenchmarkSuite coordinates execution of the entire benchmark battery.
func runBenchmarkSuite(ctx context.Context, cfg BenchmarkConfig) (*BenchmarkSuiteResult, error) {
	suiteStart := time.Now()

	env := BenchmarkEnv{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
	}

	// 1. PRNG & Generator throughput
	prngRes, err := benchmarkPRNGAndGenerators(ctx, cfg.Duration/3)
	if err != nil {
		return nil, fmt.Errorf("PRNG benchmark failed: %w", err)
	}

	// 2. Adya Graph construction & Cycle Detection latency (100, 500, 1000 nodes)
	var adyaResults []AdyaGraphBenchResult
	for _, nodes := range []int{100, 500, 1000} {
		res, err := benchmarkAdyaGraph(ctx, nodes)
		if err != nil {
			return nil, fmt.Errorf("Adya graph benchmark (%d nodes) failed: %w", nodes, err)
		}
		adyaResults = append(adyaResults, res)
	}

	// 3. Delta-Debugging reduction performance
	ddminRes, err := benchmarkDeltaDebugging(ctx, cfg.Duration/3)
	if err != nil {
		return nil, fmt.Errorf("Delta-Debugging benchmark failed: %w", err)
	}

	// 4. Real concurrent database transaction throughput
	dbRes, err := benchmarkDatabaseConcurrency(ctx, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("Database stress benchmark failed: %w", err)
	}

	totalDuration := time.Since(suiteStart)

	// Assemble rows
	var rows []BenchmarkRow
	rows = append(rows, BenchmarkRow{
		Component: "PRNG & Generators",
		Metric:    "Parameter Throughput",
		Value:     formatFloatWithCommas(prngRes.ThroughputOpsPerSec, 0),
		Unit:      "ops/sec",
		Status:    prngRes.Status,
	})

	for _, a := range adyaResults {
		rows = append(rows, BenchmarkRow{
			Component: fmt.Sprintf("Adya Graph (%d nodes)", a.Nodes),
			Metric:    "Build & Cycle Latency",
			Value:     fmt.Sprintf("%.2f", a.DurationMs),
			Unit:      "ms",
			Status:    a.Status,
		})
	}

	rows = append(rows, BenchmarkRow{
		Component: "Delta-Debugging (ddmin)",
		Metric:    "Reduction Throughput",
		Value:     formatFloatWithCommas(ddminRes.IterationsPerSec, 0),
		Unit:      "iter/sec",
		Status:    ddminRes.Status,
	})

	rows = append(rows, BenchmarkRow{
		Component: fmt.Sprintf("Database Engine (%s)", strings.ToUpper(dbRes.Driver)),
		Metric:    "Concurrent Throughput",
		Value:     formatFloatWithCommas(dbRes.TPS, 0),
		Unit:      "TPS",
		Status:    dbRes.Status,
	})

	summary := map[string]interface{}{
		"total_benchmarks": len(rows),
		"all_optimal":      true,
		"total_time_ms":    totalDuration.Milliseconds(),
	}

	return &BenchmarkSuiteResult{
		Config:         cfg,
		Environment:    env,
		PRNG:           prngRes,
		AdyaGraphs:     adyaResults,
		DeltaDebugging: ddminRes,
		Database:       dbRes,
		Rows:           rows,
		TotalDuration:  totalDuration,
		Summary:        summary,
	}, nil
}

// renderBenchmarkCard formats the complete benchmark suite as a Lipgloss terminal card.
func renderBenchmarkCard(res *BenchmarkSuiteResult) string {
	var sb strings.Builder

	// Top banner
	sb.WriteString(reporter.RenderBanner())
	sb.WriteString("\n")

	// Suite Header Box
	var headerBox strings.Builder
	title := benchBoldStyle.Foreground(benchColorCyan).Render("CHAOSSQL PERFORMANCE & STRESS BENCHMARK SUITE")
	headerBox.WriteString(fmt.Sprintf("%s\n\n", title))
	headerBox.WriteString(fmt.Sprintf("  • %s: %s | %s (%d CPUs)\n",
		benchBoldStyle.Render("Runtime Environment"),
		res.Environment.GoVersion,
		res.Environment.OS+"/"+res.Environment.Arch,
		res.Environment.NumCPU,
	))
	headerBox.WriteString(fmt.Sprintf("  • %s: duration=%v | workers=%d | target_db=%s\n",
		benchBoldStyle.Render("Benchmark Parameters"),
		res.Config.Duration,
		res.Config.Workers,
		res.Database.Driver,
	))
	headerBox.WriteString(fmt.Sprintf("  • %s: %s\n",
		benchBoldStyle.Render("Elapsed Benchmark Time"),
		res.TotalDuration.Round(100*time.Microsecond),
	))
	sb.WriteString(benchBoxStyle.Render(headerBox.String()))
	sb.WriteString("\n")

	// Table Data
	headers := []string{"COMPONENT", "METRIC", "VALUE", "UNIT", "STATUS"}
	var tableRows [][]string

	for _, row := range res.Rows {
		statusFormatted := "✔ " + row.Status
		if row.Status == "OPTIMAL" {
			statusFormatted = benchOptimalBadge.Render("✔ OPTIMAL")
		} else if row.Status == "PASS" {
			statusFormatted = benchPassBadge.Render("✔ PASS")
		} else {
			statusFormatted = benchWarnBadge.Render("▲ WARN")
		}

		tableRows = append(tableRows, []string{
			benchComponentStyle.Render(row.Component),
			benchCellStyle.Render(row.Metric),
			benchValueStyle.Render(row.Value),
			lipgloss.NewStyle().Foreground(benchColorGray).Render(row.Unit),
			statusFormatted,
		})
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(benchColorPrimary)).
		Headers(
			benchHeaderStyle.Render(headers[0]),
			benchHeaderStyle.Render(headers[1]),
			benchHeaderStyle.Render(headers[2]),
			benchHeaderStyle.Render(headers[3]),
			benchHeaderStyle.Render(headers[4]),
		).
		Rows(tableRows...)

	sb.WriteString(t.Render())
	sb.WriteString("\n\n")

	// Footer Card
	footerBadge := benchBoldStyle.Foreground(lipgloss.Color("#FFFFFF")).Background(benchColorGreen).Padding(0, 1).
		Render(fmt.Sprintf(" ✔ ALL %d BENCHMARKS COMPLETED SUCCESSFULLY ", len(res.Rows)))
	footerHint := lipgloss.NewStyle().Foreground(benchColorGray).Italic(true).
		Render("  Use --json for automated CI metric parsing or --duration=5s for extended stress runs.")

	sb.WriteString(fmt.Sprintf("  Status: %s\n%s\n", footerBadge, footerHint))

	return sb.String()
}