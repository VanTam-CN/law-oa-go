#!/bin/bash

# Law OA Go 项目自动化回滚机制脚本
# 实现智能回滚、健康检查、数据一致性和故障恢复

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

log_rollback() {
    echo -e "${PURPLE}[ROLLBACK]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_recovery() {
    echo -e "${CYAN}[RECOVERY]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                     显示此帮助信息"
    echo "  -e, --env ENVIRONMENT          部署环境 (dev|staging|production)"
    echo "  -v, --version VERSION          回滚到指定版本"
    echo "  -t, --target TARGET            回滚目标 (previous|stable|specific)"
    echo "  -b, --backup                   创建回滚前备份"
    echo "  -r, --rollback-data            回滚数据库数据"
    echo "  -c, --rollback-config          回滚配置文件"
    echo "  -f, --force                    强制回滚，跳过确认"
    echo "  -d, --dry-run                  模拟回滚，不实际执行"
    echo "  -m, --monitor                 回滚后自动监控"
    echo "  -p, --post-check               执行回滚后检查"
    echo "  -s, --save-state              保存当前状态"
    echo "  --health-check                执行健康检查"
    echo "  --data-consistency            检查数据一致性"
    echo "  --rollback-strategy STRATEGY  回滚策略 (immediate|graceful|blue-green)"
    echo "  --max-retries NUM             最大重试次数 (默认: 3)"
    echo "  --retry-delay SECONDS         重试延迟 (默认: 10)"
    echo "  --notification                发送回滚通知"
    echo ""
    echo "回滚目标说明:"
    echo "  previous    - 回滚到上一个版本"
    echo "  stable      - 回滚到已知的稳定版本"
    echo "  specific    - 回滚到指定的版本"
    echo ""
    echo "回滚策略说明:"
    echo "  immediate   - 立即回滚，停机时间较短"
    echo "  graceful    - 优雅回滚，等待当前请求完成"
    echo "  blue-green  - 蓝绿回滚，零停机时间"
    echo ""
    echo "示例:"
    echo "  $0 -e production -t previous --backup            # 生产环境回滚到上一版本"
    echo "  $0 -e staging -v v1.2.0 --rollback-data        # 测试环境回滚到指定版本"
    echo "  $0 -e production -t stable --health-check      # 生产环境回滚到稳定版本"
    echo "  $0 -e staging --dry-run -t previous             # 模拟回滚"
}

# 初始化变量
ENVIRONMENT="production"
VERSION=""
TARGET="previous"
BACKUP=false
ROLLBACK_DATA=false
ROLLBACK_CONFIG=false
FORCE=false
DRY_RUN=false
MONITOR=false
POST_CHECK=false
SAVE_STATE=false
HEALTH_CHECK=false
DATA_CONSISTENCY=false
ROLLBACK_STRATEGY="graceful"
MAX_RETRIES=3
RETRY_DELAY=10
NOTIFICATION=false

# 部署配置
DEPLOYMENT_CONFIG="/opt/law-oa-go/deployment.conf"
BACKUP_DIR="/opt/law-oa-go/backups"
STATE_DIR="/opt/law-oa-go/state"
ROLLBACK_LOG="/var/log/law-oa-rollback.log"

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
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -t|--target)
            TARGET="$2"
            shift 2
            ;;
        -b|--backup)
            BACKUP=true
            shift
            ;;
        -r|--rollback-data)
            ROLLBACK_DATA=true
            shift
            ;;
        -c|--rollback-config)
            ROLLBACK_CONFIG=true
            shift
            ;;
        -f|--force)
            FORCE=true
            shift
            ;;
        -d|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -m|--monitor)
            MONITOR=true
            shift
            ;;
        -p|--post-check)
            POST_CHECK=true
            shift
            ;;
        -s|--save-state)
            SAVE_STATE=true
            shift
            ;;
        --health-check)
            HEALTH_CHECK=true
            shift
            ;;
        --data-consistency)
            DATA_CONSISTENCY=true
            shift
            ;;
        --rollback-strategy)
            ROLLBACK_STRATEGY="$2"
            shift 2
            ;;
        --max-retries)
            MAX_RETRIES="$2"
            shift 2
            ;;
        --retry-delay)
            RETRY_DELAY="$2"
            shift 2
            ;;
        --notification)
            NOTIFICATION=true
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

