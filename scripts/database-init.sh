#!/bin/bash

# 律师事务所管理系统数据库初始化脚本
# 创建 law_oa 用户和数据库，设置权限

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 配置变量
DB_HOST="localhost"
DB_PORT="33060"
DB_ROOT_USER="root"
DB_ROOT_PASSWORD="password"
DB_NAME="law_oa"
DB_USER="law_oa"
DB_PASSWORD="law_oa_password"

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

# 检查MySQL连接
check_mysql_connection() {
    log_info "检查MySQL连接..."

    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "SELECT 1;" > /dev/null 2>&1; then
        log_success "MySQL连接正常"
        return 0
    else
        log_error "MySQL连接失败，请检查数据库服务状态和配置"
        log_info "配置信息："
        echo "  主机: $DB_HOST"
        echo "  端口: $DB_PORT"
        echo "  用户: $DB_ROOT_USER"
        return 1
    fi
}

# 创建数据库用户
create_database_user() {
    log_section "创建数据库用户"

    # 检查用户是否已存在
    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "SELECT user FROM mysql.user WHERE user = '$DB_USER';" | grep -q "$DB_USER"; then
        log_warning "用户 $DB_USER 已存在，跳过创建"
    else
        log_info "创建用户 $DB_USER..."

        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "
            CREATE USER '$DB_USER'@'%' IDENTIFIED BY '$DB_PASSWORD';
            CREATE USER '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';
        "

        log_success "用户 $DB_USER 创建成功"
    fi
}

# 创建数据库
create_database() {
    log_section "创建数据库"

    # 检查数据库是否已存在
    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "SHOW DATABASES;" | grep -q "$DB_NAME"; then
        log_warning "数据库 $DB_NAME 已存在，跳过创建"
    else
        log_info "创建数据库 $DB_NAME..."

        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "
            CREATE DATABASE $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
        "

        log_success "数据库 $DB_NAME 创建成功"
    fi
}

# 设置用户权限
set_user_privileges() {
    log_section "设置用户权限"

    log_info "为用户 $DB_USER 设置数据库权限..."

    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_ROOT_USER" -p"$DB_ROOT_PASSWORD" -e "
        GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'%';
        GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost';
        FLUSH PRIVILEGES;
    "

    log_success "用户权限设置完成"
}

