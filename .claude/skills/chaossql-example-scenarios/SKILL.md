---
name: chaossql-example-scenarios
description: The canonical anomaly scenarios in examples/ — what each one targets, its demo alias and matrix code, observed outcomes on SQLite at the default and READ_UNCOMMITTED levels, the README format, and the full checklist of places to update when adding, renaming or changing a scenario. Use when touching examples/ or anything that lists them (demo, matrix, portal, docs, tests).
---

# Canonical Example Scenarios

Each example is a directory with `chaos.yaml`, `schema.sql`, `seed.sql`,
`README.md`. They drive `demo`, `matrix`, `swarm`, portal content, docs and
several tests.

## When to use

- Adding, renaming, or editing an example.
- A demo or matrix result looks wrong.

## Catalog

Observed with `chaossql run --json` at each spec's own seed (outcomes on a
real database are timing dependent; labels can vary between runs):

| Directory | Target | Demo alias | Matrix | SQLite default (SERIALIZABLE) | SQLite `READ_UNCOMMITTED` |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `banking_lost_update` | P4 lost update | `banking` | P4 | passed | violation, P4, 2 ops |
| `inventory_oversell` | A3/oversell | `inventory` | A3 | execution_error (`CHECK stock >= 0`) | violation, labeled P4, 2 ops |
| `hospital_write_skew` | A5B write skew | `hospital` | A5B | passed | violation, A5B, 2 ops |
| `read_skew_financial_audit` | A5A read skew | `financial` | A5A | passed | violation, labeled P4, 2 ops |
| `dirty_write_auction` | G0 dirty write | `auction` | G0 | passed | passed |
| `circular_info_crypto_arbitrage` | G1c circular information | `crypto` | G1c | passed | passed |
| `dirty_read_flash_crash` | G1a dirty read | `flash_crash` | G1a | passed | passed |
| `ticket_booking_anti_dependency` | G2 anti-dependency | `ticket` | G2 | **violation** (3 bookings) | violation, G2, 3 ops |
| `deadlock_cycle` | deadlock diagnostics | `deadlock` | G-DL | passed | passed |
| `foreign_key_cascade_deadlock` | FK cascade deadlock | `fk` | — | passed | passed |

Takeaways:
- On SQLite's default level transactions are serialized
  (`chaossql-database-drivers`); use PostgreSQL/MySQL or READ_UNCOMMITTED to
  observe anomalies.
- `ticket_booking_anti_dependency` violates `total_booked <= 2` even when every
  transaction runs serially, so its invariant also fails on serializable
  histories — a false positive with respect to `evals/02_false_positive_rate.md`.
- `inventory_oversell`'s CHECK constraint turns serialized oversell attempts
  into operation errors (`execution_error`).
- None of the examples sets `database.isolation`.

## README format

`# Scenario NN: <Title> (Anomaly <code>)`, then `## Business Context`, an
anomaly breakdown or mathematical formulation, the consistency invariant, and
`## Formal Mitigation` / `## Remediation`. English only.

## Adding or renaming a scenario (checklist)

1. `examples/<dir>/` with the four files (use `chaossql init` to scaffold).
2. `chaossql validate` it; run it on SQLite RU and, if possible, PostgreSQL.
3. `internal/domain/parser_test.go` example list.
4. Demo alias in `resolveDemoPath` (`cmd/chaossql/main.go`), its usage string,
   `cmd/chaossql/demo_test.go`, and the `demo` target in `Makefile`.
5. Matrix list in `cmd/chaossql/matrix.go` (if it is an anomaly row).
6. Portal: `site/src/data/scenarios.json` (and matrix data/pages).
7. `docs/SCENARIO_ACADEMIC_AUDIT.md`, README scenario table, specs if relevant.
8. `tools/harness_check.go` requires `examples/foreign_key_cascade_deadlock/chaos.yaml`
   by name — keep it or update the list.
9. `examples/*/mutated/` is CI output (`swarm.yml`); never commit it.

## Source map

- `examples/banking_lost_update/chaos.yaml`
- `examples/inventory_oversell/chaos.yaml`
- `examples/hospital_write_skew/chaos.yaml`
- `examples/read_skew_financial_audit/chaos.yaml`
- `examples/dirty_write_auction/chaos.yaml`
- `examples/circular_info_crypto_arbitrage/chaos.yaml`
- `examples/dirty_read_flash_crash/chaos.yaml`
- `examples/ticket_booking_anti_dependency/chaos.yaml`
- `examples/deadlock_cycle/chaos.yaml`
- `examples/foreign_key_cascade_deadlock/chaos.yaml`
- `cmd/chaossql/main.go`
- `cmd/chaossql/matrix.go`
- `cmd/chaossql/demo_test.go`
- `internal/domain/parser_test.go`
- `site/src/data/scenarios.json`
- `docs/SCENARIO_ACADEMIC_AUDIT.md`
- `evals/02_false_positive_rate.md`

## Related skills

- `chaossql-spec-format`, `chaossql-scenario-tooling`,
  `chaossql-differential-fuzzing`, `chaossql-database-drivers`,
  `chaossql-adya-anomaly-classification`
