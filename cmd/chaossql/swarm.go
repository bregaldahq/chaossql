package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/swarm"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newSwarmCmd() *cobra.Command {
	var (
		scenariosDir        string
		driversFlag         string
		concurrencyFlag     int
		jsonOutput          bool
		markdownSummaryPath string
	)

	runMatrix := func(cmd *cobra.Command, args []string) error {
		resolvedDir := scenariosDir
		if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
			resolvedDir = strings.TrimSpace(args[0])
		}
		if _, err := os.Stat(resolvedDir); os.IsNotExist(err) {
			altPath := filepath.Join("..", "..", resolvedDir)
			if _, altErr := os.Stat(altPath); altErr == nil {
				resolvedDir = altPath
			} else {
				return fmt.Errorf("scenarios directory %q not found: %w", resolvedDir, err)
			}
		}

		specs, err := discoverAndLoadSpecs(resolvedDir)
		if err != nil {
			return err
		}

		driverNames := parseDriversList(driversFlag)
		if len(driverNames) == 0 {
			driverNames = []string{"sqlite", "mock"}
		}

		if concurrencyFlag <= 0 {
			concurrencyFlag = 4
		}

		ctx := context.Background()
		report, err := swarm.ExecuteDifferentialMatrix(ctx, specs, driverNames, concurrencyFlag)
		if err != nil {
			return fmt.Errorf("differential swarm execution failed: %w", err)
		}

		if markdownSummaryPath != "" {
			summaryMd := reporter.GenerateSwarmMarkdownSummary(report)
			dir := filepath.Dir(markdownSummaryPath)
			if dir != "" && dir != "." {
				_ = os.MkdirAll(dir, 0755)
			}
			if err := os.WriteFile(markdownSummaryPath, []byte(summaryMd), 0644); err != nil {
				return fmt.Errorf("failed to write markdown summary to %s: %w", markdownSummaryPath, err)
			}
		}

		if jsonOutput {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}

		cmd.Println(reporter.RenderBanner())
		renderSwarmTerminal(cmd, report, driverNames)
		return nil
	}

	rootSwarm := &cobra.Command{
		Use:   "swarm",
		Short: "Autonomous multi-engine differential test swarm and stress runner",
		Long: `Swarm runs continuous cross-engine differential isolation fuzzing across multiple
database engines (SQLite, PostgreSQL, MySQL, Mock). It detects semantic divergence where one engine
permits an anomaly while another enforces isolation.`,
		RunE: runMatrix,
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Run differential isolation matrix across multiple database engines",
		RunE:  runMatrix,
	}

	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Execute swarm matrix across scenario suites and target drivers",
		RunE:  runMatrix,
	}

	rootSwarm.PersistentFlags().StringVar(&scenariosDir, "scenarios-dir", "./examples", "Directory containing scenario YAML specifications (chaos.yaml, variant_*.yaml)")
	rootSwarm.PersistentFlags().StringVar(&driversFlag, "drivers", "sqlite,mock", "Comma-separated list of database drivers to evaluate (e.g. sqlite,mock,postgres,mysql)")
	rootSwarm.PersistentFlags().IntVar(&concurrencyFlag, "concurrency", 4, "Maximum concurrent scenario/driver executions")
	rootSwarm.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	rootSwarm.PersistentFlags().StringVar(&markdownSummaryPath, "markdown-summary", "", "Path to write GitHub Flavored Markdown summary report (e.g. $GITHUB_STEP_SUMMARY)")

	rootSwarm.AddCommand(diffCmd, runCmd)
	return rootSwarm
}

func parseDriversList(flagVal string) []string {
	parts := strings.Split(flagVal, ",")
	seen := make(map[string]bool)
	var result []string
	for _, p := range parts {
		trimmed := strings.ToLower(strings.TrimSpace(p))
		if trimmed != "" && !seen[trimmed] {
			seen[trimmed] = true
			result = append(result, trimmed)
		}
	}
	return result
}

