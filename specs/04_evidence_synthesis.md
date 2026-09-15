# Spec 04: Evidence Synthesis (Mermaid & Repro Script)

## Objective
Transform the minimal failure trace into visual and standalone executable artifacts for immediate post-mortem diagnosis.

## Verifiable Requirements
1. **Mermaid Sequence Diagram:** Generates a `sequenceDiagram` with actors partitioned per worker, highlighting interleaving order, critical transitions, and the database invariant violation note.
2. **Standalone Repro Script (`repro_test.go`):** Synthesizes a self-contained Go test script (independent of external ChaosSQL dependencies) that deterministically reproduces the race condition in under 1 second.
3. **Terminal Report:** Emits an interactive or formatted terminal panel with an invariant summary table comparing expected vs. actual database values.
4. **Versioned Replay Artifact:** `run --export-result <path>` atomically writes a permission-restricted JSON artifact containing the complete specification, effective seed, logical schedule, exact bound operations, observed trace, and stable failure signature.
5. **Executable Verification:** `replay --verify <artifact>` resets the configured database and executes the stored operations. It succeeds only when the logical schedule and stable failure signature match the artifact.
6. **Legacy Inspection:** Existing trace-only JSON remains renderable, but version 0 payloads without execution inputs cannot be verified.
7. **Timing Boundary:** Replay verifies harness-controlled inputs and failure identity. Trace timestamps, physical completion order, and anomaly graph classification remain observations of the database and runtime.
