---
name: chaossql-spec-format
description: The chaos.yaml scenario format and its Go model (domain.Spec). Use when writing or reviewing a scenario, adding or changing a spec field, touching LoadSpec/ParseSpecBytes/Validate, or debugging "chaos spec validation failed" errors.
---

# Scenario Spec Format (`chaos.yaml`)

A scenario is a YAML document decoded into `domain.Spec`. It is the single input
of every execution surface: CLI (`run`, `demo`, `diff`, `matrix`, `swarm`,
`mutate`, `validate`, `bench`), WASM playground, and — after normalization —
the `engine` IPC protocol and `pkg/chaostest`.

## When to use

- Authoring or reviewing a scenario file.
- Adding, renaming, or changing the meaning of a spec field.
- A spec fails to load or validate.

## Schema (verified against `internal/domain/types.go`)

```yaml
version: "1.0"            # required, free-form string (not interpreted)
name: "banking_lost_update"   # required; used in reports and Cloud metadata
description: "..."        # optional prose (excluded from the Cloud fingerprint)
database:
  driver: "sqlite"        # required: sqlite|sqlite3|postgres|postgresql|mysql|mariadb|mock
  dsn: ""                 # optional; empty/":memory:" SQLite => unique shared in-memory DB
  isolation: ""           # optional: READ_UNCOMMITTED|READ_COMMITTED|REPEATABLE_READ|SERIALIZABLE
  schema: "schema.sql"    # DDL inline, or a path ending in .sql (relative to the YAML file)
  seed: "seed.sql"        # DML inline, or a path ending in .sql
engine:
  workers: 4              # <=0 => runner default 4
  iterations: 20          # number of scheduled operations; <=0 => 10
  seed: 42                # PRNG master seed; 0 is a valid literal seed
  jitter_ms: [1, 10]      # per-step delay range; must satisfy 0 <= min <= max
  faults:                 # see chaossql-fault-injection
    abort_probability: 0.0
    latency_probability: 0.0
    latency_spike_ms: [0, 0]
    disconnect_probability: 0.0   # parsed but NOT used by the runner
invariants:               # at least one; names must be unique and non-empty
  - name: "ledger_balance_consistency"
    query: "SELECT ... AS actual_balance, ... AS expected_balance ..."
    assert: "actual_balance == expected_balance and actual_balance >= 0"
temporal_invariants:      # parsed, but NOT evaluated by any runner today
  - name: "no_aborts"
    type: "no_aborts"     # no_aborts | no_error_events | monotonicity
operations:               # at least one; each needs a name
  - name: "withdraw_vulnerable"
    weight: 1.0           # parsed, but IGNORED: selection is uniform
    params:
      amount: "int(5, 25)"          # generators: chaossql-param-generators
    steps:
      - sql: "SELECT balance FROM accounts WHERE id = 1;"
        capture: "current_bal"      # first column of first row -> variable
      - sql: "UPDATE accounts SET balance = {current_bal - amount} WHERE id = 1;"
```

## Loading pipeline

1. `domain.LoadSpec(path)` reads the file and calls `ParseSpecBytes`.
2. `ParseSpecBytes` runs `yaml.Unmarshal` into `Spec`, then `Spec.Validate()`.
3. `LoadSpec` then resolves `database.schema` / `database.seed`: if the trimmed
   value ends in `.sql`, it is read relative to the YAML file's directory and
   replaced by the file content. Otherwise the value is used as inline SQL.
4. `ParseSpecString(yaml, schemaSQL, seedSQL)` parses without disk access and
   overrides schema/seed with non-empty arguments (used by in-memory callers).

`ParseSpecBytes` (no disk access) does **not** resolve `.sql` paths — the WASM
bridge and `mutate` therefore see the literal `"schema.sql"` string.

## `Spec.Validate()` rules

Errors wrap `domain.ErrSpecValidationFailed`:
- `version`, `name`, `database.driver` non-empty;
- `database.isolation` is empty or one of the four levels;
- `invariants` and `operations` non-empty;
- `ValidateInvariantNames`: every invariant has a unique, non-empty name
  (failure signatures identify a violation by invariant name);
- every operation has a name;
- `jitter_ms[0] >= 0` and `jitter_ms[1] >= jitter_ms[0]`.

It does **not** check that steps exist, that SQL parses, that the driver is
supported, or that assertions compile — `chaossql validate` adds those
(see `chaossql-scenario-tooling`). `Runner.Run`/`RunSchedule` re-check only
invariant names and isolation support, so programmatic specs can be partial.

## Gotchas

- `weight` is ignored: `GenerateSchedule` picks templates uniformly.
- `temporal_invariants` never run in `run`/`demo`/`engine`/WASM; only
  `evaluator.EvaluateTemporalInvariants` (library + tests) uses them.
- `faults.disconnect_probability` has no effect.
- Isolation support is driver specific (SQLite: SERIALIZABLE and
  READ_UNCOMMITTED only); an unsupported level fails before the reset.
- The YAML seed (`engine.seed`) is overridden by `--seed` only when the flag is
  explicitly set (so `--seed 0` works).
- `database.seed` (SQL) and `engine.seed` (PRNG) are different things.
- `mutate` re-serializes specs with `yaml.Marshal`, so variant files contain
  every field (including zero values) with the Go YAML tags.

## Change checklist (adding or changing a field)

- `internal/domain/types.go` (YAML tag, validation if needed).
- Every surface that builds a `Spec` by hand: `cmd/chaossql/engine.go`
  (IPC mirror structs), `pkg/chaostest/chaostest.go`, `cmd/chaossql-wasm/bridge.go`.
- `pkg/mutator/mutator.go` `cloneSpec` (deep copy of slices/maps).
- `cmd/chaossql/validate.go` static checks, `cmd/chaossql/init.go` template.
- Cloud fingerprint impact: `internal/cloud/fingerprint.go` hashes the whole
  spec JSON (minus DSN, description, engine seed) — any new field changes
  fingerprints and therefore baseline compatibility.
- Replay artifacts embed the full spec (`chaossql-replay-artifacts`).
- Docs: `specs/`, portal docs data in `site/src/data/docs.json` if user facing.

## Tests

- `internal/domain/types_test.go`, `internal/domain/parser_test.go`
  (loads every example), `internal/domain/parser_inmemory_test.go`
- `go test ./internal/domain/...`

## Source map

- `internal/domain/types.go`
- `internal/domain/parser.go`
- `internal/domain/errors.go`
- `internal/domain/parser_test.go`
- `internal/domain/types_test.go`
- `examples/banking_lost_update/chaos.yaml`

## Related skills

- `chaossql-param-generators`, `chaossql-invariant-evaluation`,
  `chaossql-fault-injection`, `chaossql-scenario-tooling`, `chaossql-example-scenarios`
