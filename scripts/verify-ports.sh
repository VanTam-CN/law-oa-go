#!/bin/bash

# 端口验证脚本

echo "🔍 验证律师事务所OA系统端口配置..."
echo ""

# 定义端口映射
declare -A PORTS=(
    ["前端应用"]="3003"
    ["后端API"]="8080"
    ["HTTPS"]="443"
    ["HTTP"]="80"
    ["Prometheus"]="9090"
    ["Grafana"]="3000"
    ["Alertmanager"]="9093"
    ["Node Exporter"]="9100"
    ["MySQL Exporter"]="9104"
    ["Redis Exporter"]="9121"
    ["cAdvisor"]="8080"
    ["Loki"]="3100"
    ["Promtail"]="9080"
)

echo "📋 端口配置列表："
echo "========================="
for service in "${!PORTS[@]}"; do
    echo "  $service: ${PORTS[$service]}"
done
echo ""

# 检查端口是否被占用
echo "🔍 检查端口占用情况："
echo "========================="
for service in "${!PORTS[@]}"; do
    port=${PORTS[$service]}
    if lsof -i :$port >/dev/null 2>&1; then
        echo "  ❌ 端口 $port ($service) 已被占用"
        # 显示占用进程
        lsof -i :$port | tail -n +2 | while read line; do
            echo "     $line"
        done
    else
        echo "  ✅ 端口 $port ($service) 可用"
    fi
done
echo ""

# 检查Docker端口映射
echo "🐳 检查Docker容器端口映射："
echo "==============================="
if docker ps --format "table {{.Names}}\t{{.Ports}}" | grep -E "(3003|8080|443|80|9090|3000|9093)" >/dev/null 2>&1; then
    echo "发现运行中的容器端口映射："
    docker ps --format "table {{.Names}}\t{{.Ports}}" | grep -E "(3003|8080|443|80|9090|3000|9093)"
else
    echo "未发现相关容器运行"
fi
echo ""

# 检查compose文件中的端口配置
echo "📄 检查Docker Compose配置："
echo "==============================="
if [ -f deployments/docker-compose.prod.yml ]; then
    echo "生产环境配置文件中的端口映射："
    grep -A 2 "ports:" deployments/docker-compose.prod.yml | grep -E "^\s+- \"[0-9]+:" | sed 's/^[[:space:]]*- /  /'
else
    echo "❌ 未找到生产环境Docker Compose文件"
fi
echo ""

# 检查前端配置
echo "⚛️ 检查前端配置："
echo "=================="
if [ -f frontend/package.json ]; then
    echo "前端package.json中的脚本："
    grep -A 10 '"scripts"' frontend/package.json | grep -E "(start|dev|build)" || echo "未找到相关脚本"
else
    echo "❌ 未找到前端package.json文件"
fi

if [ -f frontend/.env ]; then
    echo "前端环境变量："
    grep "REACT_APP_API_URL" frontend/.env || echo "未找到API_URL配置"
else
    echo "ℹ️ 未找到前端.env文件（可能使用环境变量）"
fi
echo ""

# 检查后端配置
echo "🔧 检查后端配置："
echo "=================="
if [ -f cmd/server/main.go ]; then
    echo "后端服务配置："
    grep -E ":(8080|3003)" cmd/server/main.go || echo "未在main.go中找到端口配置"
    echo "后端端口通常通过环境变量API_PORT配置，默认为8080"
else
    echo "❌ 未找到后端main.go文件"
fi
echo ""

# 网络连接测试建议
echo "🌐 网络连接测试建议："
echo "===================="
echo "启动系统后，可以使用以下命令测试连接："
echo ""
echo "# 测试前端"
echo "curl -I http://localhost:3003"
echo ""
echo "# 测试后端API"
echo "curl -I http://localhost:8080/health"
echo ""
echo "# 测试HTTPS"
echo "curl -I https://localhost --insecure"
echo ""
echo "# 测试Prometheus"
echo "curl -I http://localhost:9090"
echo ""
echo "# 测试Grafana"
echo "curl -I http://localhost:3000"
echo ""

echo "✅ 端口验证完成！"