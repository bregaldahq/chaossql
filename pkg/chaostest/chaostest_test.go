package chaostest_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/drivers"
	"github.com/bregaldahq/chaossql/pkg/chaostest"
)

type spyTB struct {
	testing.TB
	fatalCalled bool
	fatalMsg    string
}

func (s *spyTB) Fatalf(format string, args ...any) {
	s.fatalCalled = true
	s.fatalMsg = fmt.Sprintf(format, args...)
}

func (s *spyTB) Helper() {}

func TestChaostest_AssertNoAnomalies_Pass(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("file:chaostest_pass?mode=memory&cache=shared")
	defer driver.Close()

	tester := chaostest.New(t).
		WithDriver(driver).
		WithSchema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);").
		WithSeed("INSERT INTO accounts (id, balance) VALUES (1, 1000), (2, 1000);").
		WithInvariant("total_balance", "SELECT SUM(balance) AS total FROM accounts;", "total == 2000").
		AddOperation("transfer_1_to_2",
			"UPDATE accounts SET balance = balance - 100 WHERE id = 1;",
			"UPDATE accounts SET balance = balance + 100 WHERE id = 2;",
		).
		AddOperation("transfer_2_to_1",
			"UPDATE accounts SET balance = balance - 50 WHERE id = 2;",
			"UPDATE accounts SET balance = balance + 50 WHERE id = 1;",
		)

	tester.AssertNoAnomalies(ctx, 4, 20, 42)
}

func TestChaostest_LostUpdate_RunAndShrink(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("file:chaostest_lost_update?mode=memory&cache=shared")
	defer driver.Close()

	tester := chaostest.New(t).
		WithDriver(driver).
		WithSchema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT); CREATE TABLE ledger (id INTEGER PRIMARY KEY AUTOINCREMENT, amount INT);").
		WithSeed("INSERT INTO accounts (id, balance) VALUES (1, 1000);").
		WithInvariant("ledger_consistency",
			"SELECT (SELECT balance FROM accounts WHERE id = 1) AS actual, (1000 - COALESCE(SUM(amount), 0)) AS expected FROM ledger;",
			"actual == expected",
		).
		AddOperation("withdraw_vulnerable",
			"SELECT balance FROM accounts WHERE id = 1; -> current_bal",
			"UPDATE accounts SET balance = {current_bal - 100} WHERE id = 1;",
			"INSERT INTO ledger (amount) VALUES (100);",
		)

	execRes, shrinkRes, err := tester.Run(ctx, 4, 20, 42)
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
	if execRes == nil {
		t.Fatalf("expected non-nil ExecutionResult")
	}
	if !execRes.ViolationDetected {
		t.Fatalf("expected ViolationDetected == true, got false")
	}
	if execRes.FailingInvariant == nil {
		t.Fatalf("expected non-nil FailingInvariant")
	}
	if execRes.FailingInvariant.Name != "ledger_consistency" {
		t.Errorf("expected FailingInvariant name 'ledger_consistency', got %q", execRes.FailingInvariant.Name)
	}
	if execRes.FailingInvariant.Passed {
		t.Errorf("expected FailingInvariant.Passed == false")
	}

	if shrinkRes == nil {
		t.Fatalf("expected non-nil ShrinkResult")
	}
	if shrinkRes.OriginalSize != 20 {
		t.Errorf("expected OriginalSize == 20, got %d", shrinkRes.OriginalSize)
	}
	if shrinkRes.ReducedSize > shrinkRes.OriginalSize {
		t.Errorf("expected ReducedSize <= OriginalSize, got %d > %d", shrinkRes.ReducedSize, shrinkRes.OriginalSize)
	}
	if shrinkRes.ReducedSize < 2 {
		t.Errorf("expected ReducedSize >= 2, got %d", shrinkRes.ReducedSize)
	}
	if len(shrinkRes.MinimalOps) != shrinkRes.ReducedSize {
		t.Errorf("expected len(MinimalOps) == %d, got %d", shrinkRes.ReducedSize, len(shrinkRes.MinimalOps))
	}
	if shrinkRes.ReductionRatio <= 0 {
		t.Errorf("expected positive ReductionRatio, got %f", shrinkRes.ReductionRatio)
	}
}

func TestChaostest_AssertNoAnomalies_ViolationReportsFatal(t *testing.T) {
	ctx := context.Background()
	driver := drivers.NewSQLiteDriver("file:chaostest_fatal?mode=memory&cache=shared")
	defer driver.Close()

	spy := &spyTB{TB: t}

	tester := chaostest.New(spy).
		WithDriver(driver).
		WithSchema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT); CREATE TABLE ledger (id INTEGER PRIMARY KEY AUTOINCREMENT, amount INT);").
		WithSeed("INSERT INTO accounts (id, balance) VALUES (1, 1000);").
		WithInvariant("ledger_consistency",
			"SELECT (SELECT balance FROM accounts WHERE id = 1) AS actual, (1000 - COALESCE(SUM(amount), 0)) AS expected FROM ledger;",
			"actual == expected",
		).
		AddOperation("withdraw_vulnerable",
			"SELECT balance FROM accounts WHERE id = 1; -> current_bal",
			"UPDATE accounts SET balance = {current_bal - 100} WHERE id = 1;",
			"INSERT INTO ledger (amount) VALUES (100);",
		)

	tester.AssertNoAnomalies(ctx, 4, 20, 42)

	if !spy.fatalCalled {
		t.Fatalf("expected Fatalf to be called when anomaly detected")
	}
	if !strings.Contains(spy.fatalMsg, "ledger_consistency") {
		t.Errorf("expected fatalMsg to mention failing invariant, got: %s", spy.fatalMsg)
	}
}

func TestChaostest_DefaultDriver(t *testing.T) {
	ctx := context.Background()

	tester := chaostest.New(t).
		WithSchema("CREATE TABLE counter (id INT PRIMARY KEY, val INT);").
		WithSeed("INSERT INTO counter (id, val) VALUES (1, 0);").
		WithInvariant("counter_non_negative", "SELECT val FROM counter WHERE id = 1;", "val >= 0").
		AddOperation("inc", "UPDATE counter SET val = val + 1 WHERE id = 1;")

	execRes, _, err := tester.Run(ctx, 2, 5, 1)
	if err != nil {
		t.Fatalf("unexpected Run error: %v", err)
	}
	if execRes == nil || execRes.ViolationDetected {
		t.Fatalf("expected valid successful run without violations")
	}
}
