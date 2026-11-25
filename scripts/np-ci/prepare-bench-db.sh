#!/usr/bin/env bash
# Bootstrap the frozen runner template and clone it for the benchmark workload.
set -euo pipefail

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

template=np_first_slice_dev
bench=np_benchmark
temp_dir="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
[[ "${DB_DRIVER:-}/${DB_HOST:-}/${DB_PORT:-}/${DB_USERNAME:-}/${DB_DATABASE:-}" == "postgres/127.0.0.1/18790/np_m1/$template" ]] || {
  echo "Unexpected benchmark database target" >&2
  exit 1
}
psql_cmd=(psql -X -qAt -v ON_ERROR_STOP=1 -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USERNAME")

go run ./cmd/migrate -command bootstrap >"$temp_dir/bootstrap.log" 2>&1
bash scripts/np-first-slice/apply-main-np86.sh >"$temp_dir/schema.log" 2>&1
"${psql_cmd[@]}" -d postgres -c "CREATE DATABASE $bench TEMPLATE $template"
"${psql_cmd[@]}" -d "$bench" -c "SELECT to_regclass('public.case_intake_write_requests') IS NOT NULL AND (SELECT version FROM schema_bootstrap_state WHERE id = 1) = 'postgres-mvp-2026-08-27-v14'" | grep -qx t
