# Spec 15: Autonomous Multi-Agent QA Swarm & Concurrency Stress Testing (v1.4)

## 1. Domain Theory & Motivation
- **Context**: ChaosSQL v1.0–v1.3 verified database isolation levels using static canonical YAML scenarios and deterministic PRNG scheduling.
- **Swarm Intelligence Architecture**:
  - High-throughput concurrency stress testing requires continuous schedule mutation (AST manipulation, statement delay perturbations, savepoint injections, and lock hierarchy inversion).
  - Multi-Agent coordination distributes adversarial scenario mutation, cross-engine differential execution (SQLite, PostgreSQL, MySQL), and browser Web Worker memory/rendering profiling across specialized agents.

## 2. Adversarial Mutation Subsystem (`pkg/mutator/`)
- Mutation operators for `domain.Spec`:
  - `InterleaveDelayMutation`: Injects stochastic micro-jitter sleep intervals between transaction steps.
  - `SavepointRollbackMutation`: Adds nested `SAVEPOINT sp` and conditional `ROLLBACK TO sp` to test partial transaction recovery.
  - `LockOrderInversionMutation`: Permutes write orders across multiple shared keys to provoke deadlocks.
  - `StepShuffleMutation`: Shuffles independent non-causal transaction operations.
- CLI: `chaossql mutate <scenario.yaml> --variants <N> --output-dir <dir>`.

## 3. Multi-Engine Differential Swarm (`internal/swarm/`)
- Concurrent cross-database test dispatcher:
  - Takes mutated scenario suites and executes synchronized interleaved runs across multiple database targets.
  - Computes Isolation Divergence Matrix: isolates transactions where Driver A (e.g. SQLite Read Committed) permits an anomaly while Driver B (e.g. Postgres Serializable) blocks or aborts with serialization failure.
- CLI: `chaossql swarm run --scenarios-dir <dir> --drivers <csv>`.

## 4. Headless WASM & Performance Profiling (`tools/headless_worker_stress.js`)
- Long-running Web Worker stress harness:
  - Executes 100+ consecutive scenario runs inside headless V8 / Node.js VM.
  - Tracks WebAssembly linear memory growth (`memory.buffer.byteLength`), ensuring zero heap leakage over extended fuzzing runs.
  - Measures main thread event loop latency, validating the 60 FPS non-blocking guarantee.

## 5. Automated Evidence & CI Integration (`internal/reporter/swarm_summary.go`)
- Aggregates multi-agent test matrices into:
  - Standardized Markdown Step Summaries.
  - OASIS SARIF 2.1.0 security advisories.
  - Auto-generated standalone reproduction tests (`repro_test.go`).
