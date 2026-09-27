---
name: chaossql-transparent-proxy
description: The chaossql proxy Layer-7 reverse proxy for PostgreSQL and MySQL wire protocols — connection handling, statement inspection, PCT-style jitter and commit barriers, the live shadow serialization graph, anomaly alerts, SARIF export, live UI endpoints, and known limitations. Use when changing pkg/proxy or cmd/chaossql/proxy.go, or when running the proxy against an application.
---

# Transparent Database Proxy (`chaossql proxy`)

Sits between an application and its database, perturbs timing of real
traffic, and detects isolation anomalies from the observed statements — no
application code changes. Unlike the scenario engine it has no invariants and
no shrinking.

## When to use

- Changing `pkg/proxy/*` or `cmd/chaossql/proxy.go`.
- Running the proxy in an integration test environment.

## Command

```
chaossql proxy --listen 127.0.0.1:5433 --upstream 127.0.0.1:5432 --protocol postgres \
  --jitter-min 10µs --jitter-max 2ms --commit-barrier-min 50µs --commit-barrier-max 5ms \
  --pct-depth 2 --seed 42 [--export-sarif FILE] [--ui-port N] [--fail-on-anomaly=true]
```

Runs until SIGINT/SIGTERM, then shuts down (3 s), writes SARIF, prints a
summary, and exits non-zero if anomalies were found and `--fail-on-anomaly`.

## Data path (`pkg/proxy/proxy.go`)

Per client connection: dial upstream, create `ProxySession`, run two
forwarding goroutines (client→server, server→client).
- **PostgreSQL**: handshake loop relays `SSLRequest` + the 1-byte answer, then
  forwards the StartupMessage. Client `Q` (simple query) and `P` (parse)
  messages are inspected; `X` terminates. Server `Z` (ReadyForQuery) status
  `I`/`T` updates the session's in-transaction flag.
- **MySQL**: `COM_QUERY` and `COM_STMT_PREPARE` are inspected; OK/ERR packets
  are parsed for status and deadlock codes.

`InspectSQL` (`parser.go`) classifies: boundaries `BEGIN|BEGIN TRANSACTION|
START TRANSACTION`, `COMMIT|END`, `ROLLBACK|ABORT`, `SET [SESSION
CHARACTERISTICS AS] TRANSACTION ISOLATION LEVEL ...`; otherwise read/write with
one item `table` or `table:id` (from `WHERE id = N` / INSERT id regexes).

Per statement:
- BEGIN → `ShadowGraphEngine.StartTx`.
- COMMIT → sleep a commit barrier, forward, `RecordCommit` (promote writes,
  run cycle detection, emit alerts).
- ROLLBACK → forward, `RecordRollback` (G1a if someone read this tx's writes).
- DML/SELECT inside a transaction → sleep `CalculateJitter`, then
  `RecordWrite`/`RecordRead`, then forward.

## Scheduler (`jitter.go`)

`PCTScheduler` draws from one `math/rand` source seeded with `--seed`
(0 → 42) behind a mutex. Jitter is uniform in `[min, max]`; for DML on
sessions with `id % depth == 0` (depth > 1) it is pushed toward `max`
(priority inversion). Commit barriers are uniform in their range.

## Shadow graph (`shadow_graph.go`)

Tracks active tx read/write sets, per-item readers, committed last writers and
uncommitted writers; adds WR/WW/RW edges into an `analyzer.AdyaGraph`,
runs `analyzer.FindCycles` + `ClassifyCycle` on every commit, de-duplicates by
cycle fingerprint, logs alerts to stderr, and records a trace
(`GetTrace`, `GetAnomalies`, `GetGraph`).

Live UI (`--ui-port`): `/` dashboard, `/api/status`, `/api/anomalies`,
`/api/sarif`, `/api/trace`.

## Limitations and gotchas (verified in code)

- Timing is not reproducible: the shared RNG is consumed in real arrival order.
- TLS is not supported in practice (after an `S` answer to `SSLRequest` the
  stream is encrypted and cannot be parsed) — use `sslmode=disable`.
- Only statement **text** is inspected: re-executions of prepared statements
  (PG Bind/Execute, MySQL `COM_STMT_EXECUTE`) are invisible.
- Statements outside `BEGIN` share one `S<id>-autocommit` node per session.
- The graph is never pruned; cycle detection cost grows with session length.
- In `RecordWrite`, `uncommittedWriters[item]` is overwritten with the current
  tx **before** the uncommitted-writer check, so that WW edge is never added
  (only committed last writers produce WW edges).
- Item extraction is separate from `internal/analyzer` (`extractItemFromSQL`);
  keep them aligned.

## Tests

- `pkg/proxy/*_test.go`, `cmd/chaossql/proxy_test.go`
- Manual end-to-end: `tools/run_manual_proxy_test.sh` (needs a real database).

## Source map

- `cmd/chaossql/proxy.go`
- `pkg/proxy/proxy.go`
- `pkg/proxy/types.go`
- `pkg/proxy/parser.go`
- `pkg/proxy/postgres.go`
- `pkg/proxy/mysql.go`
- `pkg/proxy/jitter.go`
- `pkg/proxy/shadow_graph.go`
- `pkg/proxy/sarif.go`
- `pkg/proxy/ui.go`
- `tools/run_manual_proxy_test.sh`
- `specs/16_transparent_database_proxy.md`

## Related skills

- `chaossql-adya-anomaly-classification`, `chaossql-report-exporters`
