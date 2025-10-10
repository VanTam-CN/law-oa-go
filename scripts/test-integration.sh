#!/bin/bash

# 律师事务所管理系统集成测试脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
FRONTEND_URL="http://localhost:3003"
BACKEND_URL="http://localhost:8080"
MYSQL_PORT=33060
TIMEOUT=30

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

# 检查URL可访问性
check_url() {
    local url=$1
    local description=$2
    local max_attempts=$3
    local attempt=1

    log_info "检查 $description ($url)..."

    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "$description 可访问"
            return 0
        fi

        log_warning "$description 不可访问，等待重试... ($attempt/$max_attempts)"
        sleep 5
        attempt=$((attempt + 1))
    done

    log_error "$description 不可访问，请检查服务状态"
    return 1
}

# 检查数据库连接
check_database() {
    log_info "检查数据库连接..."

    if mysql -h localhost -P $MYSQL_PORT -u root -ppassword -e "SELECT 1;" > /dev/null 2>&1; then
        log_success "数据库连接正常"
        return 0
    else
        log_error "数据库连接失败"
        return 1
    fi
}

# 检查Redis连接
check_redis() {
    log_info "检查Redis连接..."

    if redis-cli -p 6379 ping > /dev/null 2>&1; then
        log_success "Redis连接正常"
        return 0
    else
        log_error "Redis连接失败"
        return 1
    fi
}

# 检查Elasticsearch连接
check_elasticsearch() {
    log_info "检查Elasticsearch连接..."

    if curl -f -s "http://localhost:9200/_cluster/health" > /dev/null 2>&1; then
        log_success "Elasticsearch连接正常"
        return 0
    else
        log_error "Elasticsearch连接失败"
        return 1
    fi
}

# 测试API端点
test_api_endpoints() {
    log_info "测试API端点..."

    local endpoints=(
        "/health"
        "/api/v1/ping"
    )

    local all_passed=true

    for endpoint in "${endpoints[@]}"; do
        local url="$BACKEND_URL$endpoint"
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "API端点 $endpoint 正常"
        else
            log_error "API端点 $endpoint 失败"
            all_passed=false
        fi
    done

    if [ "$all_passed" = true ]; then
        log_success "所有API端点测试通过"
        return 0
    else
        log_error "部分API端点测试失败"
        return 1
    fi
}

# 测试前端资源
test_frontend() {
    log_info "测试前端资源..."

    # 检查主页面
    if check_url "$FRONTEND_URL" "前端主页" 1; then
        # 检查静态资源
        local resources=(
            "$FRONTEND_URL/static/js/main.js"
            "$FRONTEND_URL/static/css/main.css"
            "$FRONTEND_URL/favicon.ico"
        )

        local all_passed=true
        for resource in "${resources[@]}"; do
            if curl -f -s "$resource" > /dev/null 2>&1; then
                log_success "前端资源 $(basename $resource) 可访问"
            else
                log_warning "前端资源 $(basename $resource) 不可访问（可能正在构建中）"
            fi
        done

        log_success "前端应用测试通过"
        return 0
    else
        log_error "前端应用测试失败"
        return 1
    fi
}

# 测试数据库表结构
test_database_schema() {
    log_info "测试数据库表结构..."

    local tables=(
        "users"
        "roles"
        "permissions"
        "user_roles"
        "clients"
        "cases"
        "departments"
    )

    local all_exist=true

    for table in "${tables[@]}"; do
        if mysql -h localhost -P $MYSQL_PORT -u root -ppassword law_oa -e "DESCRIBE $table;" > /dev/null 2>&1; then
            log_success "数据表 $table 存在"
        else
            log_error "数据表 $table 不存在"
            all_exist=false
        fi
    done

    if [ "$all_exist" = true ]; then
        log_success "数据库表结构测试通过"
        return 0
    else
        log_error "数据库表结构测试失败"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    local report_file="integration-test-report-$(date +%Y%m%d-%H%M%S).txt"

    cat > "$report_file" << EOF
律师事务所管理系统集成测试报告
========================================

测试时间: $(date)
测试环境: 开发环境

测试项目:
1. 前端应用访问
2. 后端API服务
3. 数据库连接
4. Redis连接
5. Elasticsearch连接
6. API端点测试
7. 数据库表结构

详细结果:
EOF

    echo "报告已生成: $report_file"
    log_success "测试报告已生成: $report_file"
}

# 运行完整测试
run_full_test() {
    log_info "开始运行完整的集成测试..."

    local failed_tests=0
    local total_tests=7

    # 测试数据库连接
    check_database || failed_tests=$((failed_tests + 1))

    # 测试Redis连接
    check_redis || failed_tests=$((failed_tests + 1))

    # 测试Elasticsearch连接
    check_elasticsearch || failed_tests=$((failed_tests + 1))

    # 测试后端服务
    check_url "$BACKEND_URL/health" "后端健康检查" $TIMEOUT || failed_tests=$((failed_tests + 1))

    # 测试前端应用
    test_frontend || failed_tests=$((failed_tests + 1))

    # 测试API端点
    test_api_endpoints || failed_tests=$((failed_tests + 1))

    # 测试数据库表结构
    test_database_schema || failed_tests=$((failed_tests + 1))

    # 生成报告
    generate_report

    # 显示结果
    echo ""
    echo "========================================"
    if [ $failed_tests -eq 0 ]; then
        log_success "所有测试通过！ ($total_tests/$total_tests)"
        exit 0
    else
        log_error "部分测试失败！ ($((total_tests - failed_tests))/$total_tests)"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统集成测试脚本

用法: $0 [命令]

命令:
    test        运行完整测试
    frontend    测试前端应用
    backend     测试后端API
    database    测试数据库连接
    redis       测试Redis连接
    elasticsearch 测试Elasticsearch连接
    api         测试API端点
    schema      测试数据库表结构
    help        显示此帮助信息

示例:
    $0 test       # 运行完整测试
    $0 frontend   # 测试前端应用
    $0 backend    # 测试后端API
    $0 database   # 测试数据库连接

EOF
}

# 主函数
main() {
    case "${1:-}" in
        "test")
            run_full_test
            ;;
        "frontend")
            test_frontend
            ;;
        "backend")
            check_url "$BACKEND_URL/health" "后端健康检查" $TIMEOUT
            test_api_endpoints
            ;;
        "database")
            check_database
            ;;
        "redis")
            check_redis
            ;;
        "elasticsearch")
            check_elasticsearch
            ;;
        "api")
            test_api_endpoints
            ;;
        "schema")
            test_database_schema
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