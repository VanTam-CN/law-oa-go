#!/usr/bin/env bash
# One fresh synthetic actor and comparison window per suite. Ports belong to this fixture.
set -euo pipefail
umask 077
suite=${1:?select route-b, storage, navigation, editing, tabs, legacy or prod}
root=$(cd "$(dirname "$0")/../.." && pwd)
cd "$root"
config=playwright.np-m4.config.ts
spec=()
case "$suite" in
  route-b) : "${NP_TEST_DSN:?route-b requires isolated PostgreSQL}"; config=playwright.np.config.ts ;;
  prod) config=np-entry/c10-sw/playwright.c10-sw.config.ts ;;
  storage|navigation|editing|tabs|legacy) spec=("$suite.spec.ts") ;;
  *) echo "Unknown M4 suite" >&2; exit 2 ;;
esac
[[ ${NP_LOCAL_CHECK:-} == synthetic-only && ${DB_HOST:-} == 127.0.0.1 && ${DB_PORT:-} == 18790 &&
   ${DB_DRIVER:-} == postgres && ${DB_DATABASE:-} == np_first_slice_dev && ${DB_USERNAME:-} == np_m1 ]] || {
  echo "Dedicated loopback v14+86 PostgreSQL required" >&2; exit 2
}
export VITE_API_PROXY_TARGET=http://127.0.0.1:18789 VITE_NP_MEMORY_ONLY=true
work=$(mktemp -d "${TMPDIR:-/tmp}/np-m4-run.XXXXXXXX")
export NP_BROWSER_SEED="$work/seed.json" NP_BROWSER_OTHER_SEED="$work/other.json"
server_pid=
cleanup() {
  if [[ -n "$server_pid" ]]; then kill "$server_pid" 2>/dev/null || true; wait "$server_pid" 2>/dev/null || true; fi
  rm -rf -- "$work"
}
trap cleanup EXIT
python_bin=${PYTHON:-python3}
compare=("$python_bin" scripts/np-first-slice/compare-main-browser-db.py)
"${compare[@]}" seed "$NP_BROWSER_SEED"
"${compare[@]}" seed "$NP_BROWSER_OTHER_SEED"
"${compare[@]}" before "$NP_BROWSER_SEED" "$work/before.json"
go build -tags np_local_check -o "$work/server" ./cmd/np-local-check
"$work/server" >"$work/backend.log" 2>&1 & server_pid=$!
ready=false
for _ in {1..100}; do
  kill -0 "$server_pid" || exit 1
  if curl --noproxy '*' -s -o /dev/null --max-time 1 http://127.0.0.1:18789/; then ready=true; break; fi
  sleep 0.1
done
[[ $ready == true ]] && kill -0 "$server_pid" || { echo "Isolated server unavailable" >&2; exit 1; }
(cd frontend && node node_modules/@playwright/test/cli.js test -c "$config" ${spec[@]+"${spec[@]}"})
case "$suite" in
  route-b)
    "${compare[@]}" after "$NP_BROWSER_SEED" "$work/before.json"
    NP_GO_SHA_MANIFEST="$work/before.json" go test -count=1 ./internal/handlers -run '^TestNPGoDigestMatchesBrowserManifest$' -v >"$work/digest.log"
    grep -q '^--- PASS: TestNPGoDigestMatchesBrowserManifest ' "$work/digest.log" || { echo "Go digest missing or skipped" >&2; exit 1; }
    echo GO_BROWSER_DIGEST_MATCH ;;
  storage|navigation|prod)
    "${compare[@]}" unchanged "$NP_BROWSER_SEED" "$work/before.json" ;;
esac
