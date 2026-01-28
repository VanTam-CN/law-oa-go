#!/bin/bash

# 蓝绿部署自动化脚本
# 实现零停机部署，支持健康检查和自动回滚

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 显示使用方法
show_usage() {
    echo "用法: $0 [选项]"
    echo "选项:"
    echo "  -h, --help                 显示此帮助信息"
    echo "  -e, --env ENVIRONMENT      部署环境 (staging|production)"
    echo "  -i, --image IMAGE          Docker镜像名称"
    echo "  -t, --tag TAG              Docker镜像标签"
    echo "  -p, --port PORT            应用端口 (默认: 8080)"
    echo "  -w, --health-wait TIME     健康检查等待时间 (默认: 30s)"
    echo "  -r, --health-retries COUNT 健康检查重试次数 (默认: 10)"
    echo "  -d, --deployment-dir DIR   部署目录 (默认: /opt/law-oa-go)"
    echo "  -f, --force                强制部署（忽略健康检查）"
    echo "  -k, --keep-versions COUNT  保留版本数量 (默认: 3)"
    echo ""
    echo "示例:"
    echo "  $0 -e production -i law-oa-go-production -t v1.0.0"
    echo "  $0 -e staging -i law-oa-go-staging -t v1.0.0 -p 8081"
}

# 部署配置
DEPLOYMENT_CONFIG="/opt/law-oa-go/deployment.conf"
NGINX_CONFIG="/etc/nginx/sites-available/law-oa-go"

# 加载配置
load_config() {
    local environment=$1
    
    log_info "加载 $environment 环境配置..."
    
    if [ ! -f "$DEPLOYMENT_CONFIG" ]; then
        log_error "部署配置文件不存在: $DEPLOYMENT_CONFIG"
        return 1
    fi
    
    # 读取配置
    source "$DEPLOYMENT_CONFIG"
    
    # 验证环境配置
    if [ -z "${DEPLOYMENT_DIR[$environment]}" ]; then
        log_error "未找到环境配置: $environment"
        return 1
    fi
    
    DEPLOYMENT_PATH="${DEPLOYMENT_DIR[$environment]}"
    HEALTH_CHECK_URL="${HEALTH_CHECK_URLS[$environment]:-http://localhost:8080/health}"
    ROLLBACK_ENABLED="${ROLLBACK_ENABLED[$environment]:-true}"
    
    log_success "配置加载完成: $DEPLOYMENT_PATH"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        return 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装"
        return 1
    fi
    
    # 检查Nginx
    if ! command -v nginx &> /dev/null; then
        log_warning "Nginx 未安装，将跳过负载均衡配置"
    fi
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装"
        return 1
    fi
    
    log_success "依赖检查通过"
}

# 健康检查
perform_health_check() {
    local url=$1
    local max_wait=$2
    local max_retries=$3
    
    log_info "执行健康检查: $url"
    
    local attempt=1
    local total_wait=0
    
    while [ $attempt -le $max_retries ]; do
        if curl -f -s "$url" > /dev/null 2>&1; then
            log_success "健康检查通过"
            return 0
        fi
        
        if [ $total_wait -ge $max_wait ]; then
            log_error "健康检查超时"
            return 1
        fi
        
        log_info "健康检查失败，等待重试... ($attempt/$max_retries)"
        sleep 5
        ((attempt++))
        ((total_wait+=5))
    done
    
    log_error "健康检查失败"
    return 1
}

# 获取当前活跃环境
get_current_active() {
    local deployment_path=$1
    
    if [ -L "$deployment_path/current" ]; then
        readlink -f "$deployment_path/current"
    else
        echo ""
    fi
}

# 获取新版本目录
get_new_version_dir() {
    local deployment_path=$1
    local tag=$2
    
    echo "$deployment_path/versions/$tag"
}

# 创建版本目录
create_version_directory() {
    local version_dir=$1
    
    log_info "创建版本目录: $version_dir"
    
    mkdir -p "$version_dir"
    
    # 创建必要的子目录
    mkdir -p "$version_dir/logs"
    mkdir -p "$version_dir/config"
    mkdir -p "$version_dir/temp"
    
    log_success "版本目录创建完成"
}

