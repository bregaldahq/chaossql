---
name: chaossql-server-operations
description: Running and operating the ChaosSQL Cloud control plane — the two server entry points (chaossql-server and chaossql server) and their flags/env, bootstrap owner token rules, organization and token admin commands, static dashboard serving, and the deployment targets (Docker/Compose with nginx, Helm chart, Fly.io with Litestream backups, restore drill, smoke test). Use when changing server startup, admin CLI, or any deployment artifact.
---

# Server Operations and Deployment

## When to use

- Changing `cmd/chaossql-server/main.go`, `cmd/chaossql/server.go`,
  `internal/server/bootstrap.go`, `internal/server/provisioning.go`,
  `internal/serveradmin`, `deploy/`, `charts/`, `Dockerfile*`, `docker-compose.yml`.
- Provisioning tenants or rotating the owner token.

## Two entry points (keep them in sync)

| | `chaossql-server` (`cmd/chaossql-server`) | `chaossql server` (`cmd/chaossql/server.go`) |
| :--- | :--- | :--- |
| Start | root command or `start` | `server start` |
| Admin | `create-token`, `org create|list|set-plan` | same, under `server` |
| Static dashboard | no | `--static-dir` / `STATIC_DIR`: `/v1/` and `/health` → API, everything else → SPA file server |
| Used by | `Dockerfile` (Compose, Helm) | Fly image (`deploy/fly/run.sh`, `admin.sh`) |

Shared flags/env: `--port/-p` (`PORT`, 8080), `--db` (`DB_PATH`,
`chaossql-cloud.db`), `--token` (`CHAOSSQL_ADMIN_TOKEN`, required),
`--public-url` (`PUBLIC_URL`, `http://localhost:8080`, used in run URLs),
`--enforce-retention` (`CHAOSSQL_ENFORCE_RETENTION=true`).

Startup: `ValidateBootstrapToken` → open SQLite (`modernc`) → `NewStore` →
`AutoMigrate` → `ConfigureBootstrapOwner` → `NewRouter` → HTTP server
(15 s read/write timeouts) → optional retention loop → graceful shutdown on
SIGINT/SIGTERM (5 s).

## Bootstrap owner

`ValidateBootstrapToken` rejects empty tokens, surrounding whitespace, and the
public example tokens (`chaossql_dev_token`, `chaossql_prod_secret`, ...).
`ConfigureBootstrapOwner` ensures `org_default` ("Default Organization", plan
`pro`) and upserts token ID `tok_admin` with role owner — restarting with a
new token **rotates** the owner credential.

## Admin CLI (`internal/serveradmin`)

- `org create --name N [--plan developer|team|pro|enterprise]` → new
  `org_<hex>` plus an owner token printed **once**.
- `org list` → organizations with plan and usage.
- `org set-plan <org-id> <plan>`.
- `create-token [--org org_default] [--name "CI Token"] [--role member|admin|owner]`
  → requires an existing organization.
All open the DB at `--db` directly (run them where the volume is mounted; on
Fly use `deploy/fly/admin.sh`, which drops root to the service user).

## Deployment targets

- **Docker/Compose** (`Dockerfile`, `Dockerfile.dashboard`, `docker-compose.yml`):
  server image runs `chaossql-server start` as UID 10001 with `/data` volume
  and a `/v1/health` healthcheck; the dashboard image builds `site/` and serves
  it with nginx, proxying `/v1/` to the server (`deploy/docker/nginx*.conf`);
  optional PostgreSQL service. `.env.example` lists variables.
- **Helm** (`charts/chaossql-server`): single replica (SQLite = one writer),
  PVC, secret for the admin token (`secrets.adminToken` or `existingSecret`),
  image `ghcr.io/bregaldahq/chaossql-server:<tag>`.
- **Fly.io** (`deploy/fly/`): one image containing `chaossql`, the built site
  and Litestream; `run.sh` refuses to start without `LITESTREAM_BUCKET` unless
  `CHAOSSQL_ALLOW_NO_BACKUP=true`, restores the latest replica on an empty
  volume, then runs `litestream replicate -exec "chaossql server start"`.
  `fly.toml`: region `gru`, one machine, auto-stop off (outbox + retention run
  in-process), `/v1/health` check, `CHAOSSQL_ENFORCE_RETENTION=true`.
  `restore-check.sh` restores into a scratch file and lists tenants;
  `smoke.py` is a disposable end-to-end test with MinIO.
  Runbook: `docs/managed-deployment.md`.

## Gotchas

- Never scale beyond one replica/machine: SQLite single writer, in-process
  dispatcher and retention loop.
- Changes to startup wiring must be applied to **both** entry points
  (bootstrap and retention are shared helpers for that reason).
- Helm `image.tag`/`appVersion` and chart `version` are bumped on release
  (`chaossql-release-process`).

## Tests

- `cmd/chaossql/server_test.go`, `cmd/chaossql/server_bootstrap_test.go`,
  `cmd/chaossql/spa_file_server_test.go`, `internal/serveradmin/commands_test.go`,
  `internal/server/provisioning_test.go`; Fly: `deploy/fly/smoke.py`.

## Source map

- `cmd/chaossql-server/main.go`
- `cmd/chaossql/server.go`
- `internal/server/bootstrap.go`
- `internal/server/provisioning.go`
- `internal/serveradmin/commands.go`
- `Dockerfile`
- `Dockerfile.dashboard`
- `docker-compose.yml`
- `deploy/docker/nginx.conf`
- `deploy/fly/fly.toml`
- `deploy/fly/run.sh`
- `deploy/fly/litestream.yml`
- `deploy/fly/restore-check.sh`
- `charts/chaossql-server/values.yaml`
- `docs/self-hosted.md`
- `docs/managed-deployment.md`
- `.env.example`

## Related skills

- `chaossql-control-plane-api`, `chaossql-store-migrations`,
  `chaossql-plans-retention`, `chaossql-release-process`
