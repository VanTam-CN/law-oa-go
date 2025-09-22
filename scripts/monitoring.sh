#!/bin/bash

# Law OA Go 项目部署监控脚本
# 实现实时监控、告警通知、性能分析和健康检查

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

log_monitor() {
    echo -e "${PURPLE}[MONITOR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_alert() {
    echo -e "${CYAN}[ALERT]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                     显示此帮助信息"
    echo "  -e, --env ENVIRONMENT          监控环境 (dev|staging|production)"
    echo "  -d, --duration MINUTES         监控持续时间 (分钟, 默认: 60)"
    echo "  -i, --interval SECONDS         监控间隔 (秒, 默认: 30)"
    echo "  -u, --url URL                  应用URL (默认: http://localhost:8080)"
    echo "  -H, --health-url URL           健康检查URL (默认: /health)"
    echo "  -m, --metrics-url URL          指标URL (默认: /metrics)"
    echo "  -o, --output FILE              输出文件 (默认: monitoring.log)"
    echo "  -n, --notifications           启用通知 (邮件/钉钉/企业微信)"
    echo "  -c, --config FILE              配置文件 (默认: monitoring.conf)"
    echo "  -t, --threshold-threshold      告警阈值 (默认: 5次连续失败)"
    echo "  -r, --report                  生成监控报告"
    echo "  --check-docker                检查Docker容器状态"
    echo "  --check-system                检查系统资源"
    echo "  --check-database              检查数据库连接"
    echo "  --check-redis                 检查Redis连接"
    echo "  --check-elasticsearch         检查Elasticsearch连接"
    echo "  --performance-analysis        性能分析"
    echo "  --auto-restart               自动重启失败服务"
    echo "  --dry-run                     仅显示监控计划"
    echo ""
    echo "示例:"
    echo "  $0 -e production -d 120 -i 15                  # 生产环境监控2小时"
    echo "  $0 -e staging -o staging_monitoring.log      # 测试环境监控到文件"
    echo "  $0 -e production --check-docker --performance-analysis  # 完整监控"
    echo "  $0 --dry-run -e production                    # 预览监控计划"
}

# 初始化变量
ENVIRONMENT="dev"
DURATION=60
INTERVAL=30
URL="http://localhost:8080"
HEALTH_URL="/health"
METRICS_URL="/metrics"
OUTPUT_FILE="monitoring.log"
NOTIFICATIONS=false
CONFIG_FILE="monitoring.conf"
ALERT_THRESHOLD=5
REPORT=false
CHECK_DOCKER=false
CHECK_SYSTEM=false
CHECK_DATABASE=false
CHECK_REDIS=false
CHECK_ELASTICSEARCH=false
PERFORMANCE_ANALYSIS=false
AUTO_RESTART=false
DRY_RUN=false

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
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        -i|--interval)
            INTERVAL="$2"
            shift 2
            ;;
        -u|--url)
            URL="$2"
            shift 2
            ;;
        -H|--health-url)
            HEALTH_URL="$2"
            shift 2
            ;;
        -m|--metrics-url)
            METRICS_URL="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_FILE="$2"
            shift 2
            ;;
        -n|--notifications)
            NOTIFICATIONS=true
            shift
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -t|--threshold)
            ALERT_THRESHOLD="$2"
            shift 2
            ;;
        -r|--report)
            REPORT=true
            shift
            ;;
        --check-docker)
            CHECK_DOCKER=true
            shift
            ;;
        --check-system)
            CHECK_SYSTEM=true
            shift
            ;;
        --check-database)
            CHECK_DATABASE=true
            shift
            ;;
        --check-redis)
            CHECK_REDIS=true
            shift
            ;;
        --check-elasticsearch)
            CHECK_ELASTICSEARCH=true
            shift
            ;;
        --performance-analysis)
            PERFORMANCE_ANALYSIS=true
            shift
            ;;
        --auto-restart)
            AUTO_RESTART=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_usage
            exit 1
            ;;
    esac
done

# 验证环境参数
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|production)$ ]]; then
    log_error "无效的环境参数: $ENVIRONMENT"
    echo "支持的环境: dev, staging, production"
    exit 1
fi

# 验证数值参数
if ! [[ "$DURATION" =~ ^[0-9]+$ ]] || [ "$DURATION" -lt 1 ]; then
    log_error "无效的监控持续时间: $DURATION"
    exit 1
fi

if ! [[ "$INTERVAL" =~ ^[0-9]+$ ]] || [ "$INTERVAL" -lt 1 ]; then
    log_error "无效的监控间隔: $INTERVAL"
    exit 1
fi

# 计算总监控次数
TOTAL_CHECKS=$((DURATION * 60 / INTERVAL))

