#!/bin/sh
# Restore drill: restore the latest replica into a scratch file and list the
# tenants it contains. It never starts the server or Litestream replication,
# so it cannot write to the replica it reads.
set -eu
: "${LITESTREAM_BUCKET:?LITESTREAM_BUCKET is required}"
export LITESTREAM_PATH="${LITESTREAM_PATH:-chaossql-cloud}"
target="$(mktemp -d)/restore-check.db"
litestream restore -config /etc/litestream.yml -o "$target" "$DB_PATH"
chaossql server org list --db "$target"
echo "[chaossql-cloud] Restore check succeeded from ${LITESTREAM_BUCKET}/${LITESTREAM_PATH}"
