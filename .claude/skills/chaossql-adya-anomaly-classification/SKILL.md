---
name: chaossql-adya-anomaly-classification
description: How traces become Adya dependency graphs (WR/WW/RW edges), how cycles are found and classified into anomaly types (G0, G1a, G1c, P4, A5A, A5B, G2), the per-caller priority order, and the Elle-style register checker. Use when an anomaly label is wrong, when adding an anomaly type, or when changing SQL item extraction.
---

# Adya Anomaly Classification

Invariants say **that** something broke; the analyzer says **what kind** of
isolation anomaly the trace exhibits, by building a direct serialization graph
from the trace and classifying its cycles.

## When to use

- The reported `anomaly_type` looks wrong or is `UNKNOWN_INVARIANT_VIOLATION`.
- Adding an anomaly type or changing SQL item extraction.
- Touching `internal/analyzer` or any of its many callers.

## Pipeline

### 1. Item extraction (`extractItem`)

Per `EXEC` event SQL (case-insensitive prefix):
- `UPDATE t ... WHERE id = N` → write `t:N` (or `t` without the id match)
- `INSERT INTO t` → write `t` (table level)
- `DELETE FROM t ... WHERE id = N` → write `t:N` / `t`
- `SELECT ... FROM t ... WHERE id = N` → read `t:N` / `t` (table scan)
- fallback regexes for other shapes; otherwise ignored.

Only the literal pattern `WHERE ID = <int>` (single spaces, column named `id`)
produces row items. `account_id = 1`, `id=1`, `id IN (...)` become table-level
items.

### 2. Graph (`BuildGraph`)

- Transaction node ID: `T<workerID>-<opIndex>`.
- Aborted transactions: any tx with a `ROLLBACK` event (pre-scan).
- Only `EXEC` events are analyzed, in **trace order** (real completion order).
- Write on item: `WW` from last writer; `RW` from every reader of the item and
  of its table (when the item is a row); then it becomes last writer of the
  item and the table; readers of the item are cleared.
- Read on item: `WR` from last writer; a table-level read gets `WR` from every
  last writer of any row of that table.
- Edges carry `IsAbortedWriter` when the source tx rolled back; duplicates
  (same from/to/type/item) are skipped.

### 3. Cycles (`FindCycles`)

DFS with a recursion stack over `g.Nodes` (a map — iteration order is random,
so the order and exact set of reported cycles can vary between calls).

### 4. Classification (`ClassifyCycle`)

In this order:
1. any `WR` edge from an aborted writer → `G1A_DIRTY_READ`
2. only `WW` → `G0_DIRTY_WRITE`
3. only `WR` → `G1C_CIRCULAR_INFO`
4. `RW` and `WR` (with or without `WW`) → `A5A_READ_SKEW`
5. `WW` and `RW`, no `WR` → `P4_LOST_UPDATE`
6. only `RW`: length > 2 → `G2_ANTI_DEPENDENCY`, else `A5B_WRITE_SKEW`
7. otherwise → `UNKNOWN_INVARIANT_VIOLATION`

`A3_PHANTOM_READ`, `G1B_INTERMEDIATE_READ` and `FRACTURED_READ` are never
produced by `ClassifyCycle` (the register checker emits G1b/fractured reads).

### 5. Choosing one label per run (caller-specific!)

Callers iterate cycles and pick a label, and the rules differ:
- `cmd/chaossql/main.go` uses `dominantAnomaly(cycles, fallback)`: the **first**
  cycle (in cycle order) classified as G1a, G0, G1c, G2, A5B or A5A wins
  immediately; P4 only replaces the fallback while scanning continues; with an
  unknown fallback and cycles present, `cycles[0]` decides. For the minimal
  trace the full-trace label is the fallback, so it survives when the minimal
  trace has no cycles.
- `cmd/chaossql/engine.go`: first non-unknown cycle.
- `pkg/chaostest`, `internal/swarm`, `internal/reporter/html.go`/`ui.go`/
  `sarif.go`, `cmd/chaossql-wasm/bridge.go`: their own loops or `cycles[0]`.

Combined with random cycle order, the same trace can be labeled differently by
different surfaces. Classification is advisory; the invariant decides status.

## Register checker (`CheckRegisterLinearizability`)

Elle-inspired analysis over explicit `RegisterEvent`s (`READ/WRITE/APPEND/
COMMIT/ROLLBACK`): G1a (read from aborted tx), G1b (read of a non-final write
of another tx) and fractured reads on list-append registers. It is a library
used only by its tests; no command feeds it.

## Change checklist

- New anomaly type → `domain.AnomalyType`, `ClassifyCycle`, every priority
  loop listed above, SARIF rule catalog (`internal/reporter/sarif.go`), proxy
  SARIF mapping (`pkg/proxy/sarif.go`), Cloud allow-list
  (`cloud.IsAllowedAnomalyType`), Python `AnomalyType`/`anomaly_code`
  (`sdks/python/chaossql/types.py`), TypeScript `anomalyCode` mapping
  (`sdks/typescript/src/`), portal docs,
  `specs/05_advanced_anomaly_taxonomy.md`.
- Changing extraction also changes the live proxy (it has its own extractor in
  `pkg/proxy/parser.go` — keep semantics aligned).
- Consider reusing `dominantAnomaly` (today private to `cmd/chaossql`) instead
  of adding another priority loop.

## Tests

- `internal/analyzer/adya_test.go` (one test per anomaly + acyclic),
  `internal/analyzer/register_test.go`

## Source map

- `internal/analyzer/adya.go`
- `internal/analyzer/register.go`
- `internal/analyzer/adya_test.go`
- `internal/analyzer/register_test.go`
- `internal/domain/types.go`
- `cmd/chaossql/main.go`
- `internal/analyzer/adya_paths_test.go`
- `specs/05_advanced_anomaly_taxonomy.md`
- `docs/THEORY.md`

## Related skills

- `chaossql-report-exporters`, `chaossql-transparent-proxy`,
  `chaossql-runner-execution`, `chaossql-example-scenarios`
