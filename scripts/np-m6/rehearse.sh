#!/usr/bin/env bash
# Run serially on the exclusive M5 loopback ports with a new private np-m5.* window.
set -euo pipefail
umask 077
cd "$(git rev-parse --show-toplevel)"
work=${NP_M5_WORK:?new private directory required}
[[ -d "$work" && -O "$work" && ! -L "$work" && $(basename "$work") == np-m5.* ]] || exit 2
m5=scripts/np-m5/environment.sh
candidate=49e5c153c01fd4d21abea99f2792eea279628a5b
protected=89f2eb9eace9a2f72f31375a83cc34d97289c4dc
legacy=d9c92fbd583a0cbbf7f9b4c071e1f2329981f4d7
export PYTHONDONTWRITEBYTECODE=1
py=${PYTHON:-/usr/bin/python3}
archive_build() {
  mkdir "$work/$1"
  cp "$work/bin/root" "$work/binary.sha256" "$work/build-info" "$work/$1/"
}
if [[ "${1:-}" == prepare ]]; then
  [[ ! -e "$work/env" ]]
  NP_M5_BUILD_REVISION="$candidate" bash "$m5" init-v14
  archive_build candidate
  # Use the M5 acceptance binary exactly, including the clean Go VCS stamp.
  [[ $(cut -d' ' -f1 "$work/candidate/binary.sha256") == 213bb27532851e03bcc17ee630665e6fd2edd914cc1b711b477269decfae48d0 ]]
  NP_M5_BUILD_REVISION="$protected" bash "$m5" build
  archive_build rollback
  NP_M5_BUILD_REVISION="$legacy" bash "$m5" build
  archive_build legacy
  exit 0
