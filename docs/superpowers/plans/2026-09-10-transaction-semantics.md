# Transaction Semantics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Execute every scheduled operation in a real SQL transaction and report trustworthy isolation, cancellation, invariant, and execution outcomes.

**Architecture:** The domain defines portable isolation and result contracts. Database adapters validate and translate transaction options. The runner delegates each operation to one transaction-owning function and finalizes all outcomes through one status function.

**Tech Stack:** Go 1.25, `database/sql`, modernc SQLite, pgx stdlib, go-sql-driver/mysql, YAML specifications, GitHub Actions database services.

**Spec:** `docs/superpowers/specs/2026-09-10-transaction-semantics-design.md`

## Global Constraints

- The same operation must use exactly one `drivers.Tx` from begin through commit or rollback.
- The domain must not import database adapters.
- Existing YAML remains valid when `database.isolation` is absent.
- Existing JSON fields remain available; new status fields add safer semantics.
- PostgreSQL must reject `READ_UNCOMMITTED` instead of claiming semantics it does not provide.
- MySQL must support all four declared isolation levels.
- SQLite must reject unsupported levels instead of silently changing them.
- Intentional fault aborts are successful test actions when rollback succeeds.
- Every opened transaction receives one commit or rollback attempt.
- Native and WASM builds remain CGO-free.

---

### Task 1: Domain isolation and outcome contracts

**Files:**
- Modify: `internal/domain/types.go`
- Modify: `internal/domain/types_test.go`
- Modify: `internal/domain/parser_test.go`

**Interfaces:**
- Produces: `DatabaseConfig.Isolation domain.IsolationLevel`
- Produces: `ExecutionStatus`, `OperationError`, `ExecutionResult.Status`, `ExecutionResult.OperationErrors`, `ExecutionResult.Isolation`
- Produces: `TraceEvent.Phase`

- [x] **Step 1: Write failing validation and serialization tests**

Add table tests proving each supported isolation value passes, an unknown value returns `ErrSpecValidationFailed`, and an omitted value remains valid. Add JSON assertions for the five exact status strings and the new structured fields.

```go
func TestSpecValidateIsolation(t *testing.T) {
    valid := []domain.IsolationLevel{"", domain.LevelReadUncommitted, domain.LevelReadCommitted, domain.LevelRepeatableRead, domain.LevelSerializable}
    for _, level := range valid {
        spec := validSpec()
        spec.Database.Isolation = level
        if err := spec.Validate(); err != nil { t.Fatalf("isolation %q: %v", level, err) }
    }
    spec := validSpec()
    spec.Database.Isolation = "SNAPSHOT"
    if err := spec.Validate(); !errors.Is(err, domain.ErrSpecValidationFailed) {
        t.Fatalf("expected validation error, got %v", err)
    }
}
```

- [x] **Step 2: Run the domain tests and confirm RED**

Run: `go test ./internal/domain -run 'TestSpecValidateIsolation|TestExecutionStatusJSON' -count=1`

Expected: compile failure because the new fields and statuses do not exist.

- [x] **Step 3: Add the domain types and validation**

Add the five status constants, `OperationError`, `DatabaseConfig.Isolation`, `TraceEvent.Phase`, and result fields. Validate a non-empty isolation through a `switch` over the four constants.

- [x] **Step 4: Run domain tests and the parser suite**

Run: `go test ./internal/domain -count=1`

Expected: PASS.

- [x] **Step 5: Commit**

```bash
git add internal/domain/types.go internal/domain/types_test.go internal/domain/parser_test.go
git commit -m "feat: define transaction execution outcomes"
```

### Task 2: Adapter transaction options and isolation validation

**Files:**
- Create: `internal/drivers/isolation.go`
- Create: `internal/drivers/isolation_test.go`
- Modify: `internal/drivers/driver.go`
- Modify: `internal/drivers/sqlite.go`
- Modify: `internal/drivers/postgres.go`
- Modify: `internal/drivers/mysql.go`
- Modify: `internal/drivers/driver_mock.go`
- Modify: adapter tests under `internal/drivers/*_test.go`

