#!/usr/bin/env bash
# Extract probe diagnostics + artifacts for runs 45379251327 (0ca6eb2c) and 45379862576 (52f9d40a). bash-3.2 safe.
set -uo pipefail
R=repos/VanTam-CN/law-oa-go
RAW="$(cd "$(dirname "$0")" && pwd)/raw"
cd "$RAW"

for pair in "162:45379251327" "163:45379855010" "164:45379862576"; do
  num=${pair%%:*}; rid=${pair##*:}
  gh api "$R/actions/runs/$rid/artifacts" --jq '{total_count:.total_count,artifacts:[.artifacts[]|{name:.name,size_in_bytes:.size_in_bytes}]}' > "run-$num-artifacts.json" || echo "artifacts fail $num"
  gh api "$R/actions/runs/$rid/jobs" --jq '.jobs[]|select(.name|startswith("np-i11 ("))|"\(.id) \(.name)"' > ".jobs-$num.txt" || echo "jobs fail $num"
done

probe() { # probe jobid outfile
  {
    echo "### probe=$1 job=$2"
    gh api "$R/actions/jobs/$2/logs" 2>/dev/null \
      | grep -E "I11_EVENT|M2_[A-Z_]+[[:space:]]*$|I11_FUZZ[A-Z_ ]*[[:graph:]]|##\[error\]Process" \
      | grep -v "safe.directory" \
      || true
    echo
  } >> "$3"
}

for num in 162 164; do
  : > "diag-$num-probes.txt"
  while read -r jid jname; do
    [ -n "$jid" ] || continue
    p=${jname#np-i11 (}; p=${p%)}
    probe "$p" "$jid" "diag-$num-probes.txt"
  done < ".jobs-$num.txt"
done

for num in 162 164; do
  echo "== run $num probes =="
  awk '/^### probe=/{p=$2} /M2_[A-Z_]+[[:space:]]*$|I11_FUZZ/{print p": "$0}' "diag-$num-probes.txt" | sed 's/2026-11-24T[0-9:.]*Z //'
done
echo "== artifacts =="
cat run-162-artifacts.json run-163-artifacts.json run-164-artifacts.json 2>/dev/null
echo "== job counts =="
wc -l .jobs-162.txt .jobs-164.txt
