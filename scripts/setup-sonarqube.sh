#!/bin/bash

# SonarQube 设置和配置脚本
# Law OA Go 项目 - SonarQube 代码质量分析平台

set -e

# 配置变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SONAR_HOME="$PROJECT_ROOT/.sonar"
SONAR_CONFIG_DIR="$SONAR_HOME/config"
SONAR_DATA_DIR="$SONAR_HOME/data"
SONAR_LOGS_DIR="$SONAR_HOME/logs"

# 颜色配置
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[SONAR-SETUP]${NC} $1"
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
SonarQube 设置脚本 - Law OA Go 项目

用法: $0 [选项]

选项:
    -h, --help              显示帮助信息
    --install-local         在本地安装 SonarQube (Docker)
    --install-server        在服务器安装 SonarQube
    --docker-compose        使用 Docker Compose 启动 SonarQube
    --configure-quality     配置质量门禁
    --setup-webhook         设置 Git Webhook
    --import-issues         导入外部问题
    --backup-config         备份 SonarQube 配置
    --restore-config        恢复 SonarQube 配置
    --create-projects       创建项目配置
    --clean                 清理 SonarQube 数据

示例:
    $0 --docker-compose         # 使用 Docker Compose 启动
    $0 --configure-quality      # 配置质量门禁
    $0 --create-projects        # 创建项目配置

EOF
}

# 创建 SonarQube 目录结构
create_sonar_directories() {
    log "创建 SonarQube 目录结构..."

    mkdir -p "$SONAR_HOME"
    mkdir -p "$SONAR_CONFIG_DIR"
    mkdir -p "$SONAR_DATA_DIR"
    mkdir -p "$SONAR_LOGS_DIR"
    mkdir -p "$PROJECT_ROOT/reports/sonar"

    log_success "SonarQube 目录结构创建完成"
}

# 生成 SonarQube Docker Compose 配置
generate_sonar_docker_compose() {
    log "生成 SonarQube Docker Compose 配置..."

    cat > "$PROJECT_ROOT/docker-compose.sonarqube.yml" << 'EOF'
version: '3.8'

services:
  # SonarQube Community Edition
  sonarqube:
    image: sonarqube:community
    container_name: law-oa-sonarqube
    restart: unless-stopped
    ports:
      - "9000:9000"
      - "9092:9092"
    environment:
      # 基本配置
      - SONAR_ES_BOOTSTRAP_CHECKS_DISABLE=true
      - SONAR_WEB_JAVAADDITIONALOPTS=-server
      - SONAR_CE_JAVAADDITIONALOPTS=-server
      - SONAR_SEARCH_JAVAADDITIONALOPTS=-server

      # 性能优化
      - SONAR_WEB_JAVAOPTS=-Xmx1G -Xms1G -XX:+HeapDumpOnOutOfMemoryError
      - SONAR_CE_JAVAOPTS=-Xmx1G -Xms512m
      - SONAR_SEARCH_JAVAOPTS=-Xmx1G -Xms512m

      # 数据库配置
      - SONAR_JDBC_URL=jdbc:postgresql://sonarqube-db:5432/sonar
      - SONAR_JDBC_USERNAME=sonar
      - SONAR_JDBC_PASSWORD=sonar

      # 安全配置
      - SONAR_FORCEAUTHENTICATION=true
      - SONAR_UPDATECENTER_ACTIVATED=true
      - SONAR_WEB_CONTEXT=/

      # 日志配置
      - SONAR_LOG_LEVEL=INFO
      - SONAR_LOG_CONSOLE=true
    volumes:
      - sonarqube_data:/opt/sonarqube/data
      - sonarqube_logs:/opt/sonarqube/logs
      - sonarqube_extensions:/opt/sonarqube/extensions
      - sonarqube_conf:/opt/sonarqube/conf
    networks:
      - sonar-network
    depends_on:
      sonarqube-db:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/api/system/status"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 30s

  # PostgreSQL 数据库
  sonarqube-db:
    image: postgres:15-alpine
    container_name: law-oa-sonarqube-db
    restart: unless-stopped
    environment:
      - POSTGRES_USER=sonar
      - POSTGRES_PASSWORD=sonar
      - POSTGRES_DB=sonar
      - PGDATA=/var/lib/postgresql/data/pgdata
    volumes:
      - sonarqube_db_data:/var/lib/postgresql/data
      - ./scripts/sonar-postgres.conf:/etc/postgresql/postgresql.conf
    networks:
      - sonar-network
    command: postgres -c config_file=/etc/postgresql/postgresql.conf
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sonar"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s

  # Nginx 反向代理 (可选)
  sonarqube-nginx:
    image: nginx:alpine
    container_name: law-oa-sonarqube-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./scripts/sonar-nginx.conf:/etc/nginx/conf.d/default.conf
      - ./scripts/ssl:/etc/nginx/ssl
    networks:
      - sonar-network
    depends_on:
      - sonarqube
    profiles:
      - production

volumes:
  sonarqube_data:
    driver: local
  sonarqube_logs:
    driver: local
  sonarqube_extensions:
    driver: local
  sonarqube_conf:
    driver: local
  sonarqube_db_data:
    driver: local

networks:
  sonar-network:
    driver: bridge
    ipam:
      driver: default
      config:
        - subnet: 172.21.0.0/16
EOF

    log_success "Docker Compose 配置生成完成"
}

