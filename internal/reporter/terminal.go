package reporter

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
)

var (
	// Theme Colors
	colorPrimary = lipgloss.Color("#7D56F4")
	colorGreen   = lipgloss.Color("#04B575")
	colorRed     = lipgloss.Color("#FF5F87")
	colorYellow  = lipgloss.Color("#FFAF00")
	colorCyan    = lipgloss.Color("#00D7FF")
	colorGray    = lipgloss.Color("#626262")

	// Base Styles
	boldStyle = lipgloss.NewStyle().Bold(true)

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorPrimary).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	successBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorGreen).
			Padding(0, 1)

	violationBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorRed).
			Padding(0, 1)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorCyan).
				Padding(0, 1)

	tableRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E0E0E0")).
			Padding(0, 1)

	passStyle = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	failStyle = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
)

// RenderBanner returns the ChaosSQL CLI ASCII Banner.
func RenderBanner() string {
	rawBanner := `  ____ _                     ____   ___  _     
 / ___| |__   __ _  ___  ___/ ___| / _ \| |    
| |   | '_ \ / _` + "`" + ` |/ _ \/ __\___ \| | | | |    
| |___| | | | (_| | (_) \__ \___) | |_| | |___ 
 \____|_| |_|\__,_|\___/|___/____/ \__\_\_____|`

	tagline := lipgloss.NewStyle().
		Foreground(colorCyan).
		Italic(true).
		Render("ChaosSQL • Causal Concurrency Stress Testing & Anomaly Synthesis")

	content := fmt.Sprintf("%s\n\n%s", rawBanner, tagline)
	return bannerStyle.Render(content)
}

// RenderRunSummary formats the execution summary box.
func RenderRunSummary(spec domain.Spec, result *engine.RunResult, anomaly domain.AnomalyType) string {
	var sb strings.Builder

	title := boldStyle.Foreground(colorCyan).Render("EXECUTION SUMMARY")
	sb.WriteString(fmt.Sprintf("%s\n\n", title))

	sb.WriteString(fmt.Sprintf("  • %s: %s\n", boldStyle.Render("Scenario"), spec.Name))
	if spec.Description != "" {
		sb.WriteString(fmt.Sprintf("  • %s: %s\n", boldStyle.Render("Description"), spec.Description))
	}
	sb.WriteString(fmt.Sprintf("  • %s: %s\n", boldStyle.Render("Database Driver"), spec.Database.Driver))
	sb.WriteString(fmt.Sprintf("  • %s: %d workers | %d iterations | seed=%d\n",
		boldStyle.Render("Engine Parameters"),
		spec.Engine.Workers,
		spec.Engine.Iterations,
		spec.Engine.Seed,
	))

	if result != nil {
		sb.WriteString(fmt.Sprintf("  • %s: %s\n\n", boldStyle.Render("Elapsed Time"), result.Duration.Round(100*1000)))

		if result.ViolationDetected {
			anomalyStr := string(anomaly)
			if anomalyStr == "" || anomalyStr == string(domain.AnomalyUnknown) {
				anomalyStr = "INVARIANT_VIOLATION"
			}
			badge := violationBadge.Render(" ✘ ISOLATION ANOMALY DETECTED ")
			sb.WriteString(fmt.Sprintf("  Status: %s  %s\n", badge, boldStyle.Foreground(colorRed).Render("["+anomalyStr+"]")))
		} else {
			badge := successBadge.Render(" ✔ ALL INVARIANTS SATISFIED ")
			sb.WriteString(fmt.Sprintf("  Status: %s\n", badge))
		}
	}

	return boxStyle.Render(sb.String())
}

// RenderInvariantTable formats invariant evaluation results in a styled table.
func RenderInvariantTable(invariants []domain.InvariantResult) string {
	if len(invariants) == 0 {
		return ""
	}

	var sb strings.Builder
	title := boldStyle.Foreground(colorCyan).Render("INVARIANT INTEGRITY AUDIT")
	sb.WriteString(fmt.Sprintf("%s\n\n", title))

	// Header
	sb.WriteString(fmt.Sprintf("  %-30s %-8s %-30s %s\n",
		tableHeaderStyle.Render("INVARIANT"),
		tableHeaderStyle.Render("STATUS"),
		tableHeaderStyle.Render("ASSERTION EXPRESSION"),
		tableHeaderStyle.Render("ACTUAL DATABASE STATE"),
	))
	sb.WriteString("  " + strings.Repeat("─", 100) + "\n")

	for _, inv := range invariants {
		statusStr := passStyle.Render("PASS")
		if !inv.Passed {
			statusStr = failStyle.Render("FAIL")
		}

		stateStr := fmt.Sprintf("%v", inv.ActualValues)
		if inv.Error != nil {
			stateStr = fmt.Sprintf("Error: %v", inv.Error)
		}

		exprStr := inv.Expression
		if len(exprStr) > 28 {
			exprStr = exprStr[:25] + "..."
		}

		sb.WriteString(fmt.Sprintf("  %-30s %-16s %-30s %s\n",
			tableRowStyle.Render(inv.Name),
			statusStr,
			tableRowStyle.Render(exprStr),
			tableRowStyle.Render(stateStr),
		))
	}

	return boxStyle.Render(sb.String())
}

