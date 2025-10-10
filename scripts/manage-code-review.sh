#!/bin/bash

# 代码审查工具管理脚本
# Law OA Go 项目 - 统一的代码审查工具管理

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.code-review.yml"
DEFAULT_PROFILE="basic"

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[CODE-REVIEW-MANAGER]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
代码审查工具管理脚本 - Law OA Go 项目

用法: $0 [命令] [选项]

命令:
    start [profile]       启动代码审查工具
    stop [profile]        停止代码审查工具
    restart [profile]     重启代码审查工具
    status               查看服务状态
    logs [service]        查看日志
    run-analysis         运行代码分析
    generate-report      生成质量报告
    install-tools        安装依赖工具
    clean                清理所有数据
    backup               备份配置和数据
    restore              恢复配置和数据

配置文件:
    $COMPOSE_FILE

可用配置:
    basic          - SonarQube + 基础工具 (默认)
    full           - 所有服务完整配置
    tools          - 仅代码审查工具
    frontend-tools - 仅前端工具
    monitoring     - 仅监控面板
    reports        - 仅报告服务
    proxy          - 仅代理服务

示例:
    $0 start basic                 # 启动基础服务
    $0 start full                  # 启动完整环境
    $0 run-analysis                # 运行代码分析
    $0 logs sonarqube              # 查看 SonarQube 日志
    $0 status                     # 查看所有服务状态

EOF
}

# 检查 Docker 和 Docker Compose
check_dependencies() {
    log "检查依赖..."

    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker 未安装"
        exit 1
    fi

    if ! command -v docker-compose >/dev/null 2>&1; then
        log_error "Docker Compose 未安装"
        exit 1
    fi

    # 检查 Docker 守护进程
    if ! docker info >/dev/null 2>&1; then
        log_error "Docker 守护进程未运行"
        exit 1
    fi

    log_success "依赖检查通过"
}

# 验证配置文件
validate_config() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        log_error "未找到 Docker Compose 配置文件: $COMPOSE_FILE"
        exit 1
    fi

    # 检查必要的环境文件
    if [ ! -f "$PROJECT_ROOT/.golangci.yml" ]; then
        log_warning "未找到 .golangci.yml 配置文件"
    fi

    if [ ! -f "$PROJECT_ROOT/sonar-project.properties" ]; then
        log_warning "未找到 sonar-project.properties 配置文件"
    fi
}

# 启动服务
start_services() {
    local profile=${1:-$DEFAULT_PROFILE}
    log "启动代码审查工具 (配置: $profile)..."

    validate_config

    case $profile in
        basic)
            log_info "启动基础配置 (SonarQube + 数据库)..."
            docker-compose -f "$COMPOSE_FILE" up -d sonarqube sonarqube-db
            ;;
        full)
            log_info "启动完整配置..."
            docker-compose -f "$COMPOSE_FILE" --profile tools --profile frontend-tools --profile monitoring --profile reports --profile proxy up -d
            ;;
        tools)
            log_info "启动代码审查工具..."
            docker-compose -f "$COMPOSE_FILE" --profile tools up -d code-review-tools
            ;;
        frontend-tools)
            log_info "启动前端工具..."
            docker-compose -f "$COMPOSE_FILE" --profile frontend-tools up -d frontend-review-tools
            ;;
        monitoring)
            log_info "启动监控面板..."
            docker-compose -f "$COMPOSE_FILE" --profile monitoring up -d monitoring-dashboard
            ;;
        reports)
            log_info "启动报告服务..."
            docker-compose -f "$COMPOSE_FILE" --profile reports up -d quality-report-web
            ;;
        proxy)
            log_info "启动代理服务..."
            docker-compose -f "$COMPOSE_FILE" --profile proxy up -d code-review-nginx
            ;;
        *)
            log_error "未知配置: $profile"
            show_help
            exit 1
            ;;
    esac

    log_success "服务启动完成"
    show_access_urls
}

