-- PostgreSQL 完整数据库结构补全脚本
-- 基于MySQL完整结构创建PostgreSQL对应的所有表
-- 执行前请确保已安装必要的扩展：uuid-ossp, pg_trgm

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建枚举类型
DO $$ BEGIN
    -- 用户角色枚举
    CREATE TYPE user_role AS ENUM ('admin', 'lawyer', 'user', 'assistant', 'manager');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 用户状态枚举
    CREATE TYPE user_status AS ENUM ('active', 'inactive', 'pending');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 客户类型枚举
    CREATE TYPE client_type AS ENUM ('individual', 'company', '个人', '企业');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 案件类型枚举
    CREATE TYPE case_type AS ENUM ('民事', '刑事', '行政', '商事', '劳动', '知识产权', '其他');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 案件优先级枚举
    CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'urgent', 'normal');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 案件状态枚举
    CREATE TYPE case_status AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived', 'draft', 'ongoing', 'closed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 冲突风险等级枚举
    CREATE TYPE conflict_risk_level AS ENUM ('HIGH', 'MEDIUM', 'LOW', 'MINIMAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 检查状态枚举
    CREATE TYPE check_status AS ENUM ('PROCESSING', 'COMPLETED', 'FAILED', 'pending');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 关系类型枚举
    CREATE TYPE relation_type AS ENUM ('PARENT', 'SUBSIDIARY', 'SISTER', 'COMPETITOR', 'ADVERSE', 'OTHER');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 文档类型枚举
    CREATE TYPE document_type AS ENUM ('pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt', 'jpg', 'png', 'gif', 'webp');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 通知类型枚举
    CREATE TYPE notification_type AS ENUM ('system', 'case', 'client', 'finance');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 日程类型枚举
    CREATE TYPE schedule_type AS ENUM ('meeting', 'hearing', 'deadline', 'task');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    -- 财务记录类型枚举
    CREATE TYPE financial_type AS ENUM ('income', 'expense');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ===================================
-- 1. 核心用户权限系统表
-- ===================================

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL,
    role_key VARCHAR(50) NOT NULL UNIQUE,
    sort INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    permission_name VARCHAR(100) NOT NULL UNIQUE,
    permission_key VARCHAR(100) NOT NULL UNIQUE,
    parent_id BIGINT,
    path VARCHAR(200),
    component VARCHAR(255),
    icon VARCHAR(100),
    sort INTEGER DEFAULT 0,
    menu_type VARCHAR(20) DEFAULT 'menu',
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, role_id)
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (role_id, permission_id)
);

-- 部门表
CREATE TABLE IF NOT EXISTS departments (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    parent_id BIGINT DEFAULT 0,
    leader_id BIGINT,
    description TEXT,
    sort_order INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- 2. 核心业务表（补充缺失字段）
-- ===================================

-- 更新用户表，添加缺失字段
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'real_name') THEN
        ALTER TABLE users ADD COLUMN real_name VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_at') THEN
        ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_ip') THEN
        ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(45);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role_id') THEN
        ALTER TABLE users ADD COLUMN role_id BIGINT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'department_id') THEN
        ALTER TABLE users ADD COLUMN department_id BIGINT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'remark') THEN
        ALTER TABLE users ADD COLUMN remark TEXT;
    END IF;
END $$;

-- 律师表
CREATE TABLE IF NOT EXISTS lawyers (
    id BIGSERIAL PRIMARY KEY,
    lawyer_name VARCHAR(50) NOT NULL,
    phone VARCHAR(20),
    email VARCHAR(100),
    license_no VARCHAR(50) UNIQUE,
    position VARCHAR(50),
    department VARCHAR(100),
    specialty TEXT,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 更新客户表，添加缺失字段
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'client_name') THEN
        ALTER TABLE clients ADD COLUMN client_name VARCHAR(100);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'company') THEN
        ALTER TABLE clients ADD COLUMN company VARCHAR(100);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'lawyer_id') THEN
        ALTER TABLE clients ADD COLUMN lawyer_id BIGINT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'remark') THEN
        ALTER TABLE clients ADD COLUMN remark TEXT;
    END IF;
END $$;