**Interfaces:**
- Consumes: `domain.IsolationLevel`
- Produces: `drivers.TransactionOptions{Isolation domain.IsolationLevel}`
- Produces: `DatabaseDriver.BeginTx(context.Context, TransactionOptions)`
- Produces: `DatabaseDriver.EffectiveIsolation(domain.IsolationLevel)`

- [x] **Step 1: Write failing adapter isolation tests**

Cover empty defaults and supported/unsupported levels per adapter. Assert PostgreSQL rejects Read Uncommitted, MySQL accepts all four, SQLite accepts Serializable and Read Uncommitted, and mock accepts all four.

```go
func TestPostgresEffectiveIsolation(t *testing.T) {
    d := drivers.NewPostgresDriver("unused")
    if got, err := d.EffectiveIsolation(""); err != nil || got != domain.LevelReadCommitted {
        t.Fatalf("got %q, %v", got, err)
    }
    if _, err := d.EffectiveIsolation(domain.LevelReadUncommitted); err == nil {
        t.Fatal("expected unsupported isolation error")
    }
}
```

- [x] **Step 2: Run adapter tests and confirm RED**

Run: `go test ./internal/drivers -run 'EffectiveIsolation|BeginTxIsolation' -count=1`

Expected: compile failure for missing transaction options and methods.

- [x] **Step 3: Implement shared translation and adapter validation**

Create an unexported `toSQLIsolation(domain.IsolationLevel) (sql.IsolationLevel, error)`. Update all `BeginTx` implementations to resolve the effective value, translate it, and pass `sql.TxOptions{Isolation: level}`. Preserve constructor defaults for an empty request.

- [x] **Step 4: Update every interface consumer and test double**

Change direct calls from `BeginTx(ctx)` to `BeginTx(ctx, drivers.TransactionOptions{})`. Ensure the mock and WASM build satisfy the interface.

- [x] **Step 5: Run native and WASM adapter checks**

Run: `go test ./internal/drivers -count=1`

Run: `CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -o bin/eng01.wasm ./cmd/chaossql-wasm`

Expected: both PASS.

- [x] **Step 6: Commit**

```bash
git add internal/drivers
git commit -m "feat: validate transaction isolation by adapter"
```

### Task 3: Real transaction lifecycle per operation

**Files:**
- Create: `internal/engine/runner_transaction_test.go`
- Modify: `internal/engine/runner.go`
- Modify: `internal/engine/runner_savepoint_test.go`
- Modify: `internal/engine/runner_trace_test.go`

**Interfaces:**
- Consumes: `drivers.TransactionOptions`
- Produces: `engine.ScheduleOutcome`
- Produces: `Runner.ExecuteSchedule(...) (ScheduleOutcome, error)`
- Produces: unexported `executeOperation(...) operationOutcome`

- [x] **Step 1: Write the rollback regression test**

Use one SQLite worker and one scheduled operation containing a valid update followed by invalid SQL. Assert status data reports a step error, the trace contains a successful real rollback, and the original value remains unchanged.

```go
ops := []domain.ScheduledOp{{ID: 1, Name: "fail", Steps: []domain.StepConfig{
    {SQL: "UPDATE accounts SET balance = 9 WHERE id = 1"},
    {SQL: "NOT VALID SQL"},
}}}
outcome, err := runner.ExecuteSchedule(ctx, spec, ops)
if err != nil { t.Fatal(err) }
if len(outcome.OperationErrors) != 1 || outcome.OperationErrors[0].Phase != "step" { t.Fatalf("%+v", outcome) }
assertBalance(t, driver, 10)
```

- [x] **Step 2: Run the regression test and confirm RED**

Run: `go test ./internal/engine -run TestRunner_RollsBackFailedOperation -count=1`

Expected: balance is 9 under the current autocommit runner.

- [x] **Step 3: Implement `ScheduleOutcome` and `executeOperation`**

Open one transaction per operation, execute captures and statements through that transaction, commit successful operations, and roll back step failures. Record begin only after success and commit only after success. Collect structured errors and sort them before returning.

- [x] **Step 4: Add successful commit and lifecycle tests**

Assert a successful update persists; each successful operation has one begin and one commit; failed operations have one begin and one rollback; begin failures have an error without synthetic transaction events.

- [x] **Step 5: Adapt savepoint and trace tests to `ScheduleOutcome.Trace`**

