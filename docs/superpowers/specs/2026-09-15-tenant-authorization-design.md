# SEC-01 Tenant Authorization Design

## Problem

The SaaS router authenticates only some endpoints and passes organization identity through the caller-visible `X-Org-ID` header. Run lookups and repository queries are global, organization path parameters are trusted without comparison to the authenticated organization, and repository names are globally unique. These behaviors allow one tenant to address another tenant's resources.

## Identity model

Bearer tokens resolve to an internal `Principal` containing token ID, organization ID, and one of three roles: `owner`, `admin`, or `member`. Middleware stores the principal in an unexported context key. Handlers never accept identity from request headers, bodies, or path parameters.

Existing tokens migrate to `member`, which preserves run ingestion and read access without granting webhook administration. Bootstrap administration tokens are created explicitly as `owner`.

## Authorization matrix

| Endpoint | Minimum role | Resource rule |
|---|---|---|
| `GET /v1/health` | Public | No tenant data |
| `GET /v1/runs` | Member | List caller organization only |
| `POST /v1/runs` | Member | Create within caller organization only |
| `GET /v1/runs/{id}` | Member | Run repository must belong to caller organization |
| `GET /v1/repositories/{owner}/{name}/runs` | Member | Repository must belong to caller organization |
| `GET /v1/organizations/{id}/subscription` | Member | `{id}` is `me` or caller organization |
| `GET /v1/organizations/{id}/webhooks` | Admin | `{id}` is `me` or caller organization |
| Webhook create, delete, and test | Admin | `{id}` is `me` or caller organization |

Cross-organization resource access returns `404` consistently so the API does not disclose whether another tenant owns an identifier. Missing or invalid credentials return `401`; an authenticated member attempting an administrative operation receives `403`.

## Store boundaries

Tenant-aware store methods require `orgID` and enforce ownership in SQL joins. Repository lookup and uniqueness use `(org_id, full_name)`. A schema migration adds token roles and rebuilds the legacy global repository uniqueness constraint while preserving repository IDs and dependent records.

## Local mode

`NewRouter` always installs the SaaS authorization matrix. Local permissive behavior is available only through the separate `NewLocalRouter` constructor, which injects an owner principal for `org_default` and exposes short dashboard webhook routes. The production server entry point uses `NewRouter`.

## Acceptance

- Every tenant-data route returns `401` without a bearer token.
- Organization A cannot read, mutate, delete, test, or ingest into organization B.
- Member tokens can ingest and read runs, repositories, and subscriptions in their tenant but cannot read webhook URLs or administer webhooks.
- Owner and admin tokens can administer webhooks only for their tenant.
- Two organizations may use the same repository full name and receive distinct records.
- Direct store methods cannot return a run or repository outside the supplied organization.
- Local permissive routes exist only on `NewLocalRouter`.
- `make verify`, race tests, and zero-CGO builds pass.
