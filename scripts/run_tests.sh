#!/bin/bash

# Law OA Go 项目自动化测试执行脚本
# 支持分层测试、并行执行、质量门禁和报告生成

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_progress() {
    echo -e "${PURPLE}[PROGRESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                     显示此帮助信息"
    echo "  -e, --env ENVIRONMENT          测试环境 (dev|staging|production)"
    echo "  -t, --test-type TYPE           测试类型 (unit|integration|e2e|performance|security|all)"
    echo "  -p, --parallel                 启用并行测试执行"
    echo "  -j, --jobs NUM                 并行任务数 (默认: CPU核心数)"
    echo "  -r, --report FILE              生成测试报告文件"
    echo "  -q, --quality-gate             启用质量门禁检查"
    echo "  -s, --strict                   严格模式（质量门禁失败则退出）"
    echo "  --coverage                    生成测试覆盖率报告"
    echo "  --bench                       运行基准测试"
    echo "  --fuzz                        运行模糊测试"
    echo "  --profile                     生成性能分析报告"
    echo "  --timeout TIMEOUT             测试超时时间 (默认: 30m)"
    echo "  --skip-build                  跳过构建步骤"
    echo "  --skip-cleanup                跳过清理步骤"
    echo "  --dry-run                     仅显示将要执行的操作"
    echo ""
    echo "测试类型说明:"
    echo "  unit          - 单元测试 (快速，核心逻辑)"
    echo "  integration   - 集成测试 (中等，组件交互)"
    echo "  e2e           - 端到端测试 (慢速，完整流程)"
    echo "  performance   - 性能测试 (基准测试，负载测试)"
    echo "  security      - 安全测试 (漏洞扫描，模糊测试)"
    echo "  all           - 所有测试类型 (默认)"
    echo ""
    echo "示例:"
    echo "  $0 -e dev -t unit --coverage                    # 开发环境单元测试"
    echo "  $0 -e staging -t integration -p -j 4          # 测试环境并行集成测试"
    echo "  $0 -e production -t all -q -s --bench --fuzz   # 生产环境完整测试"
    echo "  $0 --dry-run -t performance                    # 性能测试预览"
}

# 初始化变量
ENVIRONMENT="dev"
TEST_TYPE="all"
PARALLEL=false
JOBS=""
REPORT_FILE=""
QUALITY_GATE=false
STRICT_MODE=false
COVERAGE=false
BENCH=false
FUZZ=false
PROFILE=false
TIMEOUT="30m"
SKIP_BUILD=false
SKIP_CLEANUP=false
DRY_RUN=false

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -t|--test-type)
            TEST_TYPE="$2"
            shift 2
            ;;
        -p|--parallel)
            PARALLEL=true
            shift
            ;;
        -j|--jobs)
            JOBS="$2"
            shift 2
            ;;
        -r|--report)
            REPORT_FILE="$2"
            shift 2
            ;;
        -q|--quality-gate)
            QUALITY_GATE=true
            shift
            ;;
        -s|--strict)
            STRICT_MODE=true
            shift
            ;;
        --coverage)
            COVERAGE=true
            shift
            ;;
        --bench)
            BENCH=true
            shift
            ;;
        --fuzz)
            FUZZ=true
            shift
            ;;
        --profile)
            PROFILE=true
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-cleanup)
            SKIP_CLEANUP=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            log_error "未知参数: $1"
            show_usage
            exit 1
            ;;
    esac
done

# 验证环境参数
if [[ ! "$ENVIRONMENT" =~ ^(dev|staging|production)$ ]]; then
    log_error "无效的环境参数: $ENVIRONMENT"
    echo "支持的环境: dev, staging, production"
    exit 1
fi

# 验证测试类型
if [[ ! "$TEST_TYPE" =~ ^(unit|integration|e2e|performance|security|all)$ ]]; then
    log_error "无效的测试类型: $TEST_TYPE"
    echo "支持的测试类型: unit, integration, e2e, performance, security, all"
    exit 1
fi

# 获取CPU核心数
if [ -z "$JOBS" ]; then
    if command -v nproc &> /dev/null; then
        JOBS=$(nproc)
    else
        JOBS=4
    fi
fi

# 设置Go环境变量
export GO111MODULE=on
export CGO_ENABLED=1
export GIN_MODE=release
export ENVIRONMENT=$ENVIRONMENT

# 创建测试报告目录
mkdir -p test-reports
mkdir -p profiles

