package reporter_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateSARIFReport_AnomalyDetected(t *testing.T) {
	spec := domain.Spec{
		Version:     "1.0",
		Name:        "banking_lost_update",
		Description: "Detects concurrent lost updates under READ COMMITTED isolation",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "schema.sql",
			Seed:   "seed.sql",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 10,
			Seed:       42,
		},
	}

	invResults := []domain.InvariantResult{
		{
			Name:       "total_balance_invariant",
			Passed:     false,
			Expression: "total_balance == 1000",
			ActualValues: map[string]interface{}{
				"total_balance": 900,
			},
		},
	}

	// AdyaGraph with Lost Update (WW and RW conflicts on same item)
	graph := analyzer.NewAdyaGraph()
	graph.AddEdge("T1", "T2", analyzer.DepWW, "accounts:1")
	graph.AddEdge("T2", "T1", analyzer.DepRW, "accounts:1")

	shrink := &domain.ShrinkResult{
		OriginalSize:   10,
		ReducedSize:    2,
		ReductionRatio: 80.0,
		Iterations:     3,
		Duration:       45 * time.Millisecond,
		MinimalOps: []domain.ScheduledOp{
			{
				ID:   1,
				Name: "withdraw",
				Steps: []domain.StepConfig{
					{SQL: "UPDATE accounts SET balance = balance - 100 WHERE id = 1;"},
				},
			},
			{
				ID:   2,
				Name: "deposit",
				Steps: []domain.StepConfig{
					{SQL: "UPDATE accounts SET balance = balance + 100 WHERE id = 1;"},
				},
			},
		},
	}

	sarifJSON, err := reporter.GenerateSARIFReport(spec, invResults, graph, shrink)
	if err != nil {
		t.Fatalf("GenerateSARIFReport returned unexpected error: %v", err)
	}

	if !json.Valid([]byte(sarifJSON)) {
		t.Fatalf("GenerateSARIFReport output is not valid JSON: %s", sarifJSON)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal([]byte(sarifJSON), &report); err != nil {
		t.Fatalf("Failed to unmarshal SARIF report into struct: %v", err)
	}

	// Assert SARIF version and schema
	if report.Version != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %q", report.Version)
	}
	if !strings.Contains(report.Schema, "sarif") || !strings.Contains(report.Schema, "2.1.0") {
		t.Errorf("expected SARIF $schema 2.1.0, got %q", report.Schema)
	}

	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
	run := report.Runs[0]

	// Assert Driver
	if run.Tool.Driver.Name != "ChaosSQL" && run.Tool.Driver.Name != "chaossql" {
		t.Errorf("expected driver name ChaosSQL or chaossql, got %q", run.Tool.Driver.Name)
	}

	// Assert Rules contain chaossql/P4-lost-update with level error
	var p4Rule *reporter.SarifRule
	for i := range run.Tool.Driver.Rules {
		if run.Tool.Driver.Rules[i].ID == "chaossql/P4-lost-update" {
			p4Rule = &run.Tool.Driver.Rules[i]
			break
		}
	}
	if p4Rule == nil {
		t.Fatalf("expected driver rules to contain chaossql/P4-lost-update")
	}
	if p4Rule.DefaultConfiguration.Level != "error" {
		t.Errorf("expected rule default level 'error', got %q", p4Rule.DefaultConfiguration.Level)
	}

	// Assert Results contain finding
	if len(run.Results) == 0 {
		t.Fatalf("expected results array to contain at least 1 finding")
	}
	res := run.Results[0]
	if res.RuleID != "chaossql/P4-lost-update" {
		t.Errorf("expected result RuleID 'chaossql/P4-lost-update', got %q", res.RuleID)
	}
	if res.Level != "error" {
		t.Errorf("expected result Level 'error', got %q", res.Level)
	}

	// Assert results point to spec artifact
	if len(res.Locations) == 0 {
		t.Fatalf("expected result to have locations pointing to spec artifact")
	}
	artifactURI := res.Locations[0].PhysicalLocation.ArtifactLocation.URI
	if !strings.Contains(artifactURI, spec.Name) && !strings.Contains(artifactURI, "chaos.yaml") {
		t.Errorf("expected location artifact URI to point to spec artifact, got %q", artifactURI)
	}
	if res.Locations[0].PhysicalLocation.Region.StartLine < 1 {
		t.Errorf("expected valid startLine >= 1, got %d", res.Locations[0].PhysicalLocation.Region.StartLine)
	}

	// Assert markdown message contains minimal causal reproducing operations
	if !strings.Contains(res.Message.Markdown, "withdraw") || !strings.Contains(res.Message.Markdown, "deposit") {
		t.Errorf("expected result markdown to contain minimal causal operations, got: %s", res.Message.Markdown)
	}
	if !strings.Contains(res.Message.Markdown, "total_balance_invariant") {
		t.Errorf("expected result markdown to mention failing invariant, got: %s", res.Message.Markdown)
	}
}

