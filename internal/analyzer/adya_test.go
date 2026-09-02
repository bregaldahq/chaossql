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
