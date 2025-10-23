#!/bin/bash
# Law OA Go 开发环境启动脚本
# 简化版本，兼容zsh和bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Docker Compose命令检测
DOCKER_COMPOSE=""
if command -v docker-compose >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker-compose"
elif docker compose version >/dev/null 2>&1; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"  # 默认值，会在check_dependencies中处理
fi

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
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

# 显示帮助
show_help() {
    cat << EOF
Law OA Go 开发环境启动脚本

用法: $0 [命令]

命令:
    start    启动所有服务 (默认)
    stop     停止所有服务
    restart  重启所有服务
    status   查看服务状态
    logs     查看服务日志
    clean    清理数据和缓存
    help     显示此帮助信息

示例:
    $0                # 启动服务
    $0 start          # 启动服务
    $0 stop           # 停止服务
    $0 logs           # 查看日志
EOF
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."

    # 检查Docker
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker未安装或未启动"
        exit 1
    fi

      # 验证Docker Compose命令可用性
    if ! $DOCKER_COMPOSE version >/dev/null 2>&1; then
        log_error "Docker Compose未安装或无法使用"
        log_info "请安装Docker Compose或更新Docker到最新版本"
        exit 1
    fi

    # 检查必要文件
    if [[ ! -f "docker-compose.yml" ]]; then
        log_error "缺少 docker-compose.yml 文件"
        exit 1
    fi

    log_success "系统依赖检查通过"
}

# 创建环境配置
setup_environment() {
    log_info "设置环境配置..."

    if [[ ! -f ".env.development" ]]; then
        if [[ -f ".env.example" ]]; then
            log_info "从模板创建开发环境配置..."
            cp .env.example .env.development
            log_success "环境配置文件创建完成"
        fi
    fi
}

# 启动服务
start_services() {
    log_info "启动开发环境服务..."

    # 设置环境变量
    export COMPOSE_FILE="docker-compose.yml"

    # 启动服务
    $DOCKER_COMPOSE up -d

    if [[ $? -eq 0 ]]; then
        log_success "服务启动成功"

        # 显示服务状态
        sleep 5
        $DOCKER_COMPOSE ps

        # 显示访问信息
        show_access_info
    else
        log_error "服务启动失败"
        exit 1
    fi
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."
    $DOCKER_COMPOSE down
    log_success "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    stop_services
    sleep 2
    start_services
}

# 显示服务状态
show_status() {
    log_info "服务状态:"
    echo ""
    if $DOCKER_COMPOSE ps >/dev/null 2>&1; then
        $DOCKER_COMPOSE ps
        echo ""
        echo "资源使用情况:"
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" 2>/dev/null || echo "无法获取资源使用情况"
    else
        echo "没有运行的服务"
    fi
}

# 显示日志
show_logs() {
    log_info "显示服务日志..."
    $DOCKER_COMPOSE logs -f
}

# 清理数据
clean_data() {
    log_info "清理数据和缓存..."
    read -p "这将删除所有容器和数据，确认继续吗？(y/N): " confirm
    if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
        $DOCKER_COMPOSE down -v
        docker system prune -f
        log_success "清理完成"
    else
        log_info "清理操作已取消"
    fi
}

# 显示访问信息
show_access_info() {
    echo ""
    echo -e "${GREEN}🚀 Law OA Go 服务已启动${NC}"
    echo ""
    echo -e "${BLUE}📱 前端应用:${NC}      http://localhost:3003"
    echo -e "${BLUE}🔧 后端API:${NC}        http://localhost:8080"
    echo -e "${BLUE}📊 健康检查:${NC}       http://localhost:8080/health"
    echo -e "${BLUE}📈 监控指标:${NC}       http://localhost:8080/metrics"
    echo ""
    echo -e "${BLUE}🗄️  MySQL:${NC}          localhost:33060"
    echo -e "${BLUE}🔴 Redis:${NC}          localhost:6379"
    echo -e "${BLUE}🔍 Elasticsearch:${NC}  http://localhost:9200"
    echo ""
    echo -e "${YELLOW}💡 提示: 使用 '$0 logs' 查看服务日志${NC}"
    echo -e "${YELLOW}💡 提示: 使用 '$0 stop' 停止所有服务${NC}"
    echo ""
}

# 主函数
main() {
    local command="start"

    if [[ $# -gt 0 ]]; then
        command="$1"
    fi

    case $command in
        start|"")
            check_dependencies
            setup_environment
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_data
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 显示标题
echo -e "${BLUE}"
cat << "EOF"
 ____  _   _ ____    _       ____    _    _     _
/ ___|| | | |  _ \  | |     |  _ \  / \  | |   | |
\___ \| |_| | | | | | |     | | | |/ _ \ | |   | |
 ___) |  _  | |_| | | |___  | |_| / ___ \| |___| |___
|____/|_| |_|____/  |_____| |____/_/   \_\_____|_____|

    开发环境快速启动脚本 v2.1.0
EOF
echo -e "${NC}"
echo ""

# 执行主函数
main "$@"