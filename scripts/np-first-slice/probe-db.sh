#!/bin/sh
# 只读探针：核对 127.0.0.1:18790 确为 np-first-slice 隔离 PG 16.14 / np_first_slice_dev。不改数据。
set -eu

cd "$(dirname "$0")"
ENV_FILE="${1:-./.env}"
if [ ! -f "$ENV_FILE" ]; then
  echo "FAIL: 缺少私有 env：$ENV_FILE"
  exit 2
fi
set -a
. "$ENV_FILE"
set +a
: "${POSTGRES_DB:?missing POSTGRES_DB}"
: "${POSTGRES_USER:?missing POSTGRES_USER}"

compose() { docker compose --env-file "$ENV_FILE" "$@"; }
compose ps --status running db >/dev/null 2>&1 || { echo "FAIL: np-first-slice/db 未运行"; exit 3; }
INFO=$(compose exec -T db psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -Atc "SELECT current_database() || '|server=' || current_setting('server_version') || '|user=' || current_user")
nc -z 127.0.0.1 18790 || { echo "FAIL: 回环 18790 不可达"; exit 4; }
case "$INFO" in
  np_first_slice_dev\|server=16.14*\|*) echo "PROBE_OK $INFO loopback=127.0.0.1:18790" ;;
  *) echo "FAIL: 实例不匹配，期望 np_first_slice_dev/16.14，实际 $INFO"; exit 5 ;;
esac
