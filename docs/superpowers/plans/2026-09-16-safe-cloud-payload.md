# SEC-02 Safe Cloud Payload Implementation Plan

**Goal:** Ensure hosted ingestion cannot receive or persist SQL, customer values, schemas, or reproduction artifacts under the default product contract.

**Design:** `docs/superpowers/specs/2026-09-16-safe-cloud-payload-design.md`

### Task 1: Metadata allowlist

- [x] Add failing capture tests with representative secrets and personal data in every forbidden field.
- [x] Add an immutable metadata-only projection at the HTTP client boundary.
- [x] Make structural traces omit SQL and schema identifiers.
- [x] Run focused cloud tests and commit.

### Task 2: Server-side privacy boundary

- [x] Add failing tests for detailed fields, unknown fields, and oversized payloads.
- [x] Reject forbidden content before repository or run persistence.
- [x] Enforce strict JSON decoding and a 64 KiB request limit.
- [x] Run server and command integration tests and commit.

### Task 3: Contract and product claims

- [x] Add the normative hosted payload specification.
- [x] Correct README privacy language to match proven behavior.
- [x] Update client, server, serialization, and PR report compatibility tests.
- [x] Run focused tests and commit.

### Task 4: Final audit, review, PR, and merge

- [x] Run `gofmt`, `git diff --check`, race tests, and `make verify`.
- [x] Run zero-CGO native and WASM builds.
- [x] Request independent review and resolve every Critical and Important finding.
- [x] Push `codex/sec-02-safe-cloud-payload`, open a PR to `main`, and wait for all checks.
- [x] Mark this plan complete, merge with a merge commit, and mark SEC-02 complete in the ignored commercialization plan.
