#!/usr/bin/env bash
# Run only against a disposable container owned by this invocation.
set -euo pipefail
probe=${1:-normal}
case "$probe" in normal|missing-pg|skip|none) ;; *) exit 2 ;; esac
cd "$(git rev-parse --show-toplevel)"
tmp=$(mktemp -d)
cleanup() {
  if [[ -s "$tmp/cid" ]]; then docker rm -f "$(cat "$tmp/cid")" >/dev/null; fi
  rm -rf "$tmp"
}
trap cleanup EXIT
umask 077
export DB_DRIVER=postgres DB_HOST=127.0.0.1 DB_PORT=18790 DB_DATABASE=np_first_slice_dev DB_USERNAME=np_m1 DB_SSLMODE=disable
PGPASSWORD=$(openssl rand -hex 24)
export PGPASSWORD DB_PASSWORD="$PGPASSWORD" POSTGRES_PASSWORD="$PGPASSWORD" POSTGRES_USER="$DB_USERNAME" POSTGRES_DB="$DB_DATABASE"
export NP_TEST_DSN="host=$DB_HOST port=$DB_PORT dbname=$DB_DATABASE user=$DB_USERNAME sslmode=disable"
unset PGSERVICE PGHOSTADDR PGOPTIONS
psql_bin=${NP86_PSQL:-$(command -v psql || echo /usr/local/opt/libpq/bin/psql)}
[[ -x "$psql_bin" ]] || { echo M2_PSQL_MISSING; exit 1; }
if [[ "$probe" != missing-pg ]]; then
  docker run --rm -d --cidfile "$tmp/cid" --tmpfs /var/lib/postgresql/data -p 127.0.0.1:18790:5432 \
    -e POSTGRES_USER -e POSTGRES_DB -e POSTGRES_PASSWORD postgres:16.14 >/dev/null
  for ((try=0; try<30; try++)); do
    "$psql_bin" -XAt "$NP_TEST_DSN" -c 'select 1' >/dev/null 2>&1 && break
    sleep 1
  done
fi
identity=$("$psql_bin" -XAt "$NP_TEST_DSN" -c "select current_database() || ':' || current_user || ':' || current_setting('server_version_num')" 2>/dev/null) || { echo M2_PG_UNAVAILABLE; exit 1; }
[[ "$identity" == "np_first_slice_dev:np_m1:160014" && -s "$tmp/cid" ]] || { echo M2_WRONG_DATABASE; exit 1; }
go run ./cmd/migrate -command bootstrap >"$tmp/bootstrap" 2>&1 || { echo M2_BOOTSTRAP_FAILED; exit 1; }
bash scripts/np-first-slice/apply-main-np86.sh >"$tmp/schema" 2>&1 || { echo M2_SCHEMA_FAILED; exit 1; }
selector='^(TestNativeConcurrentDraftAndHistory|TestSchemaRequiresFrozenKeysAndTypes)$'
args=()
if [[ "$probe" == skip ]]; then
  # Only the temporary compiler overlay skips the critical test.
  /usr/bin/python3 - "$tmp" <<'PY'
import json, pathlib, sys
tmp = pathlib.Path(sys.argv[1]).resolve()
src = pathlib.Path("tests/integration/m1/native_pg_test.go").resolve()
signature = "func TestNativeConcurrentDraftAndHistory(t *testing.T) {"
body = src.read_text()
assert body.count(signature) == 1
overlay = tmp / "skip.go"
overlay.write_text(body.replace(signature, signature + '\n\tt.Skip("M2 synthetic skip")'))
(tmp / "overlay.json").write_text(json.dumps({"Replace": {str(src): str(overlay)}}))
PY
  args+=("-overlay=$tmp/overlay.json")
elif [[ "$probe" == none ]]; then
  selector='^TestM2DeliberatelyAbsent$'
fi
rc=0
go test -json -race -count=1 -tags np_m1_pg -run "$selector" ${args[@]+"${args[@]}"} ./tests/integration/m1 >"$tmp/events" 2>"$tmp/stderr" || rc=$?
/usr/bin/python3 scripts/np-ci/check-go-events.py "$tmp/events" "$rc"
