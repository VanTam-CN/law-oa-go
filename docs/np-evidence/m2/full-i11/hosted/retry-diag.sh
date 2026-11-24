#!/usr/bin/env bash
# Retry the probe sections whose log fetch returned empty.
set -uo pipefail
R=repos/VanTam-CN/law-oa-go
RAW="$(cd "$(dirname "$0")" && pwd)/raw"
cd "$RAW"

redo() { # run probe jobid
  local run=$1 probe=$2 jid=$3
  local body
  body=$(gh api "$R/actions/jobs/$jid/logs" 2>/dev/null \
    | grep -E "I11_EVENT|M2_[A-Z_]+[[:space:]]*$|I11_FUZZ[A-Z_ ]*[[:graph:]]|##\[error\]Process" \
    | grep -v "safe.directory" || true)
  if [[ -z "$body" ]]; then echo "STILL EMPTY: run$run $probe"; return 1; fi
  # rewrite the section for this probe in the diag file
  local f="diag-$run-probes.txt" tmp="$RAW/.tmp-section"
  { echo "### probe=$probe job=$jid"; echo "$body"; echo; } > "$tmp"
  awk -v sect="$probe" -v repl="$tmp" '
    BEGIN { skip=0 }
    /^### probe=/ {
      p=$0; sub(/^### probe=/,"",p); sub(/ job=.*$/,"",p)
      skip=(p==sect)
    }
    { if (!skip) print }
    END {}
  ' "$f" > "$f.new"
  # insert replacement before the next section or at end
  if grep -q "^### probe=$probe job=$jid$" "$f.new"; then
    awk -v ins="$tmp" '/^### probe=/ && !done { while ((getline line < ins) > 0) print line; close(ins); done=1 } { print }' "$f.new" > "$f.final"
  else
    cat "$f.new" "$tmp" > "$f.final"
  fi
  mv "$f.final" "$f"; rm -f "$f.new" "$tmp"
  echo "OK: run$run $probe"
}

redo 158 fuzz-selected-failed 138389756903 || true
redo 158 fuzz-missing-proof 138389756857 || true
redo 160 missing-pg 138391311628 || true
redo 160 missing-dsn 138391311704 || true
redo 160 fuzz-selected-failed 138391311712 || true
redo 160 partial 138391311802 || true
