# Changelog

All notable changes to **ChaosSQL** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.6.0] - 2026-09-23

Reliability and security release for ChaosSQL Cloud, plus the first managed
deployment tooling. Upgrading self-hosted servers runs one-time database
migrations automatically.

### Added
- **Multiple organizations per server**: `server org create|list|set-plan`
  provisions isolated tenants with a one-time owner token (#25).
- **Plan retention enforcement** (opt-in, `--enforce-retention` or
  `CHAOSSQL_ENFORCE_RETENTION=true`), preserving baseline runs and pending
  alerts (#26).
- **Managed deployment** (`deploy/fly/`): single-origin API and dashboard with
  continuous Litestream backups, restore drill, and runbook
  (`docs/managed-deployment.md`) (#30).
- **Self-hosted distribution**: Docker Compose, Helm chart, and `chaossql server`
  (#22).
- Admin endpoint to issue member tokens; `--version` on `chaossql` and
  `chaossql-server`.
- Deterministic schedule generation, trustworthy shrinking, and executable
  replay (#17, #18); real transaction semantics in drivers (#16).

### Changed
- Ingestion is transactional and idempotent per execution; identical retries
  replay the stored response and conflicting content returns 409 (#21, #23).
- Baselines use real commit time and a locally computed scenario fingerprint;
  a one-time migration clears legacy upload times stored as commit times (#23).
- Hosted payloads carry metadata only: no SQL, parameters, schema, traces, or
  reproduction code (#20).
- Webhook alerts go through a transactional outbox with retries (#21).
- `server create-token` requires an existing organization instead of silently
  creating one (#25).
- Site sections use crawlable path URLs (`/dashboard`, `/docs`); old hash links
  are migrated (#27, #29). Cloud run links now point to `/dashboard?run=<id>`.
- The release version has a single source (`internal/version`), used by the
  CLI, SARIF reports, the Cloud client User-Agent, and the WASM engine.

### Fixed
- Tenant authorization on every hosted route (#19).
- Webhook token leak and SSRF protections, including DNS pinning (#13, #23).
- GitHub Action no longer interpolates inputs into shell source and exposes
  `cloud-run-id`, `cloud-run-url`, and `is-regression` outputs (#23).
- Generated TypeScript reproductions fail loudly without a SQLite driver
  instead of reporting success against a no-op database (#28).
- Dependency security updates (#14, #24).

---

## [1.4.0] - 2026-09-06

### Added
- **Multi-Language SDKs & Embedded Engine Binary IPC (Python, TypeScript & Node.js)**:
  - Zero-CGO Embedded Engine Binary IPC architecture (`chaossql engine`) utilizing bidirectional JSON streaming over standard I/O (GIL bypass, memory isolation, and cross-language type safety).
  - Python SDK (`chaossql-py` / `pip install chaossql` at `sdks/python/`) featuring fluent `ChaosHarness`, `@pytest.mark.chaossql` marker, and `chaossql_runner` pytest fixture.
  - TypeScript / Node.js SDK (`@chaossql/test` / `npm install @chaossql/test` at `sdks/typescript/`) featuring fluent `ChaosHarness`, TypeScript 5.0+ declaration typings (`.d.ts`), and support for Vitest, Jest, and Node test runner (`node:test`).
  - Zero-dependency standalone regression test synthesizers (`internal/reporter/repro_python.go` and `repro_ts.go`) producing executable test files.
  - Formal capability specification [specs/17_multi_language_sdks.md](specs/17_multi_language_sdks.md).
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
