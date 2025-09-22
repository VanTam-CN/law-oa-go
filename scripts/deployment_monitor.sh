#!/bin/bash

# 部署监控脚本
# 监控部署后的应用状态，收集性能指标和错误日志

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -e, --env ENVIRONMENT      监控环境 (dev|staging|production)"
    echo "  -i, --interval SECONDS     监控间隔 (默认: 60)"
    echo "  -d, --duration MINUTES     监控时长 (默认: 30)"
    echo "  -o, --output DIR           输出目录 (默认: monitoring)"
    echo "  -a, --alert-threshold COUNT 错误阈值 (默认: 10)"
    echo "  -r, --report               生成监控报告"
    echo ""
    echo "示例:"
    echo "  $0 -e production -i 30 -d 60          # 生产环境监控30分钟，每30秒检查一次"
    echo "  $0 -e staging -r                      # 生成测试环境监控报告"
}

# 监控配置
declare -A MONITORING_ENDPOINTS=(
    ["health"]="http://localhost:8080/health"
    ["metrics"]="http://localhost:8080/metrics"
    ["status"]="http://localhost:8080/status"
)

# 监控数据
MONITORING_DATA=()

# 检查应用健康状态
check_application_health() {
    local endpoint=$1
    local environment=$2
    
    log_info "检查应用健康状态: $endpoint"
    
    local response_code=$(curl -s -o /dev/null -w "%{http_code}" "$endpoint" 2>/dev/null || echo "000")
    
    case $response_code in
        200)
            log_success "应用健康检查通过"
            return 0
            ;;
        000)
            log_error "无法连接到应用"
            return 1
            ;;
        401|403)
            log_warning "应用认证问题"
            return 2
            ;;
        500)
            log_error "应用内部错误"
            return 3
            ;;
        *)
            log_warning "应用返回异常状态码: $response_code"
            return 4
            ;;
    esac
}

# 收集性能指标
collect_performance_metrics() {
    local environment=$1
    local output_dir=$2
    
    log_info "收集性能指标..."
    
    local metrics_file="$output_dir/metrics_$(date +%Y%m%d_%H%M%S).json"
    
    # 收集系统指标
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    local memory_usage=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')
    local disk_usage=$(df -h / | awk 'NR==2{print $5}' | cut -d'%' -f1)
    
    # 收集应用指标
    local app_metrics=""
    if curl -s "${MONITORING_ENDPOINTS["metrics"]}" > /dev/null 2>&1; then
        app_metrics=$(curl -s "${MONITORING_ENDPOINTS["metrics"]}")
    fi
    
    # 收集Docker指标
    local docker_stats=""
    if command -v docker &> /dev/null; then
        docker_stats=$(docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" 2>/dev/null || echo "")
    fi
    
    # 生成指标JSON
    cat > "$metrics_file" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "environment": "$environment",
  "system": {
    "cpu_usage": ${cpu_usage:-0},
    "memory_usage": ${memory_usage:-0},
    "disk_usage": ${disk_usage:-0}
  },
  "application": {
    "metrics": $([ -n "$app_metrics" ] && echo "$app_metrics" || echo "{}")
  },
  "docker": {
    "stats": "$docker_stats"
  }
}
EOF
    
    log_success "性能指标已收集: $metrics_file"
    
    # 检查阈值
    check_thresholds "$cpu_usage" "$memory_usage" "$disk_usage"
}

