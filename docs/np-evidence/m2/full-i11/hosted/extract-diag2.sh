#!/usr/bin/env bash
# Extract per-job diagnostics for the four hosted I11 events (pure bash + gh api).
set -uo pipefail
R=repos/VanTam-CN/law-oa-go
RAW="$(cd "$(dirname "$0")" && pwd)/raw"
cd "$RAW"

# artifacts snapshots (retry, keep small)
for pair in "157:45378121147" "158:45378128854" "159:45378568403" "160:45378578245"; do
  num=${pair%%:*}; id=${pair##*:}
  gh api "$R/actions/runs/$id/artifacts" --jq '{total_count:.total_count,artifacts:[.artifacts[]|{name:.name,size_in_bytes:.size_in_bytes,expired:.expired,expires_at:.expires_at}]}' > "run-$num-artifacts.json"
done

FULL157=$1
FULL159=$2

dump_full() { # outfile jobid
  gh api "$R/actions/jobs/$2/logs" 2>/dev/null \
    | grep -E "I11_(EVENT|SOURCE|START|END|FULL|PG_PRESERVED|BROWSER_PASS|TEST_PASS|REDIS)|Traceback|JSONDecodeError|check-i11-tests\.py|##\[error\]Process" \
    > "$1" || true
}

dump_full "diag-157-full.log" "$FULL157"
dump_full "diag-159-full.log" "$FULL159"

probe() { # outfile probe jobid
  {
    echo "### probe=$1 job=$2"
    gh api "$R/actions/jobs/$2/logs" 2>/dev/null \
      | grep -E "I11_EVENT|M2_[A-Z_]+[[:space:]]*$|I11_FUZZ[A-Z_ ]*[[:graph:]]|##\[error\]Process" \
      | grep -v "safe.directory" \
      || true
    echo
  } >> "$3"
}

: > diag-158-probes.txt
probe missing-pg 138389756868 diag-158-probes.txt
probe missing-dsn 138389757084 diag-158-probes.txt
probe skip 138389756411 diag-158-probes.txt
probe partial 138389757014 diag-158-probes.txt
probe none 138389757164 diag-158-probes.txt
probe fuzz-unknown 138389756843 diag-158-probes.txt
probe fuzz-selected-failed 138389756903 diag-158-probes.txt
probe fuzz-missing-proof 138389756857 diag-158-probes.txt

: > diag-160-probes.txt
probe missing-pg 138391311628 diag-160-probes.txt
probe missing-dsn 138391311704 diag-160-probes.txt
probe fuzz-selected-failed 138391311712 diag-160-probes.txt
probe partial 138391311802 diag-160-probes.txt
probe fuzz-missing-proof 138391311805 diag-160-probes.txt
probe fuzz-unknown 138391312144 diag-160-probes.txt
probe none 138391312173 diag-160-probes.txt
probe skip 138391312453 diag-160-probes.txt

wc -l diag-*.log diag-*-probes.txt run-*-artifacts.json
