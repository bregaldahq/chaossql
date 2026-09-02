# Milestone 5: Stochastic Fault Injection, Dirty Read (G1a) Anomaly & Trace Replayer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement stochastic fault injection (`internal/faults`), $G1a$ Dirty Read / Aborted Read anomaly classification, interactive trace replay CLI command (`chaossql replay`), and the 7th flagship demonstration scenario (`examples/dirty_read_flash_crash`).

**Architecture:**
- Create `internal/faults/fault.go` with `FaultInjector` supporting forced rollbacks, latency spikes, and disconnects.
- Update `internal/domain/types.go` to support `FaultConfig` and `AnomalyG1aDirtyRead`.
- Update `internal/analyzer/adya.go` to classify $G1a$ Dirty Read cycles ($w_1 \dots r_2 \dots a_1$).
- Implement `cmd/chaossql/replay.go` (`chaossql replay <trace.json>`).
- Create `examples/dirty_read_flash_crash/` (Scenario 7).

**Tech Stack:** Go 1.23+, `modernc.org/sqlite`, `github.com/charmbracelet/lipgloss`, `github.com/spf13/cobra`.

**Spec:** `specs/08_fault_injection_and_dirty_reads.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Stochastic Fault Injector (`internal/faults/fault.go`)

**Files:**
- Create: `internal/faults/fault.go`
- Test: `internal/faults/fault_test.go`
- Modify: `internal/domain/types.go`
- Modify: `internal/engine/runner.go`

- [ ] **Step 1: Update `domain.EngineConfig` with `Faults domain.FaultConfig`**
- [ ] **Step 2: Write failing tests for `FaultInjector` in `internal/faults/fault_test.go`**
- [ ] **Step 3: Implement `FaultInjector` with `ShouldAbort()`, `GetLatencySpike()`, `ShouldDisconnect()`**
- [ ] **Step 4: Integrate `FaultInjector` into `internal/engine/runner.go`**
- [ ] **Step 5: Run tests and verify 100% pass with `-race`**
- [ ] **Step 6: Commit and push**

---

### Task 2: G1a Dirty Read / Aborted Read Anomaly Classification (`internal/analyzer/adya.go`)

**Files:**
- Modify: `internal/domain/types.go`
- Modify: `internal/analyzer/adya.go`
- Test: `internal/analyzer/adya_test.go`

- [ ] **Step 1: Add `AnomalyG1aDirtyRead` to `domain.AnomalyType`**
- [ ] **Step 2: Write failing unit test `TestAdyaDirtyReadG1a`**
- [ ] **Step 3: Implement $G1a$ cycle classification in `internal/analyzer/adya.go`**
- [ ] **Step 4: Run tests and verify 100% pass**
- [ ] **Step 5: Commit and push**

---

### Task 3: Interactive Trace Replayer CLI Command (`chaossql replay`)

**Files:**
- Create: `cmd/chaossql/replay.go`
- Test: `cmd/chaossql/replay_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing tests for `replayCmd`**
- [ ] **Step 2: Implement `chaossql replay <trace.json>` with Lipgloss chronological event swimlane rendering**
- [ ] **Step 3: Run tests and verify pass**
- [ ] **Step 4: Commit and push**

---

### Task 4: Flagship Scenario 7: Flash Crash Liquidation Dirty Read ($G1a$)

**Files:**
- Create: `examples/dirty_read_flash_crash/schema.sql`
- Create: `examples/dirty_read_flash_crash/seed.sql`
- Create: `examples/dirty_read_flash_crash/chaos.yaml`
- Create: `examples/dirty_read_flash_crash/README.md`
- Modify: `internal/domain/parser_test.go`
- Modify: `cmd/chaossql/main.go`
- Modify: `cmd/chaossql/matrix.go`
- Modify: `Makefile`

- [ ] **Step 1: Create Flash Crash $G1a$ scenario files**
- [ ] **Step 2: Register in CLI demo and matrix command**
- [ ] **Step 3: Run demo and verify detection and $ddmin$ reduction**
- [ ] **Step 4: Commit and push**

---

### Task 5: Milestone 5 Verification & Quality Gate

**Files:**
- Modify: `README.md`
- Modify: `tools/harness_check.go`

- [ ] **Step 1: Update `tools/harness_check.go` for 27 artifacts**
- [ ] **Step 2: Run `make verify` and `make demo` (7 scenarios)**
- [ ] **Step 3: Update documentation and badges**
- [ ] **Step 4: Commit and push to GitHub, verifying CI turns green**
