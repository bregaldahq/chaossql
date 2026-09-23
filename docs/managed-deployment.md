# Managed deployment (Fly.io + S3-compatible backups)

This runbook deploys one ChaosSQL Cloud control plane that serves the API and
the dashboard on the same origin, with continuous SQLite replication to
S3-compatible object storage through [Litestream](https://litestream.io).
The files live in `deploy/fly/`:

| File | Purpose |
| :--- | :--- |
| `Dockerfile` | Builds the `chaossql` binary, the dashboard, and pins Litestream by SHA-256. |
| `run.sh` | Restores the latest replica on an empty volume, then runs the server under Litestream. |
| `litestream.yml` | Replica settings, read from environment variables. |
| `admin.sh` | `chaossql-admin`: operator commands that drop root to the service user. |
| `restore-check.sh` | `chaossql-restore-check`: restore drill that never replicates. |
| `fly.toml` | One machine, one volume, health check, plan retention enabled. |
| `smoke.py` | Disposable local end-to-end check with MinIO (Docker required). |

The same image runs on any container host with a persistent volume; Fly.io is
only the documented default.

## Guarantees and limits

- Exactly one machine writes the SQLite database. Do not scale beyond one
  machine or attach the volume elsewhere.
- The container **refuses to start** unless replication is configured. For a
  local test only, `CHAOSSQL_ALLOW_NO_BACKUP=true` disables this guard.
- Litestream ships changes continuously (about one second behind). A machine or
  volume loss can lose at most the last seconds of writes.
- On a fresh, empty volume the container restores the latest replica before it
  serves traffic. An existing database file is never overwritten.
- Plan retention is enabled (`CHAOSSQL_ENFORCE_RETENTION=true`); see the
  self-hosted guide for what is deleted.
- Only metadata reaches the control plane: no SQL, parameters, schema, or
  reproduction code.

## 1. Object storage

Create a private bucket (for example Cloudflare R2 `chaossql-cloud-backups`) and
an access key limited to **object read and write on that bucket only**. Record:

| Secret | Example |
| :--- | :--- |
| `LITESTREAM_BUCKET` | `chaossql-cloud-backups` |
| `LITESTREAM_ENDPOINT` | `https://<account-id>.r2.cloudflarestorage.com` |
| `LITESTREAM_ACCESS_KEY_ID` | access key ID |
| `LITESTREAM_SECRET_ACCESS_KEY` | secret access key |

`LITESTREAM_PATH` (optional, default `chaossql-cloud`) is the prefix inside the
bucket. Use a different prefix for staging.

## 2. Application, volume, and secrets

From the repository root, with `flyctl` authenticated:

```bash
fly apps create chaossql-cloud
fly volumes create chaossql_data --app chaossql-cloud --region gru --size 1
fly secrets set --app chaossql-cloud \
  CHAOSSQL_ADMIN_TOKEN="$(openssl rand -hex 32)" \
  LITESTREAM_BUCKET=... LITESTREAM_ENDPOINT=... \
  LITESTREAM_ACCESS_KEY_ID=... LITESTREAM_SECRET_ACCESS_KEY=...
```

Store the admin token in your password manager before running the command; Fly
secrets cannot be read back. Adjust `app`, `primary_region`, and `PUBLIC_URL` in
`deploy/fly/fly.toml` if you use different names.

## 3. Deploy

```bash
fly deploy . --config deploy/fly/fly.toml --dockerfile deploy/fly/Dockerfile --ha=false
```

`--ha=false` keeps a single machine. Check the logs for the restore step,
`Plan retention enforcement enabled`, and the listening address:

```bash
fly logs --app chaossql-cloud
```

## 4. Domain

Point the public name (default `cloud.chaossql.bregalda.com`) at the app and
issue a certificate:

```bash
fly certs add cloud.chaossql.bregalda.com --app chaossql-cloud
```

Create the DNS records that command prints (a `CNAME` to
`chaossql-cloud.fly.dev` is typical). With Cloudflare DNS, keep the record
**DNS only** (not proxied) so Fly can validate and serve the certificate.
Then verify:

```bash
curl -fsS https://cloud.chaossql.bregalda.com/v1/health
```

## 5. Onboard an organization

```bash
fly ssh console --app chaossql-cloud -C "chaossql-admin org create --name 'Acme' --plan team"
```

Always use `chaossql-admin` (not `chaossql server`) inside the machine. Shell
access runs as root, and the wrapper switches to the service user so SQLite
never creates root-owned `-wal` or `-shm` files the server cannot open.

The command prints the organization ID and its owner token once. Send the token
privately. The owner opens `https://cloud.chaossql.bregalda.com/#/dashboard`,
connects with that token, and issues member tokens for CI.

## 6. Monitoring

- Fly runs the `/v1/health` check every 15 seconds and restarts a failing
  machine.
- Add an external uptime monitor on `https://<domain>/v1/health` that alerts a
  person; Fly checks alone do not notify you.
- Watch for `Retention pass failed` and Litestream errors in `fly logs`.

## 7. Restore drill (run before onboarding customers, then quarterly)

A backup is only proven by a restore. `chaossql-restore-check` restores the
latest replica into a scratch file and lists its organizations. It never starts
the server or Litestream replication, so it cannot write to the replica it
reads. Run it anywhere the image and the storage secrets are available, for
example on a workstation:

```bash
docker build -f deploy/fly/Dockerfile -t chaossql-cloud:drill .
docker run --rm --entrypoint chaossql-restore-check \
  -e LITESTREAM_BUCKET=... -e LITESTREAM_ENDPOINT=... \
  -e LITESTREAM_ACCESS_KEY_ID=... -e LITESTREAM_SECRET_ACCESS_KEY=... \
  chaossql-cloud:drill
```

Or inside the running machine:
`fly ssh console --app chaossql-cloud -C chaossql-restore-check`.

Never start the full image (its default entrypoint) against the production
bucket from a second place: two writers on one replica prefix corrupt the
backup history.

## 8. Disaster recovery

If the volume or machine is lost:

```bash
fly volumes create chaossql_data --app chaossql-cloud --region gru --size 1
fly deploy . --config deploy/fly/fly.toml --dockerfile deploy/fly/Dockerfile --ha=false
```

The new, empty volume triggers an automatic restore from the replica. Verify
health and `chaossql-admin org list` before announcing recovery.

## 9. Credential rotation

- **Admin token:** `fly secrets set CHAOSSQL_ADMIN_TOKEN=...` restarts the
  machine and revokes the previous `org_default` owner token.
- **Storage keys:** create the new key, update the two `LITESTREAM_*_KEY`
  secrets, confirm replication in the logs, then revoke the old key.
- **Customer tokens:** issue a replacement from the dashboard or
  `chaossql-admin create-token`; revocation from the dashboard is not
  available yet.
