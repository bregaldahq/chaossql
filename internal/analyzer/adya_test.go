package analyzer

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestAdyaLostUpdate(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 1"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyLostUpdate {
		t.Errorf("Expected %v, got %v", domain.AnomalyLostUpdate, anomaly)
	}
}

func TestAdyaWriteSkew(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 2"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 2"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 1"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyWriteSkew {
		t.Errorf("Expected %v, got %v", domain.AnomalyWriteSkew, anomaly)
	}
}

func TestAdyaReadSkew(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 2"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 2"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyA5AReadSkew {
		t.Errorf("Expected %v, got %v", domain.AnomalyA5AReadSkew, anomaly)
	}
}

func TestAdyaDirtyWrite(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 30 WHERE id = 2"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 40 WHERE id = 2"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyG0DirtyWrite {
		t.Errorf("Expected %v, got %v", domain.AnomalyG0DirtyWrite, anomaly)
	}
}

func TestAdyaDirtyRead_G1a(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventRollback, SQL: "ROLLBACK"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 2"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 2"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyG1aDirtyRead {
		t.Errorf("Expected %v, got %v", domain.AnomalyG1aDirtyRead, anomaly)
	}
}

func TestAdyaCircularInfo(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 2"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 2"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) == 0 {
		t.Fatalf("Expected cycle, got none")
	}

	anomaly := ClassifyCycle(cycles[0])
	if anomaly != domain.AnomalyG1cCircularInfo {
		t.Errorf("Expected %v, got %v", domain.AnomalyG1cCircularInfo, anomaly)
	}
}

func TestAdyaAcyclic(t *testing.T) {
	trace := domain.ExecutionTrace{
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 10 WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT balance FROM accounts WHERE id = 1"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 20 WHERE id = 1"},
	}

	g := BuildGraph(trace)
	cycles := FindCycles(g)

	if len(cycles) > 0 {
		t.Fatalf("Expected no cycles, got %d", len(cycles))
	}
}

func TestClassifyCycleDirect(t *testing.T) {
	tests := []struct {
		name     string
		cycle    Cycle
		expected domain.AnomalyType
	}{
		{
			name: "Aborted Writer (G1a Dirty Read)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepWR, Item: "accounts:1", IsAbortedWriter: true},
				{From: "T2", To: "T1", Type: DepWR, Item: "accounts:2"},
			},
			expected: domain.AnomalyG1aDirtyRead,
		},
		{
			name: "Only WW (G0 Dirty Write)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepWW, Item: "accounts:1"},
				{From: "T2", To: "T1", Type: DepWW, Item: "accounts:2"},
			},
			expected: domain.AnomalyG0DirtyWrite,
		},
		{
			name: "Only WR (G1c Circular Info)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepWR, Item: "accounts:1"},
				{From: "T2", To: "T1", Type: DepWR, Item: "accounts:2"},
			},
			expected: domain.AnomalyG1cCircularInfo,
		},
		{
			name: "RW and WR (A5A Read Skew)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepRW, Item: "accounts:1"},
				{From: "T2", To: "T1", Type: DepWR, Item: "accounts:2"},
			},
			expected: domain.AnomalyA5AReadSkew,
		},
		{
			name: "WW and RW (P4 Lost Update)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepWW, Item: "accounts:1"},
				{From: "T2", To: "T1", Type: DepRW, Item: "accounts:1"},
			},
			expected: domain.AnomalyLostUpdate,
		},
		{
			name: "Only RW (A5B Write Skew)",
			cycle: Cycle{
				{From: "T1", To: "T2", Type: DepRW, Item: "accounts:1"},
				{From: "T2", To: "T1", Type: DepRW, Item: "accounts:2"},
			},
			expected: domain.AnomalyWriteSkew,
		},
		{
			name:     "Empty Cycle (Unknown)",
			cycle:    Cycle{},
			expected: domain.AnomalyUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyCycle(tt.cycle)
			if got != tt.expected {
				t.Errorf("ClassifyCycle() = %v, want %v", got, tt.expected)
			}
		})
	}
}
