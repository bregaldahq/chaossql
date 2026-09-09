package main

import (
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/charmbracelet/lipgloss"
	"github.com/expr-lang/expr"
	"github.com/spf13/cobra"
)

type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Level   string `json:"level"` // "ERROR" or "WARNING"
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate <chaos.yaml>",
		Short: "Statically validate and lint a chaos testing specification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specPath := args[0]
			spec, err := domain.LoadSpec(specPath)
			if err != nil {
				return fmt.Errorf("spec load error: %w", err)
			}

			issues := validateScenarioSpec(spec)

			cmd.Println(reporter.RenderBanner())
			renderValidationTerminal(cmd, spec, specPath, issues)

			hasErrors := false
			for _, iss := range issues {
				if iss.Level == "ERROR" {
					hasErrors = true
					break
				}
			}

			if hasErrors {
				return fmt.Errorf("scenario validation failed with %d error(s)", len(issues))
			}

			return nil
		},
	}

	return cmd
}

func validateScenarioSpec(spec *domain.Spec) []ValidationIssue {
	var issues []ValidationIssue

	if spec.Database.Driver != "sqlite" && spec.Database.Driver != "postgres" && spec.Database.Driver != "mysql" {
		issues = append(issues, ValidationIssue{
			Field:   "database.driver",
			Message: fmt.Sprintf("driver %q may not be officially supported (supported: sqlite, postgres, mysql)", spec.Database.Driver),
			Level:   "WARNING",
		})
	}

	if len(spec.Database.Schema) == 0 {
		issues = append(issues, ValidationIssue{
			Field:   "database.schema",
			Message: "schema SQL is empty",
			Level:   "ERROR",
		})
	}

	if len(spec.Database.Seed) == 0 {
		issues = append(issues, ValidationIssue{
			Field:   "database.seed",
			Message: "seed SQL is empty",
			Level:   "WARNING",
		})
	}

	for i, inv := range spec.Invariants {
		if inv.Name == "" {
			issues = append(issues, ValidationIssue{
				Field:   fmt.Sprintf("invariants[%d].name", i),
				Message: "invariant missing name",
				Level:   "ERROR",
			})
		}
		if inv.Query == "" {
			issues = append(issues, ValidationIssue{
				Field:   fmt.Sprintf("invariants[%d].query", i),
				Message: "invariant missing SQL query",
				Level:   "ERROR",
			})
		}
		if inv.Assert == "" {
			issues = append(issues, ValidationIssue{
				Field:   fmt.Sprintf("invariants[%d].assert", i),
				Message: "invariant missing assert expression",
				Level:   "ERROR",
			})
		} else {
			// Static compile check
			_, err := expr.Compile(inv.Assert, expr.AsBool())
			if err != nil {
				issues = append(issues, ValidationIssue{
					Field:   fmt.Sprintf("invariants[%d].assert", i),
					Message: fmt.Sprintf("invalid assert expression %q: %v", inv.Assert, err),
					Level:   "ERROR",
				})
			}
		}
	}

	for i, op := range spec.Operations {
		if op.Name == "" {
			issues = append(issues, ValidationIssue{
				Field:   fmt.Sprintf("operations[%d].name", i),
				Message: "operation missing name",
				Level:   "ERROR",
			})
		}
		if len(op.Steps) == 0 {
			issues = append(issues, ValidationIssue{
				Field:   fmt.Sprintf("operations[%d].steps", i),
				Message: "operation has no steps",
				Level:   "ERROR",
			})
		}
	}

	return issues
}

func renderValidationTerminal(cmd *cobra.Command, spec *domain.Spec, path string, issues []ValidationIssue) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	cmd.Println(headerStyle.Render(fmt.Sprintf("STATIC SPECIFICATION VALIDATOR — %s", spec.Name)))

	var lines string
	lines += fmt.Sprintf("  • Path: %s\n", path)
	lines += fmt.Sprintf("  • Driver: %s\n", spec.Database.Driver)
	lines += fmt.Sprintf("  • Invariants: %d\n", len(spec.Invariants))
	lines += fmt.Sprintf("  • Operations: %d\n\n", len(spec.Operations))

	if len(issues) == 0 {
		statusStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
		lines += statusStyle.Render("  ✔ SPECIFICATION IS 100% VALID AND READY FOR EXECUTION\n")
	} else {
		lines += "  VALIDATION FINDINGS:\n"
		lines += "  ────────────────────────────────────────────────────────────────────────────\n"
		for _, iss := range issues {
			levelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
			if iss.Level == "WARNING" {
				levelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
			}
			lines += fmt.Sprintf("  [%s] %-25s: %s\n", levelStyle.Render(iss.Level), iss.Field, iss.Message)
		}
	}

	cmd.Println(cardStyle.Render(lines))
}
