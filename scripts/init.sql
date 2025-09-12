-- 律师事务所管理系统数据库初始化脚本
-- 创建数据库
CREATE DATABASE IF NOT EXISTS law_oa CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE law_oa;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(100),
    phone VARCHAR(20),
    real_name VARCHAR(50),
    avatar VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    last_login_at DATETIME,
    last_login_ip VARCHAR(45),
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_username (username),
    INDEX idx_status (status)
);

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    role_name VARCHAR(50) UNIQUE NOT NULL,
    role_key VARCHAR(50) UNIQUE NOT NULL,
    sort INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_role_key (role_key),
    INDEX idx_status (status)
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    permission_name VARCHAR(100) UNIQUE NOT NULL,
    permission_key VARCHAR(100) UNIQUE NOT NULL,
    parent_id INT,
    path VARCHAR(200),
    component VARCHAR(255),
    icon VARCHAR(100),
    sort INT DEFAULT 0,
    menu_type VARCHAR(20) DEFAULT 'menu',
    status VARCHAR(20) DEFAULT 'active',
    remark VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_permission_key (permission_key),
    INDEX idx_parent_id (parent_id),
    INDEX idx_status (status),
    FOREIGN KEY (parent_id) REFERENCES permissions(id) ON DELETE SET NULL
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    UNIQUE KEY uk_user_role (user_id, role_id)
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE,
    UNIQUE KEY uk_role_permission (role_id, permission_id)
);

-- 客户表
CREATE TABLE IF NOT EXISTS clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    client_name VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    client_type VARCHAR(20) DEFAULT 'individual',
    company VARCHAR(100),
    id_card VARCHAR(20),
    address TEXT,
    contact_person VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_client_name (client_name),
    INDEX idx_phone (phone),
    INDEX idx_client_type (client_type),
    INDEX idx_status (status)
);

-- 律师表
CREATE TABLE IF NOT EXISTS lawyers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    lawyer_name VARCHAR(50) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    license_no VARCHAR(50) UNIQUE,
    position VARCHAR(50),
    department VARCHAR(100),
    specialty TEXT,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_lawyer_name (lawyer_name),
    INDEX idx_phone (phone),
    INDEX idx_license_no (license_no),
    INDEX idx_status (status)
);

-- 案件表
CREATE TABLE IF NOT EXISTS cases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    case_no VARCHAR(50) UNIQUE NOT NULL,
    case_name VARCHAR(200) NOT NULL,
    case_type VARCHAR(50),
    client_id INT,
    lawyer_id INT,
    status VARCHAR(20) DEFAULT 'pending',
    description TEXT,
    project_code VARCHAR(50),
    contract_amount DECIMAL(12,2),
    start_date DATETIME,
    end_date DATETIME,
    team_members TEXT,
    project_type VARCHAR(50),
    principal_info TEXT,
    opponent_info TEXT,
    cause_of_action TEXT,
    assisting_lawyer_id INT,
    billing_method VARCHAR(50),
    conflict_check_status VARCHAR(20) DEFAULT 'pending',
    is_major_risk BOOLEAN DEFAULT FALSE,
    is_mass_case BOOLEAN DEFAULT FALSE,
    is_sensitive_case BOOLEAN DEFAULT FALSE,
    contract_document VARCHAR(500),
    legal_letter_document VARCHAR(500),
    other_documents TEXT,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_case_no (case_no),
    INDEX idx_case_name (case_name),
    INDEX idx_case_type (case_type),
    INDEX idx_client_id (client_id),
    INDEX idx_lawyer_id (lawyer_id),
    INDEX idx_status (status),
    INDEX idx_cause_of_action (cause_of_action),
    INDEX idx_assisting_lawyer_id (assisting_lawyer_id),
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL,
    FOREIGN KEY (lawyer_id) REFERENCES lawyers(id) ON DELETE SET NULL,
    FOREIGN KEY (assisting_lawyer_id) REFERENCES lawyers(id) ON DELETE SET NULL
);

-- 法律实体表
CREATE TABLE IF NOT EXISTS law_entities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    entity_name VARCHAR(200) NOT NULL,
    entity_type VARCHAR(50),
    entity_subtype VARCHAR(50),
    id_card VARCHAR(20),
    license_no VARCHAR(50),
    address TEXT,
    contact_info TEXT,
    risk_level VARCHAR(20) DEFAULT 'low',
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_entity_name (entity_name),
    INDEX idx_entity_type (entity_type),
    INDEX idx_entity_subtype (entity_subtype),
    INDEX idx_status (status)
);