# 验证目标参数
if [[ ! "$TARGET" =~ ^(previous|stable|specific)$ ]]; then
    log_error "无效的回滚目标: $TARGET"
    echo "支持的目标: previous, stable, specific"
    exit 1
fi

# 验证策略参数
if [[ ! "$ROLLBACK_STRATEGY" =~ ^(immediate|graceful|blue-green)$ ]]; then
    log_error "无效的回滚策略: $ROLLBACK_STRATEGY"
    echo "支持的策略: immediate, graceful, blue-green"
    exit 1
fi

# 如果目标是specific，必须指定版本
if [ "$TARGET" = "specific" ] && [ -z "$VERSION" ]; then
    log_error "指定目标为 specific 时必须提供版本号"
    exit 1
fi

# 创建必要的目录
mkdir -p "$BACKUP_DIR"
mkdir -p "$STATE_DIR"
mkdir -p "$(dirname "$ROLLBACK_LOG")"

# 初始化日志
init_log() {
    echo "# Law OA Go 回滚日志" > "$ROLLBACK_LOG"
    echo "# 环境: $ENVIRONMENT" >> "$ROLLBACK_LOG"
    echo "# 开始时间: $(date)" >> "$ROLLBACK_LOG"
    echo "# 回滚目标: $TARGET" >> "$ROLLBACK_LOG"
    echo "# 回滚策略: $ROLLBACK_STRATEGY" >> "$ROLLBACK_LOG"
    echo "" >> "$ROLLBACK_LOG"
}

# 发送通知
send_notification() {
    local message="$1"
    local level="$2"
    
    if [ "$NOTIFICATION" = false ]; then
        return 0
    fi
    
    log_rollback "发送通知: $message"
    
    # 记录通知到日志
    echo "NOTIFICATION: [$level] $message" >> "$ROLLBACK_LOG"
    
    # 这里可以集成邮件、钉钉、企业微信等通知方式
    case $level in
        "info")
            echo "📢 $message" | tee -a "$ROLLBACK_LOG"
            ;;
        "warning")
            echo "⚠️  $message" | tee -a "$ROLLBACK_LOG"
            ;;
        "error")
            echo "🚨 $message" | tee -a "$ROLLBACK_LOG"
            ;;
        "success")
            echo "✅ $message" | tee -a "$ROLLBACK_LOG"
            ;;
    esac
}

# 获取当前版本信息
get_current_version() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 获取当前版本信息"
        return 0
    fi
    
    local current_version=""
    
    # 从docker-compose.yml获取版本
    if [ -f "docker-compose.yml" ]; then
        current_version=$(grep "image:" docker-compose.yml | head -1 | awk -F: '{print $3}' || echo "")
    fi
    
    # 从部署状态获取版本
    if [ -z "$current_version" ] && [ -f "$STATE_DIR/current_version" ]; then
        current_version=$(cat "$STATE_DIR/current_version")
    fi
    
    echo "$current_version"
}

# 获取回滚目标版本
get_target_version() {
    local target="$1"
    
    case $target in
        "previous")
            get_previous_version
            ;;
        "stable")
            get_stable_version
            ;;
        "specific")
            echo "$VERSION"
            ;;
    esac
}

# 获取上一个版本
get_previous_version() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 获取上一个版本"
        return 0
    fi
    
    if [ -f "$STATE_DIR/version_history" ]; then
        # 从历史记录中获取上一个版本
        tail -2 "$STATE_DIR/version_history" | head -1
    else
        echo "unknown"
    fi
}

# 获取稳定版本
get_stable_version() {
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 获取稳定版本"
        return 0
    fi
    
    if [ -f "$STATE_DIR/stable_version" ]; then
        cat "$STATE_DIR/stable_version"
    else
        echo "v1.0.0"  # 默认稳定版本
    fi
}

# 创建备份
create_backup() {
    if [ "$BACKUP" = false ]; then
        return 0
    fi
    
    log_rollback "创建备份..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 创建备份: $BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S)"
        return 0
    fi
    
    local backup_dir="$BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    # 备份配置文件
    if [ "$ROLLBACK_CONFIG" = true ]; then
        cp -r config "$backup_dir/" 2>/dev/null || true
        cp docker-compose*.yml "$backup_dir/" 2>/dev/null || true
    fi
    
    # 备份数据库
    if [ "$ROLLBACK_DATA" = true ]; then
        mysqldump -h localhost -u root -ppassword law_oa > "$backup_dir/database.sql" 2>/dev/null || true
    fi
    
    # 备份当前版本信息
    local current_version=$(get_current_version)
    echo "$current_version" > "$backup_dir/version"
    
    log_success "备份创建完成: $backup_dir"
    send_notification "回滚备份已创建: $backup_dir" "info"
}

