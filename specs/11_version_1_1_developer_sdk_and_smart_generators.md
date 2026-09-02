# Spec 11: Version 1.1 — Go Developer Testing SDK (`pkg/chaostest`), Smart Generators & GitHub Action

## 1. Go Developer Testing SDK (`pkg/chaostest`)
- Package `github.com/bregaldahq/chaossql/pkg/chaostest` providing idiomatic programmatic API for `*testing.T`:
  - `Tester`: Fluent builder wrapping `engine.Runner` and `drivers.DatabaseDriver`.
  - Methods:
    - `New(t *testing.T) *Tester`
    - `WithDriver(driver drivers.DatabaseDriver) *Tester`
    - `WithSchema(schemaSQL string) *Tester`
    - `WithSeed(seedSQL string) *Tester`
    - `WithInvariant(name, query, assertExpr string) *Tester`
    - `AddOperation(name string, steps ...string) *Tester`
    - `Run(workers, iterations int, seed uint64) *domain.ExecutionResult`
    - `AssertNoAnomalies(workers, iterations int, seed uint64)`: Fails test with `t.Fatalf` if anomaly is detected, printing minimal $ddmin$ reduction and Mermaid sequence.

## 2. Smart Parameter Generator Extensions (`internal/engine/prng.go`)
- Extend declarative PRNG dynamic generators:
  - `$faker_email()`: Deterministic random emails (e.g. `trader_84@defi.org`).
  - `$faker_name()`: Deterministic random full names (e.g. `'Alice Johnson'`).
  - `$faker_phone()`: Deterministic formatted phone numbers.
  - `$monotonic_counter(start, step)`: Incrementing thread-safe counter.

## 3. Official GitHub Action (`action.yml`)
- Root-level composite action `action.yml` for GitHub Marketplace integration:
  - Inputs: `spec-path`, `workers`, `iterations`, `seed`, `export-html`, `export-junit`, `export-summary`.
  - Outputs: `anomaly-detected`, `anomaly-type`, `reduction-ratio`.

## 4. Quality Gate & Harness
- Add `specs/11_version_1_1_developer_sdk_and_smart_generators.md` to `tools/harness_check.go`.
