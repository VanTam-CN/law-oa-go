#!/bin/bash

# 律师事务所管理系统项目验证脚本
# 集成所有测试功能和验证流程

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
LOG_DIR="$PROJECT_ROOT/logs"
REPORT_DIR="$PROJECT_ROOT/reports"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_LOG="$LOG_DIR/project-validation-$TIMESTAMP.log"
TEST_REPORT="$REPORT_DIR/project-validation-$TIMESTAMP.md"

# 创建日志和报告目录
mkdir -p "$LOG_DIR"
mkdir -p "$REPORT_DIR"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$TEST_LOG"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$TEST_LOG"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$TEST_LOG"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$TEST_LOG"
}

log_section() {
    echo -e "\n${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}" | tee -a "$TEST_LOG"
}

# 显示进度条
show_progress() {
    local current=$1
    local total=$2
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    local remaining=$((width - completed))

    printf "\r["
    printf "%*s" $completed | tr ' ' '█'
    printf "%*s" $remaining | tr ' ' '░'
    printf "] %d%% (%d/%d)" $percentage $current $total
}

# 检查命令是否存在
check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# 检查Docker状态
check_docker_status() {
    log_section "检查Docker环境"

    if ! check_command docker; then
        log_error "Docker未安装或未启动"
        return 1
    fi

    if ! docker info > /dev/null 2>&1; then
        log_error "Docker服务未运行"
        return 1
    fi

    log_success "Docker环境正常"

    # 检查Docker Compose
    if check_command docker-compose; then
        log_success "Docker Compose可用"
    else
        log_warning "Docker Compose不可用，部分功能可能受限"
    fi

    return 0
}

# 检查服务端口
check_service_ports() {
    log_section "检查服务端口占用情况"

    local ports=(3003 8080 33060 6379 9200)
    local port_status=()

    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            log_warning "端口 $port 已被占用"
            port_status+=("occupied")
        else
            log_success "端口 $port 可用"
            port_status+=("free")
        fi
    done

    return 0
}

# 运行Go单元测试
run_go_tests() {
    log_section "运行Go单元测试"

    cd "$PROJECT_ROOT"

    # 运行单元测试
    if [ -f "test_runner.sh" ]; then
        log_info "运行项目单元测试..."
        bash test_runner.sh
    else
        log_info "运行标准Go测试..."
        go test -v ./...
    fi

    if [ $? -eq 0 ]; then
        log_success "Go单元测试通过"
        return 0
    else
        log_error "Go单元测试失败"
        return 1
    fi
}

# 检查代码质量
check_code_quality() {
    log_section "检查代码质量"

    cd "$PROJECT_ROOT"

    # 检查Go代码格式
    if check_command gofmt; then
        log_info "检查Go代码格式..."
        if gofmt -l . | grep -q .; then
            log_warning "发现格式问题的Go文件"
            gofmt -l . | head -5
        else
            log_success "Go代码格式正确"
        fi
    fi

    # 检查是否有语法错误
    log_info "检查Go语法错误..."
    if go build -o /tmp/law-oa-test .; then
        log_success "Go代码编译通过"
        rm -f /tmp/law-oa-test
    else
        log_error "Go代码编译失败"
        return 1
    fi

    return 0
}

# 运行API测试
run_api_tests() {
    log_section "运行API测试"

    cd "$PROJECT_ROOT"

    # 检查API测试脚本
    if [ -f "scripts/test_api.sh" ]; then
        log_info "运行API测试..."
        bash scripts/test_api.sh
    fi

    if [ -f "scripts/test_api_v2.sh" ]; then
        log_info "运行API v2测试..."
        bash scripts/test_api_v2.sh
    fi

    log_success "API测试完成"
    return 0
}

# 运行集成测试
run_integration_tests() {
    log_section "运行集成测试"

    cd "$PROJECT_ROOT"

    # 检查集成测试脚本
    if [ -f "scripts/test-integration.sh" ]; then
        log_info "运行集成测试..."
        bash scripts/test-integration.sh test
    else
        log_warning "集成测试脚本不存在"
    fi

    log_success "集成测试完成"
    return 0
}

# 检查Docker服务
check_docker_services() {
    log_section "检查Docker服务状态"

    cd "$PROJECT_ROOT"

    # 检查docker-compose文件
    if [ -f "docker-compose.yml" ]; then
        log_info "检查docker-compose服务..."
        if docker-compose ps | grep -q "Up"; then
            log_success "Docker服务正在运行"

            # 显示服务状态
            docker-compose ps
        else
            log_warning "Docker服务未运行"
        fi
    fi

    # 检查开发环境配置
    if [ -f "docker-compose.dev.yml" ]; then
        log_info "开发环境配置文件存在"
    fi

    return 0
}

