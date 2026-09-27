---
name: chaossql-scenario-tooling
description: Authoring-time commands — chaossql init (scaffold), chaossql validate (static lint with ERROR/WARNING levels), and chaossql mutate with the pkg/mutator adversarial operators (jitter, lock inversion, step shuffle, savepoint injection). Use when changing any of these commands or the mutator, or when generating scenario variants.
---

# Scenario Tooling: `init`, `validate`, `mutate`

## When to use

- Changing scaffolding templates, lint rules, or mutation operators.
- Producing variants for `swarm`.

## `chaossql init <dir>`

Flags: `--driver sqlite`, `--name` (default: dir base name), `--force`.
Creates the dir and writes `schema.sql` (accounts table), `seed.sql` (Alice and
Bob with 1000 each), `chaos.yaml` (4 workers, 20 iterations, seed 42, jitter
`[0, 5]`, invariant `total_balance == 2000`, two transfer operations) and a
`README.md` skeleton. Refuses if `chaos.yaml` exists unless `--force`.

The template's transfers use `balance = balance ± 100` (atomic updates), so it
passes by design — it is a starting point, not an anomaly.

## `chaossql validate <chaos.yaml>`

`LoadSpec` first (so every `Spec.Validate` rule applies and aborts early),
then `validateScenarioSpec` adds:
- WARNING: driver not in `sqlite|postgres|mysql` (aliases and `mock` warn too);
- ERROR: empty schema SQL; WARNING: empty seed SQL;
- ERROR: invariant without name/query/assert;
- ERROR: assert that fails `expr.Compile(assert, expr.AsBool())` (no env, so
  unknown variables are **not** detected — see `chaossql-invariant-evaluation`);
- ERROR: operation without name or without steps.
Any ERROR → command fails with `scenario validation failed with N error(s)`
(N counts all issues, including warnings).

## `chaossql mutate <scenario.yaml>`

Flags: `--variants 5`, `--output-dir ./mutated`, `--seed 42`, `--json`.
Uses `ParseSpecBytes` (no `.sql` resolution), copies referenced `schema.sql`/
`seed.sql` into the output dir, and writes `variant_<i>.yaml` via `yaml.Marshal`.

`mutator.MutateScenario(spec, opts)`: per variant `i`,
`variantSeed = seed + i*7919 + 1`, `rng = math/rand.New(NewSource(variantSeed))`,
deep `cloneSpec`, `Name = <name>_variant_<i>`, `Engine.Seed = variantSeed`,
then operators in this order:
1. **Jitter**: `jitter_ms = [min, min+2+rand(25)]` with `min = rand(10)`.
2. **Lock inversion**: in ops with >=2 write/`FOR UPDATE` steps matching
   `id|account_id|order_id|item_id|doctor_id|user_id|book_id = N`, take the
   first such step and the first later one on a different id and swap them
   with p=0.75 — unless the later step references a capture made by any step
   from the first one up to it.
3. **Step shuffle**: randomized topological sort that preserves capture
   dependencies, savepoint statements, and same-table orderings when either
   step writes.
4. **Savepoints**: per op (skipped with p=0.2 when >1 op), wrap each write (or
   a single-step op) in `SAVEPOINT spN` … `RELEASE SAVEPOINT spN`, adding
   `ROLLBACK TO SAVEPOINT spN` with p=0.25; otherwise wrap one random step.

Deterministic for a given input and `--seed` (math/rand with fixed seeds).

## Gotchas

- Mutations can make scenarios fail for reasons unrelated to isolation (e.g.
  lock inversion creates deadlocks → `execution_error`; savepoints are not
  supported identically across engines).
- Lock inversion only checks whether the *moved-earlier* step uses a capture;
  steps between the two that use the moved-later step's capture can break.
- Variant YAML contains every field with zero values and keeps the original
  `schema:`/`seed:` file references (copied next to the variants).
- `swarm.yml` writes variants to `examples/*/mutated/` in CI; do not commit them.

## Tests

- `cmd/chaossql/init_test.go`, `cmd/chaossql/validate_test.go`,
  `cmd/chaossql/mutate_test.go`, `pkg/mutator/mutator_test.go`

## Source map

- `cmd/chaossql/init.go`
- `cmd/chaossql/validate.go`
- `cmd/chaossql/mutate.go`
- `pkg/mutator/mutator.go`
- `pkg/mutator/mutator_test.go`
- `specs/10_developer_tooling_and_static_validator.md`

## Related skills

- `chaossql-spec-format`, `chaossql-differential-fuzzing`,
  `chaossql-invariant-evaluation`, `chaossql-example-scenarios`
