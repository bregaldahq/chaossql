# CLAUDE.md — ChaosSQL Agent Instructions

@AGENTS.md

This file is the top-level instruction set for AI agents working in this
repository. `AGENTS.md` (imported above) holds the engineering contract;
this file adds the agent skill harness and its maintenance rules.

---

## 1. Mandatory Rules

1. **Keep the skills in sync — always.** Every change to code, configuration,
   CLI flags, file formats, protocols, workflows, or documentation that is
   covered by a skill under `.claude/skills/` **must update that skill in the
   same commit**. Before finishing any task:
   - look up the touched paths in `docs/harness/flow-map.md`;
   - update every affected `SKILL.md` (behavior, gotchas, checklists, source map);
   - if you add a new flow, create a new skill and register it in
     `docs/harness/flow-map.md` and in the index below;
   - if you rename, move, or delete a file, fix every `## Source map` that
     lists it.
   `make check-harness` fails when a skill is missing from the index, when an
   indexed skill does not exist, or when a source map path no longer resolves.
   The procedure is detailed in the `chaossql-skill-maintenance` skill.
2. **No co-authors.** Never add `Co-Authored-By` trailers, "Generated with"
   footers, or any other AI or co-author attribution to commit messages or
   pull request descriptions.
3. **English only** outside `site/` and `docs/superpowers/`, including skills
   (enforced by `tools/test_english_purity.js`).
4. **Never break `make verify`.** Run the narrowest relevant check while
   iterating (see `chaossql-quality-gate`) and the full gate before a PR.

---

## 2. Skill Index

Load the skill that matches the flow you are touching. Start with
`chaossql-harness-overview` when you do not know where something lives.

**Navigation**
- `chaossql-harness-overview` — repository map, layers, where each flow lives
- `chaossql-skill-maintenance` — how to keep skills, flow map and this file in sync

**Deterministic core engine**
- `chaossql-spec-format` — `chaos.yaml` schema, parsing, validation
- `chaossql-param-generators` — parameter generators, substitution, captures
- `chaossql-deterministic-schedule` — schedule plan, worker assignment, golden corpus
- `chaossql-runner-execution` — runner lifecycle, trace events, statuses
- `chaossql-invariant-evaluation` — SQL invariants, assertions, temporal invariants
- `chaossql-ddmin-shrinker` — failure signatures and delta debugging
- `chaossql-adya-anomaly-classification` — dependency graph, cycles, anomaly classes
- `chaossql-fault-injection` — latency spikes and aborts

**Ports and adapters**
- `chaossql-database-drivers` — SQLite, PostgreSQL, MySQL, Mock adapters

**CLI flows**
- `chaossql-cli-run-pipeline` — `run` and `demo`
- `chaossql-replay-artifacts` — `--export-result` and `replay --verify`
- `chaossql-differential-fuzzing` — `diff`, `matrix`, `swarm`
- `chaossql-scenario-tooling` — `init`, `validate`, `mutate`
- `chaossql-benchmarks` — `bench`

**Evidence and reporting**
- `chaossql-repro-synthesis` — Go, Python, TypeScript reproduction scripts
- `chaossql-report-exporters` — terminal, Mermaid, HTML, UI, OTLP, JUnit, summary, SARIF

**Embedding surfaces**
- `chaossql-engine-ipc` — `chaossql engine` JSON protocol
- `chaossql-go-testing-sdk` — `pkg/chaostest`
- `chaossql-sdk-python` — `sdks/python`
- `chaossql-sdk-typescript` — `sdks/typescript`
- `chaossql-transparent-proxy` — `chaossql proxy`
- `chaossql-wasm-playground` — WASM engine and browser worker

**ChaosSQL Cloud**
- `chaossql-cloud-publishing` — CLI-side Cloud client and privacy projection
- `chaossql-control-plane-api` — HTTP API, auth, roles, tenancy
- `chaossql-ingestion-baselines` — idempotent ingestion and regression detection
- `chaossql-webhooks-outbox` — webhook alerts, SSRF guard, outbox
- `chaossql-store-migrations` — control-plane SQLite schema and migrations
- `chaossql-plans-retention` — plans, limits, retention
- `chaossql-server-operations` — server entry points, admin CLI, deployment

**Web surfaces**
- `chaossql-website-portal` — React/Vite portal
- `chaossql-edge-worker` — Cloudflare worker and waitlist function

**Quality, process, distribution**
- `chaossql-quality-gate` — `make verify`, tools, CI
- `chaossql-github-action` — composite GitHub Action
- `chaossql-example-scenarios` — canonical anomaly scenarios
- `chaossql-docs-specs-adrs` — specs, ADRs, evals, ledgers
- `chaossql-release-process` — versioning, changelog, release notes