# 停止服务
stop_services() {
    local profile=${1:-$DEFAULT_PROFILE}
    log "停止代码审查工具 (配置: $profile)..."

    case $profile in
        basic)
            docker-compose -f "$COMPOSE_FILE" stop sonarqube sonarqube-db
            ;;
        full)
            docker-compose -f "$COMPOSE_FILE" down
            ;;
        tools)
            docker-compose -f "$COMPOSE_FILE" stop code-review-tools
            ;;
        frontend-tools)
            docker-compose -f "$COMPOSE_FILE" stop frontend-review-tools
            ;;
        monitoring)
            docker-compose -f "$COMPOSE_FILE" stop monitoring-dashboard
            ;;
        reports)
            docker-compose -f "$COMPOSE_FILE" stop quality-report-web
            ;;
        proxy)
            docker-compose -f "$COMPOSE_FILE" stop code-review-nginx
            ;;
        *)
            log_error "未知配置: $profile"
            show_help
            exit 1
            ;;
    esac

    log_success "服务已停止"
}

# 重启服务
restart_services() {
    local profile=${1:-$DEFAULT_PROFILE}
    log "重启代码审查工具 (配置: $profile)..."
    stop_services "$profile"
    sleep 2
    start_services "$profile"
}

# 查看服务状态
show_status() {
    log "查看服务状态..."
    echo ""

    if [ -f "$COMPOSE_FILE" ]; then
        echo "=== Docker Compose 服务状态 ==="
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""
    fi

    echo "=== 端口使用情况 ==="
    echo "SonarQube:       http://localhost:9000"
    echo "Monitoring:      http://localhost:3001"
    echo "Quality Reports: http://localhost:8081"
    echo "Nginx Proxy:     http://localhost:8888"
    echo ""

    echo "=== 健康检查 ==="
    check_sonarqube_health
    check_monitoring_health
    check_reports_health
}

# 检查 SonarQube 健康状态
check_sonarqube_health() {
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "law-oa-sonarqube"; then
        if curl -s http://localhost:9000/api/system/status | grep -q "UP"; then
            echo "✅ SonarQube: 健康运行"
        else
            echo "⚠️ SonarQube: 容器运行但服务未就绪"
        fi
    else
        echo "❌ SonarQube: 未运行"
    fi
}

# 检查监控服务健康状态
check_monitoring_health() {
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "law-oa-monitoring-dashboard"; then
        if curl -s http://localhost:3001/api/health | grep -q "ok"; then
            echo "✅ Monitoring: 健康运行"
        else
            echo "⚠️ Monitoring: 容器运行但服务未就绪"
        fi
    else
        echo "❌ Monitoring: 未运行"
    fi
}

# 检查报告服务健康状态
check_reports_health() {
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "law-oa-quality-report-web"; then
        if curl -s http://localhost:8081/health | grep -q "healthy"; then
            echo "✅ Quality Reports: 健康运行"
        else
            echo "⚠️ Quality Reports: 容器运行但服务未就绪"
        fi
    else
        echo "❌ Quality Reports: 未运行"
    fi
}

# 显示访问地址
show_access_urls() {
    echo ""
    echo "=== 访问地址 ==="
    echo "SonarQube:       http://localhost:9000"
    echo "Monitoring:      http://localhost:3001"
    echo "Quality Reports: http://localhost:8081"
    echo "Nginx Proxy:     http://localhost:8888"
    echo ""
    echo "=== 默认凭据 ==="
    echo "SonarQube:       admin / admin"
    echo "Monitoring:      admin / admin123"
    echo ""
}

# 查看日志
show_logs() {
    local service=${1:-""}

    if [ -z "$service" ]; then
        log "显示所有服务日志..."
        docker-compose -f "$COMPOSE_FILE" logs -f --tail=100
    else
        log "显示 $service 服务日志..."
        docker-compose -f "$COMPOSE_FILE" logs -f --tail=100 "$service"
    fi
}

