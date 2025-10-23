#!/bin/bash
# Law OA Go Kubernetes 部署脚本
# 基于最新Kubernetes部署最佳实践

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
K8S_DIR="$PROJECT_ROOT/k8s"

# 环境配置
NAMESPACE=${NAMESPACE:-law-oa}
CONTEXT=${CONTEXT:-}
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}

# 构建配置
BUILD_IMAGES=${BUILD_IMAGES:-true}
PUSH_IMAGES=${PUSH_IMAGES:-false}
REGISTRY=${REGISTRY:-your-registry.com}
IMAGE_TAG=${IMAGE_TAG:-latest}

# 部署配置
SKIP_NAMESPACES=${SKIP_NAMESPACES:-false}
SKIP_SECRETS=${SKIP_SECRETS:-false}
SKIP_CONFIGMAPS=${SKIP_CONFIGMAPS:-false}
SKIP_DEPLOYMENTS=${SKIP_DEPLOYMENTS:-false}
SKIP_SERVICES=${SKIP_SERVICES:-false}

# 验证配置
VERIFY=${VERIFY:-true}
WAIT_FOR_ROLLOUT=${WAIT_FOR_ROLLOUT:-true}
HEALTH_CHECK_TIMEOUT=${HEALTH_CHECK_TIMEOUT:-300}

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

log_step() {
    echo -e "${PURPLE}[STEP]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
Law OA Go Kubernetes 部署脚本

用法: $0 [选项] [命令]

命令:
    deploy          部署应用到Kubernetes集群 (默认)
    delete          删除应用部署
    status          查看应用状态
    logs             查看应用日志
    rollout         执行滚动更新
    scale           扩缩应用
    health          健康检查
    validate        验证部署配置
    clean            清理资源

选项:
    -n, --namespace NAMESPACE  指定命名空间 (默认: law-oa)
    -c, --context CONTEXT      指定kubectl上下文
    -k, --kubeconfig KUBECONFIG  指定kubeconfig文件
    -t, --tag TAG              指定镜像标签 (默认: latest)
    -r, --registry REGISTRY     指定镜像仓库
    -b, --build                 构建镜像
    -p, --push                  推送镜像
    --skip-namespaces           跳过命名空间创建
    --skip-secrets              跳过密钥创建
    --skip-configmaps           跳过配置映射创建
    --skip-deployments          跳过部署创建
    --skip-services             跳过服务创建
    --no-verify                 跳过部署验证
    --no-wait                   跳过等待部署完成
    --health-timeout TIMEOUT    健康检查超时时间 (默认: 300秒)
    -v, --verbose               详细输出
    -h, --help                  显示此帮助信息

示例:
    $0 deploy                   # 部署应用
    $0 delete -n law-oa          # 删除命名空间law-oa中的应用
    $0 rollout -t v2.1.0        # 执行滚动更新到v2.1.0版本
    $0 scale backend 5          # 扩展backend到5个副本
    $0 health                   # 检查应用健康状态

环境变量:
    NAMESPACE                    命名空间
    CONTEXT                      kubectl上下文
    KUBECONFIG                   kubeconfig文件路径
    BUILD_IMAGES                 是否构建镜像
    PUSH_IMAGES                  是否推送镜像
    REGISTRY                     镜像仓库地址
    IMAGE_TAG                    镜像标签
    SKIP_NAMESPACES              是否跳过命名空间
    SKIP_SECRETS                  是否跳过密钥
    SKIP_CONFIGMAPS              是否跳过配置映射
    SKIP_DEPLOYMENTS             是否跳过部署
    SKIP_SERVICES                是否跳过服务
    VERIFY                       是否验证部署
    WAIT_FOR_ROLLOUT             是否等待部署完成
    HEALTH_CHECK_TIMEOUT         健康检查超时时间

EOF
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖工具..."

    # 检查kubectl
    if ! command -v kubectl >/dev/null 2>&1; then
        log_error "kubectl未安装或不在PATH中"
        exit 1
    fi

    # 检查Docker (如果需要构建镜像)
    if [[ "$BUILD_IMAGES" == "true" ]] && ! command -v docker >/dev/null 2>&1; then
        log_error "Docker未安装或不在PATH中"
        exit 1
    fi

    # 检查kubeconfig
    if [[ ! -f "$KUBECONFIG" ]]; then
        log_error "kubeconfig文件不存在: $KUBECONFIG"
        exit 1
    fi

    # 设置kubectl配置
    export KUBECONFIG="$KUBECONFIG"

    # 检查集群连接
    if ! kubectl cluster-info >/dev/null 2>&1; then
        log_error "无法连接到Kubernetes集群"
        exit 1
    fi

    # 设置上下文
    if [[ -n "$CONTEXT" ]]; then
        if ! kubectl config use-context "$CONTEXT" >/dev/null 2>&1; then
            log_error "无法设置kubectl上下文: $CONTEXT"
            exit 1
        fi
    fi

    log_success "依赖检查通过"
}

