# ChaosSQL Flow Map — Agent Skill Harness

This document is the inventory of every flow in the repository that has a
dedicated agent skill under `.claude/skills/`. It is the single source of truth
for the mapping **flow → skill → source paths**. `make check-harness` verifies
that every skill listed here exists and that every skill on disk is listed here.

Rules for keeping this map and the skills in sync live in `CLAUDE.md` and in
the `chaossql-skill-maintenance` skill.

---

## 1. Navigation

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-harness-overview` | Repository map, layer boundaries, which skill covers what | `AGENTS.md`, `ARCHITECTURE.md`, `CLAUDE.md` |
| `chaossql-skill-maintenance` | Keeping skills, this map and `CLAUDE.md` in sync with the code | `.claude/skills/`, `docs/harness/flow-map.md`, `tools/harness_check.go` |

## 2. Deterministic Core Engine

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-spec-format` | `chaos.yaml` schema, parsing, validation, external SQL resolution, honored vs. ignored fields | `internal/domain/types.go`, `internal/domain/parser.go`, `internal/domain/errors.go` |
| `chaossql-param-generators` | `$random_*` / `$faker_*` / `$uuid` / `$monotonic_counter` generators, `{param}` substitution, inline arithmetic, `capture` | `internal/engine/prng.go`, `internal/engine/runner.go` |
| `chaossql-deterministic-schedule` | Operation selection, worker assignment, per-step jitter/latency/abort decisions, schedule plan v1 and golden corpus | `internal/engine/schedule.go`, `internal/engine/runner.go`, `internal/engine/testdata/deterministic_schedule_v1.json` |
| `chaossql-runner-execution` | Reset → schedule → concurrent workers → trace → status resolution | `internal/engine/runner.go` |
| `chaossql-invariant-evaluation` | SQL invariant query + `expr-lang` assertion; temporal invariants library | `internal/evaluator/evaluator.go`, `internal/evaluator/temporal.go` |
| `chaossql-ddmin-shrinker` | Failure signatures, Zeller ddmin with memoization, baseline check, 1-minimality audit | `internal/shrinker/ddmin.go` |
| `chaossql-adya-anomaly-classification` | Trace → Adya dependency graph → cycles → anomaly class; Elle-style register checker | `internal/analyzer/adya.go`, `internal/analyzer/register.go` |
| `chaossql-fault-injection` | `engine.faults` → planned latency spikes and aborts; legacy `internal/faults` | `internal/engine/schedule.go`, `internal/faults/fault.go` |

## 3. Ports and Adapters

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-database-drivers` | `DatabaseDriver` port, SQLite/PostgreSQL/MySQL/Mock adapters, isolation resolution, error mapping, reset semantics, WASM build tags | `internal/drivers/` |

## 4. CLI Flows (`cmd/chaossql`)

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-cli-run-pipeline` | `run` and `demo`: flag overrides, run → classify → shrink → exports → cloud publish → exit code | `cmd/chaossql/main.go`, `cmd/chaossql/root.go` |
| `chaossql-replay-artifacts` | `--export-result` artifact v1, `replay`, `replay --verify` | `cmd/chaossql/replay.go`, `cmd/chaossql/replay_file_unix.go`, `cmd/chaossql/replay_file_windows.go` |
| `chaossql-differential-fuzzing` | `diff`, `matrix`, `swarm` cross-engine comparison | `cmd/chaossql/diff.go`, `cmd/chaossql/matrix.go`, `cmd/chaossql/swarm.go`, `internal/engine/diff.go`, `internal/swarm/diff_runner.go` |
| `chaossql-scenario-tooling` | `init`, `validate`, `mutate` (adversarial mutator) | `cmd/chaossql/init.go`, `cmd/chaossql/validate.go`, `cmd/chaossql/mutate.go`, `pkg/mutator/mutator.go` |
| `chaossql-benchmarks` | `bench` micro-benchmarks and database stress | `cmd/chaossql/bench.go` |

