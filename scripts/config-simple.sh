#!/bin/bash

# 简单配置管理脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
配置管理脚本

用法:
    $0 [command] [options]

命令:
    help                    显示此帮助信息
    env                     显示当前环境
    switch <env>           切换到指定环境
    generate <env>          生成环境配置文件模板
    list                    列出所有可用环境
    test                    测试配置加载

示例:
    $0 env                  # 显示当前环境
    $0 switch production   # 切换到生产环境
    $0 generate staging     # 生成staging环境配置
    $0 test                 # 测试配置加载

EOF
}

get_project_root() {
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    project_root="$(dirname "$script_dir")"
    echo "$project_root"
}

get_current_env() {
    cd "$(get_project_root)"

    if [[ -f ".env" ]]; then
        env=$(grep "^APP_ENV=" .env 2>/dev/null | cut -d'=' -f2)
        if [[ -n "$env" ]]; then
            echo "$env"
            return 0
        fi
    fi

    echo "development"
}

switch_env() {
    local target_env="$1"

    if [[ -z "$target_env" ]]; then
        log_error "请指定目标环境"
        show_help
        exit 1
    fi

    case "$target_env" in
        development|production|test|staging)
            ;;
        *)
            log_error "无效的环境: $target_env"
            log_info "有效环境: development, production, test, staging"
            exit 1
            ;;
    esac

    cd "$(get_project_root)"

    local env_file="configs/environments/.env.$target_env"
    if [[ -f "$env_file" ]]; then
        cp "$env_file" ".env"
        log_success "已切换到 $target_env 环境"
        log_info "配置文件: $env_file"
    else
        log_error "环境文件不存在: $env_file"
        log_info "请先创建配置文件"
    fi
}

generate_env_config() {
    local env="$1"

    if [[ -z "$env" ]]; then
        log_error "请指定环境名称"
        exit 1
    fi

    cd "$(get_project_root)"

    local env_file="configs/environments/.env.$env"

    if [[ -f "$env_file" ]]; then
        log_warning "配置文件已存在: $env_file"
        return 0
    fi

    mkdir -p configs/environments

    case "$env" in
        development)
            cat > "$env_file" << 'EOF'
# 开发环境配置
APP_ENV=development
DEBUG=true
LOG_LEVEL=debug

# 服务器配置
SERVER_HOST=localhost
SERVER_PORT=8080

# 数据库配置
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=law_oa_user
DB_PASSWORD=<set-local-dev-password>
DB_NAME=law_oa_db

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Elasticsearch配置
ES_ENABLED=true
ES_HOSTS=localhost:9200

# JWT配置
JWT_SECRET=<set-local-dev-jwt-secret>
EOF
            ;;
        test)
            cat > "$env_file" << 'EOF'
# 测试环境配置
APP_ENV=test
DEBUG=false
LOG_LEVEL=error

# 服务器配置
SERVER_HOST=localhost
SERVER_PORT=0

# 数据库配置
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=test
DB_PASSWORD=<set-test-password>
DB_NAME=law_oa_test

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=1

# Elasticsearch配置
ES_ENABLED=false

# JWT配置
JWT_SECRET=<set-test-jwt-secret>
EOF
            ;;
    esac

    log_success "已生成 $env 环境配置文件: $env_file"
}

list_environments() {
    cd "$(get_project_root)"

    echo "可用环境:"
    local config_dir="configs/environments"
    if [[ -d "$config_dir" ]]; then
        for env_file in "$config_dir"/.env.*; do
            if [[ -f "$env_file" ]]; then
                local env_name=$(basename "$env_file" | sed 's/^\.env\.//')
                local current_env=$(get_current_env)
                if [[ "$env_name" == "$current_env" ]]; then
                    echo "  ✓ $env_name (当前)"
                else
                    echo "  - $env_name"
                fi
            fi
        done
    else
        echo "  development"
        echo "  production"
        echo "  test"
        echo "  staging"
    fi
}

test_config() {
    cd "$(get_project_root)"

    log_info "测试配置加载..."

    if ! command -v go &> /dev/null; then
        log_error "Go 命令未找到，无法测试配置"
        return 1
    fi

    if timeout 30s go test ./internal/config -v > /dev/null 2>&1; then
        log_success "配置测试通过"
    else
        log_error "配置测试失败"
        return 1
    fi
}

main() {
    local command="${1:-help}"

    case "$command" in
        help|--help|-h)
            show_help
            ;;
        env)
            local current_env=$(get_current_env)
            log_info "当前环境: $current_env"
            ;;
        switch)
            switch_env "$2"
            ;;
        generate)
            generate_env_config "$2"
            ;;
        list)
            list_environments
            ;;
        test)
            test_config
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
