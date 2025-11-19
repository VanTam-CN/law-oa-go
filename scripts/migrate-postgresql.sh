#!/bin/bash

# PostgreSQL Migration Script
# 用于执行PostgreSQL数据库迁移

set -e

# 默认配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-law_oa_user}
DB_PASSWORD=${DB_PASSWORD:-law_oa_password}
DB_NAME=${DB_NAME:-law_oa_db}
MIGRATIONS_DIR=${MIGRATIONS_DIR:-./migrations}

# 构建PostgreSQL DSN
PG_DSN="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

echo "🚀 开始PostgreSQL数据库迁移..."
echo "数据库: ${PG_DSN}"

# 检查psql是否安装
if ! command -v psql &> /dev/null; then
    echo "❌ 错误: psql 未安装"
    echo "请安装PostgreSQL客户端工具"
    exit 1
fi

# 检查数据库连接
echo "🔍 检查数据库连接..."
if ! PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;" > /dev/null 2>&1; then
    echo "❌ 数据库连接失败"
    echo "请检查数据库配置和连接信息"
    exit 1
fi

echo "✅ 数据库连接成功"

# 获取已执行的迁移版本
echo "📋 检查已执行的迁移..."
if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" | grep -q "t"; then
    echo "✅ 迁移表存在"
else
    echo "📝 创建迁移表..."
    PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version BIGINT PRIMARY KEY,
            dirty BOOLEAN NOT NULL DEFAULT FALSE
        );
    "
fi

# 执行迁移
echo "🔧 执行迁移文件..."

# 查找所有up迁移文件并按版本号排序
for migration_file in $(ls -1 ${MIGRATIONS_DIR}/*_up.sql | sort -V); do
    # 提取版本号 (文件名开头的数字)
    version=$(basename "$migration_file" | sed 's/_.*//' | sed 's/^0*//')

    # 如果版本号为空，跳过
    if [ -z "$version" ]; then
        echo "⚠️  跳过无效的迁移文件: $(basename "$migration_file")"
        continue
    fi

    # 检查是否已经执行过此迁移
    if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "SELECT version FROM schema_migrations WHERE version = $version;" | grep -q "$version"; then
        echo "⏭️  跳过已执行的迁移: v${version}"
        continue
    fi

    echo "📄 执行迁移: v${version} - $(basename "$migration_file")"

    # 执行迁移文件
    if PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -f "$migration_file"; then
        # 记录迁移版本
        PGPASSWORD=${DB_PASSWORD} psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USER} -d ${DB_NAME} -c "INSERT INTO schema_migrations (version, dirty) VALUES ($version, FALSE);"
        echo "✅ 迁移 v${version} 执行成功"
    else
        echo "❌ 迁移 v${version} 执行失败"
        exit 1
    fi
done

echo "🎉 所有迁移执行完成！"