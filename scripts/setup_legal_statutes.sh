#!/bin/bash

# 法条数据设置脚本
# 用于运行数据库迁移和导入基础数据

set -e

echo "🚀 开始设置法条数据系统..."

# 检查必要的环境变量
if [ -z "$DB_HOST" ]; then
    echo "⚠️  DB_HOST 未设置，使用默认值 localhost"
    export DB_HOST=localhost
fi

if [ -z "$DB_PORT" ]; then
    echo "⚠️  DB_PORT 未设置，使用默认值 5432"
    export DB_PORT=5432
fi

if [ -z "$DB_USER" ]; then
    echo "❌ 错误: DB_USER 环境变量必须设置"
    exit 1
fi

if [ -z "$DB_PASSWORD" ]; then
    echo "❌ 错误: DB_PASSWORD 环境变量必须设置"
    exit 1
fi

if [ -z "$DB_NAME" ]; then
    echo "⚠️  DB_NAME 未设置，使用默认值 law_oa_db"
    export DB_NAME=law_oa_db
fi

echo "📋 数据库配置:"
echo "   主机: $DB_HOST"
echo "   端口: $DB_PORT"
echo "   用户: $DB_USER"
echo "   数据库: $DB_NAME"
echo ""

# 检查golang-migrate是否安装
if ! command -v migrate &> /dev/null; then
    echo "❌ 错误: golang-migrate 未安装"
    echo "请运行以下命令安装:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    echo "或者参考: https://github.com/golang-migrate/migrate"
    exit 1
fi

# 检查数据库连接
echo "🔍 检查数据库连接..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 数据库连接正常"
else
    echo "❌ 数据库连接失败，请检查配置"
    exit 1
fi

# 运行数据库迁移
echo ""
echo "🔄 运行数据库迁移..."
cd migrations

# 创建迁移记录表（如果不存在）
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL
);" 2>/dev/null

# 执行迁移
migrate -path . -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up

if [ $? -eq 0 ]; then
    echo "✅ 数据库迁移完成"
else
    echo "❌ 数据库迁移失败"
    exit 1
fi

cd ..

# 编译并运行数据导入脚本
echo ""
echo "📦 准备导入法条数据..."
cd scripts

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "❌ 错误: Go 未安装或不在PATH中"
    exit 1
fi

# 编译数据导入程序
echo "🔨 编译数据导入程序..."
go build -o migrate_legal_statutes migrate_legal_statutes.go

if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

# 运行数据导入
echo "📥 导入法条数据..."
./migrate_legal_statutes

if [ $? -eq 0 ]; then
    echo "✅ 法条数据导入完成"
else
    echo "❌ 法条数据导入失败"
    exit 1
fi

# 清理临时文件
rm -f migrate_legal_statutes

cd ..

# 检查Elasticsearch状态（可选）
echo ""
echo "🔍 检查Elasticsearch状态..."
if command -v curl &> /dev/null; then
    if curl -s http://localhost:9200/_cluster/health > /dev/null 2>&1; then
        echo "✅ Elasticsearch 正在运行"
        echo "💡 提示: 请手动运行以下命令导入数据到Elasticsearch:"
        echo "   curl -X POST 'localhost:9200/legal_statutes/_bulk' -H 'Content-Type: application/json' --data-binary @scripts/legal_statutes_es_data.json"
    else
        echo "⚠️  Elasticsearch 未运行或不可访问"
        echo "💡 启动Elasticsearch后，运行以下命令导入数据:"
        echo "   curl -X POST 'localhost:9200/legal_statutes/_bulk' -H 'Content-Type: application/json' --data-binary @scripts/legal_statutes_es_data.json"
    fi
else
    echo "⚠️  curl 命令不可用，无法检查Elasticsearch状态"
fi

echo ""
echo "🎉 法条数据系统设置完成！"
echo ""
echo "📊 数据统计:"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT
    'legal_categories' as table_name, COUNT(*) as record_count
FROM legal_categories
UNION ALL
SELECT
    'legal_statutes' as table_name, COUNT(*) as record_count
FROM legal_statutes
UNION ALL
SELECT
    'legal_tags' as table_name, COUNT(*) as record_count
FROM legal_tags
UNION ALL
SELECT
    'legal_search_history' as table_name, COUNT(*) as record_count
FROM legal_search_history
ORDER BY table_name;
" 2>/dev/null

echo ""
echo "📁 生成的文件:"
echo "   - scripts/legal_statutes_es_data.json (Elasticsearch数据文件)"
echo ""
echo "🔧 下一步操作:"
echo "   1. 确保Elasticsearch正在运行"
echo "   2. 导入数据到Elasticsearch索引"
echo "   3. 启动后端服务测试法条搜索功能"
echo "   4. 更新前端界面以支持新的法条搜索功能"