#!/bin/bash
# 基于最新Docker最佳实践的自动化构建脚本
# 用于生产环境Docker镜像构建和部署

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# 默认配置
DEFAULT_REGISTRY="your-registry.com"
DEFAULT_PROJECT_NAME="law-oa"
DEFAULT_VERSION="latest"

# 解析命令行参数
REGISTRY=${REGISTRY:-$DEFAULT_REGISTRY}
PROJECT_NAME=${PROJECT_NAME:-$DEFAULT_PROJECT_NAME}
VERSION=${VERSION:-$DEFAULT_VERSION}
ENVIRONMENT=${ENVIRONMENT:-production}
PUSH_IMAGES=${PUSH_IMAGES:-false}
CLEANUP=${CLEANUP:-true}

# 显示帮助信息
show_help() {
    cat << EOF
Law OA Go - Docker构建脚本

用法: $0 [选项]

选项:
    -r, --registry REGISTRY     Docker镜像仓库地址 (默认: $DEFAULT_REGISTRY)
    -p, --project PROJECT       项目名称 (默认: $DEFAULT_PROJECT_NAME)
    -v, --version VERSION       版本标签 (默认: $DEFAULT_VERSION)
    -e, --environment ENV       环境类型 (development|staging|production, 默认: production)
    --push                      推送镜像到仓库
    --no-cleanup                构建后不清理临时文件
    -h, --help                  显示此帮助信息

示例:
    # 构建生产环境镜像
    $0 -e production -v v2.1.0

    # 构建并推送镜像
    $0 -v v2.1.0 --push

    # 使用自定义仓库
    $0 -r docker.io/myorg -p law-oa-prod -v v2.1.0

环境变量:
    REGISTRY                  Docker仓库地址
    PROJECT_NAME              项目名称
    VERSION                   版本标签
    ENVIRONMENT               环境类型
    PUSH_IMAGES               是否推送镜像 (true/false)
    CLEANUP                   是否清理临时文件 (true/false)

EOF
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -p|--project)
            PROJECT_NAME="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        --push)
            PUSH_IMAGES=true
            shift
            ;;
        --no-cleanup)
            CLEANUP=false
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 验证环境类型
case $ENVIRONMENT in
    development|staging|production)
        ;;
    *)
        log_error "无效的环境类型: $ENVIRONMENT"
        exit 1
        ;;
esac

# 构建信息
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

log_info "开始构建 Docker 镜像"
log_info "项目: $PROJECT_NAME"
log_info "版本: $VERSION"
log_info "环境: $ENVIRONMENT"
log_info "仓库: $REGISTRY"
log_info "提交: $BUILD_COMMIT"
log_info "分支: $BUILD_BRANCH"

# 检查Docker是否运行
if ! docker info >/dev/null 2>&1; then
    log_error "Docker未运行或无法访问"
    exit 1
fi

# 检查必要文件
required_files=("Dockerfile.optimized" ".dockerignore")
for file in "${required_files[@]}"; do
    if [[ ! -f "$PROJECT_ROOT/$file" ]]; then
        log_error "缺少必要文件: $file"
        exit 1
    fi
done

cd "$PROJECT_ROOT"

# 清理之前的构建
if [[ "$CLEANUP" == "true" ]]; then
    log_info "清理之前的构建缓存..."
    docker system prune -f --filter until=24h || true
fi

# 构建后端镜像
log_info "构建后端应用镜像..."

BACKEND_IMAGE_NAME="${REGISTRY}/${PROJECT_NAME}/backend:${VERSION}"
BACKEND_LATEST_NAME="${REGISTRY}/${PROJECT_NAME}/backend:latest"

# 构建参数
BUILD_ARGS=(
    "--build-arg" "BUILD_COMMIT=${BUILD_COMMIT}"
    "--build-arg" "BUILD_DATE=${BUILD_DATE}"
    "--build-arg" "BUILD_VERSION=${VERSION}"
    "--build-arg" "VERSION=${VERSION}"
    "--build-arg" "TARGETPLATFORM=linux/amd64"
    "--build-arg" "BUILDPLATFORM=linux/amd64"
)