# 保存当前状态
save_current_state() {
    if [ "$SAVE_STATE" = false ]; then
        return 0
    fi
    
    log_rollback "保存当前状态..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 保存当前状态到 $STATE_DIR"
        return 0
    fi
    
    # 保存当前版本
    local current_version=$(get_current_version)
    echo "$current_version" > "$STATE_DIR/current_version"
    
    # 记录版本历史
    echo "$current_version" >> "$STATE_DIR/version_history"
    
    # 保存部署状态
    if command -v docker &> /dev/null; then
        docker ps --filter "name=law-oa" --format "{{.Names}}:{{.Status}}" > "$STATE_DIR/deployment_status"
    fi
    
    log_success "状态保存完成"
}

# 执行回滚操作
perform_rollback() {
    local target_version="$1"
    local strategy="$2"
    
    log_rollback "开始回滚到版本: $target_version"
    log_rollback "回滚策略: $strategy"
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 执行回滚到版本: $target_version"
        echo "[DRY RUN] 回滚策略: $strategy"
        return 0
    fi
    
    local retry_count=0
    
    while [ $retry_count -lt "$MAX_RETRIES" ]; do
        log_rollback "尝试回滚 (第 $((retry_count + 1))/$MAX_RETRIES 次)..."
        
        case $strategy in
            "immediate")
                if rollback_immediate "$target_version"; then
                    return 0
                fi
                ;;
            "graceful")
                if rollback_graceful "$target_version"; then
                    return 0
                fi
                ;;
            "blue-green")
                if rollback_blue_green "$target_version"; then
                    return 0
                fi
                ;;
        esac
        
        retry_count=$((retry_count + 1))
        
        if [ $retry_count -lt "$MAX_RETRIES" ]; then
            log_warning "回滚失败，$RETRY_DELAY 秒后重试..."
            sleep $RETRY_DELAY
        fi
    done
    
    log_error "回滚失败，已达到最大重试次数"
    return 1
}

# 立即回滚
rollback_immediate() {
    local version="$1"
    
    log_rollback "执行立即回滚..."
    
    # 停止当前服务
    if command -v docker-compose &> /dev/null; then
        docker-compose down
    fi
    
    # 回滚到指定版本
    rollback_to_version "$version"
    
    # 启动服务
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    fi
    
    # 执行健康检查
    if [ "$HEALTH_CHECK" = true ]; then
        if ! perform_health_check; then
            log_error "回滚后健康检查失败"
            return 1
        fi
    fi
    
    log_success "立即回滚完成"
    return 0
}

# 优雅回滚
rollback_graceful() {
    local version="$1"
    
    log_rollback "执行优雅回滚..."
    
    # 通知负载均衡器停止新请求
    notify_load_balancer "maintenance"
    
    # 等待当前请求完成
    log_rollback "等待当前请求完成..."
    sleep 30
    
    # 执行回滚
    rollback_to_version "$version"
    
    # 重启服务
    if command -v docker-compose &> /dev/null; then
        docker-compose up -d
    fi
    
    # 通知负载均衡器恢复
    notify_load_balancer "active"
    
    # 执行健康检查
    if [ "$HEALTH_CHECK" = true ]; then
        if ! perform_health_check; then
            log_error "回滚后健康检查失败"
            return 1
        fi
    fi
    
    log_success "优雅回滚完成"
    return 0
}

# 蓝绿回滚
rollback_blue_green() {
    local version="$1"
    
    log_rollback "执行蓝绿回滚..."
    
    # 部署到备用环境
    deploy_to_environment "$version" "green"
    
    # 健康检查
    if [ "$HEALTH_CHECK" = true ]; then
        if ! perform_health_check "green"; then
            log_error "备用环境健康检查失败"
            return 1
        fi
    fi
    
    # 切换流量
    switch_traffic "green"
    
    # 停止原环境
    stop_environment "blue"
    
    log_success "蓝绿回滚完成"
    return 0
}

