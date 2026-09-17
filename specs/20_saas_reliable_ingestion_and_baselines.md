# Specification 20: SaaS Reliable Ingestion, Baselines, and Alert Outbox

## Status

Normative for hosted run ingestion, baseline regression evaluation, and transactional alert notifications.

## Idempotent Ingestion

1. The server MUST accept an optional idempotency key from the `Idempotency-Key` HTTP header or the `idempotency_key` metadata payload field.
2. An idempotency key MUST be unique within a repository scope `(repo_id, idempotency_key)`.
3. If a run ingestion request is submitted with an `idempotency_key` that already exists for the repository:
   - The server MUST return HTTP 200 with the previously saved run record.
   - The server MUST NOT insert duplicate rows into `runs` or `findings`.
   - The server MUST NOT re-enqueue duplicate webhook notifications.
4. If an ingestion request fails at any point (such as during finding creation or outbox staging), all database writes for that run MUST be rolled back in a single atomic transaction.

## Baseline Regression & Chronology

1. A run MUST record a `scenario_fingerprint` capturing the scenario configuration, driver, and invariant identities.
2. A run MUST record a `commit_timestamp` representing the author/commit creation time in UTC.
3. Baseline evaluation on the default branch:
   - Only runs with `status == "passed"` and `pr_number == 0` on the default branch are eligible to update the baseline.
   - An eligible run MUST NOT overwrite an existing baseline if its `commit_timestamp` is earlier than the existing baseline's `commit_timestamp`.
   - An eligible run MUST have a compatible `scenario_fingerprint` to update or evaluate against that scenario's baseline.

## Transactional Webhook Outbox

1. Alerts for regressions and run failures MUST NOT be dispatched via detached, unbounded goroutines.
2. When an ingestion triggers an alert event, the server MUST insert delivery records into the `webhook_outbox` table within the same transaction that persists the run.
3. The outbox dispatcher MUST deliver pending notifications asynchronously with bounded concurrency.
4. The outbox dispatcher MUST support automatic retry with exponential backoff for up to 3 attempts.
5. SSRF protection:
   - Webhook destination URLs MUST use `http` or `https` schemes.
   - The dispatcher MUST resolve the destination hostname and reject requests targeting loopback (`127.0.0.0/8`, `::1`), private (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), or link-local (`169.254.0.0/16`, `fe80::/10`) addresses.
   - HTTP redirects MUST be validated through the same SSRF resolver before being followed.
6. The outbox record MUST store delivery status (`pending`, `delivered`, `failed`), retry count, error details, and delivery timestamps.

## Verification

Tests MUST verify:
- Ingestion atomicity under simulated database failures.
- Idempotent deduplication under concurrent duplicate requests.
- Protection against out-of-order commits overwriting baselines.
- Persistent outbox queueing, worker retry, and SSRF blocking.