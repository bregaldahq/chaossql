# Specification 20: SaaS Reliable Ingestion, Baselines, and Alert Outbox

## Status

Normative for hosted run ingestion, baseline regression evaluation, and transactional alert notifications.

## Idempotent Ingestion

1. The server MUST accept an optional idempotency key from the `Idempotency-Key` HTTP header or the `idempotency_key` metadata payload field. When both are present, they MUST agree; disagreement returns HTTP 400. A CI workflow's `run_id` MUST NOT be used as the execution key. Legacy requests without a key create distinct executions.
2. An idempotency key MUST be unique within a repository scope `(repo_id, idempotency_key)`.
3. If a run ingestion request is submitted with an `idempotency_key` that already exists for the repository:
   - Identical execution content MUST return HTTP 200 with the original saved response, including the original baseline comparison and regression decision. Concurrent retries MUST create only one execution.
   - Different execution content MUST return HTTP 409 without changing the original execution. A missing measurement and an explicitly measured zero are distinct content.
   - Keys from legacy records without a durable response MUST return HTTP 409 rather than inventing a replay result.
   - The server MUST NOT insert duplicate rows into `runs` or `findings`.
   - The server MUST NOT re-enqueue duplicate webhook notifications.
4. Repository quota checks, repository/scenario resolution, run/finding insertion, baseline evaluation, response persistence, and required outbox writes MUST share one transaction. Any storage failure MUST roll back all those writes and return an error. Recovery followed by a retry MUST produce one execution and one notification per matching destination.
5. Repository plan limits MUST be enforced atomically during creation. Existing repositories remain usable at capacity. A request requiring a new repository beyond capacity returns HTTP 403 with a repository limit error.
6. Returned run URLs MUST target the dashboard's supported `/#/dashboard?run=<id>` route. They MUST NOT imply that hosted SQL or reproduction artifacts are available.

## Baseline Regression & Chronology

1. Official clients MUST provide a stable `scenario_fingerprint` covering scenario semantics, driver, database seed statements, and invariant identities. The execution's PRNG seed MUST NOT change that fingerprint. SQL contributes to the local fingerprint but MUST NOT be uploaded.
2. Commit chronology MUST use `ci.commit_timestamp`, representing the actual commit time in UTC. The upload's `timestamp` MUST NOT substitute for commit time. Legacy clients may omit the fingerprint or commit time; unknown values remain unknown.
3. Baseline evaluation on the default branch:
   - Only runs with `status == "passed"` and `pr_number == 0` on the default branch are eligible to update the baseline.
   - An eligible run MUST NOT overwrite an existing baseline if its `commit_timestamp` is earlier than the existing baseline's `commit_timestamp`.
   - A run without commit time MUST NOT replace a baseline with known commit time.
   - Different known fingerprints MUST NOT be compared or replace one another. A missing fingerprint MUST NOT compare against or replace a known fingerprint.
   - A passing run with a known fingerprint may establish a trusted baseline over a legacy baseline lacking a fingerprint, subject to the chronology checks. That legacy baseline MUST NOT be returned as a compatible comparison.
   - Legacy runs whose fingerprints are both missing retain legacy comparison behavior; this does not establish known scenario compatibility.
4. Upgrade migration `2026-09-21-correct-commit-timestamp-provenance` MUST run once, inside a transaction, and clear `runs.commit_timestamp` for rows without a `run_ingestions` provenance record. Releases before corrected ingestion stored upload time in that column, so those values are not trusted commit times. Rows with corrected-ingestion provenance, and rows written after the migration is recorded, MUST be preserved across restarts.

## Public Metadata and Administration

1. Run list and detail responses MUST use snake_case and expose known `repo_full_name`, `scenario_name`, `driver`, `is_regression`, `total_schedules`, and `failed_schedules`. Legacy records without saved decisions or measurements return null for those values. Omitted measurements MUST NOT become measured zeros.
2. The saved regression decision MUST be returned without reevaluating against a newer baseline. SQL, traces, reproduction code, tokens, and webhook credentials MUST remain absent from general run metadata.
3. `POST /v1/organizations/me/tokens` requires admin or owner authorization and accepts only a `name`. It MUST generate a new member token, store its hash, return `token`, `role: "member"`, and `organization_id`, and set `Cache-Control: no-store`. Requests containing elevated role fields MUST be rejected.
4. Webhook creation and listing MUST return redacted destination URLs without user information, paths, query strings, or fragments. Saved destinations are tested through `POST /v1/organizations/me/webhooks/{wh_id}/test`; the server resolves the full destination by tenant and saved ID. Testing and deletion require admin or owner authorization, and another tenant's ID is not found.
5. Deleting a webhook MUST remove its queued delivery records in the same tenant-scoped transaction so foreign keys do not prevent deletion and the removed destination retains no queued alerts.

## Transactional Webhook Outbox

1. Alerts for regressions and run failures MUST NOT be dispatched via detached, unbounded goroutines.
2. When an ingestion triggers an alert event, the server MUST insert delivery records into the `webhook_outbox` table within the same transaction that persists the run.
3. The outbox dispatcher MUST deliver pending notifications asynchronously with bounded concurrency.
4. The outbox dispatcher MUST support automatic retry with exponential backoff for up to 3 attempts.
5. SSRF protection:
   - New and explicitly tested webhook destination URLs MUST use `https` and MUST NOT contain URL user credentials.
   - The dispatcher MUST resolve the destination hostname and reject requests targeting loopback (`127.0.0.0/8`, `::1`), private (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), or link-local (`169.254.0.0/16`, `fe80::/10`) addresses.
   - HTTP redirects MUST be validated through the same SSRF resolver before being followed.
   - Production dispatchers MUST use the guarded HTTP client by default. The dialer MUST connect to the validated literal IP address instead of resolving the original hostname again.
6. The outbox record MUST store delivery status (`pending`, `delivered`, `failed`), retry count, error details, and delivery timestamps.

## Plan Retention

1. Retention enforcement MUST be opt-in (`--enforce-retention` or `CHAOSSQL_ENFORCE_RETENTION=true`). Without it, no history is deleted.
2. When enabled, a pass MUST run at startup and then hourly. The clock MUST be injectable for tests.
3. For each organization with a finite `RetentionDays`, a pass MUST delete runs created before `now - RetentionDays`, with their findings and `run_ingestions` rows, in one transaction per organization.
4. A run referenced by any current baseline MUST NOT be deleted, regardless of age.
5. Only `delivered` or `failed` outbox alerts older than the window are deleted. `pending` alerts MUST be kept.
6. Repositories, scenarios, webhooks, tokens, and organizations are configuration and MUST NOT be deleted by retention.
7. A repeated pass with the same clock MUST delete nothing further.

## Verification

Tests MUST verify:
- Ingestion atomicity under simulated database failures.
- Idempotent deduplication under concurrent duplicate requests.
- Content conflicts, immutable replay responses, and distinct scenarios from one CI workflow.
- Protection against out-of-order commits overwriting baselines.
- Unknown and incompatible fingerprints, safe migration from legacy baselines, and faithful missing measurements.
- The one-time commit-timestamp provenance migration, including preservation of interim corrected rows and restarts.
- Concurrent repository quota enforcement, member-token authorization, and tenant-scoped saved webhook management.
- Persistent outbox queueing, worker retry, default SSRF blocking, and connection-time DNS pinning.
- Plan retention windows per organization, baseline preservation, pending-alert preservation, and idempotent repeated passes.
