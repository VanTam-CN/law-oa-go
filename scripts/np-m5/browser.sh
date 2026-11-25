#!/usr/bin/env bash
# Run a named suite against a separately booted M5 root. Seeds and all raw output stay private.
set -euo pipefail
umask 077
suite=${1:?suite required}
window=${2:?new directory below NP_M5_WORK required}
: "${NP_BROWSER_SEED:?}" "${NP_BROWSER_OTHER_SEED:?}" "${NP_BROWSER_PASSWORD:?}" "${NP_M5_WORK:?}"
[[ "$window" == "$NP_M5_WORK/"* && ! -e "$window" && ${NP_LOCAL_CHECK:-} == synthetic-only ]] || exit 2
[[ $(docker inspect -f '{{index .Config.Labels "law-oa-m5"}}' "$(cat "$NP_M5_WORK/app")") == "$(basename "$NP_M5_WORK")" ]]
root=$(git rev-parse --show-toplevel)
cd "$root"
mkdir "$window"
py=${PYTHON:-/usr/bin/python3}
spec="np-m4/$suite.spec.ts"
config=playwright.m5.config.ts
case "$suite" in
  route-b) spec=e2e/np-first-slice.live.spec.ts ;;
  selection) spec=np-entry/first-selection.spec.ts ;;
  prod) config=np-entry/c10-sw/playwright.c10-sw.config.ts; spec=c10-sw.spec.ts ;;
  navigation|storage|tabs|editing|legacy) ;;
  *) echo "Unknown M5 suite" >&2; exit 2 ;;
esac
export ROOT_URL=http://127.0.0.1:18789 VITE_API_PROXY_TARGET=http://127.0.0.1:18789 VITE_NP_MEMORY_ONLY=true
unset NP_M5_PROXY_PORT
export NP_M5_PROXY_EVENT_FILE="$window/events.jsonl"
sha=$(docker exec "$(cat "$NP_M5_WORK/app")" sha256sum /fixture/root)
[[ "${sha%% *}" == "$(cut -d' ' -f1 "$NP_M5_WORK/binary.sha256")" ]]
"$py" scripts/np-first-slice/compare-main-browser-db.py before "$NP_BROWSER_SEED" "$window/before.json"
"$py" - "$suite" >"$window/names" <<'PY'
import json, re, sys
with open("scripts/np-ci/i11-browser-cases.json") as f:
    for name in json.load(f)[sys.argv[1]]:
        print(re.escape(name) + "$")
PY
n=0
while IFS= read -r name; do
  # Each reopened page performs several API reads. Keep production throttling active.
  if [[ "$n" -gt 0 ]]; then sleep 55; fi
  n=$((n+1))
  export PLAYWRIGHT_JSON_OUTPUT_FILE="$window/case-$n.json"
  (cd frontend && node node_modules/@playwright/test/cli.js test -c "$config" "$spec" \
      -g "$name" --reporter=json </dev/null >"$window/case-$n.log" 2>&1)
done <"$window/names"
[[ $(docker exec "$(cat "$NP_M5_WORK/app")" sha256sum /fixture/root) == "$sha" ]]
"$py" scripts/np-m5/check-browser.py "$suite" "$NP_BROWSER_SEED" "$window/before.json" "$window"/case-*.json >"$window/summary.json"
cat "$window/summary.json"
