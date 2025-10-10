#!/bin/bash

# Docker镜像管理脚本
# 用于构建和管理law-oa项目的所有Docker镜像

set -e

# 配置变量
REGISTRY="law-oa"
VERSION="latest"
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

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

# 构建函数
build_image() {
    local image_name=$1
    local dockerfile_path=$2
    local build_context=$3
    local extra_args=$4

    local full_image_name="${REGISTRY}/${image_name}:${VERSION}"

    log_info "构建镜像: ${full_image_name}"

    docker build \
        --build-arg BUILD_COMMIT="${BUILD_COMMIT}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        -t "${full_image_name}" \
        -f "${dockerfile_path}" \
        ${extra_args} \
        "${build_context}"

    if [ $? -eq 0 ]; then
        log_success "镜像构建成功: ${full_image_name}"
    else
        log_error "镜像构建失败: ${full_image_name}"
        exit 1
    fi
}

# 拉取基础镜像
pull_base_images() {
    log_info "拉取基础镜像..."

    # MySQL基础镜像
    docker pull mysql:8.0
    docker tag mysql:8.0 ${REGISTRY}/mysql:8.0

    # Redis基础镜像
    docker pull redis:7-alpine
    docker tag redis:7-alpine ${REGISTRY}/redis:7-alpine

    # Elasticsearch基础镜像
    docker pull elasticsearch:8.8.0
    docker tag elasticsearch:8.8.0 ${REGISTRY}/elasticsearch:8.8.0

    # Kibana基础镜像
    docker pull kibana:8.8.0
    docker tag kibana:8.8.0 ${REGISTRY}/kibana:8.8.0

    log_success "基础镜像拉取和标记完成"
}

# 构建所有镜像
build_all() {
    log_info "开始构建所有law-oa镜像..."

    # 构建后端镜像
    build_image "backend" "Dockerfile" "." ""

    # 构建前端镜像
    build_image "frontend" "frontend/Dockerfile" "frontend" ""

    log_success "所有镜像构建完成"
}

# 推送镜像到仓库
push_images() {
    log_info "推送镜像到仓库..."

    images=("backend" "frontend" "mysql:8.0" "redis:7-alpine" "elasticsearch:8.8.0" "kibana:8.8.0")

    for image in "${images[@]}"; do
        local full_image_name="${REGISTRY}/${image}:${VERSION}"
        log_info "推送镜像: ${full_image_name}"
        docker push "${full_image_name}"
    done

    log_success "所有镜像推送完成"
}

# 清理未使用的镜像
cleanup_images() {
    log_info "清理未使用的镜像..."

    # 清理悬空镜像
    docker image prune -f

    # 清理law-oa相关的旧版本镜像（保留最新版本）
    docker images "${REGISTRY}/*" --format "{{.Repository}}:{{.Tag}}" | \
        grep -v "${VERSION}" | \
        xargs -r docker rmi

    log_success "镜像清理完成"
}

# 显示镜像列表
list_images() {
    log_info "law-oa镜像列表:"
    docker images "${REGISTRY}/*" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
}

# 显示帮助信息
show_help() {
    cat << EOF
Docker镜像管理脚本

用法: $0 [命令]

命令:
    build       构建所有项目镜像
    pull        拉取基础镜像并标记
    push        推送镜像到仓库
    cleanup     清理未使用的镜像
    list        显示镜像列表
    help        显示此帮助信息

示例:
    $0 build     # 构建所有镜像
    $0 pull      # 拉取基础镜像
    $0 push      # 推送镜像
    $0 cleanup   # 清理镜像
    $0 list      # 显示镜像列表

EOF
}

# 主函数
main() {
    case "${1:-}" in
        "build")
            build_all
            ;;
        "pull")
            pull_base_images
            ;;
        "push")
            push_images
            ;;
        "cleanup")
            cleanup_images
            ;;
        "list")
            list_images
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