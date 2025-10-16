#!/bin/bash

# 律师事务所OA系统监控配置脚本
# 用于设置监控指标收集和导出

set -e

echo "🔧 开始配置律师事务所OA系统监控..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 创建监控网络
echo "🌐 创建监控网络..."
docker network create monitoring-network 2>/dev/null || echo "监控网络已存在"

# 启动Node Exporter（系统指标）
echo "📊 启动Node Exporter..."
docker run -d \
  --name node-exporter \
  --network monitoring-network \
  --restart unless-stopped \
  -p 9100:9100 \
  -v "/:/host:ro,rslave" \
  --pid="host" \
  quay.io/prometheus/node-exporter:latest \
  --path.rootfs=/host

# 启动MySQL Exporter
echo "🗄️ 启动MySQL Exporter..."
docker run -d \
  --name mysql-exporter \
  --network monitoring-network \
  --restart unless-stopped \
  -p 9104:9104 \
  -e DATA_SOURCE_NAME="grafana:${GRAFANA_MYSQL_PASSWORD:-password}@(mysql:3306)/" \
  prom/mysqld-exporter:latest

# 启动Redis Exporter
echo "🧠 启动Redis Exporter..."
docker run -d \
  --name redis-exporter \
  --network monitoring-network \
  --restart unless-stopped \
  -p 9121:9121 \
  -e REDIS_ADDR="redis://redis:6379" \
  -e REDIS_PASSWORD="${REDIS_PASSWORD:-}" \
  oliver006/redis_exporter:latest

# 启动Nginx Exporter
echo "🌐 启动Nginx Exporter..."
docker run -d \
  --name nginx-exporter \
  --network monitoring-network \
  --restart unless-stopped \
  -p 9113:9113 \
  -v "/var/run/docker.sock:/var/run/docker.sock:ro" \
  nginx/nginx-prometheus-exporter:latest \
  -nginx.scrape-uri="http://nginx:80/nginx_status"

# 启动cAdvisor（容器监控）
echo "🐳 启动cAdvisor..."
docker run -d \
  --name cadvisor \
  --network monitoring-network \
  --restart unless-stopped \
  -p 8080:8080 \
  -v "/:/rootfs:ro" \
  -v "/var/run:/var/run:ro" \
  -v "/sys:/sys:ro" \
  -v "/var/lib/docker/:/var/lib/docker:ro" \
  -v "/dev/disk/:/dev/disk:ro" \
  --privileged=true \
  --device=/dev/kmsg \
  gcr.io/cadvisor/cadvisor:latest

# 等待导出器启动
echo "⏳ 等待监控导出器启动..."
sleep 10

# 验证导出器是否正常工作
echo "✅ 验证监控导出器状态..."

check_exporter() {
    local name=$1
    local port=$2
    local url=$3

    echo -n "检查 $name ... "
    if curl -s --connect-timeout 5 "$url" > /dev/null; then
        echo "✅ 正常"
        return 0
    else
        echo "❌ 异常"
        return 1
    fi
}

# 检查各个导出器
check_exporter "Node Exporter" 9100 "http://localhost:9100/metrics"
check_exporter "MySQL Exporter" 9104 "http://localhost:9104/metrics"
check_exporter "Redis Exporter" 9121 "http://localhost:9121/metrics"
check_exporter "cAdvisor" 8080 "http://localhost:8080/metrics"

# 创建监控配置目录
echo "📁 创建监控配置目录..."
mkdir -p configs/prometheus/rules
mkdir -p configs/alertmanager/templates
mkdir -p configs/grafana/provisioning/{datasources,dashboards}
mkdir -p configs/grafana/dashboards/{system,business}

# 创建告警模板目录
mkdir -p configs/alertmanager/templates

# 生成Alertmanager邮件模板
echo "📧 生成Alertmanager邮件模板..."
cat > configs/alertmanager/templates/email.tmpl << 'EOF'
{{ define "email.default.subject" }}
[{{ .Status | toUpper }}] {{ .GroupLabels.alertname }} 告警
{{ end }}

{{ define "email.default.body" }}
{{ range .Alerts }}
🚨 告警详情 🚨

告警名称: {{ .Annotations.summary }}
告警描述: {{ .Annotations.description }}
告警级别: {{ .Labels.severity }}
影响服务: {{ .Labels.service }}
实例: {{ .Labels.instance }}

开始时间: {{ .StartsAt.Format "2006-01-02 15:04:05" }}
{{ if .EndsAt }}结束时间: {{ .EndsAt.Format "2006-01-02 15:04:05" }}{{ end }}

{{ if .Annotations.runbook_url }}
📖 运行手册: {{ .Annotations.runbook_url }}
{{ end }}

{{ end }}
{{ end }}
EOF

# 生成业务指标收集脚本
echo "📈 生成业务指标收集脚本..."
cat > scripts/collect-business-metrics.sh << 'EOF'
#!/bin/bash

# 业务指标收集脚本
# 通过API获取业务数据并转换为Prometheus指标

API_BASE="http://localhost:8080/api"
METRICS_FILE="/tmp/business-metrics.prom"

# 清空指标文件
> "$METRICS_FILE"

# 获取系统统计信息
echo "📊 收集业务指标..."

