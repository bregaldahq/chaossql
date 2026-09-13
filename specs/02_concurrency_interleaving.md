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
