#!/bin/sh
# Entrypoint for the managed image. A managed database without off-machine
# replication is one volume failure away from total loss, so startup refuses to
# continue without Litestream unless explicitly allowed for local testing.
set -eu

: "${DB_PATH:?DB_PATH is required}"
: "${CHAOSSQL_ADMIN_TOKEN:?CHAOSSQL_ADMIN_TOKEN is required}"
export LITESTREAM_PATH="${LITESTREAM_PATH:-chaossql-cloud}"

if [ -z "${LITESTREAM_BUCKET:-}" ]; then
  if [ "${CHAOSSQL_ALLOW_NO_BACKUP:-}" = "true" ]; then
    echo "[chaossql-cloud] WARNING: running without replication (CHAOSSQL_ALLOW_NO_BACKUP=true)" >&2
    exec chaossql server start
  fi
  echo "[chaossql-cloud] LITESTREAM_BUCKET is not set; refusing to start without backups" >&2
  exit 1
fi
: "${LITESTREAM_ENDPOINT:?LITESTREAM_ENDPOINT is required with LITESTREAM_BUCKET}"
: "${LITESTREAM_ACCESS_KEY_ID:?LITESTREAM_ACCESS_KEY_ID is required with LITESTREAM_BUCKET}"
: "${LITESTREAM_SECRET_ACCESS_KEY:?LITESTREAM_SECRET_ACCESS_KEY is required with LITESTREAM_BUCKET}"

# On a fresh volume, recover the latest replica before serving traffic.
litestream restore -config /etc/litestream.yml -if-db-not-exists -if-replica-exists "$DB_PATH"

# Litestream supervises the server and forwards shutdown signals to it.
exec litestream replicate -config /etc/litestream.yml -exec "chaossql server start"