# 多平台构建支持
if command -v docker buildx >/dev/null 2>&1; then
    log_info "使用 buildx 进行多平台构建..."

    # 创建构建器实例
    BUILDER_NAME="law-oa-builder"
    if ! docker buildx inspect "$BUILDER_NAME" >/dev/null 2>&1; then
        docker buildx create --name "$BUILDER_NAME" --driver docker-container --bootstrap
    fi

    docker buildx use "$BUILDER_NAME"

    # 构建镜像
    docker buildx build \
        --platform linux/amd64,linux/arm64 \
        -f Dockerfile.optimized \
        -t "$BACKEND_IMAGE_NAME" \
        -t "$BACKEND_LATEST_NAME" \
        "${BUILD_ARGS[@]}" \
        --push \
        .
else
    log_warning "未找到 buildx，使用标准构建..."

    docker build \
        -f Dockerfile.optimized \
        -t "$BACKEND_IMAGE_NAME" \
        -t "$BACKEND_LATEST_NAME" \
        "${BUILD_ARGS[@]}" \
        .

    if [[ "$PUSH_IMAGES" == "true" ]]; then
        log_info "推送后端镜像..."
        docker push "$BACKEND_IMAGE_NAME"
        docker push "$BACKEND_LATEST_NAME"
    fi
fi

# 构建前端镜像（如果存在）
if [[ -d "frontend" && -f "frontend/Dockerfile.prod" ]]; then
    log_info "构建前端应用镜像..."

    FRONTEND_IMAGE_NAME="${REGISTRY}/${PROJECT_NAME}/frontend:${VERSION}"
    FRONTEND_LATEST_NAME="${REGISTRY}/${PROJECT_NAME}/frontend:latest"

    cd frontend

    # 前端构建参数
    FRONTEND_BUILD_ARGS=(
        "--build-arg" "REACT_APP_VERSION=${VERSION}"
        "--build-arg" "REACT_APP_ENVIRONMENT=${ENVIRONMENT}"
        "--build-arg" "BUILD_DATE=${BUILD_DATE}"
        "--build-arg" "BUILD_COMMIT=${BUILD_COMMIT}"
    )

    if command -v docker buildx >/dev/null 2>&1; then
        docker buildx build \
            --platform linux/amd64,linux/arm64 \
            -f Dockerfile.prod \
            -t "$FRONTEND_IMAGE_NAME" \
            -t "$FRONTEND_LATEST_NAME" \
            "${FRONTEND_BUILD_ARGS[@]}" \
            --push \
            .
    else
        docker build \
            -f Dockerfile.prod \
            -t "$FRONTEND_IMAGE_NAME" \
            -t "$FRONTEND_LATEST_NAME" \
            "${FRONTEND_BUILD_ARGS[@]}" \
            .

        if [[ "$PUSH_IMAGES" == "true" ]]; then
            log_info "推送前端镜像..."
            docker push "$FRONTEND_IMAGE_NAME"
            docker push "$FRONTEND_LATEST_NAME"
        fi
    fi

    cd ..
fi

# 生成镜像清单
log_info "生成镜像清单..."
MANIFEST_FILE="build-manifest-${VERSION}.json"
cat > "$MANIFEST_FILE" << EOF
{
  "build_info": {
    "version": "${VERSION}",
    "environment": "${ENVIRONMENT}",
    "build_date": "${BUILD_DATE}",
    "git_commit": "${BUILD_COMMIT}",
    "git_branch": "${BUILD_BRANCH}"
  },
  "images": {
    "backend": {
      "image": "${BACKEND_IMAGE_NAME}",
      "latest": "${BACKEND_LATEST_NAME}"
    }
EOF

if [[ -d "frontend" ]]; then
    cat >> "$MANIFEST_FILE" << EOF
    ,
    "frontend": {
      "image": "${FRONTEND_IMAGE_NAME}",
      "latest": "${FRONTEND_LATEST_NAME}"
    }
EOF
fi

cat >> "$MANIFEST_FILE" << EOF
  },
  "deployment": {
    "compose_file": "docker-compose.prod.yml",
    "environment_file": ".env.${ENVIRONMENT}"
  }
}
EOF

