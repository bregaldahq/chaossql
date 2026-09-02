# Milestone 6: Temporal Invariants, Multi-Transaction G2 Anti-Dependency Cycles & CI/CD Reporters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement temporal/event-stream invariants (`internal/evaluator/temporal.go`), multi-transaction $G2$ anti-dependency cycle inference ($T_1 \xrightarrow{rw} \dots \xrightarrow{rw} T_k \xrightarrow{rw} T_1$), JUnit XML and GitHub step summary exporters, and Flagship Demo 8 (`examples/ticket_booking_anti_dependency`).

**Architecture:**
- Create `internal/evaluator/temporal.go` to evaluate temporal invariants over traces.
- Update `internal/domain/types.go` to add `AnomalyG2AntiDependency` and `TemporalInvariantConfig`.
- Update `internal/analyzer/adya.go` to classify $G2$ anti-dependency cycles for arbitrary cycle length $k \ge 2$.
- Create `internal/reporter/junit.go` and `internal/reporter/summary.go`.
- Create `examples/ticket_booking_anti_dependency/` (Scenario 8).

**Tech Stack:** Go 1.23+, `modernc.org/sqlite`, `github.com/spf13/cobra`.

**Spec:** `specs/09_temporal_invariants_and_g2_cycles.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Temporal & Event-Stream Invariant Evaluator (`internal/evaluator/temporal.go`)

**Files:**
- Create: `internal/evaluator/temporal.go`
- Test: `internal/evaluator/temporal_test.go`
- Modify: `internal/domain/types.go`

- [ ] **Step 1: Add `TemporalInvariantConfig` to `internal/domain/types.go`**
- [ ] **Step 2: Write failing unit tests for temporal invariants in `internal/evaluator/temporal_test.go`**
- [ ] **Step 3: Implement `EvaluateTemporalInvariants(trace domain.ExecutionTrace, invs []domain.TemporalInvariantConfig) []domain.InvariantResult`**
- [ ] **Step 4: Run tests and verify 100% pass with `-race`**
- [ ] **Step 5: Commit and push**

---

### Task 2: Multi-Transaction G2 Anti-Dependency Cycle Inference (`internal/analyzer/adya.go`)

**Files:**
- Modify: `internal/domain/types.go`
- Modify: `internal/analyzer/adya.go`
- Test: `internal/analyzer/adya_test.go`

- [ ] **Step 1: Add `AnomalyG2AntiDependency` to `domain.AnomalyType`**
- [ ] **Step 2: Write failing test `TestAdyaMultiTransaction_G2` (3-transaction cycle $T_1 \xrightarrow{rw} T_2 \xrightarrow{rw} T_3 \xrightarrow{rw} T_1$)**
- [ ] **Step 3: Implement generalized $G2$ cycle classification in `internal/analyzer/adya.go`**
- [ ] **Step 4: Run tests and verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 3: JUnit XML & GitHub Actions Summary Exporters (`internal/reporter/junit.go` & `internal/reporter/summary.go`)

**Files:**
- Create: `internal/reporter/junit.go`
- Create: `internal/reporter/summary.go`
- Test: `internal/reporter/junit_test.go`
- Test: `internal/reporter/summary_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing tests for JUnit XML and GitHub summary generators**
- [ ] **Step 2: Implement `GenerateJUnitXML` and `GenerateGitHubSummaryMarkdown`**
- [ ] **Step 3: Add `--export-junit` and `--export-summary` CLI flags in `cmd/chaossql/main.go`**
- [ ] **Step 4: Run tests and verify pass**
- [ ] **Step 5: Commit and push**

---

### Task 4: Flagship Scenario 8: Ticket Seat Reservation Anti-Dependency Cycle ($G2$)

**Files:**
- Create: `examples/ticket_booking_anti_dependency/schema.sql`
- Create: `examples/ticket_booking_anti_dependency/seed.sql`
- Create: `examples/ticket_booking_anti_dependency/chaos.yaml`
- Create: `examples/ticket_booking_anti_dependency/README.md`
- Modify: `internal/domain/parser_test.go`
- Modify: `cmd/chaossql/main.go`
- Modify: `cmd/chaossql/matrix.go`
- Modify: `Makefile`

- [ ] **Step 1: Create Ticket Booking $G2$ scenario files**
- [ ] **Step 2: Register in CLI demo and matrix command**
- [ ] **Step 3: Run demo and verify $G2$ detection and $ddmin$ reduction**
- [ ] **Step 4: Commit and push**

---

### Task 5: Milestone 6 Verification & Quality Gate

**Files:**
- Modify: `README.md`
- Modify: `tools/harness_check.go`

- [ ] **Step 1: Update `tools/harness_check.go` for 28 artifacts**
- [ ] **Step 2: Run `make verify` and `make demo` (8 scenarios)**
- [ ] **Step 3: Update documentation and badges**
- [ ] **Step 4: Commit and push to GitHub, verifying CI turns green**
