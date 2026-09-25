---
name: chaossql-fault-injection
description: Stochastic fault injection (engine.faults) — how abort_probability, latency_probability and latency_spike_ms become deterministic per-step schedule decisions, what the runner does with them, and the unused legacy internal/faults injector and disconnect_probability. Use when configuring or changing faults.
---

# Fault Injection

Faults perturb the schedule to provoke more interleavings and exercise
rollback paths, without breaking determinism: every fault is a precomputed,
seed-derived decision in the schedule plan.

## When to use

- Configuring `engine.faults` in a scenario.
- Changing how aborts or latency spikes are decided or applied.
- Wondering why `disconnect_probability` does nothing.

## Configuration (`domain.FaultConfig`)

```yaml
engine:
  faults:
    abort_probability: 0.1       # per step
    latency_probability: 0.2     # per step
    latency_spike_ms: [5, 50]    # spike range
    disconnect_probability: 0.0  # parsed only
```

## Decision rules (`internal/engine/schedule.go`)

For every step of every scheduled op, using `decisionRand(seed, opID, step, stream)`:
- **Latency** (stream 2): if `latency_probability <= 0` → 0. Otherwise draw
  `Float64()`; if `>= probability` → 0. Else clamp `min` to `>= 0`; if
  `max <= min` then `max = min + 10`; spike = `min + IntN(max-min+1)` ms.
- **Abort** (stream 3): `abort_probability > 0 && Float64() < abort_probability`.
- Jitter (stream 1) is independent of faults.

## Runner behavior (`executeOperation`)

Before each step: wait jitter, then wait latency (both context-aware), then if
`Abort` → roll back the transaction immediately. An intentional abort:
- records a `ROLLBACK` event for that step;
- does **not** record an `OperationError` — the run can still be `passed` or
  `violation`;
- counts toward the temporal `no_aborts` invariant (library only).

## Legacy and unused pieces

- `internal/faults.FaultInjector` (math/rand seeded with `seed + 99991`,
  mutex-guarded `ShouldAbort`/`GetLatencySpike`/`ShouldDisconnect`) is not
  wired into any execution path; only its own tests use it. Its draws depend on
  call order, which is why the engine moved to per-step decisions.
- `disconnect_probability` is read by nothing but that legacy injector.

If you implement disconnects, add a new decision stream (4) in
`BuildSchedulePlan`, a new `ScheduleDecision` field, runner handling, and treat
it as a schedule format change (`chaossql-deterministic-schedule`).

## Tests

- `internal/engine/runner_schedule_test.go` (`TestExecuteSchedule_UsesPrecomputedAbortDecisions`)
- `internal/engine/runner_transaction_test.go` (`TestRunner_IntentionalAbortRollsBackWithoutExecutionError`)
- `internal/engine/schedule_test.go` (golden corpus includes faults)
- `internal/faults/fault_test.go` (legacy injector)

## Source map

- `internal/engine/schedule.go`
- `internal/engine/runner.go`
- `internal/faults/fault.go`
- `internal/domain/types.go`
- `specs/08_fault_injection_and_dirty_reads.md`

## Related skills

- `chaossql-deterministic-schedule`, `chaossql-runner-execution`, `chaossql-spec-format`