Keep their database assertions. They must now prove savepoint SQL executes through the real transaction rather than autocommit.

- [x] **Step 6: Run engine transaction tests**

Run: `go test -race ./internal/engine -run 'Runner_(RollsBack|Commits|Savepoint|Trace)' -count=1`

Expected: PASS.

- [x] **Step 7: Commit**

```bash
git add internal/engine/runner.go internal/engine/runner_transaction_test.go internal/engine/runner_savepoint_test.go internal/engine/runner_trace_test.go
git commit -m "feat: execute operations in real transactions"
```

### Task 4: Result finalization and invariant error semantics

**Files:**
- Create: `internal/engine/runner_status_test.go`
- Modify: `internal/engine/runner.go`
- Test: `cmd/`, `internal/reporter/`, and `pkg/` packages for compatibility with the additive result fields

**Interfaces:**
- Consumes: `ScheduleOutcome`
- Produces: one shared result finalizer used by `Run` and `RunSchedule`

- [x] **Step 1: Write failing status precedence tests**

Add cases for passed, invariant false, invariant query error, operation error, and cancellation. Assert exact status and compatibility booleans. Assert invariants are not evaluated after operation errors or cancellation.

```go
if result.Status != domain.StatusExecutionError || result.Success || result.ViolationDetected {
    t.Fatalf("unexpected result: %+v", result)
}
```

- [x] **Step 2: Run status tests and confirm RED**

Run: `go test ./internal/engine -run 'TestRunner_Status|TestRunner_Inconclusive' -count=1`

Expected: current runner reports invariant errors as violations and operation failures can report success.

- [x] **Step 3: Implement the shared finalizer**

Apply precedence `canceled > execution_error > inconclusive > violation > passed`. Join operation errors for `ExecutionResult.Error`, copy the effective isolation, and derive compatibility booleans solely from status.

- [x] **Step 4: Use finalization from both run entry points**

Remove duplicated invariant loops from `Run` and `RunSchedule`. Preserve scheduled operations and duration.

- [x] **Step 5: Run engine, CLI, reporter, and SDK-facing tests**

Run: `go test -race ./internal/engine ./cmd/... ./internal/reporters ./pkg/... -count=1`

Expected: PASS after updating explicit expected values for the safer boolean behavior.

- [x] **Step 6: Commit**

```bash
git add internal/engine/runner.go internal/engine/runner_status_test.go
git commit -m "feat: report trustworthy execution statuses"
```

### Task 5: Context-aware waits and transaction cancellation

**Files:**
- Add cancellation cases to `internal/engine/runner_transaction_test.go`
- Modify: `internal/engine/runner.go`

**Interfaces:**
- Produces: unexported `waitForContext(context.Context, time.Duration) bool`

- [x] **Step 1: Write a failing cancellation test**

Start one operation with fixed jitter longer than the context timeout. Assert return occurs well before the jitter duration, status data is canceled, and the active transaction is rolled back.

```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
defer cancel()
started := time.Now()
outcome, err := runner.ExecuteSchedule(ctx, specWithJitter(500), ops)
if !errors.Is(err, context.DeadlineExceeded) || !outcome.Canceled { t.Fatalf("%+v %v", outcome, err) }
if time.Since(started) > 200*time.Millisecond { t.Fatal("cancellation was not prompt") }
```

- [x] **Step 2: Run cancellation test and confirm RED**

Run: `go test ./internal/engine -run TestRunner_CancellationRollsBackPromptly -count=1`

Expected: current `time.Sleep` delays return or no real rollback occurs.

- [x] **Step 3: Replace both sleeps with context-aware timers**

Use one timer helper for jitter and injected latency. On cancellation, stop operation execution, roll back any open transaction, and return `context.Cause(ctx)` or `ctx.Err()`.

- [x] **Step 4: Run cancellation repeatedly under the race detector**

Run: `go test -race ./internal/engine -run TestRunner_CancellationRollsBackPromptly -count=20`

Expected: PASS with no race report.

- [x] **Step 5: Commit**

```bash
git add internal/engine/runner.go internal/engine/runner_transaction_test.go
git commit -m "feat: cancel transactional operations promptly"
```

