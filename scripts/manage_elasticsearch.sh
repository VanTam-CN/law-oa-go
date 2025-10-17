#!/bin/bash

# Elasticsearch 管理脚本
# 用于管理法条搜索系统的 Elasticsearch 实例

set -e

# 配置
ELASTICSEARCH_CONTAINER="law-oa-elasticsearch"
ELASTICSEARCH_IMAGE="docker.elastic.co/elasticsearch/elasticsearch:8.11.0"
ELASTICSEARCH_PORT=9200
ELASTICSEARCH_HOST="localhost"
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)

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

# 检查Docker是否运行
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不在 PATH 中"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker 守程未运行，请启动 Docker"
        exit 1
    fi
}

# 检查Elasticsearch状态
check_elasticsearch() {
    if curl -s http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/_cluster/health > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# 启动Elasticsearch
start_elasticsearch() {
    log_info "启动 Elasticsearch 容器..."

    # 检查是否已经运行
    if docker ps | grep -q $ELASTICSEARCH_CONTAINER; then
        log_warning "Elasticsearch 容器已经在运行"
        return 0
    fi

    # 创建网络
    if ! docker network ls | grep -q law-oa-network; then
        log_info "创建 Docker 网络..."
        docker network create law-oa-network
    fi

    # 创建数据目录
    mkdir -p $PROJECT_ROOT/data/elasticsearch/{data,logs,plugins}
    chmod 777 $PROJECT_ROOT/data/elasticsearch

    # 启动容器
    docker run -d \
        --name $ELASTICSEARCH_CONTAINER \
        --network law-oa-network \
        -p $ELASTICSEARCH_PORT:9200 \
        -p 9300:9300 \
        -e "discovery.type=single-node" \
        -e "ES_JAVA_OPTS=-Xms1g -Xmx1g" \
        -e "xpack.security.enabled=false" \
        -e "bootstrap.memory_lock=false" \
        -v $PROJECT_ROOT/data/elasticsearch/data:/usr/share/elasticsearch/data \
        -v $PROJECT_ROOT/data/elasticsearch/logs:/usr/share/elasticsearch/logs \
        -v $PROJECT_ROOT/config/elasticsearch/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml \
        $ELASTICSEARCH_IMAGE

    log_info "等待 Elasticsearch 启动..."

    # 等待Elasticsearch启动
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if check_elasticsearch; then
            log_success "Elasticsearch 启动成功"
            return 0
        fi

        log_info "等待 Elasticsearch 启动... ($attempt/$max_attempts)"
        sleep 10
        attempt=$((attempt + 1))
    done

    log_error "Elasticsearch 启动超时"
    return 1
}

# 停止Elasticsearch
stop_elasticsearch() {
    log_info "停止 Elasticsearch 容器..."

    if docker ps | grep -q $ELASTICSEARCH_CONTAINER; then
        docker stop $ELASTICSEARCH_CONTAINER
        docker rm $ELASTICSEARCH_CONTAINER
        log_success "Elasticsearch 容器已停止并删除"
    else
        log_warning "Elasticsearch 容器未运行"
    fi
}

# 重启Elasticsearch
restart_elasticsearch() {
    log_info "重启 Elasticsearch..."
    stop_elasticsearch
    sleep 2
    start_elasticsearch
}

# 查看Elasticsearch状态
status_elasticsearch() {
    log_info "检查 Elasticsearch 状态..."

    if check_elasticsearch; then
        log_success "Elasticsearch 运行正常"

        # 显示集群健康状态
        echo ""
        echo "=== 集群健康状态 ==="
        curl -s http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/_cluster/health?pretty

        # 显示节点信息
        echo ""
        echo "=== 节点信息 ==="
        curl -s http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/_nodes?pretty

        # 显示索引信息
        echo ""
        echo "=== 索引列表 ==="
        curl -s http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/_cat/indices?v

    else
        log_error "Elasticsearch 未运行"
        return 1
    fi
}

# 创建法条索引
create_law_index() {
    log_info "创建法条搜索索引..."

    if ! check_elasticsearch; then
        log_error "Elasticsearch 未运行"
        return 1
    fi

    # 删除现有索引（如果存在）
    curl -X DELETE "http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/legal_statutes" 2>/dev/null || true

    # 创建新索引
    curl -X PUT "http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/legal_statutes" \
        -H 'Content-Type: application/json' \
        -d @config/elasticsearch/legal_statutes_mapping.json

    if [ $? -eq 0 ]; then
        log_success "法条索引创建成功"
    else
        log_error "法条索引创建失败"
        return 1
    fi
}

# 删除法条索引
delete_law_index() {
    log_info "删除法条搜索索引..."

    if ! check_elasticsearch; then
        log_error "Elasticsearch 未运行"
        return 1
    fi

    curl -X DELETE "http://$ELASTICSEARCH_HOST:$ELASTICSEARCH_PORT/legal_statutes"

    if [ $? -eq 0 ]; then
        log_success "法条索引删除成功"
    else
        log_error "法条索引删除失败"
        return 1
    fi
}

# 同步数据到Elasticsearch
sync_data_to_elasticsearch() {
    log_info "同步法条数据到 Elasticsearch..."

    if ! check_elasticsearch; then
        log_error "Elasticsearch 未运行"
        return 1
    fi

    # 调用后端API进行数据同步
    curl -X POST "http://localhost:8080/api/v1/legal/admin/sync-elasticsearch" \
        -H 'Content-Type: application/json'

    if [ $? -eq 0 ]; then
        log_success "数据同步成功"
    else
        log_error "数据同步失败，请检查后端服务"
        return 1
    fi
}

# 重建搜索索引
rebuild_search_index() {
    log_info "重建搜索索引..."

    if ! check_elasticsearch; then
        log_error "Elasticsearch 未运行"
        return 1
    fi

    # 调用后端API重建索引
    curl -X POST "http://localhost:8080/api/v1/legal/admin/rebuild-index" \
        -H 'Content-Type: application/json'

    if [ $? -eq 0 ]; then
        log_success "索引重建成功"
    else
        log_error "索引重建失败，请检查后端服务"
        return 1
    fi
}

# 安装IK分词插件
install_ik_plugin() {
    log_info "安装IK中文分词插件..."

    if docker ps | grep -q $ELASTICSEARCH_CONTAINER; then
        docker exec -it $ELASTICSEARCH_CONTAINER elasticsearch-plugin install https://github.com/medcl/elasticsearch-analysis-ik/releases/download/v8.11.0/elasticsearch-analysis-ik-8.11.0.zip
        log_success "IK插件安装完成，重启容器以生效"
        restart_elasticsearch
    else
        log_error "Elasticsearch 容器未运行，无法安装插件"
        return 1
    fi
}

# 查看Elasticsearch日志
show_logs() {
    if docker ps | grep -q $ELASTICSEARCH_CONTAINER; then
        docker logs -f $ELASTICSEARCH_CONTAINER
    else
        log_error "Elasticsearch 容器未运行"
        return 1
    fi
}

# 清理数据
cleanup() {
    log_info "清理 Elasticsearch 数据..."

    stop_elasticsearch

    # 清理数据目录
    if [ -d "$PROJECT_ROOT/data/elasticsearch" ]; then
        log_warning "删除数据目录: $PROJECT_ROOT/data/elasticsearch"
        rm -rf "$PROJECT_ROOT/data/elasticsearch"
    fi

    # 删除网络
    if docker network ls | grep -q law-oa-network; then
        log_info "删除 Docker 网络"
        docker network rm law-oa-network
    fi

    log_success "清理完成"
}

# 显示帮助信息
show_help() {
    echo "用法: $0 {start|stop|restart|status|create-index|delete-index|sync|rebuild|install-plugin|logs|cleanup|help}"
    echo ""
    echo "命令说明:"
    echo "  start         - 启动 Elasticsearch"
    echo "  stop          - 停止 Elasticsearch"
    echo "  restart       - 重启 Elasticsearch"
    echo "  status        - 查看 Elasticsearch 状态"
    echo "  create-index  - 创建法条搜索索引"
    echo "  delete-index  - 删除法条搜索索引"
    echo "  sync          - 同步数据到 Elasticsearch"
    echo "  rebuild       - 重建搜索索引"
    echo "  install-plugin - 安装IK分词插件"
    echo "  logs          - 查看Elasticsearch日志"
    echo "  cleanup       - 清理所有数据"
    echo "  help          - 显示此帮助信息"
    echo ""
}

# 主函数
main() {
    check_docker

    case "${1:-help}" in
        start)
            start_elasticsearch
            ;;
        stop)
            stop_elasticsearch
            ;;
        restart)
            restart_elasticsearch
            ;;
        status)
            status_elasticsearch
            ;;
        create-index)
            create_law_index
            ;;
        delete-index)
            delete_law_index
            ;;
        sync)
            sync_data_to_elasticsearch
            ;;
        rebuild)
            rebuild_search_index
            ;;
        install-plugin)
            install_ik_plugin
            ;;
        logs)
            show_logs
            ;;
        cleanup)
            cleanup
            ;;
        help|*)
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"