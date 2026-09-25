---
name: chaossql-benchmarks
description: The chaossql bench command — PRNG/generator throughput, Adya graph build and cycle detection latency, ddmin throughput, and a database transfer-throughput stress, with JSON and terminal output. Use when changing bench, reading its numbers, or benchmarking a scenario.
---

# Benchmarks (`chaossql bench [scenario.yaml]`)

## When to use

- Changing `cmd/chaossql/bench.go` or `make bench`.
- Interpreting benchmark output or comparing performance across changes.

## Flags

`-d/--duration 2s` (must be > 0), `-w/--workers 4` (must be > 0), `--json`.
A scenario path is resolved from the CWD, then from the repo root.

## Suite (`runBenchmarkSuite`)

1. **PRNG & generators** for `duration/3`: evaluates generator expressions
   with `EvaluateGenerator`, plus `WorkerSeed` and `Jitter` → ops/sec.
2. **Adya graph** for 100, 500, 1000 nodes: synthetic trace → `BuildGraph`,
   `FindCycles`, `ClassifyCycle` → ms and cycles found.
3. **ddmin** for `duration/3`: synthetic oracle → iterations/sec.
4. **Database concurrency** for the full `duration`: `Reset` with the built-in
   SQLite `accounts` schema (4 accounts × 100000) or the given scenario's
   schema/seed, then `workers` goroutines loop:
   `UPDATE accounts SET balance = balance - a WHERE id = x`,
   `UPDATE ... + a WHERE id = y`, `SELECT balance ...` — autocommit statements,
   not transactions. TPS counts loops where both updates succeeded; status
   `OPTIMAL` unless TPS < 50 (`PASS`).

Output: `BenchmarkSuiteResult{config, environment, prng, adya_graphs,
delta_debugging, database, rows, summary}` as JSON, or a Lipgloss card.

## Gotchas

- With a scenario argument only its driver/DSN/schema/seed are used; the
  workload is always the hard-coded `accounts` transfer, so the scenario schema
  must contain `accounts(id, balance)` with ids 1–4 or TPS is 0.
- The database workload uses wall-clock-seeded RNGs — it is intentionally not
  deterministic and must not be used as a correctness oracle.
- `summary.all_optimal` is always `true`.
- `bench` is the only caller of the process-global `$monotonic_counter` path.

## Tests

- `cmd/chaossql/bench_test.go`

## Source map

- `cmd/chaossql/bench.go`
- `cmd/chaossql/bench_test.go`
- `Makefile`

## Related skills

- `chaossql-param-generators`, `chaossql-adya-anomaly-classification`,
  `chaossql-ddmin-shrinker`, `chaossql-wasm-playground` (browser bench)