# 初始化监控数据
declare -a health_check_history=()
declare -a response_time_history=()
declare -a error_rate_history=()
declare -a system_metrics_history=()

HEALTH_FAILURE_COUNT=0
TOTAL_REQUESTS=0
FAILED_REQUESTS=0
START_TIME=$(date +%s)
MONITOR_PID=$$

# 加载配置文件
load_config() {
    if [ -f "$CONFIG_FILE" ]; then
        log_info "加载配置文件: $CONFIG_FILE"
        source "$CONFIG_FILE"
    else
        log_info "配置文件不存在，使用默认配置"
    fi
}

# 健康检查
perform_health_check() {
    local check_url="$URL$HEALTH_URL"
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] curl -f -s $check_url"
        return 0
    fi
    
    local start_time=$(date +%s%N)
    local response_code=$(curl -f -s -w "%{http_code}" -o /dev/null "$check_url" 2>/dev/null || echo "000")
    local end_time=$(date +%s%N)
    
    local response_time=$(( (end_time - start_time) / 1000000 ))
    
    if [ "$response_code" = "200" ]; then
        echo "healthy|$response_time"
    else
        echo "unhealthy|$response_time"
    fi
}

# 性能指标检查
check_performance_metrics() {
    local metrics_url="$URL$METRICS_URL"
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] curl -s $metrics_url"
        return 0
    fi
    
    local metrics=$(curl -s "$metrics_url" 2>/dev/null || echo "")
    
    # 解析关键指标
    local requests_per_second=$(echo "$metrics" | grep "http_requests_total" | tail -1 | awk '{print $2}' || echo "0")
    local error_rate=$(echo "$metrics" | grep "http_requests_total.*status=~\"5..\"" | tail -1 | awk '{print $2}' || echo "0")
    local memory_usage=$(echo "$metrics" | grep "process_memory_bytes" | tail -1 | awk '{print $2}' || echo "0")
    local cpu_usage=$(echo "$metrics" | grep "process_cpu_seconds_total" | tail -1 | awk '{print $2}' || echo "0")
    
    echo "$requests_per_second|$error_rate|$memory_usage|$cpu_usage"
}

# Docker容器状态检查
check_docker_status() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] docker ps --filter \"name=law-oa\" --format \"{{.Names}}:{{.Status}}\""
        return 0
    fi
    
    local container_status=$(docker ps --filter "name=law-oa" --format "{{.Names}}:{{.Status}}")
    echo "$container_status"
}

# 系统资源检查
check_system_resources() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] # 检查CPU、内存、磁盘使用率"
        return 0
    fi
    
    local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//' || echo "0")
    local memory_usage=$(free -m | awk 'NR==2{printf "%.2f", $3*100/$2 }')
    local disk_usage=$(df -h / | awk 'NR==2{print $5}' | sed 's/%//')
    
    echo "$cpu_usage|$memory_usage|$disk_usage"
}

# 数据库连接检查
check_database_connection() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] mysql -h localhost -u root -ppassword -e \"SELECT 1\" law_oa"
        return 0
    fi
    
    if mysql -h localhost -u root -ppassword -e "SELECT 1" law_oa >/dev/null 2>&1; then
        echo "connected"
    else
        echo "disconnected"
    fi
}

# Redis连接检查
check_redis_connection() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] redis-cli ping"
        return 0
    fi
    
    if redis-cli ping >/dev/null 2>&1; then
        echo "connected"
    else
        echo "disconnected"
    fi
}

# Elasticsearch连接检查
check_elasticsearch_connection() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] curl -f http://localhost:9200/_cluster/health"
        return 0
    fi
    
    local health=$(curl -s http://localhost:9200/_cluster/health 2>/dev/null | jq -r '.status' || echo "unknown")
    echo "$health"
}

# 发送通知
send_notification() {
    local message="$1"
    local level="$2"
    
    if [ "$NOTIFICATIONS" = false ]; then
        return 0
    fi
    
    log_alert "发送通知: $message"
    
    # 这里可以集成邮件、钉钉、企业微信等通知方式
    # 目前先记录到日志
    echo "NOTIFICATION: [$level] $message" >> "$OUTPUT_FILE"
}

# 自动重启服务
auto_restart_service() {
    if [ "$AUTO_RESTART" = false ]; then
        return 0
    fi
    
    log_warning "检测到服务异常，尝试自动重启..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] docker-compose restart"
        return 0
    fi
    
    if docker-compose restart >/dev/null 2>&1; then
        log_success "服务重启成功"
        send_notification "Law OA Go 服务已自动重启" "warning"
    else
        log_error "服务重启失败"
        send_notification "Law OA Go 服务重启失败" "error"
    fi
}

