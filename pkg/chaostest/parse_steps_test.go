package chaostest

import (
	"reflect"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
)

func TestParseSteps(t *testing.T) {
	got := parseSteps([]string{
		"  SELECT balance FROM accounts WHERE id = 1; -> current  ",
		"SELECT stock FROM items => stock",
		"UPDATE accounts SET balance = {current} - 10 WHERE id = 1;",
	})
	want := []domain.StepConfig{
		{SQL: "SELECT balance FROM accounts WHERE id = 1;", Capture: "current"},
		{SQL: "SELECT stock FROM items", Capture: "stock"},
		{SQL: "UPDATE accounts SET balance = {current} - 10 WHERE id = 1;"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseSteps =\n%#v\nwant\n%#v", got, want)
	}
}

func TestBuilderOptions(t *testing.T) {
	tr := New(nil).
		WithJitter(0, 2).
		AddOperationWithParams("bid", map[string]string{"amount": "int(1, 9)"}, "INSERT INTO bids VALUES ({amount})")
	if tr.jitterMs != [2]int{0, 2} {
		t.Fatalf("jitterMs = %v, want [0 2]", tr.jitterMs)
	}
	op := tr.operations[0]
	if op.Name != "bid" || op.Weight != 1.0 || op.Params["amount"] != "int(1, 9)" || op.Steps[0].SQL != "INSERT INTO bids VALUES ({amount})" {
		t.Fatalf("operation = %+v", op)
	}
}

func TestUnshrunkResultKeepsOriginalSchedule(t *testing.T) {
	ops := []domain.ScheduledOp{{ID: 1}, {ID: 2}}
	plain := unshrunkResult(ops, nil)
	if plain.OriginalSize != 2 || plain.ReducedSize != 2 || plain.ReductionRatio != 0 || len(plain.MinimalOps) != 2 || plain.Iterations != 0 {
		t.Fatalf("unshrunkResult(nil) = %+v", plain)
	}
	withAttempt := unshrunkResult(ops, &domain.ShrinkResult{Iterations: 4, Trials: 9, MinimalOps: ops[:1]})
	if withAttempt.Iterations != 4 || withAttempt.Trials != 9 || len(withAttempt.MinimalOps) != 2 {
		t.Fatalf("unshrunkResult(attempt) = %+v, want attempt statistics with the full schedule", withAttempt)
	}
}
