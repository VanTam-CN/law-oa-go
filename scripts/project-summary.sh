#!/bin/bash

# 律师事务所管理系统项目概览脚本
# 提供项目整体状态和配置信息

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SUMMARY_FILE="$PROJECT_ROOT/project-summary-$(date +%Y%m%d_%H%M%S).md"

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

log_section() {
    echo -e "\n${PURPLE}========================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}========================================${NC}"
}

# 获取项目基本信息
get_project_info() {
    log_section "项目基本信息"

    echo "项目名称:律师事务所管理系统"
    echo "项目路径: $PROJECT_ROOT"
    echo "当前时间: $(date)"
    echo "Git仓库: $(git remote get-url origin 2>/dev/null || echo '本地项目')"
    echo "当前分支: $(git branch --show-current 2>/dev/null || echo '未知')"
    echo "最新提交: $(git log -1 --format='%h - %s (%cr)' 2>/dev/null || echo '无Git信息')"
    echo ""
}

# 获取技术栈信息
get_tech_stack() {
    log_section "技术栈信息"

    echo "后端技术:"
    echo "  - 语言: Go $(go version 2>/dev/null | cut -d' ' -f3 || echo '未安装')"
    echo "  - 框架: Gin"
    echo "  - 数据库: MySQL 8.0"
    echo "  - 缓存: Redis 7.0"
    echo "  - 搜索: Elasticsearch 8.0"
    echo ""

    echo "前端技术:"
    echo "  - 框架: React Bootstrap (frontend/)"
    echo "  - 框架: Vue 3 (frontend-vue/)"
    echo "  - 端口: 3003"
    echo ""

    echo "容器化:"
    echo "  - Docker: $(docker --version 2>/dev/null | cut -d' ' -f3 || echo '未安装')"
    echo "  - Docker Compose: $(docker-compose --version 2>/dev/null | cut -d' ' -f3 || echo '未安装')"
    echo ""
}

# 获取端口配置
get_port_configuration() {
    log_section "端口配置"

    echo "服务端口分配:"
    echo "  - 前端应用: 3003"
    echo "  - 后端API: 8080"
    echo "  - MySQL数据库: 33060"
    echo "  - Redis: 6379"
    echo "  - Elasticsearch: 9200"
    echo "  - Kibana: 5601"
    echo ""

    echo "管理界面:"
    echo "  - phpMyAdmin: 8081"
    echo "  - Redis Commander: 8082"
    echo ""

    # 检查端口占用情况
    echo "端口占用状态:"
    local ports=(3003 8080 33060 6379 9200)
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo "  - $port: ${RED}占用${NC}"
        else
            echo "  - $port: ${GREEN}空闲${NC}"
        fi
    done
    echo ""
}

# 获取Docker信息
get_docker_info() {
    log_section "Docker信息"

    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        return
    fi

    echo "Docker版本: $(docker --version)"
    echo "Docker状态: $(docker info --format='{{.ServerStatus}}' 2>/dev/null || echo '未运行')"
    echo ""

    if [ -f "$PROJECT_ROOT/docker-compose.yml" ]; then
        echo "Docker Compose配置:"
        echo "  - 主配置: docker-compose.yml"
    fi

    if [ -f "$PROJECT_ROOT/docker-compose.dev.yml" ]; then
        echo "  - 开发配置: docker-compose.dev.yml"
    fi
    echo ""

    # 显示镜像信息
    echo "Docker镜像:"
    docker images | grep "law-oa" | head -5 || echo "  暂无law-oa镜像"
    echo ""

    # 显示容器状态
    echo "Docker容器:"
    if docker-compose ps 2>/dev/null | grep -q "Up"; then
        docker-compose ps
    else
        echo "  当前没有运行的容器"
    fi
    echo ""
}

