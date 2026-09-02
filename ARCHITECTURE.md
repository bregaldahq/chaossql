# ARCHITECTURE.md — ChaosSQL System Architecture

Este documento define a arquitetura, os limites entre camadas e as garantias de estado do **ChaosSQL**.

---

## 1. Diagrama geral de Camadas

```mermaid
flowchart TB
    subgraph PRESENTATION [Camada de Apresentação]
        CLI[chaossql CLI - Typer]
        TERM[Terminal Reporter - Rich]
        MERMAID[Mermaid Timeline Generator]
        REPRO[Standalone Repro Scripts]
    end

    subgraph APPLICATION [Camada de Aplicação]
        EXECUTOR[Chaos Executor & Scheduler]
        SHRINKER[Delta-Debugging Shrinker]
    end

    subgraph DOMAIN [Domínio Determinístico]
        PRNG[Gerador de Parâmetros Seeded]
        INV_MODEL[Modelo de Invariantes & AST]
        TRACE[Execution Trace Log]
    end

    subgraph PORTS [Portas e Adaptadores]
        DRIVER_PORT[Protocolo DatabaseDriver]
        SQLITE[SQLite Adapter - aiosqlite]
        PG[Postgres Adapter - asyncpg]
    end

    CLI --> EXECUTOR
    EXECUTOR --> PRNG
    EXECUTOR --> DRIVER_PORT
    EXECUTOR --> INV_MODEL
    EXECUTOR --> TRACE
    EXECUTOR --> SHRINKER
    SHRINKER --> EXECUTOR
    DRIVER_PORT --> SQLITE
    DRIVER_PORT --> PG
    EXECUTOR --> TERM
    EXECUTOR --> MERMAID
    EXECUTOR --> REPRO
```

---

## 2. Garantias de Engenharia

| Garantia | Como é Assegurada |
| :--- | :--- |
| **Determinismo** | Gerador PRNG isolado por `seed` e filas de workers determinísticas |
| **Reprodutibilidade** | O banco é resetado atomicamente (schema + seed) antes de cada execução |
| **Minimalidade** | Algoritmo $ddmin$ (Delta-Debugging) garante que nenhuma operação possa ser removida sem sumir o bug |
| **Segurança de Asserções** | Avaliação de expressões SQL com safe eval isolado (sem acesso a builtins perigosos) |
| **Concorrência Real** | Transações executadas em conexões assíncronas independentes com jitter de latência |

---

## 3. Ciclo de Vida da Execução

```mermaid
sequenceDiagram
    participant C as CLI
    participant E as ChaosExecutor
    participant D as DatabaseDriver
    participant S as TraceShrinker

    C->>E: run(spec, seed=42)
    E->>D: reset(schema, seed)
    E->>D: evaluate_invariants() [Inicial PASS]
    E->>D: dispatch N workers (concurrent transactions)
    D-->>E: all transactions finished
    E->>D: evaluate_invariants()
    D-->>E: INVARIANT VIOLATED!
    E->>S: shrink(failing_plan)
    loop Delta Debugging
        S->>D: reset() & run(subset)
        D-->>S: violation status
    end
    S-->>C: Minimal Trace (2 ops) + Repro Script + Mermaid
```
