# ChaosSQL site copy guide

How the marketing pages speak, and where every claim comes from. The copy lives
in `src/i18n/en.ts` (canonical) and `src/i18n/pt.ts`; the type system and
`src/i18n/i18n.test.ts` keep them in sync.

## Audience and order

Three readers, in this order of attention:

1. **Backend engineer** (user): has an intermittent balance or stock bug they
   cannot reproduce. Needs to *see the bug happen* and get a test.
2. **Tech lead** (champion): wants the bug to stay fixed. Needs CI evidence.
3. **CTO / head of engineering** (buyer): wants the risk gone before a launch.
   Needs a clear audit scope, price and timeline.

Every topic is written in three layers, and a page never skips a layer:

1. **Pain in business terms**: money, stock, on-call coverage.
2. **Mechanism in engineering terms**: interleaving, invariant, seed, replay.
3. **Theory as an optional depth layer**: Adya, dependency graphs, ddmin. Link to
   the docs; do not lead with it.

## Voice

- Short sentences. Concrete verbs: *reads, writes, erases, replays, keeps*.
- English first, written for a global audience; Portuguese is an adaptation,
  not a literal translation.
- No em or en dashes (a test enforces it on the dictionaries). Use a period,
  a comma or a colon.
- No emojis, no "VIP", no *seamless, revolutionary, next-gen, unleash, elevate*.
- One label per intent, everywhere on the site:

  | Intent | EN | PT |
  | :--- | :--- | :--- |
  | See it working | See the bug happen | Ver o bug acontecer |
  | Get the CLI | Install | Instalar |
  | Cloud interest | Join the Cloud waitlist | Entrar na lista do Cloud |
  | Buy the service | Book an audit | Agendar auditoria |

- Numbers are either measured (with a source below) or not written. "In
  milliseconds", "< 200ms" and invented seeds or balances are not allowed.
- Say what the product does *not* do when a reader would otherwise assume it
  (see "Will a violation fail my build?").

## Claims register

Re-check this table when the engine, the action or the prices change.

| Claim in the copy | Source of truth |
| :--- | :--- |
| Banking story numbers (seed 42, 1000 → 981 / 989, expected 970, 2 transactions) | Derived at runtime from `src/data/traces/banking_lost_update.json` by `src/data/story.ts`; asserted in `src/data/story.test.ts` |
| "From 20 transactions to 2", "90% smaller", replay count | `shrink` block of the same recorded run |
| "Under a second" for shrinking | Recorded runs take 0.2 to 0.7 s; `evals/01_shrinking_ratio.md` requires < 2 s |
| "At least 85% reduction" | `evals/01_shrinking_ratio.md` |
| Same spec and seed give the same schedule | `evals/03_deterministic_replay.md` (schedule identity); physical timing is explicitly *not* promised |
| Every run resets schema and seed data | Runner lifecycle: `driver.Reset` before each run and shrink trial (`internal/engine/runner.go`) |
| SQLite serializes writes at its default level | `chaossql-database-drivers` skill; example outcomes in `chaossql-example-scenarios` |
| A violation does not fail the build | `chaossql run` exits 0 on `violation` (`unreliableRunError`); `action.yml` step stays green |
| JUnit report and job summary from the action | `action.yml` inputs `export-junit`, `export-summary` (defaults to `$GITHUB_STEP_SUMMARY`) |
| PR comment and `is-regression` need Cloud | Both are produced only after a successful Cloud publish (`publishToCloud` in `cmd/chaossql/main.go`) |
| One static Go binary, no CGO | `AGENTS.md` principle 6; `go install ./cmd/chaossql` with `CGO_ENABLED=0` |
| MIT license | `LICENSE` |
| Cloud from $39 per month, early access | `src/pages/PricingPage.tsx` (Team, monthly) and `internal/server/billing.go` |
| Audit: $1,490, one week, 45 min kickoff, 5 transactions, 100,000+ schedules, 3 months Cloud Team | Confirmed offer; `pricing.audit` in `src/i18n/en.ts` |
| Cloud plan limits (1/10/30 repositories, 7/90/365 days) and prices ($0, $39/$31, $99/$79) | `internal/server/billing.go`; `CLOUD_PLANS` in `src/pages/PricingPage.tsx` |
| "Every Cloud plan includes": metadata-only uploads, baseline on main and `is-regression`, PR comments, Discord/Slack/generic webhooks, unlimited users | `chaossql-cloud-publishing`, `chaossql-ingestion-baselines` and `chaossql-webhooks-outbox` skills; plan feature flags are informational only, so no feature is sold as exclusive to a tier |
| Docs: exit codes, `run`/`demo` flag table, shrinker and scheduler descriptions | `cmd/chaossql/main.go` (flags, `unreliableRunError`), `internal/shrinker/ddmin.go`, `pkg/proxy/jitter.go` (PCT-style scheduler, proxy only); guarded by `src/data/docs-copy.test.ts` |
| Docs: "85% or more" and "under 2 s" for the shrinker | `evals/01_shrinking_ratio.md` targets, not per-run measurements |
| Not claimed on purpose | Nightly scheduled fuzzing, downloadable reproducers from Cloud (reproductions stay in the runner), PagerDuty, "advanced isolation policies", "most popular" (no customers yet) |

## What the recorded runs are (and are not)

`node tools/export_site_traces.mjs` runs three examples (`banking_lost_update`,
`inventory_oversell`, `hospital_write_skew`) on SQLite at `READ_UNCOMMITTED`
with their own seeds. Only the schedule, the shrink result and the minimal
trace are stable across runs; durations and the full-run invariant values vary
with machine timing, so the copy never quotes them exactly.

Two examples are deliberately not featured: `read_skew_financial_audit` (its
label flips between P4 and A5A across runs) and
`ticket_booking_anti_dependency` (its invariant also fails on serial
histories).
