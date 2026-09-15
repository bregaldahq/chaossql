# ENG-03 Executable Replay and Trustworthy Shrinking Plan

**Goal:** Ensure minimization preserves the original failure and produce a portable artifact that can reexecute and verify the same finding.

**Design:** `docs/superpowers/specs/2026-09-14-replay-shrinker-design.md`

### Task 1: Stable failure identity

**Files:** `internal/domain/types.go`, `internal/shrinker/ddmin.go`, tests

- [x] Add failing tests for same-invariant matching and different-failure rejection.
- [x] Define the stable failure signature and shared reproduction predicate.
- [x] Update every shrink call site to use the original failure signature.
- [x] Run focused domain, shrinker, CLI, IPC, and WASM tests.
- [x] Commit the failure identity change.

### Task 2: Auditable 1-minimal shrinking

**Files:** `internal/shrinker/ddmin.go`, `internal/shrinker/ddmin_test.go`, `internal/domain/types.go`

- [x] Add failing tests for cancellation before the first oracle call, deterministic ties, explicit 1-minimality, and trial counts.
- [x] Count actual oracle trials and make cancellation checks surround oracle execution.
- [x] Add a final single-removal audit that converges to a 1-minimal set.
- [x] Run focused tests repeatedly under the race detector.
- [x] Commit the shrinker auditability change.

### Task 3: Versioned replay artifact export

**Files:** `cmd/chaossql/replay.go`, `cmd/chaossql/main.go`, tests

- [x] Add failing round-trip tests for a complete version 1 artifact.
- [x] Add `run --export-result <path>` and write the minimal reproducible artifact atomically.
- [x] Rebuild and persist the logical schedule for the exported operation subset.
- [x] Preserve legacy trace-only replay compatibility.
- [x] Commit replay artifact export.

### Task 4: Executable replay verification

**Files:** `cmd/chaossql/replay.go`, `cmd/chaossql/replay_test.go`

- [x] Add failing tests for successful SQLite reexecution and tampered artifacts.
- [x] Add `replay --verify` with artifact validation, database reset, schedule comparison, and failure-signature comparison.
- [x] Return actionable errors for unsupported versions and missing execution inputs.
- [x] Run replay and CLI integration tests under the race detector.
- [x] Commit executable replay.

### Task 5: Specifications and regression evidence

**Files:** `specs/03_delta_debugging_shrinker.md`, `specs/04_evidence_synthesis.md`, `evals/01_shrinking_ratio.md`, `evals/03_deterministic_replay.md`

- [x] Document failure identity, trial metrics, artifact schema, and physical timing boundary.
- [x] Add deterministic fixtures proving repeated replay and minimization preserve the same failure.
- [x] Run documentation and compatibility checks.
- [x] Commit specifications and fixtures.

### Task 6: Final audit, review, PR, and merge

- [x] Run `gofmt`, `git diff --check`, race tests, and `make verify`.
- [x] Run zero-CGO native and WASM builds.
- [x] Request independent review and resolve every Critical and Important finding.
- [ ] Push `codex/eng-03-replay-shrinker`, open a PR to `main`, and wait for all checks.
- [ ] Mark this plan complete, merge with a merge commit, and mark ENG-03 complete in the ignored commercialization plan.
