---
name: chaossql-invariant-evaluation
description: How SQL invariants are evaluated (one query, first row, columns as variables, expr-lang boolean assertion) and how the temporal invariant library works. Use when writing invariants, changing internal/evaluator, or debugging inconclusive results and assertion compile errors.
---

# Invariant Evaluation

Invariants are the oracle of every run (principle #2 of `AGENTS.md`). After all
workers finish, each invariant query runs against the database and its
assertion decides pass/fail.

## When to use

- Writing or reviewing `invariants:` in a scenario.
- A run ends `inconclusive`.
- Changing `internal/evaluator`.

## SQL invariants (`Evaluator.Evaluate`)

1. `driver.Query(ctx, inv.Query)` — outside any transaction.
2. Read the column names; **only the first row** is used; zero rows → error
   `invariant query returned no rows`.
3. Scan every column into `interface{}`; `[]byte` values become `string`.
4. Build `env[column] = value` (also stored in `ActualValues`).
5. `expr.Compile(inv.Assert, expr.Env(env), expr.AsBool())` then `expr.Run`.
6. Result `Passed = output`.

Any error (query, no rows, scan, compile, run, non-bool) is returned **and**
set on `InvariantResult.Error`; the runner then marks the run `inconclusive`
and stops evaluating further invariants.

### Assertion language

[expr-lang/expr](https://expr-lang.org): `==`, `!=`, `<`, `and`, `or`, `not`,
arithmetic, `int()`, `float()`, `len()`, `nil`, etc. Column aliases are the
variable names, so alias every expression column (`SUM(x) AS total`).

## Gotchas (verified)

- Types come from the actual values, so the assertion is type-checked against
  data: a string column compared to a number fails to compile —
  `total == 2000` with `total = "2000"` → `mismatched types string and int` →
  `inconclusive`. Use `int(total) == 2000`. This affects any driver that
  returns text or `[]byte` (MySQL text protocol values).
- `NULL` becomes `nil`; `COALESCE` in SQL is usually clearer than nil checks.
- Mixed `int64`/`float64` arithmetic works (`n + f > 7`).
- `chaossql validate` compiles assertions **without** an environment
  (`expr.Compile(assert, expr.AsBool())`), so it catches syntax errors but not
  unknown columns or type mismatches.
- Invariants run in spec order and the first failure wins; later invariants are
  not evaluated. Reports therefore carry at most one invariant result.
- On the mock driver every column of every query is `int64(0)`
  (see `chaossql-database-drivers`), so assertions like `total == 2000` fail.

## Temporal invariants (`EvaluateTemporalInvariants`)

Pure function over an `ExecutionTrace`:
- `no_aborts` — passes iff no `ROLLBACK` events (`aborted_count`).
- `no_error_events` — passes iff no `ERROR` events and no event with `Error`.
- `monotonicity` — passes iff timestamps never decrease.
- unknown type → failed result with an error.

**No execution surface calls it today**; `temporal_invariants` in a spec are
parsed and ignored. Wiring it in would need a status decision in
`finalizeResult` and reporter support.

## Change checklist

- Changing value conversion changes assertion typing for every scenario —
  re-run all examples on SQLite (`go test ./internal/domain -run LoadSpec`,
  `make demo`) and the driver integration tests.
- Keep the generated repro scripts' assertion evaluators in mind: they
  re-implement a subset of expression evaluation (`chaossql-repro-synthesis`).
- Update `specs/01_invariant_evaluation.md` and `specs/09_*` for semantic changes.

## Tests

- `internal/evaluator/evaluator_test.go`, `internal/evaluator/temporal_test.go`
- `internal/engine/runner_status_test.go` (inconclusive/violation precedence)

## Source map

- `internal/evaluator/evaluator.go`
- `internal/evaluator/temporal.go`
- `internal/evaluator/evaluator_test.go`
- `internal/evaluator/temporal_test.go`
- `cmd/chaossql/validate.go`
- `specs/01_invariant_evaluation.md`
- `specs/09_temporal_invariants_and_g2_cycles.md`

## Related skills

- `chaossql-runner-execution`, `chaossql-spec-format`,
  `chaossql-database-drivers`, `chaossql-scenario-tooling`
