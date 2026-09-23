# Cloud integration corrections

The user approved correcting the September 17 review. This change repairs the existing product contract across ingestion, dashboard, client, and self-hosted distribution. It does not add badges, automatic SQL repair, payment-provider integration, or a new hosted identity provider.

## Required behavior

1. Each CLI execution has a unique ingestion key, retained across HTTP retries. Different scenarios, jobs and attempts cannot silently overwrite one another. Reusing an explicit key for different content returns a conflict. Legacy callers without a key remain supported without interpreting a whole CI workflow as a single execution.
2. Run, finding, baseline decision and required outbox records commit atomically. Database errors propagate. Retrying an acknowledged execution returns its original regression decision. Baseline selection uses actual commit time and scenario fingerprints, never upload time. Missing legacy identity must not override a trusted baseline.
3. The client projects only safe metadata, including an explicit idempotency key and CI commit timestamp. It computes a stable fingerprint from scenario semantics without transmitting SQL. Defaults retry transient failures, preserve cancellation, and do not retry terminal client errors. The GitHub Action exposes run ID, run URL, and regression outputs and treats inputs as data.
4. Dashboard requests authenticate with a user-supplied token kept in component/session memory, not browser persistent storage. Public API responses use snake_case. Measured metadata is displayed faithfully; unknown values are unavailable rather than replaced with demo values. The API exposes repository name, scenario name, driver and the saved regression decision where available. Organization-scoped webhook routes and their actual request/response types are used.
5. Onboarding issues a real member CI token through an admin-authorized endpoint, or explains the existing CLI command when the user lacks permission. Tokens and full webhook URLs do not enter general metadata responses. Hosted SQL and reproduction code remain forbidden. Finding links must not masquerade sample traces as live CI evidence.
6. Pricing remains an explicit request-for-contact flow. It confirms success only after acknowledged server delivery, preserves selected plan and audit intent, and reports failures. Existing repository plan limits are enforced atomically; payment subscriptions remain outside this corrective patch.
7. Both server entrypoints share owner-token bootstrap, require configured credentials and propagate errors. Production templates contain no usable fixed token. SQLite is supported as a single writer replica; unsupported autoscaling fails validation. The dashboard image runs nonroot and serves its WASM assets. Documentation includes a usable local startup and backup/restore procedure.
8. Every confirmed bug receives a focused regression test. `CGO_ENABLED=0` builds and `make verify` must pass. Ignored local planning documents do not invalidate the English source gate. Real SQL driver tests are exercised against disposable databases when available.

## Shared interfaces

- `cloud.RunIngestRequest.IdempotencyKey string` serializes as optional `idempotency_key`; HTTP `Idempotency-Key` is the transport override. Conflicting body/header values are rejected.
- `cloud.CIContext.CommitTimestamp *time.Time` serializes as optional `commit_timestamp`; the safe metadata projection/decoder preserves it.
- Public run fields remain snake_case and gain `repo_full_name`, `scenario_name`, `driver`, `is_regression`, `total_schedules`, `failed_schedules` when known. The latter values must not be invented for legacy records.
- `POST /v1/organizations/me/tokens` requires admin and accepts `{ "name": "CI Token" }`; it creates only a member token, returns `{ "token": "...", "role": "member", "organization_id": "..." }`, and applies `Cache-Control: no-store`.
- Webhook CRUD uses `/v1/organizations/me/webhooks`, `target_type` and comma-separated `events`. GET returns the existing array of redacted records. Test existing destinations by their ID, not by exposing saved secrets.

## Implementation boundaries

Backend owns `internal/server` except shared bootstrap files. Client owns `internal/cloud`, `cmd/chaossql/main.go` and `action.yml`. Frontend owns `site/src` and frontend tests. Integration owns bootstrap entrypoints, Docker/Helm/nginx, workers, quality scripts and documentation. Independent implementation is followed by focused review and a whole-change verification.
