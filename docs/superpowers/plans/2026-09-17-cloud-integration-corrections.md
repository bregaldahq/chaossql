# Cloud Integration Corrections Implementation Plan

> **For agentic workers:** Use superpowers:dispatching-parallel-agents for the independent file domains below, test-driven development for behavior changes, and requesting-code-review before completion.

**Goal:** Resolve the confirmed September 17 integration and self-hosted bugs with executable regressions.

**Architecture:** Keep the existing metadata-only API and pure-Go SQLite store. Make ingestion a transaction boundary, align the browser and CLI with the real API, and share administrative bootstrap between entrypoints.

**Tech Stack:** Go 1.25+, modernc SQLite, React/TypeScript/Vitest, Cloudflare worker, Docker Compose, Helm.

**Spec:** `docs/superpowers/specs/2026-09-17-cloud-integration-corrections-design.md`.

## Global constraints

- All binaries and WASM build with `CGO_ENABLED=0`.
- No SQL, parameter values, schema, local reproduction code or credentials in hosted run metadata.
- Prefix shell commands with `rtk` on the host; use its proxy for WSL commands.
- Do not publish, submit external leads, send live alerts, modify existing database services, or merge into main during implementation.
- Preserve the existing design and bilingual interface. Unknown live values must not be demo substitutions.

## Task 1: Reliable backend and dashboard API

Files: `internal/server/store.go`, `handlers.go`, `regression.go`, corresponding tests, new focused ingestion files as needed, `specs/20_saas_reliable_ingestion_and_baselines.md`.

- [x] Add regressions from the review: two scenarios share `CI.RunID` but persist two runs; injected outbox insertion failure rolls back; retry after recovery creates one alert.
- [x] Test concurrent duplicate requests, idempotency conflicts, replay of original response, old commits arriving late, incompatible/missing fingerprints, and transaction rollback.
- [x] Implement transactional ingestion and persisted comparison metadata without bypassing tenant roles or safe payload validation.
- [x] Add faithful public metadata and admin-only member-token creation with no-store and authorization tests.
- [x] Enforce repository limits under concurrent creation; return a meaningful limit response and preserve existing repositories at capacity.
- [x] Run `go test -race -count=1 ./internal/server` and record results.

Regression shape:
```go
// Two distinct payloads from a single CI workflow must not share an ID.
first := postRun("scenario_one", "workflow-123")
second := postRun("scenario_two", "workflow-123")
if first.RunID == second.RunID { t.Fatal("workflow is not an execution identity") }
// A trigger rejecting webhook_outbox INSERT must leave zero runs/findings.
```

## Task 2: Official client and Action

Files: `internal/cloud/{types,payload,ci,client}.go`, tests, `cmd/chaossql/main.go`, CLI tests, `action.yml`.

- [x] Add failing tests for default retries after HTTP 503, stable retry key, distinct execution keys, cancellation, terminal 4xx, safe commit timestamp and fingerprint projection.
- [x] Implement the shared fields from the design. Obtain commit time for the actual SHA. Hash stable scenario semantics; do not upload input SQL.
- [x] Emit GitHub Action outputs from successful publication without embedding secrets or untrusted strings into shell source. Preserve execution failures and cloud-fail-fast in text and JSON modes.
- [x] Run `go test -race -count=1 ./internal/cloud ./cmd/chaossql` and targeted Action checks.

## Task 3: Authenticated dashboard and truthful product feedback

Files: `site/src/pages/DashboardPage.tsx`, `VisualizerPage.tsx`, `PricingPage.tsx`, focused `site/src/lib` API and test modules; frontend test setup when required.

- [x] Add behavior tests for authenticated requests, snake_case mapping, empty data, 401/403, HTTP error propagation, real webhook IDs and token issuance.
- [x] Extract a typed API client and replace the old mapper, unscoped webhook calls and demo token.
- [x] Display known values accurately; label summary scope. Offer local artifacts honestly; prevent CI links from rendering unrelated demo traces as evidence.
- [x] Await lead acknowledgment and retain plan/billing/audit intent in requests. Label pricing as contact/request flow.
- [x] Run frontend typecheck, build and tests. Exercise successful and failed UI paths locally. (Covered by component tests, 30/30; no manual browser review was performed.)

