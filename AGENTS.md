# AGENTS.md — ChaosSQL Harness Engineering Protocol

This document formalizes the operational rules, architectural boundaries, and state guarantees that human contributors and AI agents must follow when developing **ChaosSQL**.

---

## 1. Philosophy and Non-Negotiable Principles

1. **Strict Determinism:**
   Given the same specification (`chaos.yaml`) and the same `--seed`, ChaosSQL must produce the exact same sequence of operations, parameter assignments, and interleaving schedules. Randomness must never depend on uncontrolled wall-clock time or mutable global state.
2. **Verifiability via Invariants:**
   Every concurrency test scenario must define a mathematical or domain-specific consistency invariant evaluated against the database. We reject tests that rely on arbitrary sleep delays or timing luck.
3. **Mandatory Minimization (Delta-Debugging):**
   Upon discovering an invariant violation in a trace of $N$ operations, the engine is required to reduce the trace using Zeller's delta-debugging algorithm ($ddmin$) to an isolated, 1-minimal reproducible counterexample.
4. **Layer Isolation (Clean Architecture & Ports/Adapters):**
   * The **Domain** (invariants, trace models, reduction) has zero awareness of specific database drivers.
   * The **Adapters** (SQLite, PostgreSQL, MySQL, Mock) strictly implement the `DatabaseDriver` interface.
   * The **CLI and Reporters** consume domain artifacts without duplicating execution logic.
5. **Closed Regression Loop:**
   Any reported bug must become a minimal fixture or test case before being marked as resolved.
6. **Zero CGO:**
   All Go binaries and WebAssembly modules must build statically without CGO dependencies (`CGO_ENABLED=0`).

---

## 2. Harness Surface

| Component | Location | Responsibility |
| :--- | :--- | :--- |
| **Operational Rules** | `AGENTS.md` | Working contract for agents and contributors |
| **Boundaries & Design** | `ARCHITECTURE.md` | Formal definition of layers, ports, and state guarantees |
| **Architectural Decisions** | `docs/adrs/` | Immutable records of technical decisions (ADRs) |
| **Formal Specifications** | `specs/` | Formal requirements per system capability (01-15) |
| **Reference Scenarios** | `examples/` | Canonical anomaly test cases (Lost Update, Write Skew, G2, etc.) |
| **Quality Criteria (Evals)**| `evals/` | Shrinking ratio benchmarks and false-positive criteria |
| **Unified Quality Gate** | `make verify` | Single command validating harness, lint, tests, WASM, and English purity |

---

## 3. Agent Development Workflow

1. **Never Break `make verify`:** No changes are accepted if the quality gate (`check-harness`, `lint`, `test`, `stress-wasm`, or `test_english_purity.js`) fails.
2. **Living Documentation:** If modifying engine fuzzer, shrinker, or mutator behaviors, update the corresponding specifications in `specs/`.
3. **Strict Zero CGO:** Ensure `CGO_ENABLED=0` remains functional for all Go targets.

---

## 4. Essential Commands

```bash
make bootstrap   # Download and verify Go dependencies
make test        # Run unit and integration tests (-race)
make lint        # Run go vet and static analysis
make verify      # Execute unified quality gate
make demo        # Run 10 interactive anomaly demonstrations
```
