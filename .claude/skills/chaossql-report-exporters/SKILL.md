---
name: chaossql-report-exporters
description: Every report format produced from a run — terminal (Lipgloss), Mermaid sequence diagram, standalone HTML report, embedded trace viewer served by chaossql ui / --ui, OpenTelemetry OTLP JSON, JUnit XML, GitHub Step Summary, SARIF 2.1.0 rules and results, and the swarm Markdown summary. Use when adding or changing an export, a SARIF rule, or the trace viewer.
---

# Report Exporters (`internal/reporter`)

Reporters are pure presentation: they receive domain artifacts (spec,
result, trace, graph, shrink result, invariant results) and render them. They
must not execute anything (`AGENTS.md` layer rule).

## When to use

- Adding an output format or changing an existing one.
- Changing the SARIF rule catalog or anomaly → rule mapping.
- Working on the local trace viewer.

## Formats

| Format | Function | Trigger |
| :--- | :--- | :--- |
| Terminal | `PrintTerminalReport` / `RenderFullReport` (banner, run summary, invariant table, shrink summary) | default `run`/`demo` |
| Mermaid | `GenerateMermaidSequence(trace)` | `--export-mermaid` (`./trace.mermaid`), `--json`, IPC |
| HTML | `GenerateStandaloneHTMLReport(trace, spec, graph, shrink, invs)` | `--export-html PATH`, `--json` |
| Trace viewer | `GenerateEmbeddedTraceViewerHTML(...)` + assets in `ui_assets.go` | `chaossql ui <file>`, `run --ui` |
| OTLP JSON | `GenerateOTLPTraceJSON(trace, spec)` | `--export-otel PATH`, `--json` |
| JUnit XML | `GenerateJUnitXML(spec, result, anomaly)` | `--export-junit PATH` |
| Step Summary | `GenerateGitHubSummaryMarkdown(spec, result, shrink, anomaly)` | `--export-summary PATH` |
| SARIF 2.1.0 | `GenerateSARIFReport(spec, invs, graph, shrink)` | `--export-sarif PATH` |
| Swarm summary | `GenerateSwarmMarkdownSummary(report)` | `swarm --markdown-summary` |

`run` passes the **minimal** trace/ops when shrinking succeeded, and the graph
built from the full trace to HTML/SARIF/UI.

## Format notes

- **Mermaid**: one participant per worker plus `DB`; BEGIN/EXEC/SAVEPOINT/
  COMMIT/ROLLBACK as arrows, errors as `--x` plus a note; appends a note per
  detected cycle with its classification. SQL whitespace is collapsed.
- **OTLP**: a root span named after the scenario, one child span per
  operation (events grouped by `op_<id>`), one grandchild span per event
  (latency = gap to the next event, minimum 100 µs); error events mark the
  transaction span as failed; base time fixed at 2026-01-01 UTC plus trace
  offsets; trace and
  span IDs come from `crypto/rand` — output is **not** byte-stable across runs.
- **JUnit**: one suite `chaossql.<name>`, one test case
  `isolation_invariants_<name>`, class `chaossql.drivers.<driver>`; violation →
  `<failure>`; other non-passed statuses → `<error>`. Identifiers that fail
  `safeSummaryIdentifier` are replaced by fallbacks.
- **SARIF**: `StandardRulesCatalog` has 11 rules (`chaossql/P4-lost-update`,
  `A5B-write-skew`, `A5A-read-skew`, `G0-dirty-write`, `G1a-dirty-read`,
  `G1b-intermediate-read`, `G1c-circular-info`, `G2-anti-dependency`,
  `G-DL-deadlock`, `A3-phantom-read`, `unknown-invariant-violation`). At most
  one result: a deadlock result when the scenario name, an invariant name, or
  an invariant error contains "deadlock"; else an anomaly result when an
  invariant failed **or any cycle exists**. The location URI is guessed as
  `examples/<spec.name>/chaos.yaml` (or the name if it looks like a path).
- **Trace viewer**: `chaossql ui` accepts a result/replay JSON with `trace`,
  a bare trace array, or an `ExecutionResult`; serves on `127.0.0.1:<port>`
  (default 8090) and opens a browser unless `--no-open`. Styling follows the
  Studio Bregalda tokens noted in `ui.go`.

## Gotchas

- SARIF can emit a finding for a **passed** run (cycles exist, or the scenario
  name contains "deadlock").
- HTML, UI, SARIF and Mermaid each pick an anomaly label with their own
  cycle-priority loop — labels can disagree with the terminal/JSON label
  (`chaossql-adya-anomaly-classification`).
- Reporters embed SQL text and data values; never ship their output to Cloud
  (the Cloud client strips them — `chaossql-cloud-publishing`).
- The pkg/proxy SARIF generator is separate (`pkg/proxy/sarif.go`).

## Change checklist

- New export: generator here, flag + export step in `executeChaos`
  (`chaossql-cli-run-pipeline`), `action.yml` input if CI-facing, docs.
- New SARIF rule: catalog entry, `mapAnomalyToRuleID`, proxy mapping, tests.
- Keep generators deterministic where possible (sorted maps) so tests stay stable.

## Tests

- `internal/reporter/*_test.go` (`html_test.go`, `junit_summary_test.go`,
  `otel_test.go`, `sarif_test.go`, `swarm_summary_test.go`, `ui_test.go`,
  `reporter_test.go`), `cmd/chaossql/ui_test.go`, `cmd/chaossql/sarif_cli_test.go`

## Source map

- `internal/reporter/terminal.go`
- `internal/reporter/mermaid.go`
- `internal/reporter/html.go`
- `internal/reporter/ui.go`
- `internal/reporter/ui_assets.go`
- `internal/reporter/otel.go`
- `internal/reporter/junit_summary.go`
- `internal/reporter/sarif.go`
- `internal/reporter/swarm_summary.go`
- `cmd/chaossql/ui.go`
- `specs/13_interactive_visualizer_sarif_and_register_checker.md`
- `docs/adrs/0006-bubbletea-terminal-ux.md`

## Related skills

- `chaossql-cli-run-pipeline`, `chaossql-repro-synthesis`,
  `chaossql-adya-anomaly-classification`, `chaossql-transparent-proxy`
