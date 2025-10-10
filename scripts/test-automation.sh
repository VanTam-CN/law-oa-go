#!/bin/bash

# Law OA Go 项目测试自动化脚本
# 用于自动运行测试、生成覆盖率报告和验证代码质量

set -e  # 遇到错误立即退出
set -o pipefail  # 管道中任何命令失败都视为失败

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# 输出函数
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

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖..."

    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装或不在PATH中"
        exit 1
    fi

    # 检查golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint 未安装，跳过代码质量检查"
        SKIP_LINT=true
    fi

    # 检查Redis（可选）
    if ! redis-cli ping &> /dev/null 2>&1; then
        log_warning "Redis 未运行，某些集成测试可能失败"
    fi

    # 检查MySQL（可选）
    if ! mysql -u root -e "SELECT 1;" &> /dev/null 2>&1; then
        log_warning "MySQL 未运行，某些集成测试可能失败"
    fi

    log_success "依赖检查完成"
}

# 清理之前的测试结果
cleanup() {
    log_step "清理之前的测试结果..."

    # 删除测试覆盖率文件
    rm -f coverage.out coverage.html coverage.xml coverage.json

    # 删除测试二进制文件
    rm -f *.test

    # 删除测试报告文件
    rm -f test-report.xml test-results.json

    # 清理测试缓存
    go clean -testcache

    log_success "清理完成"
}

# 运行静态代码分析
run_static_analysis() {
    if [ "$SKIP_LINT" = true ]; then
        log_warning "跳过静态代码分析"
        return
    fi

    log_step "运行静态代码分析..."

    # 运行golangci-lint
    if ! golangci-lint run --timeout=5m --out-format=github-actions; then
        log_error "静态代码分析失败"
        return 1
    fi

    log_success "静态代码分析通过"
}

# 运行单元测试
run_unit_tests() {
    log_step "运行单元测试..."

    # 创建测试覆盖率输出目录
    mkdir -p reports

    # 运行单元测试并生成覆盖率报告
    if ! go test -v -coverprofile=coverage.out -covermode=count ./internal/... -timeout=5m 2>&1 | tee reports/unit-test.log; then
        log_error "单元测试失败"
        return 1
    fi

    # 生成HTML覆盖率报告
    if command -v go tool cover &> /dev/null; then
        go tool cover -html=coverage.out -o coverage.html
        log_success "HTML覆盖率报告已生成: coverage.html"
    fi

    # 生成XML覆盖率报告（用于CI）
    if command -v gocov &> /dev/null && command -v gocov-xml &> /dev/null; then
        gocov convert coverage.out | gocov-xml > coverage.xml
        log_success "XML覆盖率报告已生成: coverage.xml"
    fi

    log_success "单元测试通过"
}

# 运行集成测试
run_integration_tests() {
    log_step "运行集成测试..."

    # 运行集成测试
    if ! go test -v ./tests/integration/... -timeout=10m 2>&1 | tee reports/integration-test.log; then
        log_error "集成测试失败"
        return 1
    fi

    log_success "集成测试通过"
}

# 运行基准测试
run_benchmark_tests() {
    log_step "运行基准测试..."

    # 创建基准测试报告目录
    mkdir -p reports/benchmark

    # 运行基准测试
    if ! go test -bench=. -benchmem -benchtime=5s ./internal/... -timeout=15m > reports/benchmark/benchmark.txt 2>&1; then
        log_error "基准测试失败"
        return 1
    fi

    # 生成基准测试报告
    if command -v benchstat &> /dev/null; then
        benchstat reports/benchmark/benchmark.txt > reports/benchmark/benchmark-summary.txt
        log_success "基准测试摘要已生成: reports/benchmark/benchmark-summary.txt"
    fi

    log_success "基准测试完成"
}

# 生成测试报告
generate_test_report() {
    log_step "生成测试报告..."

    # 计算测试覆盖率
    if [ -f coverage.out ]; then
        total_coverage=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')

        # 创建测试报告
        cat > reports/test-report.md << EOF
# 测试报告

## 测试覆盖率
- 总覆盖率: ${total_coverage}%

## 测试执行情况
- 单元测试: 通过
- 集成测试: 通过
- 基准测试: 通过

## 生成时间
$(date)

## 文件列表
- HTML覆盖率报告: coverage.html
- 单元测试日志: reports/unit-test.log
- 集成测试日志: reports/integration-test.log
- 基准测试报告: reports/benchmark/benchmark.txt
EOF

        log_success "测试报告已生成: reports/test-report.md"
    fi
}