# 运行代码分析
run_analysis() {
    log "运行代码分析..."

    # 检查工具容器是否运行
    if docker ps --format "table {{.Names}}" | grep -q "law-oa-code-review-tools"; then
        log_info "在工具容器中运行分析..."

        docker exec law-oa-code-review-tools /workspace/scripts/code-review-tools.sh --all --generate-report
    else
        log_info "直接运行分析脚本..."
        "$SCRIPT_DIR/code-review-tools.sh" --all --generate-report
    fi

    log_success "代码分析完成"
}

# 生成质量报告
generate_report() {
    log "生成质量报告..."

    # 检查报告服务是否运行
    if docker ps --format "table {{.Names}}" | grep -q "law-oa-quality-report-web"; then
        log_info "触发报告服务生成报告..."
        curl -X POST http://localhost:8081/api/trigger-analysis
    else
        log_info "直接运行报告生成..."
        "$SCRIPT_DIR/code-review-tools.sh" --generate-report
    fi

    log_success "质量报告生成完成"
}

# 安装依赖工具
install_tools() {
    log "安装依赖工具..."

    if ! command -v golangci-lint >/dev/null 2>&1; then
        log_info "安装 golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi

    if ! command -v eslint >/dev/null 2>&1; then
        log_info "安装 eslint..."
        npm install -g eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
    fi

    if ! command -v sonar-scanner >/dev/null 2>&1; then
        log_info "SonarScanner 需要手动安装，请参考官方文档"
    fi

    log_success "工具安装完成"
}

# 清理数据
clean_all() {
    log_warning "⚠️  这将删除所有容器、镜像和数据！"
    read -p "确认继续? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "清理所有数据..."

        # 停止并删除容器
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans

        # 删除相关镜像
        docker images | grep law-oa | awk '{print $3}' | xargs -r docker rmi -f

        # 删除本地报告
        rm -rf "$PROJECT_ROOT/reports"

        log_success "清理完成"
    else
        log_info "操作已取消"
    fi
}

# 备份配置和数据
backup_data() {
    local backup_dir="$PROJECT_ROOT/backups/code-review"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$backup_dir/code-review-backup-$timestamp.tar.gz"

    log "备份数据到 $backup_file..."

    mkdir -p "$backup_dir"

    # 创建备份
    tar czf "$backup_file" \
        -C "$PROJECT_ROOT" \
        docker-compose.code-review.yml \
        .golangci.yml \
        sonar-project.properties \
        scripts/ \
        reports/ 2>/dev/null || true

    log_success "备份完成: $backup_file"

    # 清理旧备份 (保留最近7天)
    find "$backup_dir" -name "code-review-backup-*.tar.gz" -mtime +7 -delete
}

# 恢复配置和数据
restore_data() {
    local backup_file=$1

    if [ -z "$backup_file" ]; then
        log_error "请指定备份文件路径"
        exit 1
    fi

    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在: $backup_file"
        exit 1
    fi

    log_warning "⚠️  恢复操作将覆盖现有配置和数据！"
    read -p "确认继续? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        log "从 $backup_file 恢复数据..."

        # 解压备份
        tar xzf "$backup_file" -C "$PROJECT_ROOT"

        log_success "恢复完成"
    else
        log_info "操作已取消"
    fi
}

# 主函数
main() {
    case ${1:-help} in
        start)
            start_services "$2"
            ;;
        stop)
            stop_services "$2"
            ;;
        restart)
            restart_services "$2"
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        run-analysis)
            run_analysis
            ;;
        generate-report)
            generate_report
            ;;
        install-tools)
            install_tools
            ;;
        clean)
            clean_all
            ;;
        backup)
            backup_data
            ;;
        restore)
            restore_data "$2"
            ;;
        -h|--help|help)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"