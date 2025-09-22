#!/bin/bash

# Law OA Go 项目 CI/CD 部署脚本
# 支持多种部署环境和配置

set -e

# 颜色输出
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

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -e, --env ENVIRONMENT      部署环境 (dev|staging|production)"
    echo "  -b, --build-target TARGET  构建目标 (standard|pgo)"
    echo "  -t, --tag TAG              Docker镜像标签"
    echo "  -d, --deploy                执行部署"
    echo "  -r, --rollback VERSION     回滚到指定版本"
    echo "  -c, --clean                清理旧版本"
    echo "  -s, --strict               严格模式（质量门禁失败则停止部署）"
    echo "  --skip-quality-gate        跳过质量门禁检查"
    echo "  --skip-fuzzing            跳过Fuzzing测试"
    echo "  --skip-security           跳过安全扫描"
    echo ""
    echo "质量门禁标准:"
    echo "  - 开发环境: 测试覆盖率 > 60%, Fuzzing crashes < 10"
    echo "  - 测试环境: 测试覆盖率 > 75%, Fuzzing crashes < 5"
    echo "  - 生产环境: 测试覆盖率 > 85%, Fuzzing crashes < 2, PGO性能提升 > 5%"
    echo ""
    echo "示例:"
    echo "  $0 -e staging -b pgo -t v1.0.0 -d           # 部署PGO版本到测试环境"
    echo "  $0 -e production -b pgo -t v1.0.0 -d -s    # 生产环境严格模式部署"
    echo "  $0 -e dev -b standard -d                   # 开发环境标准构建部署"
    echo "  $0 -e staging -b pgo -t v1.0.0 -d --skip-fuzzing  # 跳过Fuzzing测试"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        exit 1
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go 未安装"
        exit 1
    fi
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        log_error "Git 未安装"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 构建应用
build_application() {
    local build_target=$1
    local tag=$2
    local environment=$3
    
    log_info "构建应用 (目标: $build_target, 标签: $tag, 环境: $environment)..."
    
    # 设置构建参数
    export BUILD_COMMIT=$(git rev-parse --short HEAD)
    export BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    export BUILD_TARGET=$build_target
    export PGO_PROFILE=default.pgo
    
    # 如果是PGO构建，先生成性能数据
    if [ "$build_target" = "pgo" ]; then
        log_info "生成PGO性能数据..."
        ./scripts/quick_pgo.sh -p
    fi
    
    # 构建Docker镜像
    local image_name="law-oa-go"
    if [ "$environment" != "dev" ]; then
        image_name="law-oa-go-$environment"
    fi
    
    docker build \
        --build-arg BUILD_COMMIT=$BUILD_COMMIT \
        --build-arg BUILD_DATE=$BUILD_DATE \
        --build-arg BUILD_TARGET=$BUILD_TARGET \
        --build-arg PGO_PROFILE=$PGO_PROFILE \
        -t $image_name:$tag \
        -t $image_name:latest \
        .
    
    log_success "应用构建完成: $image_name:$tag"
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    # 单元测试
    log_info "运行单元测试..."
    go test -v -race -coverprofile=coverage.out ./...
    
    # 集成测试
    log_info "运行集成测试..."
    go test -v -race -tags=integration ./...
    
    # Fuzzing测试（快速版本）
    log_info "运行Fuzzing测试..."
    chmod +x scripts/fuzzing_test.sh
    ./scripts/fuzzing_test.sh -a -t 30s -f 2
    
    log_success "所有测试通过"
}

# 部署到环境
deploy_to_environment() {
    local environment=$1
    local tag=$2
    
    log_info "部署到 $environment 环境..."
    
    # 根据环境选择配置
    case $environment in
        dev)
            log_info "部署到开发环境..."
            docker-compose down
            docker-compose up -d
            ;;
        staging)
            log_info "部署到测试环境..."
            # 使用测试环境配置
            export GIN_MODE=release
            export ENVIRONMENT=staging
            docker-compose -f docker-compose.yml -f docker-compose.staging.yml down
            docker-compose -f docker-compose.yml -f docker-compose.staging.yml up -d
            ;;
        production)
            log_info "部署到生产环境..."
            # 生产环境使用蓝绿部署策略
            export GIN_MODE=release
            export ENVIRONMENT=production
            
            # 蓝绿部署
            perform_blue_green_deployment $tag $environment
            ;;
        *)
            log_error "未知环境: $environment"
            exit 1
            ;;
    esac
    
    log_success "部署到 $environment 环境完成"
}

