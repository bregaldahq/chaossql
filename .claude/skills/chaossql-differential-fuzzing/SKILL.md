---
name: chaossql-differential-fuzzing
description: Cross-engine comparison flows — chaossql diff (two drivers, one schedule), chaossql matrix (Hermitage-style permitted/prevented table over the examples), and chaossql swarm (scenarios x drivers matrix with divergence rules, DSN env resolution and Markdown summary). Use when changing any of these commands, internal/engine/diff.go, internal/swarm, or interpreting divergence results.
---

# Differential Fuzzing: `diff`, `matrix`, `swarm`

All three run the **same deterministic schedule** on different engines and
compare invariant outcomes. None of them shrink or export artifacts.

## When to use

- Changing `cmd/chaossql/diff.go`, `matrix.go`, `swarm.go`,
  `internal/engine/diff.go`, or `internal/swarm/diff_runner.go`.
- Reading a "semantic divergence" report or a matrix row.

## `chaossql diff <spec.yaml>`

Flags: `--driver-a sqlite`, `--driver-b sqlite`, `--dsn-a :memory:`,
`--dsn-b :memory:`, `--seed` (only when set), `--json`.

`engine.RunDifferentialFuzzing(ctx, spec, A, B, seed)`:
1. `GenerateSchedule` once; `RunSchedule` on A, then on B (sequential).
2. Any run error aborts (`driver A (x) execution error: ...`).
3. Divergent when `ViolationDetected` differs, or both violate with different
   failing invariant names. Output: `domain.DiffResult`.

`database.driver`/`dsn` in the spec are ignored; the flags choose engines.

## `chaossql matrix`

Flags: `--driver sqlite`, `--dsn :memory:`, `--json`, `--markdown`.

Runs a hard-coded list of 9 examples (P4, A3, A5B, A5A, G0, G1c, G1a, G2,
G-DL — **not** the FK cascade scenario) with `runner.Run` on one driver
instance. A row is "PERMITTED (Vulnerable)" iff the run returns no error and
`ViolationDetected`; otherwise "PREVENTED (Safe)" — so execution errors,
inconclusive runs and load failures (silently skipped) all read as safe.
Paths are tried from the CWD, then `../../` (test working directory).

## `chaossql swarm [dir]` (also `swarm diff`, `swarm run`)

Flags (persistent): `--scenarios-dir ./examples`, `--drivers sqlite,mock`,
`--concurrency 4`, `--json`, `--markdown-summary PATH`.

1. Discovery (`discoverAndLoadSpecs`): walk the dir for `chaos.yaml`/`chaos.yml`
   and `variant_*.yaml|yml` (outputs of `mutate`); if none, any YAML. Sorted;
   files that fail `LoadSpec` are skipped silently.
2. `swarm.ExecuteDifferentialMatrix`: pre-generate one schedule per scenario
   from its own seed; fan out (scenario, driver) tasks over `concurrency`
   goroutines.
3. `executeDriverRun` → `resolveDSN`: spec DSN if the spec's driver matches, else env:
   PostgreSQL `DATABASE_URL` (must start with `postgres`) or `POSTGRES_DSN`;
   MySQL `MYSQL_DSN` or `DATABASE_URL` starting with `mysql`. Each run has a
   15 s timeout; errors are recorded per driver instead of failing the swarm.
4. `EvaluateScenarioDivergence` over drivers without errors: pairwise
   divergence on violation flag, differing failing invariant, or (both
   violating) differing detected anomaly labels. 0 valid drivers → "All
   drivers encountered execution or connection errors"; 1 → single-driver summary.
5. Optional GFM summary via `reporter.GenerateSwarmMarkdownSummary`.

## Gotchas

- Every driver runs with the spec's `isolation`; a level unsupported by one
  engine (e.g. READ_UNCOMMITTED on PostgreSQL) becomes an error for that
  driver and removes it from the comparison.
- Schema/seed SQL must be portable across the compared engines.
- The mock driver answers every query with zeros, so comparing against
  `mock` mostly tests the assertion's behavior on zeros.
- On SQLite's default SERIALIZABLE level transactions are serialized (see
  `chaossql-database-drivers`), so `matrix --driver sqlite` reports most
  anomalies as prevented.
- Detected anomaly labels come from possibly random cycle order
  (`chaossql-adya-anomaly-classification`), so label-based divergence can be noisy.
- Exit code is 0 even when divergence is found.

## Change checklist

- New matrix scenario → add to the `scenarios` slice in `matrix.go`,
  the portal matrix data (`site/src/pages/MatrixPage.tsx` / data files) and docs.
- Divergence rule change → `EvaluateScenarioDivergence`,
  `engine.RunDifferentialFuzzing`, swarm Markdown summary, specs 07/15.
- DSN resolution change → `.github/workflows/swarm.yml` env and docs.

## Tests

- `internal/engine/diff_test.go`, `internal/swarm/diff_runner_test.go`,
  `cmd/chaossql/diff_matrix_test.go`, `cmd/chaossql/swarm_test.go`,
  `internal/swarm/diff_runner_internal_test.go`, `internal/reporter/swarm_summary_test.go`

## Source map

- `cmd/chaossql/diff.go`
- `cmd/chaossql/matrix.go`
- `cmd/chaossql/swarm.go`
- `internal/engine/diff.go`
- `internal/swarm/diff_runner.go`
- `internal/reporter/swarm_summary.go`
- `.github/workflows/swarm.yml`
- `specs/07_differential_fuzzing_and_matrix.md`
- `specs/15_multiagent_qa_and_swarm_fuzzing.md`

## Related skills

- `chaossql-database-drivers`, `chaossql-deterministic-schedule`,
  `chaossql-scenario-tooling`, `chaossql-report-exporters`, `chaossql-github-action`
