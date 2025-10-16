#!/bin/bash

# 律师事务所OA系统生产环境启动脚本

set -e

echo "🚀 启动律师事务所OA系统生产环境..."

# 检查Docker和Docker Compose
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

# 检查环境变量文件
if [ ! -f .env.production ]; then
    echo "❌ 未找到.env.production文件"
    echo "请复制.env.example并配置生产环境变量"
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p logs/nginx
mkdir -p logs/app
mkdir -p uploads
mkdir -p ssl

# 生成自签名SSL证书（仅用于测试）
if [ ! -f ssl/cert.pem ]; then
    echo "🔐 生成SSL证书..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ssl/key.pem \
        -out ssl/cert.pem \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=Law OA/CN=localhost" \
        2>/dev/null || echo "SSL证书生成失败，请手动配置"
fi

# 加载环境变量
export $(cat .env.production | grep -v '^#' | xargs)

# 启动监控系统（可选）
read -p "是否启动监控系统？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📊 启动监控系统..."
    ./scripts/monitoring-setup.sh
fi

# 启动生产环境
echo "🏗️ 构建和启动生产环境..."
docker-compose -f deployments/docker-compose.prod.yml --env-file .env.production up -d --build

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 30

# 健康检查
echo "🏥 执行健康检查..."
./scripts/monitoring-health-check.sh 2>/dev/null || echo "监控检查失败，但服务可能正在启动中"

# 显示服务状态
echo ""
echo "📋 服务状态："
docker-compose -f deployments/docker-compose.prod.yml ps

echo ""
echo "🎉 律师事务所OA系统启动完成！"
echo ""
echo "📱 访问地址："
echo "  🌐 前端应用: http://localhost:3003"
echo "  🔒 HTTPS访问: https://localhost"
echo "  📊 后端API: http://localhost:8080/api"
echo "  📈 Prometheus: http://localhost:9090"
echo "  📊 Grafana: http://localhost:3000 (admin/admin)"
echo "  🚨 Alertmanager: http://localhost:9093"
echo ""
echo "🔧 管理命令："
echo "  查看日志: docker-compose -f deployments/docker-compose.prod.yml logs -f [service]"
echo "  停止系统: docker-compose -f deployments/docker-compose.prod.yml down"
echo "  重启服务: docker-compose -f deployments/docker-compose.prod.yml restart [service]"
echo ""
echo "📖 更多信息请查看 docs/API_DOCUMENTATION.md"