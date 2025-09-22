#!/bin/bash

# Go 1.23 PGO (Profile-Guided Optimization) 构建脚本
# 此脚本用于执行基于性能剖析的优化构建流程

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目配置
PROJECT_NAME="law-oa-go"
BINARY_NAME="law-oa-server"
BUILD_DIR="build"
PROFILE_DIR="profiles"
DEFAULT_PGO_FILE="default.pgo"

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
Go 1.23 PGO 构建脚本

用法: $0 [选项]

选项:
    -h, --help          显示此帮助信息
    -p, --pgo           启用PGO优化构建
    -w, --workload      运行工作负载生成剖析数据
    -c, --clean         清理构建文件
    -d, --debug         构建调试版本
    -r, --race          启用竞态检测
    -s, --static        构建静态链接二进制文件
    -t, --test          运行测试并生成剖析数据
    -b, --benchmark     运行基准测试
    -o, --output DIR    指定输出目录 (默认: $BUILD_DIR)
    -v, --version VERSION 指定版本信息
    -j, --jobs NUM      并行构建作业数 (默认: CPU核心数)

示例:
    $0 -p               # 使用PGO优化构建
    $0 -w -p            # 生成工作负载剖析数据后PGO构建
    $0 -t -p            # 运行测试生成剖析数据后PGO构建
    $0 -c -d            # 清理并构建调试版本

EOF
}

# 清理函数
cleanup() {
    log_info "清理构建文件..."
    rm -rf "$BUILD_DIR"
    rm -f "$BINARY_NAME"
    rm -rf "$PROFILE_DIR"
    find . -name "*.prof" -delete
    find . -name "*.test" -delete
    log_success "清理完成"
}

# 运行工作负载生成剖析数据
run_workload() {
    log_info "运行工作负载生成剖析数据..."
    
    # 确保服务器在运行
    if ! pgrep -f "$BINARY_NAME" > /dev/null; then
        log_warning "服务器未运行，启动服务器进行剖析..."
        
        # 先构建基本版本
        go build -o "$BINARY_NAME" ./main.go
        
        # 启动服务器
        ./"$BINARY_NAME" &
        SERVER_PID=$!
        
        # 等待服务器启动
        sleep 5
        
        # 设置清理函数
        trap "kill $SERVER_PID 2>/dev/null || true" EXIT
    fi
    
    # 创建剖析目录
    mkdir -p "$PROFILE_DIR"
    
    # 运行HTTP工作负载
    log_info "运行HTTP工作负载..."
    go run scripts/pgo_workload.go &
    HTTP_WORKLOAD_PID=$!
    
    # 运行综合工作负载
    log_info "运行综合工作负载..."
    go run scripts/comprehensive_pgo_workload.go &
    COMPREHENSIVE_WORKLOAD_PID=$!
    
    # 等待工作负载完成
    wait $HTTP_WORKLOAD_PID
    wait $COMPREHENSIVE_WORKLOAD_PID
    
    log_success "工作负载完成"
}

# 运行测试生成剖析数据
run_tests_with_profiling() {
    log_info "运行测试生成剖析数据..."
    
    # 创建剖析目录
    mkdir -p "$PROFILE_DIR"
    
    # 运行测试并生成CPU剖析
    log_info "运行测试并生成CPU剖析..."
    go test -cpuprofile="$PROFILE_DIR/cpu.prof" -coverprofile="$PROFILE_DIR/coverage.out" ./...
    
    # 运行测试并生成内存剖析
    log_info "运行测试并生成内存剖析..."
    go test -memprofile="$PROFILE_DIR/memory.prof" ./...
    
    # 运行基准测试
    log_info "运行基准测试..."
    go test -bench=. -benchmem -cpuprofile="$PROFILE_DIR/bench_cpu.prof" ./...
    
    log_success "测试剖析完成"
}

# 处理剖析数据
process_profiles() {
    log_info "处理剖析数据..."
    
    # 检查是否有剖析文件
    if [ ! -d "$PROFILE_DIR" ] || [ -z "$(ls -A "$PROFILE_DIR"/*.prof 2>/dev/null)" ]; then
        log_error "没有找到剖析文件"
        return 1
    fi
    
    # 生成默认的PGO配置（如果不存在）
    if [ ! -f "$DEFAULT_PGO_FILE" ]; then
        log_warning "未找到默认PGO配置文件，使用剖析数据生成..."
        
        # 合并所有CPU剖析文件
        if [ -f "$PROFILE_DIR/cpu.prof" ]; then
            go tool pprof -output="$PROFILE_DIR/merged_cpu.prof" -text "$PROFILE_DIR/cpu.prof"
        fi
        
        if [ -f "$PROFILE_DIR/bench_cpu.prof" ]; then
            go tool pprof -output="$PROFILE_DIR/merged_bench_cpu.prof" -text "$PROFILE_DIR/bench_cpu.prof"
        fi
        
        log_info "剖析数据处理完成"
    fi
}