# 获取数据库信息
get_database_info() {
    log_section "数据库信息"

    echo "MySQL配置:"
    echo "  - 主机: localhost"
    echo "  - 端口: 33060"
    echo "  - 数据库: law_oa"
    echo "  - 用户: root"
    echo ""

    # 测试数据库连接
    if mysql -h localhost -P 33060 -u root -ppassword -e "SELECT 1;" > /dev/null 2>&1; then
        echo "数据库连接: ${GREEN}正常${NC}"

        # 显示数据库信息
        echo "数据库统计:"
        mysql -h localhost -P 33060 -u root -ppassword law_oa -e "
            SELECT
                'users' as table_name, COUNT(*) as record_count FROM users UNION
                SELECT 'clients' as table_name, COUNT(*) as record_count FROM clients UNION
                SELECT 'cases' as table_name, COUNT(*) as record_count FROM cases UNION
                SELECT 'departments' as table_name, COUNT(*) as record_count FROM departments;
        " 2>/dev/null || echo "  无法获取统计信息"
    else
        echo "数据库连接: ${RED}失败${NC}"
    fi
    echo ""
}

# 获取Redis信息
get_redis_info() {
    log_section "Redis信息"

    echo "Redis配置:"
    echo "  - 主机: localhost"
    echo "  - 端口: 6379"
    echo "  - 数据库: 0-15"
    echo ""

    # 测试Redis连接
    if redis-cli -p 6379 ping > /dev/null 2>&1; then
        echo "Redis连接: ${GREEN}正常${NC}"

        # 显示Redis信息
        echo "Redis统计:"
        redis-cli -p 6379 info memory | grep -E "used_memory:|used_memory_human:|mem_fragmentation_ratio:"
    else
        echo "Redis连接: ${RED}失败${NC}"
    fi
    echo ""
}

# 获取Elasticsearch信息
get_elasticsearch_info() {
    log_section "Elasticsearch信息"

    echo "Elasticsearch配置:"
    echo "  - 主机: localhost"
    echo "  - 端口: 9200"
    echo "  - 集群: law-oa-cluster"
    echo ""

    # 测试Elasticsearch连接
    if curl -f -s "http://localhost:9200/_cluster/health" > /dev/null 2>&1; then
        echo "Elasticsearch连接: ${GREEN}正常${NC}"

        # 显示集群信息
        echo "集群健康状态:"
        curl -s "http://localhost:9200/_cluster/health" | jq -r '.status'
        echo "节点数量:"
        curl -s "http://localhost:9200/_cluster/health" | jq -r '.number_of_nodes'
        echo "索引数量:"
        curl -s "http://localhost:9200/_cat/indices" | wc -l
    else
        echo "Elasticsearch连接: ${RED}失败${NC}"
    fi
    echo ""
}

# 获取服务状态
get_service_status() {
    log_section "服务状态"

    echo "服务健康检查:"

    # 检查后端服务
    if curl -f -s "http://localhost:8080/health" > /dev/null 2>&1; then
        echo "  - 后端API: ${GREEN}正常${NC}"
    else
        echo "  - 后端API: ${RED}异常${NC}"
    fi

    # 检查前端服务
    if curl -f -s "http://localhost:3003" > /dev/null 2>&1; then
        echo "  - 前端应用: ${GREEN}正常${NC}"
    else
        echo "  - 前端应用: ${RED}异常${NC}"
    fi

    # 检查API心跳
    if curl -f -s "http://localhost:8080/api/v1/ping" > /dev/null 2>&1; then
        echo "  - API心跳: ${GREEN}正常${NC}"
    else
        echo "  - API心跳: ${RED}异常${NC}"
    fi
    echo ""
}

# 获取项目结构
get_project_structure() {
    log_section "项目结构"

    echo "主要目录:"
    echo "  - frontend/          React Bootstrap前端"
    echo "  - frontend-vue/      Vue前端"
    echo "  - internal/          Go内部包"
    echo "  - scripts/           管理脚本"
    echo "  - config/            配置文件"
    echo "  - docs/              文档"
    echo ""

    echo "重要文件:"
    echo "  - docker-compose.yml         主Docker配置"
    echo "  - docker-compose.dev.yml     开发环境配置"
    echo "  - .env.example              环境变量示例"
    echo "  - README.md                 项目文档"
    echo "  - go.mod                    Go模块定义"
    echo ""

    echo "脚本文件:"
    echo "  - scripts/start-dev.sh             开发环境启动"
    echo "  - scripts/docker-manage.sh         Docker镜像管理"
    echo "  - scripts/test-integration.sh      集成测试"
    echo "  - scripts/project-validation.sh    项目验证"
    echo "  - scripts/deployment-verification.sh 部署验证"
    echo "  - scripts/project-summary.sh       项目概览 (本脚本)"
    echo ""
}

