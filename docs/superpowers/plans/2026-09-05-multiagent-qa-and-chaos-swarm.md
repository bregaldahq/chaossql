# ChaosSQL Multi-Agent QA & Swarm Concurrency Fuzzing Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:dispatch-parallel-agents to execute this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish an autonomous Multi-Agent QA & Stress Testing Swarm for ChaosSQL that runs concurrent adversarial SQL mutations, cross-database differential isolation benchmarks, headless browser WebAssembly stress testing, and automated evidence synthesis with zero human intervention.

**Architecture:**
- **Coordinator Agent (Lead Orchestrator)**: Manages test cycles, distributes scenarios to worker agents, and aggregates results.
- **Fuzzing & Mutation Agent (Adversarial Swarm)**: Uses AST and grammar-based SQL mutation to generate complex transaction schedules (CTEs, nested transactions, savepoints, foreign key cascades).
- **Differential Verification Agent (Matrix Auditor)**: Orchestrates simultaneous differential executions across SQLite (in-memory & file), PostgreSQL, and MySQL, verifying Hermitage isolation invariants.
- **WASM & Client-Side Performance Agent (Headless Browser QA)**: Automates headless browser Web Worker load testing, measures WASM heap allocations, and asserts 60 FPS SVG rendering during high-frequency cycles.
- **Evidence & Synthesis Agent (Reporting Scribe)**: Generates automated GitHub Step Summaries, SARIF security diagnostics, reproducible Go standalone tests (`repro_test.go`), and bilingual documentation updates.

**Tech Stack:** Go 1.23+, Python 3, Node.js v20+, WebAssembly (`GOOS=js GOARCH=wasm`), SQLite, PostgreSQL, MySQL, GitHub Actions, SARIF 2.1.0, JUnit XML.

**Spec:** `specs/15_multiagent_qa_and_swarm_fuzzing.md` (to be implemented).

## Global Constraints
- Zero CGO (`CGO_ENABLED=0`) across all Go packages.
- Zero server dependency for client-side WASM verification.
- All generated scenarios must pass `chaossql validate` before execution.
- Deterministic reproducibility: every anomaly must output a PRNG seed and minimal causal reproduction schedule ($ddmin$).
- Subagents execute with `Model: "inherit"`.

---

### Task 1: Scenario Mutation Engine & Adversarial Generator (`pkg/mutator/`)

**Files:**
- Create: `pkg/mutator/mutator.go`
- Create: `pkg/mutator/mutator_test.go`
- Create: `cmd/chaossql/mutate.go`

**Interfaces:**
- Consumes: Canonical scenario YAML files (`examples/*/chaos.yaml`).
- Produces:
  - `mutator.MutateScenario(spec domain.Spec, opts MutationOptions) ([]domain.Spec, error)`
  - CLI: `chaossql mutate <scenario.yaml> --variants 10 --output-dir /tmp/mutated/`

- [ ] **Step 1: Write failing unit test for scenario mutator**
- [ ] **Step 2: Implement transaction interleaving and statement mutators**
- [ ] **Step 3: Implement CLI `chaossql mutate` command**
- [ ] **Step 4: Run test to verify it passes**
- [ ] **Step 5: Commit and review**

---

### Task 2: Multi-Engine Differential Runner (`internal/swarm/diff_runner.go`)

**Files:**
- Create: `internal/swarm/diff_runner.go`
- Create: `internal/swarm/diff_runner_test.go`
- Create: `cmd/chaossql/swarm.go`

**Interfaces:**
- Consumes: Mutated scenarios, database connection strings (SQLite, Postgres, MySQL).
- Produces:
  - `swarm.ExecuteDifferentialMatrix(ctx context.Context, specs []domain.Spec, drivers []string) (*DifferentialReport, error)`
  - CLI: `chaossql swarm run --scenarios-dir ./mutated --drivers sqlite,postgres`

- [ ] **Step 1: Write failing test for multi-engine differential coordinator**
- [ ] **Step 2: Implement concurrent multi-driver schedule dispatch**
- [ ] **Step 3: Implement divergence detector (isolating differences in anomaly classification)**
- [ ] **Step 4: Verify with test suite**
- [ ] **Step 5: Commit and review**

---

### Task 3: Headless Browser WASM & UI Stress Harness (`tools/headless_worker_stress.js`)

**Files:**
- Create: `tools/headless_worker_stress.js`
- Create: `site/assets/wasm-bench.js`

**Interfaces:**
- Consumes: `http://localhost:8080`, `chaossql.wasm`, `site/app.js`.
- Produces:
  - Stress testing report: throughput (ops/sec), Web Worker memory leakage over 1,000 iterations, frame rendering benchmarks.
  - Automated detection of UI thread blocking (> 16.6ms frame time).

- [ ] **Step 1: Implement headless worker loop running 100 consecutive scenarios**
- [ ] **Step 2: Measure memory growth in V8 WebAssembly linear memory**
- [ ] **Step 3: Measure SVG DOM node layout latency**
- [ ] **Step 4: Integrate into Makefile `make test-wasm-stress`**
- [ ] **Step 5: Commit and review**

---

### Task 4: Automated Evidence Synthesizer & GitHub Actions Matrix (`.github/workflows/swarm.yml`)

**Files:**
- Create: `.github/workflows/swarm.yml`
- Create: `internal/reporter/swarm_summary.go`
- Create: `internal/reporter/swarm_summary_test.go`

**Interfaces:**
- Consumes: Differential and mutation reports.
- Produces:
  - Aggregated markdown matrix for PR comments.
  - SARIF 2.1.0 security findings upload to GitHub Code Scanning.
  - Standalone `repro_test.go` files archived as CI artifacts.

- [ ] **Step 1: Implement Swarm summary aggregator**
- [ ] **Step 2: Configure matrix workflow in `.github/workflows/swarm.yml`**
- [ ] **Step 3: Test local GitHub step summary generation**
- [ ] **Step 4: Commit and review**

---

### Task 5: Multi-Agent Parallel Execution Dispatcher

**Files:**
- Create: `docs/superpowers/plans/swarm-execution-ledger.md`

**Interfaces:**
- Coordinates:
  - Subagent A: Mutation generation
  - Subagent B: SQLite vs Postgres differential stress
  - Subagent C: Browser WASM memory & stress audit
  - Subagent D: Evidence aggregation & documentation

- [ ] **Step 1: Dispatch parallel worker subagents using `dispatch-parallel-agents`**
- [ ] **Step 2: Consolidate findings into central ledger**
- [ ] **Step 3: Final whole-swarm review gate**