## 5. Evidence and Reporting

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-repro-synthesis` | Standalone Go / Python / TypeScript reproduction scripts | `internal/reporter/repro.go`, `internal/reporter/repro_python.go`, `internal/reporter/repro_ts.go` |
| `chaossql-report-exporters` | Terminal, Mermaid, HTML, trace viewer (`ui`), OTLP, JUnit, Step Summary, SARIF, swarm summary | `internal/reporter/`, `cmd/chaossql/ui.go` |

## 6. Embedding Surfaces

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-engine-ipc` | `chaossql engine` JSON stdin/stdout protocol used by SDKs | `cmd/chaossql/engine.go` |
| `chaossql-go-testing-sdk` | `pkg/chaostest` fluent Go testing API | `pkg/chaostest/chaostest.go` |
| `chaossql-sdk-python` | `chaossql-py` harness, binary discovery, pytest plugin | `sdks/python/` |
| `chaossql-sdk-typescript` | `@chaossql/test` harness and IPC client | `sdks/typescript/` |
| `chaossql-transparent-proxy` | Layer-7 PostgreSQL/MySQL proxy, PCT jitter, live shadow graph, SARIF, live UI | `pkg/proxy/`, `cmd/chaossql/proxy.go` |
| `chaossql-wasm-playground` | Go → WASM engine, JS worker protocol, browser bridge, headless stress tests | `cmd/chaossql-wasm/`, `site/assets/wasm-worker.js`, `site/src/lib/wasm-bridge.ts` |

## 7. ChaosSQL Cloud (SaaS)

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-cloud-publishing` | CLI side: fingerprint, CI detection, metadata-only projection, retries, idempotency, PR comment, action outputs | `internal/cloud/` |
| `chaossql-control-plane-api` | HTTP routes, bearer auth, roles, tenant scoping, public projections, local router | `internal/server/handlers.go`, `internal/server/identity.go`, `internal/server/member_tokens.go` |
| `chaossql-ingestion-baselines` | Transactional idempotent ingestion, regression engine, baselines | `internal/server/ingestion.go`, `internal/server/regression.go`, `internal/server/run_metadata.go` |
| `chaossql-webhooks-outbox` | Webhook registration, SSRF guard, transactional outbox, Discord/Slack/generic dispatch | `internal/server/webhooks.go` |
| `chaossql-store-migrations` | SQLite schema, `AutoMigrate` and one-time migrations, tenant-scoped queries | `internal/server/store.go` |
| `chaossql-plans-retention` | Plans, repository limits, retention purge loop | `internal/server/billing.go`, `internal/server/retention.go` |
| `chaossql-server-operations` | Two server entry points, bootstrap owner token, org/token admin CLI, Docker/Compose/Helm/Fly/Litestream | `cmd/chaossql-server/`, `cmd/chaossql/server.go`, `internal/server/bootstrap.go`, `internal/server/provisioning.go`, `internal/serveradmin/`, `deploy/`, `charts/` |

## 8. Web Surfaces

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-website-portal` | React/Vite portal, path routing, route metadata, i18n, dashboard client, build output | `site/` |
| `chaossql-edge-worker` | Cloudflare worker: waitlist, webhook test relay, rate limits, per-route SEO shell | `worker.ts`, `site/_worker.js`, `functions/api/waitlist.ts`, `wrangler.toml` |

## 9. Quality, Process and Distribution

| Skill | Flow | Primary paths |
| :--- | :--- | :--- |
| `chaossql-quality-gate` | `make verify` stages, English purity, harness check, CI workflows, database integration env | `Makefile`, `tools/`, `.github/workflows/ci.yml`, `codecov.yml` |
| `chaossql-github-action` | Composite GitHub Action and the workflows that dogfood it | `action.yml`, `.github/workflows/concurrency-ci.yml`, `.github/workflows/swarm.yml` |
| `chaossql-example-scenarios` | Canonical anomaly catalog and the checklist for adding a scenario | `examples/` |
| `chaossql-docs-specs-adrs` | Living documentation: specs, ADRs, evals, planning ledgers | `specs/`, `docs/` , `evals/` |
| `chaossql-release-process` | Version sources, changelog, release notes | `internal/version/version.go`, `CHANGELOG.md`, `docs/releases/` |

---

## 10. Cross-Cutting Facts Every Skill Must Respect

- **Determinism:** the logical schedule depends only on the spec and the seed;
  wall-clock time only affects trace timestamps and real database interleaving.
- **Language purity:** everything outside `site/` and `docs/superpowers/` is
  English (`tools/test_english_purity.js`). Skills are English too.
- **Zero CGO:** all Go targets build with `CGO_ENABLED=0`.
- **Privacy:** hosted Cloud payloads carry metadata only (no SQL, parameters,
  schema, traces, or reproduction code).
