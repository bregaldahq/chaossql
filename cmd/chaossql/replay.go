package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/shrinker"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type ReplayPayload struct {
	Version           int                     `json:"version,omitempty"`
	Spec              *domain.Spec            `json:"spec,omitempty"`
	Seed              uint64                  `json:"seed"`
	Schedule          domain.SchedulePlan     `json:"schedule"`
	Trace             domain.ExecutionTrace   `json:"trace,omitempty"`
	ScheduledOps      []domain.ScheduledOp    `json:"scheduled_ops,omitempty"`
	AnomalyType       domain.AnomalyType      `json:"anomaly_type,omitempty"`
	ViolationDetected bool                    `json:"violation_detected"`
	FailingInvariant  *domain.InvariantResult `json:"failing_invariant,omitempty"`
	Status            domain.ExecutionStatus  `json:"status,omitempty"`
	FailureSignature  domain.FailureSignature `json:"failure_signature,omitempty"`
}

const replayArtifactVersion = 1

func buildReplayArtifact(
	spec domain.Spec,
	result *engine.RunResult,
	ops []domain.ScheduledOp,
	trace domain.ExecutionTrace,
	anomaly domain.AnomalyType,
) (ReplayPayload, error) {
	if result == nil {
		return ReplayPayload{}, fmt.Errorf("cannot build replay artifact from a nil result")
	}
	signature, err := shrinker.FailureSignatureFor(result)
	if err != nil {
		return ReplayPayload{}, err
	}
	spec.Engine.Seed = result.Seed
	storedOps := append([]domain.ScheduledOp(nil), ops...)
	storedTrace := append(domain.ExecutionTrace(nil), trace...)
	return ReplayPayload{
		Version:           replayArtifactVersion,
		Spec:              &spec,
		Seed:              result.Seed,
		Schedule:          engine.BuildSchedulePlan(spec, storedOps, engine.NewPRNG(result.Seed)),
		Trace:             storedTrace,
		ScheduledOps:      storedOps,
		AnomalyType:       anomaly,
		ViolationDetected: result.ViolationDetected,
		FailingInvariant:  result.FailingInvariant,
		Status:            result.Status,
		FailureSignature:  signature,
	}, nil
}

func writeReplayArtifact(path string, payload ReplayPayload) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode replay artifact: %w", err)
	}
	data = append(data, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create replay artifact directory: %w", err)
	}
	temp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create replay artifact: %w", err)
	}
	tempPath := temp.Name()
	defer func() { _ = os.Remove(tempPath) }()
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("secure replay artifact: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write replay artifact: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close replay artifact: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("publish replay artifact: %w", err)
	}
	return nil
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
