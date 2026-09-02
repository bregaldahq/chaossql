package reporter_test

import (
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/engine"
	"github.com/bregaldahq/chaossql/internal/reporter"
)

func TestGenerateMermaidSequence_Valid(t *testing.T) {
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
			WorkerID:  2,
			OpIndex:   2,
			OpName:    "withdraw_vulnerable",
			StepIndex: 0,
			Type:      domain.EventBegin,
			SQL:       "BEGIN",
		},
		{
			Timestamp: 20 * time.Millisecond,
			WorkerID:  1,
			OpIndex:   1,
			OpName:    "withdraw_vulnerable",
			StepIndex: 1,
			Type:      domain.EventExec,
			SQL:       "SELECT balance FROM accounts WHERE id = 1;",
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

	mermaid := reporter.GenerateMermaidSequence(trace)
	if mermaid == "" {
		t.Fatalf("expected non-empty mermaid output")
	}

	if !strings.HasPrefix(strings.TrimSpace(mermaid), "sequenceDiagram") {
		t.Errorf("expected sequenceDiagram header, got: %s", mermaid)
	}

	if !strings.Contains(mermaid, "participant W1 as Worker 1") {
		t.Errorf("expected participant W1, got: %s", mermaid)
	}
	if !strings.Contains(mermaid, "participant W2 as Worker 2") {
		t.Errorf("expected participant W2, got: %s", mermaid)
	}
	if !strings.Contains(mermaid, "participant DB as Database") {
		t.Errorf("expected participant DB, got: %s", mermaid)
	}

	if !strings.Contains(mermaid, "W1->>DB: [withdraw_vulnerable #1] BEGIN") {
		t.Errorf("expected W1 BEGIN event in diagram, got: %s", mermaid)
	}
	if !strings.Contains(mermaid, "SELECT balance FROM accounts WHERE id = 1;") {
		t.Errorf("expected SQL query in diagram, got: %s", mermaid)
	}
	if !strings.Contains(mermaid, "UPDATE accounts SET balance = 900 WHERE id = 1;") {
		t.Errorf("expected UPDATE query in diagram, got: %s", mermaid)
	}
}

func TestGenerateMermaidSequence_EmptyTrace(t *testing.T) {
	mermaid := reporter.GenerateMermaidSequence(domain.ExecutionTrace{})
	if !strings.HasPrefix(strings.TrimSpace(mermaid), "sequenceDiagram") {
		t.Errorf("expected valid sequenceDiagram for empty trace, got: %s", mermaid)
	}
}

func TestGenerateStandaloneGoRepro_ValidSyntax(t *testing.T) {
	spec := domain.Spec{
		Version:     "1.0",
		Name:        "banking_lost_update",
		Description: "Detects Lost Update (P4)",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);",
			Seed:   "INSERT INTO accounts VALUES (1, 1000);",
		},
		Engine: domain.EngineConfig{
			Workers:    2,
			Iterations: 2,
			Seed:       42,
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "ledger_balance_consistency",
				Query:  "SELECT balance FROM accounts WHERE id = 1;",
				Assert: "balance == 800",
			},
		},
	}

	ops := []domain.ScheduledOp{
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
	}

	failingInv := &domain.InvariantResult{
		Name:       "ledger_balance_consistency",
		Passed:     false,
		Expression: "balance == 800",
		ActualValues: map[string]interface{}{
			"balance": int64(900),
		},
	}

	code := reporter.GenerateStandaloneGoRepro(spec, ops, failingInv)
	if code == "" {
		t.Fatalf("expected non-empty generated Go repro code")
	}

	// Verify the generated Go code parses successfully
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "repro_test.go", code, parser.AllErrors)
	if err != nil {
		t.Fatalf("generated Go code has syntax errors: %v\nCode:\n%s", err, code)
	}
	if node == nil {
		t.Fatalf("parsed AST node is nil")
	}

	if !strings.Contains(code, "TestReproduceAnomaly") {
		t.Errorf("expected TestReproduceAnomaly in generated code")
	}
	if !strings.Contains(code, "CREATE TABLE accounts") {
		t.Errorf("expected schema SQL in generated code")
	}
	if !strings.Contains(code, "INSERT INTO accounts") {
		t.Errorf("expected seed SQL in generated code")
	}
}

