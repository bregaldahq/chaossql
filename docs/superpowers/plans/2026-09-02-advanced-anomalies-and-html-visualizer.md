# Advanced Isolation Taxonomy, Dynamic Generators, HTML Visualizer & Benchmarks Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand ChaosSQL with the full Adya anomaly taxonomy ($A5A$ Read Skew, $G0$ Dirty Write, $G1c$ Circular Flow), rich deterministic parameter generators, interactive standalone HTML reporting with interactive vis.js graph visualization, a dedicated financial audit scenario, and a high-performance benchmarking command.

**Architecture:** Extend `internal/domain` with new anomaly types and generator functions; extend `internal/analyzer` with cycle classifiers for $A5A$, $G0$, $G1c$; enhance `internal/engine` with regex-based parametric generators; implement `internal/reporter/html.go` with embedded standalone HTML/SVG/vis.js; add `cmd/chaossql/bench.go` CLI command.

**Tech Stack:** Go 1.23+, `modernc.org/sqlite`, `github.com/jackc/pgx/v5`, `github.com/expr-lang/expr`, `github.com/charmbracelet/lipgloss`, `github.com/spf13/cobra`.

**Spec:** `specs/05_advanced_anomaly_taxonomy.md`

## Global Constraints
- Pure Go / Zero CGO compilation (`CGO_ENABLED=0`).
- Strict TDD: Write failing test first, verify failure, implement minimal code, verify pass with `-race`.
- Commit and push to Git after each completed task.

---

### Task 1: Domain & Analyzer Anomaly Taxonomy Expansion ($A5A$, $G0$, $G1c$)

**Files:**
- Modify: `internal/domain/types.go`
- Modify: `internal/analyzer/adya.go`
- Test: `internal/analyzer/adya_test.go`

**Interfaces:**
- `domain.AnomalyA5AReadSkew domain.AnomalyType = "A5A_READ_SKEW"`
- `domain.AnomalyG0DirtyWrite domain.AnomalyType = "G0_DIRTY_WRITE"`
- `domain.AnomalyG1cCircularInfo domain.AnomalyType = "G1C_CIRCULAR_INFO"`

- [ ] **Step 1: Write failing tests in `internal/analyzer/adya_test.go` for $A5A$, $G0$, and $G1c$ cycles**
- [ ] **Step 2: Run tests and verify failure**
- [ ] **Step 3: Implement cycle classification logic in `internal/analyzer/adya.go`**
- [ ] **Step 4: Run tests and verify 100% pass with `-race`**
- [ ] **Step 5: Commit and push**

---

### Task 2: Deterministic Rich Parameter Generators in Engine

**Files:**
- Modify: `internal/engine/prng.go`
- Modify: `internal/engine/runner.go`
- Test: `internal/engine/prng_test.go`

**Interfaces:**
- Support functions: `$random_int(min, max)`, `$random_choice(val1, val2, ...)`, `$random_string(len)`, `$uuid()`
- Deterministic sub-seed derivation per worker goroutine.

- [ ] **Step 1: Write failing tests for rich parameter generator expressions**
- [ ] **Step 2: Run tests and verify failure**
- [ ] **Step 3: Implement generator functions in `internal/engine/prng.go`**
- [ ] **Step 4: Run tests and verify 100% pass with `-race`**
- [ ] **Step 5: Commit and push**

---

### Task 3: Interactive Standalone HTML Report Synthesizer (`internal/reporter/html.go`)

**Files:**
- Create: `internal/reporter/html.go`
- Test: `internal/reporter/html_test.go`
- Modify: `cmd/chaossql/main.go`

**Interfaces:**
- `GenerateStandaloneHTMLReport(trace domain.ExecutionTrace, spec domain.Spec, graph *analyzer.AdyaGraph, shrinkResult *domain.ShrinkResult, invResults []domain.InvariantResult) string`

- [ ] **Step 1: Write failing tests in `internal/reporter/html_test.go`**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement `internal/reporter/html.go` with embedded CSS and interactive Vis.js graph**
- [ ] **Step 4: Run tests and verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 4: Flagship Scenario: Financial Audit Read Skew ($A5A$)

**Files:**
- Create: `examples/read_skew_financial_audit/chaos.yaml`
- Create: `examples/read_skew_financial_audit/schema.sql`
- Create: `examples/read_skew_financial_audit/seed.sql`
- Create: `examples/read_skew_financial_audit/README.md`
- Test: `internal/domain/parser_test.go`

- [ ] **Step 1: Write schema, seed, and chaos.yaml for account balance transfer vs audit read**
- [ ] **Step 2: Verify spec parsing in `parser_test.go`**
- [ ] **Step 3: Run scenario and verify $A5A$ Read Skew detection and $ddmin$ reduction**
- [ ] **Step 4: Commit and push**

---

### Task 5: High-Performance Benchmarking Command (`chaossql bench`)

**Files:**
- Create: `cmd/chaossql/bench.go`
- Modify: `cmd/chaossql/main.go`
- Test: `cmd/chaossql/bench_test.go`

- [ ] **Step 1: Write failing benchmark test for CLI execution**
- [ ] **Step 2: Implement `chaossql bench` command measuring TPS, PRNG throughput, and cycle detection latency**
- [ ] **Step 3: Verify execution in terminal**
- [ ] **Step 4: Commit and push**

---

### Task 6: Unified Quality Gate & CI Verification

**Files:**
- Modify: `Makefile`
- Modify: `README.md`

- [ ] **Step 1: Run `make verify`**
- [ ] **Step 2: Run `make demo` (now testing 4 scenarios)**
- [ ] **Step 3: Commit and push to GitHub, verifying CI turns green**
