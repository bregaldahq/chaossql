# Spec 10: Developer Tooling, Static Scenario Validator & Deadlock Diagnostics

## 1. Scenario Scaffolding (`chaossql init <path>`)
- Creates a new scenario boilerplate with:
  - `chaos.yaml`: Declarative chaos configuration with example invariants and operations.
  - `schema.sql`: Initial DDL table schema.
  - `seed.sql`: Initial seed data records.
  - `README.md`: Formal business and isolation anomaly documentation template.

## 2. Static Scenario Linter & Validator (`chaossql validate <chaos.yaml>`)
- Performs comprehensive pre-flight verification:
  - Checks YAML syntax and required fields (`version`, `name`, `database.driver`, `invariants`, `operations`).
  - Verifies accessibility of `schema` and `seed` SQL files.
  - Statically compiles all `assert` invariant expressions with `expr.Compile` to detect syntax or type errors without executing queries.
  - Validates operation weights ($> 0$) and non-empty steps.

## 3. Scenario 9: Deadlock Cycle & Timeout Diagnostics (`examples/deadlock_cycle/`)
- Demonstrates concurrent resource locking in inverse order:
  - $T_1$: Locks Account 1, pauses, attempts to lock Account 2.
  - $T_2$: Locks Account 2, pauses, attempts to lock Account 1.
- Invariant: `deadlock_handled_gracefully` or `ledger_preservation`.
