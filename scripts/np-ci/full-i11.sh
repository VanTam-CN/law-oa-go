#!/usr/bin/env bash
# Called inside early-pg's owned, identity-checked PG container window.
set -euo pipefail
umask 077
work=${1:?dedicated evidence directory required}
[[ -s "$work/cid" && -n ${NP_TEST_DSN:-} && ${DB_DATABASE:-} == np_first_slice_dev ]] || exit 2
export NP_LOCAL_CHECK=synthetic-only ENVIRONMENT=development VITE_NP_MEMORY_ONLY=true
export VITE_API_PROXY_TARGET=http://127.0.0.1:18789 PYTHON=/usr/bin/python3
NP_BROWSER_PASSWORD=$(openssl rand -hex 18)
JWT_SECRET=$(openssl rand -hex 32)
SUBJECT_DATA_KEY=$(openssl rand -base64 32)
export NP_BROWSER_PASSWORD JWT_SECRET SUBJECT_DATA_KEY

# Raw outputs stay in the disposable directory. Only fixed labels, times, counts and digests leave it.
run_step() {
  local label=$1 rc=0
  shift
  echo "I11_START $label $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  "$@" >"$work/$label.log" 2>&1 || rc=$?
  echo "I11_END $label rc=$rc $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  [[ $rc == 0 ]]
}
echo "I11_SOURCE $(git rev-parse HEAD)"
run_step fuzz-normal bash scripts/np-ci/i11-fuzz-probe.sh fuzz-normal
run_step build go build ./...
docker run --rm -d --cidfile "$work/redis-cid" -p 127.0.0.1::6379 redis:7-alpine redis-server --save '' --appendonly no > /dev/null
trap 'docker rm -f "$(cat "$work/redis-cid")" >/dev/null' EXIT
NP_I11_REDIS=$(docker port "$(cat "$work/redis-cid")" 6379/tcp)
export NP_I11_REDIS
for ((try=0; try<30; try++)); do
  if [[ $(docker exec "$(cat "$work/redis-cid")" redis-cli ping 2>/dev/null) == PONG ]]; then break; fi
  sleep 1
done
run_step redis go test -race -count=1 -tags np_i11_redis -run '^TestI11ResponseCacheRealRedis$' -v .
grep -q '^--- PASS: TestI11ResponseCacheRealRedis ' "$work/redis.log" || { echo I11_REDIS_PROOF_MISSING; exit 1; }
go_selector=$("$PYTHON" - <<'PY'
import json
items = set(json.load(open('docs/np-evidence/m3/independent-review/verification.json'))['go']['expected'])
items.discard('TestNPGoDigestMatchesBrowserManifest')  # real manifest is produced inside the route-b window
items.update(['TestIntakeEarlyResponsesNoStore', 'TestNarrowPilotCacheZeroReadWrite'])
print('^(' + '|'.join(sorted(items)) + ')$')
PY
)
run_step go go test -p 1 -race -count=1 -json . ./internal/handlers ./internal/database ./internal/middleware ./internal/router -run "$go_selector"
"$PYTHON" scripts/np-ci/check-i11-tests.py go "$work/go.log"
run_step tagged go test -p 1 -json -race -count=1 -tags np_m1_pg ./tests/integration/m1 -run '^(TestNativeConcurrentDraftAndHistory|TestAsyncPGFormalDetectorCompletesOnce|TestConflictPGLinkAdvancesRevisionOnce|TestM3ConflictCompletionAfterSourceSave|TestSchemaRequiresFrozenKeysAndTypes)$'
"$PYTHON" scripts/np-ci/check-i11-tests.py tagged "$work/tagged.log"
(
  cd frontend
  run_step typecheck npm run type-check
  run_step jest node node_modules/jest/bin/jest.js --runInBand --coverage=false --json "--outputFile=$work/jest.json" \
    src/components/intake src/pages/auth/__tests__ src/pages/batch01/__tests__/narrowPilotI13.test.ts \
    src/pages/batch01/__tests__/CaseIntakeConflictAction.test.ts src/components/layout/__tests__/Header.test.tsx src/utils/__tests__/storage.test.ts
  run_step frontend-build npm run build
)
"$PYTHON" scripts/np-ci/check-i11-tests.py jest "$work/jest.json"
for suite in route-b selection storage navigation editing tabs legacy prod native; do
  export PLAYWRIGHT_JSON_OUTPUT_FILE="$work/$suite.json"
  if [[ "$suite" == native ]]; then
    (cd frontend && VITE_NP_MEMORY_ONLY=false run_step native node node_modules/@playwright/test/cli.js test -c playwright.m1.config.ts --reporter=json cases.spec.ts conflict-actions.spec.ts -g '创建重放保留|直接审批|重复保存|立案冲突门禁')
  else
    run_step "$suite" bash scripts/np-first-slice/run-m4-browser.sh "$suite"
  fi
  "$PYTHON" scripts/np-ci/check-i11-browser.py "$suite" "$PLAYWRIGHT_JSON_OUTPUT_FILE"
  if [[ "$suite" == route-b ]]; then
    "$PYTHON" - "$work/$suite.log" <<'PY'
import json, re, sys
lines = open(sys.argv[1]).read().splitlines()
rows = [json.loads(line) for line in lines if line.startswith('{"protected_sha256":')]
assert len(rows) == 1 and 'GO_BROWSER_DIGEST_MATCH' in lines and 'BROWSER_PG_AFTER_OK' in lines
d = rows[0]
assert re.fullmatch('[0-9a-f]{64}', d['protected_sha256'])
assert [d[key] for key in ('sources', 'receipts', 'linked', 'manual')] == [3, 4, 3, 1]
print('I11_PG_PRESERVED', d['protected_sha256'], 'sources=3 receipts=4 linked=3 manual=1')
PY
  fi
done
echo I11_FULL_PASS
