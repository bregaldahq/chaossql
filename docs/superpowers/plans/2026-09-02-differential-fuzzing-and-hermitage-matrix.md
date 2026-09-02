# Milestone 4: Differential Fuzzing, Hermitage Matrix & Circular Info Flow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement cross-engine differential isolation fuzzing (`chaossql diff`), automated Hermitage isolation matrix generator (`chaossql matrix`), and a 6th flagship demonstration scenario for $G1c$ Circular Information Flow (`examples/circular_info_crypto_arbitrage`).

**Architecture:** Add `internal/engine/diff.go`; implement `cmd/chaossql/diff.go` and `cmd/chaossql/matrix.go`; create crypto arbitrage $G1c$ demo.

**Tech Stack:** Go 1.23+, `modernc.org/sqlite`, `github.com/jackc/pgx/v5`, `github.com/go-sql-driver/mysql`, `github.com/charmbracelet/lipgloss`.

**Spec:** `specs/07_differential_fuzzing_and_matrix.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Cross-Engine Differential Fuzzer Engine (`internal/engine/diff.go`)

**Files:**
- Create: `internal/engine/diff.go`
- Test: `internal/engine/diff_test.go`
- Modify: `internal/domain/types.go`

- [ ] **Step 1: Write failing tests in `internal/engine/diff_test.go`**
- [ ] **Step 2: Implement `RunDifferentialFuzzing` comparing 2 database drivers**
- [ ] **Step 3: Run tests and verify pass**
- [ ] **Step 4: Commit and push**

---

### Task 2: Differential Fuzzing & Hermitage Matrix CLI Commands (`chaossql diff` & `chaossql matrix`)

**Files:**
- Create: `cmd/chaossql/diff.go`
- Create: `cmd/chaossql/matrix.go`
- Test: `cmd/chaossql/diff_test.go`
- Test: `cmd/chaossql/matrix_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing tests for `diff` and `matrix` commands**
- [ ] **Step 2: Implement CLI commands and Lipgloss matrix table rendering**
- [ ] **Step 3: Run tests and verify pass**
- [ ] **Step 4: Commit and push**

---

### Task 3: Flagship Scenario: Crypto Arbitrage Circular Information Flow ($G1c$)

**Files:**
- Create: `examples/circular_info_crypto_arbitrage/schema.sql`
- Create: `examples/circular_info_crypto_arbitrage/seed.sql`
- Create: `examples/circular_info_crypto_arbitrage/chaos.yaml`
- Create: `examples/circular_info_crypto_arbitrage/README.md`
- Modify: `internal/domain/parser_test.go`
- Modify: `cmd/chaossql/main.go`
- Modify: `Makefile`

- [ ] **Step 1: Create crypto arbitrage $G1c$ scenario**
- [ ] **Step 2: Register in CLI demo and Makefile**
- [ ] **Step 3: Run demo and verify $G1c$ detection and $ddmin$ reduction**
- [ ] **Step 4: Commit and push**

---

### Task 4: Milestone 4 Verification & Quality Gate

**Files:**
- Modify: `README.md`
- Modify: `tools/harness_check.go`

- [ ] **Step 1: Run `make verify` and `make demo` (now 6 scenarios)**
- [ ] **Step 2: Update documentation and badges**
- [ ] **Step 3: Commit and push to GitHub, verifying CI turns green**
