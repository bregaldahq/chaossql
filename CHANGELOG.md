# Changelog

All notable changes to **ChaosSQL** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.4.0] - 2026-09-06

### Added
- **Layer-7 Transparent Database Reverse Proxy (chaossql proxy)**:
  - Zero-CGO PostgreSQL Wire Protocol 3.0 streaming decoder supporting SSLRequest negotiation, Simple Query ('Q'), Parse ('P'), CommandComplete ('C'), and ReadyForQuery ('Z').
  - Zero-CGO MySQL Client/Server protocol streaming decoder supporting packet framing, COM_QUERY, COM_STMT_PREPARE, COM_STMT_EXECUTE, OK_Packet, and ERR_Packet (deadlock 1213 handling).
  - Stochastic PCT micro-jitter engine injecting microsecond delays (\mu\text{s}$ to \text{ms}$) with deterministic PRNG seed (--seed) and transaction priority scheduling.
  - Commit barrier injector stalling outgoing COMMIT packets to expose critical concurrency race windows across concurrent transactions.
  - Online Shadow Dependency Serialization Graph (Shadow DSG) tracking active sessions and transaction boundaries to record , ww, rw$ conflict edges.
  - Real-time Adya cycle detector and anomaly classifier identifying Lost Update ($), Write Skew ($), Dirty Read ($), and Anti-Dependency Cycles ($).
  - OASIS SARIF 2.1.0 security report exporter (--export-sarif) compatible with GitHub Code Scanning.
  - Embedded real-time diagnostic web dashboard and JSON endpoints (/api/status, /api/anomalies, /api/sarif, /api/trace) via --ui-port.
  - CLI command chaossql proxy with Lipgloss summary table, SIGINT/SIGTERM graceful drain, and automated integration script (	ools/run_manual_proxy_test.sh).
  - Formal capability specification [specs/16_transparent_database_proxy.md](specs/16_transparent_database_proxy.md).
- **Autonomous Multi-Engine Differential Swarm & Concurrency Stress Testing**:
  - Multi-engine differential fuzzer chaossql swarm [diff|run] with bounded worker pools comparing isolation levels across SQLite, PostgreSQL, and MySQL.
  - Stochastic DSL scenario mutator chaossql mutate featuring micro-jitter delay perturbation, nested LIFO savepoint rollback lifecycle, causal DAG topological step shuffling, and lock order inversion.
  - Headless WebAssembly & Web Worker stress harness executing 100 consecutive scenarios with bounded RSS memory and 60 FPS non-blocking layout verification.
  - Formal capability specification [specs/15_multiagent_qa_and_swarm_fuzzing.md](specs/15_multiagent_qa_and_swarm_fuzzing.md) and academic foundations [docs/ACADEMIC_FOUNDATIONS.md](docs/ACADEMIC_FOUNDATIONS.md).

---

## [1.3.0] - 2026-09-05

### Added
- **In-Browser WebAssembly (WASM) Playground & Client-Side Verification Engine**:
  - Virtual SQLite compiled to WebAssembly (`GOOS=js GOARCH=wasm`) running 100% inside client-side browser Web Workers with zero backend server dependencies.
  - Pure-Go zero-CGO compilation target with `-ldflags="-s -w -X main.version=1.3.0" -trimpath` under 8MB uncompressed (< 2.2MB gzipped / brotli).
  - Dedicated Web Worker RPC Protocol (`site/assets/wasm-worker.js`) supporting asynchronous event streaming (`INIT`, `VALIDATE`, `RUN`, `CANCEL`, `PROGRESS`, `CYCLE_DETECTED`, `COMPLETE`) to preserve 60 FPS UI performance.
  - Interactive Web Playground Studio (`site/#/playground`) featuring 1-click loading of 10 canonical concurrency anomaly scenarios.
  - Live YAML scenario editor with real-time validation and interactive runtime sliders for concurrency workers (1–8), iterations (5–50), and micro-jitter (0–50ms).
  - Client-side execution of the Adya Direct Serialization Graph (DSG) with responsive SVG conflict graph ($ww, wr, rw$) and pulsing cycle animations.
  - Microsecond Gantt swimlane timeline visualizing worker interleavings and operation latency.
  - Causal Delta-Debugging ($ddmin$) inspector comparing raw schedules against 1-minimal counterexample schedules.
  - Dynamic bilingual localization (PT / EN) with real-time translation of traces, timeline markers, anomaly badges, and console logs.
  - In-browser benchmark telemetry (`site/assets/wasm-bench.js`) and headless stress harness (`tools/headless_worker_stress.js`) validating V8 heap growth < 15MB and WASM linear memory stability.
  - Automated quality gates including structural DOM hierarchy assertions (`tools/test_playground_ui.js`) and English purity enforcement (`tools/test_english_purity.js`).
  - Formal capability specification [specs/14_wasm_in_browser_playground.md](specs/14_wasm_in_browser_playground.md) and release notes [docs/releases/v1.3.0.md](docs/releases/v1.3.0.md).


---

## [1.2.0] - 2026-09-04

### Added
- **Interactive Trace Visualizer & SARIF Reporting**:
  - Embedded local web inspector chaossql ui <trace.json> with microsecond Gantt swimlane and pulsing Adya cycle graph.
  - Standardized OASIS SARIF 2.1.0 export (--export-sarif) mapped to CWE-362 and rules CHAOS001 through CHAOS009.
  - 10th flagship scenario: Foreign Key Cascade Deadlock (\text{-DL}$) with referential lock hierarchy inversion.
  - Official documentation and landing portal at [chaossql.bregalda.com](https://chaossql.bregalda.com) featuring Bregalda visual identity and bilingual switcher.
  - Formal capability specification [specs/13_interactive_visualizer_sarif_and_register_checker.md](specs/13_interactive_visualizer_sarif_and_register_checker.md).

---

## [1.1.0] - 2026-09-02

### Added
- **Developer SDK & Automated Quality Gates**:
  - Programmatic Go testing SDK (pkg/chaostest) for testing transactional code in unit tests (TestChaos_WithDriver).
  - Dynamic parameter generators (, , , ).
  - Official GitHub Action (ction.yml) for automated pull request quality gates.
  - Formal capability specification [specs/11_version_1_1_developer_sdk_and_smart_generators.md](specs/11_version_1_1_developer_sdk_and_smart_generators.md).

---

## [1.0.0] - 2026-09-01

### Added
- **Initial Release of ChaosSQL**:
  - Deterministic PRNG scheduler with pseudo-random concurrency interleaving.
  - Causal Delta-Debugging ($) trace shrinker reducing 100-operation failure traces to 1-minimal counterexamples.
  - Adya Direct Serialization Graph (DSG) anomaly classifier (, A5B, G1a, G2$).
  - Multi-engine driver abstractions supporting pure-Go SQLite (modernc.org/sqlite), PostgreSQL (pgx), and MySQL (go-sql-driver/mysql).
