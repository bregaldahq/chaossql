# Spec 13: Interactive Trace Visualizer, SARIF 2.1.0 Reporter & Elle Register Linearizability Checker (v1.2)

## 1. Domain Theory: Elle-Style Register & List-Append Linearizability (`internal/analyzer/register.go`)
- Inspired by Kyle Kingsbury's Elle / Jepsen black-box linearizability auditing.
- Data Model:
  - `RegisterEvent`: `{TxID string, Var string, Op string, Val string, ReadFromTx string, MonotonicIndex int64}`.
  - Linearizability Checker:
    - **G1b (Intermediate Reads)**: Detects when transaction $T_2$ reads a version of variable $X$ that was modified by $T_1$, but $T_1$ subsequently wrote another version to $X$ before committing.
    - **Monotonic Version Regressions**: Detects when a client observes a variable's state moving backwards in logical time.
    - **Fractured List-Appends**: Detects when an append-only collection read omits previously committed elements.

## 2. Security & CI/CD Observability: SARIF 2.1.0 Exporter (`internal/reporter/sarif.go`)
- Complies with OASIS Standard SARIF 2.1.0 (Static Analysis Results Interchange Format).
- Direct integration with GitHub Actions Code Scanning:
  - Driver `chaossql` with rules:
    - `chaossql/P4-lost-update` (Error)
    - `chaossql/A5B-write-skew` (Error)
    - `chaossql/A5A-read-skew` (Warning)
    - `chaossql/G0-dirty-write` (Error)
    - `chaossql/G1a-dirty-read` (Error)
    - `chaossql/G1b-intermediate-read` (Error)
    - `chaossql/G1c-circular-info` (Error)
    - `chaossql/G2-anti-dependency` (Error)
    - `chaossql/G-DL-deadlock` (Warning)
  - `results`: points to file path (`chaos.yaml` / `schema.sql`), line numbers, and formatted markdown explanations with minimal causal reproduction traces.

## 3. Developer Tooling: Interactive Trace Visualizer Server (`cmd/chaossql/ui.go` & `internal/reporter/ui.go`)
- CLI command: `chaossql ui <trace.json> [--port 8090]` and flag `chaossql run scenario.yaml --ui`.
- Embedded HTTP server serving a single-page interactive visual trace inspector:
  - **Transaction Gantt Timeline**: Horizontal swimlane showing worker goroutines, microsecond statement latencies, commit/abort boundaries, and lock contention.
  - **Force-Directed Adya Graph**: SVG graph rendering transaction nodes and color-coded conflict edges ($rw, wr, ww$) with pulsing cycle paths.
  - **Delta-Debugging Scrubber**: Step-by-step slider comparing the raw execution trace against the reduced 1-minimal reproduction.
  - **Statement Inspector**: Clicking any statement shows parameters evaluated, execution time, and SQL returned values.

## 4. 10th Flagship Scenario: `examples/foreign_key_cascade_deadlock/`
- Tables: `customers`, `orders`, `order_items` with `FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE`.
- Workload:
  - $T_1$: Concurrent checkout adding items to existing order.
  - $T_2$: Concurrent order cancellation deleting order with cascaded lock contention.
- Invariant: Referential consistency and zero orphan items.