func TestGenerateSARIFReport_InvariantsSatisfied(t *testing.T) {
	spec := domain.Spec{
		Version: "1.0",
		Name:    "clean_run",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
	}

	invResults := []domain.InvariantResult{
		{
			Name:       "balance_preservation",
			Passed:     true,
			Expression: "balance == 1000",
			ActualValues: map[string]interface{}{
				"balance": 1000,
			},
		},
	}

	sarifJSON, err := reporter.GenerateSARIFReport(spec, invResults, nil, nil)
	if err != nil {
		t.Fatalf("GenerateSARIFReport returned unexpected error: %v", err)
	}

	if !json.Valid([]byte(sarifJSON)) {
		t.Fatalf("GenerateSARIFReport output is not valid JSON: %s", sarifJSON)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal([]byte(sarifJSON), &report); err != nil {
		t.Fatalf("Failed to unmarshal SARIF report: %v", err)
	}

	if report.Version != "2.1.0" {
		t.Errorf("expected version 2.1.0, got %q", report.Version)
	}
	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}

	// Invariants satisfied (clean run): output contains empty results array and valid SARIF structure
	if report.Runs[0].Results == nil {
		t.Errorf("expected results to be non-nil empty slice")
	}
	if len(report.Runs[0].Results) != 0 {
		t.Errorf("expected empty results array for clean run, got %d results", len(report.Runs[0].Results))
	}
	// Assert JSON literally contains `"results": []`
	if !strings.Contains(sarifJSON, `"results": []`) {
		t.Errorf("expected JSON to contain '\"results\": []', got: %s", sarifJSON)
	}
}

func TestGenerateSARIFReport_DeadlockCycle(t *testing.T) {
	spec := domain.Spec{
		Version: "1.0",
		Name:    "deadlock_cycle",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
	}

	// Deadlock lock contention cycle
	graph := analyzer.NewAdyaGraph()
	graph.AddEdge("T1", "T2", analyzer.DepWW, "bank_ledgers:1")
	graph.AddEdge("T2", "T1", analyzer.DepWW, "bank_ledgers:2")

	sarifJSON, err := reporter.GenerateSARIFReport(spec, nil, graph, nil)
	if err != nil {
		t.Fatalf("GenerateSARIFReport returned unexpected error: %v", err)
	}

	if !json.Valid([]byte(sarifJSON)) {
		t.Fatalf("GenerateSARIFReport output is not valid JSON: %s", sarifJSON)
	}

	var report reporter.SarifReport
	if err := json.Unmarshal([]byte(sarifJSON), &report); err != nil {
		t.Fatalf("Failed to unmarshal SARIF report: %v", err)
	}

	if len(report.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(report.Runs))
	}
	run := report.Runs[0]

	// Assert rule "chaossql/G-DL-deadlock" exists in driver rules
	var dlRule *reporter.SarifRule
	for i := range run.Tool.Driver.Rules {
		if run.Tool.Driver.Rules[i].ID == "chaossql/G-DL-deadlock" {
			dlRule = &run.Tool.Driver.Rules[i]
			break
		}
	}
	if dlRule == nil {
		t.Fatalf("expected driver rules to contain 'chaossql/G-DL-deadlock'")
	}
	if dlRule.DefaultConfiguration.Level != "warning" {
		t.Errorf("expected rule default level 'warning', got %q", dlRule.DefaultConfiguration.Level)
	}

	// Assert results contains finding with rule chaossql/G-DL-deadlock and level warning
	if len(run.Results) == 0 {
		t.Fatalf("expected at least 1 result for deadlock cycle")
	}
	res := run.Results[0]
	if res.RuleID != "chaossql/G-DL-deadlock" {
		t.Errorf("expected result RuleID 'chaossql/G-DL-deadlock', got %q", res.RuleID)
	}
	if res.Level != "warning" {
		t.Errorf("expected result Level 'warning', got %q", res.Level)
	}
	if len(res.Locations) == 0 {
		t.Fatalf("expected location for deadlock result")
	}
	if !strings.Contains(res.Locations[0].PhysicalLocation.ArtifactLocation.URI, "deadlock_cycle") {
		t.Errorf("expected location to point to deadlock_cycle spec, got %q", res.Locations[0].PhysicalLocation.ArtifactLocation.URI)
	}
}

