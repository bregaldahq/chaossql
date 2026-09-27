---
name: chaossql-engine-ipc
description: The chaossql engine JSON-over-stdio protocol used by the Python and TypeScript SDKs — request normalization (flat or nested fields, defaults, inline capture syntax), execution with shrink and classification, the response shape with three repro scripts, streaming of multiple payloads, and error behavior. Use when changing cmd/chaossql/engine.go or anything the SDKs send or parse.
---

# Engine IPC Protocol (`chaossql engine`)

Host-language SDKs spawn the static `chaossql` binary with the `engine`
subcommand, write one JSON scenario to stdin, and read one JSON result from
stdout (ARCHITECTURE §4). No CGO, no shared memory, no Python GIL issues.

## When to use

- Changing `IPCPayload`/`IPCResponse` or `executeIPCPayload`.
- Changing what the SDKs send or parse.

## Transport

`runEngineIPC` decodes a stream of JSON values from stdin; for each payload it
writes one encoded `IPCResponse` line to stdout. It stops at EOF or when no
further value is buffered (`dec.More()`). A decode error writes an
`execution_error` response with `JSON decoding error: ...` and exits non-zero.
SIGINT/SIGTERM cancel the context.

## Request (`IPCPayload`)

Fields may be flat or nested; nested non-empty values win:

| Concern | Flat | Nested | Default |
| :--- | :--- | :--- | :--- |
| driver | `driver` | `database.driver` | `sqlite` |
| DSN | `dsn` | `database.dsn` | SQLite empty/`:memory:` → `file:chaossql_ipc_<nanos>?mode=memory&cache=shared` |
| isolation | `isolation` | `database.isolation` | driver default |
| schema / seed SQL | `schema`, `seed` | `database.schema`, `database.seed` | "" |
| workers | `workers` | `engine.workers` | **2** |
| iterations | `iterations` | `engine.iterations` | 10 |
| PRNG seed | `seed_value` | `engine.seed` | 0 |
| jitter | — | `engine.jitter_ms` (used when `[1] > 0`) | `[1, 5]` |
| name / version | `name`, `version` | — | `chaossql_ipc_scenario`, `1.1` |

`invariants[] {name, query, assert}` and `operations[] {name, weight, params,
steps[] {sql, capture}}`. A step without `capture` may use the inline form
`"SELECT ... -> var"` or `"... => var"` (split on the first occurrence).
Weight `<= 0` becomes 1.0 (and is ignored by scheduling anyway).
Faults are not expressible over IPC. `Spec.Validate()` is **not** called; the
runner still rejects duplicate/empty invariant names and unsupported isolation.

## Execution

Mirrors the CLI pipeline without exports: `GetDriver` → `Open` →
`runner.Run` → classify (first non-unknown cycle) → shrink + verify on
violation → reclassify → Mermaid + Go/Python/TypeScript repros from the
minimal ops. Cancellation during shrink returns a `canceled` response with
the original run metadata.

## Response (`IPCResponse`)

`status`, `isolation`, `seed`, `schedule`, `operation_errors`, `success`,
`violation_detected`, `anomaly_type`, `failing_invariant`, `duration_ms`,
`trace_events_count`, `shrink`, `minimal_operations`, `mermaid`, `repro_go`,
`repro_python`, `repro_typescript`, `error`. Driver/open failures return
`status: execution_error` with only `error` set. For a successful run,
`error` carries `runResult.Error` (e.g. joined operation errors).

## Gotchas

- The inline capture split also triggers on SQL that legitimately contains
  `->` or `=>` (PostgreSQL JSON operators) when `capture` is empty — pass
  `capture` explicitly for such steps.
- Defaults differ from the YAML runner (workers 2 vs 4, jitter `[1,5]`).
- JSON numbers above 2^53 lose precision in JavaScript; the TS SDK rejects
  unsafe seeds before spawning.
- The trace itself is not returned (only its length).

## Change checklist

- Field added/renamed → `IPCPayload`/`IPCResponse`, Python
  `engine.execute_ipc` + `types.ChaosResult`, TypeScript `engine.ts` +
  `types.ts`, SDK READMEs, `specs/17_multi_language_sdks.md`.
- Keep responses backwards compatible (SDKs read keys with defaults).

## Tests

- `cmd/chaossql/engine_test.go`; SDK suites via `make test-sdks`.

## Source map

- `cmd/chaossql/engine.go`
- `cmd/chaossql/engine_test.go`
- `sdks/python/chaossql/engine.py`
- `sdks/typescript/src/engine.ts`
- `specs/17_multi_language_sdks.md`
- `ARCHITECTURE.md`

## Related skills

- `chaossql-sdk-python`, `chaossql-sdk-typescript`, `chaossql-cli-run-pipeline`,
  `chaossql-repro-synthesis`
