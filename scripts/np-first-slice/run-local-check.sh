#!/bin/sh
# Synthetic loopback runner only; never starts or updates the trial deployment.
set -eu
cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
ENV_FILE="${1:-./.env}"
set -a
. "$ENV_FILE"
set +a
: "${NP_FIXTURE_PASSWORD:?set the original synthetic fixture password in private .env}"
sh ./probe-db.sh "$ENV_FILE"
mkdir -p .bin
(cd "$ROOT/backend" && go build -tags np_local_check -o "$ROOT/scripts/np-first-slice/.bin/np-local-check" ./cmd/np-local-check)
export DB_DRIVER=postgres DB_HOST=127.0.0.1 DB_PORT=18790
export DB_USERNAME="$POSTGRES_USER" DB_PASSWORD="$POSTGRES_PASSWORD" DB_DATABASE="$POSTGRES_DB"
export ENVIRONMENT=development NP_LOCAL_CHECK=synthetic-only
export JWT_SECRET="$(openssl rand -hex 32)" SUBJECT_DATA_KEY="$(openssl rand -hex 16)"
cd "$ROOT/backend"
exec "$ROOT/scripts/np-first-slice/.bin/np-local-check"