# 检查指标阈值
check_thresholds() {
    local cpu_usage=$1
    local memory_usage=$2
    local disk_usage=$3
    
    local threshold_cpu=80
    local threshold_memory=85
    local threshold_disk=90
    
    local alerts=()
    
    if (( $(echo "$cpu_usage > $threshold_cpu" | bc -l) )); then
        alerts+=("CPU使用率过高: ${cpu_usage}%")
    fi
    
    if (( $(echo "$memory_usage > $threshold_memory" | bc -l) )); then
        alerts+=("内存使用率过高: ${memory_usage}%")
    fi
    
    if [ "$disk_usage" -gt "$threshold_disk" ]; then
        alerts+=("磁盘使用率过高: ${disk_usage}%")
    fi
    
    if [ ${#alerts[@]} -gt 0 ]; then
        log_error "触发监控告警:"
        for alert in "${alerts[@]}"; do
            log_error "  - $alert"
        done
        
        # 发送告警通知
        send_alert_notification "${alerts[@]}"
    fi
}

# 分析错误日志
analyze_error_logs() {
    local environment=$1
    local output_dir=$2
    local alert_threshold=$3
    
    log_info "分析错误日志..."
    
    local log_files=()
    
    # 查找日志文件
    if [ -d "/var/log/law-oa-go" ]; then
        log_files+=("/var/log/law-oa-go/*.log")
    fi
    
    if [ -d "./logs" ]; then
        log_files+=("./logs/*.log")
    fi
    
    local error_count=0
    local error_report="$output_dir/error_analysis_$(date +%Y%m%d_%H%M%S).txt"
    
    # 分析错误日志
    for log_pattern in "${log_files[@]}"; do
        for log_file in $log_pattern; do
            if [ -f "$log_file" ]; then
                local file_errors=$(grep -i "error\|exception\|panic\|fatal" "$log_file" | wc -l)
                error_count=$((error_count + file_errors))
                
                # 提取最近的错误
                echo "=== 错误日志分析: $log_file ===" >> "$error_report"
                grep -i "error\|exception\|panic\|fatal" "$log_file" | tail -10 >> "$error_report"
                echo "" >> "$error_report"
            fi
        done
    done
    
    log_info "发现 $error_count 个错误"
    
    # 检查错误阈值
    if [ "$error_count" -gt "$alert_threshold" ]; then
        log_error "错误数量超过阈值: $error_count > $alert_threshold"
        send_alert_notification "错误数量超过阈值: $error_count"
    fi
}

# 发送告警通知
send_alert_notification() {
    local alerts=("$@")
    
    log_warning "发送告警通知..."
    
    # 这里可以集成邮件、Slack、钉钉等通知方式
    # 目前只是记录到日志
    for alert in "${alerts[@]}"; do
        echo "$(date '+%Y-%m-%d %H:%M:%S') ALERT: $alert" >> alert.log
    done
    
    # 如果有webhook配置，可以发送HTTP通知
    if [ -n "$WEBHOOK_URL" ]; then
        local payload=$(cat << EOF
{
  "timestamp": "$(date -Iseconds)",
  "environment": "$ENVIRONMENT",
  "alerts": [$(printf '"%s",' "${alerts[@]}" | sed 's/,$//')]
}
EOF
)
        curl -X POST -H "Content-Type: application/json" -d "$payload" "$WEBHOOK_URL" 2>/dev/null || true
    fi
}

# 生成监控报告
generate_monitoring_report() {
    local environment=$1
    local output_dir=$2
    
    log_info "生成监控报告..."
    
    local report_file="$output_dir/monitoring_report_$(date +%Y%m%d_%H%M%S).html"
    
    # 收集监控数据
    local health_status=$(check_application_health "${MONITORING_ENDPOINTS["health"]}" "$environment" 2>/dev/null || echo "failed")
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>Law OA Go 监控报告 - $environment</title>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { background-color: #d4edda; border-color: #c3e6cb; }
        .error { background-color: #f8d7da; border-color: #f5c6cb; }
        .warning { background-color: #fff3cd; border-color: #ffeaa7; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Law OA Go 监控报告</h1>
        <p><strong>环境:</strong> $environment</p>
        <p><strong>生成时间:</strong> $(date)</p>
        <p><strong>健康状态:</strong> $health_status</p>
    </div>
    
    <div class="section">
        <h2>系统资源</h2>
        <table>
            <tr><th>指标</th><th>当前值</th><th>状态</th></tr>
            <tr><td>CPU使用率</td><td>$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%</td><td>正常</td></tr>
            <tr><td>内存使用率</td><td>$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')%</td><td>正常</td></tr>
            <tr><td>磁盘使用率</td><td>$(df -h / | awk 'NR==2{print $5}')</td><td>正常</td></tr>
        </table>
    </div>
    
    <div class="section">
        <h2>Docker容器状态</h2>
        <pre>$(docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "Docker不可用")</pre>
    </div>
    
    <div class="section">
        <h2>最近的监控数据</h2>
        <p>监控数据文件: $output_dir/metrics_*.json</p>
        <p>错误分析报告: $output_dir/error_analysis_*.txt</p>
    </div>
</body>
</html>
EOF
    
    log_success "监控报告已生成: $report_file"
}

# 持续监控
continuous_monitoring() {
    local environment=$1
    local interval=$2
    local duration=$3
    local output_dir=$4
    local alert_threshold=$5
    
    log_info "开始持续监控 (环境: $environment, 间隔: ${interval}s, 时长: ${duration}分钟)..."
    
    local end_time=$(date -d "+$duration minutes" +%s)
    
    while [ $(date +%s) -lt $end_time ]; do
        # 收集性能指标
        collect_performance_metrics "$environment" "$output_dir"
        
        # 分析错误日志
        analyze_error_logs "$environment" "$output_dir" "$alert_threshold"
        
        # 等待下次检查
        sleep $interval
    done
    
    log_info "持续监控完成"
}

# 主函数
main() {
    # 默认参数
    ENVIRONMENT="dev"
    INTERVAL=60
    DURATION=30
    OUTPUT_DIR="monitoring"
    ALERT_THRESHOLD=10
    GENERATE_REPORT=false
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -i|--interval)
                INTERVAL="$2"
                shift 2
                ;;
            -d|--duration)
                DURATION="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -a|--alert-threshold)
                ALERT_THRESHOLD="$2"
                shift 2
                ;;
            -r|--report)
                GENERATE_REPORT=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    log_info "启动部署监控 (环境: $ENVIRONMENT)..."
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"
    
    if [ "$GENERATE_REPORT" = true ]; then
        # 生成监控报告
        generate_monitoring_report "$ENVIRONMENT" "$OUTPUT_DIR"
    else
        # 持续监控
        continuous_monitoring "$ENVIRONMENT" "$INTERVAL" "$DURATION" "$OUTPUT_DIR" "$ALERT_THRESHOLD"
    fi
    
    log_success "监控完成"
}

# 运行主函数
main "$@"