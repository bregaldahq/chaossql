#!/bin/sh
# Operator commands for the managed image (`chaossql server ...`). Shell access
# such as `fly ssh console` runs as root; drop to the service user so SQLite
# never creates root-owned database, -wal, or -shm files.
set -eu
if [ "$(id -u)" = "0" ]; then
  exec su-exec chaossql chaossql server "$@"
fi
exec chaossql server "$@"
