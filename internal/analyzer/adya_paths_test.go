package analyzer

import (
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestExtractItem(t *testing.T) {
	cases := []struct {
		sql       string
		wantWrite bool
		wantItem  string
	}{
		{"UPDATE accounts SET balance = 0 WHERE id = 7;", true, "accounts:7"},
		{"INSERT INTO orders (id, total) VALUES (1, 10)", true, "orders"},
		{"DELETE FROM sessions WHERE id = 3;", true, "sessions:3"},
		{"DELETE FROM sessions", true, "sessions"},
		{"SELECT balance FROM accounts WHERE id = 2", false, "accounts:2"},
		{"SELECT * FROM accounts;", false, "accounts"},
		// Non-numeric ids fall back to table granularity.
		{"UPDATE accounts SET balance = 0 WHERE id = abc", true, "accounts"},
		// Leading whitespace skips the fast path and uses the regex fallback.
		{"  update ledger set v = 1 where id = 5", true, "ledger:5"},
		{"  select v from ledger where id = 9", false, "ledger:9"},
		{"BEGIN", false, ""},
	}
	for _, tc := range cases {
		isWrite, item := extractItem(tc.sql)
		if isWrite != tc.wantWrite || item != tc.wantItem {
			t.Errorf("extractItem(%q) = (%t, %q), want (%t, %q)", tc.sql, isWrite, item, tc.wantWrite, tc.wantItem)
		}
	}
}

func TestAddEdge_DeduplicatesIdenticalEdges(t *testing.T) {
	g := NewAdyaGraph()
	g.AddEdge("T1", "T2", DepWW, "accounts:1")
	g.AddEdge("T1", "T2", DepWW, "accounts:1")
	g.AddEdge("T1", "T2", DepRW, "accounts:1")
	if got := len(g.Edges["T1"]); got != 2 {
		t.Fatalf("got %d edges from T1, want 2 (duplicate WW dropped)", got)
	}
}

func TestBuildGraph_TableScanDependencies(t *testing.T) {
	trace := domain.ExecutionTrace{
		// T1 scans the whole table, T2 writes a row: RW anti-dependency on the scan.
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT * FROM accounts"},
		{WorkerID: 1, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT * FROM accounts"},
		{WorkerID: 2, OpIndex: 1, Type: domain.EventExec, SQL: "UPDATE accounts SET balance = 0 WHERE id = 1"},
		// T3 scans after T2 wrote a row: WR dependency on that row.
		{WorkerID: 3, OpIndex: 1, Type: domain.EventExec, SQL: "SELECT * FROM accounts"},
		{WorkerID: 3, OpIndex: 1, Type: domain.EventExec, SQL: "COMMIT"},
	}
	g := BuildGraph(trace)

	assertEdge(t, g, "T1-1", "T2-1", DepRW, "accounts:1")
	assertEdge(t, g, "T2-1", "T3-1", DepWR, "accounts:1")
}

func assertEdge(t *testing.T, g *AdyaGraph, from, to string, dep DependencyType, item string) {
	t.Helper()
	for _, e := range g.Edges[from] {
		if e.To == to && e.Type == dep && e.Item == item {
			return
		}
	}
	t.Fatalf("missing edge %s -%s-> %s on %s; edges from %s: %+v", from, dep, to, item, from, g.Edges[from])
}

func TestFindCycles_SelfLoop(t *testing.T) {
	g := NewAdyaGraph()
	g.AddNode("T1")
	g.AddEdge("T1", "T1", DepWW, "accounts:1")
	cycles := FindCycles(g)
	if len(cycles) != 1 || len(cycles[0]) != 1 || cycles[0][0].From != "T1" {
		t.Fatalf("cycles = %+v, want a single self-loop", cycles)
	}
}

func TestClassifyCycle(t *testing.T) {
	cases := []struct {
		name  string
		cycle Cycle
		want  domain.AnomalyType
	}{
		{"empty", nil, domain.AnomalyUnknown},
		{"aborted writer read", Cycle{{Type: DepWR, IsAbortedWriter: true}, {Type: DepWR}}, domain.AnomalyG1aDirtyRead},
		{"ww only", Cycle{{Type: DepWW}, {Type: DepWW}}, domain.AnomalyG0DirtyWrite},
		{"wr only", Cycle{{Type: DepWR}, {Type: DepWR}}, domain.AnomalyG1cCircularInfo},
		{"rw and wr", Cycle{{Type: DepRW}, {Type: DepWR}}, domain.AnomalyA5AReadSkew},
		{"ww and rw", Cycle{{Type: DepWW}, {Type: DepRW}}, domain.AnomalyLostUpdate},
		{"two rw", Cycle{{Type: DepRW}, {Type: DepRW}}, domain.AnomalyWriteSkew},
		{"three rw", Cycle{{Type: DepRW}, {Type: DepRW}, {Type: DepRW}}, domain.AnomalyG2AntiDependency},
		{"ww and wr", Cycle{{Type: DepWW}, {Type: DepWR}}, domain.AnomalyUnknown},
	}
	for _, tc := range cases {
		if got := ClassifyCycle(tc.cycle); got != tc.want {
			t.Errorf("%s: ClassifyCycle = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestRegister_RejectsUnknownOp(t *testing.T) {
	if _, err := CheckRegisterLinearizability([]RegisterEvent{{TxID: "T1", Var: "x", Op: "TRUNCATE"}}); err == nil {
		t.Fatal("expected unknown op to be rejected")
	}
}

func TestRegister_DirtyReadFromAbortedTransaction(t *testing.T) {
	result, err := CheckRegisterLinearizability([]RegisterEvent{
		{TxID: "T1", Var: "x", Op: OpWrite, Val: "5"},
		{TxID: "T2", Var: "x", Op: OpRead, Val: "5", ReadFromTx: "T1"},
		{TxID: "T1", Op: OpRollback},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Linearizable || len(result.Violations) != 1 || result.Violations[0].Type != domain.AnomalyG1aDirtyRead || result.Violations[0].OffendingTx != "T1" {
		t.Fatalf("result = %+v, want one G1a violation from T1", result)
	}
}

func TestRegister_IntermediateReadWithoutKnownWriter(t *testing.T) {
	result, err := CheckRegisterLinearizability([]RegisterEvent{
		{TxID: "T1", Var: "x", Op: OpWrite, Val: "1"},
		{TxID: "T1", Var: "x", Op: OpWrite, Val: "2"},
		{TxID: "T1", Op: OpCommit},
		{TxID: "T2", Var: "x", Op: OpRead, Val: "1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Linearizable || len(result.Violations) != 1 {
		t.Fatalf("result = %+v, want one violation", result)
	}
	v := result.Violations[0]
	if v.Type != domain.AnomalyG1bIntermediateRead || v.OffendingTx != "T1" || v.Expected != "2" || v.Actual != "1" {
		t.Fatalf("violation = %+v, want G1b from T1 (expected 2, actual 1)", v)
	}
}
