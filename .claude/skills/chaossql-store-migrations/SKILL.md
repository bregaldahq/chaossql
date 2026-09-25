---
name: chaossql-store-migrations
description: The control-plane SQLite store in internal/server/store.go — single-connection design, table schema, token hashing, transaction binding, AutoMigrate and the one-time named migrations tracked in schema_migrations. Use when changing tables, queries, or adding a migration.
---

# Control Plane Store and Migrations

## When to use

- Adding a column/table, changing a query, or writing a migration.
- Upgrading a deployment with existing data.

## Design

- `NewStore(db)` forces `db.SetMaxOpenConns(1)`: SQLite foreign-key pragmas are
  per connection and SQLite has one writer; every query shares one connection.
- `Store{db, tx}`: `withTransaction` returns a bound store whose `queryer()`
  is the transaction, so helpers compose inside `ingest` and migrations.
- Tokens are stored as SHA-256 hex (`hashToken`); plaintext is shown once.
- IDs: `randomIdentifier(prefix)` = prefix + 48 hex (`run_`, `find_`,
  `outbox_`, `org_`, `tok_`); repositories use `repo_<unixnano>`.

## Tables (from `AutoMigrate` and migrations)

`organizations(id, name, plan, created_at)` ·
`api_tokens(id, org_id, token_hash UNIQUE, name, role DEFAULT 'member', created_at)` ·
`repositories(id, org_id, full_name, default_branch, created_at, UNIQUE(org_id, full_name))` ·
`scenarios(id, repo_id, name, driver, created_at, UNIQUE(repo_id, name))` ·
`runs(id, repo_id, scenario_id, commit_sha, branch, pr_number, status, anomaly_type, seed, duration_ms, created_at, idempotency_key, scenario_fingerprint, commit_timestamp)` ·
`findings(id, run_id, anomaly_type, assertion, minimal_ops, repro_code, trace_json, created_at)` ·
`webhooks(id, org_id, target_type, url, events, active, created_at)` ·
`baselines(id, repo_id, scenario_id, branch, run_id, updated_at, UNIQUE(repo_id, scenario_id, branch))` ·
`webhook_outbox(...)` · `run_ingestions(run_id PK, request_hash, response_json, driver, total_schedules, failed_schedules)` ·
`schema_migrations(version PK, applied_at)`.
Unique index `idx_runs_repo_idempotency ON runs(repo_id, idempotency_key) WHERE idempotency_key IS NOT NULL`.

## `AutoMigrate` order

1. `PRAGMA foreign_keys = ON` + `CREATE TABLE IF NOT EXISTS` for the base tables.
2. `ensureAPITokenRoleColumn` — adds `role` if missing.
3. `ensureTenantRepositoryUniqueness` — rebuilds `repositories` with
   `UNIQUE(org_id, full_name)` (foreign keys off during the copy).
4. `purgeLegacyFindingDetails` — migration `2026-09-16-purge-hosted-finding-details`
   blanks `assertion`, `repro_code`, `trace_json` of old findings.
5. `ensureIdempotentRunsAndOutbox` — migration `2026-09-17-idempotent-runs-and-outbox`
   adds idempotency/fingerprint/commit-timestamp columns, the unique index and
   `webhook_outbox`.
6. `ensureIngestionMetadata` — `run_ingestions` and migration
   `2026-09-21-correct-commit-timestamp-provenance` (clears upload times stored
   as commit times on runs without ingestion metadata).

Both server entry points call `AutoMigrate` on every start; it must stay
idempotent.

## Adding a migration

- Use a dated, descriptive `const migration = "YYYY-MM-DD-..."`, check
  `schema_migrations`, run inside a transaction, insert the version at the end.
- Detect columns with `PRAGMA table_info` before `ALTER TABLE ... ADD COLUMN`.
- Append the call at the end of `AutoMigrate`.
- Add an upgrade test that starts from the previous schema
  (`internal/server/ingestion_upgrade_test.go` is the pattern).
- Note it in `CHANGELOG.md` ("runs one-time migrations").

## Gotchas

- `findings.repro_code`/`trace_json`/`assertion` columns still exist but must
  stay empty (privacy); ingestion writes only metadata.
- Timestamps are compared in Go, not SQL (DATETIME text ordering is unreliable).

## Source map

- `internal/server/store.go`
- `internal/server/ingestion.go`
- `internal/server/store_test.go`
- `internal/server/ingestion_upgrade_test.go`
- `docs/adrs/0007-context-principal-and-tenant-scoped-storage.md`

## Related skills

- `chaossql-ingestion-baselines`, `chaossql-plans-retention`,
  `chaossql-control-plane-api`, `chaossql-server-operations`
