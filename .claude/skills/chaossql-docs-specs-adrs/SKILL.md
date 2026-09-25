---
name: chaossql-docs-specs-adrs
description: The living documentation system — numbered capability specs (specs/01-20), ADRs (docs/adrs), evals, theory and operations docs, and the historical superpowers planning ledgers — what each holds, which ones the harness check requires, and when a code change must update them. Use when a change alters documented behavior, when recording a decision, or when adding a spec.
---

# Specs, ADRs, Evals and Docs

`AGENTS.md` §3.2: behavior changes to the fuzzer, shrinker or mutator must
update the matching spec. Skills are the agent-facing layer; specs/ADRs are
the project's formal record. Both must stay true.

## Map

| Location | Content | Style |
| :--- | :--- | :--- |
| `specs/01..20_*.md` | one capability each (invariants, interleaving, shrinker, evidence, taxonomy, MySQL/OTEL, diff/matrix, faults, temporal/G2, tooling, SDK/generators, portal, visualizer/SARIF/register, WASM, swarm, proxy, SDKs, tenant authorization, payload privacy, ingestion/baselines) | Objective / design / acceptance; later ones carry a `Status` |
| `docs/adrs/000N-*.md` | immutable decisions (PRNG + replay, async interleaving, ddmin, pure-Go SQLite, repro synthesis, terminal UX, context principal + tenant scoping) | Status, date, context, decision, consequences |
| `evals/0N_*.md` | quality criteria: shrinking ratio > 85%, 0% false positives, 100% deterministic replay | Objective + method |
| `docs/THEORY.md`, `docs/ACADEMIC_FOUNDATIONS.md`, `docs/SCENARIO_ACADEMIC_AUDIT.md` | theory and per-scenario academic audit | prose |
| `docs/self-hosted.md`, `docs/managed-deployment.md`, `docs/cloud-early-access.md` | operations and onboarding | runbooks |
| `docs/releases/vX.Y.Z.md` | release notes | see `chaossql-release-process` |
| `docs/superpowers/{specs,plans}/` | historical design docs and implementation plans (may be Portuguese; excluded from the purity gate) | ledger, not a source of truth |
| `docs/harness/flow-map.md` | flow → skill → paths | see `chaossql-skill-maintenance` |

## Rules

- Specs 01–17 plus most ADRs and all evals are required files in
  `tools/harness_check.go`; renaming them requires updating that list
  (specs 18–20 and ADR 0007 are not in the list yet).
- A new architectural decision → new ADR with the next number; never rewrite
  an accepted ADR — supersede it.
- A new capability → new spec with the next number, referenced from the
  owning skill's source map.
- Do not treat `docs/superpowers/` or older specs as current behavior — verify
  against code (several specs describe intent that the code does not do, e.g.
  temporal invariants are not wired into runs).
- English only (except `docs/superpowers/`).

## Source map

- `specs/01_invariant_evaluation.md`
- `specs/20_saas_reliable_ingestion_and_baselines.md`
- `docs/adrs/0001-deterministic-prng-and-replay.md`
- `docs/adrs/0007-context-principal-and-tenant-scoped-storage.md`
- `evals/01_shrinking_ratio.md`
- `evals/02_false_positive_rate.md`
- `evals/03_deterministic_replay.md`
- `docs/THEORY.md`
- `docs/SCENARIO_ACADEMIC_AUDIT.md`
- `docs/superpowers/plans/swarm-execution-ledger.md`
- `docs/harness/flow-map.md`
- `tools/harness_check.go`

## Related skills

- `chaossql-skill-maintenance`, `chaossql-release-process`, `chaossql-quality-gate`
