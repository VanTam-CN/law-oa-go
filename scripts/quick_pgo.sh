#!/bin/bash

# 简化的PGO构建脚本
# 快速执行PGO优化构建

set -e

# 配置
BINARY_NAME="bin/law-oa-go"
PROFILE_DIR="profiles"
DEFAULT_PGO_FILE="default.pgo"

# 颜色输出
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
    echo "  -p, --profile     生成剖析数据然后PGO构建"
    echo "  -t, --test        使用测试数据PGO构建"
    echo "  -b, --benchmark   使用基准测试数据PGO构建"
    echo "  -c, --clean       清理构建文件"
    echo "  -h, --help        显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 -p     # 生成剖析数据并PGO构建"
    echo "  $0 -t     # 使用测试数据PGO构建"
    echo "  $0 -c     # 清理构建文件"
}

# 清理构建
cleanup() {
    log_info "清理构建文件..."
    rm -rf build/
    rm -f "$BINARY_NAME"
    rm -rf "$PROFILE_DIR"
    find . -name "*.prof" -delete
    find . -name "*.test" -delete
    log_success "清理完成"
}

# 生成剖析数据
generate_profile() {
    log_info "生成剖析数据..."

    mkdir -p "$PROFILE_DIR"

    # 运行测试生成CPU剖析
    log_info "运行测试并生成CPU剖析..."
    go test -cpuprofile="$PROFILE_DIR/cpu.prof" ./...

    # 运行基准测试
    log_info "运行基准测试..."
    go test -bench=. -benchmem -cpuprofile="$PROFILE_DIR/bench.prof" ./...

    log_success "剖析数据生成完成"
}

# PGO构建
pgo_build() {
    local profile_arg=""

    if [ -f "$DEFAULT_PGO_FILE" ]; then
        profile_arg="-pgo=$DEFAULT_PGO_FILE"
        log_info "使用PGO配置文件: $DEFAULT_PGO_FILE"
    elif [ -f "$PROFILE_DIR/cpu.prof" ]; then
        profile_arg="-pgo=$PROFILE_DIR/cpu.prof"
        log_info "使用CPU剖析文件: $PROFILE_DIR/cpu.prof"
    elif [ -f "$PROFILE_DIR/bench.prof" ]; then
        profile_arg="-pgo=$PROFILE_DIR/bench.prof"
        log_info "使用基准测试剖析文件: $PROFILE_DIR/bench.prof"
    else
        log_warning "未找到剖析文件，进行标准构建"
    fi

    log_info "开始PGO构建..."
    mkdir -p "$(dirname "$BINARY_NAME")"

    # 构建参数
    local build_args=(
        -o "$BINARY_NAME"
        -ldflags="-s -w"
        -tags="netgo osusergo"
    )

    if [ -n "$profile_arg" ]; then
        build_args+=("$profile_arg")
    fi

    # 执行构建
    if go build "${build_args[@]}" ./main.go; then
        log_success "PGO构建成功"

        # 显示文件信息
        if [ -f "$BINARY_NAME" ]; then
            local size=$(du -h "$BINARY_NAME" | cut -f1)
            log_info "二进制文件大小: $size"
        fi
    else
        log_error "构建失败"
        return 1
    fi
}

# 主函数
main() {
    case "${1:-}" in
        -h|--help)
            show_usage
            exit 0
            ;;
        -c|--clean)
            cleanup
            exit 0
            ;;
        -p|--profile)
            generate_profile
            pgo_build
            ;;
        -t|--test)
            log_info "使用测试数据进行PGO构建..."
            generate_profile
            pgo_build
            ;;
        -b|--benchmark)
            log_info "使用基准测试数据进行PGO构建..."
            generate_profile
            pgo_build
            ;;
        "")
            log_warning "未指定选项，使用现有数据进行PGO构建..."
            pgo_build
            ;;
        *)
            log_error "未知选项: $1"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
