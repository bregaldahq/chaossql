---
name: chaossql-harness-overview
description: Repository map for ChaosSQL. Use first when you need to find where a flow lives, which layer owns a behavior, which skill to load next, or how the engine, CLI, SDKs, proxy, WASM playground, Cloud control plane and website fit together.
---

# ChaosSQL Harness Overview

ChaosSQL is a deterministic concurrency fuzzer for SQL databases. A scenario
(`chaos.yaml`) declares a schema, seed data, SQL invariants and transactional
operations. The engine generates a seeded schedule of operations, runs them on
concurrent workers against a real database, checks the invariants, classifies
the anomaly from the trace (Adya dependency cycles), and shrinks the failing
schedule with delta debugging (ddmin) to a 1-minimal reproduction.

## When to use

- You do not yet know which package or skill owns a behavior.
- You need the end-to-end picture before a cross-cutting change.
- You are onboarding to the repository.

## The pipeline in one picture

```
chaos.yaml ──LoadSpec──▶ domain.Spec
                           │
            GenerateSchedule (seeded PCG)      ◀── chaossql-deterministic-schedule
                           │  []ScheduledOp (IDs 1..N)
            BuildSchedulePlan (per-step jitter/latency/abort)
                           │
  Runner.Run: driver.Reset(schema, seed) → N worker goroutines
            → BEGIN / steps / COMMIT per op → ExecutionTrace
                           │                    ◀── chaossql-runner-execution
            finalizeResult: canceled > execution_error > invariants
                           │                    ◀── chaossql-invariant-evaluation
            analyzer.BuildGraph → FindCycles → ClassifyCycle
                           │                    ◀── chaossql-adya-anomaly-classification
            if violation: shrinker.Shrink (ddmin over ScheduledOps)
                           │                    ◀── chaossql-ddmin-shrinker
            reporters (terminal, repro, Mermaid, HTML, SARIF, OTLP, JUnit, ...)
                           │                    ◀── chaossql-report-exporters / chaossql-repro-synthesis
            optional Cloud publish (metadata only)
                                                ◀── chaossql-cloud-publishing
```

## Layers (from `ARCHITECTURE.md`)

| Layer | Packages | Rule |
| :--- | :--- | :--- |
| Domain | `internal/domain`, `internal/analyzer`, `internal/evaluator` (expression side) | No database driver awareness |
| Application | `internal/engine`, `internal/shrinker`, `internal/swarm` | Talks to databases only through `drivers.DatabaseDriver` |
| Ports/Adapters | `internal/drivers` | Implement the `DatabaseDriver` interface |
| Presentation | `cmd/*`, `internal/reporter` | Consume domain artifacts; never re-implement execution |
| Embedding | `pkg/chaostest`, `pkg/proxy`, `pkg/mutator`, `sdks/*`, `cmd/chaossql-wasm` | Public or out-of-process surfaces |
| Cloud | `internal/cloud` (client), `internal/server`, `internal/serveradmin`, `cmd/chaossql-server` | Metadata-only SaaS control plane |
| Web | `site/`, `worker.ts`, `functions/` | Portal, playground, dashboard, edge worker |

## Binaries

| Binary | Entry | Notes |
| :--- | :--- | :--- |
| `chaossql` | `cmd/chaossql` | Commands: `run demo bench diff matrix replay init validate ui mutate swarm proxy engine server` (see `cmd/chaossql/root.go`) |
| `chaossql-server` | `cmd/chaossql-server` | Standalone control plane; duplicates `chaossql server` wiring |
| `chaossql.wasm` | `cmd/chaossql-wasm` | Browser engine, always uses the mock driver |

## Where to go next

Every flow has a skill; the authoritative mapping (flow → skill → paths) is
`docs/harness/flow-map.md`, and `CLAUDE.md` has the grouped index.

Quick routing:
- Scenario syntax → `chaossql-spec-format`, `chaossql-param-generators`
- "Why did the same seed produce a different result?" → `chaossql-deterministic-schedule`, `chaossql-runner-execution`
- Wrong anomaly label → `chaossql-adya-anomaly-classification`
- Shrink did not reduce / did not reproduce → `chaossql-ddmin-shrinker`
- A driver behaves differently → `chaossql-database-drivers`
- New CLI flag or export → `chaossql-cli-run-pipeline`, `chaossql-report-exporters`
- Anything under `internal/server` → the Cloud skills
- CI red → `chaossql-quality-gate`

## Non-negotiables (from `AGENTS.md`)

- Same spec + same seed ⇒ same logical schedule (operation choice, params,
  worker assignment, jitter/latency/abort decisions).
- Every scenario has an invariant; no sleep-based assertions.
- Violations are minimized with ddmin.
- `CGO_ENABLED=0` for every Go target.
- English outside `site/` and `docs/superpowers/`.
- Update the matching skill in the same commit as the code change.

## Source map

- `AGENTS.md` — engineering contract
- `ARCHITECTURE.md` — layer diagram and lifecycle
- `CLAUDE.md` — agent rules and skill index
- `docs/harness/flow-map.md` — flow → skill → paths
- `cmd/chaossql/root.go` — command tree
- `go.mod` — module `github.com/bregaldahq/chaossql`, Go 1.25

## Related skills

- `chaossql-skill-maintenance`
- `chaossql-quality-gate`