-- 法律实体别名表
CREATE TABLE IF NOT EXISTS law_entity_aliases (
    id INT AUTO_INCREMENT PRIMARY KEY,
    entity_id INT NOT NULL,
    alias_name VARCHAR(200) NOT NULL,
    alias_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_entity_id (entity_id),
    INDEX idx_alias_name (alias_name),
    FOREIGN KEY (entity_id) REFERENCES law_entities(id) ON DELETE CASCADE
);

-- 法律实体关系表
CREATE TABLE IF NOT EXISTS law_entity_relations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source_entity_id INT NOT NULL,
    target_entity_id INT NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    relation_desc TEXT,
    start_date DATETIME,
    end_date DATETIME,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_source_entity_id (source_entity_id),
    INDEX idx_target_entity_id (target_entity_id),
    INDEX idx_relation_type (relation_type),
    FOREIGN KEY (source_entity_id) REFERENCES law_entities(id) ON DELETE CASCADE,
    FOREIGN KEY (target_entity_id) REFERENCES law_entities(id) ON DELETE CASCADE
);

-- 利益冲突检查记录表
CREATE TABLE IF NOT EXISTS conflict_check_records (
    id INT AUTO_INCREMENT PRIMARY KEY,
    case_id INT NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    target_id INT NOT NULL,
    target_name VARCHAR(200) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    conflict_level VARCHAR(20) NOT NULL,
    conflict_desc TEXT NOT NULL,
    related_case_id INT,
    recommendation TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    checked_by VARCHAR(50),
    checked_at DATETIME,
    resolved_by VARCHAR(50),
    resolved_at DATETIME,
    resolution TEXT,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_case_id (case_id),
    INDEX idx_target_id (target_id),
    INDEX idx_conflict_level (conflict_level),
    INDEX idx_status (status),
    INDEX idx_related_case_id (related_case_id),
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (related_case_id) REFERENCES cases(id) ON DELETE SET NULL
);

