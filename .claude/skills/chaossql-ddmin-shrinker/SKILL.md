---
name: chaossql-ddmin-shrinker
description: The delta-debugging shrinker (internal/shrinker) — failure signatures, the memoized ddmin loop, the empty-schedule baseline, the 1-minimality audit, and how each caller wires the oracle and verifies the result. Use when changing shrinking, debugging a shrink that does not reduce or does not reproduce, or adding a new caller.
---

# ddmin Shrinker

When a run ends in `violation`, ChaosSQL reduces the scheduled operations to a
1-minimal subset that still produces the **same failure signature**
(principle #3 of `AGENTS.md`, ADR 0003).

## When to use

- Changing `internal/shrinker/ddmin.go`.
- A shrink reports 0% reduction, `ErrBaselineFailure`, or "minimal operations
  did not reproduce target failure".
- Adding a new place that shrinks (new SDK, surface, command).

## Failure signature

`FailureSignatureFor(result)`:
- `passed` → `{Status: passed}`;
- `violation` → `{Status: violation, FailingInvariant: <name>}` (requires the
  violation marker and a named failing invariant);
- any other status → error (no stable replay signature).

`ReproducesFailure(result, target)` requires the same status and, for
violations, the same invariant name. A different invariant failing is **not**
a reproduction.

## `Shrink(ctx, testFn, ops)`

Oracle convention: `testFn(subset) == true` means **PASS** (does not
reproduce); `false` means **FAIL** (reproduces).

1. Memoization by the comma-joined op IDs in order; `Trials` counts only real
   oracle calls. Context is checked before and after each call.
2. The full set must fail, else error `initial operations do not fail the test`.
3. The empty set must pass, else `ErrBaselineFailure` (the failure exists with
   no operations — e.g. a broken seed or an invariant that is always false).
4. ddmin loop with `n = 2` while `len(c) >= 2`:
   - partition `c` into `n` contiguous chunks (earlier chunks get the remainder);
   - try each **complement** first; on failure `c = complement`, `n = max(n-1, 2)`;
   - else try each **subset**; on failure `c = subset`, `n = 2`;
   - else stop if `n == len(c)`, otherwise `n = min(2n, len(c))`.
5. 1-minimality audit: repeatedly try removing each single op; keep any removal
   that still fails. Removing the last op and still failing → `ErrBaselineFailure`.
6. Result: `OriginalSize`, `ReducedSize`, `ReductionRatio` (percent 0–100),
   `MinimalOps`, `Iterations` (ddmin rounds), `Trials`, `Duration`.

## How callers wire it (all follow the same pattern)

```go
target, _ := shrinker.FailureSignatureFor(runResult)          // only when ViolationDetected
testFn := func(subset []domain.ScheduledOp) bool {
    res, err := runner.RunSchedule(ctx, spec, subset)          // reset + execute + finalize
    if err != nil { return true }                               // errors count as "not reproduced"
    return !shrinker.ReproducesFailure(res, target)
}
shrunk, err := shrinker.Shrink(ctx, testFn, runResult.ScheduledOps)
// then re-run shrunk.MinimalOps once and accept only if it reproduces
```

Callers: `cmd/chaossql/main.go` (`run`/`demo`), `cmd/chaossql/engine.go`
(IPC), `pkg/chaostest/chaostest.go`, `cmd/chaossql-wasm/bridge.go`.
`ShrinkExecution` implements the same flow but none of them call it (only tests).
On cancellation all callers return the context error; on other shrink errors
they fall back to the unshrunk ops (`chaostest` returns a 0% `ShrinkResult`).

## Gotchas

- Each trial resets the database and re-executes with real concurrency, so the
  oracle can be flaky; memoization freezes the first answer per subset, and
  the final verification run can still fail — callers then keep the original
  trace.
- Because decisions and worker assignment depend only on op ID (see
  `chaossql-deterministic-schedule`), subsets keep their original timing
  decisions; op IDs are **not** renumbered.
- Complements are tried before subsets (reverse of the textbook order).
- Anomaly classification is redone on the minimal trace and may differ from
  the full trace.

## Change checklist

- Keep the oracle convention; update all four callers if the wiring changes
  (or migrate them to `ShrinkExecution`).
- Evals: `evals/01_shrinking_ratio.md`, `evals/03_deterministic_replay.md`.
- `bench` measures ddmin throughput (`chaossql-benchmarks`).

## Tests

- `internal/shrinker/ddmin_test.go` (synthetic oracle, no-bug, cancellation,
  baseline failure, trial counting, deterministic tie-breaking, invariant
  preservation across runs)

## Source map

- `internal/shrinker/ddmin.go`
- `internal/shrinker/ddmin_test.go`
- `cmd/chaossql/main.go`
- `cmd/chaossql/engine.go`
- `pkg/chaostest/chaostest.go`
- `cmd/chaossql-wasm/bridge.go`
- `docs/adrs/0003-causal-delta-debugging-shrinker.md`
- `specs/03_delta_debugging_shrinker.md`
- `evals/01_shrinking_ratio.md`

## Related skills

- `chaossql-runner-execution`, `chaossql-deterministic-schedule`,
  `chaossql-replay-artifacts`, `chaossql-repro-synthesis`