# 验证镜像
log_info "验证构建的镜像..."
docker images --filter "reference=${REGISTRY}/${PROJECT_NAME}/*"

# 生成部署文件
log_info "生成部署配置..."

# 生成环境特定的docker-compose文件
COMPOSE_FILE="docker-compose.${ENVIRONMENT}.yml"
if [[ -f "docker-compose.prod.yml" ]]; then
    cp docker-compose.prod.yml "$COMPOSE_FILE"

    # 替换镜像标签
    if command -v sed >/dev/null 2>&1; then
        sed -i.bak "s|image: law-oa/backend:latest|image: ${BACKEND_IMAGE_NAME}|g" "$COMPOSE_FILE"
        sed -i.bak "s|image: law-oa/frontend:latest|image: ${FRONTEND_IMAGE_NAME}|g" "$COMPOSE_FILE"
        rm "${COMPOSE_FILE}.bak"
    fi
fi

# 生成.env文件模板
ENV_FILE=".env.${ENVIRONMENT}"
cat > "$ENV_FILE" << EOF
# Law OA Go - ${ENVIRONMENT^} Environment Configuration
# Generated on: ${BUILD_DATE}
# Version: ${VERSION}

# 应用配置
VERSION=${VERSION}
ENVIRONMENT=${ENVIRONMENT}
BUILD_DATE=${BUILD_DATE}
BUILD_COMMIT=${BUILD_COMMIT}

# 镜像配置
BACKEND_IMAGE=${BACKEND_IMAGE_NAME}
FRONTEND_IMAGE=${FRONTEND_IMAGE_NAME}

# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=lawuser
DB_PASSWORD_FILE=/run/secrets/mysql_user_password
DB_NAME=law_oa

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD_FILE=/run/secrets/redis_password
REDIS_DB=0

# Elasticsearch配置
ES_HOST=http://elasticsearch:9200

# JWT配置
JWT_SECRET_FILE=/run/secrets/jwt_secret

# 监控配置
PROMETHEUS_ENABLED=true
OPENTELEMETRY_ENABLED=true
OTLP_ENDPOINT=http://jaeger:4318

# 时区配置
TZ=Asia/Shanghai

# 日志级别
LOG_LEVEL=info

# 应用模式
GIN_MODE=release
EOF

# 清理
if [[ "$CLEANUP" == "true" ]]; then
    log_info "清理构建缓存..."
    docker builder prune -f >/dev/null 2>&1 || true
fi

# 生成部署命令
log_info "生成部署命令..."
cat << EOF

${GREEN}构建完成！${NC}

部署命令:
---------

1. 检查配置文件:
   cat $ENV_FILE

2. 启动服务:
   docker-compose -f $COMPOSE_FILE --env-file $ENV_FILE up -d

3. 检查服务状态:
   docker-compose -f $COMPOSE_FILE ps

4. 查看日志:
   docker-compose -f $COMPOSE_FILE logs -f

5. 停止服务:
   docker-compose -f $COMPOSE_FILE down

镜像信息:
---------
后端: $BACKEND_IMAGE_NAME
$(if [[ -d "frontend" ]]; then echo "前端: $FRONTEND_IMAGE_NAME"; fi)

构建信息:
---------
版本: $VERSION
环境: $ENVIRONMENT
提交: $BUILD_COMMIT
分支: $BUILD_BRANCH
日期: $BUILD_DATE

生成的文件:
-----------
- 构建清单: $MANIFEST_FILE
- 部署配置: $COMPOSE_FILE
- 环境变量: $ENV_FILE

EOF

log_success "Docker镜像构建完成！"