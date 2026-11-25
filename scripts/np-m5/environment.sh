#!/usr/bin/env bash
# Isolated root binary. Keep credentials and raw logs in the caller's private directory.
set -euo pipefail
umask 077
work=${NP_M5_WORK:?set new private temporary directory}
[[ -d "$work" && -O "$work" && ! -L "$work" && $(basename "$work") == np-m5.* ]] || exit 2
[[ $(ls -ld "$work" | cut -c2-10) == rwx------ ]] || { echo "Private mode 0700 required" >&2; exit 2; }
root=$(git rev-parse --show-toplevel)
cd "$root"
pg_image=postgres:16.14
redis_image=redis:7-alpine
runtime_image=alpine:3.20
label="law-oa-m5=$(basename "$work")"
network="law-oa-$(basename "$work" | tr '[:upper:].' '[:lower:]-')"
remove_owned() {
  local kind=$1 file=$2 id tag
  [[ -s "$file" ]] || return 0
  id=$(cat "$file")
  if [[ "$kind" == network ]]; then
    tag=$(docker network inspect -f '{{index .Labels "law-oa-m5"}}' "$id")
    [[ "$tag" == "$(basename "$work")" ]] || return 1
    docker network rm "$id" >/dev/null
  else
    tag=$(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$id")
    [[ "$tag" == "$(basename "$work")" ]] || return 1
    docker rm -f "$id" >/dev/null
  fi
  rm "$file"
}
case ${1:-} in
  init)
    [[ ! -e "$work/env" && ! -e "$work/network" ]] || exit 2
    docker network create --internal --label "$label" "$network" >"$work/network"
    docker network create --label "$label" "$network-ingress" >"$work/ingress"
    # Docker Desktop suppresses published ports on an internal-only network.
    # Only the forwarding container has an ingress interface; root, PG and Redis have no egress.
    docker run -d --cidfile "$work/forward" --label "$label" --network "$network-ingress" \
      -p 127.0.0.1:18789:8080 -p 127.0.0.1:18790:5432 node:18-alpine node -e '
        const net = require("node:net");
        for (const [port, host] of [[8080,"root"],[5432,"pg"]]) {
          net.createServer(client => {
            const upstream = net.connect(port, host);
            client.on("error", () => upstream.destroy());
            upstream.on("error", () => client.destroy());
            client.on("close", () => upstream.destroy());
            upstream.on("close", () => client.destroy());
            client.pipe(upstream).pipe(client);
          }).listen(port);
        }' >/dev/null
    docker network connect "$network" "$(cat "$work/forward")"
    export POSTGRES_DB=np_first_slice_dev POSTGRES_USER=np_m1
    POSTGRES_PASSWORD=$(openssl rand -hex 24); export POSTGRES_PASSWORD
    docker run -d --cidfile "$work/pg" --label "$label" --network "$network" --network-alias pg \
      --tmpfs /var/lib/postgresql/data \
      -e POSTGRES_USER -e POSTGRES_DB -e POSTGRES_PASSWORD "$pg_image" >/dev/null
    docker run -d --cidfile "$work/redis" --label "$label" --network "$network" --network-alias redis \
      "$redis_image" redis-server --save '' --appendonly no >/dev/null
    export DB_USERNAME="$POSTGRES_USER" DB_DATABASE="$POSTGRES_DB" DB_PASSWORD="$POSTGRES_PASSWORD" PGPASSWORD="$POSTGRES_PASSWORD"
    export DB_DRIVER=postgres DB_HOST=127.0.0.1 DB_PORT=18790 DB_SSLMODE=disable
    export NP_LOCAL_CHECK=synthetic-only NP_TEST_DSN='host=127.0.0.1 port=18790 dbname=np_first_slice_dev user=np_m1 sslmode=disable'
    export NP_M5_WORK="$work" NP_REDIS_CID="$(cat "$work/redis")"
    NP_BROWSER_PASSWORD=$(openssl rand -hex 20); JWT_SECRET=$(openssl rand -hex 32)
    APP_SECRET=$(openssl rand -hex 32); SUBJECT_DATA_KEY=$(openssl rand -base64 32)
    export JWT_SECRET APP_SECRET SUBJECT_DATA_KEY NP_BROWSER_PASSWORD
    for key in DB_DRIVER DB_HOST DB_PORT DB_SSLMODE DB_USERNAME DB_DATABASE DB_PASSWORD PGPASSWORD NP_LOCAL_CHECK NP_TEST_DSN NP_BROWSER_PASSWORD JWT_SECRET APP_SECRET SUBJECT_DATA_KEY NP_M5_WORK NP_REDIS_CID; do
      printf 'export %s=%q\n' "$key" "${!key}"
    done >"$work/env"
    psql_bin=${NP86_PSQL:-$(command -v psql || echo /usr/local/opt/libpq/bin/psql)}
    for ((i=0;i<60;i++)); do
      if "$psql_bin" -XwAt "$NP_TEST_DSN" -c 'SELECT 1' >/dev/null 2>&1; then break; fi
      sleep 1
    done
    [[ $("$psql_bin" -XwAt "$NP_TEST_DSN" -c "select current_database() || ':' || current_user || ':' || current_setting('server_version_num')") == np_first_slice_dev:np_m1:160014 ]]
    [[ $(docker exec "$NP_REDIS_CID" redis-cli ping) == PONG ]]
    go run ./cmd/migrate -command bootstrap >"$work/bootstrap.log" 2>&1
    bash scripts/np-first-slice/apply-main-np86.sh >"$work/schema.log" 2>&1
    case $(docker info --format '{{.Architecture}}') in
      aarch64|arm64) arch=arm64 ;;
      x86_64|amd64) arch=amd64 ;;
      *) echo "Unsupported Docker CPU" >&2; exit 2 ;;
    esac
    mkdir -p "$work/bin"
    CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -o "$work/bin/root" .
    ;;
  root)
    . "$work/env"
    : "${NP_ASSISTANT_ACTORS?specify an actor list or an explicitly empty value}"
    [[ -x "$work/bin/root" && -s "$work/network" ]] || exit 2
    pg=$(cat "$work/pg")
    [[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$pg") == "$(basename "$work")" ]]
    if [[ $(docker exec "$pg" psql -U "$DB_USERNAME" -d "$DB_DATABASE" -Atc 'SHOW ssl') == off ]]; then
      docker exec -u postgres "$pg" sh -ec 'cd "$PGDATA"; umask 077
        openssl req -x509 -newkey rsa:2048 -nodes -days 2 -subj /CN=pg -keyout server.key -out server.crt >/dev/null 2>&1
        psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -c "ALTER SYSTEM SET ssl=on" -c "SELECT pg_reload_conf()" >/dev/null'
    fi
    [[ $(docker exec "$pg" psql -U "$DB_USERNAME" -d "$DB_DATABASE" -Atc 'SHOW ssl') == on ]]
    remove_owned container "$work/app"
    cp "$(go env GOROOT)/lib/time/zoneinfo.zip" "$work/bin/zoneinfo.zip"
    docker run -d --cidfile "$work/app" --label "$label" --network "$network" \
      --network-alias root --mount "type=bind,src=$work/bin/root,dst=/fixture/root,readonly" \
      --mount "type=bind,src=$work/bin/zoneinfo.zip,dst=/fixture/zoneinfo.zip,readonly" -e ZONEINFO=/fixture/zoneinfo.zip -e ENVIRONMENT=production -e PORT=8080 -e DB_DRIVER -e DB_HOST=pg -e DB_PORT=5432 \
      -e DB_USERNAME -e DB_DATABASE -e DB_PASSWORD -e DB_SSLMODE=require -e REDIS_HOST=redis -e REDIS_PORT=6379 \
      -e JWT_SECRET -e APP_SECRET -e SUBJECT_DATA_KEY -e NP_ASSISTANT_ACTORS -e CORS_ALLOWED_ORIGINS=https://m5.invalid \
      -e ONLYOFFICE_ENABLED=false -e EXTERNAL_HEALTHCHECK_ENABLED=false "$runtime_image" /fixture/root >/dev/null
    for ((i=0;i<60;i++)); do
      if [[ $(docker inspect -f '{{.State.Running}}' "$(cat "$work/app")") != true ]]; then echo M5_ROOT_EXITED >&2; exit 1; fi
      if curl -fs --noproxy '*' --max-time 2 http://127.0.0.1:18789/health/live >/dev/null; then echo M5_ROOT_LIVE; exit 0; fi
      sleep 1
    done
    exit 1 ;;
  clean)
    for name in app forward redis pg; do remove_owned container "$work/$name"; done
    remove_owned network "$work/network"
    remove_owned network "$work/ingress" ;;
  *) echo 'expected init, root or clean' >&2; exit 2 ;;
esac
