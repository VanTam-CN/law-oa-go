#!/usr/bin/env bash
# I14 compatibility control: a fresh v14 database and pre-NP root, no published ports.
set -euo pipefail
umask 077
work=${NP_M5_WORK:?private M5 workspace required}
. "$work/env"
pg=$(cat "$work/pg")
net=$(cat "$work/network")
tag=$(basename "$work")
[[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$pg") == "$tag" ]]
[[ $(docker network inspect -f '{{.Internal}}' "$net") == true ]]
w=$(mktemp -d "$work/legacy-baseline.XXXXXXXX")
sha=d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7
git clone --quiet --shared --no-checkout "$(git rev-parse --show-toplevel)" "$w/src"
git -C "$w/src" checkout --quiet --detach "$sha"
case $(docker info --format '{{.Architecture}}') in
  aarch64|arm64) arch=arm64 ;;
  amd64|x86_64) arch=amd64 ;;
  *) exit 2 ;;
esac
(cd "$w/src" && CGO_ENABLED=0 GOOS=linux GOARCH="$arch" go build -trimpath -o "$w/root" .)
go version -m "$w/root" > "$w/build-info"
grep -Fqx "$(printf '\tbuild\tvcs.revision=%s' "$sha")" "$w/build-info"
grep -Fqx "$(printf '\tbuild\tvcs.modified=false')" "$w/build-info"
export DB_DATABASE="np_m5_legacy_$(openssl rand -hex 6)"
docker exec "$pg" createdb -U "$DB_USERNAME" -T template0 "$DB_DATABASE"
(cd "$w/src" && go run ./cmd/migrate -command bootstrap >"$w/bootstrap.log" 2>&1)
query=(docker exec -i "$pg" psql -XqAtw -v ON_ERROR_STOP=1 -U "$DB_USERNAME" -d "$DB_DATABASE")
guard="SELECT version || ':' || (to_regclass('case_intake_source_snapshots') IS NULL AND to_regclass('case_intake_write_requests') IS NULL)::text FROM schema_bootstrap_state WHERE id=1"
[[ $("${query[@]}" -c "$guard") == postgres-mvp-2026-08-27-v14:true ]]
export NP_BROWSER_PASSWORD=$(openssl rand -hex 20)
docker exec -i -e NP_BROWSER_PASSWORD "$pg" psql -XqAtw -v ON_ERROR_STOP=1 -U "$DB_USERNAME" -d "$DB_DATABASE" <<'SQL'
\getenv password NP_BROWSER_PASSWORD
INSERT INTO users(username,email,name,password,role,status)
VALUES ('legacy_m5','legacy@np.invalid','Synthetic legacy lawyer',crypt(:'password',gen_salt('bf')),'lawyer','active');
SQL
cleanup() {
  for file in "$w/app" "$w/redis"; do
    if [[ -s "$file" ]]; then
      id=$(cat "$file")
      [[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$id") == "$tag" ]] && docker rm -f "$id" >/dev/null
    fi
  done
}
trap cleanup EXIT
docker run -d --cidfile "$w/redis" --network "$net" --network-alias legacy-redis --label "law-oa-m5=$tag" redis:7-alpine redis-server --save '' --appendonly no >/dev/null
start_root() {
  docker run -d --cidfile "$w/app" --network "$net" --network-alias legacy-root --label "law-oa-m5=$tag" \
  --mount "type=bind,src=$1,dst=/fixture/root,readonly" --mount "type=bind,src=$work/bin/zoneinfo.zip,dst=/fixture/zoneinfo.zip,readonly" \
  -e ZONEINFO=/fixture/zoneinfo.zip -e ENVIRONMENT=production -e PORT=8080 -e DB_DRIVER=postgres -e DB_HOST=pg -e DB_PORT=5432 \
  -e DB_USERNAME -e DB_PASSWORD -e DB_DATABASE -e DB_SSLMODE=require -e REDIS_HOST=legacy-redis -e REDIS_PORT=6379 \
  -e JWT_SECRET -e APP_SECRET -e SUBJECT_DATA_KEY -e ONLYOFFICE_ENABLED=false -e CORS_ALLOWED_ORIGINS=https://m5.invalid \
  -e EXTERNAL_HEALTHCHECK_ENABLED=false alpine:3.20 /fixture/root >/dev/null
}
start_root "$w/root"
for ((i=0;i<40;i++)); do
  [[ $(docker inspect -f '{{.State.Running}}' "$(cat "$w/app")") == true ]]
  if docker exec "$(cat "$w/app")" wget -qO /dev/null http://127.0.0.1:8080/health/live 2>/dev/null; then break; fi
  sleep 1
done
docker exec -i -e NP_BROWSER_PASSWORD "$(cat "$work/forward")" node --input-type=module - <<'JS'
import assert from 'node:assert/strict'
const origin = 'http://legacy-root:8080/api/v1'
let token
async function call(method, path, body) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers.Authorization = 'Bearer ' + token
  const r = await fetch(origin + path, { method, headers, body: body ? JSON.stringify(body) : undefined })
  assert.ok([200,201].includes(r.status), method + ' HTTP ' + r.status)
  return (await r.json()).data
}
token = (await call('POST','/auth/login',{account:'legacy_m5',password:process.env.NP_BROWSER_PASSWORD})).token
assert.ok(token)
const draft = await call('POST','/case-intakes',{title:'M5 legacy',case_type:'civil',description:'first'})
assert.match(draft.id, /^[a-f0-9-]{36}$/)
await call('PUT','/case-intakes/'+draft.id,{title:'M5 legacy revised',description:'second'})
const got = await call('GET','/cases/intake-workbench/'+draft.id)
assert.equal(got.intake.description, 'second')
assert.equal(got.intake.title, 'M5 legacy revised')
JS
[[ $("${query[@]}" -c "$guard") == postgres-mvp-2026-08-27-v14:true ]]
[[ $("${query[@]}" -c "SELECT count(*) FROM case_intakes WHERE title='M5 legacy revised' AND description='second'") == 1 ]]
echo "M5_LEGACY_NATIVE_READ_WRITE_PASS 86_absent_before_after=true"
old=$(cat "$w/app")
[[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$old") == "$tag" ]]
docker rm -f "$old" >/dev/null
rm "$w/app"
(cd "$work/bin" && shasum -a 256 -c "$work/binary.sha256")
start_root "$work/bin/root"
candidate=$(cat "$w/app")
for ((i=0;i<25;i++)); do
  [[ $(docker inspect -f '{{.State.Running}}' "$candidate") == true ]] || break
  sleep 1
done
[[ $(docker inspect -f '{{.State.ExitCode}}' "$candidate") == 1 ]]
docker logs "$candidate" >"$w/schema-rejection.log" 2>&1
grep -Fq '接案来源或回执表结构与 000086 不一致' "$w/schema-rejection.log"
[[ $("${query[@]}" -c "$guard") == postgres-mvp-2026-08-27-v14:true ]]
echo M5_CANDIDATE_REJECTED_UNMIGRATED_SCHEMA