# 部署新版本
deploy_new_version() {
    local image_name=$1
    local tag=$2
    local version_dir=$3
    local environment=$4
    
    log_info "部署新版本: $image_name:$tag"
    
    # 创建docker-compose.yml
    cat > "$version_dir/docker-compose.yml" << EOF
version: '3.8'
services:
  app:
    image: $image_name:$tag
    container_name: law-oa-go-$environment-${tag}
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - GIN_MODE=release
      - ENVIRONMENT=$environment
      - VERSION=$tag
    volumes:
      - ./logs:/app/logs
      - ./config:/app/config
      - ./temp:/app/temp
    networks:
      - law-oa-go-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M

networks:
  law-oa-go-network:
    driver: bridge
EOF
    
    # 启动新版本
    cd "$version_dir"
    docker-compose up -d
    
    log_success "新版本启动完成"
}

# 切换流量
switch_traffic() {
    local version_dir=$1
    local deployment_path=$2
    local environment=$3
    
    log_info "切换流量到新版本..."
    
    # 创建符号链接
    ln -sfn "$version_dir" "$deployment_path/current"
    
    # 重新加载Nginx配置
    if command -v nginx &> /dev/null; then
        nginx -t && nginx -s reload
        log_success "Nginx配置重新加载完成"
    fi
    
    log_success "流量切换完成"
}

# 停止旧版本
stop_old_version() {
    local deployment_path=$1
    local keep_versions=$2
    
    log_info "停止旧版本..."
    
    # 获取所有版本目录
    local versions_dir="$deployment_path/versions"
    if [ -d "$versions_dir" ]; then
        # 找到所有正在运行的容器
        cd "$versions_dir"
        for version_dir in */; do
            if [ "$version_dir" != "*" ] && [ "$version_dir" != "$(basename $deployment_path/current)" ]; then
                local version_path="$versions_dir/$version_dir"
                
                # 停止容器
                if [ -f "$version_path/docker-compose.yml" ]; then
                    cd "$version_path"
                    docker-compose down
                    log_info "已停止版本: $version_dir"
                fi
            fi
        done
    fi
    
    # 清理旧版本
    cleanup_old_versions "$deployment_path" "$keep_versions"
}

