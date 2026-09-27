---
name: chaossql-deterministic-schedule
description: How ChaosSQL turns a spec and a seed into a reproducible logical schedule (GenerateSchedule + BuildSchedulePlan), including worker assignment, per-step jitter/latency/abort decisions, schedule plan version 1 and the golden corpus. Use when touching schedule generation, PRNG streams, worker partitioning, or when a seed stops reproducing.
---

# Deterministic Schedule

Determinism is principle #1 of `AGENTS.md`. What is deterministic is the
**logical schedule**: which operations run, with which parameters, on which
worker, and with which injected delays/aborts. The physical interleaving on a
real database (and trace timestamps) still depends on timing.

## When to use

- Changing `GenerateSchedule`, `BuildSchedulePlan`, `assignedWorker`,
  `partitionScheduledOps`, or `decisionRand`.
- A failing seed no longer reproduces, or two runs with one seed disagree.
- The golden corpus test fails.

## Stage 1 — operations (`GenerateSchedule`, `internal/engine/runner.go`)

- `iterations <= 0` → 10 operations; no operation templates → `nil`.
- Fresh `PRNG` from the master seed + `masterRng = rand.New(rand.NewPCG(seed, 0))`.
- For `i` in `0..N-1`: template = `Operations[masterRng.IntN(len)]` (uniform —
  `weight` is ignored), params evaluated in sorted name order from the same
  `masterRng`, result `ScheduledOp{ID: i+1, Name, Params, Steps}`.
- `Steps` shares the template's slice (do not mutate it).

## Stage 2 — decisions (`BuildSchedulePlan`, `internal/engine/schedule.go`)

- `workers <= 0` → 4. Ops are sorted by ID (input order does not matter).
- Worker: `assignedWorker(id, W) = ((id-1) mod W) + 1` — round-robin by **ID**.
- One `ScheduleDecision` per step (1-based `StepIndex`) with a global
  `Sequence`, `JitterMs`, `LatencyMs`, `Abort`.
- Each decision field comes from its own RNG:
  `decisionRand(seed, opID, stepIndex, stream)` builds a PCG from
  `mix64(seed ^ opID*φ ^ step*c1)` and `mix64(first ^ stream*c2)`; streams are
  1 = jitter, 2 = latency, 3 = abort.
- Jitter: `min + IntN(max-min+1)` when `max > 0 && max >= min`, else 0.
- Latency/abort rules: see `chaossql-fault-injection`.
- `SchedulePlan{Version: 1, Seed, Workers, Decisions}`; `Decisions` encodes as
  `[]` (never `null`) when empty.

### Why decisions depend on (seed, opID, step) only

ddmin replays **subsets** of the original operations. Because a decision is a
pure function of the op's identity and the worker is a function of its ID,
removing other operations never changes the surviving operations' worker,
jitter, latency or abort. That is what makes shrinking candidates comparable.

## Stage 3 — execution queues (`partitionScheduledOps`)

`ExecuteSchedule` rebuilds the plan, validates IDs (positive, unique), and gives
each worker the ID-sorted list of its operations; each worker runs its queue
sequentially, one transaction per operation. See `chaossql-runner-execution`.

## Contracts

- Same spec + same seed ⇒ identical `[]ScheduledOp` and `SchedulePlan`
  (checked byte-for-byte by the golden corpus).
- Seed `0` is a literal seed, not "random".
- `PRNG.WorkerSeed` and `PRNG.Jitter` exist but are only used by `bench`.
- The plan is persisted in results (`ExecutionResult.Schedule`), JSON output,
  replay artifacts (validated with `reflect.DeepEqual`), and Cloud payloads
  (the client drops it from hosted metadata).

## Changing the algorithm (format change)

Any change that alters generated ops or decisions for an existing seed is a
schedule format change:
1. Bump `schedulePlanVersion` in `internal/engine/schedule.go`.
2. Bump `scheduleVersionForReplay` in `cmd/chaossql/replay.go` (replay
   artifacts pin version 1 today) and decide how old artifacts are handled.
3. Regenerate the corpus intentionally:
   `UPDATE_GOLDEN=1 go test ./internal/engine -run TestDeterministicScheduleV1GoldenCorpus`
   (consider a new `deterministic_schedule_v2.json` instead of overwriting v1).
4. Update `docs/adrs/0001-deterministic-prng-and-replay.md` / a new ADR, and
   the replay and runner skills.

Never introduce `time.Now()`, `math/rand` globals, map iteration order, or
goroutine completion order into these functions.

## Tests

- `internal/engine/schedule_test.go` (determinism, order independence, empty
  decisions as array, golden corpus)
- `internal/engine/runner_schedule_test.go` (worker queues, invalid IDs,
  precomputed aborts, persisted seed and schedule)
- `internal/engine/prng_test.go` (`TestGenerateSchedule_*`, seed zero)

## Source map

- `internal/engine/schedule.go`
- `internal/engine/runner.go`
- `internal/engine/prng.go`
- `internal/engine/testdata/deterministic_schedule_v1.json`
- `internal/engine/schedule_test.go`
- `internal/engine/runner_schedule_test.go`
- `cmd/chaossql/replay.go`
- `docs/adrs/0001-deterministic-prng-and-replay.md`
- `docs/adrs/0002-async-step-interleaving.md`

## Related skills

- `chaossql-param-generators`, `chaossql-runner-execution`,
  `chaossql-fault-injection`, `chaossql-ddmin-shrinker`, `chaossql-replay-artifacts`