# 构建函数
build_binary() {
    local build_args=()
    local build_mode="standard"
    
    # 添加构建标签
    build_args+=(-tags="netgo osusergo")
    
    # 处理构建参数
    if [ "$ENABLE_PGO" = true ]; then
        log_info "启用PGO优化构建..."
        
        # 检查PGO文件是否存在
        if [ -f "$DEFAULT_PGO_FILE" ]; then
            build_args+=(-pgo="$DEFAULT_PGO_FILE")
            log_info "使用PGO配置文件: $DEFAULT_PGO_FILE"
        else
            log_warning "未找到PGO配置文件，尝试使用剖析数据..."
            
            if [ -d "$PROFILE_DIR" ] && [ -n "$(ls -A "$PROFILE_DIR"/*.prof 2>/dev/null)" ]; then
                # 使用剖析数据文件
                for prof_file in "$PROFILE_DIR"/*.prof; do
                    if [ -f "$prof_file" ]; then
                        build_args+=(-pgo="$prof_file")
                        log_info "使用剖析文件: $prof_file"
                        break
                    fi
                done
            else
                log_warning "没有找到剖析数据，进行标准构建"
            fi
        fi
        
        build_mode="pgo"
    fi
    
    # 调试构建
    if [ "$DEBUG_BUILD" = true ]; then
        build_args+=(-gcflags="all=-N -l")
        build_mode="debug"
    fi
    
    # 竞态检测
    if [ "$RACE_DETECTION" = true ]; then
        build_args+=(-race)
    fi
    
    # 静态链接
    if [ "$STATIC_LINK" = true ]; then
        build_args+=(-ldflags="-extldflags=-static")
        build_mode="static"
    fi
    
    # 版本信息
    if [ -n "$VERSION" ]; then
        build_args+=(-ldflags="-X main.Version=$VERSION")
    fi
    
    # 输出目录
    if [ -n "$OUTPUT_DIR" ]; then
        mkdir -p "$OUTPUT_DIR"
        build_args+=(-o="$OUTPUT_DIR/$BINARY_NAME")
    else
        mkdir -p "$BUILD_DIR"
        build_args+=(-o="$BUILD_DIR/$BINARY_NAME")
    fi
    
    # 并行构建
    if [ -n "$JOBS" ]; then
        build_args+=(-p="$JOBS")
    fi
    
    # 构建信息
    log_info "构建模式: $build_mode"
    log_info "构建参数: ${build_args[*]}"
    
    # 执行构建
    log_info "开始构建..."
    if go build "${build_args[@]}" ./main.go; then
        log_success "构建成功"
        
        # 显示构建信息
        local binary_path="${OUTPUT_DIR:-$BUILD_DIR}/$BINARY_NAME"
        if [ -f "$binary_path" ]; then
            local file_size=$(du -h "$binary_path" | cut -f1)
            log_info "二进制文件大小: $file_size"
            
            # 显示构建时间
            local build_time=$(date -r "$binary_path" "+%Y-%m-%d %H:%M:%S")
            log_info "构建时间: $build_time"
        fi
    else
        log_error "构建失败"
        return 1
    fi
}

# 运行基准测试
run_benchmarks() {
    log_info "运行基准测试..."
    
    mkdir -p "$PROFILE_DIR"
    
    # 运行所有基准测试
    go test -bench=. -benchmem -count=3 -benchtime=1s \
        -cpuprofile="$PROFILE_DIR/benchmark_cpu.prof" \
        -memprofile="$PROFILE_DIR/benchmark_mem.prof" \
        ./...
    
    log_success "基准测试完成"
}

# 验证构建
verify_build() {
    local binary_path="${OUTPUT_DIR:-$BUILD_DIR}/$BINARY_NAME"
    
    if [ ! -f "$binary_path" ]; then
        log_error "二进制文件不存在: $binary_path"
        return 1
    fi
    
    # 检查二进制文件信息
    log_info "验证二进制文件..."
    
    # 显示文件信息
    file "$binary_path"
    
    # 显示依赖信息
    if command -v ldd >/dev/null 2>&1; then
        ldd "$binary_path" | head -5
    fi
    
    # 尝试运行版本检查
    if ./"$binary_path" --version >/dev/null 2>&1; then
        log_info "版本检查:"
        ./"$binary_path" --version
    fi
    
    log_success "构建验证完成"
}

# 生成构建报告
generate_report() {
    local report_file="${BUILD_DIR}/build_report.txt"
    
    log_info "生成构建报告..."
    
    {
        echo "=== PGO 构建报告 ==="
        echo "项目: $PROJECT_NAME"
        echo "二进制文件: $BINARY_NAME"
        echo "构建时间: $(date)"
        echo "Go版本: $(go version)"
        echo "构建模式: ${BUILD_MODE:-standard}"
        echo ""
        
        if [ -f "$DEFAULT_PGO_FILE" ]; then
            echo "PGO配置: 使用 ($DEFAULT_PGO_FILE)"
        else
            echo "PGO配置: 未使用"
        fi
        
        echo ""
        echo "=== 构建参数 ==="
        echo "输出目录: ${OUTPUT_DIR:-$BUILD_DIR}"
        echo "调试构建: ${DEBUG_BUILD:-false}"
        echo "竞态检测: ${RACE_DETECTION:-false}"
        echo "静态链接: ${STATIC_LINK:-false}"
        echo "并行作业: ${JOBS:-auto}"
        echo ""
        
        if [ -f "${OUTPUT_DIR:-$BUILD_DIR}/$BINARY_NAME" ]; then
            echo "=== 二进制文件信息 ==="
            local binary_path="${OUTPUT_DIR:-$BUILD_DIR}/$BINARY_NAME"
            echo "文件大小: $(du -h "$binary_path" | cut -f1)"
            echo "文件权限: $(ls -l "$binary_path" | awk '{print $1}')"
            echo "修改时间: $(date -r "$binary_path" "+%Y-%m-%d %H:%M:%S")"
            echo ""
            
            echo "=== 依赖信息 ==="
            if command -v ldd >/dev/null 2>&1; then
                ldd "$binary_path" | head -5
            else
                echo "ldd 命令不可用"
            fi
        fi
        
        if [ -d "$PROFILE_DIR" ] && [ -n "$(ls -A "$PROFILE_DIR" 2>/dev/null)" ]; then
            echo ""
            echo "=== 剖析数据 ==="
            ls -la "$PROFILE_DIR"/*.prof 2>/dev/null || echo "无剖析文件"
        fi
        
    } > "$report_file"
    
    log_success "构建报告已生成: $report_file"
}

# 主函数
main() {
    # 初始化变量
    ENABLE_PGO=false
    RUN_WORKLOAD=false
    CLEAN_BUILD=false
    DEBUG_BUILD=false
    RACE_DETECTION=false
    STATIC_LINK=false
    RUN_TESTS=false
    RUN_BENCHMARK=false
    OUTPUT_DIR=""
    VERSION=""
    JOBS=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            -p|--pgo)
                ENABLE_PGO=true
                shift
                ;;
            -w|--workload)
                RUN_WORKLOAD=true
                shift
                ;;
            -c|--clean)
                CLEAN_BUILD=true
                shift
                ;;
            -d|--debug)
                DEBUG_BUILD=true
                shift
                ;;
            -r|--race)
                RACE_DETECTION=true
                shift
                ;;
            -s|--static)
                STATIC_LINK=true
                shift
                ;;
            -t|--test)
                RUN_TESTS=true
                shift
                ;;
            -b|--benchmark)
                RUN_BENCHMARK=true
                shift
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -j|--jobs)
                JOBS="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 设置默认并行作业数
    if [ -z "$JOBS" ]; then
        JOBS=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4)
    fi
    
    # 设置构建模式
    if [ "$ENABLE_PGO" = true ]; then
        BUILD_MODE="pgo"
    elif [ "$DEBUG_BUILD" = true ]; then
        BUILD_MODE="debug"
    elif [ "$STATIC_LINK" = true ]; then
        BUILD_MODE="static"
    else
        BUILD_MODE="standard"
    fi
    
    log_info "开始 $PROJECT_NAME PGO 构建流程"
    log_info "Go版本: $(go version)"
    log_info "构建模式: $BUILD_MODE"
    
    # 清理构建
    if [ "$CLEAN_BUILD" = true ]; then
        cleanup
    fi
    
    # 生成剖析数据
    if [ "$RUN_WORKLOAD" = true ]; then
        run_workload
    fi
    
    if [ "$RUN_TESTS" = true ]; then
        run_tests_with_profiling
    fi
    
    if [ "$RUN_BENCHMARK" = true ]; then
        run_benchmarks
    fi
    
    # 处理剖析数据
    if [ "$ENABLE_PGO" = true ] && ([ "$RUN_WORKLOAD" = true ] || [ "$RUN_TESTS" = true ]); then
        process_profiles
    fi
    
    # 构建二进制文件
    build_binary
    
    # 验证构建
    verify_build
    
    # 生成报告
    generate_report
    
    log_success "PGO 构建流程完成"
}

# 运行主函数
main "$@"