# 生成 PostgreSQL 配置
generate_postgres_config() {
    log "生成 PostgreSQL 配置..."

    cat > "$SCRIPT_DIR/sonar-postgres.conf" << 'EOF'
# PostgreSQL 优化配置 for SonarQube

# 内存配置
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

# 连接配置
max_connections = 200
listen_addresses = '*'

# 日志配置
log_statement = 'all'
log_duration = on
log_min_duration_statement = 1000

# 性能配置
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100

# 安全配置
ssl = off
password_encryption = scram-sha-256
EOF

    log_success "PostgreSQL 配置生成完成"
}

# 生成 Nginx 配置
generate_nginx_config() {
    log "生成 Nginx 配置..."

    cat > "$SCRIPT_DIR/sonar-nginx.conf" << 'EOF'
# Nginx 配置 for SonarQube

upstream sonarqube {
    server sonarqube:9000;
}

server {
    listen 80;
    server_name localhost;

    # 重定向到 HTTPS (生产环境)
    # return 301 https://$server_name$request_uri;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    # 代理配置
    location / {
        proxy_pass http://sonarqube;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # 超时配置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # 缓冲配置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;

        # 大文件上传
        client_max_body_size 50M;
    }

    # 健康检查
    location /health {
        access_log off;
        return 200 "healthy\n";
        add_header Content-Type text/plain;
    }
}

# HTTPS 配置 (生产环境)
# server {
#     listen 443 ssl http2;
#     server_name localhost;
#
#     ssl_certificate /etc/nginx/ssl/cert.pem;
#     ssl_certificate_key /etc/nginx/ssl/key.pem;
#
#     ssl_protocols TLSv1.2 TLSv1.3;
#     ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
#     ssl_prefer_server_ciphers off;
#
#     # 其他配置与 HTTP 相同...
# }
EOF

    log_success "Nginx 配置生成完成"
}

# 配置质量门禁
configure_quality_gate() {
    log "配置 SonarQube 质量门禁..."

    cat > "$SONAR_CONFIG_DIR/quality-gate.json" << 'EOF'
{
  "name": "Law OA Go Quality Gate",
  "conditions": [
    {
      "metric": "coverage",
      "operator": "LT",
      "value": "70",
      "error": "测试覆盖率必须 >= 70%"
    },
    {
      "metric": "duplicated_lines_density",
      "operator": "GT",
      "value": "3",
      "error": "代码重复率必须 <= 3%"
    },
    {
      "metric": "new_violations",
      "operator": "GT",
      "value": "0",
      "error": "不允许新增的违反规则"
    },
    {
      "metric": "new_security_hotspots",
      "operator": "GT",
      "value": "0",
      "error": "不允许新增的安全热点"
    },
    {
      "metric": "new_bugs",
      "operator": "GT",
      "value": "0",
      "error": "不允许新增的 bug"
    },
    {
      "metric": "new_vulnerabilities",
      "operator": "GT",
      "value": "0",
      "error": "不允许新增的漏洞"
    },
    {
      "metric": "maintainability_rating",
      "operator": "GT",
      "value": "1",
      "error": "可维护性评级必须是 A"
    },
    {
      "metric": "reliability_rating",
      "operator": "GT",
      "value": "1",
      "error": "可靠性评级必须是 A"
    },
    {
      "metric": "security_rating",
      "operator": "GT",
      "value": "1",
      "error": "安全性评级必须是 A"
    }
  ]
}
EOF

    # 生成质量配置脚本
    cat > "$SCRIPT_DIR/setup-sonar-quality-gate.sh" << 'EOF'
#!/bin/bash

# SonarQube 质量门禁设置脚本

SONAR_URL="http://localhost:9000"
SONAR_TOKEN=""  # 需要设置实际的 token

echo "设置 SonarQube 质量门禁..."

# 检查 SonarQube 连接
if ! curl -s "$SONAR_URL/api/system/status" > /dev/null; then
    echo "错误: 无法连接到 SonarQube ($SONAR_URL)"
    exit 1
fi

echo "SonarQube 连接成功"

# TODO: 实际的质量门禁配置需要通过 Web UI 或 API 完成
echo "请访问 $SONAR_URL 手动配置质量门禁"

EOF
    chmod +x "$SCRIPT_DIR/setup-sonar-quality-gate.sh"

    log_success "质量门禁配置生成完成"
}

