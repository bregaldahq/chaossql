package reporter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateStandaloneHTMLReport_AnomalyDetected(t *testing.T) {
	spec := domain.Spec{
		Version:     "1.0",
		Name:        "banking_lost_update",
		Description: "Detects concurrent Lost Update anomaly on accounts",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			Seed:   "INSERT INTO accounts VALUES (1, 1000);",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 20,
			Seed:       1337,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "ledger_balance_consistency",
				Query:  "SELECT balance FROM accounts WHERE id = 1;",
				Assert: "balance == 800",
			},
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 15 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 25 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
		},
		{
			Timestamp: 30 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 35 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 3,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
		},
		{
			Timestamp: 40 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 45 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 3,
			Type:      domain.EventCommit,
			SQL:       "COMMIT",
		},
	}

	graph := analyzer.NewAdyaGraph()
	graph.AddEdge("T1-1", "T2-2", analyzer.DepWW, "accounts:1")
	graph.AddEdge("T2-2", "T1-1", analyzer.DepRW, "accounts:1")

	shrink := &domain.ShrinkResult{
		OriginalSize:   20,
		ReducedSize:    2,
		ReductionRatio: 0.90,
		Iterations:     5,
		Duration:       25 * time.Millisecond,
		MinimalOps: []domain.ScheduledOp{
			{
				ID:   1,
				Name: "withdraw_vulnerable",
				Params: map[string]string{
					"amount": "100",
				},
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;", Capture: "current_bal"},
					{SQL: "UPDATE accounts SET balance = {current_bal - amount} WHERE id = 1;"},
				},
			},
			{
				ID:   2,
				Name: "withdraw_vulnerable",
				Params: map[string]string{
					"amount": "100",
				},
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;", Capture: "current_bal"},
					{SQL: "UPDATE accounts SET balance = {current_bal - amount} WHERE id = 1;"},
				},
			},
		},
	}

	invResults := []domain.InvariantResult{
		{
			Name:       "ledger_balance_consistency",
			Passed:     false,
			Expression: "balance == 800",
			ActualValues: map[string]interface{}{
				"balance": int64(900),
			},
		},
	}

	html := reporter.GenerateStandaloneHTMLReport(trace, spec, graph, shrink, invResults)
	if html == "" {
		t.Fatalf("expected non-empty HTML report")
	}

	// 1. DOCTYPE & Structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in report")
	}
	if !strings.Contains(html, "<html") || !strings.Contains(html, "</html>") {
		t.Errorf("expected valid html tags")
	}

	// 2. CSS Styles & Dark Mode
	if !strings.Contains(html, "Inter") {
		t.Errorf("expected Inter font in styles")
	}
	if !strings.Contains(html, "--accent-purple") || !strings.Contains(html, "--accent-cyan") {
		t.Errorf("expected purple/cyan neon accent variables in CSS")
	}

	// 3. Execution Summary Badge
	if !strings.Contains(html, "ISOLATION ANOMALY DETECTED") {
		t.Errorf("expected 'ISOLATION ANOMALY DETECTED' badge in report")
	}
	if !strings.Contains(html, "banking_lost_update") {
		t.Errorf("expected spec name 'banking_lost_update'")
	}
	if !strings.Contains(html, "sqlite") {
		t.Errorf("expected driver 'sqlite'")
	}

	// 4. SVG Serialization Graph
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected svg graph container")
	}
	if !strings.Contains(html, "T1-1") || !strings.Contains(html, "T2-2") {
		t.Errorf("expected transaction nodes T1-1 and T2-2 in graph")
	}
	if !strings.Contains(html, "WW") || !strings.Contains(html, "RW") {
		t.Errorf("expected edge dependency types WW and RW")
	}
	if !strings.Contains(html, "cycle") {
		t.Errorf("expected cycle highlighting in graph")
	}

	// 5. Invariant Audit Table
	if !strings.Contains(html, "ledger_balance_consistency") {
		t.Errorf("expected invariant name in table")
	}
	if !strings.Contains(html, "balance == 800") {
		t.Errorf("expected invariant assertion in table")
	}
	if !strings.Contains(html, "900") {
		t.Errorf("expected actual db state 900 in table")
	}
	if !strings.Contains(html, "FAIL") {
		t.Errorf("expected FAIL status in invariant table")
	}

	// 6. Delta-Debugging Shrink Card
	if !strings.Contains(html, "90.0%") && !strings.Contains(html, "90%") {
		t.Errorf("expected reduction percentage in shrink card")
	}
	if !strings.Contains(html, "ddmin") {
		t.Errorf("expected ddmin mention in shrink card")
	}
	if !strings.Contains(html, "withdraw_vulnerable") {
		t.Errorf("expected minimal op name in shrink card")
	}

	// 7. Chronological Timeline Swimlane
	if !strings.Contains(html, "Worker 1") || !strings.Contains(html, "Worker 2") {
		t.Errorf("expected worker names in timeline")
	}
	if !strings.Contains(html, "BEGIN") || !strings.Contains(html, "COMMIT") {
		t.Errorf("expected BEGIN/COMMIT events in timeline")
	}
	if !strings.Contains(html, "SELECT balance FROM accounts") {
		t.Errorf("expected SQL statements in timeline")
	}
}

