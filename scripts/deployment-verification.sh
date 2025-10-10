#!/bin/bash

# 律师事务所管理系统部署验证脚本
# 用于验证部署后的系统状态和功能

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOYMENT_REPORT="$PROJECT_ROOT/deployment-verification-$(date +%Y%m%d_%H%M%S).md"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$DEPLOYMENT_REPORT"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$DEPLOYMENT_REPORT"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$DEPLOYMENT_REPORT"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$DEPLOYMENT_REPORT"
}

log_section() {
    echo -e "\n${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}" | tee -a "$DEPLOYMENT_REPORT"
}

# 检查URL可访问性
check_url() {
    local url=$1
    local description=$2
    local max_attempts=$3
    local attempt=1

    log_info "检查 $description ($url)..."

    while [ $attempt -le $max_attempts ]; do
        if curl -f -s --max-time 10 "$url" > /dev/null 2>&1; then
            log_success "$description 可访问"
            return 0
        fi

        log_warning "$description 不可访问，等待重试... ($attempt/$max_attempts)"
        sleep 5
        attempt=$((attempt + 1))
    done

    log_error "$description 不可访问"
    return 1
}

# 检查服务健康状态
check_service_health() {
    log_section "检查服务健康状态"

    local services=(
        "http://localhost:8080/health:后端健康检查"
        "http://localhost:8080/api/v1/ping:API心跳检查"
        "http://localhost:3003:前端应用检查"
    )

    local failed_services=0

    for service_info in "${services[@]}"; do
        local url=$(echo "$service_info" | cut -d: -f1-2)
        local description=$(echo "$service_info" | cut -d: -f3-)

        if check_url "$url" "$description" 3; then
            log_success "$description 正常"
        else
            log_error "$description 异常"
            failed_services=$((failed_services + 1))
        fi
    done

    if [ $failed_services -eq 0 ]; then
        log_success "所有服务健康状态正常"
        return 0
    else
        log_error "部分服务健康状态异常"
        return 1
    fi
}

# 检查数据库连接
check_database_connection() {
    log_section "检查数据库连接"

    # 检查MySQL连接
    log_info "检查MySQL数据库连接..."
    if mysql -h localhost -P 33060 -u root -ppassword -e "SELECT 1;" > /dev/null 2>&1; then
        log_success "MySQL数据库连接正常"

        # 检查数据库表结构
        log_info "检查数据库表结构..."
        local tables=("users" "roles" "clients" "cases" "departments")
        local missing_tables=0

        for table in "${tables[@]}"; do
            if mysql -h localhost -P 33060 -u root -ppassword law_oa -e "DESCRIBE $table;" > /dev/null 2>&1; then
                log_success "数据表 $table 存在"
            else
                log_error "数据表 $table 不存在"
                missing_tables=$((missing_tables + 1))
            fi
        done

        if [ $missing_tables -eq 0 ]; then
            log_success "数据库表结构完整"
        else
            log_error "数据库表结构不完整"
            return 1
        fi
    else
        log_error "MySQL数据库连接失败"
        return 1
    fi

    return 0
}

# 检查Redis连接
check_redis_connection() {
    log_section "检查Redis连接"

    if redis-cli -p 6379 ping > /dev/null 2>&1; then
        log_success "Redis连接正常"

        # 检查Redis信息
        log_info "检查Redis信息..."
        redis-cli -p 6379 info server | head -3
        return 0
    else
        log_error "Redis连接失败"
        return 1
    fi
}

# 检查Elasticsearch连接
check_elasticsearch_connection() {
    log_section "检查Elasticsearch连接"

    if curl -f -s "http://localhost:9200/_cluster/health" > /dev/null 2>&1; then
        log_success "Elasticsearch连接正常"

        # 检查集群健康状态
        log_info "检查Elasticsearch集群健康状态..."
        curl -s "http://localhost:9200/_cluster/health" | jq -r '.status'
        return 0
    else
        log_error "Elasticsearch连接失败"
        return 1
    fi
}

