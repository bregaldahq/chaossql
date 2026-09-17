# Specification 19: SaaS Cloud Payload Privacy

## Status

Normative for hosted ingestion and remote CI reporting.

## Metadata allowlist

1. The client MUST construct a new metadata payload at the network boundary. It MUST NOT serialize an execution object directly.
2. Hosted ingestion MAY receive protocol time/version, repository and revision identifiers, scenario identity and engine configuration, aggregate execution status, anomaly classification, invariant name, and numeric reduction metrics.
3. SQL, table or schema names, operation parameters, query results, invariant expressions and actual values, schedule decisions, traces, generated source, diagrams, and CI actor identity MUST remain local.
4. Adding a field to an internal execution type MUST NOT add that field to the hosted payload automatically.
5. The client MUST reject a metadata payload larger than 64 KiB before opening an HTTP request.
6. Every allowed string MUST use bounded structural-identifier syntax. Free text, email addresses, credential-bearing URLs, DSNs, query strings, and fragments MUST be rejected.
7. Repository discovery MUST strip remote user information, query strings, fragments, and local filesystem paths before projection.

## Server enforcement

1. The server MUST authenticate the caller before decoding tenant data.
2. The server MUST reject unknown JSON fields and multiple JSON values.
3. The server MUST reject bodies larger than 64 KiB with HTTP `413`.
4. The server MUST reject known legacy detail fields with HTTP `400` before creating a repository, scenario, run, or finding.
5. Detailed upload is unsupported until project consent, exact preview, access policy, retention, and deletion are implemented together.
6. The privacy migration MUST clear historical assertion values, reproduction source, and trace JSON exactly once.
7. Run detail and list responses MUST suppress unsafe historical run strings. Finding responses MUST omit reproduction and trace storage fields and suppress unsafe legacy assertion or anomaly values.

## Remote reports

GitHub comments and step summaries MAY contain the same allowed metadata and a structural list of worker/operation categories. They MUST NOT include SQL, schemas, invariant expressions, actual values, or generated reproduction source.

## Verification

Tests MUST capture exact outbound bytes seeded with representative emails, credentials, DSNs, SQL literals, schema names, results, and reproduction source. None may occur in the captured payload or remote report. The server tests MUST prove rejection happens before persistence.
