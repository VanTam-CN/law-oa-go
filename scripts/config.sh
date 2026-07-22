#!/bin/bash

# 配置管理脚本
# 用法: ./scripts/config.sh [command] [options]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 显示帮助信息
show_help() {
    cat << EOF
配置管理脚本

用法:
    $0 [command] [options]

命令:
    help                    显示此帮助信息
    env                     显示当前环境
    switch <env>           切换到指定环境 (development, production, test)
    validate                验证当前配置
    generate <env>          生成环境配置文件模板
    list                    列出所有可用环境
    show                    显示当前配置摘要
    test                    测试配置加载
    backup                  备份当前配置
    restore <backup>        恢复配置备份

示例:
    $0 env                  # 显示当前环境
    $0 switch production   # 切换到生产环境
    $0 validate             # 验证配置
    $0 generate staging     # 生成staging环境配置
    $0 show                 # 显示配置摘要

EOF
}

# 获取项目根目录
get_project_root() {
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    project_root="$(dirname "$script_dir")"
    echo "$project_root"
}

# 获取当前环境
get_current_env() {
    cd "$(get_project_root)"

    if [[ -f ".env" ]]; then
        env=$(grep "^APP_ENV=" .env | cut -d'=' -f2)
        if [[ -n "$env" ]]; then
            echo "$env"
            return 0
        fi
    fi

    # 检查环境变量
    if [[ -n "$APP_ENV" ]]; then
        echo "$APP_ENV"
        return 0
    fi

    echo "development"
}

# 切换环境
switch_env() {
    local target_env="$1"

    if [[ -z "$target_env" ]]; then
        log_error "请指定目标环境"
        show_help
        exit 1
    fi

    # 验证环境名称
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

    # 备份当前配置
    backup_config

    # 切换环境配置
    local env_file="configs/environments/.env.$target_env"
    if [[ -f "$env_file" ]]; then
        cp "$env_file" ".env"
        log_success "已切换到 $target_env 环境"
        log_info "配置文件: $env_file"
    else
        log_warning "环境文件不存在: $env_file"
        log_info "正在创建默认配置..."
        generate_env_config "$target_env"
    fi

    # 验证新配置
    validate_config
}

# 生成环境配置文件
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

    # 创建目录
    mkdir -p configs/environments

    # 根据环境生成配置
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
DB_NAME=law_oa_dev

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
        production)
            cat > "$env_file" << 'EOF'
# 生产环境配置
APP_ENV=production
DEBUG=false
LOG_LEVEL=warn

# 服务器配置
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# 数据库配置
DB_DRIVER=postgres
DB_HOST=${DB_HOST}
DB_PORT=5432
DB_USERNAME=${DB_USERNAME}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}

# Redis配置
REDIS_HOST=${REDIS_HOST}
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD}

# Elasticsearch配置
ES_ENABLED=true
ES_HOSTS=${ES_HOSTS}

# JWT配置
JWT_SECRET=${JWT_SECRET}
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
        staging)
            cat > "$env_file" << 'EOF'
# 预发布环境配置
APP_ENV=staging
DEBUG=false
LOG_LEVEL=info

# 服务器配置
SERVER_HOST=localhost
SERVER_PORT=8080

# 数据库配置
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=law_oa_staging
DB_PASSWORD=<set-staging-password>
DB_NAME=law_oa_staging

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=2

# Elasticsearch配置
ES_ENABLED=true
ES_HOSTS=localhost:9200

# JWT配置
JWT_SECRET=<set-staging-jwt-secret>
EOF
            ;;
    esac

    log_success "已生成 $env 环境配置文件: $env_file"
    log_warning "请根据实际环境修改配置文件中的敏感信息"
}

# 验证配置
validate_config() {
    cd "$(get_project_root)"

    log_info "验证配置..."

    # 检查必需的环境变量
    local required_vars=("DB_USERNAME" "DB_PASSWORD" "DB_NAME" "JWT_SECRET")
    local missing_vars=()

    for var in "${required_vars[@]}"; do
        if [[ -f ".env" ]]; then
            if ! grep -q "^$var=" .env; then
                missing_vars+=("$var")
            fi
        else
            if [[ -z "${!var}" ]]; then
                missing_vars+=("$var")
            fi
        fi
    done

    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "缺少必需的环境变量: ${missing_vars[*]}"
        return 1
    fi

    log_success "配置验证通过"
}

# 显示配置摘要
show_config() {
    cd "$(get_project_root)"

    local current_env=$(get_current_env)
    local config_file=".env"

    echo "================================"
    echo "配置摘要"
    echo "================================"
    echo "环境: $current_env"
    echo "配置文件: $config_file"
    echo ""

    if [[ -f "$config_file" ]]; then
        echo "主要配置项:"
        grep -E "^(APP_ENV|DEBUG|LOG_LEVEL|SERVER_PORT|DB_HOST|ES_ENABLED)" "$config_file" | while IFS= read -r line; do
            echo "  $line"
        done
    fi

    echo "================================"
}

# 列出可用环境
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

# 测试配置加载
test_config() {
    cd "$(get_project_root)"

    log_info "测试配置加载..."

    if ! command -v go &> /dev/null; then
        log_error "Go 命令未找到，无法测试配置"
        return 1
    fi

    # 尝试构建和运行配置测试
    if timeout 30s go test ./internal/config -v > /dev/null 2>&1; then
        log_success "配置测试通过"
    else
        log_error "配置测试失败"
        return 1
    fi
}

# 备份配置
backup_config() {
    cd "$(get_project_root)"

    local backup_dir="config-backups"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$backup_dir/config_$timestamp.tar.gz"

    mkdir -p "$backup_dir"

    # 备份配置文件
    tar -czf "$backup_file" .env configs/ 2>/dev/null || {
        log_warning "没有找到配置文件可以备份"
        return 0
    }

    log_success "配置已备份到: $backup_file"
}

# 恢复配置
restore_config() {
    local backup_file="$1"

    if [[ -z "$backup_file" ]]; then
        log_error "请指定备份文件"
        return 1
    fi

    cd "$(get_project_root)"

    if [[ ! -f "$backup_file" ]]; then
        log_error "备份文件不存在: $backup_file"
        return 1
    fi

    # 备份当前配置
    backup_config

    # 恢复配置
    tar -xzf "$backup_file"

    log_success "配置已从 $backup_file 恢复"
}

# 主函数
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
        validate)
            validate_config
            ;;
        generate)
            generate_env_config "$2"
            ;;
        list)
            list_environments
            ;;
        show)
            show_config
            ;;
        test)
            test_config
            ;;
        backup)
            backup_config
            ;;
        restore)
            restore_config "$2"
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
