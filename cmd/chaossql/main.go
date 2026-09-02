package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
)

var (
	seedFlag          uint64
	workersFlag       int
	iterationsFlag    int
	jsonFlag          bool
	exportReproFlag   bool
	exportMermaidFlag bool
	exportHTMLFlag    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "chaossql",
		Short: "ChaosSQL: Causal concurrency chaos engineering and isolation anomaly detector",
		Long: `ChaosSQL is a causal chaos engineering framework for relational databases.
It executes PCT-scheduled concurrent transaction interleavings to detect isolation anomalies
(such as Lost Updates, Write Skew, and Phantom Reads) and synthesizes minimal reproductions.`,
	}

	runCmd := &cobra.Command{
		Use:   "run <spec.yaml>",
		Short: "Execute a chaos test battery from a YAML specification",
		Args:  cobra.ExactArgs(1),
		RunE:  runChaosTest,
	}

	runCmd.Flags().Uint64Var(&seedFlag, "seed", 0, "PRNG seed for deterministic replay (default from spec)")
	runCmd.Flags().IntVar(&workersFlag, "workers", 0, "Number of concurrent worker goroutines (default from spec)")
	runCmd.Flags().IntVar(&iterationsFlag, "iterations", 0, "Number of operations to schedule (default from spec)")
	runCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results in structured JSON format")
	runCmd.Flags().BoolVar(&exportReproFlag, "export-repro", false, "Export standalone Go reproduction test (repro_test.go)")
	runCmd.Flags().BoolVar(&exportMermaidFlag, "export-mermaid", false, "Export Mermaid sequence diagram (trace.mermaid)")
	runCmd.Flags().StringVar(&exportHTMLFlag, "export-html", "", "Export standalone HTML report to file (e.g. report.html)")

	demoCmd := &cobra.Command{
		Use:   "demo [banking|inventory|hospital|financial]",
		Short: "Run an end-to-end interactive demo of a known concurrency anomaly",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDemo,
	}

	demoCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output demo results in structured JSON format")
	demoCmd.Flags().BoolVar(&exportReproFlag, "export-repro", false, "Export standalone Go reproduction test")
	demoCmd.Flags().BoolVar(&exportMermaidFlag, "export-mermaid", false, "Export Mermaid sequence diagram")
	demoCmd.Flags().StringVar(&exportHTMLFlag, "export-html", "", "Export standalone HTML report to file (e.g. report.html)")

	rootCmd.AddCommand(runCmd, demoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runChaosTest(cmd *cobra.Command, args []string) error {
	specPath := args[0]
	return executeChaos(specPath)
}

func runDemo(cmd *cobra.Command, args []string) error {
	scenario := "banking"
	if len(args) > 0 {
		scenario = strings.ToLower(strings.TrimSpace(args[0]))
	}

	var specPath string
	switch scenario {
	case "banking", "banking_lost_update", "lost_update":
		specPath = "examples/banking_lost_update/chaos.yaml"
	case "inventory", "inventory_oversell", "oversell":
		specPath = "examples/inventory_oversell/chaos.yaml"
	case "hospital", "hospital_write_skew", "write_skew":
		specPath = "examples/hospital_write_skew/chaos.yaml"
	case "financial", "financial_audit", "read_skew", "read_skew_financial_audit":
		specPath = "examples/read_skew_financial_audit/chaos.yaml"
	default:
		return fmt.Errorf("unknown demo scenario %q. Available scenarios: banking, inventory, hospital, financial", scenario)
	}

	// If running from subfolder or root, locate the spec file
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		// Try root relative
		altPath := filepath.Join("/root/chaossql", specPath)
		if _, err := os.Stat(altPath); err == nil {
			specPath = altPath
		} else {
			return fmt.Errorf("demo spec file %q not found on disk: %w", specPath, err)
		}
	}

	return executeChaos(specPath)
}

