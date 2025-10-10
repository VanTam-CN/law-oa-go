-- 律师事务所管理系统数据库表结构
-- 创建时间: $(date)

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

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
    last_login_at TIMESTAMP NULL COMMENT '最后登录时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_email (email),
    INDEX idx_role (role_id),
    INDEX idx_department (department_id),
    INDEX idx_status (status)
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

-- 财务记录表
CREATE TABLE IF NOT EXISTS financial_records (
    id INT AUTO_INCREMENT PRIMARY KEY,
    case_id INT COMMENT '关联案件ID',
    client_id INT COMMENT '关联客户ID',
    type VARCHAR(50) NOT NULL COMMENT '类型：income-收入，expense-支出',
    category VARCHAR(50) NOT NULL COMMENT '分类',
    amount DECIMAL(15,2) NOT NULL COMMENT '金额',
    currency VARCHAR(10) DEFAULT 'CNY' COMMENT '货币',
    description TEXT COMMENT '描述',
    transaction_date DATE NOT NULL COMMENT '交易日期',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态：pending-待处理，completed-已完成，cancelled-已取消',
    payment_method VARCHAR(50) COMMENT '支付方式',
    invoice_number VARCHAR(100) COMMENT '发票号',
    created_by INT NOT NULL COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT '删除时间',
    INDEX idx_case (case_id),
    INDEX idx_client (client_id),
    INDEX idx_type (type),
    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_transaction_date (transaction_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务记录表';

-- 消息通知表
CREATE TABLE IF NOT EXISTS notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '接收用户ID',
    type VARCHAR(50) NOT NULL COMMENT '通知类型：system-系统，case-案件，client-客户，finance-财务',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    related_id INT COMMENT '相关记录ID',
    related_type VARCHAR(50) COMMENT '相关记录类型',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读：1-已读，0-未读',
    priority VARCHAR(20) DEFAULT 'normal' COMMENT '优先级：low-低，normal-正常，high-高',
    expires_at TIMESTAMP NULL COMMENT '过期时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_type (type),
    INDEX idx_is_read (is_read),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息通知表';

-- 日程安排表
CREATE TABLE IF NOT EXISTS schedules (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL COMMENT '用户ID',
    title VARCHAR(200) NOT NULL COMMENT '标题',
    description TEXT COMMENT '描述',
    start_time TIMESTAMP NOT NULL COMMENT '开始时间',
    end_time TIMESTAMP NOT NULL COMMENT '结束时间',
    type VARCHAR(50) NOT NULL COMMENT '类型：meeting-会议，hearing-开庭，deadline-截止日期，task-任务',
    related_id INT COMMENT '相关记录ID',
    related_type VARCHAR(50) COMMENT '相关记录类型',
    location VARCHAR(200) COMMENT '地点',
    participants JSON COMMENT '参与者ID列表',
    reminder_time TIMESTAMP NULL COMMENT '提醒时间',
    is_all_day TINYINT DEFAULT 0 COMMENT '是否全天：1-是，0-否',
    status VARCHAR(20) DEFAULT 'scheduled' COMMENT '状态：scheduled-已安排，in_progress-进行中，completed-已完成，cancelled-已取消',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_type (type),
    INDEX idx_start_time (start_time),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日程安排表';

-- 开启外键检查
SET FOREIGN_KEY_CHECKS = 1;