# 构建应用
build_application() {
    if [ "$SKIP_BUILD" = true ]; then
        log_info "跳过构建步骤"
        return 0
    fi
    
    log_info "构建应用..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go build -o bin/law-oa-server ."
        return 0
    fi
    
    # 清理之前的构建
    rm -f bin/law-oa-server
    
    # 构建应用
    if go build -o bin/law-oa-server .; then
        log_success "应用构建成功"
        return 0
    else
        log_error "应用构建失败"
        return 1
    fi
}

# 运行单元测试
run_unit_tests() {
    log_info "运行单元测试..."
    
    local args=""
    if [ "$PARALLEL" = true ]; then
        args="$args -parallel $JOBS"
    fi
    args="$args -v -race -timeout=$TIMEOUT"
    
    if [ "$COVERAGE" = true ]; then
        args="$args -coverprofile=test-reports/unit_coverage.out"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go test $args ./internal/..."
        return 0
    fi
    
    if go test $args ./internal/...; then
        log_success "单元测试通过"
        return 0
    else
        log_error "单元测试失败"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "运行集成测试..."
    
    local args=""
    if [ "$PARALLEL" = true ]; then
        args="$args -parallel $JOBS"
    fi
    args="$args -v -race -tags=integration -timeout=$TIMEOUT"
    
    if [ "$COVERAGE" = true ]; then
        args="$args -coverprofile=test-reports/integration_coverage.out"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go test $args ./tests/..."
        return 0
    fi
    
    if go test $args ./tests/...; then
        log_success "集成测试通过"
        return 0
    else
        log_error "集成测试失败"
        return 1
    fi
}

# 运行端到端测试
run_e2e_tests() {
    log_info "运行端到端测试..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] # 启动测试环境"
        echo "[DRY RUN] docker-compose -f docker-compose.test.yml up -d"
        echo "[DRY RUN] # 等待服务就绪"
        echo "[DRY RUN] # 运行E2E测试"
        echo "[DRY RUN] go test -v -timeout=$TIMEOUT ./tests/e2e/..."
        return 0
    fi
    
    # 启动测试环境
    docker-compose -f docker-compose.test.yml up -d
    
    # 等待服务就绪
    log_info "等待测试环境就绪..."
    sleep 30
    
    # 运行E2E测试
    if go test -v -timeout=$TIMEOUT ./tests/e2e/...; then
        log_success "端到端测试通过"
        docker-compose -f docker-compose.test.yml down
        return 0
    else
        log_error "端到端测试失败"
        docker-compose -f docker-compose.test.yml down
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "运行性能测试..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go test -bench=. -benchmem -timeout=$TIMEOUT ./tests/performance/..."
        return 0
    fi
    
    local args="-bench=. -benchmem -timeout=$TIMEOUT"
    
    if [ "$PROFILE" = true ]; then
        args="$args -cpuprofile=profiles/cpu.prof"
        args="$args -memprofile=profiles/mem.prof"
    fi
    
    if go test $args ./tests/performance/... > test-reports/performance.txt 2>&1; then
        log_success "性能测试通过"
        return 0
    else
        log_error "性能测试失败"
        return 1
    fi
}

# 运行基准测试
run_benchmark_tests() {
    log_info "运行基准测试..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go test -bench=. -benchmem -count=3 ./..."
        return 0
    fi
    
    if go test -bench=. -benchmem -count=3 ./... > test-reports/benchmark.txt 2>&1; then
        log_success "基准测试通过"
        return 0
    else
        log_error "基准测试失败"
        return 1
    fi
}

# 运行模糊测试
run_fuzz_tests() {
    log_info "运行模糊测试..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go test -fuzz=Fuzz -fuzztime=30s ./..."
        return 0
    fi
    
    # 创建模糊测试结果目录
    mkdir -p fuzzing_results
    
    if go test -fuzz=Fuzz -fuzztime=30s ./... > test-reports/fuzz.txt 2>&1; then
        log_success "模糊测试通过"
        return 0
    else
        log_error "模糊测试失败"
        return 1
    fi
}

# 运行安全测试
run_security_tests() {
    log_info "运行安全测试..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] # 运行安全扫描"
        echo "[DRY RUN] gosec -fmt=json -out=test-reports/security.json ./..."
        echo "[DRY RUN] # 运行漏洞检查"
        echo "[DRY RUN] govulncheck -json ./... > test-reports/vulnerabilities.json"
        return 0
    fi
    
    # 运行安全扫描
    if command -v gosec &> /dev/null; then
        log_info "运行gosec安全扫描..."
        gosec -fmt=json -out=test-reports/security.json ./...
    fi
    
    # 运行漏洞检查
    if command -v govulncheck &> /dev/null; then
        log_info "运行govulncheck漏洞检查..."
        govulncheck -json ./... > test-reports/vulnerabilities.json 2>/dev/null || true
    fi
    
    log_success "安全测试完成"
    return 0
}

