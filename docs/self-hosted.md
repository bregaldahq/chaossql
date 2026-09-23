# Self-hosted deployment

The control plane receives execution metadata, maintains baselines, and queues
regression alerts. Tests run in the CLI against your own target database. SQL,
parameter values, schemas, and local reproduction artifacts stay with the runner.
The dashboard displays stored summaries; inspect a local trace or reproduction
artifact to investigate the actual operations.

The API uses organization-scoped bearer tokens. The configured bootstrap token
owns the default organization and can create member tokens for CI. Authentication
is token based; the pricing form requests contact and does not activate a paid
subscription. Plan limits in the API are separate from payment processing.

## Docker Compose

From the repository root, create independent credentials in a private file:

```bash
umask 077
cp .env.example .env
python3 - <<'PY'
from pathlib import Path
import secrets
path = Path('.env')
text = path.read_text()
for key in ('CHAOSSQL_ADMIN_TOKEN', 'POSTGRES_PASSWORD'):
    text = text.replace(key + '=\n', key + '=' + secrets.token_hex(32) + '\n')
path.write_text(text)
PY
docker compose up --build -d
docker compose ps
```

Compose refuses to start without both credentials. The server also rejects known
example tokens. Keep `.env` private and out of source control. Avoid printing
resolved Compose configuration because it contains credentials.

Default listeners bind only to localhost:

| Service | Address | Purpose |
| --- | --- | --- |
| Dashboard | `http://localhost:3000/#/dashboard` | Token login and metadata summaries |
| API | `http://localhost:8080/v1/health` | Health and authenticated API |
| PostgreSQL | `localhost:5432` | Test target, database `test` |

The dashboard proxies `/v1/` to the server on the private Compose network. Set
`PUBLIC_URL` to the dashboard's externally reachable origin when placing a TLS
reverse proxy in front of it. Adjust `PORT`, `DASHBOARD_PORT`, and `POSTGRES_PORT`
in `.env` if those host ports are occupied. The API image runs as UID 10001 and
nginx as UID 101 on container port 8080; `/wasm/` assets are included.

Enter the owner token in the dashboard, then generate a member token for CI in
onboarding. The browser retains credentials only in memory; reloading requires
authentication again. Member tokens can publish runs but cannot administer tokens
or webhooks. Save the issued token in your CI secret store.
Rotating `CHAOSSQL_ADMIN_TOKEN` and recreating the server revokes the previous
bootstrap token; separately issued member tokens remain valid.

The default data volumes remain `chaossql_enterprise_data` and
`chaossql_postgres_data` for existing installations. Override
`CHAOSSQL_DATA_VOLUME` and `POSTGRES_DATA_VOLUME` for isolated installations. Do
not use `docker compose down -v` when retaining data.

## SQLite storage, backup, and restore

Run exactly one server instance per SQLite database. The Helm chart rejects
multiple replicas and autoscaling; multiple independent database files do not
provide shared state. Keep `/data/chaossql-cloud.db` on persistent local storage.
No automatic snapshot schedule or recovery-time guarantee is provided.

For an online backup, use SQLite's backup API from a host or maintenance image
with SQLite tooling and access to the data volume. The API runtime image does not
include the `sqlite3` utility. Do not copy a live database file directly. For
example, when the volume is available at `/srv/chaossql` on the maintenance host:

```bash
sqlite3 /srv/chaossql/chaossql-cloud.db ".backup '/srv/backups/chaossql.db'"
sqlite3 /srv/backups/chaossql.db 'PRAGMA integrity_check;'
```

Restore first into a separate empty data directory or volume. Stop the server,
retain the original volume, copy the validated backup as `chaossql-cloud.db`, and
give UID/GID 10001 read/write access to the directory and file. Point
`CHAOSSQL_DATA_VOLUME` at the restored volume, restart the server, and verify both
health and authenticated run history. Restore the matching private environment
configuration as well. Test this procedure before relying on a backup.

### Plan retention