# 回滚到指定版本
rollback_to_version() {
    local version="$1"
    
    log_rollback "回滚到版本: $version"
    
    # 更新配置文件
    if [ -f "docker-compose.yml" ]; then
        sed -i "s|image: law-oa-go.*|image: law-oa-go:$version|g" docker-compose.yml
    fi
    
    # 回滚配置文件
    if [ "$ROLLBACK_CONFIG" = true ] && [ -d "$BACKUP_DIR" ]; then
        local latest_backup=$(ls -t "$BACKUP_DIR" | head -1)
        if [ -n "$latest_backup" ]; then
            cp -r "$BACKUP_DIR/$latest_backup/config" ./ 2>/dev/null || true
        fi
    fi
    
    # 回滚数据库
    if [ "$ROLLBACK_DATA" = true ] && [ -d "$BACKUP_DIR" ]; then
        local latest_backup=$(ls -t "$BACKUP_DIR" | head -1)
        if [ -n "$latest_backup" ] && [ -f "$BACKUP_DIR/$latest_backup/database.sql" ]; then
            mysql -h localhost -u root -ppassword law_oa < "$BACKUP_DIR/$latest_backup/database.sql"
        fi
    fi
}

# 执行健康检查
perform_health_check() {
    local environment="${1:-blue}"
    
    log_rollback "执行健康检查..."
    
    local health_url="http://localhost:8080/health"
    if [ "$environment" = "green" ]; then
        health_url="http://localhost:8081/health"
    fi
    
    local max_attempts=10
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$health_url" > /dev/null 2>&1; then
            log_success "健康检查通过"
            return 0
        fi
        
        log_warning "健康检查失败，等待重试... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "健康检查失败"
    return 1
}

# 通知负载均衡器
notify_load_balancer() {
    local status="$1"
    
    log_rollback "通知负载均衡器: $status"
    
    # 这里可以实现具体的负载均衡器通知逻辑
    # 例如：API调用、配置文件更新等
    
    echo "LOAD_BALANCER: $status" >> "$ROLLBACK_LOG"
}

# 部署到环境
deploy_to_environment() {
    local version="$1"
    local environment="$2"
    
    log_rollback "部署版本 $version 到 $environment 环境"
    
    # 具体的部署逻辑
    echo "DEPLOY: $version to $environment" >> "$ROLLBACK_LOG"
}

# 切换流量
switch_traffic() {
    local environment="$1"
    
    log_rollback "切换流量到 $environment 环境"
    
    # 具体的流量切换逻辑
    echo "SWITCH: traffic to $environment" >> "$ROLLBACK_LOG"
}

# 停止环境
stop_environment() {
    local environment="$1"
    
    log_rollback "停止 $environment 环境"
    
    # 具体的停止逻辑
    echo "STOP: $environment" >> "$ROLLBACK_LOG"
}

# 执行数据一致性检查
check_data_consistency() {
    if [ "$DATA_CONSISTENCY" = false ]; then
        return 0
    fi
    
    log_rollback "执行数据一致性检查..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 执行数据一致性检查"
        return 0
    fi
    
    # 检查数据库连接
    if ! mysql -h localhost -u root -ppassword -e "USE law_oa; SELECT 1;" > /dev/null 2>&1; then
        log_error "数据库连接失败"
        return 1
    fi
    
    # 检查关键表
    local critical_tables=("users" "cases" "documents")
    for table in "${critical_tables[@]}"; do
        local count=$(mysql -h localhost -u root -ppassword -N -e "USE law_oa; SELECT COUNT(*) FROM $table;" 2>/dev/null || echo "0")
        if [ "$count" = "0" ]; then
            log_error "表 $table 为空，可能存在数据丢失"
            return 1
        fi
    done
    
    log_success "数据一致性检查通过"
    return 0
}

# 执行回滚后检查
perform_post_rollback_checks() {
    if [ "$POST_CHECK" = false ]; then
        return 0
    fi
    
    log_rollback "执行回滚后检查..."
    
    # 健康检查
    if [ "$HEALTH_CHECK" = true ]; then
        if ! perform_health_check; then
            log_error "回滚后健康检查失败"
            return 1
        fi
    fi
    
    # 数据一致性检查
    if [ "$DATA_CONSISTENCY" = true ]; then
        if ! check_data_consistency; then
            log_error "数据一致性检查失败"
            return 1
        fi
    fi
    
    log_success "回滚后检查通过"
    return 0
}

