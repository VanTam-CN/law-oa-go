#!/bin/bash

# Law OA Go 数据库迁移管理脚本
# 支持MySQL到PostgreSQL的数据迁移

set -e

# 颜色定义
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
        log_error "Docker未安装或未运行"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "无法连接到Docker daemon"
        exit 1
    fi
}

# 检查Docker Compose
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装"
        exit 1
    fi
}

# 加载环境变量
load_env() {
    if [ -f ".env.postgresql" ]; then
        export $(cat .env.postgresql | grep -v '^#' | xargs)
        log_info "已加载环境变量配置"
    else
        log_warning "未找到.env.postgresql文件，使用默认配置"
        export POSTGRES_DB=law_oa_db
        export POSTGRES_USER=law_oa_user
        export POSTGRES_PASSWORD=law_oa_password_2024
        export POSTGRES_PORT=5432
    fi
}

# 启动PostgreSQL和Elasticsearch
start_services() {
    log_info "启动PostgreSQL和Elasticsearch服务..."

    # 停止现有容器
    docker-compose -f docker-compose.postgresql.yml down || true

    # 启动数据库和搜索服务
    docker-compose -f docker-compose.postgresql.yml up -d postgresql elasticsearch redis

    log_info "等待服务启动..."
    sleep 10

    # 检查PostgreSQL是否就绪
    for i in {1..30}; do
        if docker-compose -f docker-compose.postgresql.yml exec postgresql pg_isready -U $POSTGRES_USER -d $POSTGRES_DB &> /dev/null; then
            log_success "PostgreSQL已就绪"
            break
        fi
        echo -n "."
        sleep 2
    done

    # 检查Elasticsearch是否就绪
    for i in {1..30}; do
        if curl -s http://localhost:9200/_cluster/health &> /dev/null; then
            log_success "Elasticsearch已就绪"
            break
        fi
        echo -n "."
        sleep 2
    done

    log_success "所有服务已启动"
}

# 停止服务
stop_services() {
    log_info "停止所有服务..."
    docker-compose -f docker-compose.postgresql.yml down
    log_success "服务已停止"
}

# 数据库备份
backup_mysql() {
    log_info "备份MySQL数据库..."

    backup_file="backup/mysql_backup_$(date +%Y%m%d_%H%M%S).sql"
    mkdir -p backup

    mysqldump -h localhost -u root -p$MYSQL_PASSWORD \
        --single-transaction \
        --routines \
        --triggers \
        $MYSQL_DATABASE > "$backup_file"

    log_success "MySQL备份已保存到: $backup_file"
}

# 执行数据迁移
run_migration() {
    log_info "开始执行数据迁移..."

    # 检查PostgreSQL连接
    if ! docker-compose -f docker-compose.postgresql.yml exec postgresql pg_isready -U $POSTGRES_USER -d $POSTGRES_DB &> /dev/null; then
        log_error "PostgreSQL未就绪"
        exit 1
    fi

    # 运行迁移脚本
    cd scripts
    go run migrate-to-postgresql.go

    log_success "数据迁移完成"
}

# 初始化数据库
init_database() {
    log_info "初始化PostgreSQL数据库..."

    # 创建初始化脚本目录
    mkdir -p scripts

    # 运行数据库初始化
    docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -f /docker-entrypoint-initdb.d/01-init.sql

    log_success "数据库初始化完成"
}

# 验证迁移结果
verify_migration() {
    log_info "验证迁移结果..."

    # 检查用户数量
    user_count=$(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM users;" | tr -d ' ')
    client_count=$(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM clients;" | tr -d ' ')
    case_count=$(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM cases;" | tr -d ' ')

    echo ""
    echo "📊 迁移结果统计:"
    echo "  用户数量: $user_count"
    echo "  客户数量: $client_count"
    echo "  案件数量: $case_count"
    echo ""

    if [ "$user_count" -gt 0 ] && [ "$client_count" -gt 0 ] && [ "$case_count" -gt 0 ]; then
        log_success "数据迁移验证通过"
    else
        log_warning "部分数据可能未成功迁移，请检查日志"
    fi
}

# 生成迁移报告
generate_report() {
    log_info "生成迁移报告..."

    report_file="migration_report_$(date +%Y%m%d_%H%M%S).md"

    cat > "$report_file" << EOF
# Law OA Go 数据库迁移报告

## 迁移概述
- **迁移时间**: $(date)
- **源数据库**: MySQL
- **目标数据库**: PostgreSQL
- **迁移工具**: Go迁移脚本

## 环境信息
- **PostgreSQL版本**: $(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT version();" | head -1)
- **Elasticsearch版本**: $(curl -s http://localhost:9200 | grep -o '"number" : "[^"]*"' | head -1)

## 迁移结果
- **用户数量**: $(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM users;")
- **客户数量**: $(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM clients;")
- **案件数量**: $(docker-compose -f docker-compose.postgresql.yml exec postgresql psql -U $POSTGRES_USER -d $POSTGRES_DB -t -c "SELECT COUNT(*) FROM cases;")

## 后续步骤
1. 验证应用程序功能
2. 更新生产环境配置
3. 逐步切换到PostgreSQL
4. 监控系统性能

## 回滚计划
如需回滚到MySQL，请执行：
1. 停止应用服务
2. 恢复MySQL数据
3. 更新配置文件
4. 重启应用服务
EOF

    log_success "迁移报告已生成: $report_file"
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    # 清理可能的临时文件
    rm -f scripts/migrate-to-postgresql
    log_success "清理完成"
}

# 显示帮助信息
show_help() {
    cat << EOF
Law OA Go 数据库迁移工具

用法: $0 [选项]

选项:
    start       启动PostgreSQL和Elasticsearch服务
    stop        停止所有服务
    backup      备份MySQL数据库
    migrate     执行数据迁移
    init        初始化PostgreSQL数据库
    verify      验证迁移结果
    report      生成迁移报告
    cleanup     清理临时文件
    help        显示此帮助信息

示例:
    $0 start        # 启动服务
    $0 backup      # 备份数据
    $0 migrate     # 执行迁移
    $0 verify      # 验证结果
    $0 report      # 生成报告

注意:
- 请确保已正确配置.env.postgresql文件
- 迁移前请先备份重要数据
- 建议在测试环境中先验证迁移流程
EOF
}

# 主函数
main() {
    case "${1:-}" in
        "start")
            check_docker
            check_docker_compose
            load_env
            start_services
            ;;
        "stop")
            check_docker
            stop_services
            ;;
        "backup")
            load_env
            backup_mysql
            ;;
        "migrate")
            check_docker
            load_env
            run_migration
            ;;
        "init")
            check_docker
            load_env
            start_services
            init_database
            ;;
        "verify")
            check_docker
            load_env
            verify_migration
            ;;
        "report")
            generate_report
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            show_help
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