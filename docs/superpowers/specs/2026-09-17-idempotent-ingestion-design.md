# API-01 Reliable Ingestion, Baselines, and Alert Outbox Design

## Problem

Hosted run ingestion currently exhibits three failure modes under retries, high concurrency, and real CI execution patterns:
1. `Store.SaveRun` executes independent `INSERT INTO runs` and `INSERT INTO findings` statements without a database transaction. If finding insertion fails or times out, the run remains orphaned and partially persisted.
2. Ingestion requests lack an idempotency boundary. Retrying a failed HTTP request generates a new run ID (`run_<timestamp>`), polluting the repository history, re-triggering regression evaluation, and spamming incident webhooks.
3. Baseline evaluation naively adopts any passing run on the default branch as the latest baseline. In asynchronous CI, a job for an older commit finishing late will overwrite the baseline established by a newer commit.
4. Webhook alerts are fired in unbuffered, detached goroutines directly in the HTTP request lifecycle. A server crash or restart permanently drops undelivered alerts, and destination endpoints are vulnerable to SSRF.

## Decision

We introduce atomic, idempotent run persistence, chronological and fingerprint-aware baseline management, and a transactional outbox table for reliable webhook delivery.

### Schema Migration

A new migration `2026-09-17-idempotent-runs-and-outbox`:
1. Alters `runs` to add:
   - `idempotency_key TEXT`
   - `scenario_fingerprint TEXT NOT NULL DEFAULT ''`
   - `commit_timestamp DATETIME`
2. Creates unique index `idx_runs_repo_idempotency` on `(repo_id, idempotency_key)` where `idempotency_key IS NOT NULL`.
3. Creates table `webhook_outbox`:
   - `id TEXT PRIMARY KEY`
   - `org_id TEXT NOT NULL`
   - `webhook_id TEXT NOT NULL`
   - `event_type TEXT NOT NULL`
   - `payload_json TEXT NOT NULL`
   - `status TEXT NOT NULL DEFAULT 'pending'`
   - `attempts INTEGER NOT NULL DEFAULT 0`
   - `next_retry_at DATETIME NOT NULL`
   - `created_at DATETIME NOT NULL`
   - `delivered_at DATETIME`
   - `last_error TEXT NOT NULL DEFAULT ''`
4. Creates table `schema_migrations` tracking if not present.

### Atomic and Idempotent Ingestion

`Store.SaveRunTx(ctx, run, finding, outboxItems)`:
- Executes within an atomic `BeginTx(ctx)`.
- If `idempotency_key` is provided and already exists for `repo_id`, queries the existing `RunRecord` and returns it with a boolean `created=false` (idempotent no-op).
- Otherwise, inserts `runs`, `findings`, and any queued `webhook_outbox` entries in the same transaction.

### Chronological and Fingerprint-Aware Baseline

`RegressionEngine.Evaluate`:
- Compares `run.ScenarioFingerprint` against the baseline's fingerprint. If mismatched, the baseline is considered non-comparable rather than producing false regressions.
- If `run.Branch == defaultBranch && run.PRNumber == 0 && run.Status == "passed"`:
  - If a baseline already exists, updates only if `run.CommitTimestamp >= baseRun.CommitTimestamp`.
  - Rejects baseline degradation from late-arriving older commits.

### Persistent Outbox Dispatcher

`OutboxDispatcher`:
- Polls `webhook_outbox` for `status = 'pending' AND next_retry_at <= ?`.
- Validates the target URL with a custom `net.Dialer` that rejects loopback, RFC 1918 private, and link-local IP addresses before connection.
- Follows redirects only if the redirect target passes the same SSRF validation.
- On HTTP 2xx: marks `status = 'delivered'`, records `delivered_at`.
- On error / 5xx: increments `attempts`. If `attempts < 3`, sets `next_retry_at` with exponential backoff (e.g. 2s, 8s). If `attempts >= 3`, marks `status = 'failed'`.

## Compatibility

- Existing runs with `NULL` idempotency key remain valid.
- The `Idempotency-Key` header is optional; if omitted, runs receive unique IDs as before.
- Existing webhook integrations work seamlessly without client-side modifications.