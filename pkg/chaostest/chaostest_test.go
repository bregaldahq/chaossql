package chaostest_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/bregaldahq/chaossql/internal/domain"
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

func TestChaostest_InvariantViolation_RunAndShrink(t *testing.T) {
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
		AddOperation("withdraw_without_ledger_entry",
			"UPDATE accounts SET balance = balance - 100 WHERE id = 1;",
		)

	execRes, shrinkRes, err := tester.Run(ctx, 1, 20, 42)
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
	if shrinkRes.ReducedSize < 1 {
		t.Errorf("expected ReducedSize >= 1, got %d", shrinkRes.ReducedSize)
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
		AddOperation("withdraw_without_ledger_entry",
			"UPDATE accounts SET balance = balance - 100 WHERE id = 1;",
		)

	tester.AssertNoAnomalies(ctx, 1, 20, 42)

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

func TestChaostest_RunReturnsExecutionError(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	tester := chaostest.New(t).
		WithDriver(driver).
		WithSchema("CREATE TABLE items (id INT PRIMARY KEY);").
		WithSeed("INSERT INTO items VALUES (1);").
		WithInvariant("must_not_run", "SELECT value FROM missing_table", "value == 1").
		AddOperation("invalid", "NOT VALID SQL")

	result, shrink, err := tester.Run(context.Background(), 1, 1, 42)
	if err == nil || result == nil || result.Status != domain.StatusExecutionError || shrink != nil {
		t.Fatalf("result=%+v shrink=%+v error=%v", result, shrink, err)
	}
}

func TestChaostest_RunWithParamsAndCapture(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	tester := chaostest.New(t).
		WithDriver(driver).
		WithJitter(0, 1).
		WithSchema("CREATE TABLE bids (id INTEGER PRIMARY KEY AUTOINCREMENT, amount INT);").
		WithInvariant("bids_in_range", "SELECT COUNT(*) AS bad FROM bids WHERE amount < 10 OR amount > 20;", "bad == 0").
		AddOperationWithParams("bid", map[string]string{"amount": "int(10, 20)"},
			"SELECT COUNT(*) FROM bids -> seen",
			"INSERT INTO bids (amount) VALUES ({amount});",
		)

	result, shrink, err := tester.Run(context.Background(), 2, 10, 7)
	if err != nil || result == nil || result.ViolationDetected || shrink != nil {
		t.Fatalf("result=%+v shrink=%+v err=%v", result, shrink, err)
	}
}

func TestChaostest_RunReportsDriverOpenFailure(t *testing.T) {
	driver, err := drivers.GetDriver("postgres", "postgres://nobody@127.0.0.1:1/none?sslmode=disable&connect_timeout=2")
	if err != nil {
		t.Fatal(err)
	}
	defer driver.Close()
	_, _, err = chaostest.New(t).WithDriver(driver).Run(context.Background(), 1, 1, 1)
	if err == nil || !strings.Contains(err.Error(), "failed to open database driver") {
		t.Fatalf("err = %v, want an open failure", err)
	}
}

func TestChaostest_AssertNoAnomalies_ReportsExecutionError(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	spy := &spyTB{TB: t}
	chaostest.New(spy).
		WithDriver(driver).
		WithSchema("CREATE TABLE items (id INT PRIMARY KEY);").
		WithInvariant("must_not_run", "SELECT 1 AS one", "one == 1").
		AddOperation("invalid", "NOT VALID SQL").
		AssertNoAnomalies(context.Background(), 1, 1, 42)

	if !spy.fatalCalled || !strings.Contains(spy.fatalMsg, "execution failed") {
		t.Fatalf("fatalCalled=%t msg=%q, want an execution failure report", spy.fatalCalled, spy.fatalMsg)
	}
}

func TestChaostest_AssertNoAnomalies_ReportsCapturedSteps(t *testing.T) {
	driver := drivers.NewSQLiteDriver("file:chaostest_capture?mode=memory&cache=shared")
	defer driver.Close()
	spy := &spyTB{TB: t}
	chaostest.New(spy).
		WithDriver(driver).
		WithSchema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT);").
		WithSeed("INSERT INTO accounts VALUES (1, 100);").
		WithInvariant("balance_unchanged", "SELECT balance FROM accounts WHERE id = 1;", "balance == 100").
		AddOperation("drain",
			"SELECT balance FROM accounts WHERE id = 1; -> bal",
			"UPDATE accounts SET balance = {bal} - 1 WHERE id = 1;",
		).
		AssertNoAnomalies(context.Background(), 1, 3, 42)

	if !spy.fatalCalled || !strings.Contains(spy.fatalMsg, "(capture: bal)") {
		t.Fatalf("fatalCalled=%t msg=%q, want the minimal schedule with its capture variable", spy.fatalCalled, spy.fatalMsg)
	}
}

func TestChaostest_RunHonoursCancelledContext(t *testing.T) {
	driver := drivers.NewSQLiteDriver("")
	defer driver.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := chaostest.New(t).
		WithDriver(driver).
		WithSchema("CREATE TABLE t (id INT PRIMARY KEY, v INT);").
		WithSeed("INSERT INTO t VALUES (1, 0);").
		WithInvariant("zero", "SELECT v FROM t WHERE id = 1", "v == 0").
		AddOperation("bump", "UPDATE t SET v = v + 1 WHERE id = 1").
		Run(ctx, 1, 5, 3)
	if err == nil {
		t.Fatal("a run with a cancelled context must return an error")
	}
}