# 生成覆盖率报告
generate_coverage_report() {
    if [ "$COVERAGE" = false ]; then
        return 0
    fi
    
    log_info "生成覆盖率报告..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html"
        return 0
    fi
    
    # 合并覆盖率文件
    local coverage_files=()
    if [ -f "test-reports/unit_coverage.out" ]; then
        coverage_files+=("test-reports/unit_coverage.out")
    fi
    if [ -f "test-reports/integration_coverage.out" ]; then
        coverage_files+=("test-reports/integration_coverage.out")
    fi
    
    if [ ${#coverage_files[@]} -gt 0 ]; then
        # 合并覆盖率文件
        echo "mode: set" > test-reports/coverage.out
        for file in "${coverage_files[@]}"; do
            tail -n +2 "$file" >> test-reports/coverage.out
        done
        
        # 生成HTML报告
        go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html
        
        # 生成文本报告
        go tool cover -func=test-reports/coverage.out > test-reports/coverage.txt
        
        log_success "覆盖率报告已生成"
    fi
}

# 质量门禁检查
quality_gate_check() {
    if [ "$QUALITY_GATE" = false ]; then
        return 0
    fi
    
    log_info "执行质量门禁检查..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] ./scripts/deployment_quality_gate.sh -e $ENVIRONMENT"
        return 0
    fi
    
    if [ -f "scripts/deployment_quality_gate.sh" ]; then
        chmod +x scripts/deployment_quality_gate.sh
        
        local gate_args="-e $ENVIRONMENT"
        if [ "$STRICT_MODE" = true ]; then
            gate_args="$gate_args -s"
        fi
        
        if ./scripts/deployment_quality_gate.sh $gate_args; then
            log_success "质量门禁检查通过"
            return 0
        else
            log_error "质量门禁检查失败"
            return 1
        fi
    else
        log_warning "质量门禁脚本不存在，跳过检查"
        return 0
    fi
}

# 生成测试报告
generate_test_report() {
    if [ -z "$REPORT_FILE" ]; then
        return 0
    fi
    
    log_info "生成测试报告: $REPORT_FILE"
    
    local report_content="# Law OA Go 测试报告

## 测试执行信息
- **执行时间**: $(date)
- **测试环境**: $ENVIRONMENT
- **测试类型**: $TEST_TYPE
- **并行执行**: $PARALLEL
- **并行任务数**: $JOBS
- **超时设置**: $TIMEOUT

## 测试结果概览
"
    
    # 添加各个测试的结果
    if [[ "$TEST_TYPE" =~ ^(unit|all)$ ]]; then
        report_content+="### 单元测试
- **状态**: ${UNIT_TEST_STATUS:-未执行}
- **覆盖率**: ${UNIT_COVERAGE:-0}%
"
    fi
    
    if [[ "$TEST_TYPE" =~ ^(integration|all)$ ]]; then
        report_content+="### 集成测试
- **状态**: ${INTEGRATION_TEST_STATUS:-未执行}
- **覆盖率**: ${INTEGRATION_COVERAGE:-0}%
"
    fi
    
    if [[ "$TEST_TYPE" =~ ^(e2e|all)$ ]]; then
        report_content+="### 端到端测试
- **状态**: ${E2E_TEST_STATUS:-未执行}
"
    fi
    
    if [[ "$TEST_TYPE" =~ ^(performance|all)$ ]]; then
        report_content+="### 性能测试
- **状态**: ${PERFORMANCE_TEST_STATUS:-未执行}
"
    fi
    
    if [ "$BENCH" = true ]; then
        report_content+="### 基准测试
- **状态**: ${BENCHMARK_TEST_STATUS:-未执行}
"
    fi
    
    if [ "$FUZZ" = true ]; then
        report_content+="### 模糊测试
- **状态**: ${FUZZ_TEST_STATUS:-未执行}
"
    fi
    
    if [[ "$TEST_TYPE" =~ ^(security|all)$ ]]; then
        report_content+="### 安全测试
- **状态**: ${SECURITY_TEST_STATUS:-未执行}
"
    fi
    
    report_content+="
## 详细报告

### 文件清单
"
    
    # 添加文件清单
    if [ -f "test-reports/coverage.html" ]; then
        report_content+="- [覆盖率报告](test-reports/coverage.html)
"
    fi
    if [ -f "test-reports/performance.txt" ]; then
        report_content+="- [性能测试报告](test-reports/performance.txt)
"
    fi
    if [ -f "test-reports/benchmark.txt" ]; then
        report_content+="- [基准测试报告](test-reports/benchmark.txt)
"
    fi
    if [ -f "test-reports/security.json" ]; then
        report_content+="- [安全扫描报告](test-reports/security.json)
"
    fi
    if [ -f "test-reports/vulnerabilities.json" ]; then
        report_content+="- [漏洞检查报告](test-reports/vulnerabilities.json)
"
    fi
    
    report_content+="
## 执行日志

$(cat test-execution.log 2>/dev/null || echo "无执行日志")

---
*报告生成时间: $(date)*
"
    
    echo "$report_content" > "$REPORT_FILE"
    log_success "测试报告已生成: $REPORT_FILE"
}

# 清理
cleanup() {
    if [ "$SKIP_CLEANUP" = true ]; then
        log_info "跳过清理步骤"
        return 0
    fi
    
    log_info "清理测试环境..."
    
    if [ "$DRY_RUN" = true ]; then
        echo "[DRY RUN] # 清理Docker容器"
        echo "[DRY RUN] docker-compose -f docker-compose.test.yml down -v"
        echo "[DRY RUN] # 清理临时文件"
        return 0
    fi
    
    # 清理Docker容器
    if [ -f "docker-compose.test.yml" ]; then
        docker-compose -f docker-compose.test.yml down -v 2>/dev/null || true
    fi
    
    # 清理临时文件
    rm -f /tmp/go-build*
    rm -f /tmp/go-runtime*
    
    log_success "清理完成"
}

# 信号处理
trap cleanup EXIT

# 创建执行日志
exec > >(tee -a test-execution.log)
exec 2>&1

# 主函数
main() {
    log_info "开始执行 $ENVIRONMENT 环境的 $TEST_TYPE 测试..."
    log_info "并行模式: $PARALLEL, 任务数: $JOBS"
    
    # 构建应用
    if ! build_application; then
        log_error "应用构建失败，停止测试执行"
        exit 1
    fi
    
    # 执行测试
    local test_failed=false
    
    # 单元测试
    if [[ "$TEST_TYPE" =~ ^(unit|all)$ ]]; then
        if run_unit_tests; then
            UNIT_TEST_STATUS="通过"
            # 提取覆盖率
            if [ -f "test-reports/unit_coverage.out" ]; then
                local coverage=$(go tool cover -func=test-reports/unit_coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
                UNIT_COVERAGE="${coverage:-0}"
            fi
        else
            UNIT_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 集成测试
    if [[ "$TEST_TYPE" =~ ^(integration|all)$ ]]; then
        if run_integration_tests; then
            INTEGRATION_TEST_STATUS="通过"
            # 提取覆盖率
            if [ -f "test-reports/integration_coverage.out" ]; then
                local coverage=$(go tool cover -func=test-reports/integration_coverage.out | tail -1 | awk '{print $3}' | sed 's/%//')
                INTEGRATION_COVERAGE="${coverage:-0}"
            fi
        else
            INTEGRATION_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 端到端测试
    if [[ "$TEST_TYPE" =~ ^(e2e|all)$ ]]; then
        if run_e2e_tests; then
            E2E_TEST_STATUS="通过"
        else
            E2E_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 性能测试
    if [[ "$TEST_TYPE" =~ ^(performance|all)$ ]]; then
        if run_performance_tests; then
            PERFORMANCE_TEST_STATUS="通过"
        else
            PERFORMANCE_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 基准测试
    if [ "$BENCH" = true ]; then
        if run_benchmark_tests; then
            BENCHMARK_TEST_STATUS="通过"
        else
            BENCHMARK_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 模糊测试
    if [ "$FUZZ" = true ]; then
        if run_fuzz_tests; then
            FUZZ_TEST_STATUS="通过"
        else
            FUZZ_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 安全测试
    if [[ "$TEST_TYPE" =~ ^(security|all)$ ]]; then
        if run_security_tests; then
            SECURITY_TEST_STATUS="通过"
        else
            SECURITY_TEST_STATUS="失败"
            test_failed=true
        fi
    fi
    
    # 生成覆盖率报告
    generate_coverage_report
    
    # 质量门禁检查
    if [ "$QUALITY_GATE" = true ]; then
        if ! quality_gate_check; then
            test_failed=true
        fi
    fi
    
    # 生成测试报告
    generate_test_report
    
    # 输出最终结果
    if [ "$test_failed" = true ]; then
        log_error "部分测试失败"
        if [ "$STRICT_MODE" = true ]; then
            log_error "严格模式：测试失败，退出流程"
            exit 1
        fi
        exit 1
    else
        log_success "所有测试通过"
        exit 0
    fi
}

# 运行主函数
main "$@"