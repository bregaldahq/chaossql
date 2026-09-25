---
name: chaossql-ingestion-baselines
description: Server-side run ingestion (POST /v1/runs) — strict metadata-only decoding, idempotency keys and request hashing with replayed responses and 409 conflicts, the single transaction that stores repository/scenario/run/finding/ingestion metadata, the regression engine and default-branch baselines, and alert enqueueing. Use when changing ingestion, idempotency, regression detection, or baselines.
---

# Ingestion, Idempotency and Baselines

## When to use

- Changing `handleIngestRun`, `Store.ingest`, `ingestionKey`,
  `RegressionEngine.Evaluate`, baselines, or `run_ingestions`.
- Debugging duplicate runs, 409s, or unexpected regression verdicts.

## Request handling (`handleIngestRun`)

1. Org from the principal; body limited to `cloud.MaxPayloadBytes` (64 KiB,
   413 when exceeded).
2. `cloud.DecodeMetadataOnlyRequest`: `DisallowUnknownFields`, exactly one JSON
   object, same identifier validation as the client → 400 on failure.
3. `ingestionKey(header, body)`: the `Idempotency-Key` header and the body
   `idempotency_key` must agree when both are present.
4. Measurements (`total_schedules`, `failed_schedules`) decoded separately as
   pointers to distinguish zero from absent.
5. `Store.ingest(...)` → 201 when created, 200 when replayed; 409 on
   `ErrIdempotencyConflict`; 403 on `ErrPlanLimitReached`; else 500.
6. On creation, trigger the outbox dispatcher.

## `Store.ingest` (one SQLite transaction)

1. `requestHash` = SHA-256 of `{request without idempotency key, measurements}`.
2. Repository from CI (`local/chaossql-project`, branch `main`, commit
   `unknown` when absent); `GetOrCreateRepo` enforces the plan's repository
   limit and records the default branch **on creation only** (CI base branch,
   else `main`).
3. Idempotency (key scoped per repository): same key + same hash → return the
   stored response verbatim; different hash → conflict; a key present only on
   a legacy run (no stored response) → conflict.
4. `GetOrCreateScenario(repo, name|"default", driver|"sqlite")`.
5. Insert run (random `run_` ID, status, anomaly type, seed, duration,
   fingerprint, commit timestamp) and, when a violation or reproduction exists,
   a finding with only anomaly type, failing invariant **name**, and minimal
   op count.
6. `RegressionEngine.Evaluate` (below) → response
   `{success, run_id, url: <public>/dashboard?run=<id>, is_regression, baseline, message}`.
7. Store `run_ingestions(run_id, request_hash, response_json, driver, totals)`.
8. If regression or not passed: enqueue outbox items for active webhooks whose
   events include `regression`/`failure` or `all`.

## Regression engine (`regression.go`)

- Default branch = repo default (else `main`); a run is on it iff
  `branch == default && pr_number == 0`.
- Baseline = latest baseline run for (repo, scenario, default branch).
- Fingerprints: compatible iff no baseline or equal fingerprints.
- Passed run on the default branch updates the baseline unless: incompatible
  fingerprint vs a fingerprinted baseline, the baseline has a commit timestamp
  and this run does not, or this run's commit is older than the baseline's.
  It never counts as a regression.
- Other runs: no baseline or incompatible → no regression; otherwise
  regression iff baseline `passed` and run `failed`.

## Gotchas

- `failed` includes execution errors, so an infrastructure failure on a PR
  counts as a regression against a passing baseline.
- Plan `has_regression_gating` is not enforced — every plan gets verdicts.
- The response is frozen at first ingestion; retries return it even if the
  baseline moved since.
- Idempotency keys are per repository, not per organization.

## Tests

- `internal/server/ingestion_regression_test.go`,
  `internal/server/ingestion_upgrade_test.go`, `internal/server/regression_test.go`,
  `internal/server/handlers_test.go`, `internal/cloud/reliability_test.go`

## Source map

- `internal/server/ingestion.go`
- `internal/server/regression.go`
- `internal/server/run_metadata.go`
- `internal/server/handlers.go`
- `internal/server/store.go`
- `internal/cloud/payload.go`
- `specs/20_saas_reliable_ingestion_and_baselines.md`

## Related skills

- `chaossql-cloud-publishing`, `chaossql-control-plane-api`,
  `chaossql-webhooks-outbox`, `chaossql-store-migrations`, `chaossql-plans-retention`