# 创建项目配置
create_project_configs() {
    log "创建 SonarQube 项目配置..."

    cat > "$SONAR_CONFIG_DIR/projects.json" << 'EOF'
{
  "projects": [
    {
      "key": "law-oa-go-backend",
      "name": "Law OA Go Backend",
      "language": "go",
      "paths": [
        "cmd/",
        "internal/",
        "pkg/",
        "api/"
      ],
      "exclusions": [
        "**/*_test.go",
        "**/vendor/**",
        "**/migrations/**",
        "**/scripts/**",
        "**/docs/**"
      ]
    },
    {
      "key": "law-oa-go-frontend",
      "name": "Law OA Go Frontend (React)",
      "language": "ts",
      "paths": [
        "frontend/src/"
      ],
      "exclusions": [
        "**/*.test.*",
        "**/node_modules/**",
        "**/dist/**",
        "**/build/**"
      ]
    },
    {
      "key": "law-oa-go-frontend-vue",
      "name": "Law OA Go Frontend (Vue)",
      "language": "ts",
      "paths": [
        "frontend-vue/src/"
      ],
      "exclusions": [
        "**/*.test.*",
        "**/node_modules/**",
        "**/dist/**"
      ]
    }
  ]
}
EOF

    log_success "项目配置创建完成"
}

# 生成备份脚本
generate_backup_script() {
    log "生成 SonarQube 备份脚本..."

    cat > "$SCRIPT_DIR/backup-sonarqube.sh" << 'EOF'
#!/bin/bash

# SonarQube 数据备份脚本

set -e

BACKUP_DIR="./backups/sonarqube"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/sonarqube_backup_$TIMESTAMP.tar.gz"

echo "开始备份 SonarQube 数据..."

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 停止 SonarQube 服务
docker-compose -f docker-compose.sonarqube.yml stop sonarqube

# 备份数据
docker run --rm \
    -v sonarqube_data:/data \
    -v "$(pwd)/$BACKUP_DIR:/backup" \
    alpine:latest \
    tar czf "/backup/$(basename $BACKUP_FILE)" -C /data .

# 重启 SonarQube 服务
docker-compose -f docker-compose.sonarqube.yml start sonarqube

echo "备份完成: $BACKUP_FILE"

# 清理旧备份 (保留最近7天)
find "$BACKUP_DIR" -name "sonarqube_backup_*.tar.gz" -mtime +7 -delete

echo "备份脚本执行完成"
EOF
    chmod +x "$SCRIPT_DIR/backup-sonarqube.sh"

    log_success "备份脚本生成完成"
}

# 启动 SonarQube (Docker Compose)
start_sonarqube_docker() {
    log "启动 SonarQube (Docker Compose)..."

    cd "$PROJECT_ROOT"

    if [ ! -f "docker-compose.sonarqube.yml" ]; then
        log_error "未找到 docker-compose.sonarqube.yml 文件"
        return 1
    fi

    # 启动服务
    docker-compose -f docker-compose.sonarqube.yml up -d

    log_info "等待 SonarQube 启动..."

    # 等待 SonarQube 启动
    local max_wait=300  # 5分钟
    local wait_time=0

    while [ $wait_time -lt $max_wait ]; do
        if curl -s http://localhost:9000/api/system/status | grep -q "UP"; then
            log_success "SonarQube 启动成功！"
            log_info "访问地址: http://localhost:9000"
            log_info "默认用户名: admin"
            log_info "默认密码: admin"
            break
        fi

        echo -n "."
        sleep 10
        wait_time=$((wait_time + 10))
    done

    if [ $wait_time -ge $max_wait ]; then
        log_error "SonarQube 启动超时"
        return 1
    fi
}

# 清理 SonarQube 数据
clean_sonarqube() {
    log "清理 SonarQube 数据..."

    cd "$PROJECT_ROOT"

    if docker-compose -f docker-compose.sonarqube.yml ps -q | grep -q .; then
        log_info "停止并删除 SonarQube 容器..."
        docker-compose -f docker-compose.sonarqube.yml down -v
    fi

    # 删除 Docker volumes
    docker volume ls -q | grep sonarqube | xargs -r docker volume rm

    log_success "SonarQube 数据清理完成"
}

# 主函数
main() {
    local action=""

    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_help
                exit 0
                ;;
            --install-local)
                action="install_local"
                shift
                ;;
            --docker-compose)
                action="docker_compose"
                shift
                ;;
            --configure-quality)
                action="configure_quality"
                shift
                ;;
            --create-projects)
                action="create_projects"
                shift
                ;;
            --clean)
                action="clean"
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # 如果没有指定操作，默认安装 Docker Compose
    if [ -z "$action" ]; then
        action="docker_compose"
    fi

    # 创建目录
    create_sonar_directories

    # 执行相应操作
    case $action in
        install_local|docker_compose)
            generate_sonar_docker_compose
            generate_postgres_config
            generate_nginx_config
            start_sonarqube_docker
            configure_quality_gate
            create_project_configs
            generate_backup_script
            ;;
        configure_quality)
            configure_quality_gate
            ;;
        create_projects)
            create_project_configs
            ;;
        clean)
            clean_sonarqube
            ;;
    esac

    log_success "SonarQube 设置完成！"
}

# 运行主函数
main "$@"