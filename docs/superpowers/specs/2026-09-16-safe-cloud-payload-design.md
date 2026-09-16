# SEC-02 Safe Cloud Payload Design

## Problem

The cloud client currently sends SQL text, table names, invariant result values, replay schedules, generated reproduction code, diagrams, and CI actor metadata. Pattern-based replacement cannot prove that arbitrary SQL literals, errors, schemas, or generated artifacts are free of customer data.

## Decision

The default and only supported hosted ingestion policy is a metadata allowlist. The client creates a new payload containing only fields required for tenant routing, regression comparison, aggregate execution status, and numeric reduction metrics. It never serializes the original request directly.

Allowed metadata:

- protocol version and timestamp;
- CI provider, repository, commit, branch, base branch, pull request number, and provider run ID;
- scenario name, fingerprint, database engine/version, workers, iterations, and effective seed;
- execution status, success flags, anomaly classification, duration, and schedule counts;
- failing invariant name;
- minimal operation count and shrink duration.

The following remain local:

- SQL and table/schema names;
- operation parameters and query results;
- invariant query, expression, and actual values;
- deterministic schedule decisions;
- traces, generated reproduction source, and diagrams;
- CI actor identity.

The server independently rejects requests containing forbidden detail fields. It also rejects unknown JSON fields and bodies larger than 64 KiB before persistence. Detailed upload is intentionally unavailable until a future design provides project-level consent, preview, access control, retention, and deletion as one complete feature.

## Compatibility

Older clients that submit detailed fields receive HTTP 400 with a privacy-policy error. Structural types retain deprecated detail fields temporarily so the server can recognize and reject old payloads instead of silently accepting them.

## Verification

- Capture the exact HTTP body produced from a request seeded with email addresses, credentials, DSNs, SQL literals, result values, schema names, traces, and generated reproductions.
- Assert none of the sentinels or forbidden JSON keys appear in the body.
- Assert the original request is unchanged.
- Assert the server rejects forbidden and unknown fields before writing a run.
- Assert client and server enforce the payload size limit.
- Keep local reports useful without claiming that detailed evidence is uploaded.
