---
name: chaossql-param-generators
description: Operation parameter generators ($random_int, $random_choice, $random_string, $uuid, $faker_*, $monotonic_counter, legacy int()), step captures, and {placeholder}/{a - b} SQL substitution. Use when writing operation params, adding a generator, or debugging SQL that contains unresolved or mangled placeholders.
---

# Parameter Generators, Captures and Substitution

Operations are templates. At schedule time each `params` entry is evaluated
into a concrete string; at execution time `{name}` placeholders in each step's
SQL are replaced from the operation's params plus values captured by earlier
steps of the same transaction.

## When to use

- Declaring `params` in a scenario.
- Adding or changing a generator.
- SQL reaching the database with literal `$random_...`, `{...}` or byte-slice text.

## Generators (`internal/engine/prng.go`)

| Expression | Result |
| :--- | :--- |
| `$random_int(min, max)` / legacy `int(min, max)` | uniform integer in `[min, max]`; `min == max` returns `min` |
| `$random_choice('a', "b", c)` | one element; quotes optional, `\` escapes, commas inside quotes kept |
| `$random_string(n)` | `n` chars from `[a-zA-Z0-9]` |
| `$uuid()` | RFC 4122 v4 UUID built from the seeded RNG (deterministic) |
| `$faker_email()` | `<prefix>_<0..999>@<domain>` from fixed lists |
| `$faker_name()` | `First Last` from fixed lists |
| `$faker_phone()` | `+1-555-NNNN` |
| `$monotonic_counter(start[, step])` | `start`, `start+step`, ... per `(start, step)` key |
| anything not starting with `$` | returned verbatim (static value) |

### Evaluation path

`GenerateSchedule` creates a fresh `PRNG` from the master seed and one
`masterRng = rand.New(rand.NewPCG(seed, 0))`. For every scheduled operation it
first draws the template index, then evaluates params **in sorted name order**
with `PRNG.EvaluateParam(expr, masterRng)`:

1. `$monotonic_counter(...)` → `PRNG.evalMonotonicCounter`, a counter map owned
   by that PRNG instance (so counters restart for every `GenerateSchedule`
   call, i.e. per run).
2. otherwise → `EvaluateGenerator(expr, masterRng)`.
3. **On any error the raw expression is returned unchanged** and ends up in SQL.

`EvaluateGenerator` called directly handles `$monotonic_counter` with a
process-global `sync.Map` (not reset per run; `ResetMonotonicCounters` exists
for tests). Only `bench` calls it directly.

## Captures

A step with `capture: var` is executed with `tx.QueryRowContext(...).Scan(&v)`
where `v` is `interface{}`; the first column of the first row is stored as
`fmt.Sprintf("%v", v)` in the operation's local state. A scan error (including
no rows) is a step failure → rollback → `execution_error`.

The IPC protocol and `pkg/chaostest` also accept the inline form
`"SELECT ...; -> var"` or `"=> var"`.

## Substitution (`SubstituteParams`, `internal/engine/runner.go`)

1. Exact pass: every `{key}` for keys in state is replaced, keys ordered by
   length desc then lexicographically (so `{aa}` never collides with `{a}`).
2. Expression pass: while the SQL still contains `{` and a later `}`, take the
   first `{...}`, replace every state key occurring as a substring of the
   inner text with its value, then `evalSimpleArithmetic`:
   exactly one binary `-` (checked first) or `+` between two base-10 integers.
   The braces are removed and the evaluated (or unchanged) inner text inserted.

## Gotchas

- Invalid generator arguments are silent: `$random_int(5)` reaches SQL as-is.
- Changing, adding, or renaming a param changes the RNG draw sequence of every
  later operation → a different schedule for the same seed (the golden corpus
  test will catch engine-level changes; scenario edits simply change results).
- Any `{...}` in SQL is consumed by the expression pass, including JSON
  literals such as `'{"a":1}'` — the braces are stripped.
- Arithmetic supports only one `+` or `-` on integers: negative operands
  (`-5 - 3`), floats (`12.5`) and multiplication are left as literal text.
- Captured `[]byte` values are formatted with `%v` (e.g. `[49 48 48 48]`).
  go-sql-driver/mysql returns text-protocol values as `[]byte`, so captures on
  MySQL are unreliable; SQLite and PostgreSQL integers arrive as `int64`.
  Only SQLite captures are covered by tests.
- Captured values are substituted as raw text — quote strings in the SQL
  template yourself (`'{name}'`).

## Change checklist (new generator)

- Add the prefix branch in `EvaluateGenerator` (and in `PRNG.EvaluateParam` if
  it needs per-run state like the monotonic counter).
- Draw only from the passed `*rand.Rand` — never `time`, global rand, or maps
  iterated in random order.
- Add tests in `internal/engine/prng_test.go`; if the golden corpus changes,
  treat it as a schedule format change (see `chaossql-deterministic-schedule`).
- Mirror the substitution helpers if you change them: the generated repro
  scripts re-implement `substituteParams`/`evalSimpleArithmetic`
  (`chaossql-repro-synthesis`).
- Document the generator in the portal docs data and `specs/11_*`.

## Tests

- `internal/engine/prng_test.go` (every generator, stable parameter order,
  per-run monotonic counters)
- `internal/engine/runner_internal_test.go` (`TestSubstituteParams*`)
- `go test ./internal/engine -run 'Generator|PRNG|Substitute|GenerateSchedule'`

## Source map

- `internal/engine/prng.go`
- `internal/engine/runner.go`
- `internal/engine/prng_test.go`
- `internal/engine/runner_internal_test.go`
- `specs/11_version_1_1_developer_sdk_and_smart_generators.md`

## Related skills

- `chaossql-deterministic-schedule`, `chaossql-runner-execution`,
  `chaossql-repro-synthesis`, `chaossql-spec-format`
