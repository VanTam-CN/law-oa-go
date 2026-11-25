#!/usr/bin/env bash
# Exercise the exact shell used by the registered fuzz workflow with synthetic Actions inputs.
set -euo pipefail
probe=${1:?probe required}
case "$probe" in fuzz-normal|fuzz-unknown|fuzz-selected-failed|fuzz-missing-proof) ;; *) exit 2 ;; esac
work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT
export GITHUB_OUTPUT="$work/output" GITHUB_STEP_SUMMARY="$work/summary"
workflow=.github/workflows/fuzzing.yml
step_script() {
  ruby -ryaml -e 'steps=YAML.load_file(ARGV[0]).fetch("jobs").fetch(ARGV[1]).fetch("steps"); run=steps.fetch(0).fetch("run"); abort if run.empty?; print run' "$workflow" "$1"
}
step_script preflight > "$work/preflight"
step_script fuzzing-summary > "$work/check"
export RAW_TEST_TYPE=validators RAW_DURATION=1 RAW_FUZZERS=1
if [[ "$probe" == fuzz-unknown ]]; then RAW_TEST_TYPE=unknown; fi
if ! bash -e "$work/preflight"; then echo I11_FUZZ_PREFLIGHT_REJECTED; exit 1; fi
grep -qx 'test_type=validators' "$GITHUB_OUTPUT" || { echo I11_FUZZ_INPUT_MISMATCH; exit 1; }
export TEST_TYPE=validators
RESULTS=$(jq -cn --arg probe "$probe" '
  {preflight:{result:"success"},"check-fuzzing":{result:"success"}} +
  (["security","validators","repositories","cache","concurrency","comprehensive"] | map({
    key:(. + "-fuzzing"),
    value:(if . != "validators" then {result:"skipped"} elif $probe == "fuzz-selected-failed" then {result:"failure",outputs:{verified:"true"}}
      elif $probe == "fuzz-missing-proof" then {result:"success",outputs:{}}
      else {result:"success",outputs:{verified:"true"}} end)
  }) | from_entries)')
export RESULTS
if ! bash -e "$work/check"; then echo "I11_FUZZ_SUMMARY_REJECTED $probe"; exit 1; fi
grep -qx 'Verified selected fuzz jobs: validators' "$GITHUB_STEP_SUMMARY" || exit 1
echo I11_FUZZ_ACCEPTED