func TestGenerateSARIFReport_OtherAnomalies(t *testing.T) {
	testCases := []struct {
		name          string
		specName      string
		edges         []struct {
			from, to string
			dep      analyzer.DependencyType
			item     string
			aborted  bool
		}
		expectedRule  string
		expectedLevel string
	}{
		{
			name:     "Write Skew A5B",
			specName: "hospital_write_skew",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepRW, "doctors:1", false},
				{"T2", "T1", analyzer.DepRW, "doctors:2", false},
			},
			expectedRule:  "chaossql/A5B-write-skew",
			expectedLevel: "error",
		},
		{
			name:     "Read Skew A5A",
			specName: "read_skew_financial_audit",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepRW, "accounts:1", false},
				{"T2", "T1", analyzer.DepWR, "accounts:2", false},
			},
			expectedRule:  "chaossql/A5A-read-skew",
			expectedLevel: "warning",
		},
		{
			name:     "Dirty Write G0",
			specName: "dirty_write_auction",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepWW, "lots:1", false},
				{"T2", "T1", analyzer.DepWW, "lots:1", false},
			},
			expectedRule:  "chaossql/G0-dirty-write",
			expectedLevel: "error",
		},
		{
			name:     "Dirty Read G1a",
			specName: "dirty_read_flash_crash",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepWR, "orders:1", true},
				{"T2", "T1", analyzer.DepWR, "orders:2", false},
			},
			expectedRule:  "chaossql/G1a-dirty-read",
			expectedLevel: "error",
		},
		{
			name:     "Circular Info G1c",
			specName: "circular_info_crypto_arbitrage",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepWR, "tokens:1", false},
				{"T2", "T1", analyzer.DepWR, "tokens:2", false},
			},
			expectedRule:  "chaossql/G1c-circular-info",
			expectedLevel: "error",
		},
		{
			name:     "Anti-Dependency G2",
			specName: "ticket_booking_anti_dependency",
			edges: []struct {
				from, to string
				dep      analyzer.DependencyType
				item     string
				aborted  bool
			}{
				{"T1", "T2", analyzer.DepRW, "seats:1", false},
				{"T2", "T3", analyzer.DepRW, "seats:2", false},
				{"T3", "T1", analyzer.DepRW, "seats:3", false},
			},
			expectedRule:  "chaossql/G2-anti-dependency",
			expectedLevel: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			spec := domain.Spec{
				Version: "1.0",
				Name:    tc.specName,
				Database: domain.DatabaseConfig{
					Driver: "sqlite",
				},
			}
			graph := analyzer.NewAdyaGraph()
			for _, e := range tc.edges {
				graph.AddEdgeWithAbort(e.from, e.to, e.dep, e.item, e.aborted)
			}
			invResults := []domain.InvariantResult{
				{
					Name:       "test_inv",
					Passed:     false,
					Expression: "x == 1",
				},
			}

			sarifJSON, err := reporter.GenerateSARIFReport(spec, invResults, graph, nil)
			if err != nil {
				t.Fatalf("GenerateSARIFReport failed: %v", err)
			}

			var report reporter.SarifReport
			if err := json.Unmarshal([]byte(sarifJSON), &report); err != nil {
				t.Fatalf("Failed to unmarshal SARIF: %v", err)
			}

			if len(report.Runs) == 0 || len(report.Runs[0].Results) == 0 {
				t.Fatalf("expected at least 1 result")
			}

			res := report.Runs[0].Results[0]
			if res.RuleID != tc.expectedRule {
				t.Errorf("expected rule %q, got %q", tc.expectedRule, res.RuleID)
			}
			if res.Level != tc.expectedLevel {
				t.Errorf("expected level %q, got %q", tc.expectedLevel, res.Level)
			}
		})
	}
}
