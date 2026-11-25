#!/usr/bin/env bash
# Controlled offline append of the frozen 000086 NP migration to exactly one
# v14 synthetic candidate (np_first_slice_dev / np_m1). First run appends in a
# single advisory-locked transaction and re-verifies with the shared catalog
# check; repeat runs only verify. Partial table presence fails and rolls back.
# This is not a general migration platform.
set -euo pipefail

NP86_HOST="127.0.0.1"
NP86_PORT="18790"
NP86_DB="np_first_slice_dev"
NP86_USER="np_m1"
NP86_PSQL="${NP86_PSQL:-$(command -v psql || echo /usr/local/opt/libpq/bin/psql)}"
expected_sha256="415ef4f4b4a0437a403988810c52c3dca213fd0f6a9d9654d8161106fb7c3be2"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
[[ $# -eq 0 ]] || { echo "usage: apply-main-np86.sh" >&2; exit 2; }
cd "$repo_root"
sql_file="migrations/000086_narrow_pilot.up.sql"
schema_check="$repo_root/internal/database/narrow_pilot_schema.sql"

[[ -x "$NP86_PSQL" ]] || { echo "FAIL: psql not found: $NP86_PSQL" >&2; exit 1; }
[[ -r "$sql_file" ]] || { echo "FAIL: migration file not readable: $sql_file" >&2; exit 1; }
[[ -r "$schema_check" ]] || { echo "FAIL: shared schema check missing: $schema_check" >&2; exit 1; }
actual_sha256="$(shasum -a 256 "$sql_file" | awk '{print $1}')"
[[ "$actual_sha256" == "$expected_sha256" ]] || {
    echo "FAIL: 000086 sha256 mismatch: $actual_sha256" >&2
    exit 1
}

unset PGHOSTADDR PGSERVICE PGOPTIONS
psql_cmd=("$NP86_PSQL" -X -qAt -v ON_ERROR_STOP=1 -h "$NP86_HOST" -p "$NP86_PORT" -U "$NP86_USER" -d "$NP86_DB")

{
    cat <<'SQL'
BEGIN;
SET LOCAL search_path = public, pg_catalog;
SELECT pg_advisory_xact_lock(hashtext('law-oa-main-np86-apply'));
DO $guard$
BEGIN
    IF current_database() <> 'np_first_slice_dev' OR current_user <> 'np_m1' THEN
        RAISE EXCEPTION 'FAIL: target identity mismatch: % / %', current_database(), current_user;
    END IF;
    IF (SELECT version FROM schema_bootstrap_state WHERE id = 1) IS DISTINCT FROM 'postgres-mvp-2026-08-27-v14' THEN
        RAISE EXCEPTION 'FAIL: baseline is not v14: %', (SELECT version FROM schema_bootstrap_state WHERE id = 1);
    END IF;
END
$guard$;
SELECT to_regclass('public.case_intake_source_snapshots') IS NOT NULL AS src_exists,
       to_regclass('public.case_intake_write_requests') IS NOT NULL AS wr_exists
\gset
\if :src_exists
\if :wr_exists
SELECT 'repeat: verify-only mode' AS mode;
\else
DO $do$ BEGIN RAISE EXCEPTION 'FAIL: partial 000086 state: only case_intake_source_snapshots present'; END $do$;
\endif
\else
\if :wr_exists
DO $do$ BEGIN RAISE EXCEPTION 'FAIL: partial 000086 state: only case_intake_write_requests present'; END $do$;
\else
\i migrations/000086_narrow_pilot.up.sql
SELECT 'first apply: appending frozen 000086' AS mode;
\endif
\endif
SQL
    cat "$schema_check"
    cat <<'PSQL'
\gset
\if :ready
SELECT 'PASS: NP schema verified inside transaction' AS result;
\else
DO $do$ BEGIN RAISE EXCEPTION 'FAIL: shared catalog check returned not-ready'; END $do$;
\endif
COMMIT;
PSQL
} | "${psql_cmd[@]}"
