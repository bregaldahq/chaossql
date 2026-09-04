# ChaosSQL

```
   ____ _                     ____   ___  _                       
  / ___| |__   __ _  ___  ___/ ___| / _ \| |                      
 | |   | '_ \ / _` |/ _ \/ __\___ \| | | | |                      
 | |___| | | | (_| | (_) \__ \___) | |_| | |___                   
  \____|_| |_|\__,_|\___/|___/____/ \__\_\_____|                  
                                                                  
 Causal Concurrency Stress Testing & Anomaly Synthesis
```

> **Deterministic Concurrency & Invariant Fuzzer for SQL Databases**  
> *Finding isolation anomalies (Lost Update, Write Skew, Read Skew, Dirty Write, Dirty Read, Intermediate Read, Circular Info, G2 Anti-Dependency, Deadlocks, Phantom Reads) and shrinking chaotic traces to 1-minimal reproductions.*

[![Documentation Portal](https://img.shields.io/badge/Docs-chaossql.bregalda.com-4B2E83?style=flat&logo=cloudflare)](https://chaossql.bregalda.com)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Version](https://img.shields.io/badge/Version-1.2.0-blue?style=flat)](https://github.com/bregaldahq/chaossql/releases)
[![Zero CGO](https://img.shields.io/badge/CGO-Disabled_(Pure_Go)-success)](https://modernc.org/sqlite)
[![CI Pipeline](https://github.com/bregaldahq/chaossql/actions/workflows/ci.yml/badge.svg)](https://github.com/bregaldahq/chaossql/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Harness Engineering](https://img.shields.io/badge/Harness-Verified_(41_Artifacts)-blueviolet)](AGENTS.md)
[![Security Policy](https://img.shields.io/badge/Security-Defensive_Sandboxed-blue)](SECURITY.md)

---

## 🌐 Official Documentation & Interactive Portal

👉 **Visit the live portal:** **[https://chaossql.bregalda.com](https://chaossql.bregalda.com)**

Featuring:
* **Interactive 10 Scenarios Explorer** (SQL schema, seed, workload, and $ddmin$ reduction).
* **Live Terminal Simulator** (step-by-step causal trace reduction).
* **Academic Isolation Theory** (Adya directed dependency graph interactive diagrams).
* **Go SDK Playground** (`pkg/chaostest`).
* **Hermitage Empirical Isolation Matrix** across SQLite, PostgreSQL, and MySQL.

---

## ⚡ The Problem: Concurrency Anomalies in SQL Databases

Concurrency bugs in transactional systems (such as *Lost Updates*, *Write Skew*, *Read Skew*, *Dirty Writes*, *Dirty Reads*, *Intermediate Reads*, *Circular Info Flows*, *Deadlocks*, and *G2 Anti-Dependency Cycles*) are among the hardest defects to detect and debug:
1. **Flaky & Non-Deterministic:** They depend on microsecond race conditions between concurrent worker threads and OS scheduling.
2. **Untraceable in Production:** When a database invariant is violated (e.g. an account balance going negative or stock being oversold), production logs contain thousands of interleaved operations, making root-cause analysis nearly impossible.
3. **Engine-Specific Isolation Quirks:** `READ COMMITTED` and `SNAPSHOT ISOLATION` exhibit subtle semantic differences across SQLite, PostgreSQL, and MySQL.

**ChaosSQL solves this by:**
* Injecting **stochastic micro-jitter**, **PCT-SQL priority scheduling**, and **fault injection** (forced aborts, latency spikes, simulated disconnects) to reliably trigger rare race conditions.
* Constructing client-observed **Serialization Graphs** $SG(S) = (V, E)$ to formally classify isolation anomalies ($P4, A5B, A5A, G0, G1a, G1b, G1c, G2, G\text{-DL}$).
* Applying **Elle-Style Register Linearizability Checking** to detect intermediate uncommitted reads ($G1b$) and fractured collection reads.
* Applying **Causal Delta-Debugging ($ddmin$)** to shrink a 100-operation noisy trace down to the **exact 2 or 3 operations** that caused the bug in $< 200\text{ms}$.
* Launching an instant **Interactive Trace Visualizer (`chaossql ui`)** with worker Gantt swimlanes and SVG dependency graphs.
* Generating **OASIS SARIF 2.1.0 Security Reports (`--export-sarif`)** for native GitHub Code Scanning and Security alerts.
* Providing a native **Go Testing SDK (`pkg/chaostest`)** and **Official GitHub Action (`action.yml`)**.

---

## 🖥️ Interactive Trace Visualizer (`chaossql ui`)

Launch a local visual trace inspector server on demand or directly after a chaos run:

```bash
# Launch interactive trace viewer for any execution trace
chaossql ui trace.json --port 8090

# Or run a scenario and automatically launch the visualizer
chaossql run examples/banking_lost_update/chaos.yaml --ui
```

Features:
* **Gantt Timeline Swimlane**: Microsecond-accurate concurrent worker interleavings.
* **Force-Directed Adya Graph**: SVG rendering of transactions ($T_1, T_2$) and directed conflict edges ($rw, wr, ww$) with pulsing cycle paths.
* **Delta-Debugging Comparison**: Side-by-side comparison of the raw 100-operation trace vs the 1-minimal 2-operation reproduction.
* **Statement Inspector**: Detailed view of query text, parameters evaluated, returned rows, and latencies.

---

## 🛡️ OASIS SARIF 2.1.0 Security Reporting

Export standardized SARIF 2.1.0 reports for native integration into **GitHub Code Scanning & Security Alerts**:

```bash
chaossql run examples/banking_lost_update/chaos.yaml --export-sarif results.sarif
```

Rules mapped:
* `chaossql/P4-lost-update` (Error)
* `chaossql/A5B-write-skew` (Error)
* `chaossql/A5A-read-skew` (Warning)
* `chaossql/G0-dirty-write` (Error)
* `chaossql/G1a-dirty-read` (Error)
* `chaossql/G1b-intermediate-read` (Error)
* `chaossql/G1c-circular-info` (Error)
* `chaossql/G2-anti-dependency` (Error)
* `chaossql/G-DL-deadlock` (Warning)

---

## 💻 Go Developer Testing SDK (`pkg/chaostest`)

Embed deterministic concurrency stress tests directly into your standard Go test suite:

```go
package myapp_test

