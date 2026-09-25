---
name: chaossql-database-drivers
description: The DatabaseDriver port and its SQLite (modernc), PostgreSQL (pgx), MySQL (go-sql-driver) and Mock adapters — driver selection, reset semantics, isolation support and defaults, connection pooling, error mapping, WASM build tags, and integration test setup. Use when changing or adding a driver, choosing isolation levels, or when results differ between engines.
---

# Database Drivers (Ports and Adapters)

All database access goes through `drivers.DatabaseDriver`
(`internal/drivers/driver.go`). The engine never imports a concrete driver.

## When to use

- Changing an adapter, adding a new engine, or touching isolation handling.
- A scenario behaves differently per engine, or finds nothing on SQLite.
- Setting up PostgreSQL/MySQL integration tests.

## The port

```go
type DatabaseDriver interface {
    DriverName() string
    Open(ctx) error; Close() error
    Reset(ctx, schemaSQL, seedSQL string) error
    BeginTx(ctx, TransactionOptions{Isolation}) (Tx, error)
    EffectiveIsolation(requested IsolationLevel) (IsolationLevel, error)
    QueryRow / Query / Exec (outside transactions; used by invariants)
}
```

`Tx` exposes `ExecContext`, `QueryContext`, `QueryRowContext`, `Commit`, `Rollback`.

## Selection (`GetDriver(name, dsn)`)

| Build | Names → adapter |
| :--- | :--- |
| native (`!js || !wasm`) | `sqlite`, `sqlite3`, `""` → SQLite; `postgres`, `postgresql` → PostgreSQL; `mysql`, `mariadb` → MySQL; `mock` → Mock |
| `js && wasm` | `sqlite`, `sqlite3`, `""`, `mock` → **Mock**; anything else → error |

## Adapter matrix (verified)

| | SQLite | PostgreSQL | MySQL | Mock |
| :--- | :--- | :--- | :--- | :--- |
| Default isolation | SERIALIZABLE | READ_COMMITTED | REPEATABLE_READ | SERIALIZABLE |
| Supported | SERIALIZABLE, READ_UNCOMMITTED | RC, RR, SERIALIZABLE | all four | all four |
| Pool | 1 connection (raised to 50 permanently once READ_UNCOMMITTED is used) | 50 | 50 | database/sql default |
| Reset | drop every table in `sqlite_master`, run schema, run seed | `DROP SCHEMA public CASCADE; CREATE SCHEMA public;` + schema + seed | drop every BASE TABLE in `DATABASE()` with FK checks off + schema + seed | no-op |
| Error mapping | none | `40P01` → `ErrDeadlockDetected`, `40001` → `ErrSerializationFailure`, connection strings → `ErrConnectionDropped` | `1213` → `ErrDeadlockDetected`, `1205` → `ErrTimeout`, connection strings → `ErrConnectionDropped` | none |

Details:
- **SQLite** (`modernc.org/sqlite`, pure Go — ADR 0004): empty/`:memory:` DSN
  becomes `file:chaos_mem_<nanos>_<seq>?mode=memory&cache=shared` (unique per
  driver instance). `Open` sets WAL (errors ignored) and `busy_timeout = 250`.
  `BeginTx` takes a dedicated `*sql.Conn`, sets `PRAGMA read_uncommitted`
  (0/1) and busy timeout, begins a deferred transaction; commit/rollback reset
  the pragma and release the connection.
- **PostgreSQL** (`pgx/v5/stdlib`): `Open` pings; isolation via `sql.TxOptions`.
- **MySQL**: `Open` rewrites the DSN with `MultiStatements = true` (schema and
  seed files are multi-statement) and pings.
- **Mock**: registers the `chaossql_mock` database/sql driver. `Exec` always
  reports 1 row affected; `Query` returns exactly one row whose columns are
  parsed from the SELECT list (aliases after `AS`), **every value `int64(0)`**.
  Counts opened/committed/rolled-back transactions (`TransactionStats`).
- `MaskDSN` redacts passwords in URL and MySQL DSNs for display.

## Gotchas (verified by running the CLI)

- On SQLite at the default SERIALIZABLE level the single pooled connection is
  held for the whole transaction, so transactions run one at a time and the
  canonical examples **pass** (e.g. `banking_lost_update`). With
  `isolation: READ_UNCOMMITTED` the same scenario produces `violation`
  `P4_LOST_UPDATE` shrunk to 2 operations. Real anomaly hunting needs
  READ_UNCOMMITTED on SQLite, or PostgreSQL/MySQL.
- The Mock driver makes every invariant see zeros, so outcomes depend only on
  whether the assertion holds for zeros (the banking invariant passes); it is
  a plumbing/perf driver, not an isolation model. The WASM playground always
  uses it.
- PostgreSQL `Reset` destroys the whole `public` schema and MySQL `Reset` drops
  every table of the current database — never point a scenario at a database
  you care about.
- SQLite `Reset` drops tables only; views and triggers from previous schemas survive.
- `ErrConnectionDropped` matching includes the substring `"closed"`, which is broad.
- MySQL text-protocol values arrive as `[]byte` (captures render as byte
  lists; invariant columns become strings) — see `chaossql-param-generators`
  and `chaossql-invariant-evaluation`.
- `Tx.QueryRowContext` errors are not mapped (the error surfaces on `Scan`).

## Adding a driver (checklist)

1. New adapter file with `//go:build !js || !wasm`; implement every port
   method, `EffectiveIsolation` via `resolveIsolation`, error mapping to the
   domain errors.
2. Register names in `get_driver_native.go` (and decide the WASM behavior in
   `get_driver_wasm.go`).
3. Keep `CGO_ENABLED=0` (pure-Go driver only) — `make build` and `make wasm`.
4. Update: `validate` supported list (`cmd/chaossql/validate.go`), `init`
   `--driver` help, swarm DSN env resolution (`internal/swarm/diff_runner.go`),
   Cloud allowed drivers (`isAllowedDriver` in `internal/cloud/payload.go`),
   CI services (`.github/workflows/ci.yml`), docs/specs, portal matrix data.
5. Integration test with the skip/require helper below.

## Tests

- Unit: `isolation_test.go`, `mask_test.go`, `driver_mock_test.go`, `sqlite_test.go`.
- Integration: `postgres_test.go` (`DATABASE_URL`), `mysql_test.go`
  (`MYSQL_DSN`, fallback `DATABASE_URL`). They skip when the database is
  unreachable unless `CHAOSSQL_REQUIRE_DATABASES=1` (set in CI), which turns a
  skip into a failure (`integration_test.go`).
- Local services: `docker compose up postgres` (`docker-compose.yml` has no
  MySQL service; use the `mysql:8.4.2` image from `.github/workflows/ci.yml`).
  Note `sqlite3` is accepted by `GetDriver` but rejected by the Cloud
  allow-list (`isAllowedDriver`).

## Source map

- `internal/drivers/driver.go`
- `internal/drivers/isolation.go`
- `internal/drivers/sqlite.go`
- `internal/drivers/postgres.go`
- `internal/drivers/mysql.go`
- `internal/drivers/driver_mock.go`
- `internal/drivers/mask.go`
- `internal/drivers/get_driver_native.go`
- `internal/drivers/get_driver_wasm.go`
- `internal/drivers/integration_test.go`
- `docs/adrs/0004-pure-go-sqlite-vs-cgo.md`
- `specs/06_mysql_savepoints_and_otel.md`

## Related skills

- `chaossql-runner-execution`, `chaossql-differential-fuzzing`,
  `chaossql-wasm-playground`, `chaossql-quality-gate`
