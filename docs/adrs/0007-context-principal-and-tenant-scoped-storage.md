# ADR 0007: Context Principal and Tenant-Scoped Storage

## Status

Accepted

## Context

The initial control plane authenticated selected endpoints and copied the token organization into `X-Org-ID`. Several reads remained public, handlers trusted organization path values, and repository lookup used a globally unique full name. This allowed object identifiers and repository names to cross tenant boundaries.

## Decision

Bearer authentication produces a `Principal` with token ID, organization ID, and role. Middleware stores it under a private context key. Hosted handlers use that principal for all tenant access and apply the Owner/Admin/Member matrix in Specification 18.

Store methods used by HTTP handlers accept the organization ID and enforce ownership in their SQL queries. Repository uniqueness is `(org_id, full_name)`. Startup migration adds token roles and rebuilds the legacy repository constraint while preserving identifiers.

The hosted and local routers use separate constructors. The local constructor injects an `org_default` owner principal; no environment switch can relax the hosted router.

## Consequences

Existing API tokens receive the least-privileged `member` role. Operators must issue an admin or owner token for webhook management. Run and repository URLs no longer reveal resources owned by another organization. Local dashboard integrations must opt into `NewLocalRouter` explicitly.