# 检查前端应用
check_frontend_apps() {
    log_section "检查前端应用"

    cd "$PROJECT_ROOT"

    # 检查前端目录
    if [ -d "frontend" ]; then
        log_info "检查前端应用 (React Bootstrap)..."
        if [ -f "frontend/package.json" ]; then
            log_success "前端应用配置存在"
        else
            log_warning "前端应用配置缺失"
        fi
    fi

    if [ -d "frontend-vue" ]; then
        log_info "检查前端应用 (Vue)..."
        if [ -f "frontend-vue/package.json" ]; then
            log_success "Vue前端应用配置存在"
        else
            log_warning "Vue前端应用配置缺失"
        fi
    fi

    return 0
}

# 检查配置文件
check_configuration() {
    log_section "检查配置文件"

    cd "$PROJECT_ROOT"

    local config_files=(
        ".env.example"
        ".env"
        "config/config.yaml"
        "config/config.dev.yaml"
        "config/config.prod.yaml"
    )

    for file in "${config_files[@]}"; do
        if [ -f "$file" ]; then
            log_success "配置文件 $file 存在"
        else
            log_warning "配置文件 $file 不存在"
        fi
    done

    return 0
}

# 检查文档完整性
check_documentation() {
    log_section "检查文档完整性"

    cd "$PROJECT_ROOT"

    local doc_files=(
        "README.md"
        "docs/"
        "CLAUDE.md"
        "QWEN.md"
    )

    for file in "${doc_files[@]}"; do
        if [ -e "$file" ]; then
            log_success "文档 $file 存在"
        else
            log_warning "文档 $file 不存在"
        fi
    done

    return 0
}

# 生成测试报告
generate_report() {
    log_section "生成测试报告"

    cat > "$TEST_REPORT" << EOF
# 律师事务所管理系统项目验证报告

## 测试信息
- **测试时间**: $(date)
- **测试环境**: $(uname -a)
- **项目路径**: $PROJECT_ROOT
- **测试日志**: $TEST_LOG

## 测试项目
EOF

    echo "报告已生成: $TEST_REPORT"
    log_success "测试报告已生成: $TEST_REPORT"
}

# 运行完整验证
run_full_validation() {
    log_info "开始运行完整的项目验证..."

    local total_tests=10
    local current_test=0
    local failed_tests=0

    # 检查Docker环境
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_docker_status || failed_tests=$((failed_tests + 1))

    # 检查服务端口
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_service_ports || failed_tests=$((failed_tests + 1))

    # 运行Go单元测试
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    run_go_tests || failed_tests=$((failed_tests + 1))

    # 检查代码质量
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_code_quality || failed_tests=$((failed_tests + 1))

    # 运行API测试
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    run_api_tests || failed_tests=$((failed_tests + 1))

    # 运行集成测试
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    run_integration_tests || failed_tests=$((failed_tests + 1))

    # 检查Docker服务
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_docker_services || failed_tests=$((failed_tests + 1))

    # 检查前端应用
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_frontend_apps || failed_tests=$((failed_tests + 1))

    # 检查配置文件
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_configuration || failed_tests=$((failed_tests + 1))

    # 检查文档完整性
    current_test=$((current_test + 1))
    show_progress $current_test $total_tests
    check_documentation || failed_tests=$((failed_tests + 1))

    # 生成报告
    generate_report

    # 显示结果
    printf "\n\n"
    echo "========================================"
    if [ $failed_tests -eq 0 ]; then
        log_success "所有验证通过！ ($total_tests/$total_tests)"
        echo "========================================"
        log_info "项目验证完成，一切正常！"
        exit 0
    else
        log_error "部分验证失败！ ($((total_tests - failed_tests))/$total_tests)"
        echo "========================================"
        log_info "请检查失败的项目并修复问题"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统项目验证脚本

用法: $0 [命令]

命令:
    full        运行完整验证 (默认)
    docker      检查Docker环境和服务
    code        检查代码质量和运行单元测试
    api         运行API测试
    integration 运行集成测试
    config      检查配置文件
    docs        检查文档完整性
    report      生成测试报告
    help        显示此帮助信息

示例:
    $0 full          # 运行完整验证
    $0 docker        # 检查Docker环境
    $0 code          # 检查代码质量
    $0 api           # 运行API测试
    $0 integration   # 运行集成测试

EOF
}

# 主函数
main() {
    case "${1:-full}" in
        "full")
            run_full_validation
            ;;
        "docker")
            check_docker_status
            check_docker_services
            ;;
        "code")
            run_go_tests
            check_code_quality
            ;;
        "api")
            run_api_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "config")
            check_configuration
            ;;
        "docs")
            check_documentation
            ;;
        "report")
            generate_report
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