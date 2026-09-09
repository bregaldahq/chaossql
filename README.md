<div align="center">

<img src="site/assets/icone_bregalda.svg" width="96" height="96" alt="ChaosSQL Logo" />

# ChaosSQL

### Catch database concurrency bugs before production does.

[![Documentation Portal](https://img.shields.io/badge/Docs-chaossql.bregalda.com-4B2E83?style=for-the-badge&logo=cloudflare&logoColor=white)](https://chaossql.bregalda.com)
[![Release Version](https://img.shields.io/badge/Release-v1.4.0-F5C400?style=for-the-badge&logo=github&labelColor=2A2140)](https://github.com/bregaldahq/chaossql/releases/tag/v1.4.0)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![Zero CGO](https://img.shields.io/badge/CGO-Disabled_(Pure_Go)-22C55E?style=for-the-badge)](https://modernc.org/sqlite)
[![CI Pipeline](https://img.shields.io/badge/CI-Passing-22C55E?style=for-the-badge&logo=githubactions&logoColor=white)](https://github.com/bregaldahq/chaossql/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](LICENSE)

<p align="center">
  <b>Deterministic Concurrency & Invariant Testing for PostgreSQL, MySQL, and SQLite.</b><br />
  Stop silent data corruption, write skew, and lost updates in CI/CD before they reach production.
</p>

<p align="center">
  <a href="https://chaossql.bregalda.com">🌐 Official Portal</a> •
  <a href="https://chaossql.bregalda.com/#/playground">🧪 Interactive WASM Playground</a> •
  <a href="#-quickstart-in-30-seconds">🚀 Quickstart</a> •
  <a href="#-3-flagship-concurrency-demos">🔬 3 Key Demos</a> •
  <a href="#️-chaossql-cloud--concurrency-audit">☁️ Cloud Early Access</a>
</p>

</div>

---

## ⚡ The Concurrency Blindspot in Modern Databases

Unit tests verify functions in isolation. In production, concurrent transactions interleave in unpredictable orders, causing silent data corruption that standard tests miss:

```text
❌ Lost Update Detected (P4 Anomaly)
─────────────────────────────────────────────────────────────────────────────
Expected Account Balance:   $2,000.00
Actual Account Balance:     $1,950.00 (Silent race condition under READ COMMITTED)
Deterministic Seed:         184729
Root Cause:                 T1 and T2 concurrently read balance=2000 before either commit
Minimal Reproduction:       4 operations synthesized in repro_test.go (< 200ms)
─────────────────────────────────────────────────────────────────────────────
```

**ChaosSQL finds these bugs deterministically, synthesizes a 1-minimal reproduction, and guards your Pull Requests from regressions.**

---

## 🚀 Quickstart in 30 Seconds

Install the standalone CLI binary with zero CGO dependencies:

```bash
# Install CLI via Go (Go 1.22+)
go install github.com/bregaldahq/chaossql/cmd/chaossql@latest

# Run the flagship banking lost-update scenario
chaossql run examples/banking_lost_update/chaos.yaml
```

ChaosSQL schedules concurrent worker transactions, detects invariant violations, and generates a standalone Go reproduction test (`repro_test.go`) ready for your test suite.

---

## 🔬 3 Flagship Concurrency Demos

Explore classic race conditions with pre-packaged scenarios, or test them 100% in your browser without installing anything:

| Scenario | Anomaly & Impact | Test in Browser | CLI Command |
| :--- | :--- | :---: | :--- |
| **🏦 Banking Transfer** | **Lost Update ($P4$):** Concurrent debits overwrite balance changes under `READ COMMITTED`. | [Launch Playground](https://chaossql.bregalda.com/#/playground) | `chaossql run examples/banking_lost_update/chaos.yaml` |
| **🏥 Hospital Shift** | **Write Skew ($A5B$):** Two doctors concurrently drop shift, leaving 0 on duty under `REPEATABLE READ`. | [Launch Playground](https://chaossql.bregalda.com/#/playground) | `chaossql run examples/hospital_write_skew/chaos.yaml` |
| **🔒 Deadlock Cycle** | **Resource Deadlock ($G	ext{-DL}$):** Inverted key lock acquisitions lock worker goroutines permanently. | [Launch Playground](https://chaossql.bregalda.com/#/playground) | `chaossql run examples/deadlock_cycle/chaos.yaml` |

---

## 🤖 Continuous Concurrency in CI/CD

Prevent concurrency regressions from ever reaching `main`. Add ChaosSQL to your GitHub Actions pipeline:

```yaml
name: Concurrency Guard
on: [pull_request, push]

jobs:
  concurrency:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: bregaldahq/chaossql@v1
        with:
          spec-path: examples/banking_lost_update/chaos.yaml
          export-summary: true
```

When an invariant fails, the action blocks the pull request, publishes a GitHub Step Summary with the minimal execution trace, and synthesizes a `repro_test.go` artifact.

---

## ☁️ ChaosSQL Cloud & Concurrency Audit

### 1. ChaosSQL Cloud (CI/CD Regression Guard)
* **Main Branch Baseline:** Automatically stores invariant baselines for your default branch.
* **Pull Request Comments:** Instantly comments on failing PRs with the root-cause trace and anomaly classification ($P4$, $A5B$, etc.).
* **Zero Sensitive Data:** Queries and database payloads remain inside your CI runner—only execution metadata and minimal traces are reported.

👉 **[Join the Cloud Early Access Waitlist](https://chaossql.bregalda.com/#waitlist)**

### 2. ChaosSQL Concurrency Audit
Preparing a major launch, financial ledger, or high-throughput reservation engine? Studio Bregalda provides dedicated **Database Concurrency Audits** ($500 – $2,000):
* Formal invariant specification of your critical transaction flows.
* Multi-engine isolation analysis (PostgreSQL vs. MySQL vs. SQLite).
* Turnkey minimal reproductions and mitigation blueprints (`SELECT ... FOR UPDATE`, OCC versioning, SSI).

👉 **[Request an Audit for Your Team](https://chaossql.bregalda.com/#waitlist)**

---

## 🔬 10 Flagship Concurrency Scenarios & Production Fixes

ChaosSQL comes pre-packaged with 10 production-grade scenarios representing classic financial, e-commerce, healthcare, and distributed systems race conditions:

| # | Anomaly & Scenario | Conflict Cycle / Structure | Business Impact | Production Fix / Mitigation |
| :-: | :--- | :--- | :--- | :--- |
| **1** | **Banking: Lost Update ($P4$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{ww} T_1$ | Silent balance loss under `READ COMMITTED` | Atomic update (`UPDATE ... balance = balance - 100`) or pessimistic lock (`FOR UPDATE`) |
| **2** | **Inventory: Oversell ($A3$)** | Anti-dependency on predicate ($G\text{-phantom}$) | Negative stock inventory under concurrent checkout | Guarded conditional decrement (`WHERE stock >= 1`) checking rows affected |
| **3** | **Hospital Shift: Write Skew ($A5B$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_1$ | 0 doctors on duty under Snapshot Isolation | Elevate to Serializable Snapshot Isolation (SSI) or explicit master row lock |
| **4** | **Financial Balances: Read Skew ($A5A$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{wr} T_1$ | Inconsistent audit balance totals between accounts | Run read transactions under `REPEATABLE READ` or `SNAPSHOT ISOLATION` |
| **5** | **Auction Bidding: Dirty Write ($G0$)** | $T_1 \xrightarrow{ww} T_2 \xrightarrow{ww} T_1$ | Disassociated bidder price vs winning bidder ID | Strict Two-Phase Locking (2PL) and monotonic bid condition check |
| **6** | **Crypto Exchange: Circular Flow ($G1c$)** | $T_1 \xrightarrow{wr} T_2 \xrightarrow{wr} T_1$ | Stale oracle arbitrage and pricing imbalances in AMMs | Optimistic Concurrency Control (OCC) with monotonic versioning and atomic oracle reads |
| **7** | **Flash Crash: Dirty Read ($G1a$)** | $w_1(\text{price}) \dots r_2 \dots a_1$ | Unwarranted collateral liquidation on aborted prices | Enforce minimum `READ COMMITTED` isolation floor in database connection pool |
| **8** | **Ticket Booking: Anti-Dependency ($G2$)** | $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_3 \xrightarrow{rw} T_1$ | Concurrent overbooking of the same reserved seat | Composite `UNIQUE(section, seat_no)` constraint or serializable predicate locks |
| **9** | **Deadlock Cycle: Lock Inversion ($G\text{-DL}$)** | $T_1 \xrightarrow{\text{waits-for}} T_2 \xrightarrow{\text{waits-for}} T_1$ | Transaction timeouts and aborted wealth transfers | Canonical global primary key lock ordering (`ORDER BY id ASC`) before locking |
| **10** | **FK Cascade Deadlock ($G\text{-DL}$)** | $T_1: \text{Child} \to \text{Parent} \leftrightarrow T_2: \text{Parent} \to \text{Child}$ | Cascading delete deadlocks and orphaned records | Add B-Tree index on foreign key column and standardize parent-before-child locking |

Run any scenario interactively:
```bash
chaossql demo banking_lost_update
chaossql demo hospital_write_skew
chaossql demo foreign_key_cascade_deadlock
```

---

## 🖥️ Interactive Trace Visualizer (`chaossql ui`)

Inspect concurrent execution interleavings, microsecond worker timings, and isolation conflict graphs directly in your browser:

```bash
# Launch visualizer from an existing execution trace
chaossql ui trace.json --port 8090

# Or execute a scenario and immediately launch the UI inspector
chaossql run examples/banking_lost_update/chaos.yaml --ui
```

* **Gantt Swimlane Timeline**: Worker-by-worker execution visualization down to microsecond resolution ($0-250\mu\text{s}$).
* **Adya Dependency Graph**: Interactive SVG Direct Serialization Graph (DSG) highlighting read-write ($rw$), write-write ($ww$), and write-read ($wr$) cycles.
* **1-Minimal Causal Comparison**: Toggle between the noisy raw trace (e.g. 20 operations) and the $ddmin$-shrunk minimal reproduction (2 operations).
* **Statement Inspector**: View exact SQL statements, bound dynamic parameters, executed latencies, and invariant assertion expressions.

---

## 💻 Go Developer SDK (`pkg/chaostest`)

Embed deterministic concurrency stress tests directly into your standard `go test` suites:

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

## 🐍 Python SDK (`chaossql-py` / `pip install chaossql`)

Embed deterministic concurrency fuzzer tests directly into `pytest` or `unittest`:

```python
from chaossql import ChaosHarness

def test_banking_lost_update_prevented():
    harness = ChaosHarness(driver="sqlite", dsn=":memory:")

    result = (
        harness.with_schema("CREATE TABLE accounts (id INT PRIMARY KEY, balance INT NOT NULL);")
               .with_seed("INSERT INTO accounts VALUES (1, 1000), (2, 1000);")
               .with_invariant(
                   name="total_wealth_conserved",
                   query="SELECT sum(balance) AS total FROM accounts;",
                   assertion="total == 2000"
               )
               .add_operation("transfer_1_to_2", [
                   "SELECT balance FROM accounts WHERE id = 1 -> cur",
                   "UPDATE accounts SET balance = {cur - 50} WHERE id = 1",
                   "UPDATE accounts SET balance = balance + 50 WHERE id = 2"
               ])
               .add_operation("transfer_2_to_1", [
                   "SELECT balance FROM accounts WHERE id = 2 -> cur",
                   "UPDATE accounts SET balance = {cur - 50} WHERE id = 2",
                   "UPDATE accounts SET balance = balance + 50 WHERE id = 1"
               ])
               .assert_no_anomalies(workers=4, iterations=50, seed=42)
    )
    assert result.all_invariants_satisfied
```

### Pytest Fixture Integration

```python
import pytest

@pytest.mark.chaossql(workers=4, duration="5s", seed=100)
def test_inventory_under_concurrency(chaossql_runner):
    report = chaossql_runner.run_scenario("examples/inventory_oversell/chaos.yaml")
    assert report.is_clean
```

---

## ⚡ TypeScript / Node.js SDK (`@chaossql/test` / `npm install @chaossql/test`)

Native, strongly-typed fluent testing SDK compatible with Vitest, Jest, and Node.js Test Runner:

```typescript
import { describe, it, expect } from 'vitest';
import { ChaosHarness } from '@chaossql/test';

describe('Concurrency Isolation Suite', () => {
  it('should detect write skew in hospital on-call roster', async () => {
    const harness = new ChaosHarness({ driver: 'sqlite' });

    const result = await harness
      .withSchema(`
        CREATE TABLE doctors (id INT PRIMARY KEY, name TEXT, on_call INT);
        INSERT INTO doctors VALUES (1, 'Alice', 1), (2, 'Bob', 1);
      `)
      .withInvariant(
        'at_least_one_doctor_on_call',
        'SELECT count(*) as active FROM doctors WHERE on_call = 1;',
        'active >= 1'
      )
      .addOperation('alice_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 1 AND {active > 1}'
      ])
      .addOperation('bob_leaves', [
        'SELECT count(*) as cnt FROM doctors WHERE on_call = 1 -> active',
        'UPDATE doctors SET on_call = 0 WHERE id = 2 AND {active > 1}'
      ])
      .runAndShrink({ workers: 2, iterations: 20, seed: 1337 });

    expect(result.anomalyDetected).toBe(true);
    expect(result.anomalyCode).toBe('A5B');
    expect(result.minimalOperations.length).toBe(2);
  });
});
```

---

## 🛡️ OASIS SARIF 2.1.0 & GitHub Code Scanning

Export industry-standard SARIF 2.1.0 security reports to display concurrency race conditions as automated **GitHub Security Advisories**:

```yaml
# .github/workflows/concurrency-audit.yml
name: Concurrency Audit
on: [push, pull_request]

jobs:
  chaos:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run ChaosSQL Fuzzer
        run: |
          go install github.com/bregaldahq/chaossql/cmd/chaossql@latest
          chaossql run examples/banking_lost_update/chaos.yaml --export-sarif results.sarif || true

      - name: Upload SARIF to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Supported Security Rules:
* `chaossql/P4-lost-update` (Error - CWE-362)
* `chaossql/A5B-write-skew` (Error - CWE-362)
* `chaossql/A5A-read-skew` (Warning - CWE-362)
* `chaossql/G0-dirty-write` (Error - CWE-362)
* `chaossql/G1a-dirty-read` (Error - CWE-362)
* `chaossql/G1b-intermediate-read` (Error - CWE-362)
* `chaossql/G1c-circular-info` (Error - CWE-362)
* `chaossql/G2-anti-dependency` (Error - CWE-362)
* `chaossql/G-DL-deadlock` (Warning - CWE-362)

---

## 🌪️ Autonomous Multi-Engine Swarm & Adversarial Mutations (v1.4)

ChaosSQL v1.4 delivers automated, continuous multi-agent fuzzing, adversarial scenario mutation, and cross-engine differential isolation testing.

### 1. Adversarial Scenario Mutation (`chaossql mutate`)

Generate valid, novel variants of canonical scenarios by perturbing delays, nesting savepoints, permuting causal steps, and inverting lock hierarchies:

```bash
# Generate 5 mutated variations of a scenario into ./mutated
chaossql mutate examples/banking_lost_update/chaos.yaml --variants 5 --output-dir ./mutated

# Specify a custom deterministic seed and emit structured JSON
chaossql mutate examples/deadlock_cycle/chaos.yaml --variants 10 --seed 1337 --output-dir ./mutated --json
```

**Operational Flags:**
* `--variants int` (default `5`): Number of mutated scenario specifications to generate.
* `--output-dir string` (default `"./mutated"`): Target output directory where variants (`variant_0.yaml`, `variant_1.yaml`, ...) and referenced SQL files are saved.
* `--seed uint` (default `42`): Deterministic PRNG seed for reproducible mutation sequences.
* `--json`: Emit structured JSON describing the generated variants and applied operators.

**Adversarial Mutation Operators:**
* **Micro-Jitter Delay Perturbation (`InterleaveDelayMutation`)**: Injects randomized micro-jitter sleep intervals between transaction steps, widening race windows.
* **Nested LIFO Savepoint Lifecycle (`SavepointRollbackMutation`)**: Injects balanced `SAVEPOINT sp` ... `[ROLLBACK TO sp]` ... `RELEASE SAVEPOINT sp` stacks to test subtransaction isolation and lock retention.
* **Causal DAG Topological Step Shuffle (`StepShuffleMutation`)**: Models intra-transaction dependencies (captured variables `{var}`, table writes) as a DAG and performs random topological permutations without breaking causal validity.
* **Lock Acquisition Inversion (`LockOrderInversionMutation`)**: Inverts resource access sequences across concurrent transactions updating shared keys to deliberately provoke deadlocks ($G\text{-DL}$).

---

### 2. Multi-Engine Differential Swarm (`chaossql swarm`)

Execute mutated scenario suites concurrently across multiple database engines, comparing isolation anomalies, invariant violations, and serialization aborts:

```bash
# Run differential matrix across SQLite and Mock driver
chaossql swarm diff --scenarios-dir ./examples/banking_lost_update --drivers sqlite,mock

# Execute against production engines with Markdown Step Summary for GitHub Actions
chaossql swarm diff --scenarios-dir ./mutated --drivers sqlite,mock,postgres,mysql --concurrency 8 --markdown-summary summary.md

# Aliased execution in run mode with JSON telemetry
chaossql swarm run --scenarios-dir ./mutated --drivers sqlite,postgres --json
```

**Operational Flags:**
* `--scenarios-dir string` (default `"./examples"`): Directory or file path containing `chaos.yaml` and `variant_*.yaml` specs.
* `--drivers string` (default `"sqlite,mock"`): Comma-separated list of database drivers (`sqlite`, `mock`, `postgres`, `mysql`). Environment variables `POSTGRES_DSN` and `MYSQL_DSN` are automatically resolved.
* `--concurrency int` (default `4`): Bounded worker pool concurrency for parallel differential executions.
* `--markdown-summary string`: Path to write a GitHub Flavored Markdown (GFM) summary report (tailored for `$GITHUB_STEP_SUMMARY` and PR comments).
* `--json`: Output full differential matrix report as structured JSON.

---

### 3. Headless WebAssembly & Web Worker Stress Harness

Validate client-side concurrency fuzzer stability directly in headless Node.js V8 or modern browsers:

```bash
# Run 100 consecutive in-engine stress runs inside headless Web Worker VM
make stress-wasm
```

* **Memory Stability Bounds**: Proves bounded WebAssembly linear memory growth ($\le 32\text{MB}$) and process RSS stability ($< 100\text{MB}$) across extended stress runs without heap leakage.
* **60 FPS Non-Blocking Guarantee**: Asserts SVG Adya Direct Serialization Graph layout computation and main thread responsiveness remain strictly $< 16.6\text{ms}$ ($P_{95} = 0.184\text{ms}$, 0% jank frames).

---

## 🔌 Layer-7 Transparent Database Reverse Proxy (`chaossql proxy`) (v1.4)

Test live applications and ORMs against concurrency defects **without modifying a single line of application code**. `chaossql proxy` sits transparently between your application and your database (PostgreSQL 3.0 or MySQL 8.0 wire protocol), intercepting SQL packets, injecting stochastic PCT micro-jitter and commit barriers, and analyzing transactions in real time with a live Shadow Dependency Serialization Graph (Shadow DSG):

```
  ┌─────────────────────────┐           ┌──────────────────────────────────────┐           ┌────────────────────────┐
  │   Application / ORM     │  ──────►  │       ChaosSQL L7 Proxy (:5433)      │  ──────►  │    Upstream Database   │
  │ (psql / Prisma / pgx)   │  ◄──────  │  - Zero-CGO Protocol Stream Decoder  │  ◄──────  │ (PostgreSQL 16 / MySQL)│
  └─────────────────────────┘           │  - Stochastic PCT Micro-Jitter       │           └────────────────────────┘
                                        │  - Pre-Commit Barrier Injection      │
                                        │  - Real-time Shadow DSG Conflict DAG │
                                        └──────────────────────────────────────┘
                                                           │
                                                           ▼
                                        ┌──────────────────────────────────────┐
                                        │ Diagnostics & Live Telemetry (:8095) │
                                        │ - Real-time Adya Anomaly Detection   │
                                        │ - OASIS SARIF 2.1.0 Security Report  │
                                        │ - Web UI Dashboard & Telemetry JSON  │
                                        └──────────────────────────────────────┘
```

### Key Capabilities:
1. **Zero-CGO Layer-7 Protocol Decoders**: Fully native Go stream decoders for PostgreSQL Wire Protocol 3.0 (SSLRequest, Simple Query, Extended Query Parse/Bind/Execute) and MySQL Client/Server binary protocol.
2. **Stochastic Micro-Jitter & Commit Barriers**: Injects pseudo-random microsecond socket delays ($50\mu\text{s}$ to $5\text{ms}$) and stalls transactions right before the `COMMIT` packet is forwarded to upstream, maximizing the window for concurrent races.
3. **Live Shadow Serialization Graph (Shadow DSG)**: Constructs an online Adya conflict graph tracking active connections, identifying $wr, ww, rw$ edges on accessed keys, and classifying isolation anomalies on the fly:
   - $P4$ (Lost Update)
   - $A5B$ (Write Skew)
   - $G1a$ (Dirty Read)
   - $G2$ (Anti-Dependency Cycles)
4. **OASIS SARIF 2.1.0 & Web Dashboard**: Exports findings directly to SARIF files compatible with GitHub Code Scanning (`--export-sarif`) and provides an embedded real-time web UI and JSON telemetry API (`--ui-port`).

### Usage Example:
```bash
# Intercept PostgreSQL traffic on :5433 and forward to :5432
chaossql proxy \
  --listen 127.0.0.1:5433 \
  --upstream 127.0.0.1:5432 \
  --protocol postgres \
  --jitter-min 100us \
  --jitter-max 1ms \
  --commit-barrier-min 1ms \
  --commit-barrier-max 5ms \
  --export-sarif proxy_report.sarif \
  --ui-port 8095
```

---

## 🛠️ CLI Subcommands & Operational Flags

| Command | Syntax | Purpose |
| :--- | :--- | :--- |
| `run` | `chaossql run <config.yaml>` | Execute concurrency fuzzing, invariant checking, and causal shrinking |
| `demo` | `chaossql demo [scenario_name]` | Run one of the 10 built-in demonstration scenarios |
| `ui` | `chaossql ui <trace.json> [--port 8090]` | Launch local web inspector with Gantt swimlanes and Adya graph |
| `diff` | `chaossql diff <config.yaml> --drivers sqlite,postgres` | Run differential fuzzing across multiple database engines |
| `mutate` | `chaossql mutate <scenario.yaml> [flags]` | Generate adversarial scenario variations via stochastic AST & schedule mutations |
| `swarm` | `chaossql swarm [diff\|run] [flags]` | Execute multi-engine differential swarm fuzzing across databases |
| `proxy` | `chaossql proxy --listen :5433 --upstream :5432 [flags]` | Start transparent Layer-7 reverse proxy with PCT micro-jitter and live Shadow DSG |
| `replay` | `chaossql replay <trace.json>` | Deterministically reproduce a recorded trace using identical seed and schedule |
| `bench` | `chaossql bench [--ops 1000000]` | Measure raw engine throughput and PRNG interleaving performance |
| `validate` | `chaossql validate <config.yaml>` | Statically validate DSL syntax, expressions, and schema consistency |
| `init` | `chaossql init <directory> [--driver sqlite]` | Scaffold a new scenario folder with boilerplate schema, seed, and config |
| `matrix` | `chaossql matrix` | Print the empirical Hermitage isolation anomaly matrix for SQLite, Postgres, and MySQL |

---

## 📊 Enterprise Reliability & Quality Gate

| Verification Metric | Specification | Status |
| :--- | :--- | :--- |
| **Deterministic Replay** | Identical PRNG seed guarantees identical worker scheduling and results | `PASS` (100% Convergence) |
| **Concurrency Safety** | Verified with Go race detector (`go test -race ./...`) | `PASS` (0 Data Races) |
| **Static Code Quality** | Validated with `go vet ./...` with zero warnings | `PASS` |
| **Pure Go Portability** | Zero CGO dependencies (`CGO_ENABLED=0`) across Linux, macOS, and Windows | `PASS` (13.9M ops/s) |
| **Causal Reduction Ratio** | $ddmin$ achieves $>80\%$ reduction from noisy traces to 1-minimal reproductions | `PASS` ($<200\text{ms}$) |
| **Multi-Engine Swarm Matrix** | Cross-engine divergence detection across SQLite, Postgres 16, MySQL 8.0, and Mock | `PASS` (Deterministic Sync) |
| **WASM Memory & 60 FPS Bounds** | WebAssembly linear memory bounded (< 32MB) and zero jank frames (< 16.6ms) | `PASS` (100+ stress runs) |

---

## 📜 License & Governance

ChaosSQL is open-source software licensed under the **[MIT License](LICENSE)**.  
For release history and version notes, see **[`CHANGELOG.md`](CHANGELOG.md)**.  
For security guidelines and responsible disclosure, review **[`SECURITY.md`](SECURITY.md)**.

Architected and maintained by **[Ricardo Bregalda](https://github.com/bregaldahq)** at **[Studio Bregalda](https://bregalda.com)**.
