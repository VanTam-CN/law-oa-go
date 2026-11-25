#!/usr/bin/env bash
# Go emits one test binary per package. Keep its profiles on this runner.
set -euo pipefail
dir=$1
shift
(( $# > 0 )) || { echo "No PGO packages selected" >&2; exit 1; }
mkdir -p "$dir"/parts
for package in "$@"; do
  part=$(mktemp -d "$dir/parts/pkg.XXXXXX")
  timeout_args=()
  if [[ "$package" == law-oa-go/tests/performance ]]; then
    timeout_args=(-timeout=35m)
  fi
  go test "${timeout_args[@]}" -count=1 -bench=. -benchtime=30s -cpuprofile="$part/cpu.prof" -memprofile="$part/mem.prof" -o "$part/tests" "$package" >"$part/output" 2>&1 || {
    echo "PGO profiling failed for $package" >&2
    if [[ "$package" == law-oa-go/tests/performance ]]; then
      python3 scripts/np-ci/bench-results.py "$part/output"
    fi
    exit 1
  }
  if [[ "$package" == law-oa-go/tests/performance ]]; then
    python3 scripts/np-ci/bench-results.py "$part/output"
  fi
  test -s "$part/cpu.prof" && test -s "$part/mem.prof" || { echo "Missing profile for $package" >&2; exit 1; }
  rm "$part/tests"
done
for kind in cpu mem; do
  go tool pprof -proto -output "$dir/$kind.prof" "$dir"/parts/*/"$kind.prof"
  test -s "$dir/$kind.prof"
done