# 蓝绿部署
perform_blue_green_deployment() {
    local tag=$1
    local environment=$2
    
    log_info "执行蓝绿部署..."
    
    # 确保蓝绿部署脚本存在
    if [ ! -f "scripts/blue_green_deploy.sh" ]; then
        log_error "蓝绿部署脚本不存在: scripts/blue_green_deploy.sh"
        return 1
    fi
    
    # 给脚本执行权限
    chmod +x scripts/blue_green_deploy.sh
    
    # 构建镜像名称
    local image_name="law-oa-go"
    if [ "$environment" != "dev" ]; then
        image_name="law-oa-go-$environment"
    fi
    
    # 执行蓝绿部署
    local deploy_args="-e $environment -i $image_name -t $tag"
    
    if ! ./scripts/blue_green_deploy.sh $deploy_args; then
        log_error "蓝绿部署失败"
        return 1
    fi
    
    log_success "蓝绿部署完成"
}

# 健康检查
health_check() {
    local environment=$1
    local max_attempts=30
    local attempt=1
    
    log_info "执行健康检查..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_success "健康检查通过"
            return 0
        fi
        
        log_info "健康检查失败，等待重试... ($attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "健康检查失败"
    return 1
}

# 回滚部署
rollback_deployment() {
    local version=$1
    local environment=$2
    
    log_info "回滚到版本 $version..."
    
    # 停止当前版本
    docker-compose down
    
    # 回滚到指定版本
    docker pull law-oa-go:$version
    docker-compose up -d
    
    # 健康检查
    if health_check $environment; then
        log_success "回滚成功"
    else
        log_error "回滚失败"
        exit 1
    fi
}

# 清理旧版本
cleanup_old_versions() {
    local keep_versions=3
    
    log_info "清理旧版本，保留最新的 $keep_versions 个版本..."
    
    # 清理Docker镜像
    docker images | grep law-oa-go | grep -v latest | awk '{print $1":"$2}' | sort -V | head -n -$keep_versions | xargs -r docker rmi
    
    # 清理Docker缓存
    docker system prune -f
    
    log_success "清理完成"
}

# 质量门禁检查
quality_gate_check() {
    local environment=$1
    local strict_mode=$2
    
    log_info "执行部署质量门禁检查..."
    
    # 确保质量门禁脚本存在
    if [ ! -f "scripts/deployment_quality_gate.sh" ]; then
        log_error "质量门禁脚本不存在: scripts/deployment_quality_gate.sh"
        return 1
    fi
    
    # 给脚本执行权限
    chmod +x scripts/deployment_quality_gate.sh
    
    # 执行质量门禁检查
    local gate_args="-e $environment"
    if [ "$strict_mode" = true ]; then
        gate_args="$gate_args -s"
    fi
    
    # 生成质量报告
    local report_file="quality_gate_report_$(date +%Y%m%d_%H%M%S).md"
    gate_args="$gate_args -r $report_file"
    
    if ! ./scripts/deployment_quality_gate.sh $gate_args; then
        log_error "质量门禁检查失败"
        log_info "质量报告已生成: $report_file"
        return 1
    fi
    
    log_success "质量门禁检查通过"
    log_info "质量报告已生成: $report_file"
}

# 执行Fuzzing测试
run_fuzzing_tests() {
    local environment=$1
    
    log_info "执行Fuzzing测试..."
    
    # 根据环境选择Fuzzing测试策略
    case $environment in
        dev)
            # 开发环境：快速Fuzzing
            ./scripts/fuzzing_test.sh -a -t 30s -f 2
            ;;
        staging)
            # 测试环境：中等Fuzzing
            ./scripts/fuzzing_test.sh -a -t 60s -f 4
            ;;
        production)
            # 生产环境：全面Fuzzing
            ./scripts/fuzzing_test.sh -a -t 120s -f 8
            ;;
    esac
    
    log_success "Fuzzing测试完成"
}

