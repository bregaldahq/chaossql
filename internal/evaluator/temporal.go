package evaluator

import (
	"fmt"

	"github.com/bregaldahq/chaossql/internal/domain"
)

// EvaluateTemporalInvariants verifies temporal and event-stream invariant rules over the trace.
func EvaluateTemporalInvariants(trace domain.ExecutionTrace, invs []domain.TemporalInvariantConfig) []domain.InvariantResult {
	var results []domain.InvariantResult

	for _, inv := range invs {
		switch inv.Type {
		case "no_aborts":
			abortedCount := 0
			for _, ev := range trace {
				if ev.Type == domain.EventRollback {
					abortedCount++
				}
			}
			passed := abortedCount == 0
			results = append(results, domain.InvariantResult{
				Name:       inv.Name,
				Passed:     passed,
				Expression: "aborted_count == 0",
				ActualValues: map[string]interface{}{
					"aborted_count": abortedCount,
				},
			})

		case "no_error_events":
			errorCount := 0
			for _, ev := range trace {
				if ev.Type == domain.EventError || ev.Error != "" {
					errorCount++
				}
			}
			passed := errorCount == 0
			results = append(results, domain.InvariantResult{
				Name:       inv.Name,
				Passed:     passed,
				Expression: "error_count == 0",
				ActualValues: map[string]interface{}{
					"error_count": errorCount,
				},
			})

		case "monotonicity":
			monotonic := true
			for i := 1; i < len(trace); i++ {
				if trace[i].Timestamp < trace[i-1].Timestamp {
					monotonic = false
					break
				}
			}
			results = append(results, domain.InvariantResult{
				Name:       inv.Name,
				Passed:     monotonic,
				Expression: "timestamp_monotonic_increasing == true",
				ActualValues: map[string]interface{}{
					"monotonic": monotonic,
				},
			})

		default:
			results = append(results, domain.InvariantResult{
				Name:       inv.Name,
				Passed:     false,
				Expression: inv.Type,
				Error:      fmt.Errorf("unknown temporal invariant type: %s", inv.Type),
			})
		}
	}

	return results
}
