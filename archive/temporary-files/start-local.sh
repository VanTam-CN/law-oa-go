#!/bin/bash
# 本地开发环境启动脚本 - Law OA Go
# 不使用Docker，直接启动前后端服务

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 检查必要的工具
check_requirements() {
    log_step "检查开发环境依赖..."

    # 检查Go
    if ! command -v go >/dev/null 2>&1; then
        log_error "Go未安装，请先安装Go 1.21+"
        exit 1
    fi

    # 检查Node.js
    if ! command -v node >/dev/null 2>&1; then
        log_error "Node.js未安装，请先安装Node.js 18+"
        exit 1
    fi

    # 检查npm
    if ! command -v npm >/dev/null 2>&1; then
        log_error "npm未安装，请先安装npm"
        exit 1
    fi

    # 检查MySQL (可选)
    if command -v mysql >/dev/null 2>&1; then
        log_info "MySQL已安装"
    else
        log_warning "MySQL未安装，请确保MySQL服务正在运行"
    fi

    # 检查Redis (可选)
    if command -v redis-cli >/dev/null 2>&1; then
        log_info "Redis已安装"
    else
        log_warning "Redis未安装，请确保Redis服务正在运行"
    fi

    log_success "开发环境依赖检查完成"
}

# 启动MySQL (如果需要)
start_mysql() {
    if command -v brew >/dev/null 2>&1; then
        if brew services list | grep mysql | grep -q "stopped"; then
            log_info "启动MySQL服务..."
            brew services start mysql
            sleep 5
        fi
    elif command -v systemctl >/dev/null 2>&1; then
        if ! systemctl is-active --quiet mysql; then
            log_info "启动MySQL服务..."
            sudo systemctl start mysql
            sleep 5
        fi
    fi
}

# 启动Redis (如果需要)
start_redis() {
    if command -v brew >/dev/null 2>&1; then
        if brew services list | grep redis | grep -q "stopped"; then
            log_info "启动Redis服务..."
            brew services start redis
            sleep 2
        fi
    elif command -v systemctl >/dev/null 2>&1; then
        if ! systemctl is-active --quiet redis; then
            log_info "启动Redis服务..."
            sudo systemctl start redis
            sleep 2
        fi
    fi
}

# 初始化Go模块
setup_go() {
    log_step "初始化Go后端..."

    # 检查go.mod
    if [[ ! -f "go.mod" ]]; then
        log_error "go.mod文件不存在"
        exit 1
    fi

    # 下载依赖
    log_info "下载Go依赖..."
    go mod download
    go mod tidy

    # 检查环境配置
    if [[ ! -f ".env.local" ]]; then
        log_warning ".env.local文件不存在，使用默认配置"
        # 创建基础配置
        cat > .env.local << 'EOF'
# 本地开发环境配置
ENVIRONMENT=development
DEBUG=true
PORT=8080
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=law_oa
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=your-very-secure-jwt-secret-key-for-development-only
EOF
    fi

    log_success "Go后端初始化完成"
}

# 初始化前端
setup_frontend() {
    log_step "初始化React前端..."

    cd frontend

    # 检查package.json
    if [[ ! -f "package.json" ]]; then
        log_error "frontend/package.json文件不存在"
        cd ..
        exit 1
    fi

    # 安装依赖
    log_info "安装前端依赖..."
    npm install

    # 检查环境配置
    if [[ ! -f ".env.local" ]]; then
        log_info "创建前端环境配置..."
        cat > .env.local << 'EOF'
REACT_APP_API_URL=http://localhost:8080
REACT_APP_ENVIRONMENT=development
GENERATE_SOURCEMAP=true
EOF
    fi

    cd ..
    log_success "前端初始化完成"
}

# 启动后端服务
start_backend() {
    log_step "启动Go后端服务..."

    # 设置环境变量
    export ENVIRONMENT=development
    export DEBUG=true
    export PORT=8080
    export GIN_MODE=debug

    # 检查端口是否被占用
    if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "端口8080已被占用，尝试终止现有进程..."
        lsof -ti:8080 | xargs kill -9 2>/dev/null || true
        sleep 2
    fi

    # 检查是否有编译好的可执行文件
    if [[ -f "./main" ]]; then
        log_info "使用编译好的可执行文件启动后端..."
        nohup ./main > backend.log 2>&1 &
        BACKEND_PID=$!
        echo $BACKEND_PID > .backend.pid
    else
        log_info "编译并启动Go服务..."
        # 编译并运行
        go build -o main main.go
        nohup ./main > backend.log 2>&1 &
        BACKEND_PID=$!
        echo $BACKEND_PID > .backend.pid
    fi

    log_success "后端服务已启动 (PID: $BACKEND_PID)"
    log_info "后端日志: backend.log"
}

