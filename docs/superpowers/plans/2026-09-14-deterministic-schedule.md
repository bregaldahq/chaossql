# ENG-02 Deterministic Schedule Implementation Plan

**Goal:** Make parameter generation, worker assignment, injected timing/fault choices, and the persisted logical schedule reproducible for the same specification and seed.

**Design:** `docs/superpowers/specs/2026-09-14-deterministic-schedule-design.md`

### Task 1: Deterministic seed and parameter generation

**Files:** `internal/engine/prng.go`, `internal/engine/prng_test.go`, `internal/engine/runner.go`

- [x] Add failing tests proving seed zero remains zero, randomized map parameters are stable over 100 generations, and independent runs reset monotonic counters.
- [x] Move scheduler monotonic counters into each `PRNG` instance.
- [x] Sort parameter names before generator evaluation.
- [x] Run `go test -race ./internal/engine -run 'PRNG|GenerateSchedule' -count=1`.
- [x] Commit the deterministic generation change.

### Task 2: Versioned logical schedule

**Files:** `internal/domain/types.go`, `internal/domain/types_test.go`, `internal/engine/schedule.go`, `internal/engine/schedule_test.go`

- [x] Add failing JSON and deterministic derivation tests for `SchedulePlan` and `ScheduleDecision`.
- [x] Derive worker, jitter, latency, and abort decisions from seed plus operation/step identity.
- [x] Assert byte-identical plans across repeated construction and changed completion timing.
- [x] Run domain and schedule tests under the race detector.
- [x] Commit the schedule artifact.

### Task 3: Deterministic execution queues

**Files:** `internal/engine/runner.go`, `internal/engine/runner_schedule_test.go`

- [ ] Add failing tests that expose contested channel assignment and shared fault decision order.
- [ ] Replace the shared operation channel with deterministic per-worker queues.
- [ ] Execute each step using its precomputed schedule decision.
- [ ] Preserve prompt cancellation and structured transaction errors.
- [ ] Run transaction, schedule, status, and cancellation tests repeatedly under `-race`.
- [ ] Commit deterministic execution.

### Task 4: Persist schedule and effective seed

**Files:** `internal/domain/types.go`, `internal/engine/runner.go`, CLI IPC/JSON, cloud, reporters, Go/Python/TypeScript SDK adapters and tests

- [ ] Add failing compatibility tests for `seed` and `schedule` fields.
- [ ] Carry the effective seed and plan through `ScheduleOutcome` and `ExecutionResult`.
- [ ] Preserve the additive fields across public adapters and full-result reports.
- [ ] Run CLI, cloud, reporter, and SDK tests.
- [ ] Commit public result persistence.

### Task 5: Specification and evaluation corpus

**Files:** `specs/02_concurrency_interleaving.md`, `evals/03_deterministic_replay.md`, deterministic corpus fixtures

- [ ] Document the logical-schedule guarantee and external database timing boundary.
- [ ] Add a versioned golden corpus covering map parameters, counters, jitter, latency, aborts, and zero seed.
- [ ] Prove the corpus is identical across independent process executions.
- [ ] Commit specifications and corpus.

### Task 6: Final audit, review, PR, and merge

- [ ] Run `gofmt`, `git diff --check`, `go test -race ./... -count=1`, and `make verify`.
- [ ] Run zero-CGO native and WASM builds.
- [ ] Request independent review and resolve every Critical and Important finding.
- [ ] Push `codex/eng-02-deterministic-schedule`, open a PR to `main`, and wait for all checks.
- [ ] Mark this plan complete, merge with a merge commit, and mark ENG-02 complete in the ignored commercialization plan.