# 构建镜像
build_images() {
    if [[ "$BUILD_IMAGES" != "true" ]]; then
        log_info "跳过镜像构建"
        return
    fi

    log_step "构建Docker镜像..."

    # 构建后端镜像
    log_info "构建后端镜像..."
    if ! docker build -t "$REGISTRY/law-oa-go:$IMAGE_TAG" "$PROJECT_ROOT"; then
        log_error "后端镜像构建失败"
        exit 1
    fi

    # 构建前端镜像 (如果存在)
    if [[ -d "$PROJECT_ROOT/frontend" ]]; then
        log_info "构建前端镜像..."
        if ! docker build -t "$REGISTRY/law-oa-frontend:$IMAGE_TAG" "$PROJECT_ROOT/frontend"; then
            log_error "前端镜像构建失败"
            exit 1
        fi
    fi

    log_success "镜像构建完成"
}

# 推送镜像
push_images() {
    if [[ "$PUSH_IMAGES" != "true" ]]; then
        log_info "跳过镜像推送"
        return
    fi

    log_step "推送Docker镜像..."

    # 推送后端镜像
    log_info "推送后端镜像..."
    if ! docker push "$REGISTRY/law-oa-go:$IMAGE_TAG"; then
        log_error "后端镜像推送失败"
        exit 1
    fi

    # 推送前端镜像 (如果存在)
    if [[ -d "$PROJECT_ROOT/frontend" ]]; then
        log_info "推送前端镜像..."
        if ! docker push "$REGISTRY/law-oa-frontend:$IMAGE_TAG"; then
            log_error "前端镜像推送失败"
            exit 1
        fi
    fi

    log_success "镜像推送完成"
}

