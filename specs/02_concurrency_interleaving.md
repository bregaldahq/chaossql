# Spec 02: Concurrency Scheduler & Step Interleaving

## Objective
Orchestrate multiple concurrent asynchronous workers and inject latency jitter between transaction steps to trigger race conditions.

## Verifiable Requirements
1. **Queue Distribution:** Each worker maintains its own dedicated queue or channel and an isolated database connection.
2. **Intra-Transaction Interleaving:** If an operation consists of $k > 1$ steps, the scheduler injects micro-jitter delays between steps, allowing other workers to interleave concurrent operations against the database.
3. **Random Rollbacks:** Support aborting transactions at pseudo-random injection points to test application resilience against partial failures.
4. **Append-Only Trace:** Every transaction lifecycle event (execution, commit, rollback, driver error) is recorded with timestamp, `worker_id`, `op_id`, and executed SQL statement.
