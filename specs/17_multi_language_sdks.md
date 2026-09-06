# Spec 17: Multi-Language SDKs (Python, TypeScript & Node.js) (v1.4)

## 1. Domain Theory & Motivation
- **Context**: Enterprise applications interacting with relational database systems are overwhelmingly written in **Python (Django, FastAPI, SQLAlchemy)** and **TypeScript/Node.js (NestJS, Express, Prisma, Drizzle)**.
- **Embedded Engine Binary Architecture**:
  - Distributes zero-CGO static binaries of the core ChaosSQL fuzzer (`bin/chaossql`).
  - Host language bindings communicate with the core engine via standard I/O bidirectional JSON streaming (`chaossql engine`).
  - Completely bypasses the Python Global Interpreter Lock (GIL), avoids cross-language FFI deadlocks, and guarantees memory isolation.

## 2. Python SDK (`chaossql-py` / `pip install chaossql`) (`sdks/python/`)
- **Core Fluent Interface (`ChaosHarness`)**:
  - Matches the ergonomics of the native Go SDK (`pkg/chaostest`).
  - Builder methods: `with_schema`, `with_seed`, `with_invariant`, `add_operation`.
  - Execution methods: `run`, `run_and_shrink`, `assert_no_anomalies`.
- **Native Pytest Integration (`chaossql.pytest_plugin`)**:
  - Registered pytest plugin `pytest11 = {"chaossql" = "chaossql.pytest_plugin"}`.
  - Fixture `chaossql_runner` providing declarative YAML scenario execution (`run_scenario`).
  - Test marker `@pytest.mark.chaossql(workers=..., duration=..., seed=...)`.

## 3. TypeScript / Node.js SDK (`@chaossql/test`) (`sdks/typescript/`)
- **Core Fluent Interface (`ChaosHarness`)**:
  - Full TypeScript 5.0+ declaration typings (`.d.ts`).
  - Supports Vitest, Jest, and Node.js Test Runner (`node:test`).
  - Methods: `withSchema`, `withSeed`, `withInvariant`, `addOperation`, `run`, `runAndShrink`, `assertNoAnomalies`.

## 4. Automatic Synthesis of Native Regression Tests
- **Zero-Dependency Reproductions**:
  - Generates standalone Python scripts (`repro_python.go`) using standard library `sqlite3` and `concurrent.futures`.
  - Generates standalone TypeScript / Node.js scripts (`repro_ts.go`) using `node:test` and `Promise.all`.
  - Exposed via `harness.export_standalone_repro()` and `result.export_standalone_repro()`.

## 5. Bidirectional JSON IPC Command (`cmd/chaossql/engine.go`)
- Command: `chaossql engine` consuming JSON specification over stdin and streaming structured results over stdout.
- Emits anomaly classification (Adya cycle analysis), failing invariant evaluation, minimal counterexample traces ($ddmin$), Mermaid sequence diagrams, and multi-language repro templates.

## 6. Quality Gate & Verification
- Spec added to `tools/harness_check.go`.
- Unified validation via `make test-sdks` and `make verify`.
