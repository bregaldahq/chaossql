package reporter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/analyzer"
	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateEmbeddedTraceViewerHTML_StructureAndDesignTokens(t *testing.T) {
	spec := domain.Spec{
		Name: "banking_lost_update",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 20,
			Seed:       1337,
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "transfer",
			StepIndex: 1,
			Type:      domain.EventBegin,
			SQL:       "BEGIN TRANSACTION;",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "transfer",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
		},
		{
			Timestamp: 30 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "transfer",
			StepIndex: 1,
			Type:      domain.EventBegin,
			SQL:       "BEGIN TRANSACTION;",
		},
		{
			Timestamp: 35 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "transfer",
			StepIndex: 3,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 40 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "transfer",
			StepIndex: 4,
			Type:      domain.EventCommit,
			SQL:       "COMMIT;",
		},
		{
			Timestamp: 50 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "transfer",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE accounts SET balance = 900 WHERE id = 1;",
		},
		{
			Timestamp: 60 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "transfer",
			StepIndex: 3,
			Type:      domain.EventCommit,
			SQL:       "COMMIT;",
		},
	}

	graph := analyzer.NewAdyaGraph()
	graph.AddEdge("T1-1", "T2-2", analyzer.DepWW, "accounts:1")
	graph.AddEdge("T2-2", "T1-1", analyzer.DepRW, "accounts:1")

	shrink := &domain.ShrinkResult{
		OriginalSize:   20,
		ReducedSize:    2,
		ReductionRatio: 0.90,
		Iterations:     7,
		Duration:       15 * time.Millisecond,
		MinimalOps: []domain.ScheduledOp{
			{
				ID:     1,
				Name:   "transfer",
				Params: map[string]string{"amount": "100"},
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;"},
					{SQL: "UPDATE accounts SET balance = 900 WHERE id = 1;"},
				},
			},
			{
				ID:     2,
				Name:   "transfer",
				Params: map[string]string{"amount": "100"},
				Steps: []domain.StepConfig{
					{SQL: "SELECT balance FROM accounts WHERE id = 1;"},
					{SQL: "UPDATE accounts SET balance = 900 WHERE id = 1;"},
				},
			},
		},
	}

	invResults := []domain.InvariantResult{
		{
			Name:       "total_balance_preserved",
			Passed:     false,
			Expression: "balance == 1000",
			ActualValues: map[string]interface{}{
				"balance": 900,
			},
		},
	}

	html := reporter.GenerateEmbeddedTraceViewerHTML(trace, spec, graph, shrink, invResults)
	if html == "" {
		t.Fatalf("expected non-empty HTML output")
	}

	// 1. DOCTYPE and HTML shell
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE in HTML")
	}
	if !strings.Contains(html, "<html") || !strings.Contains(html, "</html>") {
		t.Errorf("expected html root tags")
	}

	// 2. Bregalda Design Tokens
	tokens := []string{
		"#120E1F", // Deep Ink Canvas
		"#FCFBF8", // Warm Cream
		"#4B2E83", // Bregalda Purple
		"#F5C400", // Signal Yellow
		"#22C55E", // Ship Green
		"Inter",   // Typography
		"JetBrains Mono",
	}
	for _, tok := range tokens {
		if !strings.Contains(html, tok) {
			t.Errorf("expected Bregalda design token %q in HTML styles", tok)
		}
	}

	// 3. No external script or style dependencies (zero-external-dependency SPA)
	if strings.Contains(html, "<script src=\"http") || strings.Contains(html, "<link rel=\"stylesheet\" href=\"http") {
		t.Errorf("detected external script or style dependencies in embedded HTML")
	}

	// 4. Header with Bregalda branding & scenario summary
	if !strings.Contains(html, "ChaosSQL") {
		t.Errorf("expected ChaosSQL brand in header")
	}
	if !strings.Contains(html, "banking_lost_update") {
		t.Errorf("expected spec name in header summary")
	}
	if !strings.Contains(html, "sqlite") {
		t.Errorf("expected db driver in header summary")
	}

	// 5. Timeline Swimlane Container
	if !strings.Contains(html, "timeline-swimlane") && !strings.Contains(html, "timeline-container") {
		t.Errorf("expected timeline swimlane container")
	}
	if !strings.Contains(html, "Worker 1") || !strings.Contains(html, "Worker 2") {
		t.Errorf("expected worker labels in timeline")
	}

	// 6. Adya Graph SVG Container
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected SVG container for Adya dependency graph")
	}
	if !strings.Contains(html, "adya-graph") {
		t.Errorf("expected adya-graph container in HTML")
	}
	if !strings.Contains(html, "T1-1") || !strings.Contains(html, "T2-2") {
		t.Errorf("expected transaction nodes in graph")
	}
	if !strings.Contains(html, "WW") || !strings.Contains(html, "RW") {
		t.Errorf("expected dependency edge labels WW and RW")
	}

	// 7. Delta-Debugging Comparison Card
	if !strings.Contains(html, "delta-debugging") && !strings.Contains(html, "shrink-card") {
		t.Errorf("expected delta-debugging / shrink card container")
	}
	if !strings.Contains(html, "90.0%") && !strings.Contains(html, "90%") {
		t.Errorf("expected reduction ratio 90%% in shrink comparison card")
	}

	// 8. Statement Detail Inspector Table
	if !strings.Contains(html, "statement-inspector") {
		t.Errorf("expected statement-inspector container")
	}
	if !strings.Contains(html, "SELECT balance FROM accounts") {
		t.Errorf("expected SQL statement in statement inspector")
	}
	if !strings.Contains(html, "UPDATE accounts SET balance = 900") {
		t.Errorf("expected UPDATE SQL statement in statement inspector")
	}
}

