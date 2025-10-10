#!/bin/bash

# 律师事务所管理系统开发环境启动脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
COMPOSE_FILE="docker-compose.dev.yml"
PROJECT_NAME="law-oa"
MYSQL_PORT=33060
FRONTEND_PORT=3003
BACKEND_PORT=8080

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

# 检查Docker是否运行
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        log_error "Docker未运行，请先启动Docker"
        exit 1
    fi
}

# 检查端口是否被占用
check_port() {
    local port=$1
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_error "端口 $port 已被占用"
        exit 1
    fi
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME down
    log_success "所有服务已停止"
}

# 启动服务
start_services() {
    log_info "启动律师事务所管理系统开发环境..."

    # 检查端口
    log_info "检查端口占用情况..."
    check_port $MYSQL_PORT
    check_port $FRONTEND_PORT
    check_port $BACKEND_PORT

    # 构建镜像
    log_info "构建Docker镜像..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME build

    # 启动服务
    log_info "启动所有服务..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME up -d

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 30

    # 检查服务状态
    log_info "检查服务状态..."
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps

    # 显示访问信息
    echo ""
    log_success "开发环境启动完成！"
    echo ""
    echo "访问地址："
    echo "  前端应用: http://localhost:$FRONTEND_PORT"
    echo "  后端API: http://localhost:$BACKEND_PORT"
    echo "  健康检查: http://localhost:$BACKEND_PORT/health"
    echo "  API文档: http://localhost:$BACKEND_PORT/swagger/index.html"
    echo "  数据库管理: http://localhost:8081 (phpMyAdmin)"
    echo "  Redis管理: http://localhost:8082 (Redis Commander)"
    echo "  Elasticsearch: http://localhost:9200"
    echo "  Kibana: http://localhost:5601"
    echo ""
    echo "数据库配置："
    echo "  主机: localhost"
    echo "  端口: $MYSQL_PORT"
    echo "  用户: root"
    echo "  密码: password"
    echo "  数据库: law_oa"
    echo ""
    log_warning "首次启动可能需要几分钟时间，请耐心等待"
}

# 重启服务
restart_services() {
    log_info "重启所有服务..."
    stop_services
    sleep 5
    start_services
}

# 查看日志
show_logs() {
    local service=$1
    if [ -z "$service" ]; then
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f --tail=100
    else
        docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME logs -f --tail=100 $service
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统开发环境脚本

用法: $0 [命令]

命令:
    start       启动开发环境
    stop        停止开发环境
    restart     重启开发环境
    logs        查看所有服务日志
    logs [服务] 查看特定服务日志
    status      查看服务状态
    help        显示此帮助信息

示例:
    $0 start     # 启动开发环境
    $0 stop      # 停止开发环境
    $0 restart   # 重启开发环境
    $0 logs      # 查看所有日志
    $0 logs backend # 查看后端服务日志
    $0 status    # 查看服务状态

EOF
}

# 显示服务状态
show_status() {
    log_info "服务状态："
    docker-compose -f $COMPOSE_FILE -p $PROJECT_NAME ps
}

# 主函数
main() {
    check_docker

    case "${1:-}" in
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            show_logs $2
            ;;
        "status")
            show_status
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"