# 案件统计
case_stats=$(curl -s "$API_BASE/dashboard/statistics" | jq -r '
  "# HELP law_oa_cases_total 总案件数",
  "# TYPE law_oa_cases_total gauge",
  "law_oa_cases_total " + (.data.total_cases | tostring),
  "# HELP law_oa_active_cases_total 活跃案件数",
  "# TYPE law_oa_active_cases_total gauge",
  "law_oa_active_cases_total " + (.data.active_cases | tostring),
  "# HELP law_oa_clients_total 客户总数",
  "# TYPE law_oa_clients_total gauge",
  "law_oa_clients_total " + (.data.total_clients | tostring),
  "# HELP law_oa_lawyers_total 律师总数",
  "# TYPE law_oa_lawyers_total gauge",
  "law_oa_lawyers_total " + (.data.total_lawyers | tostring)
')

echo "$case_stats" >> "$METRICS_FILE"

# 案件状态分布
case_status=$(curl -s "$API_BASE/cases/stats" | jq -r '
  .data.cases_by_type | to_entries[] |
  "# HELP law_oa_cases_by_status 按状态统计案件数",
  "# TYPE law_oa_cases_by_status gauge",
  "law_oa_cases_by_status{status=\"" + .key + "\"} " + (.value | tostring)
')

echo "$case_status" >> "$METRICS_FILE"

# 将指标推送到Pushgateway（如果配置了的话）
# curl --data-binary @"$METRICS_FILE" http://pushgateway:9091/metrics/job/business-metrics

echo "✅ 业务指标收集完成"
EOF

chmod +x scripts/collect-business-metrics.sh

# 创建监控健康检查脚本
echo "🏥 创建监控健康检查脚本..."
cat > scripts/monitoring-health-check.sh << 'EOF'
#!/bin/bash

# 监控系统健康检查脚本

FAILED_SERVICES=()

check_service() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}

    echo -n "检查 $name ... "

    status_code=$(curl -s -o /dev/null -w "%{http_code}" --connect-timeout 10 "$url" || echo "000")

    if [ "$status_code" = "$expected_status" ]; then
        echo "✅ 正常 ($status_code)"
        return 0
    else
        echo "❌ 异常 ($status_code)"
        FAILED_SERVICES+=("$name")
        return 1
    fi
}

echo "🔍 监控系统健康检查"
echo "==================="

# 检查Prometheus
check_service "Prometheus" "http://localhost:9090/-/healthy"

# 检查Grafana
check_service "Grafana" "http://localhost:3000/api/health"

# 检查Alertmanager
check_service "Alertmanager" "http://localhost:9093/-/healthy"

# 检查各个导出器
check_service "Node Exporter" "http://localhost:9100/metrics"
check_service "MySQL Exporter" "http://localhost:9104/metrics"
check_service "Redis Exporter" "http://localhost:9121/metrics"
check_service "cAdvisor" "http://localhost:8080/metrics"

# 检查应用服务
check_service "应用服务" "http://localhost:8080/health"

echo ""
echo "📋 检查结果总结"
echo "==================="

if [ ${#FAILED_SERVICES[@]} -eq 0 ]; then
    echo "🎉 所有监控服务运行正常！"
    exit 0
else
    echo "⚠️ 以下服务存在问题："
    for service in "${FAILED_SERVICES[@]}"; do
        echo "  - $service"
    done
    echo ""
    echo "请检查日志并修复问题。"
    exit 1
fi
EOF

chmod +x scripts/monitoring-health-check.sh

# 创建监控清理脚本
echo "🧹 创建监控清理脚本..."
cat > scripts/monitoring-cleanup.sh << 'EOF'
#!/bin/bash

# 监控系统清理脚本

echo "🧹 清理监控系统..."

# 停止并删除监控容器
echo "停止监控容器..."
docker stop node-exporter mysql-exporter redis-exporter nginx-exporter cadvisor 2>/dev/null || true
docker rm node-exporter mysql-exporter redis-exporter nginx-exporter cadvisor 2>/dev/null || true

# 清理监控网络
echo "清理监控网络..."
docker network rm monitoring-network 2>/dev/null || true

# 清理临时文件
echo "清理临时文件..."
rm -f /tmp/business-metrics.prom

echo "✅ 监控系统清理完成"
EOF

chmod +x scripts/monitoring-cleanup.sh

# 添加到crontab自动收集业务指标
echo "⏰ 设置定时任务..."
(crontab -l 2>/dev/null; echo "*/5 * * * * /path/to/law-oa-go/scripts/collect-business-metrics.sh") | crontab -

echo ""
echo "🎉 监控系统配置完成！"
echo ""
echo "📊 监控面板地址："
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Alertmanager: http://localhost:9093"
echo "  - Node Exporter: http://localhost:9100/metrics"
echo "  - cAdvisor: http://localhost:8080"
echo ""
echo "🔧 管理脚本："
echo "  - 健康检查: ./scripts/monitoring-health-check.sh"
echo "  - 收集指标: ./scripts/collect-business-metrics.sh"
echo "  - 清理监控: ./scripts/monitoring-cleanup.sh"
echo ""
echo "📖 下一步："
echo "1. 访问 Grafana 配置数据源"
echo "2. 导入预配置的仪表板"
echo "3. 配置告警规则"
echo "4. 设置通知渠道"