### Task 6: Live adapter transaction fixtures and specifications

**Files:**
- Modify: `internal/drivers/postgres_test.go`
- Modify: `internal/drivers/mysql_test.go`
- Modify: `cmd/chaossql/engine.go`
- Modify: `cmd/chaossql/main.go`
- Modify: Python and TypeScript SDK result adapters and compatibility fixtures
- Modify: `specs/02_concurrency_interleaving.md`
- Modify: `specs/06_mysql_savepoints_and_otel.md`
- Modify: `specs/08_fault_injection_and_dirty_reads.md`

**Interfaces:**
- Consumes: `Runner.RunSchedule`, result statuses, and transaction options
- Produces: required live rollback/commit/isolation fixtures

- [x] **Step 1: Add reusable live database test cases**

For PostgreSQL and MySQL, run one successful commit and one update-plus-invalid-SQL rollback. Query the final state through the driver and assert the effective isolation. Use `skipUnavailableDatabase` locally and the existing `CHAOSSQL_REQUIRE_DATABASES=1` behavior in CI.

- [x] **Step 2: Run live tests against configured services**

Run: `CHAOSSQL_REQUIRE_DATABASES=1 DATABASE_URL="$DATABASE_URL" MYSQL_DSN="$MYSQL_DSN" go test -race ./internal/drivers ./internal/engine -run 'TransactionIntegration|EffectiveIsolation' -count=1`

Expected: PASS when both services are available; unavailable required services must fail, never skip.

- [x] **Step 3: Update the three formal specifications**

State that each operation uses a real transaction, transaction events correspond to completed driver actions, savepoints share that transaction, intentional aborts must roll back, waits are cancellable, and unsupported isolation is an error.

- [x] **Step 4: Run the complete quality gate**

Run: `make verify`

Expected: `✔ Verification gate completed successfully!`

- [x] **Step 5: Commit**

```bash
git add internal/drivers/postgres_test.go internal/drivers/mysql_test.go internal/engine/runner_integration_test.go specs/02_concurrency_interleaving.md specs/06_mysql_savepoints_and_otel.md specs/08_fault_injection_and_dirty_reads.md
git commit -m "test: verify transactional semantics on live databases"
```

### Task 7: Final compatibility audit and task completion

**Files:**
- Verify: all files changed in Tasks 1-6
- Modify: `docs/superpowers/plans/2026-09-10-transaction-semantics.md` checkbox state

**Interfaces:**
- Verifies all contracts from the design specification

- [x] **Step 1: Check formatting, generated diff integrity, and zero-CGO builds**

Run: `gofmt -w internal/domain/types.go internal/domain/types_test.go internal/domain/parser_test.go internal/drivers/driver.go internal/drivers/isolation.go internal/drivers/isolation_test.go internal/drivers/sqlite.go internal/drivers/postgres.go internal/drivers/mysql.go internal/drivers/driver_mock.go internal/drivers/postgres_test.go internal/drivers/mysql_test.go internal/engine/runner.go internal/engine/runner_transaction_test.go internal/engine/runner_status_test.go internal/engine/runner_savepoint_test.go internal/engine/runner_trace_test.go internal/engine/runner_integration_test.go`

Run: `git diff --check origin/main...HEAD`

Run: `CGO_ENABLED=0 go build ./cmd/chaossql ./cmd/chaossql-server`

Run: `CGO_ENABLED=0 GOOS=js GOARCH=wasm go build -o bin/eng01-final.wasm ./cmd/chaossql-wasm`

- [x] **Step 2: Run the fresh final gate**

Run: `make verify`

Expected: PASS with no new skip in required CI mode.

- [x] **Step 3: Request independent code review**

Review `origin/main...HEAD` against `docs/superpowers/specs/2026-09-10-transaction-semantics-design.md`. Resolve every Critical and Important finding and rerun affected tests.

- [x] **Step 4: Push, open the ENG-01 PR, and wait for all checks**

Push `codex/eng-01-transaction-semantics`, create a PR targeting `main`, and require every GitHub check to pass.

- [x] **Step 5: Merge and record completion**

Merge the approved PR with a merge commit, then update the ignored commercialization plan so the ENG-01 row links to the merged PR and is marked complete.
