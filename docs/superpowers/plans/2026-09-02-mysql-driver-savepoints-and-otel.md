# Milestone 3: MySQL Driver, Savepoints, OpenTelemetry Traces & Auction Dirty Write Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand ChaosSQL with pure-Go MySQL driver support (`github.com/go-sql-driver/mysql`), nested transaction savepoint primitives (`SAVEPOINT`, `ROLLBACK TO`), OpenTelemetry distributed trace export (`--export-otel`), and a 5th flagship demonstration scenario for $G0$ Dirty Write (`examples/dirty_write_auction`).

**Architecture:** Implement `internal/drivers/mysql.go`; extend `internal/engine/runner.go` to handle savepoint semantics; implement `internal/reporter/otel.go` for OTLP span generation; create `examples/dirty_write_auction/` demo.

**Tech Stack:** Go 1.23+, `github.com/go-sql-driver/mysql`, `modernc.org/sqlite`, `github.com/jackc/pgx/v5`, `github.com/charmbracelet/lipgloss`.

**Spec:** `specs/06_mysql_savepoints_and_otel.md`

## Global Constraints
- Pure Go / Zero CGO compilation.
- Strict TDD with `-race` detection.
- Commit and push to Git after each completed task.

---

### Task 1: Pure-Go MySQL Database Driver (`internal/drivers/mysql.go`)

**Files:**
- Create: `internal/drivers/mysql.go`
- Test: `internal/drivers/mysql_test.go`
- Modify: `internal/domain/types.go`

- [ ] **Step 1: Write failing tests in `internal/drivers/mysql_test.go`**
- [ ] **Step 2: Implement `MySQLDriver` adhering to `drivers.DatabaseDriver`**
- [ ] **Step 3: Run tests and verify pass**
- [ ] **Step 4: Commit and push**

---

### Task 2: Savepoint & Nested Transaction Fuzzing Primitives

**Files:**
- Modify: `internal/engine/runner.go`
- Modify: `internal/domain/types.go`
- Test: `internal/engine/runner_savepoint_test.go`

- [ ] **Step 1: Write failing test for savepoint execution and rollback**
- [ ] **Step 2: Implement savepoint tracking in `Runner`**
- [ ] **Step 3: Run tests and verify pass with `-race`**
- [ ] **Step 4: Commit and push**

---

### Task 3: OpenTelemetry Distributed Trace Synthesizer (`internal/reporter/otel.go`)

**Files:**
- Create: `internal/reporter/otel.go`
- Test: `internal/reporter/otel_test.go`
- Modify: `cmd/chaossql/main.go`

- [ ] **Step 1: Write failing test for OTLP trace JSON generation**
- [ ] **Step 2: Implement `GenerateOTLPTraceJSON` and `--export-otel` CLI flag**
- [ ] **Step 3: Run tests and verify pass**
- [ ] **Step 4: Commit and push**

---

### Task 4: Flagship Scenario: Auction Bidding Dirty Write ($G0$)

**Files:**
- Create: `examples/dirty_write_auction/schema.sql`
- Create: `examples/dirty_write_auction/seed.sql`
- Create: `examples/dirty_write_auction/chaos.yaml`
- Create: `examples/dirty_write_auction/README.md`
- Modify: `cmd/chaossql/main.go`
- Modify: `Makefile`

- [ ] **Step 1: Create auction dirty write scenario**
- [ ] **Step 2: Register in CLI demo and Makefile**
- [ ] **Step 3: Run demo and verify $G0$ detection and $ddmin$ reduction**
- [ ] **Step 4: Commit and push**

---

### Task 5: Milestone 3 Verification & Documentation Update

**Files:**
- Modify: `README.md`
- Modify: `tools/harness_check.go`

- [ ] **Step 1: Run `make verify` and `make demo`**
- [ ] **Step 2: Update documentation and badges**
- [ ] **Step 3: Commit and push to GitHub, verifying CI turns green**
