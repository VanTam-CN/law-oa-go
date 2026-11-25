#!/usr/bin/env bash
# Isolate history-table fault injection from the browser, access and detection windows.
set -euo pipefail
umask 077
work=${NP_M5_WORK:?}
. "$work/env"
w="$work/history-window"
pg=$(cat "$work/pg")
net=$(cat "$work/network")
tag=$(basename "$work")
[[ "$NP_LOCAL_CHECK" == synthetic-only && "$DB_DATABASE" == np_first_slice_dev ]]
[[ $(docker network inspect -f '{{.Internal}}' "$net") == true ]]
owned() { [[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$1") == "$tag" ]]; }
owned "$pg"
remove_fixture() {
  if [[ -s "$w/$1" ]]; then
    local id
    id=$(cat "$w/$1")
    owned "$id" && docker rm -f "$id" >/dev/null && rm "$w/$1"
  fi
}
case ${1:-} in
  init)
    mkdir "$w"
    printf 'np_m5_history_%s\n' "$(openssl rand -hex 6)" >"$w/db"
    db=$(cat "$w/db")
    docker exec "$pg" createdb -U "$DB_USERNAME" -T template0 "$db"
    docker exec "$pg" pg_dump -Fc -U "$DB_USERNAME" -d "$DB_DATABASE" |
      docker exec -i "$pg" pg_restore --exit-on-error -U "$DB_USERNAME" -d "$db"
    docker run -d --cidfile "$w/redis" --label "law-oa-m5=$tag" --network "$net" --network-alias history-redis redis:7-alpine redis-server --save '' --appendonly no >/dev/null
    docker run -d --cidfile "$w/forward" --label "law-oa-m5=$tag" --network "$(cat "$work/ingress")" -p 127.0.0.1:18791:8080 node:18-alpine node -e '
      const net=require("node:net");
      net.createServer(c=>{const u=net.connect(8080,"history-root");
        c.on("error",()=>u.destroy()); u.on("error",()=>c.destroy());
        c.on("close",()=>u.destroy()); u.on("close",()=>c.destroy());
        c.pipe(u).pipe(c);}).listen(8080);' >/dev/null
    docker network connect "$net" "$(cat "$w/forward")"
    ;;
  root)
    : "${NP_ASSISTANT_ACTORS?set synthetic allowlist explicitly}"
    DB_DATABASE=$(cat "$w/db")
    [[ "$DB_DATABASE" =~ ^np_m5_history_[0-9a-f]{12}$ ]]
    remove_fixture app
    docker run -d --cidfile "$w/app" --label "law-oa-m5=$tag" --network "$net" --network-alias history-root \
      --mount "type=bind,src=$work/bin/root,dst=/fixture/root,readonly" --mount "type=bind,src=$work/bin/zoneinfo.zip,dst=/fixture/zoneinfo.zip,readonly" \
      -e ZONEINFO=/fixture/zoneinfo.zip -e ENVIRONMENT=production -e PORT=8080 -e DB_DRIVER=postgres -e DB_HOST=pg -e DB_PORT=5432 \
      -e DB_USERNAME -e DB_PASSWORD -e DB_DATABASE -e DB_SSLMODE=require -e REDIS_HOST=history-redis -e REDIS_PORT=6379 \
      -e JWT_SECRET -e APP_SECRET -e SUBJECT_DATA_KEY -e NP_ASSISTANT_ACTORS -e ONLYOFFICE_ENABLED=false -e CORS_ALLOWED_ORIGINS=https://m5.invalid \
      -e EXTERNAL_HEALTHCHECK_ENABLED=false alpine:3.20 /fixture/root >/dev/null
    for ((i=0;i<40;i++)); do
      if curl -fs --noproxy '*' --max-time 1 http://127.0.0.1:18791/health/live >/dev/null; then
        observed=$(docker exec "$(cat "$w/app")" sha256sum /fixture/root)
        [[ "${observed%% *}" == "$(cut -d ' ' -f1 "$work/binary.sha256")" ]]
        echo M5_HISTORY_ROOT_LIVE
        exit 0
      fi
      sleep 1
    done
    exit 1 ;;
  clean) for name in app forward redis; do remove_fixture "$name"; done ;;
  *) exit 2 ;;
esac