# 创建数据库表结构
create_database_tables() {
    log_section "创建数据库表结构"

    # 检查是否有表结构文件
    local schema_file="$PROJECT_ROOT/config/schema.sql"

    if [ -f "$schema_file" ]; then
        log_info "使用现有的表结构文件: $schema_file"
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$schema_file"
        log_success "表结构创建完成"
    else
        log_info "创建基础表结构..."

        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" << 'EOF'
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '姓名',
    email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码',
    phone VARCHAR(20) COMMENT '电话',
    avatar VARCHAR(255) COMMENT '头像',
    role_id INT COMMENT '角色ID',
    department_id INT COMMENT '部门ID',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_email (email),
    INDEX idx_role (role_id),
    INDEX idx_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL COMMENT '角色名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '角色编码',
    description TEXT COMMENT '角色描述',
    permissions JSON COMMENT '权限列表',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 部门表
CREATE TABLE IF NOT EXISTS departments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '部门名称',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '部门编码',
    parent_id INT DEFAULT 0 COMMENT '父部门ID',
    leader_id INT COMMENT '负责人ID',
    description TEXT COMMENT '部门描述',
    sort_order INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态：1-启用，0-禁用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_parent (parent_id),
    INDEX idx_leader (leader_id),
    INDEX idx_code (code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- 客户表
CREATE TABLE IF NOT EXISTS clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '客户姓名',
    type VARCHAR(20) NOT NULL COMMENT '客户类型：individual-个人，company-企业',
    id_card VARCHAR(18) COMMENT '身份证号',
    company_name VARCHAR(200) COMMENT '公司名称',
    tax_id VARCHAR(50) COMMENT '税号',
    phone VARCHAR(20) COMMENT '电话',
    email VARCHAR(100) COMMENT '邮箱',
    address TEXT COMMENT '地址',
    lawyer_id INT COMMENT '负责律师ID',
    source VARCHAR(50) COMMENT '客户来源',
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态：active-活跃，inactive-不活跃',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_name (name),
    INDEX idx_lawyer (lawyer_id),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户表';

-- 案件表
CREATE TABLE IF NOT EXISTS cases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(200) NOT NULL COMMENT '案件标题',
    case_number VARCHAR(50) NOT NULL UNIQUE COMMENT '案件编号',
    client_id INT NOT NULL COMMENT '客户ID',
    type VARCHAR(50) NOT NULL COMMENT '案件类型',
    category VARCHAR(50) COMMENT '案件分类',
    description TEXT COMMENT '案件描述',
    main_lawyer_id INT NOT NULL COMMENT '主办律师ID',
    assist_lawyer_ids JSON COMMENT '协办律师ID列表',
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态：draft-草稿，ongoing-进行中，completed-已完成，closed-已结案',
    priority VARCHAR(20) DEFAULT 'normal' COMMENT '优先级：low-低，normal-正常，high-高，urgent-紧急',
    amount DECIMAL(15,2) COMMENT '案件金额',
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    expected_end_date DATE COMMENT '预计结束日期',
    court VARCHAR(100) COMMENT '审理法院',
    judge VARCHAR(50) COMMENT '法官',
    opponent VARCHAR(200) COMMENT '对方当事人',
    opponent_lawyer VARCHAR(100) COMMENT '对方律师',
    created_by INT NOT NULL COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_case_number (case_number),
    INDEX idx_client (client_id),
    INDEX idx_main_lawyer (main_lawyer_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_priority (priority),
    INDEX idx_created_by (created_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='案件表';

-- 案件进度表
CREATE TABLE IF NOT EXISTS case_progress (
    id INT AUTO_INCREMENT PRIMARY KEY,
    case_id INT NOT NULL COMMENT '案件ID',
    stage VARCHAR(50) NOT NULL COMMENT '进度阶段',
    title VARCHAR(200) NOT NULL COMMENT '进度标题',
    description TEXT COMMENT '进度描述',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态：pending-待处理，in_progress-进行中，completed-已完成',
    due_date DATE COMMENT '截止日期',
    completed_at TIMESTAMP NULL COMMENT '完成时间',
    created_by INT NOT NULL COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_case (case_id),
    INDEX idx_stage (stage),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='案件进度表';

-- 案件文档表
CREATE TABLE IF NOT EXISTS case_documents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    case_id INT NOT NULL COMMENT '案件ID',
    name VARCHAR(200) NOT NULL COMMENT '文档名称',
    type VARCHAR(50) NOT NULL COMMENT '文档类型',
    file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
    file_size INT COMMENT '文件大小',
    mime_type VARCHAR(100) COMMENT 'MIME类型',
    description TEXT COMMENT '文档描述',
    uploaded_by INT NOT NULL COMMENT '上传人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_case (case_id),
    INDEX idx_type (type),
    INDEX idx_uploaded_by (uploaded_by)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='案件文档表';

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '权限名称',
    code VARCHAR(100) NOT NULL UNIQUE COMMENT '权限编码',
    resource VARCHAR(100) NOT NULL COMMENT '资源标识',
    action VARCHAR(50) NOT NULL COMMENT '操作类型',
    description TEXT COMMENT '权限描述',
    module VARCHAR(50) COMMENT '所属模块',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_resource (resource),
    INDEX idx_action (action),
    INDEX idx_module (module)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '用户ID',
    role_id INT NOT NULL COMMENT '角色ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user (user_id),
    INDEX idx_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    key_name VARCHAR(100) NOT NULL UNIQUE COMMENT '配置键',
    value TEXT COMMENT '配置值',
    type VARCHAR(20) DEFAULT 'string' COMMENT '数据类型：string, number, boolean, json',
    description TEXT COMMENT '配置描述',
    is_system TINYINT DEFAULT 0 COMMENT '是否系统配置：1-是，0-否',
    created_by INT COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_key (key_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT COMMENT '用户ID',
    user_name VARCHAR(100) COMMENT '用户名',
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    resource VARCHAR(100) COMMENT '资源类型',
    resource_id INT COMMENT '资源ID',
    description TEXT COMMENT '操作描述',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT '用户代理',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_action (action),
    INDEX idx_resource (resource),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';

EOF
        log_success "基础表结构创建完成"
    fi
}

# 插入基础数据
insert_initial_data() {
    log_section "插入基础数据"

    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" << 'EOF'
-- 插入基础权限数据
INSERT INTO permissions (name, code, resource, action, description, module) VALUES
('用户管理', 'user:view', 'user', 'view', '查看用户', 'system'),
('用户管理', 'user:create', 'user', 'create', '创建用户', 'system'),
('用户管理', 'user:update', 'user', 'update', '更新用户', 'system'),
('用户管理', 'user:delete', 'user', 'delete', '删除用户', 'system'),
('角色管理', 'role:view', 'role', 'view', '查看角色', 'system'),
('角色管理', 'role:create', 'role', 'create', '创建角色', 'system'),
('角色管理', 'role:update', 'role', 'update', '更新角色', 'system'),
('角色管理', 'role:delete', 'role', 'delete', '删除角色', 'system'),
('部门管理', 'department:view', 'department', 'view', '查看部门', 'system'),
('部门管理', 'department:create', 'department', 'create', '创建部门', 'system'),
('部门管理', 'department:update', 'department', 'update', '更新部门', 'system'),
('部门管理', 'department:delete', 'department', 'delete', '删除部门', 'system'),
('客户管理', 'client:view', 'client', 'view', '查看客户', 'crm'),
('客户管理', 'client:create', 'client', 'create', '创建客户', 'crm'),
('客户管理', 'client:update', 'client', 'update', '更新客户', 'crm'),
('客户管理', 'client:delete', 'client', 'delete', '删除客户', 'crm'),
('案件管理', 'case:view', 'case', 'view', '查看案件', 'case'),
('案件管理', 'case:create', 'case', 'create', '创建案件', 'case'),
('案件管理', 'case:update', 'case', 'update', '更新案件', 'case'),
('案件管理', 'case:delete', 'case', 'delete', '删除案件', 'case'),
('文档管理', 'document:view', 'document', 'view', '查看文档', 'document'),
('文档管理', 'document:create', 'document', 'create', '创建文档', 'document'),
('文档管理', 'document:update', 'document', 'update', '更新文档', 'document'),
('文档管理', 'document:delete', 'document', 'delete', '删除文档', 'document');

-- 插入基础角色数据
INSERT INTO roles (name, code, description) VALUES
('超级管理员', 'super_admin', '系统超级管理员，拥有所有权限'),
('管理员', 'admin', '系统管理员，拥有大部分权限'),
('律师', 'lawyer', '律师用户，可以管理自己的案件和客户'),
('助理', 'assistant', '律师助理，协助律师处理案件'),
('财务', 'finance', '财务人员，处理财务相关事务'),
('实习生', 'intern', '实习用户，权限有限');

-- 插入基础部门数据
INSERT INTO departments (name, code, description, sort_order) VALUES
('总经办', 'executive', '总经理办公室', 1),
('行政部', 'administration', '行政部门，负责人事和行政事务', 2),
('财务部', 'finance', '财务部门，负责财务和会计事务', 3),
('业务部', 'business', '业务部门，负责业务拓展', 4),
('法务部', 'legal', '法务部门，负责法律事务', 5),
('技术部', 'technology', '技术部门，负责技术支持', 6);

-- 插入基础系统配置
INSERT INTO system_config (key_name, value, type, description, is_system) VALUES
('system.name', '律师事务所管理系统', 'string', '系统名称', 1),
('system.version', '1.0.0', 'string', '系统版本', 1),
('system.company', '律师事务所', 'string', '公司名称', 0),
('system.logo', '/assets/logo.png', 'string', '系统logo路径', 0),
('system.timezone', 'Asia/Shanghai', 'string', '系统时区', 1),
('system.language', 'zh-CN', 'string', '系统语言', 1),
('system.date_format', 'YYYY-MM-DD', 'string', '日期格式', 1),
('system.time_format', 'HH:mm:ss', 'string', '时间格式', 1),
('file.max_size', '10485760', 'number', '文件上传最大大小（字节）', 1),
('file.allowed_types', 'jpg,jpeg,png,gif,pdf,doc,docx,xls,xlsx,ppt,pptx,txt', 'string', '允许上传的文件类型', 1);

EOF

    log_success "基础数据插入完成"
}

# 验证数据库初始化
verify_database_initialization() {
    log_section "验证数据库初始化"

    log_info "检查数据库连接..."
    if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT 1;" > /dev/null 2>&1; then
        log_success "数据库连接正常"
    else
        log_error "数据库连接失败"
        return 1
    fi

    log_info "检查表结构..."
    local tables_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SHOW TABLES;" | wc -l)
    log_success "共创建 $((tables_count - 1)) 个数据表"

    log_info "检查基础数据..."
    local user_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT COUNT(*) FROM roles;" | tail -1)
    local permission_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT COUNT(*) FROM permissions;" | tail -1)
    local department_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT COUNT(*) FROM departments;" | tail -1)

    log_info "基础数据统计:"
    echo "  角色: $user_count 个"
    echo "  权限: $permission_count 个"
    echo "  部门: $department_count 个"

    log_success "数据库初始化验证完成"
}

# 创建应用配置文件
create_application_config() {
    log_section "创建应用配置文件"

    local config_file="$PROJECT_ROOT/.env"

    if [ -f "$config_file" ]; then
        log_warning "配置文件 $config_file 已存在，创建备份"
        cp "$config_file" "$config_file.backup.$(date +%Y%m%d_%H%M%S)"
    fi

    cat > "$config_file" << EOF
# 应用配置
APP_ENV=development
APP_PORT=8080
APP_DEBUG=true

# 数据库配置
DB_HOST=localhost
DB_PORT=33060
DB_NAME=law_oa
DB_USER=law_oa
DB_PASSWORD=law_oa_password

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT 配置
JWT_SECRET=law_oa_jwt_secret_key_$(date +%Y%m%d)
JWT_EXPIRES_IN=24h
JWT_REFRESH_EXPIRES_IN=168h

# Elasticsearch 配置 (可选)
ES_HOSTS=http://localhost:9200
ES_USERNAME=
ES_PASSWORD=

# 文件上传配置
UPLOAD_PATH=./uploads
MAX_FILE_SIZE=10485760

# 邮件配置 (可选)
EMAIL_ENABLED=false
EMAIL_HOST=
EMAIL_PORT=587
EMAIL_USER=
EMAIL_PASSWORD=

# 日志配置
LOG_LEVEL=info
LOG_FILE=./logs/app.log
EOF

    log_success "应用配置文件已创建: $config_file"
}

# 显示数据库连接信息
show_connection_info() {
    log_section "数据库连接信息"

    echo "数据库配置信息："
    echo "  主机: $DB_HOST"
    echo "  端口: $DB_PORT"
    echo "  数据库: $DB_NAME"
    echo "  用户名: $DB_USER"
    echo "  密码: $DB_PASSWORD"
    echo ""

    echo "连接字符串："
    echo "  MySQL: mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $DB_NAME"
    echo "  URL: mysql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
    echo ""

    echo "管理界面："
    echo "  phpMyAdmin: http://localhost:8081"
    echo "  数据库: law_oa"
    echo "  用户名: law_oa"
    echo "  密码: law_oa_password"
    echo ""
}

# 运行完整初始化
run_full_initialization() {
    log_info "开始运行数据库初始化..."

    # 检查MySQL连接
    if ! check_mysql_connection; then
        exit 1
    fi

    # 创建数据库用户
    create_database_user

    # 创建数据库
    create_database

    # 设置用户权限
    set_user_privileges

    # 创建数据库表结构
    create_database_tables

    # 插入基础数据
    insert_initial_data

    # 验证数据库初始化
    verify_database_initialization

    # 创建应用配置文件
    create_application_config

    # 显示连接信息
    show_connection_info

    log_success "数据库初始化完成！"
}

# 显示帮助信息
show_help() {
    cat << EOF
律师事务所管理系统数据库初始化脚本

用法: $0 [命令]

命令:
    full        运行完整初始化 (默认)
    user        创建数据库用户
    database    创建数据库
    tables      创建表结构
    data        插入基础数据
    config      创建应用配置文件
    verify      验证数据库初始化
    info        显示数据库连接信息
    help        显示此帮助信息

示例:
    $0 full          # 运行完整初始化
    $0 user          # 创建数据库用户
    $0 database      # 创建数据库
    $0 tables        # 创建表结构
    $0 data          # 插入基础数据
    $0 config        # 创建应用配置文件

注意事项:
1. 确保MySQL服务正在运行
2. 确保root用户有权限创建用户和数据库
3. 建议先备份数据库再运行此脚本

EOF
}

# 主函数
main() {
    # 设置项目根目录
    PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

    case "${1:-full}" in
        "full")
            run_full_initialization
            ;;
        "user")
            check_mysql_connection
            create_database_user
            ;;
        "database")
            check_mysql_connection
            create_database
            ;;
        "tables")
            check_mysql_connection
            create_database_tables
            ;;
        "data")
            check_mysql_connection
            insert_initial_data
            ;;
        "config")
            create_application_config
            ;;
        "verify")
            verify_database_initialization
            ;;
        "info")
            show_connection_info
            ;;
        "help"|"--help"|"-h")
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