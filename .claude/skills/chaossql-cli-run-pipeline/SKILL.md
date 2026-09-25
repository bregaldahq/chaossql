---
name: chaossql-cli-run-pipeline
description: The chaossql run and demo commands end to end — flag overrides, driver setup, run, anomaly classification, shrink and verify, every export flag, JSON output shape, Cloud publishing hook, trace viewer, and exit codes. Use when adding or changing a run/demo flag or export, changing the JSON output, or debugging CLI behavior and exit status.
---

# CLI `run` / `demo` Pipeline

`chaossql run <chaos.yaml>` and `chaossql demo [scenario]` share
`executeChaos` in `cmd/chaossql/main.go`. It is the reference composition of
the whole engine; other surfaces (IPC, SDK, WASM) re-implement parts of it.

## When to use

- Adding/changing a flag or an export on `run`/`demo`.
- Changing the `--json` payload (SDKs and the pytest plugin parse it).
- Unexpected exit codes or missing artifacts.

## Flags (`newRunCmd`; `demo` has all except `--seed/--workers/--iterations`)

| Flag | Default / env | Effect |
| :--- | :--- | :--- |
| `--seed` | 0 | overrides `engine.seed` only when explicitly set (`--seed 0` works) |
| `--workers`, `--iterations` | 0 | override when `> 0` |
| `--json` | false | print one JSON document to stdout; suppress terminal report |
| `--export-repro` | false | write `repro_test.go` **in the current directory** |
| `--export-mermaid` | false | write `trace.mermaid` in the current directory |
| `--export-html PATH` | "" | standalone HTML report |
| `--export-otel PATH` | "" | OTLP JSON trace |
| `--export-junit PATH` | "" | JUnit XML |
| `--export-summary PATH` | "" | GitHub Step Summary markdown |
| `--export-sarif PATH` | "" | SARIF 2.1.0 |
| `--export-result PATH` | "" | replay artifact v1 (`chaossql-replay-artifacts`) |
| `--ui` | false | after the run, serve the trace viewer on `127.0.0.1:8090` (blocks) |
| `--cloud-token` | `$CHAOSSQL_CLOUD_TOKEN` | enables Cloud publishing |
| `--cloud-url` | `$CHAOSSQL_CLOUD_URL` or `https://api.chaossql.bregalda.com` | Cloud base URL |
| `--cloud-fail-fast` | false | Cloud errors become command errors |
| `--github-token` | `$GITHUB_TOKEN` | PR comment token |
| `--pr-comment` | true | post PR comment when in a PR CI context |

Flag variables are **package globals shared by `run` and `demo`**; tests that
execute commands must reset them.

## `demo`

`resolveDemoPath` maps aliases (`banking`, `inventory`, `hospital`,
`financial`, `auction`, `crypto`, `flash_crash`, `ticket`, `deadlock`, `fk` and
synonyms) to `examples/<dir>/chaos.yaml`; default `banking`. The relative path
is tried from the CWD, then from the repository root found by walking up to
`go.mod` (`findRepoRoot`). `demo` never overrides the seed.

## Pipeline (`executeChaos`)

1. `signal.NotifyContext` (SIGINT/SIGTERM) → `ctx`.
2. `domain.LoadSpec`; apply seed/workers/iterations overrides.
3. `drivers.GetDriver(driver, dsn)`, `Open`, deferred `Close`.
4. `runner.Run(ctx, spec)`; `preserveRunResult` keeps a non-nil result even
   when an error accompanied it (cancellation), else returns the error.
5. Classify the **full** trace (`chaossql-adya-anomaly-classification`).
   Classification runs for every status — a `passed` run can still carry an
   `anomaly_type` label.
6. If `ViolationDetected`: build the failure signature, `shrinker.Shrink`, re-run
   the minimal ops; only if it reproduces, adopt `minimalOps`/`minimalTrace`
   and reclassify on the minimal trace. Cancellation aborts the command.
7. Exports in this order, all using the minimal ops/trace when available:
   result artifact → Go repro (also when `--json`) → Mermaid (also when
   `--json`) → HTML (also when `--json`) → OTLP (also when `--json`) → JUnit →
   Step Summary → SARIF.
8. `--json`: encode the output map (below), optionally publish to Cloud and
   attach `cloud`/`cloud_error`, then return.
9. Otherwise: `reporter.PrintTerminalReport`, Cloud publish, optional `--ui`.
10. Return `unreliableRunError(result)`.

### JSON output keys

`spec{name,driver,workers,iterations,seed}`, `status`, `isolation`, `seed`,
`schedule`, `operation_errors`, `success`, `violation_detected`,
`anomaly_type`, `failing_invariant`, `duration_ms`, `trace_events_count`,
`shrink` (with `minimal_ops`), `mermaid`, `repro_go`, `html_report`,
`otel_trace`, optional `error`, `cloud`, `cloud_error`.
Consumers: Python `ChaosSQLRunner.run_scenario` (pytest plugin) — keep keys
stable or update `sdks/python/chaossql/pytest_plugin.py`.

## Exit codes (verified)

`unreliableRunError` returns nil for `passed` **and `violation`**; only
`execution_error`, `inconclusive`, `canceled` (and setup failures) exit 1.
So `chaossql run` exits **0 when it finds an anomaly**, and the composite
GitHub Action does not fail the step on violations. With `--cloud-fail-fast`,
Cloud errors are joined into the returned error.

## Gotchas

- `--export-repro`/`--export-mermaid` ignore any path; they always write to CWD.
- `--json` embeds full HTML/OTLP/repro strings — outputs can be large.
- Exports are written with mode `0644` except the replay artifact (`0600`, atomic).
- `--ui` blocks until interrupted; do not use in CI.
- `matrix`, `swarm`, `diff` do **not** use this pipeline (no shrink, no exports).

## Change checklist (new flag or export)

- Register the flag on both `newRunCmd` and `newDemoCmd` when it applies to both.
- Add the export step in `executeChaos` using minimal ops/trace.
- Mirror in `action.yml` inputs (`chaossql-github-action`) if CI users need it.
- Update JSON consumers (pytest plugin), README CLI docs, portal docs data.
- Tests in `cmd/chaossql/*_test.go` (e.g. `sarif_cli_test.go`, `demo_test.go`,
  `cloud_integration_test.go`, `action_integration_test.go`).

## Source map

- `cmd/chaossql/main.go`
- `cmd/chaossql/root.go`
- `cmd/chaossql/root_finder.go`
- `cmd/chaossql/demo_test.go`
- `cmd/chaossql/sarif_cli_test.go`
- `cmd/chaossql/cloud_integration_test.go`
- `README.md`

## Related skills

- `chaossql-report-exporters`, `chaossql-repro-synthesis`,
  `chaossql-replay-artifacts`, `chaossql-cloud-publishing`,
  `chaossql-github-action`, `chaossql-ddmin-shrinker`
