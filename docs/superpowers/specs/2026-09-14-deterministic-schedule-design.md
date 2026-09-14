# ENG-02 Deterministic Schedule Design

## Problem

ChaosSQL currently consumes random values while ranging over parameter maps, keeps monotonic counters in process-global state, lets workers race for a shared operation channel, and lets concurrent workers race for a shared fault injector. Two runs with the same specification and seed can therefore bind different parameters, assign operations to different workers, inject different faults, and emit different logical schedules.

For a paid service, the result must carry enough deterministic evidence to explain what ChaosSQL decided independently of operating-system and database timing.

## Guarantee

Given the same normalized specification, engine version, and seed, ChaosSQL produces the same:

1. effective seed;
2. operation sequence and bound parameters;
3. operation-to-worker assignment;
4. per-step jitter, latency, and abort decisions; and
5. versioned logical schedule artifact.

Seed zero is the literal deterministic seed `0`. Omitting `engine.seed` therefore produces a reproducible default run. Callers that want a new run must choose and persist a nonzero seed before invoking the engine.

The logical schedule is controlled by the harness. The physical completion order in `ExecutionTrace` remains an observation of the database and runtime. Lock acquisition, server scheduling, network delay, deadlock victim selection, and commit visibility are external behavior and may change even when the logical schedule is identical. ChaosSQL must expose this boundary and must not claim unrestricted bit-for-bit replay of an external distributed system.

## Data model

Add these domain artifacts:

```go
type SchedulePlan struct {
    Version   int                `json:"version"`
    Seed      uint64             `json:"seed"`
    Workers   int                `json:"workers"`
    Decisions []ScheduleDecision `json:"decisions"`
}

type ScheduleDecision struct {
    Sequence    int           `json:"sequence"`
    OperationID int           `json:"operation_id"`
    WorkerID    int           `json:"worker_id"`
    StepIndex   int           `json:"step_index"`
    JitterMs    int  `json:"jitter_ms"`
    LatencyMs   int  `json:"latency_ms"`
    Abort       bool `json:"abort"`
}
```

`ExecutionResult` includes `Seed` and `Schedule`. `ScheduleOutcome` carries the same schedule to the shared result finalizer. `SchedulePlan.Version` starts at `1`; changing the derivation algorithm requires a new version.

Each decision represents the harness inputs immediately before one SQL step. Decisions are ordered by operation ID and step index, making serialization stable. `WorkerID` is assigned round-robin as `(operationID - 1) % workers + 1`. Configured worker count is part of the specification, so changing it intentionally changes worker assignment.

## Generation

`NewPRNG(0)` preserves zero. A PRNG instance owns its monotonic counter map. Schedule generation sorts parameter names before consuming random values and uses only the instance-scoped evaluator. Independent PRNG instances with the same seed start with identical counter state.

Decision randomness is derived from `(master seed, operation ID, step index)` rather than call order. Jitter and fault choices are therefore stable even when goroutines complete in a different order. The existing fault injector may remain for callers outside the runner, but runner execution consumes the precomputed decisions only.

## Execution

Before workers start, the runner builds one schedule plan and partitions operations into deterministic per-worker queues. A worker consumes only its own queue in operation-ID order. `executeOperation` looks up the precomputed decision for each step and applies its jitter, latency, and abort fields.

Context cancellation still interrupts every planned wait. Database calls continue to use context-aware driver methods, so externally blocked steps terminate according to the execution context and produce the status defined by ENG-01.

## Compatibility

Existing specifications remain valid. Result fields are additive. Existing `ScheduledOps` and physical `Trace` remain available. JSON, cloud payloads, SDK adapters, and report metadata must preserve the effective seed and logical schedule when they expose full execution results.

The legacy package-level generator API remains available, but engine schedule generation must not use its global monotonic counter state.

## Acceptance

- One hundred in-process generations with three or more randomized map parameters are byte-identical.
- Independent PRNG instances with the same seed generate the same monotonic sequence without resets.
- Seed zero is stable across process invocations and appears as zero in the result.
- Twenty executions assign every operation to the same worker and produce the same logical schedule.
- Different goroutine completion timing does not change jitter or fault decisions.
- Cancellation during a planned wait remains prompt.
- `make verify`, zero-CGO native builds, and the zero-CGO WASM build pass.
