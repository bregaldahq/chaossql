package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type ReplayPayload struct {
	Spec              *domain.Spec            `json:"spec,omitempty"`
	Trace             domain.ExecutionTrace   `json:"trace,omitempty"`
	ScheduledOps      []domain.ScheduledOp    `json:"scheduled_ops,omitempty"`
	AnomalyType       domain.AnomalyType      `json:"anomaly_type,omitempty"`
	ViolationDetected bool                    `json:"violation_detected"`
	FailingInvariant  *domain.InvariantResult `json:"failing_invariant,omitempty"`
}

func newReplayCmd() *cobra.Command {
	var maxEvents int

	cmd := &cobra.Command{
		Use:   "replay <result.json>",
		Short: "Replay and inspect an execution trace with an interactive chronological swimlane",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read trace file: %w", err)
			}

			var payload ReplayPayload
			if err := json.Unmarshal(data, &payload); err != nil {
				// Try direct ExecutionTrace array
				var trace domain.ExecutionTrace
				if traceErr := json.Unmarshal(data, &trace); traceErr == nil {
					payload.Trace = trace
				} else {
					return fmt.Errorf("invalid json trace format: %w", err)
				}
			}

			cmd.Println(reporter.RenderBanner())
			renderReplayTerminal(cmd, payload, maxEvents)
			return nil
		},
	}

	cmd.Flags().IntVar(&maxEvents, "max-events", 50, "Maximum number of trace events to display")
	return cmd
}

func renderReplayTerminal(cmd *cobra.Command, p ReplayPayload, maxEvents int) {
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	cmd.Println(headerStyle.Render("INTERACTIVE CHRONOLOGICAL TRACE REPLAYER"))

	var lines string
	lines += fmt.Sprintf("  %-8s  %-10s  %-8s  %-12s  %s\n", "EVENT #", "TIME (µs)", "WORKER", "EVENT TYPE", "SQL STATEMENT / ACTION")
	lines += "  ──────────────────────────────────────────────────────────────────────────────────────────\n"

	total := len(p.Trace)
	if maxEvents > 0 && maxEvents < total {
		total = maxEvents
	}

	for i := 0; i < total; i++ {
		ev := p.Trace[i]
		typeStyle := lipgloss.NewStyle().Bold(true)
		switch ev.Type {
		case domain.EventBegin:
			typeStyle = typeStyle.Foreground(lipgloss.Color("63"))
		case domain.EventCommit:
			typeStyle = typeStyle.Foreground(lipgloss.Color("46"))
		case domain.EventRollback:
			typeStyle = typeStyle.Foreground(lipgloss.Color("196"))
		case domain.EventSavepoint, domain.EventRollbackTo:
			typeStyle = typeStyle.Foreground(lipgloss.Color("214"))
		default:
			typeStyle = typeStyle.Foreground(lipgloss.Color("250"))
		}

		timeUs := fmt.Sprintf("+%dµs", ev.Timestamp.Microseconds())
		workerStr := fmt.Sprintf("W%d", ev.WorkerID)
		lines += fmt.Sprintf("  %-8d  %-10s  %-8s  %-12s  %s\n", i+1, timeUs, workerStr, typeStyle.Render(string(ev.Type)), ev.SQL)
	}

	if len(p.Trace) > total {
		lines += fmt.Sprintf("\n  ... [Showing %d of %d events. Use --max-events to see full trace] ...\n", total, len(p.Trace))
	}

	cmd.Println(cardStyle.Render(lines))
}
