package reporter_test

import (
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/reporter"
	"github.com/bregaldahq/chaossql/internal/swarm"
)

func TestGenerateSwarmMarkdownSummary_NilReport(t *testing.T) {
	md := reporter.GenerateSwarmMarkdownSummary(nil)
	if !strings.Contains(md, "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit") {
		t.Errorf("expected audit header for nil report, got:\n%s", md)
	}
}

func TestGenerateSwarmMarkdownSummary_EmptyReport(t *testing.T) {
	report := &swarm.DifferentialReport{
		TotalScenarios:  0,
		DivergentCount:  0,
		TotalExecutions: 0,
		Scenarios:       []swarm.ScenarioDifferential{},
		DurationMs:      0,
	}

	md := reporter.GenerateSwarmMarkdownSummary(report)

	// Check main header
	if !strings.Contains(md, "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit") {
		t.Errorf("missing header in summary, got:\n%s", md)
	}

	// Check executive table metrics
	if !strings.Contains(md, "Total Scenarios") || !strings.Contains(md, "0") {
		t.Errorf("expected 0 total scenarios in executive metrics, got:\n%s", md)
	}
	if !strings.Contains(md, "Total Engine Executions") {
		t.Errorf("expected Total Engine Executions row in executive metrics, got:\n%s", md)
	}
	if !strings.Contains(md, "Divergences Detected") {
		t.Errorf("expected Divergences Detected row in executive metrics, got:\n%s", md)
	}
	if !strings.Contains(md, "Duration") {
		t.Errorf("expected Duration row in executive metrics, got:\n%s", md)
	}
	if !strings.Contains(md, "Target Engines") {
		t.Errorf("expected Target Engines row in executive metrics, got:\n%s", md)
	}
}

func TestGenerateSwarmMarkdownSummary_DivergentReport(t *testing.T) {
	report := &swarm.DifferentialReport{
		TotalScenarios:  1,
		DivergentCount:  1,
		TotalExecutions: 2,
		DurationMs:      150,
		Scenarios: []swarm.ScenarioDifferential{
			{
				ScenarioName: "banking_lost_update",
				Divergent:    true,
				Summary:      "Semantic divergence detected: Driver sqlite (violation=true) != Driver mock (violation=false)",
				Results: map[string]swarm.DriverExecutionResult{
					"sqlite": {
						Driver:            "sqlite",
						Success:           false,
						ViolationDetected: true,
						FailingInvariant:  "balance_positive",
						DetectedAnomaly:   "lost_update",
						DurationMs:        75,
					},
					"mock": {
						Driver:            "mock",
						Success:           true,
						ViolationDetected: false,
						DurationMs:        20,
					},
				},
			},
		},
	}

	md := reporter.GenerateSwarmMarkdownSummary(report)

	// Header check
	if !strings.Contains(md, "# 🌪️ ChaosSQL Multi-Engine Swarm Differential Audit") {
		t.Errorf("missing main header")
	}

	// Executive table checks
	requiredMetrics := []string{
		"Total Scenarios",
		"Total Engine Executions",
		"Divergences Detected",
		"Duration",
		"Target Engines",
	}
	for _, m := range requiredMetrics {
		if !strings.Contains(md, m) {
			t.Errorf("executive table missing metric %q", m)
		}
	}
	if !strings.Contains(md, "150ms") {
		t.Errorf("expected 150ms duration in metrics table")
	}

	// Matrix table checks
	tableHeaders := []string{"Scenario", "Status", "Engine Breakdown"}
	for _, h := range tableHeaders {
		if !strings.Contains(md, h) {
			t.Errorf("matrix results table missing column header %q", h)
		}
	}

	// Status badge checks
	if !strings.Contains(md, "DIVERGENT") {
		t.Errorf("expected DIVERGENT status in matrix results table")
	}

	// Engine breakdown check
	if !strings.Contains(md, "sqlite") || !strings.Contains(md, "lost_update") {
		t.Errorf("engine breakdown missing sqlite anomaly details")
	}
	if !strings.Contains(md, "mock") || !strings.Contains(md, "Safe") {
		t.Errorf("engine breakdown missing mock safe status")
	}

	// Detailed divergence section callout
	if !strings.Contains(md, "[!WARNING]") {
		t.Errorf("expected [!WARNING] callout alert for divergence")
	}
	if !strings.Contains(md, "Driver sqlite (violation=true) != Driver mock (violation=false)") {
		t.Errorf("missing detailed divergence summary in callout section")
	}

	// Reproducibility section
	if !strings.Contains(md, "chaossql swarm diff") {
		t.Errorf("expected CLI reproduction command in reproducibility section")
	}
}

func TestGenerateSwarmMarkdownSummary_ConsistentReport(t *testing.T) {
	report := &swarm.DifferentialReport{
		TotalScenarios:  1,
		DivergentCount:  0,
		TotalExecutions: 2,
		DurationMs:      80,
		Scenarios: []swarm.ScenarioDifferential{
			{
				ScenarioName: "deadlock_cycle",
				Divergent:    false,
				Summary:      "All engines satisfied invariants consistently",
				Results: map[string]swarm.DriverExecutionResult{
					"sqlite": {
						Driver:            "sqlite",
						Success:           true,
						ViolationDetected: false,
						DurationMs:        40,
					},
					"mock": {
						Driver:            "mock",
						Success:           true,
						ViolationDetected: false,
						DurationMs:        10,
					},
				},
			},
		},
	}

	md := reporter.GenerateSwarmMarkdownSummary(report)

	if !strings.Contains(md, "CONSISTENT") {
		t.Errorf("expected CONSISTENT status badge in matrix results table")
	}
	if !strings.Contains(md, "[!NOTE]") {
		t.Errorf("expected [!NOTE] callout alert for consistent run")
	}
	if strings.Contains(md, "[!WARNING]") {
		t.Errorf("unexpected [!WARNING] callout alert in consistent run")
	}
}

func TestGenerateSwarmMarkdownSummary_ErrorReport(t *testing.T) {
	report := &swarm.DifferentialReport{
		TotalScenarios:  1,
		DivergentCount:  0,
		TotalExecutions: 0,
		DurationMs:      10,
		Scenarios: []swarm.ScenarioDifferential{
			{
				ScenarioName: "network_partition",
				Divergent:    false,
				Summary:      "All drivers encountered execution or connection errors",
				Results: map[string]swarm.DriverExecutionResult{
					"postgres": {
						Driver:     "postgres",
						Success:    false,
						DurationMs: 5,
						Error:      "dial tcp 127.0.0.1:5432: connect: connection refused",
					},
					"mysql": {
						Driver:     "mysql",
						Success:    false,
						DurationMs: 5,
						Error:      "dial tcp 127.0.0.1:3306: connect: connection refused",
					},
				},
			},
		},
	}

	md := reporter.GenerateSwarmMarkdownSummary(report)

	if !strings.Contains(md, "ERROR") {
		t.Errorf("expected ERROR status badge in matrix results table")
	}
	if !strings.Contains(md, "connection refused") {
		t.Errorf("expected error details in engine breakdown")
	}
}