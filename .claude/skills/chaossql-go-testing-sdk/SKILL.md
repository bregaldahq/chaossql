---
name: chaossql-go-testing-sdk
description: The pkg/chaostest fluent Go testing API (New, WithDriver/Schema/Seed/Invariant/Jitter, AddOperation, AddOperationWithParams, Run, AssertNoAnomalies) — how it builds a Spec, its defaults, shrink fallback behavior, and failure output. Use when changing pkg/chaostest or writing Go tests that fuzz a schema with ChaosSQL.
---

# Go Testing SDK (`pkg/chaostest`)

Lets Go users write concurrency fuzz tests inside `go test` without YAML.

## When to use

- Changing `pkg/chaostest/chaostest.go`.
- Writing a Go test that uses ChaosSQL programmatically.

## API

```go
chaostest.New(t).
    WithDriver(drivers.NewSQLiteDriver("")).   // optional; default: fresh in-memory SQLite
    WithSchema(schemaSQL).
    WithSeed(seedSQL).
    WithInvariant("total", "SELECT SUM(balance) AS total FROM accounts", "total == 2000").
    WithJitter(1, 5).                          // default [1, 5] ms
    AddOperation("withdraw",
        "SELECT balance FROM accounts WHERE id = 1; -> bal",
        "UPDATE accounts SET balance = {bal - 10} WHERE id = 1").
    AddOperationWithParams("deposit", map[string]string{"amt": "$random_int(1, 9)"}, "...").
    AssertNoAnomalies(ctx, 4 /*workers*/, 50 /*iterations*/, 42 /*seed*/)
```

- Steps accept the inline capture suffix `-> var` / `=> var`.
- `Run(ctx, workers, iterations, seed)` returns
  `(*ExecutionResult, *ShrinkResult, error)`.

## `Run` behavior

1. Under a mutex; creates `file:chaostest_mem_<nanos>?mode=memory&cache=shared`
   SQLite when no driver is set, then `Open`.
2. Builds `Spec{Version "1.1", Name "chaostest", Driver: driver.DriverName(), ...}`
   (no isolation → driver default; no faults).
3. `runner.Run`. Statuses other than passed/violation → error.
4. On violation: shrink + verify (`chaossql-ddmin-shrinker`). On shrink failure
   or non-reproducing minimal ops it returns a fallback `ShrinkResult` with the
   original ops and 0% reduction instead of an error; cancellation is returned.

## `AssertNoAnomalies`

Calls `Run`; `t.Fatalf` on execution errors; on violation it re-runs the
minimal schedule to classify the anomaly and fails the test with the anomaly
class, failing invariant and minimal ops (plus reporter output).

## Gotchas

- Default SQLite driver runs at SERIALIZABLE on one connection → transactions
  serialize and classic anomalies do not appear (`chaossql-database-drivers`).
  Pass a driver configured for the level you want to test.
- The driver is reused across `Run` calls on the same `Tester` (reset each run).
- `Spec.Validate` is not called; empty invariants/operations produce a run
  that trivially passes or an empty schedule.

## Tests

- `pkg/chaostest/chaostest_test.go`

## Source map

- `pkg/chaostest/chaostest.go`
- `pkg/chaostest/chaostest_test.go`
- `specs/11_version_1_1_developer_sdk_and_smart_generators.md`

## Related skills

- `chaossql-runner-execution`, `chaossql-ddmin-shrinker`,
  `chaossql-database-drivers`, `chaossql-param-generators`
