# Specification 18: SaaS Tenant Authorization

## Status

Normative for the hosted HTTP control plane. The local CLI and engine remain independent of this specification.

## Identity

1. Every request that reads or changes tenant data MUST present a bearer token.
2. A token resolves to exactly one organization, one token identifier, and one role: `owner`, `admin`, or `member`.
3. Request identity MUST be carried in a server-owned context value. Headers, path values, query values, and bodies MUST NOT override the authenticated organization.
4. Missing, malformed, or unknown credentials return HTTP `401`.
5. Existing tokens created before role support migrate to `member`.
6. The hosted server MUST require an explicit bootstrap owner token. It MUST NOT ship a known default credential.
7. On startup, the bootstrap token record MUST be reconciled to the configured credential and `owner` role. Rotation replaces the previous credential, and persistence failures stop startup.

## Roles

1. Members MAY ingest runs and read runs, repositories, and subscriptions within their organization.
2. Admins inherit member access and MAY list, create, delete, and test webhooks within their organization. Webhook URLs are administrative secrets and MUST NOT be returned to members.
3. Owners inherit admin access.
4. An authenticated principal below the required role returns HTTP `403`.

## Object ownership

1. Run detail queries MUST join the run repository to the authenticated organization.
2. Repository lookup MUST use both organization ID and full name.
3. Recent-run lists MUST filter by organization in SQL.
4. Organization path values MUST equal the authenticated organization or use the alias `me`.
5. Cross-organization identifiers return HTTP `404` consistently.
6. Webhook deletion MUST include organization ID in the mutation and report a missing owned row as `404`.
7. Repository names are unique within an organization. Different organizations MAY store the same full name as separate records.

## Router modes

1. `NewRouter` is the hosted router and MUST enforce this matrix on every tenant route.
2. The hosted router MUST NOT register short unauthenticated dashboard routes.
3. `NewLocalRouter` is a separate, explicit constructor for a single local `org_default` owner context.
4. Environment variables and client network addresses MUST NOT turn the hosted router into local mode.

## Verification

Tests MUST use at least two organizations and prove absence of cross-tenant reads and writes for runs, repositories, subscriptions, and webhooks. Tests MUST also cover every route without credentials and member attempts to invoke administrative operations.
