package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type MatrixRow struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	AnomalyType string `json:"anomaly_type"`
	Permitted   bool   `json:"permitted"`
	Status      string `json:"status"`
}

type MatrixReport struct {
	Driver string      `json:"driver"`
	Rows   []MatrixRow `json:"rows"`
}

func newMatrixCmd() *cobra.Command {
	var (
		driverName string
		dsn        string
		jsonOutput bool
		mdOutput   bool
	)

	cmd := &cobra.Command{
		Use:   "matrix",
		Short: "Generate an empirical Hermitage transaction isolation matrix for target database",
		RunE: func(cmd *cobra.Command, args []string) error {
			scenarios := []struct {
				relPath string
				code    string
				name    string
			}{
				{"examples/banking_lost_update/chaos.yaml", "P4", "Lost Update"},
				{"examples/inventory_oversell/chaos.yaml", "A3", "Inventory Oversell"},
				{"examples/hospital_write_skew/chaos.yaml", "A5B", "Hospital Write Skew"},
				{"examples/read_skew_financial_audit/chaos.yaml", "A5A", "Financial Read Skew"},
				{"examples/dirty_write_auction/chaos.yaml", "G0", "Auction Dirty Write"},
			}

			driver, err := drivers.GetDriver(driverName, dsn)
			if err != nil {
				return fmt.Errorf("failed to instantiate driver %s: %w", driverName, err)
			}

			ctx := context.Background()
			var rows []MatrixRow

			for _, s := range scenarios {
				specPath := s.relPath
				if _, err := os.Stat(specPath); os.IsNotExist(err) {
					specPath = filepath.Join("..", "..", s.relPath)
				}
				spec, err := domain.LoadSpec(specPath)
				if err != nil {
					continue
				}

				runner := engine.NewRunner(driver, spec.Engine.Seed)
				res, err := runner.Run(ctx, *spec)
				permitted := false
				if err == nil && res.ViolationDetected {
					permitted = true
				}

				status := "PREVENTED (Safe)"
				if permitted {
					status = "PERMITTED (Vulnerable)"
				}

				rows = append(rows, MatrixRow{
					Code:        s.code,
					Name:        s.name,
					AnomalyType: s.code,
					Permitted:   permitted,
					Status:      status,
				})
			}

			report := MatrixReport{
				Driver: driverName,
				Rows:   rows,
			}

			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}

			if mdOutput {
				cmd.Printf("# Empirical Isolation Matrix: %s\n\n", driverName)
				cmd.Println("| Code | Anomaly Name | Permitted by Engine? | Isolation Protection |")
				cmd.Println("| :--- | :--- | :--- | :--- |")
				for _, r := range rows {
					cmd.Printf("| %s | %s | %v | %s |\n", r.Code, r.Name, r.Permitted, r.Status)
				}
				return nil
			}

			cmd.Println(reporter.RenderBanner())
			renderMatrixTerminal(cmd, report)
			return nil
		},
	}

	cmd.Flags().StringVar(&driverName, "driver", "sqlite", "Database driver (sqlite, postgres, mysql)")
	cmd.Flags().StringVar(&dsn, "dsn", ":memory:", "Database DSN")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	cmd.Flags().BoolVar(&mdOutput, "markdown", false, "Output results as Markdown table")

	return cmd
}

func renderMatrixTerminal(cmd *cobra.Command, rep MatrixReport) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	cmd.Println(headerStyle.Render(fmt.Sprintf("EMPIRICAL ISOLATION MATRIX — TARGET DRIVER: %s", rep.Driver)))

	var lines string
	lines += fmt.Sprintf("  %-6s  %-26s  %-12s  %s\n", "CODE", "ANOMALY PHENOMENON", "PERMITTED?", "ENGINE PROTECTION")
	lines += "  ────────────────────────────────────────────────────────────────────────────\n"

	for _, r := range rep.Rows {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")) // Green
		if r.Permitted {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
		}
		lines += fmt.Sprintf("  %-6s  %-26s  %-12v  %s\n", r.Code, r.Name, r.Permitted, statusStyle.Render(r.Status))
	}

	cmd.Println(cardStyle.Render(lines))
}