Each plan declares a history window: `developer` 7 days, `team` 90 days, `pro`
365 days, `enterprise` unlimited. Enforcement is **off by default** so an upgrade
never deletes existing history. To enable it, start the server with
`--enforce-retention` or set `CHAOSSQL_ENFORCE_RETENTION=true`.

When enabled, the server purges at startup and then hourly. For each
organization it deletes runs older than the plan window, together with their
findings and stored ingestion responses, and finished (delivered or failed)
alerts older than the window. It keeps runs that a current baseline references,
pending alerts, repositories, scenarios, webhooks, and tokens. Deletion is
permanent; take a backup before enabling enforcement on an existing database.
The bootstrap organization `org_default` uses the `pro` plan (365 days).

## Kubernetes

`charts/chaossql-server` deploys the API and its SQLite volume. It does not deploy
the dashboard or a target database. Build and push the API image from this checkout
to your own registry; set the repository and tag to an image you have verified.

Create a namespace and an existing Secret from a protected environment file whose
`CHAOSSQL_ADMIN_TOKEN` value is unique:

```bash
kubectl create namespace chaossql
kubectl -n chaossql create secret generic chaossql-credentials \
  --from-env-file=/secure/chaossql-server.env
helm upgrade --install chaossql-server ./charts/chaossql-server \
  --namespace chaossql \
  --set secrets.existingSecret=chaossql-credentials \
  --values /secure/chaossql-values.yaml
kubectl -n chaossql get pods,pvc,svc
```

Example non-secret values (replace the image and URL for your deployment):

```yaml
replicaCount: 1
image:
  repository: registry.example.com/chaossql-server
  tag: reviewed-build
config:
  publicUrl: https://chaossql.example.com
persistence:
  enabled: true
  size: 10Gi
autoscaling:
  enabled: false
```

Use a dashboard reverse proxy for both static content and `/v1/`. `publicUrl` must
point to that dashboard for run links to work. Configure ingress/TLS for the
deployment's network. `persistence.enabled=false` uses an ephemeral `emptyDir`
and loses data when the pod is removed. Updates use the `Recreate` strategy to
avoid simultaneous SQLite writers.

## Standalone server and CI

Build with Go 1.25 or newer and CGO disabled:

```bash
make build
export CHAOSSQL_ADMIN_TOKEN="$(openssl rand -hex 32)"
./bin/chaossql server start --db=/var/data/chaossql.db \
  --public-url=http://localhost:8080 --static-dir=site
```

Build the site first (`cd site && npm ci && npm run build`) when using
`--static-dir`. Ensure the database directory exists and is writable. Persist the
owner token privately before restarting. The separate `chaossql-server start`
entrypoint uses the same owner bootstrap rules.

An operator with direct database access can also create a member token:

```bash
./bin/chaossql server create-token --db=/var/data/chaossql.db \
  --org=org_default --name='CI publisher'
```

That command prints the newly issued token once; do not run it in a public CI log.
Configure CI with `CHAOSSQL_CLOUD_URL` and `CHAOSSQL_CLOUD_TOKEN` (the member token),
then run your scenario normally:

```bash
./bin/chaossql run chaos.yaml --cloud-fail-fast
```

The repository's composite GitHub Action accepts `cloud-url` and `cloud-token`
and emits `run-url` and `is-regression` after successful publication. Pin the
Action to the reviewed revision you deploy. Retain local CLI artifacts separately;
the hosted summary never includes their SQL or reproduction code.

## Air-gapped installations

Build images on a connected machine from the reviewed checkout, then transfer
the archives using your normal trusted process:

```bash
docker build -t chaossql-server:reviewed-build .
docker build -f Dockerfile.dashboard -t chaossql-dashboard:reviewed-build .
docker save -o chaossql-images.tar \
  chaossql-server:reviewed-build chaossql-dashboard:reviewed-build
# On the destination:
docker load -i chaossql-images.tar
```

Supply any target database image separately. Configure your deployment to use
these loaded images without a build or pull. External alert destinations require
outbound access; leave them unconfigured when that access is unavailable.
