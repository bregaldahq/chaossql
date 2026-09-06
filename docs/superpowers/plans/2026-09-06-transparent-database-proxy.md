# ChaosSQL Transparent Database Reverse Proxy (`chaossql proxy`) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Layer-7 TCP Database Reverse Proxy (`pkg/proxy`, `cmd/chaossql/proxy.go`) for ChaosSQL, intercepting PostgreSQL Wire 3.0 and MySQL protocols to inject stochastic micro-jitter, maintain a Live Shadow Serialization Graph (Shadow DSG), detect concurrency anomalies in real-time, and export OASIS SARIF 2.1.0 diagnostics without requiring changes to application source code.

**Architecture:**
- **Layer-7 Protocol Decoders (`pkg/proxy/postgres.go`, `pkg/proxy/mysql.go`, `pkg/proxy/parser.go`)**: Zero-CGO pure Go streaming decoders parsing wire packets for PostgreSQL (Startup, 'Q', 'P'/'B'/'E', 'Z', 'C', 'T'/'D') and MySQL (COM_QUERY, COM_STMT_PREPARE/EXECUTE, OK/ERR packets).
- **PCT Jitter & Commit Barrier Engine (`pkg/proxy/jitter.go`)**: Microsecond socket delay injector with deterministic PRNG and commit barrier stalling ($50\mu\text{s}$ to $5\text{ms}$) to expose concurrency race windows.
- **Live Shadow Serialization Graph (`pkg/proxy/shadow_graph.go`)**: Real-time dependency tracker maintaining read/write sets per session, building dynamic Adya conflict graphs ($wr, ww, rw$), and detecting anomaly cycles (P4 Lost Update, A5B Write Skew, G1a Dirty Read, G2 Anti-Dependency).
- **Proxy Server & CLI Orchestrator (`pkg/proxy/proxy.go`, `cmd/chaossql/proxy.go`)**: High-throughput bi-directional TCP proxy with graceful termination, live anomaly alerting, Lipgloss summary report, optional live UI endpoint, and SARIF 2.1.0 output.

**Tech Stack:** Go 1.25, `net`, `crypto/rand`, `github.com/spf13/cobra`, `github.com/charmbracelet/lipgloss`, SARIF 2.1.0, Zero CGO (`CGO_ENABLED=0`).

**Spec:** `internal_docs/02_transparent_database_proxy.md` & `specs/16_transparent_database_proxy.md`.

## Global Constraints
- Zero CGO (`CGO_ENABLED=0`) across all Go packages.
- All code, comments, and identifiers must be in English (strictly enforcing `tools/test_english_purity.js`).
- Never break `make verify` (`check-harness`, `lint`, `test`, `stress-wasm`, `test_english_purity.js`).
- Strict determinism: support reproducible runs via `--seed`.

---

## Tasks Summary
- [x] Task 1: Domain Isolation Types & SQL Statement Analysis
- [x] Task 2: PostgreSQL Wire Protocol 3.0 Streaming Decoder
- [x] Task 3: MySQL Client/Server Protocol Streaming Decoder
- [x] Task 4: Stochastic PCT Jitter Engine & Commit Barrier Injector
- [x] Task 5: Live Shadow Serialization Graph (Shadow DSG) & Cycle Detector
- [x] Task 6: High-Throughput Transparent TCP Proxy Core & Session Manager
- [x] Task 7: SARIF 2.1.0 Export & Embedded Trace Visualizer Endpoint
- [x] Task 8: CLI Command `chaossql proxy`, Formal Specification & Unified Verification
