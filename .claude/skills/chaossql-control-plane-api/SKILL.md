---
name: chaossql-control-plane-api
description: The ChaosSQL Cloud HTTP API in internal/server — route table, bearer-token authentication, owner/admin/member roles, tenant scoping via the context principal, the permissive local router, CORS, public response projections, and member token issuance. Use when adding or changing an endpoint, auth, roles, or what the API returns.
---

# Control Plane API

## When to use

- Adding/changing a route or handler in `internal/server/handlers.go`.
- Touching authentication, roles, tenant isolation, or response projections.

## Routes (`newRouter`)

| Method + path | Role | Handler |
| :--- | :--- | :--- |
| `GET /v1/health` | none | `{status, version: "1.0.0", time}` |
| `GET /v1/runs` | member | 50 most recent runs of the caller's org |
| `POST /v1/runs` | member | ingestion (`chaossql-ingestion-baselines`) |
| `GET /v1/runs/{id}` | member | run + finding, scoped to org |
| `GET /v1/repositories/{owner}/{name}/runs` | member | 50 latest runs of that repo |
| `GET /v1/organizations/{id}/subscription` | member | plan and usage |
| `POST /v1/organizations/{id}/tokens` | admin | issue a member token |
| `GET/POST /v1/organizations/{id}/webhooks` | admin | list / create |
| `DELETE /v1/organizations/{id}/webhooks/{wh_id}` | admin | delete |
| `POST /v1/organizations/{id}/webhooks/test` | admin | test an unsaved target |
| `POST /v1/organizations/{id}/webhooks/{wh_id}/test` | admin | test a saved webhook |
| `GET/POST /v1/webhooks`, `DELETE /v1/webhooks/{wh_id}`, `POST /v1/webhooks/test` | local router only | same handlers as local owner |

All routes are wrapped in a permissive CORS middleware (`*`, methods
GET/POST/PUT/DELETE/OPTIONS, headers Content-Type/Authorization/Idempotency-Key;
`OPTIONS` → 204).

## Authentication and authorization

- `requireAuth`: `Authorization: Bearer <token>`; `Store.AuthenticateToken`
  looks up the SHA-256 of the token (tokens are never stored in clear) and
  returns a `Principal{TokenID, OrgID, Role}` put in the request context.
- `requireRole`: `Role.Allows` ranks member(1) < admin(2) < owner(3).
- Tenant scoping: handlers read the org from the principal
  (`organizationFromContext`); `{id}` path params go through
  `organizationForRequest` — `me` or empty means the caller's org, another org
  ID returns **404** (not 403) to avoid leaking existence (ADR 0007, spec 18).
- `NewLocalRouter` swaps auth for `assumeLocalOwner` (`org_default`, owner).
  Production entry points must use `NewRouter`.

## Response projections

Stored records are never serialized directly. `publicRun` keeps status only if
`passed`/`failed`, anomaly type only if on the Cloud allow-list, and passes
identifiers through `safePublicIdentifier`; `populatePublicRuns` adds
`repo_full_name`, `scenario_name`, `driver`, `is_regression`,
`total_schedules`, `failed_schedules` from `run_ingestions` (nulls for legacy
rows). `publicFinding` exposes only safe metadata (no repro code or traces).
Webhook URLs are returned redacted (`redactedWebhookURL`).

## Member tokens

`POST /v1/organizations/{id}/tokens` accepts exactly `{"name": "..."}`
(1–128 chars, max 4 KiB body, unknown fields rejected), creates
`chaossql_<48 hex>` with role member, returns it once with
`Cache-Control: no-store`.

## Gotchas

- `/v1/health` reports a hard-coded version `1.0.0`, not `version.Version`.
- List endpoints have a fixed limit of 50 and no pagination parameters.
- `RouterConfig.Engine` is accepted but ingestion builds its own
  `RegressionEngine` inside the transaction.
- Creating the router starts an outbox dispatcher goroutine
  (`chaossql-webhooks-outbox`).

## Change checklist (new endpoint)

- Route with the minimal role through `protect(...)`; scope every query by
  org ID; return 404 for foreign resources.
- Add a public projection instead of encoding store records.
- Tests in `internal/server/handlers_test.go`; dashboard client in
  `site/src/lib/cloud-api.ts`; docs `docs/self-hosted.md`.

## Source map

- `internal/server/handlers.go`
- `internal/server/identity.go`
- `internal/server/member_tokens.go`
- `internal/server/store.go`
- `internal/server/run_metadata.go`
- `internal/server/handlers_test.go`
- `docs/adrs/0007-context-principal-and-tenant-scoped-storage.md`
- `specs/18_saas_tenant_authorization.md`

## Related skills

- `chaossql-ingestion-baselines`, `chaossql-webhooks-outbox`,
  `chaossql-store-migrations`, `chaossql-server-operations`, `chaossql-website-portal`