# 生成项目概览报告
generate_summary_report() {
    log_section "生成项目概览报告"

    cat > "$SUMMARY_FILE" << EOF
# 律师事务所管理系统项目概览报告

## 项目基本信息
- **项目名称**: 律师事务所管理系统
- **项目路径**: $PROJECT_ROOT
- **生成时间**: $(date)
- **Git仓库**: $(git remote get-url origin 2>/dev/null || echo '本地项目')
- **当前分支**: $(git branch --show-current 2>/dev/null || echo '未知')
- **最新提交**: $(git log -1 --format='%h - %s (%cr)' 2>/dev/null || echo '无Git信息')

## 技术栈

### 后端
- Go $(go version 2>/dev/null | cut -d' ' -f3 || echo '未安装')
- Gin 框架
- MySQL 8.0
- Redis 7.0
- Elasticsearch 8.0

### 前端
- React Bootstrap (frontend/)
- Vue 3 (frontend-vue/)
- 端口: 3003

### 容器化
- Docker $(docker --version 2>/dev/null | cut -d' ' -f3 || echo '未安装')
- Docker Compose $(docker-compose --version 2>/dev/null | cut -d' ' -f3 || echo '未安装')

## 端口配置

| 服务 | 端口 | 状态 |
|------|------|------|
| 前端应用 | 3003 | $(if lsof -Pi :3003 -sTCP:LISTEN -t >/dev/null 2>&1; then echo "占用"; else echo "空闲"; fi) |
| 后端API | 8080 | $(if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then echo "占用"; else echo "空闲"; fi) |
| MySQL数据库 | 33060 | $(if lsof -Pi :33060 -sTCP:LISTEN -t >/dev/null 2>&1; then echo "占用"; else echo "空闲"; fi) |
| Redis | 6379 | $(if lsof -Pi :6379 -sTCP:LISTEN -t >/dev/null 2>&1; then echo "占用"; else echo "空闲"; fi) |
| Elasticsearch | 9200 | $(if lsof -Pi :9200 -sTCP:LISTEN -t >/dev/null 2>&1; then echo "占用"; else echo "空闲"; fi) |

## 服务状态

| 服务 | 状态 |
|------|------|
| 后端API | $(if curl -f -s "http://localhost:8080/health" > /dev/null 2>&1; then echo "正常"; else echo "异常"; fi) |
| 前端应用 | $(if curl -f -s "http://localhost:3003" > /dev/null 2>&1; then echo "正常"; else echo "异常"; fi) |
| API心跳 | $(if curl -f -s "http://localhost:8080/api/v1/ping" > /dev/null 2>&1; then echo "正常"; else echo "异常"; fi) |

## 管理脚本

### 开发环境管理
- \`./scripts/start-dev.sh start\` - 启动开发环境
- \`./scripts/start-dev.sh stop\` - 停止开发环境
- \`./scripts/start-dev.sh status\` - 查看服务状态
- \`./scripts/start-dev.sh logs\` - 查看日志

### Docker镜像管理
- \`./scripts/docker-manage.sh build\` - 构建镜像
- \`./scripts/docker-manage.sh push\` - 推送镜像
- \`./scripts/docker-manage.sh cleanup\` - 清理镜像
- \`./scripts/docker-manage.sh list\` - 查看镜像

### 测试和验证
- \`./scripts/test-integration.sh test\` - 集成测试
- \`./scripts/project-validation.sh full\` - 项目验证
- \`./scripts/deployment-verification.sh full\` - 部署验证

### 项目概览
- \`./scripts/project-summary.sh\` - 项目概览 (本脚本)

## 访问地址

| 服务 | 地址 |
|------|------|
| 前端应用 | http://localhost:3003 |
| 后端API | http://localhost:8080 |
| API文档 | http://localhost:8080/swagger/index.html |
| 健康检查 | http://localhost:8080/health |
| phpMyAdmin | http://localhost:8081 |
| Redis Commander | http://localhost:8082 |
| Elasticsearch | http://localhost:9200 |
| Kibana | http://localhost:5601 |

## 数据库配置

| 参数 | 值 |
|------|------|
| 主机 | localhost |
| 端口 | 33060 |
| 数据库 | law_oa |
| 用户 | root |
| 密码 | password |

## 快速开始

1. **启动开发环境**:
   \`\`\`bash
   ./scripts/start-dev.sh start
   \`\`\`

2. **运行测试**:
   \`\`\`bash
   ./scripts/project-validation.sh full
   \`\`\`

3. **查看状态**:
   \`\`\`bash
   ./scripts/project-summary.sh
   \`\`\`

4. **访问应用**:
   - 前端: http://localhost:3003
   - 后端API: http://localhost:8080

## 注意事项

1. 确保所有依赖服务已启动 (MySQL, Redis, Elasticsearch)
2. 端口33060已更新，避免与本地MySQL冲突
3. 镜像统一使用law-oa/前缀进行管理
4. 开发环境推荐使用docker-compose.dev.yml配置
5. 定期运行验证脚本确保系统正常运行

---

*报告生成时间: $(date)*
*项目路径: $PROJECT_ROOT*
EOF

    log_success "项目概览报告已生成: $SUMMARY_FILE"
}

# 显示快速开始指南
show_quick_start() {
    log_section "快速开始指南"

    echo "常用命令:"
    echo "  启动开发环境:  ./scripts/start-dev.sh start"
    echo "  查看服务状态:  ./scripts/start-dev.sh status"
    echo "  停止开发环境:  ./scripts/start-dev.sh stop"
    echo "  项目验证:      ./scripts/project-validation.sh full"
    echo "  部署验证:      ./scripts/deployment-verification.sh full"
    echo "  项目概览:      ./scripts/project-summary.sh"
    echo ""

    echo "访问地址:"
    echo "  前端应用:      http://localhost:3003"
    echo "  后端API:       http://localhost:8080"
    echo "  API文档:       http://localhost:8080/swagger/index.html"
    echo "  phpMyAdmin:    http://localhost:8081"
    echo "  Redis管理:     http://localhost:8082"
    echo ""

    echo "数据库连接:"
    echo "  主机:           localhost"
    echo "  端口:           33060"
    echo "  用户:           root"
    echo "  密码:           password"
    echo ""

    echo "Docker镜像:"
    echo "  构建镜像:       ./scripts/docker-manage.sh build"
    echo "  查看镜像:       ./scripts/docker-manage.sh list"
    echo "  清理镜像:       ./scripts/docker-manage.sh cleanup"
    echo ""
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统项目概览脚本

用法: $0 [选项]

选项:
    info            显示项目基本信息
    tech            显示技术栈信息
    ports           显示端口配置
    docker          显示Docker信息
    database        显示数据库信息
    redis           显示Redis信息
    elasticsearch   显示Elasticsearch信息
    services        显示服务状态
    structure       显示项目结构
    quickstart      显示快速开始指南
    report          生成项目概览报告
    help            显示此帮助信息

示例:
    $0               # 显示完整概览
    $0 info          # 显示项目基本信息
    $0 ports         # 显示端口配置
    $0 services      # 显示服务状态
    $0 quickstart    # 显示快速开始指南
    $0 report        # 生成概览报告

EOF
}

# 主函数
main() {
    case "${1:-}" in
        "info")
            get_project_info
            ;;
        "tech")
            get_tech_stack
            ;;
        "ports")
            get_port_configuration
            ;;
        "docker")
            get_docker_info
            ;;
        "database")
            get_database_info
            ;;
        "redis")
            get_redis_info
            ;;
        "elasticsearch")
            get_elasticsearch_info
            ;;
        "services")
            get_service_status
            ;;
        "structure")
            get_project_structure
            ;;
        "quickstart")
            show_quick_start
            ;;
        "report")
            get_project_info
            get_tech_stack
            get_port_configuration
            get_service_status
            generate_summary_report
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        "")
            # 显示完整概览
            get_project_info
            get_tech_stack
            get_port_configuration
            get_service_status
            show_quick_start
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"