## Task 4: Self-hosted distribution, lead delivery and quality integration

Files: `cmd/chaossql/server.go`, `cmd/chaossql-server/main.go`, shared `internal/server/bootstrap.go`, tests, Docker/Compose/Helm/nginx, `worker.ts`, `site/_worker.js`, waitlist functions, quality scripts and operational docs.

- [x] Add regression tests for both bootstrap entrypoints: configured token is owner, rotated token revokes old credential, empty/default credentials fail safely.
- [x] Remove fixed secrets, require operator configuration, include `/wasm`, run nginx nonroot and reject unsupported multi-replica SQLite deployments.
- [x] Test waitlist request parsing and rejected downstream delivery without contacting external channels; preserve audit/plan selection across all worker entrypoints.
- [x] Exclude ignored confidential local plans from source language checks without ignoring shipped product code.
- [x] Validate Compose, render/lint Helm, build and smoke-test Docker with synthetic secrets and disposable local data.
- [x] Document startup, member tokens, single-replica limit, safe backup/restore and current billing scope.

## Final verification and review

- [x] Run `make verify` and explicit zero-CGO builds.
- [x] Run PostgreSQL/MySQL driver tests against disposable databases; avoid existing services.
- [x] Validate authenticated first run, baseline, PR regression and persisted alert using only local controlled endpoints. (Compose smoke covers run/baseline/regression/replay; alert persistence is covered by server outbox tests with controlled transports.)
- [x] Obtain independent review of the full diff, address important findings and rerun affected checks.
- [x] Update this checklist, record test evidence and report the completed branch and remaining external deployment limitations.

## Execution log

- Baseline: c432837; clean source passed `make verify` in the preceding review. Worktree: `.worktrees/cloud-integration-fixes`; branch: `codex/cloud-integration-fixes`.
- Scope decision: repairs to the existing token-based workflow are authorized by the user's approval of the report. A new OAuth provider, payment provider or public deployment requires separate configuration and is not implied by these bug fixes.
- Commit-timestamp provenance: review found legacy rows stored upload time as commit time. One-time transactional migration `2026-09-21-correct-commit-timestamp-provenance` clears untrusted values and preserves corrected rows; `TestUpgradeClearsLegacyUploadTimeAndPreservesNewCommitTime` and `TestUpgradePreservesCorrectedIngestionProvenance` pass with `-race`. Follow-up review of the migration found no issues. Spec 20 documents it.
- `make verify` passed (exit 0) with disposable PostgreSQL 16 and MySQL 8.4.2 containers, including real driver integration tests, the upgrade tests and 30 frontend tests. Disposable databases were removed; existing services were untouched.
- September 22 final checks: `site` rebuilt (bundle `main-BNEVPadk.js` includes the `pricing_page` lead source; build is reproducible). API and dashboard images rebuilt from the final tree; Compose smoke passed: nonroot users, same-origin API, WASM, 401/403 roles, member tokens, baseline, PR regression, idempotent replay, 409 conflict, online SQLite backup, separate-volume restore and owner token rotation. Disposable Compose resources removed.
- Helm v3.17.3 (checksum verified): lint and render pass with a token, an existing Secret, and with persistence disabled (emptyDir). Rendering fails as intended without credentials, with `replicaCount=2`, and with autoscaling enabled. No cluster installation was performed.
- `CGO_ENABLED=0` builds of `cmd/chaossql` and `cmd/chaossql-server` produce statically linked binaries; WASM is built by the gate.
- Not performed: public deployment, cluster installation, manual browser review, live lead/webhook delivery, push or merge. Known non-blocking warnings: ~609 kB frontend bundle and two moderate npm advisories.