func TestGenerateEmbeddedTraceViewerHTML_HandlesEmptyAndNilGracefully(t *testing.T) {
	spec := domain.Spec{Name: "empty_scenario"}

	html := reporter.GenerateEmbeddedTraceViewerHTML(nil, spec, nil, nil, nil)
	if html == "" {
		t.Fatalf("expected non-empty HTML even with nil/empty inputs")
	}

	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected DOCTYPE")
	}
	if !strings.Contains(html, "empty_scenario") {
		t.Errorf("expected spec name in empty scenario HTML")
	}
	if !strings.Contains(html, "#120E1F") {
		t.Errorf("expected Bregalda canvas color even in empty report")
	}
	if !strings.Contains(html, "timeline") {
		t.Errorf("expected timeline container in empty report")
	}
	if !strings.Contains(html, "adya-graph") {
		t.Errorf("expected adya graph container in empty report")
	}
	if !strings.Contains(html, "statement-inspector") {
		t.Errorf("expected statement inspector in empty report")
	}
}

func TestGenerateEmbeddedTraceViewerHTML_CycleHighlighting(t *testing.T) {
	spec := domain.Spec{
		Name: "hospital_write_skew",
		Database: domain.DatabaseConfig{
			Driver: "postgres",
		},
		Engine: domain.EngineConfig{
			Workers: 2,
		},
	}

	trace := domain.ExecutionTrace{
		{
			Timestamp: 5 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "leave_shift",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT count(*) FROM doctors WHERE on_call = true;",
		},
		{
			Timestamp: 10 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "leave_shift",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT count(*) FROM doctors WHERE on_call = true;",
		},
		{
			Timestamp: 15 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "leave_shift",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE doctors SET on_call = false WHERE id = 1;",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "leave_shift",
			StepIndex: 2,
			Type:      domain.EventExec,
			SQL:       "UPDATE doctors SET on_call = false WHERE id = 2;",
		},
	}

	graph := analyzer.NewAdyaGraph()
	graph.AddEdge("T1-1", "T2-2", analyzer.DepRW, "doctors")
	graph.AddEdge("T2-2", "T1-1", analyzer.DepRW, "doctors")

	invResults := []domain.InvariantResult{
		{
			Name:       "at_least_one_doctor",
			Passed:     false,
			Expression: "count >= 1",
			ActualValues: map[string]interface{}{
				"count": 0,
			},
		},
	}

	html := reporter.GenerateEmbeddedTraceViewerHTML(trace, spec, graph, nil, invResults)
	if html == "" {
		t.Fatalf("expected non-empty HTML")
	}

	// Cycle should be classified and highlighted
	if !strings.Contains(html, "A5B_WRITE_SKEW") && !strings.Contains(html, "WRITE_SKEW") {
		t.Errorf("expected write skew classification badge in HTML")
	}
	if !strings.Contains(html, "cycle") {
		t.Errorf("expected cycle indicator/highlighting in SVG graph")
	}
}