-- 更新案件表，添加缺失字段
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_no') THEN
        ALTER TABLE cases ADD COLUMN case_no VARCHAR(50) UNIQUE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_name') THEN
        ALTER TABLE cases ADD COLUMN case_name VARCHAR(200);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'assisting_lawyer_id') THEN
        ALTER TABLE cases ADD COLUMN assisting_lawyer_id BIGINT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'project_code') THEN
        ALTER TABLE cases ADD COLUMN project_code VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'contract_amount') THEN
        ALTER TABLE cases ADD COLUMN contract_amount DECIMAL(12,2);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'team_members') THEN
        ALTER TABLE cases ADD COLUMN team_members TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'project_type') THEN
        ALTER TABLE cases ADD COLUMN project_type VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'principal_info') THEN
        ALTER TABLE cases ADD COLUMN principal_info TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'opponent_info') THEN
        ALTER TABLE cases ADD COLUMN opponent_info TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'cause_of_action') THEN
        ALTER TABLE cases ADD COLUMN cause_of_action TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'billing_method') THEN
        ALTER TABLE cases ADD COLUMN billing_method VARCHAR(50);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'conflict_check_status') THEN
        ALTER TABLE cases ADD COLUMN conflict_check_status VARCHAR(20) DEFAULT 'pending';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'is_major_risk') THEN
        ALTER TABLE cases ADD COLUMN is_major_risk BOOLEAN DEFAULT FALSE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'is_mass_case') THEN
        ALTER TABLE cases ADD COLUMN is_mass_case BOOLEAN DEFAULT FALSE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'is_sensitive_case') THEN
        ALTER TABLE cases ADD COLUMN is_sensitive_case BOOLEAN DEFAULT FALSE;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'contract_document') THEN
        ALTER TABLE cases ADD COLUMN contract_document VARCHAR(500);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'legal_letter_document') THEN
        ALTER TABLE cases ADD COLUMN legal_letter_document VARCHAR(500);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'other_documents') THEN
        ALTER TABLE cases ADD COLUMN other_documents TEXT;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'remark') THEN
        ALTER TABLE cases ADD COLUMN remark TEXT;
    END IF;
END $$;

