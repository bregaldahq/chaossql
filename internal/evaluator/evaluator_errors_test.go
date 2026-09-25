package evaluator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/evaluator"
)

func TestEvaluator_ReportsErrorsWithoutPassing(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	if err := driver.Reset(ctx, "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT, owner TEXT);", "INSERT INTO accounts VALUES (1, 1000, 'ana');"); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	cases := []struct {
		name    string
		inv     domain.InvariantConfig
		wantErr string
	}{
		{"query fails", domain.InvariantConfig{Query: "SELECT * FROM missing_table;", Assert: "true"}, "invariant query failed"},
		{"no rows", domain.InvariantConfig{Query: "SELECT balance FROM accounts WHERE id = 99;", Assert: "balance > 0"}, "returned no rows"},
		{"invalid expression", domain.InvariantConfig{Query: "SELECT balance FROM accounts;", Assert: "balance >"}, "invalid assert expression"},
		{"non boolean", domain.InvariantConfig{Query: "SELECT balance FROM accounts;", Assert: "balance + 1"}, "invalid assert expression"},
		{"runtime failure", domain.InvariantConfig{Query: "SELECT balance FROM accounts;", Assert: "[1, 2][balance] == 1"}, "expression evaluation failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.inv.Name = tc.name
			result, err := evaluator.NewEvaluator().Evaluate(ctx, driver, tc.inv)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want it to contain %q", err, tc.wantErr)
			}
			if result.Passed || result.Error != err {
				t.Fatalf("result = %+v, want a failed result carrying the returned error", result)
			}
		})
	}
}

func TestEvaluator_ComparesTextColumnsAsStrings(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	if err := driver.Reset(ctx, "CREATE TABLE accounts (id INT PRIMARY KEY, owner BLOB);", "INSERT INTO accounts VALUES (1, CAST('ana' AS BLOB));"); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	result, err := evaluator.NewEvaluator().Evaluate(ctx, driver, domain.InvariantConfig{
		Name:   "owner_is_ana",
		Query:  "SELECT owner FROM accounts WHERE id = 1;",
		Assert: `owner == "ana"`,
	})
	if err != nil || !result.Passed {
		t.Fatalf("result = %+v, err = %v; want []byte column compared as string", result, err)
	}
	if result.ActualValues["owner"] != "ana" {
		t.Fatalf("ActualValues[owner] = %#v, want \"ana\"", result.ActualValues["owner"])
	}
}

func TestEvaluateTemporalInvariants_DetectsOutOfOrderAndUnknownTypes(t *testing.T) {
	trace := domain.ExecutionTrace{
		{Timestamp: 20, Type: domain.EventBegin},
		{Timestamp: 10, Type: domain.EventCommit},
	}
	results := evaluator.EvaluateTemporalInvariants(trace, []domain.TemporalInvariantConfig{
		{Name: "ordered", Type: "monotonicity"},
		{Name: "bogus", Type: "eventually_consistent"},
	})
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results[0].Passed || results[0].ActualValues["monotonic"] != false {
		t.Fatalf("monotonicity result = %+v, want failure for decreasing timestamps", results[0])
	}
	if results[1].Passed || results[1].Error == nil || !strings.Contains(results[1].Error.Error(), "unknown temporal invariant type") {
		t.Fatalf("unknown type result = %+v, want an unknown-type error", results[1])
	}
}
