---
name: chaossql-replay-artifacts
description: The executable replay artifact (run --export-result, version 1) and the replay command — payload fields, how it is built, strict validation rules, replay --verify re-execution, trace rendering, and the atomic 0600 file write. Use when changing the artifact format, replay verification, or the schedule format it pins.
---

# Replay Artifacts

A replay artifact freezes everything needed to re-execute a failure: the full
resolved spec, the seed, the (minimal) scheduled operations, the logical
schedule plan, the trace, and the failure signature. `chaossql replay
--verify` re-runs it and proves the same failure happens again.

## When to use

- Changing `ReplayPayload`, `buildReplayArtifact`, `validateReplayArtifact`,
  or `verifyReplayArtifact`.
- Bumping the schedule format (the artifact pins it).
- A replay fails verification.

## Producing (`run --export-result PATH`)

`buildReplayArtifact(spec, result, ops, trace, anomaly)`:
- fails if `FailureSignatureFor(result)` fails (only `passed`/`violation`
  results can be exported — `execution_error` etc. make the command fail);
- sets `spec.Engine.Seed = result.Seed`;
- stores copies of the (minimal, when shrinking succeeded) ops and trace;
- recomputes `Schedule = BuildSchedulePlan(spec, ops, NewPRNG(seed))` for
  those ops (so it describes the minimal ops, not the original run);
- `Version = 1`, plus `AnomalyType`, `ViolationDetected`, `FailingInvariant`,
  `Status`, `FailureSignature`.

`writeReplayArtifact` writes indented JSON + newline via
`writeRestrictedAtomic`: temp file in the target dir, `chmod 0600`, write,
fsync, close, rename. The Windows variant has its own implementation.

The artifact contains the **full spec including DSN, schema and seed SQL** —
treat it as sensitive.

## Replaying (`chaossql replay <file> [--verify] [--max-events 50]`)

1. Read JSON; if it is not a `ReplayPayload`, try a bare `ExecutionTrace` array.
2. Print the banner and a chronological table (event #, µs, worker, type, SQL),
   truncated to `--max-events`.
3. With `--verify`, `validateReplayArtifact`:
   - `Version == 1`, spec present and `Spec.Validate()` passes, ops non-empty;
   - `Schedule.Version == scheduleVersionForReplay` (1);
   - `Seed == Schedule.Seed == Spec.Engine.Seed`;
   - `FailureSignature.Status == Status`; for violations, the signature's
     invariant equals `FailingInvariant.Name`;
   - recomputed plan `reflect.DeepEqual` stored plan.
4. Then `verifyReplayArtifact`: `GetDriver` + `Open` from the stored spec,
   `RunSchedule(spec, ops)` (fresh reset), require an identical logical
   schedule and `ReproducesFailure(result, signature)`; prints
   `REPLAY VERIFIED: status=... seed=... operations=...`.

## Gotchas

- Verification checks the logical schedule and the failure signature, not the
  physical trace — timestamps and interleaving may differ.
- Real-database failures can be flaky; a `--verify` failure may be timing, not
  a format problem. Re-run before debugging the format.
- The stored DSN is used as-is on replay; PostgreSQL/MySQL resets are destructive.
- `replay` without `--verify` accepts legacy/partial JSON silently.

## Change checklist

- Format change → bump `replayArtifactVersion`, decide backwards compatibility,
  update tests and this skill.
- Schedule format change → bump `scheduleVersionForReplay` together with
  `schedulePlanVersion` (`chaossql-deterministic-schedule`).
- `ui` also reads result/trace JSON (`cmd/chaossql/ui.go` `loadTraceData`) —
  keep it compatible.

## Tests

- `cmd/chaossql/replay_test.go`, `cmd/chaossql/replay_file_unix_test.go`

## Source map

- `cmd/chaossql/replay.go`
- `cmd/chaossql/replay_file_unix.go`
- `cmd/chaossql/replay_file_windows.go`
- `cmd/chaossql/replay_test.go`
- `cmd/chaossql/main.go`
- `evals/03_deterministic_replay.md`
- `docs/adrs/0001-deterministic-prng-and-replay.md`

## Related skills

- `chaossql-deterministic-schedule`, `chaossql-ddmin-shrinker`,
  `chaossql-cli-run-pipeline`, `chaossql-report-exporters`