-- 案件进度表
CREATE TABLE IF NOT EXISTS case_progress (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL,
    stage VARCHAR(50) NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    due_date DATE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 案件文档表
CREATE TABLE IF NOT EXISTS case_documents (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT,
    mime_type VARCHAR(100),
    description TEXT,
    uploaded_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- 3. 利益冲突检测系统表
-- ===================================

-- 法律实体表
CREATE TABLE IF NOT EXISTS law_entities (
    id BIGSERIAL PRIMARY KEY,
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
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 法律实体别名表
CREATE TABLE IF NOT EXISTS law_entity_aliases (
    id BIGSERIAL PRIMARY KEY,
    entity_id BIGINT NOT NULL,
    alias_name VARCHAR(200) NOT NULL,
    alias_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 法律实体关系表
CREATE TABLE IF NOT EXISTS law_entity_relations (
    id BIGSERIAL PRIMARY KEY,
    source_entity_id BIGINT NOT NULL,
    target_entity_id BIGINT NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    relation_desc TEXT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 冲突检测案例表
CREATE TABLE IF NOT EXISTS conflict_cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    conflict_type VARCHAR(100) NOT NULL,
    risk_level conflict_risk_level DEFAULT 'LOW',
    description TEXT,
    opposing_parties JSONB,
    related_lawyers JSONB,
    case_no VARCHAR(100),
    case_status VARCHAR(50) DEFAULT 'ACTIVE',
    conflict_details TEXT,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 冲突检测规则表
CREATE TABLE IF NOT EXISTS conflict_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT,
    priority INTEGER DEFAULT 1,
    version INTEGER DEFAULT 1,
    mcp_source VARCHAR(255),
    active BOOLEAN DEFAULT TRUE,
    conditions JSONB,
    actions JSONB,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 冲突检测记录表
CREATE TABLE IF NOT EXISTS conflict_check_records (
    check_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    check_status check_status DEFAULT 'PROCESSING',
    has_conflict BOOLEAN DEFAULT FALSE,
    risk_level conflict_risk_level DEFAULT 'LOW',
    search_parameters JSONB,
    check_result JSONB,
    user_id BIGINT,
    check_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    case_id BIGINT,
    target_id BIGINT,
    target_name VARCHAR(200),
    target_type VARCHAR(50),
    conflict_desc TEXT,
    related_case_id BIGINT,
    recommendation TEXT,
    checked_by VARCHAR(50),
    checked_at TIMESTAMP WITH TIME ZONE,
    resolved_by VARCHAR(50),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution TEXT,
    remark TEXT
);

-- 客户关联关系表
CREATE TABLE IF NOT EXISTS client_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    related_client_id UUID NOT NULL,
    relation_type relation_type NOT NULL,
    relation_strength DECIMAL(3,2) DEFAULT 1.00,
    description TEXT,
    relation_detail VARCHAR(500),
    active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (client_id, related_client_id, relation_type)
);

-- MCP标准记录表
CREATE TABLE IF NOT EXISTS mcp_standards (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    version VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    content JSONB,
    standards JSONB,
    best_practices JSONB,
    compliance JSONB,
    risk_thresholds JSONB,
    effective_date DATE,
    source_url VARCHAR(500),
    active BOOLEAN DEFAULT TRUE,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===================================
-- 4. 文档管理系统表
-- ===================================

-- 文档表
CREATE TABLE IF NOT EXISTS documents (
    id BIGSERIAL PRIMARY KEY,
    document_no VARCHAR(50) UNIQUE NOT NULL,
    case_id BIGINT NOT NULL,
    client_id BIGINT,
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
    expire_date TIMESTAMP WITH TIME ZONE,
    uploader_id BIGINT NOT NULL,
    upload_time TIMESTAMP WITH TIME ZONE NOT NULL,
    download_count INTEGER DEFAULT 0,
    last_download_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',
    thumbnail_path VARCHAR(500),
    metadata TEXT,
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 文档版本表
CREATE TABLE IF NOT EXISTS document_versions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL,
    version_no INTEGER NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    file_size BIGINT NOT NULL,
    uploader_id BIGINT NOT NULL,
    change_log TEXT,
    upload_time TIMESTAMP WITH TIME ZONE NOT NULL,
    is_current BOOLEAN DEFAULT FALSE,
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 文档权限表
CREATE TABLE IF NOT EXISTS document_permissions (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    permission_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by BIGINT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'active',
    remark TEXT
);

-- 文档分类表
CREATE TABLE IF NOT EXISTS document_categories (
    id BIGSERIAL PRIMARY KEY,
    category_name VARCHAR(100) UNIQUE NOT NULL,
    category_key VARCHAR(100) UNIQUE NOT NULL,
    parent_id BIGINT,
    description TEXT,
    sort INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    icon VARCHAR(100),
    color VARCHAR(20),
    document_count INTEGER DEFAULT 0,
    remark TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- 5. 系统管理表
-- ===================================

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    config_type VARCHAR(20) DEFAULT 'string',
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    sort INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 操作日志表
CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    username VARCHAR(50),
    operation VARCHAR(100) NOT NULL,
    method VARCHAR(20) NOT NULL,
    path VARCHAR(255) NOT NULL,
    params TEXT,
    ip VARCHAR(45),
    user_agent TEXT,
    status INTEGER DEFAULT 200,
    error_message TEXT,
    execution_time BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===================================
-- 6. 财务记录表
-- ===================================

-- 财务记录表
CREATE TABLE IF NOT EXISTS financial_records (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT,
    client_id BIGINT,
    type financial_type NOT NULL,
    category VARCHAR(50) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(10) DEFAULT 'CNY',
    description TEXT,
    transaction_date DATE NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    payment_method VARCHAR(50),
    invoice_number VARCHAR(100),
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- ===================================
-- 7. 消息通知表
-- ===================================

-- 消息通知表
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type notification_type NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    related_id BIGINT,
    related_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    priority VARCHAR(20) DEFAULT 'normal',
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===================================
-- 8. 日程安排表
-- ===================================

-- 日程安排表
CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    type schedule_type NOT NULL,
    related_id BIGINT,
    related_type VARCHAR(50),
    location VARCHAR(200),
    participants JSONB,
    reminder_time TIMESTAMP WITH TIME ZONE,
    is_all_day BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'scheduled',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===================================
-- 9. 用户行为分析系统表（简化版本）
-- ===================================

-- 用户会话表
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(64) NOT NULL,
    ip_address VARCHAR(45),
    user_agent VARCHAR(500),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    duration BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    page_views INTEGER DEFAULT 0,
    last_active TIMESTAMP WITH TIME ZONE NOT NULL,
    referrer VARCHAR(500),
    source VARCHAR(100),
    campaign VARCHAR(100),
    device_type VARCHAR(50),
    platform VARCHAR(50),
    browser VARCHAR(100),
    location JSONB,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 页面浏览表
CREATE TABLE IF NOT EXISTS page_views (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    url VARCHAR(2000) NOT NULL,
    path VARCHAR(500) NOT NULL,
    title VARCHAR(500),
    referrer VARCHAR(500),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    duration BIGINT DEFAULT 0,
    scroll_depth INTEGER DEFAULT 0,
    viewport_size VARCHAR(20),
    screen_size VARCHAR(20),
    interaction VARCHAR(50),
    is_bounce BOOLEAN DEFAULT FALSE,
    exit_page BOOLEAN DEFAULT FALSE,
    entry_page BOOLEAN DEFAULT FALSE,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 用户事件表
CREATE TABLE IF NOT EXISTS user_events (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_category VARCHAR(50) NOT NULL,
    event_action VARCHAR(50) NOT NULL,
    event_label VARCHAR(100),
    event_value DECIMAL(10,2),
    url VARCHAR(2000),
    element VARCHAR(200),
    properties JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ===================================
-- 创建索引
-- ===================================

-- 角色表索引
CREATE INDEX IF NOT EXISTS idx_roles_role_key ON roles(role_key);
CREATE INDEX IF NOT EXISTS idx_roles_status ON roles(status);

-- 权限表索引
CREATE INDEX IF NOT EXISTS idx_permissions_permission_key ON permissions(permission_key);
CREATE INDEX IF NOT EXISTS idx_permissions_parent_id ON permissions(parent_id);
CREATE INDEX IF NOT EXISTS idx_permissions_status ON permissions(status);

-- 用户角色关联表索引
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

-- 角色权限关联表索引
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

-- 律师表索引
CREATE INDEX IF NOT EXISTS idx_lawyers_lawyer_name ON lawyers(lawyer_name);
CREATE INDEX IF NOT EXISTS idx_lawyers_phone ON lawyers(phone);
CREATE INDEX IF NOT EXISTS idx_lawyers_license_no ON lawyers(license_no);
CREATE INDEX IF NOT EXISTS idx_lawyers_status ON lawyers(status);

-- 案件进度表索引
CREATE INDEX IF NOT EXISTS idx_case_progress_case_id ON case_progress(case_id);
CREATE INDEX IF NOT EXISTS idx_case_progress_stage ON case_progress(stage);
CREATE INDEX IF NOT EXISTS idx_case_progress_status ON case_progress(status);

-- 案件文档表索引
CREATE INDEX IF NOT EXISTS idx_case_documents_case_id ON case_documents(case_id);
CREATE INDEX IF NOT EXISTS idx_case_documents_type ON case_documents(type);
CREATE INDEX IF NOT EXISTS idx_case_documents_uploaded_by ON case_documents(uploaded_by);

-- 法律实体表索引
CREATE INDEX IF NOT EXISTS idx_law_entities_entity_name ON law_entities(entity_name);
CREATE INDEX IF NOT EXISTS idx_law_entities_entity_type ON law_entities(entity_type);
CREATE INDEX IF NOT EXISTS idx_law_entities_status ON law_entities(status);

-- 法律实体别名表索引
CREATE INDEX IF NOT EXISTS idx_law_entity_aliases_entity_id ON law_entity_aliases(entity_id);
CREATE INDEX IF NOT EXISTS idx_law_entity_aliases_alias_name ON law_entity_aliases(alias_name);

-- 法律实体关系表索引
CREATE INDEX IF NOT EXISTS idx_law_entity_relations_source_entity_id ON law_entity_relations(source_entity_id);
CREATE INDEX IF NOT EXISTS idx_law_entity_relations_target_entity_id ON law_entity_relations(target_entity_id);
CREATE INDEX IF NOT EXISTS idx_law_entity_relations_relation_type ON law_entity_relations(relation_type);

-- 冲突检测相关表索引
CREATE INDEX IF NOT EXISTS idx_conflict_cases_client_id ON conflict_cases(client_id);
CREATE INDEX IF NOT EXISTS idx_conflict_cases_case_type ON conflict_cases(case_type);
CREATE INDEX IF NOT EXISTS idx_conflict_cases_risk_level ON conflict_cases(risk_level);

CREATE INDEX IF NOT EXISTS idx_conflict_rules_type ON conflict_rules(type);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_active ON conflict_rules(active);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_priority ON conflict_rules(priority);

CREATE INDEX IF NOT EXISTS idx_conflict_check_records_client_id ON conflict_check_records(client_id);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_check_status ON conflict_check_records(check_status);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_has_conflict ON conflict_check_records(has_conflict);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_risk_level ON conflict_check_records(risk_level);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_check_time ON conflict_check_records(check_time);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_user_id ON conflict_check_records(user_id);

CREATE INDEX IF NOT EXISTS idx_client_relations_client_id ON client_relations(client_id);
CREATE INDEX IF NOT EXISTS idx_client_relations_related_client_id ON client_relations(related_client_id);
CREATE INDEX IF NOT EXISTS idx_client_relations_relation_type ON client_relations(relation_type);

-- 文档管理相关表索引
CREATE INDEX IF NOT EXISTS idx_documents_document_no ON documents(document_no);
CREATE INDEX IF NOT EXISTS idx_documents_case_id ON documents(case_id);
CREATE INDEX IF NOT EXISTS idx_documents_client_id ON documents(client_id);
CREATE INDEX IF NOT EXISTS idx_documents_document_type ON documents(document_type);
CREATE INDEX IF NOT EXISTS idx_documents_file_hash ON documents(file_hash);
CREATE INDEX IF NOT EXISTS idx_documents_uploader_id ON documents(uploader_id);
CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);

CREATE INDEX IF NOT EXISTS idx_document_versions_document_id ON document_versions(document_id);
CREATE INDEX IF NOT EXISTS idx_document_versions_version_no ON document_versions(version_no);
CREATE INDEX IF NOT EXISTS idx_document_versions_file_hash ON document_versions(file_hash);

CREATE INDEX IF NOT EXISTS idx_document_permissions_document_id ON document_permissions(document_id);
CREATE INDEX IF NOT EXISTS idx_document_permissions_user_id ON document_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_document_permissions_permission_type ON document_permissions(permission_type);

CREATE INDEX IF NOT EXISTS idx_document_categories_category_key ON document_categories(category_key);
CREATE INDEX IF NOT EXISTS idx_document_categories_parent_id ON document_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_document_categories_status ON document_categories(status);

-- 系统管理表索引
CREATE INDEX IF NOT EXISTS idx_system_configs_config_key ON system_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_system_configs_status ON system_configs(status);

CREATE INDEX IF NOT EXISTS idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_operation_logs_username ON operation_logs(username);
CREATE INDEX IF NOT EXISTS idx_operation_logs_operation ON operation_logs(operation);
CREATE INDEX IF NOT EXISTS idx_operation_logs_method ON operation_logs(method);
CREATE INDEX IF NOT EXISTS idx_operation_logs_path ON operation_logs(path);
CREATE INDEX IF NOT EXISTS idx_operation_logs_ip ON operation_logs(ip);
CREATE INDEX IF NOT EXISTS idx_operation_logs_created_at ON operation_logs(created_at);

-- 财务记录表索引
CREATE INDEX IF NOT EXISTS idx_financial_records_case_id ON financial_records(case_id);
CREATE INDEX IF NOT EXISTS idx_financial_records_client_id ON financial_records(client_id);
CREATE INDEX IF NOT EXISTS idx_financial_records_type ON financial_records(type);
CREATE INDEX IF NOT EXISTS idx_financial_records_category ON financial_records(category);
CREATE INDEX IF NOT EXISTS idx_financial_records_status ON financial_records(status);
CREATE INDEX IF NOT EXISTS idx_financial_records_transaction_date ON financial_records(transaction_date);

-- 消息通知表索引
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);

-- 日程安排表索引
CREATE INDEX IF NOT EXISTS idx_schedules_user_id ON schedules(user_id);
CREATE INDEX IF NOT EXISTS idx_schedules_type ON schedules(type);
CREATE INDEX IF NOT EXISTS idx_schedules_start_time ON schedules(start_time);
CREATE INDEX IF NOT EXISTS idx_schedules_status ON schedules(status);

-- 用户行为分析表索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_start_time ON user_sessions(start_time);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_last_active ON user_sessions(last_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_source ON user_sessions(source);
CREATE INDEX IF NOT EXISTS idx_user_sessions_device_type ON user_sessions(device_type);

CREATE INDEX IF NOT EXISTS idx_page_views_session_id ON page_views(session_id);
CREATE INDEX IF NOT EXISTS idx_page_views_user_id ON page_views(user_id);
CREATE INDEX IF NOT EXISTS idx_page_views_timestamp ON page_views(timestamp);
CREATE INDEX IF NOT EXISTS idx_page_views_entry_page ON page_views(entry_page);
CREATE INDEX IF NOT EXISTS idx_page_views_exit_page ON page_views(exit_page);

CREATE INDEX IF NOT EXISTS idx_user_events_session_id ON user_events(session_id);
CREATE INDEX IF NOT EXISTS idx_user_events_user_id ON user_events(user_id);
CREATE INDEX IF NOT EXISTS idx_user_events_type ON user_events(event_type);
CREATE INDEX IF NOT EXISTS idx_user_events_category ON user_events(event_category);
CREATE INDEX IF NOT EXISTS idx_user_events_action ON user_events(event_action);
CREATE INDEX IF NOT EXISTS idx_user_events_timestamp ON user_events(timestamp);

-- ===================================
-- 创建更新时间触发器函数（如果不存在）
-- ===================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
DO $$
BEGIN
    -- 为每个需要updated_at字段的表创建触发器
    PERFORM 'CREATE TRIGGER update_' || table_name || '_updated_at BEFORE UPDATE ON ' || table_name || ' FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();'
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND column_name = 'updated_at'
    AND table_name IN (
        'users', 'clients', 'cases', 'lawyers', 'roles', 'permissions',
        'user_roles', 'role_permissions', 'departments', 'case_progress',
        'case_documents', 'law_entities', 'law_entity_aliases', 'law_entity_relations',
        'conflict_cases', 'conflict_rules', 'conflict_check_records', 'client_relations',
        'mcp_standards', 'documents', 'document_versions', 'document_permissions',
        'document_categories', 'system_configs', 'financial_records',
        'notifications', 'schedules', 'user_sessions'
    );
END $$;

-- ===================================
-- 插入初始数据
-- ===================================

-- 插入默认角色数据
INSERT INTO roles (role_name, role_key, sort, remark) VALUES
('超级管理员', 'admin', 1, '系统超级管理员'),
('律师', 'lawyer', 2, '律师角色'),
('助理', 'assistant', 3, '律师助理'),
('客户经理', 'manager', 4, '客户经理')
ON CONFLICT (role_key) DO NOTHING;

-- 插入默认权限数据
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
ON CONFLICT (permission_key) DO NOTHING;

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
ON CONFLICT (config_key) DO NOTHING;

-- 插入默认冲突检测规则
INSERT INTO conflict_rules (id, name, type, category, description, priority, active, conditions, actions) VALUES
(uuid_generate_v4(), '姓名相似性检测', 'NAME_SIMILARITY', 'GENERAL', '检测客户名称与历史案件的相似性', 5, TRUE, '{"threshold": 0.8, "algorithm": "levenshtein"}', '[]'),
(uuid_generate_v4(), '企业关联检测', 'CORPORATE_RELATION', 'GENERAL', '检测企业客户的关联关系冲突', 8, TRUE, '{"checkTypes": ["PARENT", "SUBSIDIARY", "SISTER"]}', '[]'),
(uuid_generate_v4(), '案件冲突检测', 'CASE_CONFLICT', 'GENERAL', '检测同一客户的案件冲突', 7, TRUE, '{"allowMultipleCases": false, "timeWindow": 365}', '[]'),
(uuid_generate_v4(), '对立当事人检测', 'ADVERSE_PARTY', 'GENERAL', '检测对立当事人冲突', 9, TRUE, '{"strictMode": true, "timeWindow": 1095}', '[]'),
(uuid_generate_v4(), '时间重叠检测', 'TIME_OVERLAP', 'GENERAL', '检测案件时间重叠', 3, TRUE, '{"overlapThreshold": 30, "unit": "days"}', '[]');

-- 插入默认MCP标准记录
INSERT INTO mcp_standards (id, version, title, description, effective_date, active) VALUES
(uuid_generate_v4(), '2024.1', 'ABA Model Rules', '美国律师协会利益冲突标准', '2024-01-01', TRUE),
(uuid_generate_v4(), '2024.1', 'Chinese Bar Standards', '中国律师协会利益冲突规定', '2024-01-01', TRUE);

-- 插入默认文档分类
INSERT INTO document_categories (category_name, category_key, sort, status) VALUES
('合同文档', 'contract', 1, 'active'),
('证据材料', 'evidence', 2, 'active'),
('法律文书', 'legal_docs', 3, 'active'),
('沟通记录', 'communication', 4, 'active'),
('财务文件', 'financial', 5, 'active'),
('其他文档', 'other', 6, 'active')
ON CONFLICT (category_key) DO NOTHING;

COMMIT;