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
  - Virtual SQLite compiled to WebAssembly (GOOS=js GOARCH=wasm) running 100% inside client-side browser Web Workers with zero backend server dependencies.
  - Interactive web playground deployed at [chaossql.bregalda.com/#/playground](https://chaossql.bregalda.com/#/playground).
  - Client-side execution of the Adya Direct Serialization Graph (DSG) with SVG cycle visualization and causal Delta-Debugging ($) trace shrinker.
  - Formal capability specification [specs/14_wasm_in_browser_playground.md](specs/14_wasm_in_browser_playground.md).

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
