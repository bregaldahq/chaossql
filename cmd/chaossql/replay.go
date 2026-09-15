package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
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
	return writeRestrictedAtomic(path, data)
}

func validateReplayArtifact(payload ReplayPayload) error {
	if payload.Version != replayArtifactVersion {
		return fmt.Errorf("unsupported replay artifact version %d", payload.Version)
	}
	if payload.Spec == nil {
		return fmt.Errorf("replay artifact is missing the complete specification")
	}
	if err := payload.Spec.Validate(); err != nil {
		return fmt.Errorf("invalid replay specification: %w", err)
	}
	if len(payload.ScheduledOps) == 0 {
		return fmt.Errorf("replay artifact has no scheduled operations")
	}
	if payload.Schedule.Version != scheduleVersionForReplay {
		return fmt.Errorf("unsupported logical schedule version %d", payload.Schedule.Version)
	}
	if payload.Seed != payload.Schedule.Seed || payload.Seed != payload.Spec.Engine.Seed {
		return fmt.Errorf("replay seed does not match specification and schedule metadata")
	}
	if payload.FailureSignature.Status != payload.Status {
		return fmt.Errorf("failure signature status does not match replay status")
	}
	if payload.Status == domain.StatusViolation {
		if payload.FailingInvariant == nil ||
			payload.FailureSignature.FailingInvariant != payload.FailingInvariant.Name {
			return fmt.Errorf("failure signature does not match the stored failing invariant")
		}
	}
	expected := engine.BuildSchedulePlan(*payload.Spec, payload.ScheduledOps, engine.NewPRNG(payload.Seed))
	if !reflect.DeepEqual(expected, payload.Schedule) {
		return fmt.Errorf("stored logical schedule does not match replay inputs")
	}
	return nil
}

const scheduleVersionForReplay = 1

func verifyReplayArtifact(ctx context.Context, payload ReplayPayload) (*engine.RunResult, error) {
	if err := validateReplayArtifact(payload); err != nil {
		return nil, err
	}
	driver, err := drivers.GetDriver(payload.Spec.Database.Driver, payload.Spec.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("initialize replay database driver: %w", err)
	}
	if err := driver.Open(ctx); err != nil {
		return nil, fmt.Errorf("open replay database driver: %w", err)
	}
	defer func() { _ = driver.Close() }()

	runner := engine.NewRunner(driver, payload.Seed)
	result, runErr := runner.RunSchedule(ctx, *payload.Spec, payload.ScheduledOps)
	if runErr != nil {
		return result, fmt.Errorf("execute replay schedule: %w", runErr)
	}
	if !reflect.DeepEqual(result.Schedule, payload.Schedule) {
		return result, fmt.Errorf("replayed logical schedule differs from stored schedule")
	}
	if !shrinker.ReproducesFailure(result, payload.FailureSignature) {
		return result, fmt.Errorf("stored failure signature was not reproduced")
	}
	return result, nil
}

func newReplayCmd() *cobra.Command {
	var maxEvents int
	var verify bool

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
			if verify {
				result, err := verifyReplayArtifact(cmd.Context(), payload)
				if err != nil {
					return fmt.Errorf("replay verification failed: %w", err)
				}
				cmd.Printf("REPLAY VERIFIED: status=%s seed=%d operations=%d\n", result.Status, result.Seed, len(payload.ScheduledOps))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&maxEvents, "max-events", 50, "Maximum number of trace events to display")
	cmd.Flags().BoolVar(&verify, "verify", false, "Reset the database and verify the stored schedule and failure")
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
