# Transaction Semantics Design

**Date:** September 10, 2026
**Backlog item:** ENG-01
**Status:** Approved for implementation planning

## Purpose

ChaosSQL describes each operation as a transaction, but the runner currently executes operation steps through the driver pool in autocommit mode. The trace records `BEGIN`, `COMMIT`, and `ROLLBACK` events without performing those actions. A failed operation can therefore leave earlier writes committed and the final result can still report success.

ENG-01 makes every operation an actual database transaction, exposes the requested and effective isolation level, and distinguishes a valid invariant failure from an execution that cannot support a correctness conclusion.

## Scope

This change covers transaction boundaries for scheduled operations, isolation configuration, savepoints inside those transactions, context cancellation, operation and invariant errors, trace lifecycle accuracy, result status, and tests against SQLite, PostgreSQL, MySQL, and the mock adapter.

Deterministic parameter generation and scheduling remain ENG-02. Replay identity and minimization of the same failure remain ENG-03. ENG-01 must avoid adding new scheduling randomness or claiming deterministic interleaving.

## Domain Contract

`DatabaseConfig` gains an optional `isolation` field using the existing `IsolationLevel` values:

```yaml
database:
  driver: postgres
  dsn: ${DATABASE_URL}
  isolation: READ_COMMITTED
```

The parser accepts `READ_UNCOMMITTED`, `READ_COMMITTED`, `REPEATABLE_READ`, and `SERIALIZABLE`. An empty value requests the adapter's documented default so existing scenario files remain valid. Any other value fails spec validation before database reset or schedule generation.

The driver port gains transaction options and reports the effective isolation level:

```go
type TransactionOptions struct {
    Isolation domain.IsolationLevel
}

type Tx interface {
    ExecContext(context.Context, string, ...any) (sql.Result, error)
    QueryContext(context.Context, string, ...any) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...any) *sql.Row
    Commit() error
    Rollback() error
}

type DatabaseDriver interface {
    // Existing lifecycle and non-transactional methods remain.
    BeginTx(context.Context, TransactionOptions) (Tx, error)
    EffectiveIsolation(domain.IsolationLevel) (domain.IsolationLevel, error)
}
```

`EffectiveIsolation` validates support before execution. PostgreSQL accepts `READ_COMMITTED`, `REPEATABLE_READ`, and `SERIALIZABLE`; it rejects `READ_UNCOMMITTED` because PostgreSQL silently implements that request as Read Committed. MySQL accepts all four values. SQLite accepts `SERIALIZABLE` and `READ_UNCOMMITTED`; an empty request resolves to `SERIALIZABLE`. The mock adapter accepts all values and resolves an empty request to `SERIALIZABLE`.

Driver constructors may retain their existing options for direct adapter consumers. A non-empty per-run transaction option takes precedence. An empty option uses the constructor default and the adapter reports that default as the effective level.

## Transaction Execution

Each scheduled operation follows one lifecycle on one `Tx` instance:

1. Resolve and validate the effective isolation level before workers start.
2. Call `BeginTx` when a worker receives an operation.
3. Record `BEGIN` only after `BeginTx` succeeds.
4. Copy the operation parameters into operation-local state.
5. Execute every capture query with `tx.QueryRowContext` and every other statement with `tx.ExecContext`.
6. Execute `SAVEPOINT`, `ROLLBACK TO`, and `RELEASE SAVEPOINT` statements through the same transaction. Their existing trace event types remain unchanged.
7. Commit after the last successful step and record `COMMIT` only after the commit succeeds.
8. Roll back after a step error, injected abort, or cancellation. Record the rollback attempt and its error, if any.

An operation never falls back to `DatabaseDriver.Exec` or `DatabaseDriver.QueryRow` after its transaction begins. Invariant evaluation remains outside operation transactions and runs only after all workers finish.

If `BeginTx` fails, the trace records an `ERROR` event with phase `begin`; no synthetic `BEGIN` or `ROLLBACK` event is emitted because no transaction exists. A step failure records `ERROR` for that step and then the real rollback. A commit failure records `ERROR` with phase `commit`; the runner attempts rollback to release resources and records its result. A rollback failure is retained alongside the original failure.

Injected aborts are intentional test behavior. A successful rollback caused by `abort_probability` does not make the whole run an execution error. Failure to perform that rollback does.

## Result States

`ExecutionResult` gains a `status` field and structured execution diagnostics while retaining `success`, `violation_detected`, `failing_invariant`, and `error` for compatibility.

```go
type ExecutionStatus string

const (
    StatusPassed         ExecutionStatus = "passed"
    StatusViolation      ExecutionStatus = "violation"
    StatusExecutionError ExecutionStatus = "execution_error"
    StatusInconclusive   ExecutionStatus = "inconclusive"
    StatusCanceled       ExecutionStatus = "canceled"
)

type OperationError struct {
    OperationID int    `json:"operation_id"`
    Operation   string `json:"operation"`
    StepIndex   int    `json:"step_index,omitempty"`
    Phase       string `json:"phase"`
    Message     string `json:"message"`
}
```

`ExecutionResult.OperationErrors` is ordered by operation ID, step index, and phase before return. `ExecutionResult.Isolation` contains the effective isolation level. The compatibility fields obey these rules:

