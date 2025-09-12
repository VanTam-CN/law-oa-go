#!/bin/bash

# 律师事务所管理系统启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目路径
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$PROJECT_ROOT"

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

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装，请先安装 Go 1.19 或更高版本"
        exit 1
    fi
    
    # 检查版本
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    REQUIRED_VERSION="1.19"
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
        log_error "Go 版本过低，需要 $REQUIRED_VERSION 或更高版本，当前版本：$GO_VERSION"
        exit 1
    fi
    
    log_success "Go 版本检查通过: $GO_VERSION"
    
    # 检查 MySQL
    if ! command -v mysql &> /dev/null; then
        log_warning "MySQL 客户端未安装，请确保数据库服务可用"
    fi
    
    # 检查 Redis
    if ! command -v redis-cli &> /dev/null; then
        log_warning "Redis 客户端未安装，请确保 Redis 服务可用"
    fi
}

# 检查配置文件
check_config() {
    log_info "检查配置文件..."
    
    if [ ! -f ".env" ]; then
        log_warning ".env 文件不存在，复制 .env.example"
        cp .env.example .env
        log_warning "请编辑 .env 文件配置您的环境变量"
    fi
    
    if [ ! -f "config/config.yaml" ]; then
        log_error "配置文件 config/config.yaml 不存在"
        exit 1
    fi
    
    log_success "配置文件检查通过"
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    
    mkdir -p logs
    mkdir -p uploads
    mkdir -p uploads/contract
    mkdir -p uploads/evidence
    mkdir -p uploads/letter
    mkdir -p uploads/other
    
    log_success "目录创建完成"
}

# 下载依赖
download_dependencies() {
    log_info "下载 Go 依赖..."
    
    go mod tidy
    go mod download
    
    log_success "依赖下载完成"
}

# 检查数据库连接
check_database() {
    log_info "检查数据库连接..."
    
    if [ -f ".env" ]; then
        source .env
        
        if [ -n "$DB_HOST" ] && [ -n "$DB_USER" ] && [ -n "$DB_PASSWORD" ] && [ -n "$DB_NAME" ]; then
            # 等待数据库启动
            log_info "等待数据库启动..."
            for i in {1..30}; do
                if mysql -h"$DB_HOST" -P"${DB_PORT:-3306}" -u"$DB_USER" -p"$DB_PASSWORD" -e "USE $DB_NAME" &>/dev/null; then
                    log_success "数据库连接成功"
                    break
                fi
                if [ $i -eq 30 ]; then
                    log_error "数据库连接失败，请检查配置"
                    exit 1
                fi
                sleep 1
            done
        else
            log_warning "数据库配置不完整，跳过连接检查"
        fi
    else
        log_warning "未找到 .env 文件，跳过数据库检查"
    fi
}

# 检查 Redis 连接
check_redis() {
    log_info "检查 Redis 连接..."
    
    if [ -f ".env" ]; then
        source .env
        
        if [ -n "$REDIS_HOST" ]; then
            if redis-cli -h "$REDIS_HOST" -p "${REDIS_PORT:-6379}" ping &>/dev/null; then
                log_success "Redis 连接成功"
            else
                log_error "Redis 连接失败，请检查配置"
                exit 1
            fi
        else
            log_warning "Redis 配置不完整，跳过连接检查"
        fi
    else
        log_warning "未找到 .env 文件，跳过 Redis 检查"
    fi
}

# 运行数据库迁移
run_migrations() {
    log_info "运行数据库迁移..."
    
    if [ -f "migrate/main.go" ]; then
        go run migrate/main.go up
        log_success "数据库迁移完成"
    else
        log_warning "迁移工具不存在，跳过迁移"
    fi
}

# 启动应用
start_application() {
    log_info "启动应用程序..."
    
    # 设置环境变量
    if [ -f ".env" ]; then
        set -a
        source .env
        set +a
    fi
    
    # 启动应用
    go run main.go
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统启动脚本

用法: $0 [选项]

选项:
    -h, --help          显示帮助信息
    -d, --dev           开发模式启动
    -p, --prod          生产模式启动
    -c, --check         仅检查依赖和配置
    -m, --migrate       仅运行数据库迁移
    -b, --build         构建应用
    -r, --run           运行构建后的应用

示例:
    $0                  启动开发服务器
    $0 --check          检查依赖和配置
    $0 --migrate        运行数据库迁移
    $0 --build          构建应用
    $0 --run            运行构建后的应用

EOF
}

# 主函数
main() {
    case "${1:-}" in
        -h|--help)
            show_help
            exit 0
            ;;
        -c|--check)
            check_dependencies
            check_config
            log_success "检查完成"
            exit 0
            ;;
        -m|--migrate)
            check_dependencies
            check_config
            check_database
            run_migrations
            exit 0
            ;;
        -b|--build)
            check_dependencies
            check_config
            log_info "构建应用程序..."
            go build -o bin/law-oa main.go
            log_success "构建完成，输出文件: bin/law-oa"
            exit 0
            ;;
        -r|--run)
            if [ ! -f "bin/law-oa" ]; then
                log_error "未找到构建文件，请先运行 $0 --build"
                exit 1
            fi
            log_info "运行应用程序..."
            ./bin/law-oa
            exit 0
            ;;
        -d|--dev)
            log_info "以开发模式启动..."
            ;;
        -p|--prod)
            log_info "以生产模式启动..."
            export ENVIRONMENT=production
            ;;
        "")
            log_info "以默认模式启动..."
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
    
    # 正常启动流程
    check_dependencies
    check_config
    create_directories
    download_dependencies
    check_database
    check_redis
    run_migrations
    start_application
}

# 运行主函数
main "$@"