import (
    "context"
    "testing"
    "github.com/bregaldahq/chaossql/pkg/chaostest"
)

func TestAccountTransfer_NoLostUpdates(t *testing.T) {
    ctx := context.Background()

    schema := `
    CREATE TABLE accounts (
        id INT PRIMARY KEY,
        balance INT NOT NULL
    );`

    seed := `
    INSERT INTO accounts VALUES (1, 1000), (2, 1000);`

    chaostest.New(t).
        WithSchema(schema).
        WithSeed(seed).
        WithInvariant("total_wealth", "SELECT sum(balance) AS total FROM accounts;", "total == 2000").
        AddOperation("transfer_1_to_2",
            "SELECT balance FROM accounts WHERE id = 1 -> bal1",
            "UPDATE accounts SET balance = {bal1 - 50} WHERE id = 1",
            "UPDATE accounts SET balance = balance + 50 WHERE id = 2",
        ).
        AddOperation("transfer_2_to_1",
            "SELECT balance FROM accounts WHERE id = 2 -> bal2",
            "UPDATE accounts SET balance = {bal2 - 50} WHERE id = 2",
            "UPDATE accounts SET balance = balance + 50 WHERE id = 1",
        ).
        AssertNoAnomalies(ctx, 4, 30, 42) // workers=4, iterations=30, seed=42
}
```

---

## 🚀 10 Flagship Demonstration Scenarios

| # | Anomaly / Scenario | Cycle / Structure | Result / Impact |
| :-: | :--- | :--- | :--- |
| **1** | **🏦 Banking Lost Update ($P4$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$ | Silent balance loss under `READ COMMITTED` |
| **2** | **🛒 Inventory Oversell ($A3$)** | Anti-dependency on predicate ($G\text{-phantom}$) | Negative stock inventory under concurrent checkout |
| **3** | **🏥 Hospital Write Skew ($A5B$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$ | 0 doctors on duty under Snapshot Isolation |
| **4** | **💳 Financial Read Skew ($A5A$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{wr} T_1$ | Inconsistent audit balances across accounts |
| **5** | **🏷️ Auction Dirty Write ($G0$)** | $T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$ | Disassociated bidder price vs winner ID |
| **6** | **🪙 Crypto Arbitrage ($G1c$)** | $T_1 \xrightarrow{wr} T_2 \xrightarrow{wr} T_1$ | Circular stale oracle executions in AMMs |
| **7** | **⚡ Flash Crash Dirty Read ($G1a$)** | $w_1(\text{price}) \dots r_2 \dots a_1$ | Erroneous collateral liquidation on aborted data |
| **8** | **🎟️ Ticket Anti-Dependency ($G2$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_3 \xrightarrow{rw} T_1$ | Tri-partite concurrent seat overbooking |
| **9** | **🔒 Deadlock Cycle & Recovery ($G\text{-DL}$)** | $T_1 \xrightarrow{\text{waits-for}} T_2 \xrightarrow{\text{waits-for}} T_1$ | Lock contention detection and wealth preservation |
| **10** | **🔗 Foreign Key Cascade Deadlock ($G\text{-DL}$)** | $T_1: \text{Child} \to \text{Parent} \leftrightarrow T_2: \text{Parent} \to \text{Child}$ | Referential lock hierarchy inversion and orphan safety |

---

## 📦 Quickstart & Commands

```bash
# 1. Bootstrap and verify
make bootstrap && make verify

# 2. Scaffold a new scenario
bin/chaossql init my_scenario --driver sqlite

# 3. Validate scenario statically
bin/chaossql validate my_scenario/chaos.yaml

# 4. Run all 10 interactive demos
make demo

# 5. Generate Hermitage empirical isolation matrix
make matrix

# 6. Run high-throughput stress benchmark (13.9M ops/s)
make bench

# 7. Start local documentation site
make serve-site
```

---

## 📊 Harness Engineering & Quality Gate

| Quality Gate | Requirement | Status |
| :--- | :--- | :--- |
| **Contractual Integrity** | 41 mandatory architectural and security artifacts | `PASS` (`tools/harness_check.go`) |
| **Static Analysis** | `go vet ./...` with zero warnings | `PASS` |
| **Concurrency Safety** | `go test -race ./internal/... ./cmd/... ./pkg/...` | `PASS` (0 data races) |
| **CGO Freedom** | Compiles with `CGO_ENABLED=0` | `PASS` (pure Go SQLite) |
| **Deterministic Replay** | Identical seed produces identical interleavings | `PASS` (100% convergence) |

---

## 📜 License & Security

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for terms.  
For security architecture and vulnerability disclosure, see [`SECURITY.md`](SECURITY.md).

Developed by **Ricardo Bregalda** ([@bregaldahq](https://github.com/bregaldahq)).
