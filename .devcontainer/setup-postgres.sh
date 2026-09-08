#!/usr/bin/env bash
# Initialize (once) and start a local PostgreSQL 17 cluster owned by the "node"
# user. Runs inside the dev container itself, so integration tests that connect
# to localhost:5432 work without Docker-in-Docker.
set -euo pipefail

export PATH="/usr/lib/postgresql/17/bin:${PATH}"
PGDATA="${PGDATA:-/home/node/pgdata}"

if [ ! -s "${PGDATA}/PG_VERSION" ]; then
  echo "Initializing PostgreSQL cluster at ${PGDATA}..."
  pwfile="$(mktemp)"
  printf 'easi' > "${pwfile}"
  initdb -D "${PGDATA}" -U easi \
    --auth-local=trust \
    --auth-host=scram-sha-256 \
    --pwfile="${pwfile}" >/dev/null
  rm -f "${pwfile}"
  {
    echo "listen_addresses = 'localhost'"
    echo "port = 5432"
  } >> "${PGDATA}/postgresql.conf"
fi

if pg_ctl -D "${PGDATA}" status >/dev/null 2>&1; then
  echo "PostgreSQL already running."
else
  echo "Starting PostgreSQL..."
  pg_ctl -D "${PGDATA}" -l "${PGDATA}/logfile" -w start
fi
