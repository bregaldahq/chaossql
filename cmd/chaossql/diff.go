package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func newDiffCmd() *cobra.Command {
	var (
		driverAName string
		driverBName string
		dsnA        string
		dsnB        string
		seedFlag    uint64
		jsonOutput  bool
	)

	cmd := &cobra.Command{
		Use:   "diff <spec.yaml>",
		Short: "Run cross-engine differential fuzzing comparing two database drivers",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath := args[0]
			spec, err := domain.LoadSpec(specPath)
			if err != nil {
				return fmt.Errorf("failed to load spec: %w", err)
			}

			if seedFlag > 0 {
				spec.Engine.Seed = seedFlag
			}

			driverA, err := drivers.GetDriver(driverAName, dsnA)
			if err != nil {
				return fmt.Errorf("invalid driver A: %w", err)
			}
			driverB, err := drivers.GetDriver(driverBName, dsnB)
			if err != nil {
				return fmt.Errorf("invalid driver B: %w", err)
			}

			ctx := context.Background()
			diffRes, err := engine.RunDifferentialFuzzing(ctx, *spec, driverA, driverB, spec.Engine.Seed)
			if err != nil {
				return fmt.Errorf("differential fuzzing failed: %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(diffRes)
			}

			cmd.Println(reporter.RenderBanner())
			renderDiffTerminal(cmd, diffRes)
			return nil
		},
	}

	cmd.Flags().StringVar(&driverAName, "driver-a", "sqlite", "Name of primary driver (sqlite, postgres, mysql)")
	cmd.Flags().StringVar(&driverBName, "driver-b", "sqlite", "Name of secondary driver (sqlite, postgres, mysql)")
	cmd.Flags().StringVar(&dsnA, "dsn-a", ":memory:", "DSN for driver A")
	cmd.Flags().StringVar(&dsnB, "dsn-b", ":memory:", "DSN for driver B")
	cmd.Flags().Uint64Var(&seedFlag, "seed", 0, "Override random seed")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")

	return cmd
}

func renderDiffTerminal(cmd *cobra.Command, res *domain.DiffResult) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	divergenceBadge := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	if res.Divergent {
		divergenceBadge = divergenceBadge.Background(lipgloss.Color("196")).Foreground(lipgloss.Color("231")).SetString(" SEMANTIC DIVERGENCE DETECTED ")
	} else {
		divergenceBadge = divergenceBadge.Background(lipgloss.Color("46")).Foreground(lipgloss.Color("16")).SetString(" BEHAVIOR CONSISTENT ACROSS ENGINES ")
	}

	content := fmt.Sprintf(
		"%s\n\n  • Scenario: %s\n  • Driver A: %s (Violation: %v)\n  • Driver B: %s (Violation: %v)\n\n  Status: %s\n\n  Summary: %s",
		headerStyle.Render("DIFFERENTIAL ISOLATION FUZZING REPORT"),
		res.ScenarioName,
		res.DriverA, res.ResultA.ViolationDetected,
		res.DriverB, res.ResultB.ViolationDetected,
		divergenceBadge.String(),
		res.DiffSummary,
	)

	cmd.Println(cardStyle.Render(content))
}
