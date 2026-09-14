# Spec 02: Concurrency Scheduler & Step Interleaving

## Objective
Orchestrate multiple concurrent asynchronous workers and inject latency jitter between transaction steps to trigger race conditions.

## Verifiable Requirements
1. **Transactional Operation Boundary:** Every scheduled operation opens exactly one driver transaction. All captures, statements, savepoints, and partial rollbacks in that operation execute through the same transaction until one final commit or rollback.
2. **Isolation Validation:** The requested `database.isolation` is resolved and validated by the selected adapter before workers start. An unsupported level is an execution error and must not silently fall back to another level.
3. **Intra-Transaction Interleaving:** If an operation consists of $k > 1$ steps, the scheduler injects micro-jitter delays between steps, allowing other workers to interleave concurrent operations against the database.
4. **Context-Aware Delays:** Jitter and injected latency stop promptly when the execution context is canceled. Any transaction already opened for that operation is rolled back before the worker exits.
5. **Random Rollbacks:** Support aborting transactions at pseudo-random injection points to test application resilience against partial failures. An intentional abort performs a real driver rollback and never leaves earlier steps committed.
6. **Append-Only Trace:** Every transaction lifecycle event records timestamp, `worker_id`, `op_id`, phase, and SQL statement. `BEGIN`, `COMMIT`, and `ROLLBACK` events are appended only after the corresponding driver action succeeds; failures are recorded as error events.
7. **Trustworthy Outcome:** Final status follows the precedence `canceled`, `execution_error`, `inconclusive`, `violation`, then `passed`. Infrastructure and statement failures must never be reported as confirmed invariant violations.
8. **Versioned Logical Schedule:** Before workers start, the harness records a versioned schedule containing the effective seed, normalized worker count, deterministic operation-to-worker assignment, and every per-step jitter, latency, and abort decision.
9. **Stable Decision Derivation:** Parameter names are sorted before random generation. Generator counters are scoped to one run, and injected decisions derive from operation and step identity rather than goroutine call order.
10. **Deterministic Zero Seed:** Seed `0` is a literal reproducible seed. Omitted seeds resolve to zero unless a caller explicitly chooses and persists another seed.
11. **Deterministic Worker Queues:** Operation ID deterministically selects a worker queue, and each queue processes operations in ascending ID order. Changing the configured worker count intentionally changes this assignment because worker count is part of the specification.
12. **External Timing Boundary:** The schedule is the harness-controlled execution intent. The trace is the observed physical completion order. Database locks, server scheduling, network delay, deadlock victim selection, and visibility timing are external and may vary without changing the logical schedule.