func discoverAndLoadSpecs(dirPath string) ([]domain.Spec, error) {
	fi, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access scenarios path %s: %w", dirPath, err)
	}

	var yamlPaths []string

	if !fi.IsDir() {
		yamlPaths = append(yamlPaths, dirPath)
	} else {
		err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			ext := strings.ToLower(filepath.Ext(base))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}
			if base == "chaos.yaml" || base == "chaos.yml" || strings.HasPrefix(base, "variant_") {
				yamlPaths = append(yamlPaths, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed scanning directory %s: %w", dirPath, err)
		}

		// Fallback: If no chaos.yaml or variant_*.yaml found, load any yaml files in the directory
		if len(yamlPaths) == 0 {
			_ = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".yaml" || ext == ".yml" {
					yamlPaths = append(yamlPaths, path)
				}
				return nil
			})
		}
	}

	sort.Strings(yamlPaths)

	var specs []domain.Spec
	for _, yp := range yamlPaths {
		spec, err := domain.LoadSpec(yp)
		if err != nil {
			// Skip files that are not valid chaos specifications
			continue
		}
		specs = append(specs, *spec)
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("no valid ChaosSQL scenario specifications found in %q", dirPath)
	}

	return specs, nil
}

func renderSwarmTerminal(cmd *cobra.Command, report *swarm.DifferentialReport, driverNames []string) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	cmd.Println(headerStyle.Render("AUTONOMOUS MULTI-ENGINE DIFFERENTIAL SWARM REPORT"))

	allEnginesFailed := len(report.Scenarios) > 0
	for _, sc := range report.Scenarios {
		hasValidDriver := false
		for _, dName := range driverNames {
			if res, ok := sc.Results[dName]; ok && res.Error == "" {
				hasValidDriver = true
				break
			}
		}
		if hasValidDriver {
			allEnginesFailed = false
			break
		}
	}

	divergenceBadge := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if allEnginesFailed {
		divergenceBadge = divergenceBadge.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("231")).SetString(" ALL TARGET ENGINES FAILED / OFFLINE ")
	} else if report.DivergentCount > 0 {
		divergenceBadge = divergenceBadge.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("231")).SetString(fmt.Sprintf(" %d DIVERGENCE(S) DETECTED ", report.DivergentCount))
	} else {
		divergenceBadge = divergenceBadge.Background(lipgloss.Color("46")).Foreground(lipgloss.Color("16")).SetString(" ALL ENGINES CONSISTENT ")
	}

	var lines string
	lines += fmt.Sprintf("  • Total Scenarios Evaluated: %d\n", report.TotalScenarios)
	lines += fmt.Sprintf("  • Total Engine Executions:   %d\n", report.TotalExecutions)
	lines += fmt.Sprintf("  • Target Engines:            [%s]\n", strings.Join(driverNames, ", "))
	lines += fmt.Sprintf("  • Overall Execution Time:    %dms\n", report.DurationMs)
	lines += fmt.Sprintf("  • Swarm Status:              %s\n\n", divergenceBadge.String())

	lines += fmt.Sprintf("  %-32s  %-12s  %s\n", "SCENARIO", "OUTCOME", "ENGINE BREAKDOWN")
	lines += "  ────────────────────────────────────────────────────────────────────────────────────────\n"

	for _, sc := range report.Scenarios {
		validCount := 0
		var breakdownParts []string
		for _, dName := range driverNames {
			if res, ok := sc.Results[dName]; ok {
				if res.Error != "" {
					breakdownParts = append(breakdownParts, fmt.Sprintf("%s: ERR(%s)", dName, res.Error))
				} else {
					validCount++
					if res.ViolationDetected {
						desc := res.DetectedAnomaly
						if desc == "" {
							desc = res.FailingInvariant
						}
						breakdownParts = append(breakdownParts, fmt.Sprintf("%s: VIOLATION(%s)", dName, desc))
					} else {
						breakdownParts = append(breakdownParts, fmt.Sprintf("%s: SAFE", dName))
					}
				}
			}
		}

		var statusStr string
		if validCount == 0 {
			statusStr = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("ERROR")
		} else if validCount == 1 {
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("INCONCLUSIVE")
		} else if sc.Divergent {
			statusStr = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Render("DIVERGENT")
		} else {
			statusStr = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("CONSISTENT")
		}

		lines += fmt.Sprintf("  %-32s  %-12s  %s\n", truncateString(sc.ScenarioName, 32), statusStr, strings.Join(breakdownParts, " | "))
		if sc.Divergent || validCount <= 1 {
			reasonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("209"))
			lines += fmt.Sprintf("    ↳ %s\n", reasonStyle.Render(sc.Summary))
		}
	}

	cmd.Println(cardStyle.Render(lines))
}

func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