# 清理旧版本
cleanup_old_versions() {
    local deployment_path=$1
    local keep_versions=$2
    
    log_info "清理旧版本，保留最新的 $keep_versions 个版本..."
    
    local versions_dir="$deployment_path/versions"
    if [ -d "$versions_dir" ]; then
        cd "$versions_dir"
        
        # 获取所有版本并按时间排序
        local versions=($(ls -t */ 2>/dev/null | tr -d '/'))
        
        if [ ${#versions[@]} -gt $keep_versions ]; then
            # 删除多余的版本
            for ((i=$keep_versions; i<${#versions[@]}; i++)); do
                local version="${versions[$i]}"
                local version_path="$versions_dir/$version"
                
                # 删除版本目录
                rm -rf "$version_path"
                log_info "已删除版本: $version"
            done
        fi
    fi
    
    log_success "清理完成"
}

# 回滚部署
rollback_deployment() {
    local deployment_path=$1
    local target_version=$2
    
    log_info "回滚到版本: $target_version"
    
    local target_dir="$deployment_path/versions/$target_version"
    
    if [ ! -d "$target_dir" ]; then
        log_error "目标版本不存在: $target_dir"
        return 1
    fi
    
    # 切换流量
    switch_traffic "$target_dir" "$deployment_path" ""
    
    log_success "回滚完成"
}

# 生成部署报告
generate_deployment_report() {
    local environment=$1
    local tag=$2
    local status=$3
    local start_time=$4
    local end_time=$5
    
    local report_file="deployment_report_$(date +%Y%m%d_%H%M%S).json"
    
    log_info "生成部署报告: $report_file"
    
    cat > "$report_file" << EOF
{
  "deployment": {
    "environment": "$environment",
    "tag": "$tag",
    "status": "$status",
    "start_time": "$start_time",
    "end_time": "$end_time",
    "duration": "$(date -d "$end_time" +%s) - $(date -d "$start_time" +%s) seconds"
  },
  "system": {
    "hostname": "$(hostname)",
    "kernel": "$(uname -r)",
    "docker_version": "$(docker --version)",
    "compose_version": "$(docker-compose --version)"
  },
  "metrics": {
    "disk_usage": "$(df -h /)",
    "memory_usage": "$(free -h)",
    "docker_containers": "$(docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}')"
  }
}
EOF
    
    log_success "部署报告已生成: $report_file"
}

# 主函数
main() {
    # 默认参数
    ENVIRONMENT=""
    IMAGE_NAME=""
    TAG=""
    PORT="8080"
    HEALTH_WAIT="30"
    HEALTH_RETRIES="10"
    DEPLOYMENT_DIR="/opt/law-oa-go"
    FORCE=false
    KEEP_VERSIONS="3"
    
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
            -i|--image)
                IMAGE_NAME="$2"
                shift 2
                ;;
            -t|--tag)
                TAG="$2"
                shift 2
                ;;
            -p|--port)
                PORT="$2"
                shift 2
                ;;
            -w|--health-wait)
                HEALTH_WAIT="$2"
                shift 2
                ;;
            -r|--health-retries)
                HEALTH_RETRIES="$2"
                shift 2
                ;;
            -d|--deployment-dir)
                DEPLOYMENT_DIR="$2"
                shift 2
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -k|--keep-versions)
                KEEP_VERSIONS="$2"
                shift 2
                ;;
            *)
                log_error "未知参数: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # 验证必需参数
    if [ -z "$ENVIRONMENT" ]; then
        log_error "请指定部署环境"
        exit 1
    fi
    
    if [ -z "$IMAGE_NAME" ]; then
        log_error "请指定Docker镜像名称"
        exit 1
    fi
    
    if [ -z "$TAG" ]; then
        log_error "请指定Docker镜像标签"
        exit 1
    fi
    
    # 记录开始时间
    local start_time=$(date '+%Y-%m-%d %H:%M:%S')
    
    log_info "开始蓝绿部署: $ENVIRONMENT 环境, 镜像: $IMAGE_NAME:$TAG"
    
    # 检查依赖
    check_dependencies
    
    # 加载配置
    load_config "$ENVIRONMENT"
    
    # 获取路径
    local current_active=$(get_current_active "$DEPLOYMENT_PATH")
    local new_version_dir=$(get_new_version_dir "$DEPLOYMENT_PATH" "$TAG")
    
    # 创建版本目录
    create_version_directory "$new_version_dir"
    
    # 部署新版本
    deploy_new_version "$IMAGE_NAME" "$TAG" "$new_version_dir" "$ENVIRONMENT"
    
    # 健康检查
    local health_check_url="http://localhost:$PORT/health"
    if [ "$FORCE" = false ]; then
        if ! perform_health_check "$health_check_url" "$HEALTH_WAIT" "$HEALTH_RETRIES"; then
            log_error "新版本健康检查失败，开始回滚..."
            
            # 如果有当前活跃版本，则回滚
            if [ -n "$current_active" ]; then
                rollback_deployment "$DEPLOYMENT_PATH" "$(basename $current_active)"
            fi
            
            # 记录结束时间
            local end_time=$(date '+%Y-%m-%d %H:%M:%S')
            generate_deployment_report "$ENVIRONMENT" "$TAG" "failed" "$start_time" "$end_time"
            
            exit 1
        fi
    else
        log_warning "强制部署模式，跳过健康检查"
    fi
    
    # 切换流量
    switch_traffic "$new_version_dir" "$DEPLOYMENT_PATH" "$ENVIRONMENT"
    
    # 停止旧版本
    stop_old_version "$DEPLOYMENT_PATH" "$KEEP_VERSIONS"
    
    # 记录结束时间
    local end_time=$(date '+%Y-%m-%d %H:%M:%S')
    
    # 生成部署报告
    generate_deployment_report "$ENVIRONMENT" "$TAG" "success" "$start_time" "$end_time"
    
    log_success "蓝绿部署完成"
}

# 运行主函数
main "$@"