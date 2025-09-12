#!/bin/bash

# Law OA Go 项目开发脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目信息
PROJECT_NAME="Law OA Go"
VERSION="1.0.0"

# 打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    print_info "检查依赖..."
    
    # 检查 Go
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装，请先安装 Go 1.21 或更高版本"
        exit 1
    fi
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        print_warning "Docker 未安装，Docker 相关功能将不可用"
    fi
    
    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        print_warning "Docker Compose 未安装，Docker 相关功能将不可用"
    fi
    
    print_success "依赖检查完成"
}

# 初始化项目
init_project() {
    print_info "初始化项目..."
    
    # 创建必要的目录
    mkdir -p logs
    mkdir -p scripts
    mkdir -p docs
    
    # 复制环境变量文件
    if [ ! -f .env ]; then
        cp .env.example .env
        print_success "已创建 .env 文件，请根据需要修改配置"
    fi
    
    # 下载依赖
    print_info "下载 Go 依赖..."
    go mod download
    go mod tidy
    
    print_success "项目初始化完成"
}

# 启动开发环境
start_dev() {
    print_info "启动开发环境..."
    
    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，无法启动开发环境"
        exit 1
    fi
    
    # 启动 Docker Compose
    docker-compose up -d
    
    print_success "开发环境启动完成"
    print_info "API 地址: http://localhost:8080"
    print_info "Kibana 地址: http://localhost:5601"
}

# 停止开发环境
stop_dev() {
    print_info "停止开发环境..."
    
    docker-compose down
    
    print_success "开发环境已停止"
}

# 构建项目
build() {
    print_info "构建项目..."
    
    # 构建 Go 应用
    go build -o main .
    
    print_success "项目构建完成"
}

# 运行测试
test() {
    print_info "运行测试..."
    
    go test -v ./...
    
    print_success "测试完成"
}

# 运行代码检查
lint() {
    print_info "运行代码检查..."
    
    # 检查是否安装了 golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        print_warning "golangci-lint 未安装，跳过代码检查"
        return
    fi
    
    golangci-lint run
    
    print_success "代码检查完成"
}

# 运行应用
run() {
    print_info "运行应用..."
    
    go run main.go
}

# 生成文档
docs() {
    print_info "生成文档..."
    
    # 生成 Swagger 文档
    if command -v swag &> /dev/null; then
        swag init
        print_success "Swagger 文档生成完成"
    else
        print_warning "swag 未安装，跳过 Swagger 文档生成"
    fi
    
    # 生成其他文档
    print_info "生成项目文档..."
    
    print_success "文档生成完成"
}

# 显示帮助信息
show_help() {
    echo "Law OA Go 项目开发脚本"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  init          初始化项目"
    echo "  start         启动开发环境"
    echo "  stop          停止开发环境"
    echo "  build         构建项目"
    echo "  test          运行测试"
    echo "  lint          运行代码检查"
    echo "  run           运行应用"
    echo "  docs          生成文档"
    echo "  clean         清理项目"
    echo "  help          显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 init"
    echo "  $0 start"
    echo "  $0 build"
}

# 清理项目
clean() {
    print_info "清理项目..."
    
    # 清理构建文件
    rm -f main
    
    # 清理日志文件
    rm -rf logs/*
    
    # 清理 Docker 资源
    if command -v docker &> /dev/null; then
        docker-compose down -v
        docker system prune -f
    fi
    
    print_success "项目清理完成"
}

# 主函数
main() {
    case "${1:-}" in
        "init")
            check_dependencies
            init_project
            ;;
        "start")
            start_dev
            ;;
        "stop")
            stop_dev
            ;;
        "build")
            build
            ;;
        "test")
            test
            ;;
        "lint")
            lint
            ;;
        "run")
            run
            ;;
        "docs")
            docs
            ;;
        "clean")
            clean
            ;;
        "help"|"--help"|"-h"|"")
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"