# 部署命名空间
deploy_namespaces() {
    if [[ "$SKIP_NAMESPACES" == "true" ]]; then
        log_info "跳过命名空间部署"
        return
    fi

    log_step "部署命名空间..."

    for namespace_file in "$K8S_DIR"/namespaces/*.yaml; do
        if [[ -f "$namespace_file" ]]; then
            log_info "部署命名空间: $(basename "$namespace_file" .yaml)"
            if ! kubectl apply -f "$namespace_file"; then
                log_error "命名空间部署失败: $(basename "$namespace_file" .yaml)"
                exit 1
            fi
        fi
    done

    log_success "命名空间部署完成"
}

# 部署密钥
deploy_secrets() {
    if [[ "$SKIP_SECRETS" == "true" ]]; then
        log_info "跳过密钥部署"
        return
    fi

    log_step "部署密钥..."

    for secret_file in "$K8S_DIR"/secrets/*.yaml; do
        if [[ -f "$secret_file" ]]; then
            log_info "部署密钥: $(basename "$secret_file" .yaml)"
            if ! kubectl apply -f "$secret_file" -n "$NAMESPACE"; then
                log_error "密钥部署失败: $(basename "$secret_file" .yaml)"
                exit 1
            fi
        fi
    done

    log_success "密钥部署完成"
}

# 部署配置映射
deploy_configmaps() {
    if [[ "$SKIP_CONFIGMAPS" == "true" ]]; then
        log_info "跳过配置映射部署"
        return
    fi

    log_step "部署配置映射..."

    for config_file in "$K8S_DIR"/configmaps/*.yaml; do
        if [[ -f "$config_file" ]]; then
            log_info "部署配置映射: $(basename "$config_file" .yaml)"
            if ! kubectl apply -f "$config_file" -n "$NAMESPACE"; then
                log_error "配置映射部署失败: $(basename "$config_file" .yaml)"
                exit 1
            fi
        fi
    done

    log_success "配置映射部署完成"
}

# 部署部署
deploy_deployments() {
    if [[ "$SKIP_DEPLOYMENTS" == "true" ]]; then
        log_info "跳过部署创建"
        return
    fi

    log_step "部署应用..."

    # 更新镜像标签
    log_info "更新镜像标签到: $IMAGE_TAG"
    kubectl set image deployment/law-oa-backend law-oa-backend="$REGISTRY/law-oa-go:$IMAGE_TAG" -n "$NAMESPACE" || true
    kubectl set image deployment/law-oa-frontend law-oa-frontend="$REGISTRY/law-oa-frontend:$IMAGE_TAG" -n "$NAMESPACE" || true

    # 应用部署配置
    for deploy_file in "$K8S_DIR"/deployments/*.yaml; do
        if [[ -f "$deploy_file" ]]; then
            log_info "部署应用: $(basename "$deploy_file" .yaml)"
            if ! kubectl apply -f "$deploy_file" -n "$NAMESPACE"; then
                log_error "应用部署失败: $(basename "$deploy_file" .yaml)"
                exit 1
            fi
        fi
    done

    log_success "应用部署完成"
}

# 部署服务
deploy_services() {
    if [[ "$SKIP_SERVICES" == "true" ]]; then
        log_info "跳过服务创建"
        return
    fi

    log_step "部署服务..."

    for service_file in "$K8S_DIR"/services/*.yaml; do
        if [[ -f "$service_file" ]]; then
            log_info "部署服务: $(basename "$service_file" .yaml)"
            if ! kubectl apply -f "$service_file" -n "$NAMESPACE"; then
                log_error "服务部署失败: $(basename "$service_file" .yaml)"
                exit 1
            fi
        fi
    done

    log_success "服务部署完成"
}

# 等待部署完成
wait_for_rollout() {
    if [[ "$WAIT_FOR_ROLLOUT" != "true" ]]; then
        log_info "跳过等待部署完成"
        return
    fi

    log_step "等待部署完成..."

    # 等待后端部署
    log_info "等待后端部署完成..."
    if ! kubectl rollout status deployment/law-oa-backend -n "$NAMESPACE" --timeout=300s; then
        log_error "后端部署超时"
        exit 1
    fi

    # 等待前端部署
    if kubectl get deployment law-oa-frontend -n "$NAMESPACE" >/dev/null 2>&1; then
        log_info "等待前端部署完成..."
        if ! kubectl rollout status deployment/law-oa-frontend -n "$NAMESPACE" --timeout=300s; then
            log_error "前端部署超时"
            exit 1
        fi
    fi

    log_success "部署完成"
}

# 验证部署
verify_deployment() {
    if [[ "$VERIFY" != "true" ]]; then
        log_info "跳过部署验证"
        return
    fi

    log_step "验证部署..."

    # 检查Pod状态
    log_info "检查Pod状态..."
    pod_status=$(kubectl get pods -n "$NAMESPACE" --no-headers)
    if [[ -z "$pod_status" ]]; then
        log_error "没有找到运行的Pod"
        exit 1
    fi

    # 检查Pod就绪状态
    not_ready=$(kubectl get pods -n "$NAMESPACE" --no-headers | awk '{print $3}' | grep -v '1/1' | wc -l)
    if [[ "$not_ready" -gt 0 ]]; then
        log_warning "有 $not_ready 个Pod未就绪"
    else
        log_success "所有Pod都已就绪"
    fi

    # 检查服务状态
    log_info "检查服务状态..."
    if ! kubectl get services -n "$NAMESPACE" >/dev/null 2>&1; then
        log_error "服务创建失败"
        exit 1
    fi

    log_success "部署验证通过"
}

# 健康检查
health_check() {
    log_step "执行健康检查..."

    # 检查后端健康状态
    log_info "检查后端健康状态..."
    backend_pod=$(kubectl get pods -n "$NAMESPACE" -l app=law-oa,component=backend -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -n "$backend_pod" ]]; then
        if ! kubectl exec -n "$NAMESPACE" "$backend_pod" -- curl -f http://localhost:8080/health >/dev/null 2>&1; then
            log_error "后端健康检查失败"
            exit 1
        fi
        log_success "后端健康检查通过"
    else
        log_warning "未找到后端Pod，跳过健康检查"
    fi

    # 检查前端健康状态
    frontend_pod=$(kubectl get pods -n "$NAMESPACE" -l app=law-oa,component=frontend -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -n "$frontend_pod" ]]; then
        if ! kubectl exec -n "$NAMESPACE" "$frontend_pod" -- curl -f http://localhost:80 >/dev/null 2>&1; then
            log_error "前端健康检查失败"
            exit 1
        fi
        log_success "前端健康检查通过"
    else
        log_warning "未找到前端Pod，跳过健康检查"
    fi

    log_success "健康检查完成"
}

# 删除部署
delete_deployment() {
    log_step "删除部署..."

    # 删除部署
    log_info "删除部署..."
    for deploy_file in "$K8S_DIR"/deployments/*.yaml; do
        if [[ -f "$deploy_file" ]]; then
            log_info "删除部署: $(basename "$deploy_file" .yaml)"
            kubectl delete -f "$deploy_file" -n "$NAMESPACE" --ignore-not-found=true || true
        fi
    done

    # 删除服务
    log_info "删除服务..."
    for service_file in "$K8S_DIR"/services/*.yaml; do
        if [[ -f "$service_file" ]]; then
            log_info "删除服务: $(basename "$service_file" .yaml)"
            kubectl delete -f "$service_file" -n "$NAMESPACE" --ignore-not-found=true || true
        fi
    done

    # 删除配置映射
    log_info "删除配置映射..."
    for config_file in "$K8S_DIR"/configmaps/*.yaml; do
        if [[ -f "$config_file" ]]; then
            log_info "删除配置映射: $(basename "$config_file" .yaml)"
            kubectl delete -f "$config_file" -n "$NAMESPACE" --ignore-not-found=true || true
        fi
    done

    # 删除密钥
    log_info "删除密钥..."
    for secret_file in "$K8S_DIR"/secrets/*.yaml; do
        if [[ -f "$secret_file" ]]; then
            log_info "删除密钥: $(basename "$secret_file" .yaml)"
            kubectl delete -f "$secret_file" -n "$NAMESPACE" --ignore-not-found=true || true
        fi
    done

    # 可选：删除命名空间
    read -p "是否删除命名空间 $NAMESPACE? (y/N): " confirm
    if [[ "$confirm" == "y" || "$confirm" == "Y" ]]; then
        log_info "删除命名空间: $NAMESPACE"
        kubectl delete namespace "$NAMESPACE" --ignore-not-found=true || true
    fi

    log_success "部署删除完成"
}

# 查看状态
show_status() {
    log_step "查看应用状态..."

    echo ""
    echo -e "${CYAN}=== 命名空间 ===${NC}"
    kubectl get namespace "$NAMESPACE" || true

    echo ""
    echo -e "${CYAN}=== Pod 状态 ===${NC}"
    kubectl get pods -n "$NAMESPACE" || true

    echo ""
    echo -e "${CYAN}=== 服务状态 ===${NC}"
    kubectl get services -n "$NAMESPACE" || true

    echo ""
    echo -e "${CYAN}=== 部署状态 ===${NC}"
    kubectl get deployments -n "$NAMESPACE" || true

    echo ""
    echo -e "${CYAN}=== 水平Pod自动扩缩器状态 ===${NC}"
    kubectl get hpa -n "$NAMESPACE" || true
}

# 查看日志
show_logs() {
    log_step "查看应用日志..."

    # 查看后端日志
    backend_pod=$(kubectl get pods -n "$NAMESPACE" -l app=law-oa,component=backend -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
    if [[ -n "$backend_pod" ]]; then
        log_info "查看后端日志..."
        kubectl logs -f "$backend_pod" -n "$NAMESPACE"
    else
        log_warning "未找到后端Pod"
    fi
}

# 滚动更新
perform_rollout() {
    log_step "执行滚动更新..."

    # 更新镜像标签
    log_info "更新镜像标签到: $IMAGE_TAG"
    kubectl set image deployment/law-oa-backend law-oa-backend="$REGISTRY/law-oa-go:$IMAGE_TAG" -n "$NAMESPACE"
    kubectl set image deployment/law-oa-frontend law-oa-frontend="$REGISTRY/law-oa-frontend:$IMAGE_TAG" -n "$NAMESPACE"

    # 等待部署完成
    wait_for_rollout

    log_success "滚动更新完成"
}

# 扩缩应用
scale_application() {
    local component=$1
    local replicas=$2

    if [[ -z "$component" || -z "$replicas" ]]; then
        log_error "请指定组件和副本数"
        exit 1
    fi

    log_step "扩缩应用: $component 到 $replicas 个副本"

    if ! kubectl scale deployment "law-oa-$component" --replicas="$replicas" -n "$NAMESPACE"; then
        log_error "扩缩失败: $component"
        exit 1
    fi

    # 等待扩缩完成
    kubectl rollout status "deployment/law-oa-$component" -n "$NAMESPACE" --timeout=300s

    log_success "扩缩完成: $component 到 $replicas 个副本"
}

# 验证配置
validate_config() {
    log_step "验证Kubernetes配置..."

    # 检查配置文件语法
    for dir in "$K8S_DIR"/{namespaces,secrets,configmaps,deployments,services}; do
        if [[ -d "$dir" ]]; then
            for file in "$dir"/*.yaml; do
                if [[ -f "$file" ]]; then
                    log_info "验证配置文件: $(basename "$file")"
                    if ! kubectl apply --dry-run=client -f "$file" >/dev/null 2>&1; then
                        log_error "配置文件验证失败: $(basename "$file")"
                        exit 1
                    fi
                fi
            done
        fi
    done

    log_success "配置验证通过"
}

# 清理资源
clean_resources() {
    log_step "清理Kubernetes资源..."

    # 删除所有资源
    kubectl delete namespace "$NAMESPACE" --ignore-not-found=true --grace-period=30s || true

    log_success "资源清理完成"
}

# 主函数
main() {
    local command="deploy"

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            deploy|delete|status|logs|rollout|scale|health|validate|clean)
                command="$1"
                shift
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -c|--context)
                CONTEXT="$2"
                shift 2
                ;;
            -k|--kubeconfig)
                KUBECONFIG="$2"
                shift 2
                ;;
            -t|--tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -b|--build)
                BUILD_IMAGES="true"
                shift
                ;;
            -p|--push)
                PUSH_IMAGES="true"
                shift
                ;;
            --skip-namespaces)
                SKIP_NAMESPACES="true"
                shift
                ;;
            --skip-secrets)
                SKIP_SECRETS="true"
                shift
                ;;
            --skip-configmaps)
                SKIP_CONFIGMAPS="true"
                shift
                ;;
            --skip-deployments)
                SKIP_DEPLOYMENTS="true"
                shift
                ;;
            --skip-services)
                SKIP_SERVICES="true"
                shift
                ;;
            --no-verify)
                VERIFY="false"
                shift
                ;;
            --no-wait)
                WAIT_FOR_ROLLOUT="false"
                shift
                ;;
            --health-timeout)
                HEALTH_CHECK_TIMEOUT="$2"
                shift 2
                ;;
            -v|--verbose)
                set -x
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

    # 显示标题
    echo -e "${BLUE}"
    cat << "EOF"
 ____  _   _ ____    _       ____    _    _     _
/ ___|| | | |  _ \\  | |     |  _ \\  / \\  | |   | |
\\___ \\| |_| | | | | | |     | | | |/ _ \\ | |   | |
 ___) |  _  | |_| | | |___  | |_| / ___ \\| |___| |___
|____/|_| |_|____/  |_____| |____/_/   \\_\\_____|_____|_____|

    Kubernetes 部署脚本 v2.1.0
EOF
    echo -e "${NC}"
    echo ""

    # 执行命令
    case $command in
        deploy)
            check_dependencies
            build_images
            push_images
            deploy_namespaces
            deploy_secrets
            deploy_configmaps
            deploy_deployments
            deploy_services
            wait_for_rollout
            verify_deployment
            health_check
            show_status
            ;;
        delete)
            check_dependencies
            delete_deployment
            ;;
        status)
            check_dependencies
            show_status
            ;;
        logs)
            check_dependencies
            show_logs
            ;;
        rollout)
            check_dependencies
            perform_rollout
            ;;
        scale)
            shift 2  # 移除命令名
            check_dependencies
            scale_application "$@"
            ;;
        health)
            check_dependencies
            health_check
            ;;
        validate)
            check_dependencies
            validate_config
            ;;
        clean)
            check_dependencies
            clean_resources
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"