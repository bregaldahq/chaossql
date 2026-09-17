# API-01 Reliable Ingestion, Baselines, and Alert Outbox Implementation Plan

**Goal:** Ensure hosted run ingestion is atomic, idempotent under retries, preserves baseline chronology, and persists alert dispatches to an outbox queue.

**Design:** `docs/superpowers/specs/2026-09-17-idempotent-ingestion-design.md`
**Spec:** `specs/20_saas_reliable_ingestion_and_baselines.md`

### Task 1: Spec and schema migration
- [x] Add failing schema migration tests in `internal/server/store_test.go` verifying `idempotency_key`, `scenario_fingerprint`, `commit_timestamp`, and `webhook_outbox`.
- [x] Implement migration `2026-09-17-idempotent-runs-and-outbox` in `internal/server/store.go`.
- [x] Run focused store tests and commit.

### Task 2: Atomic and idempotent persistence
- [x] Add failing tests in `internal/server/store_test.go` for `SaveRunTx` atomicity and idempotency.
- [x] Implement `SaveRunTx` with transaction and idempotent duplicate prevention in `internal/server/store.go`.
- [x] Run focused store tests and commit.

### Task 3: Chronological and fingerprint-aware baseline
- [x] Add failing tests in `internal/server/regression_test.go` for out-of-order commit protection and fingerprint checks.
- [x] Implement baseline chronology and fingerprint validation in `internal/server/regression.go`.
- [x] Run regression tests and commit.

### Task 4: Ingestion handler integration
- [x] Add failing tests in `internal/server/handlers_test.go` for `Idempotency-Key` header, duplicate submissions, and outbox staging.
- [x] Update `handleIngestRun` in `internal/server/handlers.go` to use atomic transaction and transactional outbox.
- [x] Run handler tests and commit.

### Task 5: Outbox dispatcher with SSRF protection
- [x] Add failing tests for outbox worker processing, retries, and SSRF address rejection.
- [x] Implement `OutboxDispatcher` and SSRF-safe HTTP client in `internal/server/webhooks.go`.
- [x] Run webhook tests and commit.

### Task 6: Final verification and quality gate
- [x] Run `gofmt`, race tests, and `make verify`.
- [x] Update commercialization assessment plan.