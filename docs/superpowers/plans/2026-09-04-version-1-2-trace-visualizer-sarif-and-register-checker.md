# ChaosSQL Version 1.2: Trace Visualizer, SARIF Reporter & Elle Register Checker Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Version 1.2 of ChaosSQL, featuring an Elle-style register linearizability analyzer (G1b detection), OASIS SARIF 2.1.0 security reporting for GitHub Code Scanning, an interactive local trace visualizer web UI (`chaossql ui`), and the 10th flagship scenario (foreign key cascade deadlock).

**Architecture:**
- `internal/analyzer/register.go`: Register and list-append history linearizability analyzer detecting G1b (Intermediate Reads) and fractured reads.
- `internal/reporter/sarif.go`: Standardized SARIF 2.1.0 JSON generator with GitHub Code Scanning rule definitions.
- `cmd/chaossql/ui.go` & `internal/reporter/ui.go`: Standalone HTTP server rendering an interactive SVG/HTML trace timeline and Adya force graph.
- `examples/foreign_key_cascade_deadlock/`: 10th Flagship Scenario modeling parent/child relational lock hierarchy contention.
- `tools/harness_check.go`: Update quality gate for 41 artifacts.

**Tech Stack:** Go 1.23+, modernc.org/sqlite, Lipgloss, OASIS SARIF 2.1.0, HTML5/SVG/CSS.

**Spec:** `specs/13_interactive_visualizer_sarif_and_register_checker.md`

## Global Constraints
- Zero CGO (`CGO_ENABLED=0`) across all packages.
- Zero data races (`go test -race ./...` must pass 100%).
- Commit and push to Git after each completed task.
- Subagents use `Model: "inherit"`.

---

### Task 1: Elle-Style Register Linearizability Analyzer (`internal/analyzer/register.go`)

**Files:**
- Create: `internal/analyzer/register.go`
- Create: `internal/analyzer/register_test.go`
- Modify: `internal/domain/types.go` (add `AnomalyG1bIntermediateRead`)

**Interfaces:**
- Produces: `CheckRegisterLinearizability(events []RegisterEvent) (*RegisterAnalysisResult, error)`
- Detects:
  - `G1b_INTERMEDIATE_READ`: Transaction reads an intermediate uncommitted overwrite from another transaction.
  - `FRACTURED_READ`: Reading an append-only register collection that misses prior committed items.

- [x] **Step 1: In `internal/domain/types.go`, add `AnomalyG1bIntermediateRead AnomalyType = "G1B_INTERMEDIATE_READ"`**
- [x] **Step 2: Write comprehensive unit tests in `internal/analyzer/register_test.go`**
- [x] **Step 3: Implement `internal/analyzer/register.go`**
- [x] **Step 4: Run `go test -v -race ./internal/analyzer/...` and verify 100% pass**
- [x] **Step 5: Commit and push**

---

### Task 2: OASIS SARIF 2.1.0 Security Reporter (`internal/reporter/sarif.go`)

**Files:**
- Create: `internal/reporter/sarif.go`
- Create: `internal/reporter/sarif_test.go`
- Modify: `cmd/chaossql/main.go` (add `--export-sarif <file>` flag)

**Interfaces:**
- Produces: `GenerateSARIFReport(spec domain.Spec, results []domain.InvariantResult, graph *analyzer.AdyaGraph, shrink *domain.ShrinkResult) (string, error)`
- Complies with SARIF 2.1.0 schema for GitHub Security / Code Scanning integration.

- [x] **Step 1: Write comprehensive unit tests in `internal/reporter/sarif_test.go`**
- [x] **Step 2: Implement `internal/reporter/sarif.go`**
- [x] **Step 3: Wire `--export-sarif` flag into `cmd/chaossql/main.go`**
- [x] **Step 4: Run `go test -v -race ./internal/reporter/...` and verify 100% pass**
- [x] **Step 5: Commit and push**

---

### Task 3: Interactive Trace Visualizer Web UI (`cmd/chaossql/ui.go` & `internal/reporter/ui.go`)

**Files:**
- Create: `internal/reporter/ui.go`
- Create: `cmd/chaossql/ui.go`
- Create: `cmd/chaossql/ui_test.go`
- Modify: `cmd/chaossql/main.go` (register `ui` command and `--ui` flag on `run`)

**Interfaces:**
- Produces: `chaossql ui <trace.json> [--port 8090]` and `chaossql run scenario.yaml --ui`.
- Starts a local HTTP server serving a single-page interactive trace Gantt timeline, force-directed Adya conflict graph, and $ddmin$ reduction comparison.

- [ ] **Step 1: Implement `internal/reporter/ui.go` with embedded HTML/SVG single-page trace viewer**
- [x] **Step 2: Implement `cmd/chaossql/ui.go` with Cobra command**
- [x] **Step 3: Write tests in `cmd/chaossql/ui_test.go`**
- [ ] **Step 4: Run `go test -v -race ./cmd/chaossql/...` and verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 4: 10th Flagship Scenario: Foreign Key Cascade Deadlock

**Files:**
- Create: `examples/foreign_key_cascade_deadlock/schema.sql`
- Create: `examples/foreign_key_cascade_deadlock/seed.sql`
- Create: `examples/foreign_key_cascade_deadlock/chaos.yaml`
- Create: `examples/foreign_key_cascade_deadlock/README.md`
- Modify: `internal/domain/parser_test.go`
- Modify: `cmd/chaossql/main.go` (support `chaossql demo fk` / `cascade`)
- Modify: `Makefile` (add scenario 10 to `make demo`)

- [x] **Step 1: Create the scenario files**
- [x] **Step 2: Register in `parser_test.go`, `cmd/chaossql/main.go`, and `Makefile`**
- [x] **Step 3: Run `go test -v -race ./...` and verify live execution**
- [x] **Step 4: Commit and push**

---

### Task 5: Quality Gate & Version 1.2 Release Verification

**Files:**
- Modify: `tools/harness_check.go` (audit 41 artifacts)
- Modify: `README.md` (document v1.2 features)

- [ ] **Step 1: Update `tools/harness_check.go` for 41 artifacts**
- [x] **Step 2: Run `make verify` and `make demo`**
- [x] **Step 3: Commit and push to GitHub, verifying green CI**