# 启动前端服务
start_frontend() {
    log_step "启动React前端服务..."

    cd frontend

    # 检查端口是否被占用
    if lsof -Pi :3003 -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "端口3003已被占用，尝试终止现有进程..."
        lsof -ti:3003 | xargs kill -9 2>/dev/null || true
        sleep 2
    fi

    # 检查是否有编译好的前端文件
    if [[ -f "./frontend" ]] && [[ -d "./dist" ]]; then
        log_info "使用编译好的前端文件启动静态服务器..."
        # 使用Python启动静态文件服务器（跨平台）
        if command -v python3 >/dev/null 2>&1; then
            nohup python3 -m http.server 3003 --directory dist > ../frontend.log 2>&1 &
            FRONTEND_PID=$!
            echo $FRONTEND_PID > ../.frontend.pid
        elif command -v python >/dev/null 2>&1; then
            nohup python -m SimpleHTTPServer 3003 > ../frontend.log 2>&1 &
            FRONTEND_PID=$!
            echo $FRONTEND_PID > ../.frontend.pid
        else
            log_warning "未找到Python，使用npm启动开发服务器..."
            nohup npm start > ../frontend.log 2>&1 &
            FRONTEND_PID=$!
            echo $FRONTEND_PID > ../.frontend.pid
        fi
    else
        log_info "启动前端开发服务器..."
        # 使用npm启动开发服务器
        nohup npm start > ../frontend.log 2>&1 &
        FRONTEND_PID=$!
        echo $FRONTEND_PID > ../.frontend.pid
    fi

    cd ..
    log_success "前端服务已启动 (PID: $FRONTEND_PID)"
    log_info "前端日志: frontend.log"
}

# 健康检查
health_check() {
    log_step "执行健康检查..."

    local max_attempts=30
    local attempt=1

    while [[ $attempt -le $max_attempts ]]; do
        log_info "健康检查尝试 $attempt/$max_attempts..."

        # 检查后端
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            log_success "后端服务健康检查通过"
            break
        fi

        if [[ $attempt -eq $max_attempts ]]; then
            log_error "后端健康检查失败"
            show_logs
            return 1
        fi

        sleep 3
        ((attempt++))
    done

    # 等待前端启动（React启动较慢）
    local frontend_attempts=40
    local frontend_attempt=1

    while [[ $frontend_attempt -le $frontend_attempts ]]; do
        if curl -f http://localhost:3003 >/dev/null 2>&1; then
            log_success "前端服务健康检查通过"
            break
        fi

        if [[ $frontend_attempt -eq $frontend_attempts ]]; then
            log_warning "前端服务可能还在启动中，这是正常的"
            break
        fi

        sleep 5
        ((frontend_attempt++))
    done
}

# 显示服务状态
show_status() {
    log_step "服务状态:"
    echo ""

    # 显示进程状态
    if [[ -f ".backend.pid" ]]; then
        local backend_pid=$(cat .backend.pid)
        if ps -p $backend_pid > /dev/null 2>&1; then
            echo -e "${GREEN}✅ 后端服务运行中 (PID: $backend_pid)${NC}"
        else
            echo -e "${RED}❌ 后端服务未运行${NC}"
        fi
    else
        echo -e "${RED}❌ 后端服务未启动${NC}"
    fi

    if [[ -f ".frontend.pid" ]]; then
        local frontend_pid=$(cat .frontend.pid)
        if ps -p $frontend_pid > /dev/null 2>&1; then
            echo -e "${GREEN}✅ 前端服务运行中 (PID: $frontend_pid)${NC}"
        else
            echo -e "${RED}❌ 前端服务未运行${NC}"
        fi
    else
        echo -e "${RED}❌ 前端服务未启动${NC}"
    fi

    echo ""
    echo -e "${BLUE}📊 端口占用情况:${NC}"
    lsof -i :8080 2>/dev/null || echo "  端口8080: 未占用"
    lsof -i :3003 2>/dev/null || echo "  端口3003: 未占用"
    echo ""
}

# 显示日志
show_logs() {
    log_step "显示服务日志:"
    echo ""

    if [[ -f "backend.log" ]]; then
        echo -e "${BLUE}📋 后端日志 (最后20行):${NC}"
        echo -e "${YELLOW}----------------------------------------${NC}"
        tail -20 backend.log
        echo ""
    fi

    if [[ -f "frontend.log" ]]; then
        echo -e "${BLUE}📋 前端日志 (最后20行):${NC}"
        echo -e "${YELLOW}----------------------------------------${NC}"
        tail -20 frontend.log
        echo ""
    fi
}