| Status | `success` | `violation_detected` | Meaning |
|---|---:|---:|---|
| `passed` | true | false | Every operation completed or intentionally aborted, and every invariant passed |
| `violation` | false | true | Execution completed reliably and at least one invariant returned false |
| `execution_error` | false | false | A begin, SQL, commit, or required rollback operation failed |
| `inconclusive` | false | false | Operations completed, but an invariant query or expression could not be evaluated |
| `canceled` | false | false | The caller's context ended before execution completed |

Status precedence is `canceled`, `execution_error`, `inconclusive`, `violation`, then `passed`. The runner does not evaluate invariants after cancellation or an operation execution error because the resulting state cannot support a trustworthy conclusion. Intentional aborts do not trigger this short circuit.

The existing `Error` field carries a joined Go error for in-process callers. JSON consumers use `operation_errors` and the failing invariant's existing error representation rather than relying on serialization of the `error` interface.

## Cancellation

Jitter and injected latency use a context-aware timer rather than `time.Sleep`. Cancellation while waiting stops the timer, rolls back the active transaction, records the rollback, and produces `canceled`.

Database calls continue to receive the caller context. When a call returns because the context ended, the operation is classified as canceled instead of execution error. Workers stop accepting new operations after cancellation. Every transaction that was successfully opened is finalized with commit or rollback before `ExecuteSchedule` returns.

## Trace Accuracy

Trace events gain an optional `phase` field with the values `begin`, `step`, `commit`, or `rollback`. Existing JSON fields and event type strings remain stable.

Lifecycle events describe completed database actions:

- `BEGIN`: the adapter returned a live transaction.
- `EXEC`, `SAVEPOINT`, `ROLLBACK_TO`, or `RELEASE_SAVEPOINT`: the statement succeeded on that transaction.
- `COMMIT`: commit succeeded.
- `ROLLBACK`: rollback was attempted; its `error` field is empty only when rollback succeeded.
- `ERROR`: the associated action failed, with the database error in `error`.

ENG-01 preserves append-only capture under a mutex. Timestamp and cross-worker event ordering remain observational; ENG-02 will replace uncontrolled scheduling with a logical schedule.

## Adapter Behavior

The adapters translate `domain.IsolationLevel` to `sql.IsolationLevel` in one shared helper and perform adapter-specific support validation before calling `sql.DB.BeginTx`.

- PostgreSQL maps Read Committed, Repeatable Read, and Serializable directly. Read Uncommitted returns a descriptive unsupported-isolation error.
- MySQL maps all four levels directly.
- SQLite maps Serializable directly. Read Uncommitted configures the transaction connection consistently with SQLite support; unsupported levels return an error rather than silently changing semantics.
- Mock maps values for contract tests and exposes transaction counters needed to assert lifecycle behavior.

The adapters continue to normalize deadlock, serialization, timeout, and connection errors. ENG-01 does not change database reset behavior.

## Internal Runner Structure

`ExecuteSchedule` returns a schedule outcome instead of using a single error as both cancellation and execution failure:

```go
type ScheduleOutcome struct {
    Trace           domain.ExecutionTrace
    OperationErrors []domain.OperationError
    Canceled        bool
}
```

The worker delegates one operation to `executeOperation`. That function owns transaction cleanup and returns operation events plus an optional structured error. Central trace insertion remains synchronized. This boundary keeps transaction state local to one function and makes commit, rollback, cancellation, and savepoint behavior testable without duplicating the worker loop.

`Run` and `RunSchedule` share one result-finalization function so status precedence and compatibility booleans cannot diverge.

## Compatibility

Existing YAML without `database.isolation` remains valid. Existing JSON fields remain present with their current names. Consumers that only inspect `success` or `violation_detected` receive safer behavior: execution failures no longer appear as successful runs or invariant violations.

Changing `DatabaseDriver.BeginTx` is an internal Go API change. All in-repository adapters and test doubles are updated in the same change. The public SDKs invoke the CLI and do not implement this interface.

## Verification

The implementation must add fixtures for these cases:

1. A transaction updates a value and then executes invalid SQL. The original value remains and status is `execution_error`.
2. A successful operation commits its value and status is `passed` when invariants pass.
3. An intentional abort rolls back earlier writes without producing `execution_error`.
4. Savepoint rollback removes only changes after the savepoint; release and final commit succeed.
5. Cancellation during context-aware jitter rolls back the active transaction and returns `canceled` promptly.
6. An invariant that evaluates to false produces `violation`; an invariant query or expression error produces `inconclusive`.
7. Unsupported isolation fails before schedule execution and reports the requested driver and isolation level.
8. PostgreSQL and MySQL live integration tests verify rollback, commit, and effective isolation under the required CI database mode introduced by QA-01.
9. The mock adapter proves each opened transaction receives exactly one successful commit or rollback attempt.
10. `make verify` passes, including zero-CGO native and WASM builds. WASM uses its own adapter implementation and must satisfy the updated interface without introducing CGO.

The specifications `specs/02_concurrency_interleaving.md`, `specs/06_mysql_savepoints_and_otel.md`, and `specs/08_fault_injection_and_dirty_reads.md` are updated in the implementation commit that changes the corresponding behavior.

## Completion Criteria

ENG-01 is complete when operation steps use a real transaction on every supported native adapter; rollback restores state after failures and injected aborts; commit persists successful operations; requested isolation is validated and effective isolation is reported; cancellation closes open transactions; result states distinguish passed, violation, execution error, inconclusive, and canceled; live PostgreSQL/MySQL fixtures pass in CI; `make verify` is green; independent review has no unresolved Critical or Important finding; and the merged PR is recorded as complete in the commercialization plan.