# 测试API端点
test_api_endpoints() {
    log_section "测试API端点"

    local endpoints=(
        "GET:http://localhost:8080/health:健康检查"
        "GET:http://localhost:8080/api/v1/ping:API心跳"
        "GET:http://localhost:8080/api/v1/users:用户列表"
        "GET:http://localhost:8080/api/v1/clients:客户列表"
        "GET:http://localhost:8080/api/v1/cases:案件列表"
    )

    local failed_endpoints=0

    for endpoint_info in "${endpoints[@]}"; do
        local method=$(echo "$endpoint_info" | cut -d: -f1)
        local url=$(echo "$endpoint_info" | cut -d: -f2)
        local description=$(echo "$endpoint_info" | cut -d: -f3-)

        log_info "测试 $description..."

        if curl -f -s -X "$method" "$url" > /dev/null 2>&1; then
            log_success "$description 正常"
        else
            log_error "$description 失败"
            failed_endpoints=$((failed_endpoints + 1))
        fi
    done

    if [ $failed_endpoints -eq 0 ]; then
        log_success "所有API端点测试通过"
        return 0
    else
        log_error "部分API端点测试失败"
        return 1
    fi
}

# 测试用户认证流程
test_authentication_flow() {
    log_section "测试用户认证流程"

    # 测试用户注册
    log_info "测试用户注册..."
    local register_response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "测试用户",
            "email": "test@example.com",
            "password": "password123",
            "role": "admin"
        }')

    if echo "$register_response" | jq -e '.success' > /dev/null 2>&1; then
        log_success "用户注册测试通过"
    else
        log_warning "用户注册测试失败 (用户可能已存在)"
    fi

    # 测试用户登录
    log_info "测试用户登录..."
    local login_response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "admin@example.com",
            "password": "admin123"
        }')

    if echo "$login_response" | jq -e '.success' > /dev/null 2>&1; then
        log_success "用户登录测试通过"
        local token=$(echo "$login_response" | jq -r '.data.token')

        # 测试认证API
        log_info "测试认证API..."
        if curl -s -H "Authorization: Bearer $token" "http://localhost:8080/api/v1/users/profile" > /dev/null 2>&1; then
            log_success "认证API测试通过"
        else
            log_error "认证API测试失败"
        fi
    else
        log_error "用户登录测试失败"
        return 1
    fi

    return 0
}

# 测试前端页面加载
test_frontend_pages() {
    log_section "测试前端页面加载"

    local pages=(
        "http://localhost:3003:主页"
        "http://localhost:3003/login:登录页"
        "http://localhost:3003/dashboard:仪表盘"
        "http://localhost:3003/cases:案件管理"
        "http://localhost:3003/clients:客户管理"
    )

    local failed_pages=0

    for page_info in "${pages[@]}"; do
        local url=$(echo "$page_info" | cut -d: -f1)
        local description=$(echo "$page_info" | cut -d: -f2-)

        log_info "测试 $description..."

        if curl -f -s --max-time 10 "$url" > /dev/null 2>&1; then
            log_success "$description 加载正常"
        else
            log_error "$description 加载失败"
            failed_pages=$((failed_pages + 1))
        fi
    done

    if [ $failed_pages -eq 0 ]; then
        log_success "所有前端页面加载正常"
        return 0
    else
        log_error "部分前端页面加载失败"
        return 1
    fi
}

# 检查Docker容器状态
check_docker_containers() {
    log_section "检查Docker容器状态"

    cd "$PROJECT_ROOT"

    if [ -f "docker-compose.yml" ]; then
        log_info "检查Docker容器状态..."

        # 获取容器状态
        local containers=$(docker-compose ps --services)
        local stopped_containers=0

        for container in $containers; do
            local status=$(docker-compose ps -q $container | xargs docker inspect --format='{{.State.Status}}' 2>/dev/null || echo "not_found")

            if [ "$status" = "running" ]; then
                log_success "容器 $container 正在运行"
            else
                log_warning "容器 $container 状态: $status"
                stopped_containers=$((stopped_containers + 1))
            fi
        done

        if [ $stopped_containers -eq 0 ]; then
            log_success "所有容器正常运行"
        else
            log_warning "部分容器未正常运行"
        fi
    else
        log_warning "Docker Compose配置文件不存在"
    fi

    return 0
}

# 检查系统资源使用情况
check_system_resources() {
    log_section "检查系统资源使用情况"

    # 检查磁盘空间
    log_info "检查磁盘空间..."
    df -h | grep -E "(/$|/var|/tmp)" | head -3

    # 检查内存使用
    log_info "检查内存使用..."
    if command -v free &> /dev/null; then
        free -h
    elif command -v vm_stat &> /dev/null; then
        vm_stat
    fi

    # 检查CPU使用
    log_info "检查CPU使用..."
    if command -v top &> /dev/null; then
        top -l 1 | grep "CPU usage"
    elif command -v htop &> /dev/null; then
        htop -n 1
    fi

    return 0
}