# 停止所有服务
stop_services() {
    log_step "停止所有服务..."

    # 停止后端
    if [[ -f ".backend.pid" ]]; then
        local backend_pid=$(cat .backend.pid)
        if ps -p $backend_pid > /dev/null 2>&1; then
            log_info "停止后端服务 (PID: $backend_pid)..."
            kill $backend_pid
            sleep 2
            # 强制终止
            ps -p $backend_pid > /dev/null 2>&1 && kill -9 $backend_pid
        fi
        rm -f .backend.pid
    fi

    # 停止前端
    if [[ -f ".frontend.pid" ]]; then
        local frontend_pid=$(cat .frontend.pid)
        if ps -p $frontend_pid > /dev/null 2>&1; then
            log_info "停止前端服务 (PID: $frontend_pid)..."
            kill $frontend_pid
            sleep 2
            # 强制终止
            ps -p $frontend_pid > /dev/null 2>&1 && kill -9 $frontend_pid
        fi
        rm -f .frontend.pid
    fi

    # 清理端口占用
    lsof -ti:8080 | xargs kill -9 2>/dev/null || true
    lsof -ti:3003 | xargs kill -9 2>/dev/null || true

    log_success "所有服务已停止"
}

# 重启服务
restart_services() {
    log_step "重启所有服务..."
    stop_services
    sleep 3
    start_all_services
}

# 启动所有服务
start_all_services() {
    start_mysql
    start_redis
    setup_go
    setup_frontend
    start_backend
    start_frontend
    health_check
    show_access_info
}

# 显示访问信息
show_access_info() {
    log_step "服务访问信息:"
    echo ""
    echo -e "${GREEN}🚀 Law OA Go 本地开发环境已启动${NC}"
    echo ""
    echo -e "${BLUE}📱 前端应用:${NC}      http://localhost:3003"
    echo -e "${BLUE}🔧 后端API:${NC}        http://localhost:8080"
    echo -e "${BLUE}📊 健康检查:${NC}       http://localhost:8080/health"
    echo -e "${BLUE}📈 监控指标:${NC}       http://localhost:8080/metrics"
    echo ""
    echo -e "${YELLOW}💡 提示: 使用 './start-local.sh logs' 查看服务日志${NC}"
    echo -e "${YELLOW}💡 提示: 使用 './start-local.sh status' 查看服务状态${NC}"
    echo -e "${YELLOW}💡 提示: 使用 './start-local.sh stop' 停止所有服务${NC}"
    echo ""
    echo -e "${BLUE}📋 日志文件:${NC}"
    echo -e "  后端日志: backend.log"
    echo -e "  前端日志: frontend.log"
    echo ""
}

# 主函数
main() {
    echo -e "${BLUE}"
    cat << "EOF"
 ____  _   _ ____    _       ____    _    _     _
/ ___|| | | |  _ \  | |     |  _ \  / \  | |   | |
\___ \| |_| | | | | | |     | | | |/ _ \ | |   | |
 ___) |  _  | |_| | | |___  | |_| / ___ \| |___| |___
|____/|_| |_|____/  |_____| |____/_/   \_\_____|_____|

    LOCAL DEVELOPMENT STARTUP SCRIPT v1.0
EOF
    echo -e "${NC}"
    echo ""

    case "${1:-start}" in
        start)
            check_requirements
            start_all_services
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
        backend)
            setup_go
            start_backend
            sleep 5
            health_check
            ;;
        frontend)
            setup_frontend
            start_frontend
            ;;
        help|"")
            echo "用法: $0 {start|stop|restart|status|logs|backend|frontend|help}"
            echo ""
            echo "命令说明:"
            echo "  start    - 启动前后端服务（默认）"
            echo "  stop     - 停止所有服务"
            echo "  restart  - 重启所有服务"
            echo "  status   - 查看服务状态"
            echo "  logs     - 查看服务日志"
            echo "  backend  - 仅启动后端服务"
            echo "  frontend - 仅启动前端服务"
            echo "  help     - 显示此帮助信息"
            echo ""
            echo "首次运行请确保:"
            echo "1. MySQL服务正在运行"
            echo "2. Redis服务正在运行"
            echo "3. 创建了数据库 law_oa"
            echo ""
            exit 0
            ;;
        *)
            log_error "未知命令: $1"
            main help
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap 'log_warning "操作被中断"; stop_services; exit 1' INT TERM

# 执行主函数
main "$@"