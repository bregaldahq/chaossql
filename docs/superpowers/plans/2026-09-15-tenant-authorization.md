# SEC-01 Tenant Authorization Implementation Plan

**Goal:** Enforce authentication, role checks, and organization ownership across the SaaS API and persistence layer.

**Design:** `docs/superpowers/specs/2026-09-15-tenant-authorization-design.md`

### Task 1: Identity and role contract

- [x] Add failing tests for principal lookup, invalid roles, and legacy token migration.
- [x] Add token roles, principal authentication, and context-only request identity.
- [x] Create bootstrap tokens explicitly as owners.
- [x] Run focused store and server tests and commit.

### Task 2: Tenant-scoped persistence

- [x] Add failing two-organization tests for duplicate repository names and scoped run/repository lookups.
- [x] Scope repository uniqueness and queries by organization.
- [x] Add tenant-aware run listing and detail methods.
- [x] Run store and regression tests and commit.

### Task 3: Endpoint authorization matrix

- [x] Add table-driven tests covering missing credentials, cross-tenant access, and member/admin roles on every route.
- [x] Protect all tenant-data routes and enforce ownership from the context principal.
- [x] Return consistent `401`, `403`, and `404` responses.
- [x] Run handler and webhook tests and commit.

### Task 4: Explicit local router and specification

- [x] Add tests proving the SaaS router never enables short unauthenticated routes.
- [x] Add a separate local router with an explicit `org_default` owner identity.
- [x] Document the public authorization matrix and storage invariants.
- [x] Run compatibility tests and commit.

### Task 5: Final audit, review, PR, and merge

- [x] Run `gofmt`, `git diff --check`, race tests, and `make verify`.
- [x] Run zero-CGO native and WASM builds.
- [x] Request independent review and resolve every Critical and Important finding.
- [ ] Push `codex/sec-01-tenant-authorization`, open a PR to `main`, and wait for all checks.
- [ ] Mark this plan complete, merge with a merge commit, and mark SEC-01 complete in the ignored commercialization plan.