# 执行安全扫描
run_security_scan() {
    local environment=$1
    
    log_info "执行安全扫描..."
    
    # 运行静态安全分析
    if command -v gosec &> /dev/null; then
        log_info "运行gosec安全扫描..."
        gosec -fmt=json -out=security_scan_report.json ./...
    fi
    
    # 运行依赖安全检查
    if command -v govulncheck &> /dev/null; then
        log_info "运行govulncheck漏洞检查..."
        govulncheck -json ./... > vulnerability_report.json 2>/dev/null || true
    fi
    
    log_success "安全扫描完成"
}

# 主函数
main() {
    # 默认参数
    ENVIRONMENT="dev"
    BUILD_TARGET="standard"
    TAG="latest"
    DEPLOY=false
    ROLLBACK_VERSION=""
    CLEAN=false
    STRICT_MODE=false
    SKIP_QUALITY_GATE=false
    SKIP_FUZZING=false
    SKIP_SECURITY=false
    
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
            -b|--build-target)
                BUILD_TARGET="$2"
                shift 2
                ;;
            -t|--tag)
                TAG="$2"
                shift 2
                ;;
            -d|--deploy)
                DEPLOY=true
                shift
                ;;
            -r|--rollback)
                ROLLBACK_VERSION="$2"
                shift 2
                ;;
            -c|--clean)
                CLEAN=true
                shift
                ;;
            -s|--strict)
                STRICT_MODE=true
                shift
                ;;
            --skip-quality-gate)
                SKIP_QUALITY_GATE=true
                shift
                ;;
            --skip-fuzzing)
                SKIP_FUZZING=true
                shift
                ;;
            --skip-security)
                SKIP_SECURITY=true
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 检查依赖
    check_dependencies
    
    # 回滚模式
    if [ -n "$ROLLBACK_VERSION" ]; then
        rollback_deployment $ROLLBACK_VERSION $ENVIRONMENT
        exit 0
    fi
    
    # 构建应用
    build_application $BUILD_TARGET $TAG $ENVIRONMENT
    
    # 运行基础测试
    run_tests
    
    # 运行Fuzzing测试
    if [ "$SKIP_FUZZING" = false ]; then
        run_fuzzing_tests $ENVIRONMENT
    else
        log_warning "跳过Fuzzing测试"
    fi
    
    # 运行安全扫描
    if [ "$SKIP_SECURITY" = false ]; then
        run_security_scan $ENVIRONMENT
    else
        log_warning "跳过安全扫描"
    fi
    
    # 质量门禁检查
    if [ "$SKIP_QUALITY_GATE" = false ]; then
        if ! quality_gate_check $ENVIRONMENT $STRICT_MODE; then
            if [ "$STRICT_MODE" = true ]; then
                log_error "严格模式下质量门禁失败，停止部署"
                exit 1
            else
                log_warning "质量门禁检查失败，非严格模式下继续部署"
            fi
        fi
    else
        log_warning "跳过质量门禁检查"
    fi
    
    # 部署
    if [ "$DEPLOY" = true ]; then
        deploy_to_environment $ENVIRONMENT $TAG
        
        # 健康检查
        if ! health_check $ENVIRONMENT; then
            log_error "部署后健康检查失败"
            exit 1
        fi
        
        # 启动部署后监控
        if [ "$ENVIRONMENT" = "production" ] || [ "$ENVIRONMENT" = "staging" ]; then
            log_info "启动部署后监控..."
            
            # 确保监控脚本存在
            if [ -f "scripts/deployment_monitor.sh" ]; then
                chmod +x scripts/deployment_monitor.sh
                
                # 后台启动监控（监控10分钟）
                nohup ./scripts/deployment_monitor.sh -e $ENVIRONMENT -d 10 -o deployment_monitoring > monitoring.log 2>&1 &
                MONITOR_PID=$!
                
                log_success "部署监控已启动 (PID: $MONITOR_PID)"
                
                # 等待监控收集初始数据
                sleep 30
            else
                log_warning "监控脚本不存在，跳过部署后监控"
            fi
        fi
    fi
    
    # 清理
    if [ "$CLEAN" = true ]; then
        cleanup_old_versions
    fi
    
    log_success "CI/CD 流程完成"
}

# 运行主函数
main "$@"