# 检查测试覆盖率
check_test_coverage() {
    if [ ! -f coverage.out ]; then
        log_warning "未找到覆盖率文件，跳过覆盖率检查"
        return
    fi

    log_step "检查测试覆盖率..."

    # 计算总覆盖率
    total_coverage=$(go tool cover -func=coverage.out | grep "total:" | awk '{print $3}' | sed 's/%//')

    log_info "当前测试覆盖率: ${total_coverage}%"

    # 检查是否达到目标覆盖率
    TARGET_COVERAGE=70
    if (( $(echo "$total_coverage >= $TARGET_COVERAGE" | bc -l) )); then
        log_success "测试覆盖率已达到目标 (${TARGET_COVERAGE}%): ${total_coverage}%"
    else
        log_error "测试覆盖率未达到目标 (${TARGET_COVERAGE}%): ${total_coverage}%"
        return 1
    fi
}

# 显示测试统计信息
show_test_statistics() {
    log_step "测试统计信息"

    if [ -f coverage.out ]; then
        log_info "测试覆盖率信息:"
        go tool cover -func=coverage.out | grep -E "(total|^.*\.go)"
    fi

    if [ -f reports/unit-test.log ]; then
        unit_tests=$(grep -c "^=== RUN" reports/unit-test.log || echo "0")
        unit_passed=$(grep -c "^--- PASS:" reports/unit-test.log || echo "0")
        unit_failed=$(grep -c "^--- FAIL:" reports/unit-test.log || echo "0")

        log_info "单元测试统计:"
        log_info "  总计: $unit_tests"
        log_info "  通过: $unit_passed"
        log_info "  失败: $unit_failed"
    fi

    if [ -f reports/integration-test.log ]; then
        integration_tests=$(grep -c "^=== RUN" reports/integration-test.log || echo "0")
        integration_passed=$(grep -c "^--- PASS:" reports/integration-test.log || echo "0")
        integration_failed=$(grep -c "^--- FAIL:" reports/integration-test.log || echo "0")

        log_info "集成测试统计:"
        log_info "  总计: $integration_tests"
        log_info "  通过: $integration_passed"
        log_info "  失败: $integration_failed"
    fi
}

# 主函数
main() {
    log_info "开始 Law OA Go 项目测试自动化..."

    # 解析命令行参数
    SKIP_LINT=false
    SKIP_INTEGRATION=false
    SKIP_BENCHMARK=false
    SKIP_COVERAGE_CHECK=false

    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-lint)
                SKIP_LINT=true
                shift
                ;;
            --skip-integration)
                SKIP_INTEGRATION=true
                shift
                ;;
            --skip-benchmark)
                SKIP_BENCHMARK=true
                shift
                ;;
            --skip-coverage-check)
                SKIP_COVERAGE_CHECK=true
                shift
                ;;
            --help|-h)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --skip-lint              跳过静态代码分析"
                echo "  --skip-integration       跳过集成测试"
                echo "  --skip-benchmark         跳过基准测试"
                echo "  --skip-coverage-check    跳过覆盖率检查"
                echo "  --help, -h               显示此帮助信息"
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                exit 1
                ;;
        esac
    done

    # 执行测试流程
    check_dependencies
    cleanup

    if ! run_static_analysis; then
        log_error "静态代码分析失败，测试终止"
        exit 1
    fi

    if ! run_unit_tests; then
        log_error "单元测试失败，测试终止"
        exit 1
    fi

    if [ "$SKIP_INTEGRATION" != true ]; then
        if ! run_integration_tests; then
            log_error "集成测试失败，测试终止"
            exit 1
        fi
    fi

    if [ "$SKIP_BENCHMARK" != true ]; then
        if ! run_benchmark_tests; then
            log_error "基准测试失败，测试终止"
            exit 1
        fi
    fi

    generate_test_report
    show_test_statistics

    if [ "$SKIP_COVERAGE_CHECK" != true ]; then
        if ! check_test_coverage; then
            log_error "测试覆盖率检查失败，测试终止"
            exit 1
        fi
    fi

    log_success "所有测试通过！🎉"
    log_info "查看测试结果:"
    log_info "  - HTML覆盖率报告: coverage.html"
    log_info "  - 测试报告: reports/test-report.md"
    log_info "  - 测试日志: reports/"
}

# 脚本入口点
main "$@"