fi
[[ "${1:-}" == exercise && ! -e "$work/roots.txt" ]] || exit 2
. "$work/env"
export NP_BROWSER_SEED="$work/seed.json" NP_M6_EXPECT_ENABLED=true
pg=$(cat "$work/pg")
tag=$(basename "$work")
[[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$pg") == "$tag" ]]
[[ $(docker network inspect -f '{{.Internal}}' "$(cat "$work/network")") == true ]]
[[ $(docker port "$(cat "$work/forward")" 8080/tcp) == 127.0.0.1:18789 ]]
[[ $(docker port "$(cat "$work/forward")" 5432/tcp) == 127.0.0.1:18790 ]]
sql() { docker exec "$pg" psql -XqAtw -v ON_ERROR_STOP=1 -U np_m1 -d np_first_slice_dev -c "$1"; }
[[ $(sql "SELECT current_database()||':'||current_user||':'||version FROM schema_bootstrap_state WHERE id=1") == np_first_slice_dev:np_m1:postgres-mvp-2026-08-27-v14 ]]
src="$work/source-$candidate"
apply() { bash "$src/scripts/np-first-slice/apply-main-np86.sh"; }
baseline_tables() {
  sql "SELECT encode(sha256(convert_to(string_agg(relname||':'||oid::text,',' ORDER BY relname),'UTF8')),'hex') FROM pg_class WHERE relnamespace='public'::regnamespace AND relkind IN ('r','p') AND relname NOT IN ('case_intake_source_snapshots','case_intake_write_requests')"
}
objects() { sql "SELECT coalesce(to_regclass('case_intake_source_snapshots')::oid::text,'none')||':'||coalesce(to_regclass('case_intake_write_requests')::oid::text,'none')"; }
topology() {
  local root_id=${1:-} actual expected
  actual=$(docker network inspect -f '{{range $id,$v := .Containers}}{{println $id}}{{end}}' "$(cat "$work/network")" | sed '/^$/d' | sort)
  expected=$({ for f in pg redis forward; do echo "$(cat "$work/$f")"; done; if [[ -n "$root_id" ]]; then echo "$root_id"; fi; } | sort)
  [[ "$actual" == "$expected" ]]
  if [[ -n "$root_id" ]]; then
    [[ -z $(docker port "$root_id") ]]
    [[ $(docker inspect -f '{{.HostConfig.RestartPolicy.Name}}' "$root_id") == no ]]
  fi
}
stamp() {
  local stage=$1 version=$2 id=$3 actual expected
  [[ $(docker inspect -f '{{.Path}}' "$id") == /fixture/root ]]
  actual=$(docker exec "$id" sha256sum /fixture/root | cut -d' ' -f1)
  expected=$(cut -d' ' -f1 "$work/binary.sha256")
  [[ "$actual" == "$expected" ]]
  printf '%s %s %s\n' "$stage" "$version" "$actual" >>"$work/runtime-stamps"
}
start() {
  local kind=$1 roles=$2 version
  case "$kind" in
    legacy) [[ $(objects) == none:none ]]; version="$legacy" ;;
    candidate) version="$candidate" ;;
    rollback) version="$protected" ;;
    *) return 2 ;;
  esac
  grep -Fqx "$(printf '\tbuild\tvcs.revision=%s' "$version")" "$work/$kind/build-info"
  cp "$work/$kind/root" "$work/bin/selected"
  mv "$work/bin/selected" "$work/bin/root"
  cp "$work/$kind/binary.sha256" "$work/$kind/build-info" "$work/"
  NP_ASSISTANT_ACTORS="$roles" bash "$m5" root >/dev/null
  topology "$(cat "$work/app")"
  stamp "$kind" "$version" "$(cat "$work/app")"
  echo "$kind $(cat "$work/app")" >>"$work/roots.txt"
}
stop() {
  local id
  id=$(cat "$work/app")
  [[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$id") == "$tag" ]]
  docker stop -t 15 "$id" >/dev/null
  [[ $(docker inspect -f '{{.State.Running}}' "$id") == false ]]
  if curl -sS --noproxy '*' --max-time 2 -o /dev/null http://127.0.0.1:18789/api/v1/case-intake-assist/capabilities 2>/dev/null; then
    echo 'HTTP still available after stopping root' >&2; return 1
  fi
  echo 'GAP no root HTTP response' >>"$work/steps"
}
# Reconnect with --since so a slow subscriber cannot miss the first start.
docker events --since "$(date -u +%Y-%m-%dT%H:%M:%SZ)" --filter type=container --filter "label=law-oa-m5=$tag" \
  --filter event=start --filter event=die --format '{{json .}}' >"$work/events.jsonl" &
monitor=$!
trap 'kill "$monitor" 2>/dev/null || true; wait "$monitor" 2>/dev/null || true' EXIT
topology
: >"$work/roots.txt"
start legacy ''
[[ $(curl -sS --noproxy '*' -o /dev/null -w '%{http_code}' http://127.0.0.1:18789/api/v1/case-intake-assist/capabilities) == 404 ]]
"$py" scripts/np-m6/boundary.py legacy-create
stop
baseline=$(baseline_tables)
# Deliberately incomplete schema in this owned DB must fail the bounded append.
sql 'CREATE TABLE case_intake_source_snapshots(m6_incomplete boolean)' >/dev/null
partial=$(objects)
if apply >"$work/partial.log" 2>&1; then exit 1; fi
grep -Fq 'partial 000086 state' "$work/partial.log"
[[ $(objects) == "$partial" ]]
sql 'ALTER TABLE case_intake_source_snapshots RENAME TO m6_synthetic_incomplete; DROP TABLE m6_synthetic_incomplete' >/dev/null
[[ $(objects) == none:none ]]
# Exercise a real FK privilege failure, repairing only this disposable cluster.
admin() { docker exec "$pg" psql -XqAtw -v ON_ERROR_STOP=1 -U np_m6_bootstrap -d np_first_slice_dev -c "$1" >/dev/null; }
[[ $(sql "SELECT rolsuper FROM pg_roles WHERE rolname=current_user") == f ]]
admin 'ALTER TABLE case_intakes OWNER TO np_m6_bootstrap'
if apply >"$work/permission.log" 2>&1; then admin 'ALTER TABLE case_intakes OWNER TO np_m1'; exit 1; fi
admin 'ALTER TABLE case_intakes OWNER TO np_m1'
grep -Fq 'permission denied for table case_intakes' "$work/permission.log"
[[ $(objects) == none:none ]]
apply >"$work/append.log" 2>&1
grep -Fq 'first apply: appending frozen 000086' "$work/append.log"
grep -Fq 'PASS: NP schema verified' "$work/append.log"
oids=$(objects)
[[ "$oids" =~ ^[0-9]+:[0-9]+$ ]]
sql 'ALTER TABLE case_intake_write_requests RENAME COLUMN request_digest TO m6_parked_digest' >/dev/null
if apply >"$work/mismatch.log" 2>&1; then exit 1; fi
grep -Fq 'shared catalog check returned not-ready' "$work/mismatch.log"
[[ $(objects) == "$oids" ]]
sql 'ALTER TABLE case_intake_write_requests RENAME COLUMN m6_parked_digest TO request_digest' >/dev/null
"$py" scripts/np-first-slice/compare-main-browser-db.py seed "$NP_BROWSER_SEED" >/dev/null
actor=$("$py" -c 'import json,sys;print(json.load(open(sys.argv[1]))["actor_id"])' "$NP_BROWSER_SEED")
target=$("$py" -c 'import json,sys,uuid;print(uuid.UUID(json.load(open(sys.argv[1]))["target"]))' "$NP_BROWSER_SEED")
snapshot() {
  docker exec -i "$pg" psql -XqAtw -v ON_ERROR_STOP=1 -U np_m1 -d np_first_slice_dev -v target="$target" \
    <scripts/np-m6/snapshot.sql >"$work/$1.pg.json"
  if [[ "$1" != seeded ]]; then cmp -s "$work/seeded.pg.json" "$work/$1.pg.json"; fi
}
start candidate "$actor"
"$py" scripts/np-m6/boundary.py legacy-verify
"$py" scripts/np-m6/ledger.py seed >"$work/seed-result.json"
snapshot seeded
"$py" scripts/np-m6/boundary.py open
apply >"$work/repeat.log" 2>&1
grep -Fq 'repeat: verify-only mode' "$work/repeat.log"
grep -Fq 'PASS: NP schema verified' "$work/repeat.log"
[[ $(objects) == "$oids" ]]
"$py" scripts/np-m6/ledger.py verify >"$work/repeat-result.json"
snapshot repeated
stop
id=$(cat "$work/app")
docker start "$id" >/dev/null
for ((i=0;i<40;i++)); do curl -fs --noproxy '*' --max-time 1 http://127.0.0.1:18789/health/live >/dev/null && break; sleep 1; done
[[ $(docker inspect -f '{{.State.Running}}' "$id") == true ]]
topology "$id"
stamp restart "$candidate" "$id"
echo "restart $id" >>"$work/roots.txt"
"$py" scripts/np-m6/boundary.py restarted
"$py" scripts/np-m6/ledger.py verify >"$work/restart-result.json"
snapshot restarted
stop
start rollback ''
"$py" scripts/np-m6/boundary.py disabled
NP_M6_EXPECT_ENABLED=false "$py" scripts/np-m6/ledger.py verify >"$work/rollback-result.json"
snapshot rolled-back
"$py" scripts/np-m6/boundary.py legacy-final
[[ $(objects) == "$oids" && $(wc -l <"$work/steps" | tr -d ' ') == 3 ]]
stop
for ((i=0;i<50;i++)); do
  [[ $(grep -c '"die"' "$work/events.jsonl" || true) -ge 4 ]] && break
  sleep .1
done
"$py" scripts/np-m6/switch-check.py >"$work/switch-result.json"
topology
[[ $(baseline_tables) == "$baseline" ]]
printf '%s\n' "$baseline" >"$work/baseline-tables.sha256"
printf '%s\n' "$oids" >"$work/verified-table-oids"
echo M6_I12_REHEARSAL_PASS | tee "$work/passed"