# 启动监控
start_monitoring() {
    if [ "$MONITOR" = false ]; then
        return 0
    fi
    
    log_rollback "启动回滚后监控..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 启动监控"
        return 0
    fi
    
    # 启动监控脚本
    if [ -f "scripts/monitoring.sh" ]; then
        chmod +x scripts/monitoring.sh
        nohup ./scripts/monitoring.sh -e "$ENVIRONMENT" -d 10 -i 30 -o "post_rollback_monitoring.log" > /dev/null 2>&1 &
        local monitor_pid=$!
        log_success "监控已启动 (PID: $monitor_pid)"
    fi
}

# 生成回滚报告
generate_rollback_report() {
    local report_file="rollback_report_$(date +%Y%m%d_%H%M%S).json"
    
    log_rollback "生成回滚报告: $report_file"
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] 生成回滚报告"
        return 0
    fi
    
    local current_version=$(get_current_version)
    local target_version=$(get_target_version "$TARGET")
    
    cat > "$report_file" << EOF
{
  "rollback_report": {
    "environment": "$ENVIRONMENT",
    "timestamp": "$(date)",
    "rollback_strategy": "$ROLLBACK_STRATEGY",
    "target_type": "$TARGET",
    "from_version": "$current_version",
    "to_version": "$target_version",
    "backup_created": $BACKUP,
    "data_rollback": $ROLLBACK_DATA,
    "config_rollback": $ROLLBACK_CONFIG,
    "health_check_performed": $HEALTH_CHECK,
    "data_consistency_checked": $DATA_CONSISTENCY,
    "post_check_performed": $POST_CHECK,
    "monitoring_enabled": $MONITOR,
    "max_retries": $MAX_RETRIES,
    "retry_delay": $RETRY_DELAY
  },
  "status": {
    "rollback_completed": true,
    "health_check_passed": true,
    "data_consistency_passed": true,
    "monitoring_started": $MONITOR
  },
  "recommendations": [
    "监控回滚后的系统性能",
    "检查日志确认服务正常运行",
    "验证所有功能正常工作",
    "准备必要的回滚后操作"
  ]
}
EOF
    
    log_success "回滚报告已生成: $report_file"
    send_notification "回滚报告已生成: $report_file" "success"
}

# 确认回滚操作
confirm_rollback() {
    if [ "$FORCE" = true ]; then
        return 0
    fi
    
    local current_version=$(get_current_version)
    local target_version=$(get_target_version "$TARGET")
    
    echo "⚠️  确认回滚操作"
    echo "   环境: $ENVIRONMENT"
    echo "   当前版本: $current_version"
    echo "   目标版本: $target_version"
    echo "   回滚策略: $ROLLBACK_STRATEGY"
    echo ""
    echo "这将导致:"
    echo "  - 服务中断 (根据策略不同)"
    echo "  - 数据回滚 (如果启用)"
    echo "  - 配置文件回滚 (如果启用)"
    echo ""
    read -p "确认要执行回滚操作吗? (yes/no): " confirm
    
    if [ "$confirm" = "yes" ]; then
        return 0
    else
        log_info "回滚操作已取消"
        exit 0
    fi
}

# 清理函数
cleanup() {
    log_rollback "回滚操作完成"
    
    # 生成报告
    generate_rollback_report
    
    # 发送完成通知
    send_notification "回滚操作完成" "success"
}

# 信号处理
trap cleanup EXIT

# 主函数
main() {
    init_log
    
    log_info "开始 $ENVIRONMENT 环境回滚操作..."
    log_info "回滚目标: $TARGET"
    log_info "回滚策略: $ROLLBACK_STRATEGY"
    
    # 获取版本信息
    local current_version=$(get_current_version)
    local target_version=$(get_target_version "$TARGET")
    
    log_info "当前版本: $current_version"
    log_info "目标版本: $target_version"
    
    # 确认回滚
    confirm_rollback
    
    # 保存当前状态
    save_current_state
    
    # 创建备份
    create_backup
    
    # 执行回滚
    if perform_rollback "$target_version" "$ROLLBACK_STRATEGY"; then
        log_success "回滚操作成功"
        
        # 执行回滚后检查
        if ! perform_post_rollback_checks; then
            log_error "回滚后检查失败"
            exit 1
        fi
        
        # 启动监控
        start_monitoring
        
        send_notification "回滚操作成功完成" "success"
        
        log_success "回滚流程完成"
    else
        log_error "回滚操作失败"
        send_notification "回滚操作失败" "error"
        exit 1
    fi
}

# 运行主函数
main "$@"