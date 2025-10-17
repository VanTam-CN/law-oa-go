#!/bin/bash

# 加载环境变量并运行法条数据库设置脚本

set -e

echo "🔧 加载环境变量..."

# 加载.env文件中的环境变量
if [ -f ".env" ]; then
    set -a
    source .env
    set +a

    # 映射环境变量名称以匹配脚本期望的名称
    export DB_USER=$DB_USERNAME
    export DB_NAME=$DB_DATABASE

    echo "✅ 环境变量加载完成"
    echo "   DB_USER: $DB_USER"
    echo "   DB_NAME: $DB_NAME"
    echo "   DB_HOST: $DB_HOST"
    echo "   DB_PORT: $DB_PORT"
else
    echo "❌ 错误: .env 文件未找到"
    exit 1
fi

# 确保PostgreSQL在PATH中
export PATH="/usr/local/opt/postgresql@16/bin:$PATH"

# 确保Go bin目录在PATH中
export PATH=$PATH:$(go env GOPATH)/bin

echo ""
echo "🚀 开始运行法条数据库设置..."
exec ./scripts/setup_legal_statutes.sh