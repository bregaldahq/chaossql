---
name: chaossql-plans-retention
description: Cloud plans (developer, team, pro, enterprise) and what they actually enforce — repository limits and retention windows — plus the subscription endpoint and the hourly retention purge that keeps baseline runs and pending alerts. Use when changing plans, prices, limits, or retention.
---

# Plans and Retention

## When to use

- Changing `internal/server/billing.go` or `internal/server/retention.go`.
- Changing pricing shown on the portal.

## Plans (`billing.go`)

| ID | Name | Monthly / annual USD | Max repos | Retention days |
| :--- | :--- | :--- | :--- | :--- |
| `developer` (default for unknown IDs) | Cloud Developer | 0 / 0 | 1 | 7 |
| `team` | Cloud Team | 39 / 31 | 10 | 90 |
| `pro` | Cloud Pro | 99 / 79 | 30 | 365 |
| `enterprise` | Enterprise Custom | 499 / 399 | unlimited (-1) | unlimited (-1) |

Flags `has_regression_gating`, `has_pr_comments`, `has_priority_support` are
**informational only** — nothing enforces them.

## What is enforced

- **Repository limit**: `GetOrCreateRepo` refuses a new repository when
  `count >= MaxRepositories` → `ErrPlanLimitReached` → HTTP 403 on ingestion.
- **Retention** (opt-in: `--enforce-retention` or
  `CHAOSSQL_ENFORCE_RETENTION=true`): `StartRetention` runs
  `PurgeExpiredData` immediately and every `RetentionInterval` (1 h).
  Per organization with a finite window, cutoff = now − days; deletes runs
  created before the cutoff **unless referenced by a baseline**, with their
  findings and `run_ingestions`, and deletes non-pending outbox rows older
  than the cutoff. Timestamps are compared in Go. Returns a `RetentionReport`.

`GET /v1/organizations/{id}/subscription` returns the plan, active repository
count, remaining repositories (-1 unlimited), and `can_add_repository`.
Plans change via `chaossql server org set-plan <org> <plan>`
(`chaossql-server-operations`).

## Gotchas

- The bootstrap organization `org_default` is created on the `pro` plan.
- Unknown plan strings in the DB silently behave as `developer`; the admin CLI
  validates plan IDs (`validPlanID`).
- Portal pricing is hard-coded separately in the site — keep it in sync.

## Change checklist

- Plan change → `billing.go`, portal pricing page (`site/src/pages/PricingPage.tsx`),
  `docs/cloud-early-access.md` / `docs/managed-deployment.md`, tests.
- New enforced feature → enforce server-side, not only in the UI.

## Tests

- `internal/server/billing_test.go`, `internal/server/retention_test.go`,
  `internal/server/provisioning_test.go`

## Source map

- `internal/server/billing.go`
- `internal/server/retention.go`
- `internal/server/store.go`
- `internal/server/provisioning.go`
- `internal/server/retention_test.go`
- `site/src/pages/PricingPage.tsx`

## Related skills

- `chaossql-server-operations`, `chaossql-store-migrations`,
  `chaossql-ingestion-baselines`, `chaossql-website-portal`
