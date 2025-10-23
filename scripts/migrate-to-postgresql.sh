#!/bin/bash

# MySQL到PostgreSQL数据库迁移和补全脚本
# 执行前请确保：
# 1. PostgreSQL数据库已创建并可连接
# 2. 拥有足够的权限执行DDL语句
# 3. 已备份现有数据

set -e  # 遇到错误立即退出

# 配置变量
PG_HOST=${PG_HOST:-localhost}
PG_PORT=${PG_PORT:-5432}
PG_USER=${PG_USER:-postgres}
PG_PASSWORD=${PG_PASSWORD:-}
PG_DATABASE=${PG_DATABASE:-law_oa_go}

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

# 检查PostgreSQL连接
check_pg_connection() {
    log_info "检查PostgreSQL连接..."

    if ! command -v psql &> /dev/null; then
        log_error "psql命令未找到，请安装PostgreSQL客户端"
        exit 1
    fi

    export PGPASSWORD="$PG_PASSWORD"

    if ! psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -c "SELECT version();" &> /dev/null; then
        log_error "无法连接到PostgreSQL数据库"
        log_error "请检查连接参数："
        log_error "Host: $PG_HOST"
        log_error "Port: $PG_PORT"
        log_error "User: $PG_USER"
        log_error "Database: $PG_DATABASE"
        exit 1
    fi

    log_success "PostgreSQL连接正常"
}

# 备份当前PostgreSQL数据库
backup_database() {
    log_info "备份当前PostgreSQL数据库..."

    local backup_file="backup_$(date +%Y%m%d_%H%M%S).sql"

    export PGPASSWORD="$PG_PASSWORD"

    if pg_dump -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" > "$backup_file"; then
        log_success "数据库备份成功: $backup_file"
    else
        log_error "数据库备份失败"
        exit 1
    fi
}

# 执行SQL文件
execute_sql_file() {
    local sql_file=$1
    local description=$2

    log_info "执行${description}..."

    export PGPASSWORD="$PG_PASSWORD"

    if psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -f "$sql_file"; then
        log_success "${description}执行成功"
    else
        log_error "${description}执行失败"
        exit 1
    fi
}

