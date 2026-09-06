# Spec 04: Evidence Synthesis (Mermaid & Repro Script)

## Objective
Transform the minimal failure trace into visual and standalone executable artifacts for immediate post-mortem diagnosis.

## Verifiable Requirements
1. **Mermaid Sequence Diagram:** Generates a `sequenceDiagram` with actors partitioned per worker, highlighting interleaving order, critical transitions, and the database invariant violation note.
2. **Standalone Repro Script (`repro_test.go`):** Synthesizes a self-contained Go test script (independent of external ChaosSQL dependencies) that deterministically reproduces the race condition in under 1 second.
3. **Terminal Report:** Emits an interactive or formatted terminal panel with an invariant summary table comparing expected vs. actual database values.
