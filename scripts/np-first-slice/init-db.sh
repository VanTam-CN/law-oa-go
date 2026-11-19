#!/bin/sh
# NP v0.4 首切片隔离库初始化：全新空库 -> bootstrap -> 仅 000082-000085 追加。
# 凭据只从私有 .env 读取且不打印；任何一步失败立即停止，保留首失败。
set -eu

cd "$(dirname "$0")"
ROOT="$(cd ../.. && pwd)"
ENV_FILE="${1:-./.env}"

if [ ! -f "$ENV_FILE" ]; then
  echo "FAIL: 缺少私有 env：$ENV_FILE（变量清单见 docs/np-first-slice-environment.md，值用全新虚构凭据）"
  exit 2
fi
chmod 600 "$ENV_FILE"
set -a
. "$ENV_FILE"
set +a
: "${POSTGRES_DB:?missing POSTGRES_DB}"
: "${POSTGRES_USER:?missing POSTGRES_USER}"
: "${POSTGRES_PASSWORD:?missing POSTGRES_PASSWORD}"
: "${REDIS_PASSWORD:?missing REDIS_PASSWORD}"

compose() { docker compose --env-file "$ENV_FILE" "$@"; }

echo "== 1/5 启动 db/redis（仅回环 18790/18791）"
compose up -d db redis
compose up --wait --wait-timeout 60 db redis

echo "== 2/5 核对全新空库"
TBL_COUNT=$(compose exec -T db psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -Atc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
if [ "$TBL_COUNT" != "0" ]; then
  echo "FAIL: 目标库非空（public 表数=$TBL_COUNT），拒绝在非全新卷上初始化"
  exit 3
fi

echo "== 3/5 构建 backend/cmd/migrate"
mkdir -p .bin
( cd "$ROOT/backend" && go build -o "$ROOT/scripts/np-first-slice/.bin/np-migrate" ./cmd/migrate )

echo "== 4/5 执行候选 bootstrap（指向 127.0.0.1:18790）"
DB_DRIVER=postgres \
DB_HOST=127.0.0.1 \
DB_PORT=18790 \
DB_USERNAME="$POSTGRES_USER" \
DB_PASSWORD="$POSTGRES_PASSWORD" \
DB_DATABASE="$POSTGRES_DB" \
  .bin/np-migrate -command bootstrap

echo "== 5/5 按序追加 000082-000085"
for f in 000082_intake_collaboration_mvp 000083_approval_supplement_replies 000084_matter_collaboration 000085_case_task_change_audit; do
  echo "-- $f"
  compose exec -T db sh -c 'exec psql -v ON_ERROR_STOP=1 --single-transaction -U "$POSTGRES_USER" -d "$POSTGRES_DB"' \
    < "$ROOT/backend/migrations/$f.up.sql"
done

VERSION=$(compose exec -T db psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -Atc "SELECT version FROM schema_bootstrap_state WHERE id=1")
TBL_COUNT=$(compose exec -T db psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -Atc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")
echo "INIT_OK version=$VERSION public_tables=$TBL_COUNT"
