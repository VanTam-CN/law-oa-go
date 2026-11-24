#!/usr/bin/env bash
# Only a target's real Go PASS can produce the matrix receipt.
set -euo pipefail
dir=$1
shift
cd "$dir"
for target in "$@"; do
  [[ "$target" =~ ^Fuzz_[A-Za-z0-9_]+$ ]]
  grep -Eq "^--- PASS: ${target} [(]" "fuzz-$target.log" || { echo "Missing PASS for $target" >&2; exit 1; }
  echo "- $target PASS" >> "$GITHUB_STEP_SUMMARY"
done
[[ $# -gt 0 ]]
echo 'verified=true' >> "$GITHUB_OUTPUT"