// RenderShrinkSummary formats the Delta-Debugging reduction summary.
func RenderShrinkSummary(shrink *domain.ShrinkResult) string {
	if shrink == nil {
		return ""
	}

	var sb strings.Builder
	title := boldStyle.Foreground(colorCyan).Render("DELTA-DEBUGGING CAUSAL REDUCTION (ddmin)")
	sb.WriteString(fmt.Sprintf("%s\n\n", title))

	pctReduction := shrink.ReductionRatio
	if pctReduction <= 1.0 {
		pctReduction *= 100.0
	}
	sb.WriteString(fmt.Sprintf("  • %s: %d ops  ──►  %s (%s reduction)\n",
		boldStyle.Render("Reduction"),
		shrink.OriginalSize,
		boldStyle.Foreground(colorGreen).Render(fmt.Sprintf("%d ops", shrink.ReducedSize)),
		boldStyle.Foreground(colorYellow).Render(fmt.Sprintf("%.1f%%", pctReduction)),
	))
	sb.WriteString(fmt.Sprintf("  • %s: %d iterations | Duration: %s\n\n",
		boldStyle.Render("Algorithm Cost"),
		shrink.Iterations,
		shrink.Duration.Round(100*1000),
	))

	if len(shrink.MinimalOps) > 0 {
		sb.WriteString(fmt.Sprintf("  %s:\n", boldStyle.Render("Minimal Schedule Sequence")))
		for _, op := range shrink.MinimalOps {
			paramParts := []string{}
			for k, v := range op.Params {
				paramParts = append(paramParts, fmt.Sprintf("%s=%s", k, v))
			}
			paramStr := strings.Join(paramParts, ", ")
			if paramStr != "" {
				paramStr = " {" + paramStr + "}"
			}
			sb.WriteString(fmt.Sprintf("    [%s #%d]%s (%d steps)\n",
				boldStyle.Foreground(colorCyan).Render(op.Name),
				op.ID,
				paramStr,
				len(op.Steps),
			))
			for stepIdx, step := range op.Steps {
				capStr := ""
				if step.Capture != "" {
					capStr = fmt.Sprintf(" -> capture(%s)", step.Capture)
				}
				sb.WriteString(fmt.Sprintf("      step %d: %s%s\n", stepIdx+1, sanitizeSQLForMermaid(step.SQL), capStr))
			}
		}
	}

	return boxStyle.Render(sb.String())
}

// RenderFullReport composes the entire terminal report string.
func RenderFullReport(spec domain.Spec, result *engine.RunResult, shrink *domain.ShrinkResult, anomaly domain.AnomalyType) string {
	var sb strings.Builder

	sb.WriteString(RenderBanner())
	sb.WriteString("\n")

	sb.WriteString(RenderRunSummary(spec, result, anomaly))
	sb.WriteString("\n")

	if result != nil {
		var invs []domain.InvariantResult
		if result.FailingInvariant != nil {
			invs = append(invs, *result.FailingInvariant)
		}
		if len(invs) > 0 {
			sb.WriteString(RenderInvariantTable(invs))
			sb.WriteString("\n")
		}
	}

	if shrink != nil {
		sb.WriteString(RenderShrinkSummary(shrink))
		sb.WriteString("\n")
	}

	hint := lipgloss.NewStyle().Foreground(colorGray).Italic(true).Render(
		"Artifacts: Use --export-repro to emit standalone repro_test.go | --export-mermaid to emit sequence diagram | --export-html to emit HTML report | --export-otel to emit OTLP trace JSON.",
	)
	sb.WriteString(fmt.Sprintf("  %s\n", hint))

	return sb.String()
}

// PrintTerminalReport outputs the complete report to stdout.
func PrintTerminalReport(spec domain.Spec, result *engine.RunResult, shrink *domain.ShrinkResult, anomaly domain.AnomalyType) {
	fmt.Println(RenderFullReport(spec, result, shrink, anomaly))
}
