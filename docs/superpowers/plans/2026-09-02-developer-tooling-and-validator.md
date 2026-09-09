# Milestone 7: Developer Tooling, Static Scenario Validator & Deadlock Diagnostics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `chaossql init` (scenario scaffolder), `chaossql validate` (static scenario linter and invariant expression compiler), Scenario 9 (`examples/deadlock_cycle`), and quality gate contracts.

**Architecture:**
- Create `cmd/chaossql/init.go` and unit tests.
- Create `cmd/chaossql/validate.go` and unit tests.
- Create `examples/deadlock_cycle/` (Scenario 9).
- Update `tools/harness_check.go` for 30 artifacts.

**Tech Stack:** Go 1.23+, `github.com/spf13/cobra`, `github.com/expr-lang/expr`, `modernc.org/sqlite`.

**Spec:** `specs/10_developer_tooling_and_static_validator.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Scenario Scaffolder CLI (`cmd/chaossql/init.go`)

**Files:**
- Create: `cmd/chaossql/init.go`
- Test: `cmd/chaossql/init_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing unit test for `initCmd`**
- [ ] **Step 2: Implement `newInitCmd()` creating `schema.sql`, `seed.sql`, `chaos.yaml`, and `README.md`**
- [ ] **Step 3: Register in `main.go` and verify tests pass with `-race`**
- [ ] **Step 4: Commit and push**

---

### Task 2: Static Scenario Linter & Expression Validator (`cmd/chaossql/validate.go`)

**Files:**
- Create: `cmd/chaossql/validate.go`
- Test: `cmd/chaossql/validate_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing unit test for `validateCmd` on valid and invalid scenarios**
- [ ] **Step 2: Implement `newValidateCmd()` verifying YAML schema, file existence, and compiling invariant expressions with `expr.Compile`**
- [ ] **Step 3: Register in `main.go` and verify tests pass**
- [ ] **Step 4: Commit and push**

---

### Task 3: Flagship Scenario 9: Deadlock Cycle & Timeout Diagnostics (`examples/deadlock_cycle/`)

**Files:**
- Create: `examples/deadlock_cycle/schema.sql`
- Create: `examples/deadlock_cycle/seed.sql`
- Create: `examples/deadlock_cycle/chaos.yaml`
- Create: `examples/deadlock_cycle/README.md`
- Modify: `internal/domain/parser_test.go`
- Modify: `cmd/chaossql/main.go`
- Modify: `cmd/chaossql/matrix.go`
- Modify: `Makefile`

- [ ] **Step 1: Create Deadlock scenario files**
- [ ] **Step 2: Register in parser tests, main demo, matrix and Makefile**
- [ ] **Step 3: Run `make demo` and verify execution**
- [ ] **Step 4: Commit and push**

---

### Task 4: Milestone 7 Verification & Quality Gate

**Files:**
- Modify: `README.md`
- Modify: `tools/harness_check.go`

- [ ] **Step 1: Update `tools/harness_check.go` for 30 artifacts**
- [ ] **Step 2: Run `make verify` and `make demo` (9 scenarios)**
- [ ] **Step 3: Commit and push to GitHub, verifying CI turns green**
