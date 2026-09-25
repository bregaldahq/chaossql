---
name: chaossql-repro-synthesis
description: Generation of standalone reproduction scripts from the minimal schedule — Go (repro_test.go, modernc SQLite), Python (sqlite3 + ThreadPoolExecutor) and TypeScript (node:test + node:sqlite/better-sqlite3) — their embedded data, execution model and fidelity limits. Use when changing internal/reporter/repro*.go or when a generated repro does not reproduce.
---

# Reproduction Script Synthesis

After shrinking, ChaosSQL emits self-contained scripts that embed the schema,
seed, minimal operations and the failing invariant, so a developer can re-run
the counterexample without ChaosSQL (ADR 0005).

## When to use

- Changing `GenerateStandaloneGoRepro`, `GenerateStandalonePythonRepro`, or
  `GenerateStandaloneTypeScriptRepro`.
- A generated repro passes while ChaosSQL reported a violation (or vice versa).

## Where they are produced

| Output | Producer |
| :--- | :--- |
| `repro_test.go` (CWD) / `repro_go` JSON key | `run`/`demo` with `--export-repro` or `--json` |
| `repro_go`, `repro_python`, `repro_typescript` | `chaossql engine` IPC response (always) |
| Python/TS files on disk | SDK `export_standalone_repro` / `exportStandaloneRepro` |

Inputs: `(spec, minimalOps, failingInvariant)`. The invariant query is looked
up by the failing invariant's name; if missing, the first spec invariant is used.

## Execution models

**Go** (`package main`, `TestReproduceAnomaly` + `main`): opens
`file:repro_mem?mode=memory&cache=shared` with modernc SQLite (50 connections),
runs schema and seed, starts **one goroutine per minimal op at once**, each in a
default-isolation transaction with captures and `{…}` substitution, ignores op
errors, then evaluates the invariant with a tiny evaluator: clauses split on
`" and "`, variables replaced textually, comparisons `== != >= <= > <` on
integers; unknown shapes count as true. Fails with `REPRODUCED ANOMALY: ...`.
SQL containing backticks is embedded with `strconv.Quote`.

**Python**: temp-file SQLite DB, `sqlite3.connect(..., isolation_level=None)`
(autocommit — no explicit transaction), all ops submitted to a
`ThreadPoolExecutor`, assertion evaluated with restricted `eval`
(`__builtins__` empty); any evaluation exception counts as a violation.
Runnable with pytest or `python repro.py`.

**TypeScript**: `node:test`; SQLite via `node:sqlite` `DatabaseSync(':memory:')`
(Node ≥ 22.5) or `better-sqlite3`; ops run through `Promise.all` over
synchronous calls (effectively sequential, no transactions); assertion
evaluated with `new Function(...)` as JavaScript, any exception counts as a
violation.

## Fidelity limits (important)

- Always SQLite, regardless of `database.driver`; isolation, jitter, worker
  assignment, latency and abort decisions are **not** reproduced.
- Python/TS run without transactions; TS is effectively sequential.
- Assertions are re-implemented per language: expr-lang syntax like `and`
  is valid in Python but a syntax error in JavaScript, so the TS repro reports
  a violation for any `and` assertion; the Go evaluator only handles simple
  conjunctions of integer comparisons.
- The Go repro's `substituteParams` iterates a map (random order), unlike the
  engine's length-sorted substitution; overlapping param names can differ.
- Consequently a repro may not reproduce a real-database anomaly; the
  authoritative reproduction is `replay --verify` (`chaossql-replay-artifacts`).

## Change checklist

- Keep the three generators consistent (embedded data, invariant lookup).
- If the engine's substitution or capture rules change
  (`chaossql-param-generators`), update the embedded helpers.
- Generated Go must compile in a module that has `modernc.org/sqlite`.
- Tests assert on generated content; update them with intent.

## Tests

- `internal/reporter/reporter_test.go`
- `cmd/chaossql/engine_test.go` (IPC response carries all three)
- SDK tests exercise export paths (`sdks/python/tests`, `sdks/typescript/tests`)

## Source map

- `internal/reporter/repro.go`
- `internal/reporter/repro_python.go`
- `internal/reporter/repro_ts.go`
- `internal/reporter/reporter_test.go`
- `docs/adrs/0005-repro-test-standalone-synthesis.md`
- `specs/04_evidence_synthesis.md`

## Related skills

- `chaossql-ddmin-shrinker`, `chaossql-cli-run-pipeline`, `chaossql-engine-ipc`,
  `chaossql-sdk-python`, `chaossql-sdk-typescript`, `chaossql-replay-artifacts`
