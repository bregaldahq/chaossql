# Eval 03: Deterministic Reproducibility and Convergence ($100\%$)

## Objective
Guarantee that identical specifications, engine versions, and `seed` values produce identical harness-controlled logical schedules across 100% of executions.

## Acceptance Criteria
1. **Schedule Identity:** One hundred independent process executions initialized with the same specification and seed must generate byte-identical `ScheduledOp` and versioned `SchedulePlan` artifacts.
2. **Parameter Stability:** A fixture with at least three randomized map parameters must remain identical regardless of Go map iteration order.
3. **Counter Isolation:** Independent executions must start `$monotonic_counter` generators from the configured start without process-global resets.
4. **Decision Stability:** Worker assignments and per-step jitter, latency, and abort decisions must remain identical when goroutine and database completion timing changes.
5. **Zero Seed:** An omitted or explicitly zero seed must produce the checked-in version 1 golden corpus and must be persisted as `0` in the result.
6. **Cancellation:** A context canceled during a planned jitter or latency wait must return promptly and roll back an open transaction.
7. **Shrinker Determinism:** The synthesized minimal failure trace must contain identical operation IDs across all independent runs given the same seed and logical schedule.

Physical database completion order is measured rather than promised. A schedule mismatch is a harness determinism failure; a trace mismatch with an identical schedule is evidence of external runtime or database timing and must be reported as such.