func executeChaos(specPath string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	spec, err := domain.LoadSpec(specPath)
	if err != nil {
		return fmt.Errorf("failed to load chaos spec: %w", err)
	}

	// Apply CLI flag overrides
	if seedFlag > 0 {
		spec.Engine.Seed = seedFlag
	}
	if workersFlag > 0 {
		spec.Engine.Workers = workersFlag
	}
	if iterationsFlag > 0 {
		spec.Engine.Iterations = iterationsFlag
	}

	// Initialize database driver
	var driver drivers.DatabaseDriver
	switch strings.ToLower(spec.Database.Driver) {
	case "sqlite", "sqlite3":
		driver = drivers.NewSQLiteDriver(spec.Database.DSN)
	case "postgres", "postgresql":
		driver = drivers.NewPostgresDriver(spec.Database.DSN)
	default:
		return fmt.Errorf("unsupported database driver: %s", spec.Database.Driver)
	}

	if err := driver.Open(ctx); err != nil {
		return fmt.Errorf("failed to initialize database driver: %w", err)
	}
	defer func() { _ = driver.Close() }()

	// Execute Chaos Run
	runner := engine.NewRunner(driver, spec.Engine.Seed)
	runResult, err := runner.Run(ctx, *spec)
	if err != nil {
		return fmt.Errorf("chaos execution failed: %w", err)
	}

	// Analyze Concurrency Trace with Adya Graph
	graph := analyzer.BuildGraph(runResult.Trace)
	cycles := analyzer.FindCycles(graph)
	anomaly := domain.AnomalyUnknown
	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls == domain.AnomalyWriteSkew {
			anomaly = domain.AnomalyWriteSkew
			break
		}
		if cls == domain.AnomalyA5AReadSkew {
			anomaly = domain.AnomalyA5AReadSkew
			break
		}
		if cls == domain.AnomalyLostUpdate {
			anomaly = domain.AnomalyLostUpdate
		}
	}
	if anomaly == domain.AnomalyUnknown && len(cycles) > 0 {
		anomaly = analyzer.ClassifyCycle(cycles[0])
	}

	var shrinkResult *domain.ShrinkResult
	minimalTrace := runResult.Trace
	minimalOps := runResult.ScheduledOps

	// If an invariant violation is detected, run Delta-Debugging causal reduction
	if runResult.ViolationDetected {
		testFn := func(subset []domain.ScheduledOp) bool {
			res, err := runner.RunSchedule(ctx, *spec, subset)
			if err != nil {
				return true // execution failed, not reproducible
			}
			return !res.ViolationDetected // true = PASS (no bug), false = FAIL (bug reproduced)
		}

		shrunk, err := shrinker.Shrink(ctx, testFn, runResult.ScheduledOps)
		if err == nil && shrunk != nil {
			shrinkResult = shrunk
			minimalOps = shrunk.MinimalOps

			// Capture clean minimal trace
			minRunRes, err := runner.RunSchedule(ctx, *spec, minimalOps)
			if err == nil && minRunRes != nil {
				minimalTrace = minRunRes.Trace
				minGraph := analyzer.BuildGraph(minimalTrace)
				minCycles := analyzer.FindCycles(minGraph)
				for _, c := range minCycles {
					cls := analyzer.ClassifyCycle(c)
					if cls == domain.AnomalyA5AReadSkew {
						anomaly = domain.AnomalyA5AReadSkew
						break
					}
					if cls == domain.AnomalyWriteSkew {
						anomaly = domain.AnomalyWriteSkew
						break
					}
					if cls == domain.AnomalyLostUpdate {
						anomaly = domain.AnomalyLostUpdate
					}
				}
				if anomaly == domain.AnomalyUnknown && len(minCycles) > 0 {
					anomaly = analyzer.ClassifyCycle(minCycles[0])
				}
			}
		}
	}

	// Invariant audit results
	var invResults []domain.InvariantResult
	if runResult.FailingInvariant != nil {
		invResults = append(invResults, *runResult.FailingInvariant)
	}

	// Export repro code if requested
	var reproCode string
	if exportReproFlag || jsonFlag {
		reproCode = reporter.GenerateStandaloneGoRepro(*spec, minimalOps, runResult.FailingInvariant)
		if exportReproFlag {
			reproPath := "repro_test.go"
			if err := os.WriteFile(reproPath, []byte(reproCode), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", reproPath, err)
			}
			if !jsonFlag {
				fmt.Printf("  [✓] Generated standalone Go reproduction: %s\n", reproPath)
			}
		}
	}

	// Export mermaid diagram if requested
	var mermaidCode string
	if exportMermaidFlag || jsonFlag {
		mermaidCode = reporter.GenerateMermaidSequence(minimalTrace)
		if exportMermaidFlag {
			mermaidPath := "trace.mermaid"
			if err := os.WriteFile(mermaidPath, []byte(mermaidCode), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", mermaidPath, err)
			}
			if !jsonFlag {
				fmt.Printf("  [✓] Generated Mermaid sequence diagram: %s\n", mermaidPath)
			}
		}
	}

	// Export HTML report if requested
	var htmlReport string
	if exportHTMLFlag != "" || jsonFlag {
		htmlReport = reporter.GenerateStandaloneHTMLReport(minimalTrace, *spec, graph, shrinkResult, invResults)
		if exportHTMLFlag != "" {
			if err := os.WriteFile(exportHTMLFlag, []byte(htmlReport), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", exportHTMLFlag, err)
			}
			if !jsonFlag {
				fmt.Printf("  [✓] Generated standalone HTML report: %s\n", exportHTMLFlag)
			}
		}
	}

	// Output format
	if jsonFlag {
		output := map[string]interface{}{
			"spec": map[string]interface{}{
				"name":        spec.Name,
				"driver":      spec.Database.Driver,
				"workers":     spec.Engine.Workers,
				"iterations":  spec.Engine.Iterations,
				"seed":        spec.Engine.Seed,
			},
			"success":            runResult.Success,
			"violation_detected": runResult.ViolationDetected,
			"anomaly_type":       anomaly,
			"failing_invariant":  runResult.FailingInvariant,
			"duration_ms":        runResult.Duration.Milliseconds(),
			"trace_events_count": len(runResult.Trace),
			"shrink":             shrinkResult,
			"mermaid":            mermaidCode,
			"repro_go":           reproCode,
			"html_report":        htmlReport,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Terminal Report Output
	reporter.PrintTerminalReport(*spec, runResult, shrinkResult, anomaly)
	return nil
}
