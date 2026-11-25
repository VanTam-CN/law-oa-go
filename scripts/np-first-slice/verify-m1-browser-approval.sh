#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/../.."
fixture_dir=$(mktemp -d)
trap 'rm -rf "$fixture_dir"' EXIT
NP_M1_FIXTURE_DIR="$fixture_dir" go test -count=1 -tags np_m1_pg,np_m1_browser_pg \
    ./tests/integration/m1 -run '^TestM1R4WorkbenchApprovalBrowser$' >"$fixture_dir/router.log" 2>&1 &
fixture_pid=$!
for ((i=0; i<1200; i++)); do
    [[ -f "$fixture_dir/ready.json" ]] && break
    kill -0 "$fixture_pid" 2>/dev/null || break
    sleep 0.1
done
if [[ ! -f "$fixture_dir/ready.json" ]]; then
    cat "$fixture_dir/router.log"
    exit 1
fi
rc=0
(cd frontend && NP_M1_BROWSER_FIXTURE="$(cat "$fixture_dir/ready.json")" ./node_modules/.bin/playwright test \
    --config playwright.m1.config.ts m1-repair.spec.ts -g 本机草稿恢复后 --output "$fixture_dir/browser") || rc=1
wait "$fixture_pid" || rc=1
tail -n 8 "$fixture_dir/router.log"
exit "$rc"