# 生成部署验证报告
generate_deployment_report() {
    log_section "生成部署验证报告"

    cat >> "$DEPLOYMENT_REPORT" << EOF

## 部署验证总结

### 验证时间
$(date)

### 验证项目
1. 服务健康状态检查
2. 数据库连接验证
3. Redis连接验证
4. Elasticsearch连接验证
5. API端点测试
6. 用户认证流程测试
7. 前端页面加载测试
8. Docker容器状态检查
9. 系统资源使用检查

### 验证结果
- **总体状态**: 部署验证完成
- **详细信息**: 请查看上方各检查项的结果
- **建议**: 根据测试结果进行必要的调整和优化

### 下一步行动
1. 修复所有失败的测试项
2. 优化系统性能和资源使用
3. 监控系统运行状态
4. 定期进行部署验证

EOF

    log_success "部署验证报告已生成: $DEPLOYMENT_REPORT"
}

# 运行完整部署验证
run_full_deployment_verification() {
    log_info "开始运行完整的部署验证..."

    local total_tests=9
    local current_test=0
    local failed_tests=0

    # 检查服务健康状态
    current_test=$((current_test + 1))
    check_service_health || failed_tests=$((failed_tests + 1))

    # 检查数据库连接
    current_test=$((current_test + 1))
    check_database_connection || failed_tests=$((failed_tests + 1))

    # 检查Redis连接
    current_test=$((current_test + 1))
    check_redis_connection || failed_tests=$((failed_tests + 1))

    # 检查Elasticsearch连接
    current_test=$((current_test + 1))
    check_elasticsearch_connection || failed_tests=$((failed_tests + 1))

    # 测试API端点
    current_test=$((current_test + 1))
    test_api_endpoints || failed_tests=$((failed_tests + 1))

    # 测试用户认证流程
    current_test=$((current_test + 1))
    test_authentication_flow || failed_tests=$((failed_tests + 1))

    # 测试前端页面加载
    current_test=$((current_test + 1))
    test_frontend_pages || failed_tests=$((failed_tests + 1))

    # 检查Docker容器状态
    current_test=$((current_test + 1))
    check_docker_containers || failed_tests=$((failed_tests + 1))

    # 检查系统资源使用情况
    current_test=$((current_test + 1))
    check_system_resources || failed_tests=$((failed_tests + 1))

    # 生成报告
    generate_deployment_report

    # 显示结果
    printf "\n\n"
    echo "========================================"
    if [ $failed_tests -eq 0 ]; then
        log_success "所有部署验证通过！ ($total_tests/$total_tests)"
        echo "========================================"
        log_info "系统部署成功，所有功能正常！"
        exit 0
    else
        log_error "部分部署验证失败！ ($((total_tests - failed_tests))/$total_tests)"
        echo "========================================"
        log_info "请检查失败的项目并修复问题"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统部署验证脚本

用法: $0 [命令]

命令:
    full        运行完整部署验证 (默认)
    health      检查服务健康状态
    database    检查数据库连接
    redis       检查Redis连接
    elasticsearch 检查Elasticsearch连接
    api         测试API端点
    auth        测试用户认证流程
    frontend    测试前端页面加载
    docker      检查Docker容器状态
    resources   检查系统资源使用情况
    report      生成部署验证报告
    help        显示此帮助信息

示例:
    $0 full          # 运行完整部署验证
    $0 health        # 检查服务健康状态
    $0 api           # 测试API端点
    $0 frontend      # 测试前端页面加载
    $0 docker        # 检查Docker容器状态

EOF
}

# 主函数
main() {
    case "${1:-full}" in
        "full")
            run_full_deployment_verification
            ;;
        "health")
            check_service_health
            ;;
        "database")
            check_database_connection
            ;;
        "redis")
            check_redis_connection
            ;;
        "elasticsearch")
            check_elasticsearch_connection
            ;;
        "api")
            test_api_endpoints
            ;;
        "auth")
            test_authentication_flow
            ;;
        "frontend")
            test_frontend_pages
            ;;
        "docker")
            check_docker_containers
            ;;
        "resources")
            check_system_resources
            ;;
        "report")
            generate_deployment_report
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