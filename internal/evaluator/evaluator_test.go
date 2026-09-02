package evaluator_test

import (
	"context"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/internal/evaluator"
)

func TestEvaluator_PassAndFail(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()

	schema := "CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);"
	seed := "INSERT INTO accounts VALUES (1, 1000), (2, 500);"

	if err := driver.Reset(ctx, schema, seed); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	eval := evaluator.NewEvaluator()

	// 1. Test Passing Invariant
	passInv := domain.InvariantConfig{
		Name:   "total_sum_equals_1500",
		Query:  "SELECT SUM(balance) AS total FROM accounts;",
		Assert: "total == 1500",
	}
	resultPass, err := eval.Evaluate(ctx, driver, passInv)
	if err != nil || !resultPass.Passed {
		t.Errorf("expected pass, got %v (err: %v)", resultPass, err)
	}

	// 2. Test Failing Invariant
	failInv := domain.InvariantConfig{
		Name:   "total_sum_equals_2000",
		Query:  "SELECT SUM(balance) AS total FROM accounts;",
		Assert: "total == 2000",
	}
	resultFail, err := eval.Evaluate(ctx, driver, failInv)
	if err != nil || resultFail.Passed {
		t.Errorf("expected failure, got %v (err: %v)", resultFail, err)
	}
}
