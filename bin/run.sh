#!/bin/sh
set -eu

echo "[run.sh] Starting service"

: "${DATABASE_URL:?DATABASE_URL must be set}"

set -- /app/db/migrations/*.sql
if [ -f "$1" ]; then
    echo "[run.sh] Running DB migrations"
    goose -dir /app/db/migrations postgres "$DATABASE_URL" up
else
    echo "[run.sh] No SQL migrations found; skipping migrations"
fi

echo "[run.sh] Starting Go app"
exec /app/bin/app