func TestGenerateStandaloneGoRepro_ExecutesWithGoTest(t *testing.T) {
	spec := domain.Spec{
		Version: "1.0",
		Name:    "banking_lost_update",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
			Schema: "CREATE TABLE accounts (id INTEGER PRIMARY KEY, balance INTEGER);",
			Seed:   "INSERT INTO accounts (id, balance) VALUES (1, 1000);",
		},
		Invariants: []domain.InvariantConfig{
			{
				Name:   "balance_check",
				Query:  "SELECT balance FROM accounts WHERE id = 1;",
				Assert: "balance == 800",
			},
		},
	}

	ops := []domain.ScheduledOp{
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
	}

	failingInv := &domain.InvariantResult{
		Name:       "balance_check",
		Passed:     false,
		Expression: "balance == 800",
		ActualValues: map[string]interface{}{
			"balance": int64(900),
		},
	}

	code := reporter.GenerateStandaloneGoRepro(spec, ops, failingInv)

	tmpDir, err := os.MkdirTemp("", "chaossql_repro_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy root go.mod and go.sum to allow resolution
	rootMod, err := os.ReadFile("../../go.mod")
	if err == nil {
		_ = os.WriteFile(filepath.Join(tmpDir, "go.mod"), rootMod, 0644)
	}
	rootSum, err := os.ReadFile("../../go.sum")
	if err == nil {
		_ = os.WriteFile(filepath.Join(tmpDir, "go.sum"), rootSum, 0644)
	}

	testFilePath := filepath.Join(tmpDir, "repro_test.go")
	if err := os.WriteFile(testFilePath, []byte(code), 0644); err != nil {
		t.Fatalf("failed to write repro_test.go: %v", err)
	}

	cmd := exec.Command("/usr/local/go/bin/go", "test", "-v", ".")
	cmd.Dir = tmpDir
	out, _ := cmd.CombinedOutput()
	outStr := string(out)

	if !strings.Contains(outStr, "REPRODUCED ANOMALY") && !strings.Contains(outStr, "Invariant") {
		t.Errorf("expected repro test execution output to mention anomaly or invariant, got:\n%s", outStr)
	}
}

func TestTerminal_Renderers(t *testing.T) {
	banner := reporter.RenderBanner()
	if banner == "" {
		t.Errorf("expected non-empty banner")
	}

	spec := domain.Spec{
		Name: "banking_lost_update",
		Database: domain.DatabaseConfig{
			Driver: "sqlite",
		},
		Engine: domain.EngineConfig{
			Workers:    4,
			Iterations: 20,
			Seed:       42,
		},
	}

	invRes := domain.InvariantResult{
		Name:       "ledger_balance_consistency",
		Passed:     false,
		Expression: "actual_balance == expected_balance",
		ActualValues: map[string]interface{}{
			"actual_balance":   900,
			"expected_balance": 800,
		},
	}

	runRes := &engine.RunResult{
		Success:           false,
		ViolationDetected: true,
		FailingInvariant:  &invRes,
		Duration:          50 * time.Millisecond,
	}

	summary := reporter.RenderRunSummary(spec, runRes, domain.AnomalyLostUpdate)
	if summary == "" || !strings.Contains(summary, "banking_lost_update") {
		t.Errorf("expected run summary to contain spec name, got: %s", summary)
	}

	table := reporter.RenderInvariantTable([]domain.InvariantResult{invRes})
	if table == "" || !strings.Contains(table, "ledger_balance_consistency") {
		t.Errorf("expected invariant table to contain invariant name, got: %s", table)
	}

	shrink := &domain.ShrinkResult{
		OriginalSize:   20,
		ReducedSize:    2,
		ReductionRatio: 0.90,
		Iterations:     4,
		Duration:       15 * time.Millisecond,
		MinimalOps: []domain.ScheduledOp{
			{ID: 1, Name: "withdraw_vulnerable", Params: map[string]string{"amount": "100"}},
			{ID: 2, Name: "withdraw_vulnerable", Params: map[string]string{"amount": "100"}},
		},
	}

	shrinkSummary := reporter.RenderShrinkSummary(shrink)
	if shrinkSummary == "" || !strings.Contains(shrinkSummary, "90") {
		t.Errorf("expected shrink summary to contain reduction, got: %s", shrinkSummary)
	}

	fullReport := reporter.RenderFullReport(spec, runRes, shrink, domain.AnomalyLostUpdate)
	if fullReport == "" || !strings.Contains(fullReport, "ChaosSQL") {
		t.Errorf("expected full report to contain ChaosSQL, got: %s", fullReport)
	}
}
