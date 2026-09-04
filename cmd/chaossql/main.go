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

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
	"github.com/spf13/cobra"
)

var (
	seedFlag          uint64
	workersFlag       int
	iterationsFlag    int
	jsonFlag          bool
	exportReproFlag   bool
	exportMermaidFlag bool
	exportHTMLFlag    string
	exportOTELFlag    string
	exportJUnitFlag   string
	exportSummaryFlag string
	exportSARIFFlag   string
	uiFlag            bool
)

func newRunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <chaos.yaml>",
		Short: "Execute a chaos test against the specified database configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  runChaosTest,
	}

	runCmd.Flags().Uint64Var(&seedFlag, "seed", 0, "Deterministic PRNG seed (overrides spec)")
	runCmd.Flags().IntVar(&workersFlag, "workers", 0, "Number of concurrent worker goroutines (overrides spec)")
	runCmd.Flags().IntVar(&iterationsFlag, "iterations", 0, "Total number of operations to execute (overrides spec)")
	runCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as structured JSON payload")
	runCmd.Flags().BoolVar(&exportReproFlag, "export-repro", false, "Export minimal reproduction as standalone Go test (repro_test.go)")
	runCmd.Flags().BoolVar(&exportMermaidFlag, "export-mermaid", false, "Export execution sequence diagram as Mermaid format (trace.mermaid)")
	runCmd.Flags().StringVar(&exportHTMLFlag, "export-html", "", "Export standalone interactive HTML report to file path")
	runCmd.Flags().StringVar(&exportOTELFlag, "export-otel", "", "Export execution trace as OpenTelemetry OTLP JSON to file path")
	runCmd.Flags().StringVar(&exportJUnitFlag, "export-junit", "", "Export execution test results as JUnit XML to file path")
	runCmd.Flags().StringVar(&exportSummaryFlag, "export-summary", "", "Export execution report as GitHub Step Summary markdown to file path")
	runCmd.Flags().StringVar(&exportSARIFFlag, "export-sarif", "", "Export execution security findings as OASIS SARIF 2.1.0 to file path")
	runCmd.Flags().BoolVar(&uiFlag, "ui", false, "Launch interactive trace viewer UI web server after execution")

	return runCmd
}

