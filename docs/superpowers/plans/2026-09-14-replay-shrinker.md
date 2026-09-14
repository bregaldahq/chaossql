# ENG-03 Executable Replay and Trustworthy Shrinking Plan

**Goal:** Ensure minimization preserves the original failure and produce a portable artifact that can reexecute and verify the same finding.

**Design:** `docs/superpowers/specs/2026-09-14-replay-shrinker-design.md`

### Task 1: Stable failure identity

**Files:** `internal/domain/types.go`, `internal/shrinker/ddmin.go`, tests

- [ ] Add failing tests for same-invariant matching and different-failure rejection.
- [ ] Define the stable failure signature and shared reproduction predicate.
- [ ] Update every shrink call site to use the original failure signature.
- [ ] Run focused domain, shrinker, CLI, IPC, and WASM tests.
- [ ] Commit the failure identity change.

### Task 2: Auditable 1-minimal shrinking

**Files:** `internal/shrinker/ddmin.go`, `internal/shrinker/ddmin_test.go`, `internal/domain/types.go`

- [ ] Add failing tests for cancellation before the first oracle call, deterministic ties, explicit 1-minimality, and trial counts.
- [ ] Count actual oracle trials and make cancellation checks surround oracle execution.
- [ ] Add a final single-removal audit that converges to a 1-minimal set.
- [ ] Run focused tests repeatedly under the race detector.
- [ ] Commit the shrinker auditability change.

### Task 3: Versioned replay artifact export

**Files:** `cmd/chaossql/replay.go`, `cmd/chaossql/main.go`, tests

- [ ] Add failing round-trip tests for a complete version 1 artifact.
- [ ] Add `run --export-result <path>` and write the minimal reproducible artifact atomically.
- [ ] Rebuild and persist the logical schedule for the exported operation subset.
- [ ] Preserve legacy trace-only replay compatibility.
- [ ] Commit replay artifact export.

### Task 4: Executable replay verification

**Files:** `cmd/chaossql/replay.go`, `cmd/chaossql/replay_test.go`

- [ ] Add failing tests for successful SQLite reexecution and tampered artifacts.
- [ ] Add `replay --verify` with artifact validation, database reset, schedule comparison, and failure-signature comparison.
- [ ] Return actionable errors for unsupported versions and missing execution inputs.
- [ ] Run replay and CLI integration tests under the race detector.
- [ ] Commit executable replay.

### Task 5: Specifications and regression evidence

**Files:** `specs/03_delta_debugging_shrinker.md`, `specs/04_evidence_synthesis.md`, `evals/01_shrinking_ratio.md`, `evals/03_deterministic_replay.md`

- [ ] Document failure identity, trial metrics, artifact schema, and physical timing boundary.
- [ ] Add deterministic fixtures proving repeated replay and minimization preserve the same failure.
- [ ] Run documentation and compatibility checks.
- [ ] Commit specifications and fixtures.

### Task 6: Final audit, review, PR, and merge

- [ ] Run `gofmt`, `git diff --check`, race tests, and `make verify`.
- [ ] Run zero-CGO native and WASM builds.
- [ ] Request independent review and resolve every Critical and Important finding.
- [ ] Push `codex/eng-03-replay-shrinker`, open a PR to `main`, and wait for all checks.
- [ ] Mark this plan complete, merge with a merge commit, and mark ENG-03 complete in the ignored commercialization plan.
