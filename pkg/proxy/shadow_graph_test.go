package proxy

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestShadowGraph_LostUpdate(t *testing.T) {
	engine := NewShadowGraphEngine()

	s1 := NewProxySession(1, nil, nil)
	s2 := NewProxySession(2, nil, nil)

	engine.StartTx(s1, domain.LevelReadCommitted)
	engine.StartTx(s2, domain.LevelReadCommitted)

	// S1 reads accounts:1
	engine.RecordRead(s1, []string{"accounts:1"}, "SELECT balance FROM accounts WHERE id = 1")

	// S2 reads accounts:1
	engine.RecordRead(s2, []string{"accounts:1"}, "SELECT balance FROM accounts WHERE id = 1")

	// S2 updates accounts:1 and commits
	engine.RecordWrite(s2, []string{"accounts:1"}, "UPDATE accounts SET balance = balance - 50 WHERE id = 1")
	anomalies2 := engine.RecordCommit(s2)
	if len(anomalies2) != 0 {
		t.Fatalf("did not expect anomaly on S2 commit yet, got %d", len(anomalies2))
	}

	// S1 updates accounts:1 (overwriting S2's write based on stale read) and commits
	engine.RecordWrite(s1, []string{"accounts:1"}, "UPDATE accounts SET balance = balance - 100 WHERE id = 1")
	anomalies1 := engine.RecordCommit(s1)

	if len(anomalies1) == 0 {
		t.Fatalf("expected Lost Update (P4) anomaly detected on S1 commit, got 0")
	}

	foundP4 := false
	for _, a := range anomalies1 {
		if a.Type == domain.AnomalyLostUpdate {
			foundP4 = true
			break
		}
	}
	if !foundP4 {
		t.Fatalf("expected AnomalyLostUpdate, got %v", anomalies1[0].Type)
	}
}

func TestShadowGraph_WriteSkew(t *testing.T) {
	engine := NewShadowGraphEngine()

	s1 := NewProxySession(1, nil, nil)
	s2 := NewProxySession(2, nil, nil)

	engine.StartTx(s1, domain.LevelRepeatableRead)
	engine.StartTx(s2, domain.LevelRepeatableRead)

	// Both read doctors table
	engine.RecordRead(s1, []string{"doctors"}, "SELECT * FROM doctors WHERE on_call = true")
	engine.RecordRead(s2, []string{"doctors"}, "SELECT * FROM doctors WHERE on_call = true")

	// S1 updates doctor 1
	engine.RecordWrite(s1, []string{"doctors:1"}, "UPDATE doctors SET on_call = false WHERE id = 1")
	_ = engine.RecordCommit(s1)

	// S2 updates doctor 2
	engine.RecordWrite(s2, []string{"doctors:2"}, "UPDATE doctors SET on_call = false WHERE id = 2")
	anomalies := engine.RecordCommit(s2)

	if len(anomalies) == 0 {
		t.Fatalf("expected Write Skew (A5B) anomaly detected, got 0")
	}

	foundA5B := false
	for _, a := range anomalies {
		if a.Type == domain.AnomalyWriteSkew {
			foundA5B = true
			break
		}
	}
	if !foundA5B {
		t.Fatalf("expected AnomalyWriteSkew, got %v", anomalies[0].Type)
	}
}

func TestShadowGraph_DirtyRead(t *testing.T) {
	engine := NewShadowGraphEngine()

	s1 := NewProxySession(1, nil, nil)
	s2 := NewProxySession(2, nil, nil)

	engine.StartTx(s1, domain.LevelReadUncommitted)
	engine.StartTx(s2, domain.LevelReadUncommitted)

	// S1 writes uncommitted price
	engine.RecordWrite(s1, []string{"stocks:1"}, "UPDATE stocks SET price = 10 WHERE id = 1")

	// S2 reads uncommitted write
	engine.RecordRead(s2, []string{"stocks:1"}, "SELECT price FROM stocks WHERE id = 1")

	// S1 aborts / rolls back
	anomalies := engine.RecordRollback(s1)

	if len(anomalies) == 0 {
		t.Fatalf("expected Dirty Read (G1a) anomaly detected on rollback, got 0")
	}

	if anomalies[0].Type != domain.AnomalyG1aDirtyRead {
		t.Fatalf("expected AnomalyG1aDirtyRead, got %v", anomalies[0].Type)
	}
}