func newDemoCmd() *cobra.Command {
	demoCmd := &cobra.Command{
		Use:   "demo [banking|inventory|hospital|financial|auction|crypto|flash_crash|ticket|deadlock|fk]",
		Short: "Run one of the flagship demonstration scenarios (Lost Update, Oversell, Write Skew, Read Skew, Dirty Write, Circular Info, Dirty Read, Anti-Dependency, Deadlock, Foreign Key Cascade)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runDemo,
	}
	demoCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output results as structured JSON payload")
	demoCmd.Flags().BoolVar(&exportReproFlag, "export-repro", false, "Export minimal reproduction as standalone Go test (repro_test.go)")
	demoCmd.Flags().BoolVar(&exportMermaidFlag, "export-mermaid", false, "Export execution sequence diagram as Mermaid format (trace.mermaid)")
	demoCmd.Flags().StringVar(&exportHTMLFlag, "export-html", "", "Export standalone interactive HTML report to file path")
	demoCmd.Flags().StringVar(&exportOTELFlag, "export-otel", "", "Export execution trace as OpenTelemetry OTLP JSON to file path")
	demoCmd.Flags().StringVar(&exportJUnitFlag, "export-junit", "", "Export execution test results as JUnit XML to file path")
	demoCmd.Flags().StringVar(&exportSummaryFlag, "export-summary", "", "Export execution report as GitHub Step Summary markdown to file path")
	demoCmd.Flags().StringVar(&exportSARIFFlag, "export-sarif", "", "Export execution security findings as OASIS SARIF 2.1.0 to file path")
	demoCmd.Flags().BoolVar(&uiFlag, "ui", false, "Launch interactive trace viewer UI web server after execution")

	return demoCmd
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "chaossql",
		Short: "ChaosSQL: Deterministic Concurrency & Invariant Fuzzer for SQL Databases",
		Long: `ChaosSQL bridges chaos engineering with formal academic database research (PCT, Elle, Hermitage).
It injects stochastic interleavings across database worker threads to provoke subtle isolation anomalies
(such as Lost Updates, Write Skew, and Phantom Reads) and applies causal Delta-Debugging to shrink
noisy execution traces to minimal, deterministic reproductions.`,
	}

	runCmd := newRunCmd()
	demoCmd := newDemoCmd()
	benchCmd := newBenchCmd()
	diffCmd := newDiffCmd()
	matrixCmd := newMatrixCmd()
	replayCmd := newReplayCmd()
	initCmd := newInitCmd()
	validateCmd := newValidateCmd()

	uiCmd := newUICmd()
	rootCmd.AddCommand(runCmd, demoCmd, benchCmd, diffCmd, matrixCmd, replayCmd, initCmd, validateCmd, uiCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runChaosTest(cmd *cobra.Command, args []string) error {
	specPath := args[0]
	return executeChaos(specPath)
}

func resolveDemoPath(scenario string) (string, error) {
	switch scenario {
	case "banking", "banking_lost_update", "lost_update":
		return "examples/banking_lost_update/chaos.yaml", nil
	case "inventory", "inventory_oversell", "oversell":
		return "examples/inventory_oversell/chaos.yaml", nil
	case "hospital", "hospital_write_skew", "write_skew":
		return "examples/hospital_write_skew/chaos.yaml", nil
	case "financial", "financial_audit", "read_skew", "read_skew_financial_audit":
		return "examples/read_skew_financial_audit/chaos.yaml", nil
	case "auction", "dirty_write", "auction_dirty_write":
		return "examples/dirty_write_auction/chaos.yaml", nil
	case "crypto", "circular_info", "crypto_arbitrage", "circular_info_crypto_arbitrage":
		return "examples/circular_info_crypto_arbitrage/chaos.yaml", nil
	case "flash_crash", "dirty_read", "dirty_read_flash_crash", "liquidation":
		return "examples/dirty_read_flash_crash/chaos.yaml", nil
	case "ticket", "tickets", "seat", "seats", "g2", "ticket_booking_anti_dependency":
		return "examples/ticket_booking_anti_dependency/chaos.yaml", nil
	case "deadlock", "deadlock_cycle":
		return "examples/deadlock_cycle/chaos.yaml", nil
	case "fk", "cascade", "foreign_key", "foreign_key_cascade_deadlock":
		return "examples/foreign_key_cascade_deadlock/chaos.yaml", nil
	default:
		return "", fmt.Errorf("unknown demo scenario %q. Available scenarios: banking, inventory, hospital, financial, auction, crypto, flash_crash, ticket, deadlock, fk", scenario)
	}
}

func runDemo(cmd *cobra.Command, args []string) error {
	scenario := "banking"
	if len(args) > 0 {
		scenario = strings.ToLower(strings.TrimSpace(args[0]))
	}

	specPath, err := resolveDemoPath(scenario)
	if err != nil {
		return err
	}

	if _, err := os.Stat(specPath); os.IsNotExist(err) {
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

	if seedFlag > 0 {
		spec.Engine.Seed = seedFlag
	}
	if workersFlag > 0 {
		spec.Engine.Workers = workersFlag
	}
	if iterationsFlag > 0 {
		spec.Engine.Iterations = iterationsFlag
	}

	driver, err := drivers.GetDriver(spec.Database.Driver, spec.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize database driver: %w", err)
	}

	if err := driver.Open(ctx); err != nil {
		return fmt.Errorf("failed to initialize database driver: %w", err)
	}
	defer func() { _ = driver.Close() }()

	runner := engine.NewRunner(driver, spec.Engine.Seed)
	runResult, err := runner.Run(ctx, *spec)
	if err != nil {
		return fmt.Errorf("chaos execution failed: %w", err)
	}

	graph := analyzer.BuildGraph(runResult.Trace)
	cycles := analyzer.FindCycles(graph)
	anomaly := domain.AnomalyUnknown
	for _, c := range cycles {
		cls := analyzer.ClassifyCycle(c)
		if cls == domain.AnomalyG1aDirtyRead {
			anomaly = domain.AnomalyG1aDirtyRead
			break
		}
		if cls == domain.AnomalyG0DirtyWrite {
			anomaly = domain.AnomalyG0DirtyWrite
			break
		}
		if cls == domain.AnomalyG1cCircularInfo {
			anomaly = domain.AnomalyG1cCircularInfo
			break
		}
		if cls == domain.AnomalyG2AntiDependency {
			anomaly = domain.AnomalyG2AntiDependency
			break
		}
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

	if runResult.ViolationDetected {
		testFn := func(subset []domain.ScheduledOp) bool {
			res, err := runner.RunSchedule(ctx, *spec, subset)
			if err != nil {
				return true
			}
			return !res.ViolationDetected
		}

		shrunk, err := shrinker.Shrink(ctx, testFn, runResult.ScheduledOps)
		if err == nil && shrunk != nil {
			shrinkResult = shrunk
			minimalOps = shrunk.MinimalOps

			minRunRes, err := runner.RunSchedule(ctx, *spec, minimalOps)
			if err == nil && minRunRes != nil {
				minimalTrace = minRunRes.Trace
				minGraph := analyzer.BuildGraph(minimalTrace)
				minCycles := analyzer.FindCycles(minGraph)
				for _, c := range minCycles {
					cls := analyzer.ClassifyCycle(c)
					if cls == domain.AnomalyG1aDirtyRead {
						anomaly = domain.AnomalyG1aDirtyRead
						break
					}
					if cls == domain.AnomalyG0DirtyWrite {
						anomaly = domain.AnomalyG0DirtyWrite
						break
					}
					if cls == domain.AnomalyG1cCircularInfo {
						anomaly = domain.AnomalyG1cCircularInfo
						break
					}
					if cls == domain.AnomalyG2AntiDependency {
						anomaly = domain.AnomalyG2AntiDependency
						break
					}
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

	var invResults []domain.InvariantResult
	if runResult.FailingInvariant != nil {
		invResults = append(invResults, *runResult.FailingInvariant)
	}

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

	var otelTrace string
	if exportOTELFlag != "" || jsonFlag {
		var otelErr error
		otelTrace, otelErr = reporter.GenerateOTLPTraceJSON(minimalTrace, *spec)
		if otelErr != nil {
			return fmt.Errorf("failed to generate OpenTelemetry trace: %w", otelErr)
		}
		if exportOTELFlag != "" {
			if err := os.WriteFile(exportOTELFlag, []byte(otelTrace), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", exportOTELFlag, err)
			}
			if !jsonFlag {
				fmt.Printf("  [✓] Generated OpenTelemetry OTLP trace: %s\n", exportOTELFlag)
			}
		}
	}

	if exportJUnitFlag != "" {
		junitXML := reporter.GenerateJUnitXML(*spec, *runResult, anomaly)
		if err := os.WriteFile(exportJUnitFlag, []byte(junitXML), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", exportJUnitFlag, err)
		}
		if !jsonFlag {
			fmt.Printf("  [✓] Generated JUnit XML test report: %s\n", exportJUnitFlag)
		}
	}

	if exportSummaryFlag != "" {
		summaryMD := reporter.GenerateGitHubSummaryMarkdown(*spec, *runResult, shrinkResult, anomaly)
		if err := os.WriteFile(exportSummaryFlag, []byte(summaryMD), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", exportSummaryFlag, err)
		}
		if !jsonFlag {
			fmt.Printf("  [✓] Generated GitHub Step Summary markdown: %s\n", exportSummaryFlag)
		}
	}

	if exportSARIFFlag != "" {
		sarifJSON, err := reporter.GenerateSARIFReport(*spec, invResults, graph, shrinkResult)
		if err != nil {
			return fmt.Errorf("failed to generate SARIF report: %w", err)
		}
		if err := os.WriteFile(exportSARIFFlag, []byte(sarifJSON), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", exportSARIFFlag, err)
		}
		if !jsonFlag {
			fmt.Printf("  [✓] Generated OASIS SARIF 2.1.0 security report: %s\n", exportSARIFFlag)
		}
	}

	if jsonFlag {
		output := map[string]interface{}{
			"spec": map[string]interface{}{
				"name":       spec.Name,
				"driver":     spec.Database.Driver,
				"workers":    spec.Engine.Workers,
				"iterations": spec.Engine.Iterations,
				"seed":       spec.Engine.Seed,
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
			"otel_trace":         otelTrace,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	reporter.PrintTerminalReport(*spec, runResult, shrinkResult, anomaly)

	if uiFlag {
		htmlContent := reporter.GenerateEmbeddedTraceViewerHTML(minimalTrace, *spec, graph, shrinkResult, invResults)
		return serveTraceViewer("127.0.0.1:8090", htmlContent, false)
	}

	return nil
}