func TestGenerateStandaloneHTMLReport_InvariantsSatisfied(t *testing.T) {
	spec := domain.Spec{
		Name: "hospital_safe_isolation",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
		Engine: domain.EngineConfig{
			Workers:    4,
			Iterations: 10,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "active_doctors_check",
				Query:  "SELECT count(*) as count FROM doctors WHERE on_call = true;",
				Assert: "count >= 1",
			},
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 5 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "shift_change",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT count(*) FROM doctors WHERE on_call = true;",
		},
	}

	invResults := []domain.InvariantResult{
		{
			Name:       "active_doctors_check",
			Passed:     true,
			Expression: "count >= 1",
			ActualValues: map[string]interface{}{
				"count": int64(2),
			},
		},
	}

	html := reporter.GenerateStandaloneHTMLReport(trace, spec, nil, nil, invResults)
	if html == "" {
		t.Fatalf("expected non-empty HTML report")
	}

	if !strings.Contains(html, "INVARIANTS SATISFIED") {
		t.Errorf("expected 'INVARIANTS SATISFIED' badge")
	}
	if !strings.Contains(html, "PASS") {
		t.Errorf("expected PASS status badge")
	}
	if !strings.Contains(html, "hospital_safe_isolation") {
		t.Errorf("expected spec name")
	}
}

func TestGenerateStandaloneHTMLReport_EmptyTraceAndNilInputs(t *testing.T) {
	spec := domain.Spec{Name: "empty_test"}
	html := reporter.GenerateStandaloneHTMLReport(nil, spec, nil, nil, nil)
	if html == "" {
		t.Fatalf("expected non-empty HTML output even for empty inputs")
	}
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE")
	}
	if !strings.Contains(html, "empty_test") {
		t.Errorf("expected spec name in report")
	}
}

func TestGenerateStandaloneHTMLReport_MultipleInvariantsAndGraphCycles(t *testing.T) {
	spec := domain.Spec{
		Name: "hospital_write_skew",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
		Engine: domain.EngineConfig{
			Workers: 2,
		},
	}

	graph := analyzer.NewAdyaGraph()
	// A5B Write Skew has RW cycles with no WW/WR
	graph.AddEdge("T1-1", "T2-1", analyzer.DepRW, "doctors:on_call")
	graph.AddEdge("T2-1", "T1-1", analyzer.DepRW, "doctors:on_call")

	invResults := []domain.InvariantResult{
		{
			Name:       "doctors_on_call_at_least_one",
			Passed:     false,
			Expression: "count >= 1",
			ActualValues: map[string]interface{}{
				"count": int64(0),
			},
		},
		{
			Name:       "department_open",
			Passed:     true,
			Expression: "open == true",
			ActualValues: map[string]interface{}{
				"open": true,
			},
		},
	}

	html := reporter.GenerateStandaloneHTMLReport(nil, spec, graph, nil, invResults)
	if html == "" {
		t.Fatalf("expected non-empty HTML output")
	}

	if !strings.Contains(html, "A5B_WRITE_SKEW") {
		t.Errorf("expected anomaly classification A5B_WRITE_SKEW in report")
	}
	if !strings.Contains(html, "doctors_on_call_at_least_one") || !strings.Contains(html, "department_open") {
		t.Errorf("expected both invariant names in audit table")
	}
	if !strings.Contains(html, "FAIL") || !strings.Contains(html, "PASS") {
		t.Errorf("expected both FAIL and PASS badges in audit table")
	}
}
