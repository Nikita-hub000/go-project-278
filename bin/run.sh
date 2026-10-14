#!/usr/bin/env bash
set -euo pipefail

echo "[run.sh] Starting service"

: "${DATABASE_URL:?DATABASE_URL must be set}"

echo "[run.sh] Running DB migrations"
goose -dir ./db/migrations postgres "$DATABASE_URL" up

echo "[run.sh] Starting Caddy"
caddy run --config /etc/caddy/Caddyfile &

echo "[run.sh] Starting Go app"
PORT=8080 exec /app/bin/app