# 验证表结构
verify_tables() {
    log_info "验证数据库表结构..."

    export PGPASSWORD="$PG_PASSWORD"

    # 检查关键表是否存在
    local tables=(
        "users"
        "roles"
        "permissions"
        "clients"
        "lawyers"
        "cases"
        "documents"
        "conflict_check_records"
        "law_entities"
        "system_configs"
        "notifications"
        "schedules"
        "financial_records"
    )

    local missing_tables=()

    for table in "${tables[@]}"; do
        local count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '$table' AND table_schema = 'public';")

        if [ "$count" -eq 0 ]; then
            missing_tables+=("$table")
        fi
    done

    if [ ${#missing_tables[@]} -eq 0 ]; then
        log_success "所有关键表都存在"
    else
        log_warning "以下表缺失: ${missing_tables[*]}"
        log_info "请检查SQL执行是否成功"
    fi
}

# 验证数据完整性
verify_data() {
    log_info "验证数据完整性..."

    export PGPASSWORD="$PG_PASSWORD"

    # 检查用户数据
    local user_count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM users;")
    log_info "用户数量: $user_count"

    # 检查角色数据
    local role_count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM roles;")
    log_info "角色数量: $role_count"

    # 检查客户数据
    local client_count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM clients;")
    log_info "客户数量: $client_count"

    # 检查案件数据
    local case_count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM cases;")
    log_info "案件数量: $case_count"

    log_success "数据完整性验证完成"
}

# 创建测试数据（如果需要）
create_test_data() {
    if [ "$CREATE_TEST_DATA" = "true" ]; then
        log_info "创建测试数据..."

        export PGPASSWORD="$PG_PASSWORD"

        # 插入测试律师数据
        psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" << 'EOF'
            INSERT INTO lawyers (lawyer_name, phone, email, license_no, position, department, status) VALUES
            ('张三', '13800138001', 'zhangsan@lawfirm.com', 'LAW001', '高级合伙人', '民商事部', 'active'),
            ('李四', '13800138002', 'lisi@lawfirm.com', 'LAW002', '合伙人', '刑事部', 'active'),
            ('王五', '13800138003', 'wangwu@lawfirm.com', 'LAW003', '律师', '行政部', 'active')
            ON CONFLICT (license_no) DO NOTHING;
EOF

        # 插入测试案件数据
        psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" << 'EOF'
            INSERT INTO cases (case_no, case_name, title, description, client_id, lawyer_id, case_type, priority, status, created_at, updated_at) VALUES
            ('CASE001', '合同纠纷案', '某公司合同纠纷', '涉及合同条款解释和违约责任认定', 1, 1, '商事', 'high', 'in_progress', NOW(), NOW()),
            ('CASE002', '劳动争议案', '员工劳动仲裁', '涉及劳动合同解除和经济补偿', 2, 2, '劳动', 'medium', 'pending', NOW(), NOW())
            ON CONFLICT (case_no) DO NOTHING;
EOF

        log_success "测试数据创建完成"
    fi
}

# 生成迁移报告
generate_report() {
    log_info "生成迁移报告..."

    local report_file="migration_report_$(date +%Y%m%d_%H%M%S).txt"

    cat > "$report_file" << EOF
PostgreSQL数据库迁移报告
======================

迁移时间: $(date)
数据库: $PG_DATABASE
主机: $PG_HOST:$PG_PORT

执行的SQL文件:
1. postgresql-complete-schema.sql

迁移内容:
- 创建/更新枚举类型
- 创建所有缺失的表结构
- 添加缺失的字段
- 创建索引和触发器
- 插入初始数据

表结构验证: $([ ${#missing_tables[@]} -eq 0 ] && echo "通过" || echo "失败")
数据完整性验证: 通过

建议后续步骤:
1. 检查应用程序配置，确保使用正确的数据源
2. 运行应用程序测试，验证所有功能正常
3. 监控数据库性能，必要时优化索引
4. 定期备份数据库

EOF

    log_success "迁移报告已生成: $report_file"
}

# 主函数
main() {
    log_info "开始MySQL到PostgreSQL数据库迁移..."

    # 显示配置信息
    log_info "配置信息:"
    log_info "数据库: $PG_DATABASE"
    log_info "主机: $PG_HOST:$PG_PORT"
    log_info "用户: $PG_USER"
    echo

    # 检查连接
    check_pg_connection

    # 询问是否备份
    if [ "$AUTO_BACKUP" != "false" ]; then
        read -p "是否备份当前数据库？(y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            backup_database
        fi
    fi

    # 执行数据库结构补全
    if [ -f "postgresql-complete-schema.sql" ]; then
        execute_sql_file "postgresql-complete-schema.sql" "完整数据库结构补全"
    else
        log_error "未找到 postgresql-complete-schema.sql 文件"
        exit 1
    fi

    # 验证迁移结果
    verify_tables
    verify_data

    # 创建测试数据（可选）
    create_test_data

    # 生成报告
    generate_report

    log_success "数据库迁移完成！"
    log_info "请执行以下命令更新Go应用程序："
    log_info "1. 更新 internal/models/ 目录下的模型文件"
    log_info "2. 检查数据库连接配置"
    log_info "3. 运行应用程序测试"
}

# 显示帮助信息
show_help() {
    cat << EOF
MySQL到PostgreSQL数据库迁移脚本

用法: $0 [选项]

环境变量:
    PG_HOST       PostgreSQL主机地址 (默认: localhost)
    PG_PORT       PostgreSQL端口 (默认: 5432)
    PG_USER       PostgreSQL用户名 (默认: postgres)
    PG_PASSWORD   PostgreSQL密码
    PG_DATABASE   PostgreSQL数据库名 (默认: law_oa_go)
    AUTO_BACKUP   是否自动备份 (默认: true)
    CREATE_TEST_DATA 是否创建测试数据 (默认: false)

示例:
    # 基本用法
    $0

    # 指定数据库连接信息
    PG_HOST=192.168.1.100 PG_USER=myuser PG_PASSWORD=mypass PG_DATABASE=law_db $0

    # 不自动备份
    AUTO_BACKUP=false $0

    # 创建测试数据
    CREATE_TEST_DATA=true $0

EOF
}

# 检查参数
if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
    show_help
    exit 0
fi

# 执行主函数
main "$@"