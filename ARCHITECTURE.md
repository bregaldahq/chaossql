# ARCHITECTURE.md — ChaosSQL System Architecture

This document formalizes the architecture, layer boundaries, and state guarantees of **ChaosSQL**.

---

## 1. General Layer Diagram

```mermaid
flowchart TB
    subgraph PRESENTATION [Presentation Layer]
        CLI[chaossql CLI - Cobra & Lipgloss]
        TERM[Terminal Reporter & Tables]
        MERMAID[Mermaid Timeline Generator]
        REPRO[Standalone Repro Scripts - Go]
    end

    subgraph APPLICATION [Application Layer]
        EXECUTOR[Chaos Executor & Scheduler]
        SHRINKER[Delta-Debugging Shrinker - ddmin]
        SWARM[Multi-Engine Differential Swarm]
    end

    subgraph DOMAIN [Deterministic Domain]
        PRNG[Seeded PCG64 / Math PRNG]
        INV_MODEL[Invariant Model & SQL Evaluator]
        TRACE[Execution Trace Log]
        MUTATOR[Stochastic Scenario Mutator]
    end

    subgraph PORTS [Ports and Adapters]
        DRIVER_PORT[DatabaseDriver Interface]
        SQLITE[Pure-Go SQLite Adapter - modernc]
        PG[PostgreSQL Adapter - pgx]
        MYSQL[MySQL Adapter - go-sql-driver]
        MOCK[In-Memory Mock Driver]
    end

    CLI --> EXECUTOR
    EXECUTOR --> PRNG
    EXECUTOR --> DRIVER_PORT
    EXECUTOR --> INV_MODEL
    EXECUTOR --> TRACE
    EXECUTOR --> SHRINKER
    SHRINKER --> EXECUTOR
    SWARM --> EXECUTOR
    SWARM --> DRIVER_PORT
    MUTATOR --> DOMAIN
    DRIVER_PORT --> SQLITE
    DRIVER_PORT --> PG
    DRIVER_PORT --> MYSQL
    DRIVER_PORT --> MOCK
    EXECUTOR --> TERM
    EXECUTOR --> MERMAID
    EXECUTOR --> REPRO
```

---

## 2. Engineering Guarantees

| Guarantee | Enforcement Mechanism |
| :--- | :--- |
| **Strict Determinism** | Isolated PRNG seeded via `--seed`; schedule interleavings reproducible bit-for-bit |
| **Reproducibility** | Atomic database reset (schema DDL + seed DML) executed before every scenario iteration |
| **Minimality** | Causal delta-debugging ($ddmin$) guarantees a 1-minimal trace where no operation can be removed |
| **Assertion Safety** | Isolated SQL evaluator verifying mathematical and business invariants after concurrency runs |
| **Real Concurrency** | Transactions dispatched across asynchronous worker goroutines with stochastic jitter perturbation |
| **Zero CGO** | Static binary compilation (`CGO_ENABLED=0`) across native CLI and WebAssembly targets |

---

## 3. Execution Lifecycle

```mermaid
sequenceDiagram
    participant C as CLI / Swarm
    participant E as ChaosExecutor
    participant D as DatabaseDriver
    participant S as TraceShrinker

    C->>E: run(spec, seed=42)
    E->>D: reset(schema, seed)
    E->>D: evaluate_invariants() [Initial PASS]
    E->>D: dispatch N workers (concurrent transactions)
    D-->>E: all transactions finished
    E->>D: evaluate_invariants()
    D-->>E: INVARIANT VIOLATED!
    E->>S: shrink(failing_plan)
    loop Delta Debugging (ddmin)
        S->>D: reset() & run(subset)
        D-->>S: violation status
    end
    S-->>C: Minimal Trace (2 ops) + Repro Script + Mermaid Diagram
```
