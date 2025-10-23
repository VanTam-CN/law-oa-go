#!/bin/bash

echo "🚨 PostgreSQL 终极修复脚本"
echo "=========================="

# 检查参数
if [ $# -eq 0 ]; then
    echo "用法: $0 <postgresql连接参数>"
    echo "示例: $0 'host=localhost port=5432 user=postgres password=你的密码 dbname=law_oa_go'"
    echo ""
    echo "或者设置环境变量:"
    echo "export PG_HOST=localhost"
    echo "export PG_PORT=5432"
    echo "export PG_USER=postgres"
    echo "export PG_PASSWORD=你的密码"
    echo "export PG_DATABASE=law_oa_go"
    echo "$0"
    exit 1
fi

# 获取连接字符串
if [ "$1" = "env" ]; then
    # 使用环境变量
    PG_HOST=${PG_HOST:-localhost}
    PG_PORT=${PG_PORT:-5432}
    PG_USER=${PG_USER:-postgres}
    PG_PASSWORD=${PG_PASSWORD:-postgres}
    PG_DATABASE=${PG_DATABASE:-law_oa_go}

    CONN_STR="-h $PG_HOST -p $PG_PORT -U $PG_USER -d $PG_DATABASE"
    echo "使用环境变量连接: $CONN_STR"
else
    # 使用命令行参数
    CONN_STR=$1
    echo "使用命令行参数连接: $CONN_STR"
fi

# 检查psql命令
if ! command -v psql &> /dev/null; then
    echo "❌ psql命令未找到"
    echo "macOS: brew install postgresql"
    echo "Ubuntu: sudo apt-get install postgresql-client"
    exit 1
fi

echo "✅ psql命令找到"

# 执行修复SQL
echo "🛠️ 执行PostgreSQL修复..."

# 设置密码环境变量
if [ ! -z "$PG_PASSWORD" ]; then
    export PGPASSWORD="$PG_PASSWORD"
fi

# 执行修复SQL文件
SCRIPT_DIR="$(dirname "$0")"
SQL_FILE="$SCRIPT_DIR/working-postgresql-fix.sql"

if [ ! -f "$SQL_FILE" ]; then
    echo "❌ 找不到SQL文件: $SQL_FILE"
    exit 1
fi

echo "执行SQL文件: $SQL_FILE"

psql $CONN_STR -f "$SQL_FILE"

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 PostgreSQL修复成功完成！"
    echo ""
    echo "✅ 已添加的内容:"
    echo "   - lawyers 表 (律师管理)"
    echo "   - departments 表 (部门管理)"
    echo "   - documents 表 (文档管理)"
    echo "   - users表新字段 (6个)"
    echo "   - clients表新字段 (3个)"
    echo "   - cases表新字段 (6个)"
    echo "   - 示例律师数据 (3条)"
    echo "   - 相关索引和约束"
    echo ""
    echo "🔍 请刷新数据库管理工具查看结果"
    echo "📊 现在应该能看到完整的表结构和字段"
else
    echo ""
    echo "❌ PostgreSQL修复失败"
    echo "请检查:"
    echo "1. 数据库连接参数"
    echo "2. PostgreSQL服务状态"
    echo "3. 用户权限"
    exit 1
fi