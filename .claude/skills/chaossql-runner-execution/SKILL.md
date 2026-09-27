---
name: chaossql-runner-execution
description: The Runner lifecycle in internal/engine/runner.go — reset, schedule, concurrent worker transactions, trace events, rollback semantics, cancellation and final status resolution (passed/violation/execution_error/inconclusive/canceled). Use when changing execution, trace events, status precedence, or debugging unexpected statuses.
---

# Runner Execution

`engine.Runner` executes one schedule against one `drivers.DatabaseDriver` and
produces a `domain.ExecutionResult` (alias `engine.RunResult`).

## When to use

- Modifying `Run`, `RunSchedule`, `ExecuteSchedule`, `executeOperation`,
  `finalizeResult`, or trace event emission.
- A run ends in an unexpected status or the trace looks wrong.

## Entry points

| Function | Use |
| :--- | :--- |
| `NewRunner(driver, seed)` | Holds driver, evaluator, `PRNG(seed)` |
| `Run(ctx, spec)` | Validate → reset → `GenerateSchedule` → `ExecuteSchedule` → finalize |
| `RunSchedule(ctx, spec, ops)` | Same, but with caller-provided ops (shrinker, replay, diff) |
| `ExecuteSchedule(ctx, spec, ops)` | Only the concurrent execution, no reset/finalize |

## Lifecycle of `Run` / `RunSchedule`

1. `spec.ValidateInvariantNames()`; `driver.EffectiveIsolation(spec.Database.Isolation)`
   — both fail **before** touching the database.
2. `driver.Reset(ctx, schema, seed)` — fresh state for every run (and every
   shrink trial). Error → `(nil, "database reset failed: ...")`.
3. Ops: generated (`Run`) or given (`RunSchedule`).
4. `ExecuteSchedule`:
   - validate op IDs (positive, unique), workers default 4;
   - `BuildSchedulePlan` and index decisions by `(opID, stepIndex)`;
   - one goroutine per worker, each running its ID-ordered queue sequentially;
   - a worker stops picking new ops when `ctx` is done.
5. `finalizeResult` decides the status.

## One operation = one transaction (`executeOperation`)

```
BeginTx(isolation) ── error ──▶ ERROR event (phase "begin", step 0) + OperationError
      │
   BEGIN event (step 0)
      │  for each step s = 1..n:
      │    ctx done?                  → rollback (no error recorded)
      │    wait decision.JitterMs      (ctx-aware)
      │    wait decision.LatencyMs     (ctx-aware)
      │    decision.Abort?            → rollback (intentional; NOT an execution error)
      │    sql := SubstituteParams(step.SQL, params+captures)
      │    capture? QueryRow.Scan → state[var] = fmt("%v")   else Exec
      │    error → ERROR event (phase "step") → rollback → OperationError
      │    ok    → event type from DetectEventType(sql) (EXEC/SAVEPOINT/ROLLBACK_TO/RELEASE_SAVEPOINT)
   Commit ── error ──▶ ERROR event (phase "commit", step n+1) → rollback → OperationError
      │
   COMMIT event (step n+1)
```

`rollback(...)` calls `tx.Rollback()`, ignores `sql.ErrTxDone` when the
context is already canceled, and always records a `ROLLBACK` event (with the
rollback error, if any). A rollback failure becomes its own `OperationError`
(phase `rollback`).

## Trace

- Events are appended under a mutex with `Timestamp = time.Since(start)`;
  their order is the real completion order — **not deterministic**.
- `OpIndex` is the scheduled op ID; `StepIndex` is 0 for BEGIN, 1..n for
  steps, n+1 for COMMIT; `Phase` is begin/step/commit/rollback.
- `OperationErrors` are sorted by (op ID, step, phase) for stable output.

## Status precedence (`finalizeResult`)

1. `canceled` — context done (result still returned together with `ctx.Err()`).
2. `execution_error` — **any** `OperationError` (serialization failures,
   deadlocks, lock timeouts, constraint errors, failed captures, commit errors).
   Invariants are not evaluated.
3. For each invariant in spec order: evaluation error → `inconclusive`
   (stop); `Passed == false` → `violation` with `FailingInvariant` (stop).
4. Otherwise `passed`.

`Success` = passed, `ViolationDetected` = violation. `Seed`, `Schedule`,
`Isolation` (effective), `ScheduledOps` and `Duration` are always filled.

## Gotchas

- There is no "initial invariant check" before execution, despite the
  sequence diagram in `ARCHITECTURE.md`; the shrinker's empty-schedule baseline
  plays that role.
- Under SERIALIZABLE, a database-detected conflict produces `execution_error`,
  not `passed` — CLI callers then exit non-zero (`unreliableRunError`).
- Intentional aborts produce ROLLBACK events but keep the run eligible for
  `passed`/`violation`.
- Invariants are evaluated with `driver.Query` outside any transaction after
  all workers finish.
- Worker count > number of ops simply leaves workers idle.

## Change checklist

- New event type → `domain.TraceEventType`, `DetectEventType`, every reporter
  that switches on types (Mermaid, OTLP, HTML/UI, replay colors), Adya graph
  builder (only `EXEC` events are analyzed), Cloud `ClassifyOpType`.
- New status → `domain.ExecutionStatus`, `unreliableRunError` (CLI),
  `FailureSignatureFor`, IPC/SDK types, Cloud allowed statuses
  (`internal/cloud/payload.go`), JUnit/summary reporters.

## Tests

- `internal/engine/runner_status_test.go`, `runner_transaction_test.go`,
  `runner_trace_test.go`, `runner_savepoint_test.go`, `runner_schedule_test.go`,
  `runner_test.go`
- `go test -race ./internal/engine/...`

## Source map

- `internal/engine/runner.go`
- `internal/domain/types.go`
- `internal/engine/runner_status_test.go`
- `internal/engine/runner_transaction_test.go`
- `internal/engine/runner_savepoint_test.go`
- `specs/02_concurrency_interleaving.md`

## Related skills

- `chaossql-deterministic-schedule`, `chaossql-invariant-evaluation`,
  `chaossql-database-drivers`, `chaossql-param-generators`, `chaossql-ddmin-shrinker`