-- 文档表
CREATE TABLE IF NOT EXISTS documents (
    id INT AUTO_INCREMENT PRIMARY KEY,
    document_no VARCHAR(50) UNIQUE NOT NULL,
    case_id INT NOT NULL,
    client_id INT,
    file_name VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    file_type VARCHAR(100) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    document_type VARCHAR(50) NOT NULL,
    description TEXT,
    tags TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_confidential BOOLEAN DEFAULT FALSE,
    expire_date DATETIME,
    uploader_id INT NOT NULL,
    upload_time DATETIME NOT NULL,
    download_count INT DEFAULT 0,
    last_download_time DATETIME,
    status VARCHAR(20) DEFAULT 'active',
    thumbnail_path VARCHAR(500),
    metadata TEXT,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_document_no (document_no),
    INDEX idx_case_id (case_id),
    INDEX idx_client_id (client_id),
    INDEX idx_document_type (document_type),
    INDEX idx_file_hash (file_hash),
    INDEX idx_uploader_id (uploader_id),
    INDEX idx_status (status),
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE SET NULL,
    FOREIGN KEY (uploader_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 文档版本表
CREATE TABLE IF NOT EXISTS document_versions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    document_id INT NOT NULL,
    version_no INT NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    uploader_id INT NOT NULL,
    change_log TEXT,
    upload_time DATETIME NOT NULL,
    is_current BOOLEAN DEFAULT FALSE,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_document_id (document_id),
    INDEX idx_version_no (version_no),
    INDEX idx_file_hash (file_hash),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (uploader_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 文档权限表
CREATE TABLE IF NOT EXISTS document_permissions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    document_id INT NOT NULL,
    user_id INT NOT NULL,
    permission_type VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL,
    created_by INT NOT NULL,
    expires_at DATETIME,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    INDEX idx_document_id (document_id),
    INDEX idx_user_id (user_id),
    INDEX idx_permission_type (permission_type),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- 文档分类表
CREATE TABLE IF NOT EXISTS document_categories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    category_name VARCHAR(100) UNIQUE NOT NULL,
    category_key VARCHAR(100) UNIQUE NOT NULL,
    parent_id INT,
    description TEXT,
    sort INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    icon VARCHAR(100),
    color VARCHAR(20),
    document_count INT DEFAULT 0,
    remark TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_category_key (category_key),
    INDEX idx_parent_id (parent_id),
    INDEX idx_status (status),
    FOREIGN KEY (parent_id) REFERENCES document_categories(id) ON DELETE SET NULL
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    sort INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    INDEX idx_config_key (config_key),
    INDEX idx_status (status)
);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    username VARCHAR(50),
    operation VARCHAR(100) NOT NULL,
    method VARCHAR(20) NOT NULL,
    path VARCHAR(255) NOT NULL,
    params TEXT,
    ip VARCHAR(45),
    user_agent TEXT,
    status INT DEFAULT 200,
    error_message TEXT,
    execution_time BIGINT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_username (username),
    INDEX idx_operation (operation),
    INDEX idx_method (method),
    INDEX idx_path (path),
    INDEX idx_ip (ip),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 插入默认管理员用户
INSERT INTO users (username, password, email, real_name, status) VALUES 
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin@lawoa.com', '管理员', 'active')
ON DUPLICATE KEY UPDATE username = username;

-- 插入默认角色
INSERT INTO roles (role_name, role_key, sort, remark) VALUES 
('超级管理员', 'admin', 1, '系统超级管理员'),
('律师', 'lawyer', 2, '律师角色'),
('助理', 'assistant', 3, '律师助理'),
('客户经理', 'manager', 4, '客户经理')
ON DUPLICATE KEY UPDATE role_name = role_name;

-- 插入默认权限
INSERT INTO permissions (permission_name, permission_key, path, component, sort, menu_type) VALUES 
('系统管理', 'system', '/system', 'Layout', 1, 'menu'),
('用户管理', 'system:user', '/system/user', 'User', 11, 'menu'),
('角色管理', 'system:role', '/system/role', 'Role', 12, 'menu'),
('权限管理', 'system:permission', '/system/permission', 'Permission', 13, 'menu'),
('客户管理', 'client', '/client', 'Layout', 2, 'menu'),
('客户列表', 'client:list', '/client/list', 'ClientList', 21, 'menu'),
('律师管理', 'lawyer', '/lawyer', 'Layout', 3, 'menu'),
('律师列表', 'lawyer:list', '/lawyer/list', 'LawyerList', 31, 'menu'),
('案件管理', 'case', '/case', 'Layout', 4, 'menu'),
('案件列表', 'case:list', '/case/list', 'CaseList', 41, 'menu'),
('冲突检查', 'conflict', '/conflict', 'Layout', 5, 'menu'),
('冲突检查', 'conflict:check', '/conflict/check', 'ConflictCheck', 51, 'menu'),
('文档管理', 'document', '/document', 'Layout', 6, 'menu'),
('文档列表', 'document:list', '/document/list', 'DocumentList', 61, 'menu'),
('统计分析', 'report', '/report', 'Layout', 7, 'menu'),
('仪表板', 'report:dashboard', '/report/dashboard', 'Dashboard', 71, 'menu')
ON DUPLICATE KEY UPDATE permission_name = permission_name;

-- 为管理员用户分配角色
INSERT INTO user_roles (user_id, role_id) 
SELECT u.id, r.id FROM users u CROSS JOIN roles r 
WHERE u.username = 'admin' AND r.role_key = 'admin'
ON DUPLICATE KEY UPDATE user_id = user_id;

-- 插入默认系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description) VALUES 
('system.name', '律师事务所管理系统', 'string', '系统名称'),
('system.version', '1.0.0', 'string', '系统版本'),
('system.logo', '/logo.png', 'string', '系统Logo'),
('system.description', '专业的律师事务所管理系统', 'string', '系统描述'),
('upload.max_size', '104857600', 'number', '文件上传最大大小（字节）'),
('upload.allowed_types', 'pdf,doc,docx,xls,xlsx,ppt,pptx,txt,jpg,png,gif,webp', 'string', '允许上传的文件类型'),
('upload.path', './uploads', 'string', '文件上传路径'),
('jwt.expire', '7200', 'number', 'JWT过期时间（秒）'),
('rate_limit.requests', '100', 'number', '限流请求数'),
('rate_limit.duration', '60', 'number', '限流时间窗口（秒）')
ON DUPLICATE KEY UPDATE config_key = config_key;

COMMIT;