# 生成监控报告
generate_monitoring_report() {
    local report_file="monitoring_report_$(date +%Y%m%d_%H%M%S).json"
    
    log_info "生成监控报告: $report_file"
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 生成监控报告: $report_file"
        return 0
    fi
    
    # 计算统计数据
    local total_checks=${#health_check_history[@]}
    local healthy_checks=$(printf '%s\n' "${health_check_history[@]}" | grep -c "healthy" || echo "0")
    local unhealthy_checks=$((total_checks - healthy_checks))
    
    local avg_response_time=0
    if [ ${#response_time_history[@]} -gt 0 ]; then
        local sum=0
        for time in "${response_time_history[@]}"; do
            sum=$((sum + time))
        done
        avg_response_time=$((sum / ${#response_time_history[@]}))
    fi
    
    local error_rate=0
    if [ "$TOTAL_REQUESTS" -gt 0 ]; then
        error_rate=$((FAILED_REQUESTS * 100 / TOTAL_REQUESTS))
    fi
    
    # 生成JSON报告
    cat > "$report_file" << EOF
{
  "monitoring_report": {
    "environment": "$ENVIRONMENT",
    "start_time": "$(date -d @$START_TIME)",
    "end_time": "$(date)",
    "duration_minutes": $DURATION,
    "interval_seconds": $INTERVAL,
    "total_checks": $total_checks,
    "url": "$URL"
  },
  "health_checks": {
    "total_checks": $total_checks,
    "healthy_checks": $healthy_checks,
    "unhealthy_checks": $unhealthy_checks,
    "health_percentage": $((healthy_checks * 100 / total_checks)),
    "consecutive_failures": $HEALTH_FAILURE_COUNT
  },
  "performance_metrics": {
    "total_requests": $TOTAL_REQUESTS,
    "failed_requests": $FAILED_REQUESTS,
    "error_rate": $error_rate,
    "average_response_time_ms": $avg_response_time,
    "max_response_time_ms": $(printf '%s\n' "${response_time_history[@]}" | sort -n | tail -1 || echo "0"),
    "min_response_time_ms": $(printf '%s\n' "${response_time_history[@]}" | sort -n | head -1 || echo "0")
  },
  "alerts": {
    "alert_threshold": $ALERT_THRESHOLD,
    "alerts_triggered": $([ "$HEALTH_FAILURE_COUNT" -ge "$ALERT_THRESHOLD" ] && echo "true" || echo "false"),
    "auto_restart_enabled": $AUTO_RESTART,
    "notifications_enabled": $NOTIFICATIONS
  },
  "recommendations": [
    $(if [ "$unhealthy_checks" -gt 0 ]; then echo "\"检查应用健康状态\""; fi)
    $(if [ "$error_rate" -gt 5 ]; then echo "\"检查错误率过高问题\""; fi)
    $(if [ "$avg_response_time" -gt 1000 ]; then echo "\"优化响应时间\""; fi)
    $(if [ "$HEALTH_FAILURE_COUNT" -ge "$ALERT_THRESHOLD" ]; then echo "\"考虑重启服务\""; fi)
  ]
}
EOF
    
    log_success "监控报告已生成: $report_file"
}

# 执行单次监控检查
perform_monitoring_check() {
    local check_number=$1
    
    log_monitor "执行第 $check_number/$TOTAL_CHECKS 次监控检查..."
    
    # 健康检查
    local health_result=$(perform_health_check)
    local health_status=$(echo "$health_result" | cut -d'|' -f1)
    local response_time=$(echo "$health_result" | cut -d'|' -f2)
    
    health_check_history+=("$health_status")
    response_time_history+=("$response_time")
    
    TOTAL_REQUESTS=$((TOTAL_REQUESTS + 1))
    
    if [ "$health_status" = "healthy" ]; then
        log_success "健康检查通过 (响应时间: ${response_time}ms)"
        HEALTH_FAILURE_COUNT=0
    else
        log_error "健康检查失败 (响应码: ${response_time})"
        FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
        HEALTH_FAILURE_COUNT=$((HEALTH_FAILURE_COUNT + 1))
        
        # 检查是否需要触发告警
        if [ "$HEALTH_FAILURE_COUNT" -ge "$ALERT_THRESHOLD" ]; then
            log_alert "连续 $ALERT_THRESHOLD 次健康检查失败，触发告警！"
            send_notification "Law OA Go 服务连续 $ALERT_THRESHOLD 次健康检查失败" "error"
            
            # 自动重启
            if [ "$AUTO_RESTART" = true ] && [ "$HEALTH_FAILURE_COUNT" -eq "$ALERT_THRESHOLD" ]; then
                auto_restart_service
            fi
        fi
    fi
    
    # 性能指标检查
    if [ "$PERFORMANCE_ANALYSIS" = true ]; then
        local metrics=$(check_performance_metrics)
        log_monitor "性能指标: RPS=$(echo "$metrics" | cut -d'|' -f1), 错误率=$(echo "$metrics" | cut -d'|' -f2)"
    fi
    
    # Docker状态检查
    if [ "$CHECK_DOCKER" = true ]; then
        local docker_status=$(check_docker_status)
        log_monitor "Docker状态: $docker_status"
    fi
    
    # 系统资源检查
    if [ "$CHECK_SYSTEM" = true ]; then
        local system_metrics=$(check_system_resources)
        local cpu_usage=$(echo "$system_metrics" | cut -d'|' -f1)
        local memory_usage=$(echo "$system_metrics" | cut -d'|' -f2)
        local disk_usage=$(echo "$system_metrics" | cut -d'|' -f3)
        
        log_monitor "系统资源: CPU=${cpu_usage}%, 内存=${memory_usage}%, 磁盘=${disk_usage}%"
        
        # 系统资源告警
        if (( $(echo "$cpu_usage > 80" | bc -l) )); then
            log_warning "CPU使用率过高: ${cpu_usage}%"
            send_notification "CPU使用率过高: ${cpu_usage}%" "warning"
        fi
        
        if (( $(echo "$memory_usage > 80" | bc -l) )); then
            log_warning "内存使用率过高: ${memory_usage}%"
            send_notification "内存使用率过高: ${memory_usage}%" "warning"
        fi
        
        if [ "$disk_usage" -gt 80 ]; then
            log_warning "磁盘使用率过高: ${disk_usage}%"
            send_notification "磁盘使用率过高: ${disk_usage}%" "warning"
        fi
    fi
    
    # 数据库连接检查
    if [ "$CHECK_DATABASE" = true ]; then
        local db_status=$(check_database_connection)
        log_monitor "数据库连接: $db_status"
        
        if [ "$db_status" = "disconnected" ]; then
            log_error "数据库连接失败"
            send_notification "数据库连接失败" "error"
        fi
    fi
    
    # Redis连接检查
    if [ "$CHECK_REDIS" = true ]; then
        local redis_status=$(check_redis_connection)
        log_monitor "Redis连接: $redis_status"
        
        if [ "$redis_status" = "disconnected" ]; then
            log_error "Redis连接失败"
            send_notification "Redis连接失败" "error"
        fi
    fi
    
    # Elasticsearch连接检查
    if [ "$CHECK_ELASTICSEARCH" = true ]; then
        local es_status=$(check_elasticsearch_connection)
        log_monitor "Elasticsearch状态: $es_status"
        
        if [ "$es_status" != "green" ] && [ "$es_status" != "yellow" ]; then
            log_error "Elasticsearch状态异常: $es_status"
            send_notification "Elasticsearch状态异常: $es_status" "error"
        fi
    fi
}

# 清理函数
cleanup() {
    log_info "监控会话结束"
    
    if [ "$REPORT" = true ]; then
        generate_monitoring_report
    fi
    
    # 发送监控结束通知
    if [ "$NOTIFICATIONS" = true ]; then
        send_notification "Law OA Go 监控会话结束" "info"
    fi
}

# 信号处理
trap cleanup EXIT

# 主函数
main() {
    log_info "开始 $ENVIRONMENT 环境监控..."
    log_info "监控URL: $URL"
    log_info "监控时长: $DURATION 分钟，间隔: $INTERVAL 秒"
    log_info "输出文件: $OUTPUT_FILE"
    
    # 加载配置
    load_config
    
    # 创建输出目录
    mkdir -p "$(dirname "$OUTPUT_FILE")"
    
    # 创建输出文件
    echo "# Law OA Go 监控日志" > "$OUTPUT_FILE"
    echo "# 环境: $ENVIRONMENT" >> "$OUTPUT_FILE"
    echo "# 开始时间: $(date)" >> "$OUTPUT_FILE"
    echo "# 监控配置: 时长=${DURATION}分钟, 间隔=${INTERVAL}秒" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    # 监控循环
    for ((i=1; i<=TOTAL_CHECKS; i++)); do
        perform_monitoring_check $i >> "$OUTPUT_FILE"
        
        # 如果不是最后一次检查，等待间隔
        if [ "$i" -lt "$TOTAL_CHECKS" ]; then
            sleep $INTERVAL
        fi
    done
    
    log_success "监控完成"
}

# 运行主函数
main "$@"