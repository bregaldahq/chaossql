# Version 1.1: Go Developer Testing SDK, Smart Faker Generators & GitHub Action Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deliver ChaosSQL v1.1 comprising the programmatic Go testing SDK (`pkg/chaostest`), smart parameter generators (`$faker_email`, `$faker_name`, `$faker_phone`, `$monotonic_counter`), the official `action.yml` GitHub Action, and 31 harness quality gate contracts.

**Architecture:**
- Create `pkg/chaostest/chaostest.go` providing a fluent API for `*testing.T` and automated $ddmin$ assertion helpers.
- Enhance `internal/engine/prng.go` with deterministic faker & counter generators.
- Create root-level `action.yml` composite action for GitHub Actions marketplace.
- Update `tools/harness_check.go` for 31 artifacts.

**Tech Stack:** Go 1.23+, `modernc.org/sqlite`, `github.com/expr-lang/expr`, `github.com/charmbracelet/lipgloss`.

**Spec:** `specs/11_version_1_1_developer_sdk_and_smart_generators.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Smart Parameter Generators Expansion (`internal/engine/prng.go`)

**Files:**
- Modify: `internal/engine/prng.go`
- Test: `internal/engine/prng_test.go`

**Interfaces:**
- Produces: `EvaluateGenerator(expr string, r *rand.Rand)` supporting `$faker_email()`, `$faker_name()`, `$faker_phone()`, `$monotonic_counter(start, step)`.

- [ ] **Step 1: Write failing unit test in `internal/engine/prng_test.go` for new faker generators**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement `$faker_email()`, `$faker_name()`, `$faker_phone()`, `$monotonic_counter(start, step)` in `internal/engine/prng.go`**
- [ ] **Step 4: Run test to verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 2: Go Developer Testing SDK (`pkg/chaostest`)

**Files:**
- Create: `pkg/chaostest/chaostest.go`
- Test: `pkg/chaostest/chaostest_test.go`

**Interfaces:**
- Produces: `type Tester struct`, `New(t *testing.T) *Tester`, `WithDriver`, `WithSchema`, `WithSeed`, `WithInvariant`, `AddOperation`, `Run`, `AssertNoAnomalies`.

- [ ] **Step 1: Write comprehensive test in `pkg/chaostest/chaostest_test.go` testing both passing invariant and failing lost update reproduction**
- [ ] **Step 2: Run test to verify failure**
- [ ] **Step 3: Implement `pkg/chaostest/chaostest.go`**
- [ ] **Step 4: Run test to verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 3: Official GitHub Action (`action.yml`)

**Files:**
- Create: `action.yml`

**Interfaces:**
- Inputs: `spec-path`, `workers`, `iterations`, `seed`, `export-html`, `export-junit`, `export-summary`.
- Runs: `chaossql run` binary with parameters.

- [ ] **Step 1: Create `action.yml` composite action**
- [ ] **Step 2: Validate YAML syntax**
- [ ] **Step 3: Commit and push**

---

### Task 4: Quality Gate & Version 1.1 Verification

**Files:**
- Modify: `tools/harness_check.go`
- Modify: `README.md`

- [ ] **Step 1: Update `tools/harness_check.go` for 31 artifacts**
- [ ] **Step 2: Run `make verify` and `make demo`**
- [ ] **Step 3: Commit and push to GitHub, verifying CI turns green**
