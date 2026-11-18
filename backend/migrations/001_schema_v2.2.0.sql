-- ============================================================================
-- Law OA Go Database Schema v2.2.0
-- ============================================================================
-- 版本: 2.2.0
-- 日期: 2026-02-11
-- 说明: 基于需求规格说明书 v1.2 (CODE FREEZE) 的完整数据库设计
--
-- 包含模块:
--   1. 用户扩展 (Users Extension)
--   2. 待办事项系统 (Inbox System)
--   3. 利益冲突检测 (Conflict Detection)
--   4. 文档管理 (Document Management)
--   5. 财务管理 (Financial Management)
--   6. 通知系统 (Notification System)
--   7. 客户门户 (Client Portal)
--   8. 风控核心 (Risk Control)
--   9. 数据迁移 (Migration)
--   10. 系统配置 (System Config)
-- ============================================================================

-- 设置字符集
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================================================
-- 1. 用户扩展表 (Users Extension)
-- ============================================================================

-- 扩展 users 表，添加风控相关字段
-- 注意：使用存储过程方式兼容不同 MySQL 版本
DROP PROCEDURE IF EXISTS add_users_columns;
DELIMITER //
CREATE PROCEDURE add_users_columns()
BEGIN
    -- 添加 supervisor_id
    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'users'
                   AND column_name = 'supervisor_id') THEN
        ALTER TABLE users ADD COLUMN supervisor_id BIGINT COMMENT '上级/合伙人ID' AFTER department;
    END IF;

    -- 添加 commission_rate
    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'users'
                   AND column_name = 'commission_rate') THEN
        ALTER TABLE users ADD COLUMN commission_rate DECIMAL(5,2) COMMENT '提成比例(%)' AFTER supervisor_id;
    END IF;

    -- 添加 role_type
    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'users'
                   AND column_name = 'role_type') THEN
        ALTER TABLE users ADD COLUMN role_type VARCHAR(50) COMMENT '角色类型: source/lawyer/assistant' AFTER commission_rate;
    END IF;

    -- 添加 offboarding_status
    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'users'
                   AND column_name = 'offboarding_status') THEN
        ALTER TABLE users ADD COLUMN offboarding_status VARCHAR(20) DEFAULT 'active' COMMENT '离职状态: active/offboarding/deactivated' AFTER role_type;
    END IF;
END //
DELIMITER ;

CALL add_users_columns();
DROP PROCEDURE add_users_columns;

-- 添加索引（如果不存在）
DROP PROCEDURE IF EXISTS add_users_indexes;
DELIMITER //
CREATE PROCEDURE add_users_indexes()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.statistics
                   WHERE table_schema = DATABASE()
                   AND table_name = 'users'
                   AND index_name = 'idx_supervisor') THEN
        ALTER TABLE users ADD INDEX idx_supervisor (supervisor_id);
    END IF;
END //
DELIMITER ;

CALL add_users_indexes();
DROP PROCEDURE add_users_indexes;

-- ============================================================================
-- 2. 待办事项系统 (Inbox System)
-- ============================================================================

-- 统一收件箱表
DROP TABLE IF EXISTS inbox_items;
CREATE TABLE inbox_items (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    source_type VARCHAR(50) NOT NULL COMMENT '来源类型: deadline/approval/task',
    source_id BIGINT UNSIGNED NOT NULL COMMENT '来源记录ID',
    title VARCHAR(255) NOT NULL COMMENT '待办标题',
    content TEXT COMMENT '详细内容',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级: critical/high/medium/low',
    due_date DATETIME COMMENT '到期时间',
    due_date_type VARCHAR(50) COMMENT '日期类型: hearing/appeal/evidence等',

    -- 状态字段
    is_read BOOLEAN DEFAULT FALSE COMMENT '是否已读',
    read_at DATETIME COMMENT '阅读时间',
    is_completed BOOLEAN DEFAULT FALSE COMMENT '是否完成',
    completed_at DATETIME COMMENT '完成时间',

    -- 提醒字段
    reminder_sent BOOLEAN DEFAULT FALSE COMMENT '是否已发送提醒',
    reminder_count INT DEFAULT 0 COMMENT '提醒次数',

    -- 升级机制
    escalated BOOLEAN DEFAULT FALSE COMMENT '是否已升级通知上级',
    escalated_at DATETIME COMMENT '升级时间',

    -- 延后功能
    snoozed_until DATETIME COMMENT '延后到何时提醒',
    snoozed_count INT DEFAULT 0 COMMENT '延后次数',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,

    INDEX idx_user_unread (user_id, is_read, due_date),
    INDEX idx_user_priority (user_id, priority, is_completed),
    INDEX idx_source (source_type, source_id),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='统一收件箱-待办事项表';

-- 待办提醒规则表
DROP TABLE IF EXISTS inbox_reminder_rules;
CREATE TABLE inbox_reminder_rules (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    due_date_type VARCHAR(50) NOT NULL COMMENT '日期类型',
    priority VARCHAR(20) NOT NULL COMMENT '优先级',

    -- 提醒时间点配置 (负数表示提前，正数表示延后)
    reminder_offsets JSON NOT NULL COMMENT '提醒偏移量列表，如 [-30, -15, -7, -3, -1, 0]',

    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_rule_type (due_date_type, priority)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='待办提醒规则表';

-- 初始化提醒规则数据
INSERT INTO inbox_reminder_rules (rule_name, due_date_type, priority, reminder_offsets) VALUES
('上诉期限提醒', 'appeal_deadline', 'critical', '[-30, -15, -7, -3, -1, 0]'),
('举证期限提醒', 'evidence_deadline', 'critical', '[-14, -7, -3, -1, 0]'),
('诉讼时效提醒', 'statute_of_limitations', 'critical', '[-90, -30, -7, -1, 0]'),
('执行申请期限提醒', 'execution_deadline', 'critical', '[-180, -90, -30, -7, -1]'),
('开庭日期提醒', 'hearing', 'important', '[-7, -3, -1]'),
('庭前会议提醒', 'pretrial_conference', 'important', '[-3, -1]'),
('调查取证提醒', 'investigation', 'important', '[-3, -1]'),
('缴费期限提醒', 'payment', 'normal', '[-3, -1]'),
('结案归档提醒', 'case_closing', 'normal', '[-7, -3, -1]');

-- ============================================================================
-- 3. 利益冲突检测 (Conflict Detection)
-- ============================================================================

-- 冲突检测预计算池
DROP TABLE IF EXISTS lawyer_conflict_pool;
CREATE TABLE lawyer_conflict_pool (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    lawyer_id BIGINT UNSIGNED NOT NULL COMMENT '律师ID',

    -- 实体信息
    entity_type VARCHAR(20) NOT NULL COMMENT '实体类型: company/individual',
    entity_name VARCHAR(200) NOT NULL COMMENT '标准名称',
    entity_name_standard VARCHAR(200) NOT NULL COMMENT '工商标准名称',
    entity_tax_id VARCHAR(50) COMMENT '税号/统一社会信用代码',
    entity_aliases JSON COMMENT '别名列表',

    -- 关系信息
    relationship_type VARCHAR(50) NOT NULL COMMENT '关系: client/opposing/witness',
    case_id BIGINT UNSIGNED NOT NULL COMMENT '关联案件ID',
    case_title VARCHAR(200) COMMENT '案件标题',

    -- 股权穿透数据
    shareholding_info JSON COMMENT '股权穿透信息',
    related_companies JSON COMMENT '关联公司列表',

    -- 数据来源
    data_source VARCHAR(50) DEFAULT 'manual' COMMENT '数据来源: manual/api/import',
    api_provider VARCHAR(50) COMMENT 'API提供商: qichacha/tianyancha',
    last_verified_at DATETIME COMMENT '最后验证时间',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_lawyer_entity (lawyer_id, entity_name_standard),
    INDEX idx_entity_search (entity_name_standard, entity_type),
    INDEX idx_case (case_id),
    FOREIGN KEY (lawyer_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='律师冲突检测预计算池';

-- 冲突检测报告
DROP TABLE IF EXISTS conflict_reports;
CREATE TABLE conflict_reports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    report_number VARCHAR(50) UNIQUE NOT NULL COMMENT '报告编号',

    -- 检测信息
    checked_by BIGINT UNSIGNED NOT NULL COMMENT '检测人ID',
    check_time DATETIME NOT NULL COMMENT '检测时间',
    check_duration_ms INT COMMENT '检测耗时(毫秒)',

    -- 检测对象
    client_name VARCHAR(200) NOT NULL COMMENT '客户标准名称',
    client_tax_id VARCHAR(50) COMMENT '客户税号',
    opposing_party VARCHAR(200) COMMENT '对方当事人',

    -- 检测结果
    risk_level VARCHAR(20) NOT NULL COMMENT '风险等级: CRITICAL/HIGH/MEDIUM/LOW/PASS',
    matched_cases JSON COMMENT '匹配的案件列表',
    related_companies JSON COMMENT '关联公司列表',
    conflict_details JSON COMMENT '详细冲突信息',

    -- 报告文件
    report_url VARCHAR(500) COMMENT 'PDF报告地址',
    report_generated_at DATETIME COMMENT '报告生成时间',

    -- 审批信息
    reviewed_by BIGINT UNSIGNED COMMENT '复核人ID',
    reviewed_at DATETIME COMMENT '复核时间',
    review_notes TEXT COMMENT '复核意见',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_checked_by (checked_by),
    INDEX idx_client (client_name),
    INDEX idx_risk_level (risk_level),
    INDEX idx_reviewed_by (reviewed_by),
    FOREIGN KEY (checked_by) REFERENCES users(id),
    FOREIGN KEY (reviewed_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测报告表';

-- 工商API调用记录
DROP TABLE IF EXISTS company_api_calls;
CREATE TABLE company_api_calls (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- API信息
    api_provider VARCHAR(50) NOT NULL COMMENT 'API提供商',
    api_endpoint VARCHAR(200) NOT NULL COMMENT 'API端点',
    request_params JSON COMMENT '请求参数',

    -- 搜索信息
    search_keyword VARCHAR(200) NOT NULL COMMENT '搜索关键词',
    matched_company_name VARCHAR(200) COMMENT '匹配到的公司名称',
    matched_company_tax_id VARCHAR(50) COMMENT '匹配到的税号',

    -- 响应信息
    response_status VARCHAR(20) COMMENT '响应状态: success/failed/partial',
    response_data JSON COMMENT '响应数据',
    error_message TEXT COMMENT '错误信息',

    -- 调用信息
    called_by BIGINT UNSIGNED NOT NULL COMMENT '调用者ID',
    call_duration_ms INT COMMENT '调用耗时',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_search (search_keyword),
    INDEX idx_company (matched_company_tax_id),
    INDEX idx_provider (api_provider, created_at),
    INDEX idx_called_by (called_by),
    FOREIGN KEY (called_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工商API调用记录表';

-- 利益冲突定期扫描任务
DROP TABLE IF EXISTS conflict_scan_jobs;
CREATE TABLE conflict_scan_jobs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- 扫描配置
    scan_type VARCHAR(20) NOT NULL COMMENT '扫描类型: daily/weekly/manual',
    scan_scope VARCHAR(50) DEFAULT 'all' COMMENT '扫描范围: all/new_cases/lawyer',

    -- 扫描状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/running/completed/failed',

    -- 扫描结果
    started_at DATETIME COMMENT '开始时间',
    completed_at DATETIME COMMENT '完成时间',
    scanned_cases INT DEFAULT 0 COMMENT '扫描的案件数',
    scanned_lawyers INT DEFAULT 0 COMMENT '扫描的律师数',
    found_conflicts INT DEFAULT 0 COMMENT '发现的冲突数',
    new_conflicts JSON COMMENT '新发现的冲突列表',

    -- 告警状态
    triggered_alerts BOOLEAN DEFAULT FALSE COMMENT '是否已触发告警',
    alert_sent_at DATETIME COMMENT '告警发送时间',

    -- 触发信息
    triggered_by BIGINT UNSIGNED COMMENT '触发者ID (manual时)',
    trigger_reason VARCHAR(200) COMMENT '触发原因',

    error_message TEXT COMMENT '错误信息',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_type_time (scan_type, created_at),
    INDEX idx_triggered_by (triggered_by),
    FOREIGN KEY (triggered_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突扫描任务表';

-- ============================================================================
-- 4. 文档管理 (Document Management)
-- ============================================================================

-- 文档锁表
DROP TABLE IF EXISTS document_locks;
CREATE TABLE document_locks (
    document_id BIGINT UNSIGNED PRIMARY KEY COMMENT '文档ID',
    locked_by BIGINT UNSIGNED NOT NULL COMMENT '锁定人ID',
    locked_at DATETIME NOT NULL COMMENT '锁定时间',
    expires_at DATETIME NOT NULL COMMENT '过期时间(30分钟无活动自动释放)',
    force_unlock BOOLEAN DEFAULT FALSE COMMENT '管理员强制解锁标记',

    -- 签出/签入模式
    is_checked_out BOOLEAN DEFAULT FALSE COMMENT '是否被签出(离线编辑)',
    checked_out_at DATETIME COMMENT '签出时间',
    checkout_ip VARCHAR(45) COMMENT '签出IP',

    last_activity DATETIME COMMENT '最后活动时间',

    INDEX idx_expires (expires_at),
    INDEX idx_locked_by (locked_by),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (locked_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档锁表';

-- 文档版本历史表
DROP TABLE IF EXISTS document_versions;
CREATE TABLE document_versions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    document_id BIGINT UNSIGNED NOT NULL COMMENT '文档ID',
    version INT NOT NULL COMMENT '版本号',

    -- 文件信息
    filename VARCHAR(255) NOT NULL COMMENT '文件名',
    filepath VARCHAR(500) NOT NULL COMMENT '文件路径',
    file_hash VARCHAR(64) NOT NULL COMMENT '文件SHA256哈希',
    file_size BIGINT NOT NULL COMMENT '文件大小(字节)',
    mime_type VARCHAR(100) COMMENT 'MIME类型',

    -- 版本信息
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    change_description TEXT COMMENT '变更说明',
    change_type VARCHAR(50) DEFAULT 'manual' COMMENT '变更类型: manual/checkout/auto',

    -- 状态
    is_current BOOLEAN DEFAULT FALSE COMMENT '是否当前版本',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    UNIQUE KEY uk_doc_version (document_id, version),
    INDEX idx_doc_current (document_id, is_current),
    INDEX idx_created_by (created_by),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档版本历史表';

-- 文档全文索引队列
DROP TABLE IF EXISTS document_index_queue;
CREATE TABLE document_index_queue (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    document_id BIGINT UNSIGNED NOT NULL COMMENT '文档ID',
    file_path VARCHAR(500) NOT NULL COMMENT '文件路径',
    mime_type VARCHAR(100) COMMENT 'MIME类型',

    -- 索引状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/processing/completed/failed',
    indexed_at DATETIME COMMENT '索引完成时间',
    error_message TEXT COMMENT '错误信息',
    retry_count INT DEFAULT 0 COMMENT '重试次数',

    -- 索引内容摘要
    content_preview TEXT COMMENT '内容预览(前500字符)',
    word_count INT COMMENT '字数统计',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_document (document_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档全文索引队列';

-- 案件文件夹模板表
DROP TABLE IF EXISTS case_folder_templates;
CREATE TABLE case_folder_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL COMMENT '模板名称',
    description TEXT COMMENT '模板说明',
    folder_structure JSON NOT NULL COMMENT '文件夹结构定义',

    -- 分类
    case_type VARCHAR(50) COMMENT '适用案件类型',
    is_default BOOLEAN DEFAULT FALSE COMMENT '是否默认模板',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',

    -- 模板文件
    template_files JSON COMMENT '模板文件列表',

    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_case_type (case_type),
    INDEX idx_active (is_active),
    FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='案件文件夹模板表';

-- 初始化标准文件夹模板
INSERT INTO case_folder_templates (name, description, folder_structure, case_type, is_default, created_by) VALUES
('标准民商事案件模板', '适用于一般民商事案件', '{
    "folders": [
        {"name": "01_客户证据", "description": "客户提供的证据材料"},
        {"name": "02_法律文书", "description": "起诉状、答辩状、代理词等", "subfolders": [
            {"name": "起诉状", "template": "template_indictment.docx"},
            {"name": "答辩状", "template": "template_answer.docx"},
            {"name": "代理词", "template": "template_opinion.docx"}
        ]},
        {"name": "03_法院传票与通知", "description": "法院送达的各种文书"},
        {"name": "04_研究报告与备忘录", "description": "内部分析材料"},
        {"name": "05_结案材料", "description": "判决书、裁定书、结案报告"}
    ]
}', 'civil', TRUE, 1);

-- ============================================================================
-- 5. 财务管理 (Financial Management)
-- ============================================================================

-- 合同表
DROP TABLE IF EXISTS contracts;
CREATE TABLE contracts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contract_code VARCHAR(50) UNIQUE NOT NULL COMMENT '合同编号',

    -- 关联信息
    case_id BIGINT UNSIGNED COMMENT '关联案件ID',
    client_id BIGINT UNSIGNED NOT NULL COMMENT '客户ID',

    -- 金额信息
    contract_amount DECIMAL(15,2) NOT NULL COMMENT '合同金额',
    currency VARCHAR(10) DEFAULT 'CNY' COMMENT '币种',

    -- 合同条款
    billing_cycle VARCHAR(50) COMMENT '计费周期:一次性/分期/按小时',
    payment_terms VARCHAR(100) COMMENT '付款条件',
    start_date DATE COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',

    -- 合同状态
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态: draft/active/suspended/completed/cancelled',

    -- 补充协议
    parent_contract_id BIGINT UNSIGNED COMMENT '主合同ID(补充协议时使用)',
    contract_type VARCHAR(20) DEFAULT 'original' COMMENT '合同类型: original/supplementary',

    -- 签署信息
    signed_at DATE COMMENT '签署日期',
    document_id BIGINT UNSIGNED COMMENT '合同文档ID',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME DEFAULT NULL,

    INDEX idx_case (case_id),
    INDEX idx_client (client_id),
    INDEX idx_parent (parent_contract_id),
    INDEX idx_status (status),
    INDEX idx_deleted_at (deleted_at),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (case_id) REFERENCES cases(id),
    FOREIGN KEY (parent_contract_id) REFERENCES contracts(id),
    FOREIGN KEY (document_id) REFERENCES documents(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='合同表';

-- 付款计划表
DROP TABLE IF EXISTS payment_milestones;
CREATE TABLE payment_milestones (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    contract_id BIGINT UNSIGNED NOT NULL COMMENT '合同ID',
    name VARCHAR(200) NOT NULL COMMENT '里程碑名称',
    sequence INT NOT NULL COMMENT '顺序号',
    amount DECIMAL(15,2) NOT NULL COMMENT '金额',
    percentage DECIMAL(5,2) COMMENT '占比(%)',
    due_date DATE COMMENT '到期日期',
    condition TEXT COMMENT '付款条件',

    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/billed/paid/overdue',
    invoice_id BIGINT UNSIGNED COMMENT '关联发票ID',
    paid_amount DECIMAL(15,2) DEFAULT 0 COMMENT '已付金额',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_contract (contract_id),
    INDEX idx_invoice (invoice_id),
    INDEX idx_status (status),
    FOREIGN KEY (contract_id) REFERENCES contracts(id),
    FOREIGN KEY (invoice_id) REFERENCES invoices(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='付款计划表';

-- 发票表
DROP TABLE IF EXISTS invoices;
CREATE TABLE invoices (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invoice_code VARCHAR(50) UNIQUE NOT NULL COMMENT '发票编号',

    -- 关联信息
    contract_id BIGINT UNSIGNED COMMENT '合同ID',
    milestone_id BIGINT UNSIGNED COMMENT '付款计划ID',
    client_id BIGINT UNSIGNED NOT NULL COMMENT '客户ID',

    -- 金额信息
    amount DECIMAL(15,2) NOT NULL COMMENT '发票金额(不含税)',
    tax_rate DECIMAL(5,2) DEFAULT 0 COMMENT '税率(%)',
    tax_amount DECIMAL(15,2) COMMENT '税额',
    total_amount DECIMAL(15,2) COMMENT '价税合计',

    -- 客户开票信息快照
    client_name VARCHAR(200) NOT NULL COMMENT '客户名称',
    client_tax_id VARCHAR(50) COMMENT '纳税人识别号',
    client_address TEXT COMMENT '地址',
    client_bank_name VARCHAR(100) COMMENT '开户行',
    client_bank_account VARCHAR(50) COMMENT '银行账号',

    -- 发票类型
    invoice_type VARCHAR(20) DEFAULT 'normal' COMMENT '发票类型: normal/credit(红字)',
    original_invoice_id BIGINT UNSIGNED COMMENT '原发票ID(红冲时)',
    refund_reason TEXT COMMENT '退费原因',
    write_off_amount DECIMAL(15,2) COMMENT '核销金额',

    -- 开票流程
    status VARCHAR(20) DEFAULT 'draft' COMMENT '状态: draft/submitted/approved/issued/received/cancelled',
    submitted_at DATETIME COMMENT '提交时间',
    approved_by_finance_at DATETIME COMMENT '财务复核时间',
    issued_at DATETIME COMMENT '开票时间',
    received_at DATETIME COMMENT '客户签收时间',

    -- 电子发票
    electronic_invoice_url VARCHAR(500) COMMENT '电子发票URL',
    electronic_invoice_code VARCHAR(50) COMMENT '发票代码',
    electronic_invoice_number VARCHAR(50) COMMENT '发票号码',

    -- 审批信息
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    submitted_by BIGINT UNSIGNED COMMENT '提交者ID',
    approved_by BIGINT UNSIGNED COMMENT '审批人ID',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_contract (contract_id),
    INDEX idx_client (client_id),
    INDEX idx_status (status),
    INDEX idx_original (original_invoice_id),
    INDEX idx_created_by (created_by),
    INDEX idx_approved_by (approved_by),
    FOREIGN KEY (contract_id) REFERENCES contracts(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (original_invoice_id) REFERENCES invoices(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (submitted_by) REFERENCES users(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发票表';

-- 回款记录表
DROP TABLE IF EXISTS payments;
CREATE TABLE payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_code VARCHAR(50) UNIQUE NOT NULL COMMENT '回款编号',

    invoice_id BIGINT UNSIGNED NOT NULL COMMENT '发票ID',
    amount DECIMAL(15,2) NOT NULL COMMENT '回款金额',
    payment_date DATE NOT NULL COMMENT '付款日期',

    -- 付款方式
    payment_method VARCHAR(50) COMMENT '付款方式: bank_transfer/cash/other',
    reference_no VARCHAR(100) COMMENT '银行流水号',
    payer_name VARCHAR(200) COMMENT '付款人',
    payer_account VARCHAR(100) COMMENT '付款账号',

    -- 凭证
    attachment_id BIGINT UNSIGNED COMMENT '回款凭证ID',

    -- 确认信息
    confirmed_by BIGINT UNSIGNED NOT NULL COMMENT '确认人ID',
    confirmed_at DATETIME COMMENT '确认时间',
    status VARCHAR(20) DEFAULT 'confirmed' COMMENT '状态: pending/confirmed/rejected',

    remark TEXT COMMENT '备注',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_invoice (invoice_id),
    INDEX idx_date (payment_date),
    INDEX idx_status (status),
    INDEX idx_confirmed_by (confirmed_by),
    FOREIGN KEY (invoice_id) REFERENCES invoices(id),
    FOREIGN KEY (confirmed_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回款记录表';

-- 坏账核销记录表
DROP TABLE IF EXISTS bad_debt_records;
CREATE TABLE bad_debt_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- 关联信息
    contract_id BIGINT UNSIGNED NOT NULL COMMENT '合同ID',
    invoice_id BIGINT UNSIGNED COMMENT '发票ID',

    -- 金额信息
    original_amount DECIMAL(15,2) NOT NULL COMMENT '原始应收金额',
    write_off_amount DECIMAL(15,2) NOT NULL COMMENT '核销金额',
    remaining_amount DECIMAL(15,2) COMMENT '剩余金额',

    -- 核销原因
    reason TEXT NOT NULL COMMENT '核销原因',
    reason_type VARCHAR(50) COMMENT '原因类型: bankruptcy/dispute/uncollectible/other',

    -- 审批流程
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/approved/rejected',
    approved_by BIGINT UNSIGNED COMMENT '审批人ID',
    approved_at DATETIME COMMENT '审批时间',
    approval_notes TEXT COMMENT '审批意见',

    -- 附件
    attachment_ids JSON COMMENT '证明材料ID列表',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_contract (contract_id),
    INDEX idx_status (status),
    INDEX idx_approved_by (approved_by),
    FOREIGN KEY (contract_id) REFERENCES contracts(id),
    FOREIGN KEY (invoice_id) REFERENCES invoices(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='坏账核销记录表';

-- 提成记录表
DROP TABLE IF EXISTS commission_records;
CREATE TABLE commission_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    commission_code VARCHAR(50) UNIQUE NOT NULL COMMENT '提成编号',

    -- 关联信息
    contract_id BIGINT UNSIGNED NOT NULL COMMENT '合同ID',
    payment_id BIGINT UNSIGNED NOT NULL COMMENT '回款ID',
    case_id BIGINT UNSIGNED COMMENT '案件ID',

    -- 受益人信息
    beneficiary_id BIGINT UNSIGNED NOT NULL COMMENT '受益人ID',
    beneficiary_role VARCHAR(50) NOT NULL COMMENT '角色: source/lawyer/assistant',

    -- 金额计算
    payment_amount DECIMAL(15,2) NOT NULL COMMENT '回款金额',
    cost_deduction DECIMAL(15,2) DEFAULT 0 COMMENT '成本扣除',
    commission_base DECIMAL(15,2) NOT NULL COMMENT '提成基数',
    commission_rate DECIMAL(5,2) NOT NULL COMMENT '提成比例(%)',
    commission_amount DECIMAL(15,2) NOT NULL COMMENT '提成金额',

    -- 计算/支付状态
    calculated_at DATETIME COMMENT '计算时间',
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/calculated/paid/cancelled',

    -- 支付信息
    paid_date DATE COMMENT '支付日期',
    payment_voucher VARCHAR(100) COMMENT '支付凭证号',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_contract (contract_id),
    INDEX idx_beneficiary (beneficiary_id, status),
    INDEX idx_payment (payment_id),
    INDEX idx_case (case_id),
    FOREIGN KEY (contract_id) REFERENCES contracts(id),
    FOREIGN KEY (payment_id) REFERENCES payments(id),
    FOREIGN KEY (case_id) REFERENCES cases(id),
    FOREIGN KEY (beneficiary_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='提成记录表';

-- ============================================================================
-- 6. 通知系统 (Notification System)
-- ============================================================================

-- 通知预览队列表
DROP TABLE IF EXISTS notification_queue;
CREATE TABLE notification_queue (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    -- 触发信息
    trigger_type VARCHAR(50) NOT NULL COMMENT '触发类型',
    trigger_id BIGINT UNSIGNED NOT NULL COMMENT '触发记录ID',
    case_id BIGINT UNSIGNED COMMENT '关联案件ID',

    -- 接收人信息
    recipient_type VARCHAR(20) NOT NULL COMMENT '接收人类型: client/lawyer/admin',
    recipient_id BIGINT UNSIGNED NOT NULL COMMENT '接收人ID',
    recipient_name VARCHAR(100) NOT NULL COMMENT '接收人姓名',
    recipient_contact VARCHAR(200) COMMENT '联系方式(邮箱/手机/OpenID)',

    -- 通知内容
    channel VARCHAR(20) NOT NULL COMMENT '通知渠道: email/sms/wechat',
    subject VARCHAR(200) COMMENT '标题(邮件等)',
    content TEXT NOT NULL COMMENT '通知内容',
    template_id VARCHAR(50) COMMENT '模板ID',

    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/approved/sent/cancelled/failed',
    priority VARCHAR(20) DEFAULT 'normal' COMMENT '优先级: urgent/normal/low',

    -- 审核信息
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    approved_by BIGINT UNSIGNED COMMENT '审批人ID',
    approved_at DATETIME COMMENT '审批时间',

    -- 发送信息
    sent_at DATETIME COMMENT '发送时间',
    sent_retry_count INT DEFAULT 0 COMMENT '重试次数',
    error_message TEXT COMMENT '错误信息',

    -- 敏感信息标记
    contains_sensitive_info BOOLEAN DEFAULT FALSE COMMENT '包含敏感信息',
    auto_send BOOLEAN DEFAULT FALSE COMMENT '是否自动发送',

    -- 外部消息ID
    external_message_id VARCHAR(100) COMMENT '外部消息ID(如微信msg_id)',

    INDEX idx_status (status),
    INDEX idx_recipient (recipient_id, status),
    INDEX idx_trigger (trigger_type, trigger_id),
    INDEX idx_created (created_at),
    INDEX idx_created_by (created_by),
    INDEX idx_approved_by (approved_by),
    INDEX idx_case (case_id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (approved_by) REFERENCES users(id),
    FOREIGN KEY (case_id) REFERENCES cases(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知预览队列表';

-- 通知模板表
DROP TABLE IF EXISTS notification_templates;
CREATE TABLE notification_templates (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    template_code VARCHAR(50) UNIQUE NOT NULL COMMENT '模板代码',
    template_name VARCHAR(100) NOT NULL COMMENT '模板名称',

    -- 适用信息
    channel VARCHAR(20) NOT NULL COMMENT '适用渠道: email/sms/wechat',
    recipient_type VARCHAR(20) NOT NULL COMMENT '接收人类型: client/lawyer/admin',
    trigger_event VARCHAR(100) NOT NULL COMMENT '触发事件',

    -- 模板内容
    subject_template VARCHAR(200) COMMENT '标题模板',
    content_template TEXT NOT NULL COMMENT '内容模板(支持变量替换)',

    -- 变量定义
    variables JSON COMMENT '可用变量列表',

    -- 自动发送规则
    auto_send BOOLEAN DEFAULT FALSE COMMENT '是否自动发送',
    requires_approval BOOLEAN DEFAULT TRUE COMMENT '是否需要审批',

    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_channel_event (channel, trigger_event)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知模板表';

-- 初始化通知模板
INSERT INTO notification_templates (template_code, template_name, channel, recipient_type, trigger_event, subject_template, content_template, variables, auto_send, requires_approval) VALUES
('SYSTEM_MAINTENANCE', '系统维护通知', 'email', 'lawyer', 'system_maintenance', '系统维护通知', '尊敬的{name}，系统将于{start_time}至{end_time}进行维护，届时将暂停服务。', '["name", "start_time", "end_time"]', TRUE, FALSE),
('PAYMENT_RECEIVED', '收款确认通知', 'wechat', 'client', 'payment_received', '收款确认', '尊敬的客户，我们已收到您的付款{amount}元，感谢您的配合。', '["amount", "payment_date"]', TRUE, FALSE),
('CASE_HEARING', '开庭提醒', 'wechat', 'client', 'hearing_reminder', '开庭提醒', '尊敬的客户，您的案件"{case_title}"将于{hearing_date}在{court}开庭。', '["case_title", "hearing_date", "court"]', FALSE, TRUE),
('CASE_PROGRESS', '案件进展通知', 'wechat', 'client', 'case_progress', '案件进展', '尊敬的客户，您的案件"{case_title}"有新进展：{progress}。', '["case_title", "progress"]', FALSE, TRUE);

-- ============================================================================
-- 7. 客户门户 (Client Portal)
-- ============================================================================

-- 客户账户表
DROP TABLE IF EXISTS client_accounts;
CREATE TABLE client_accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    client_id BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '客户ID',

    -- 账户信息
    username VARCHAR(50) UNIQUE NOT NULL COMMENT '用户名(手机号)',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    phone VARCHAR(20) UNIQUE COMMENT '手机号',

    -- 状态
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态: active/disabled',
    last_login_at DATETIME COMMENT '最后登录时间',
    last_login_ip VARCHAR(45) COMMENT '最后登录IP',

    -- 微信绑定
    wechat_openid VARCHAR(100) UNIQUE COMMENT '微信OpenID',
    wechat_unionid VARCHAR(100) COMMENT '微信UnionID',
    wechat_nickname VARCHAR(100) COMMENT '微信昵称',
    wechat_bound_at DATETIME COMMENT '绑定时间',

    -- 授权控制(白名单模式)
    authorized_cases JSON COMMENT '授权可见的案件ID列表',

    -- 安全设置
    password_changed_at DATETIME COMMENT '密码修改时间',
    failed_login_count INT DEFAULT 0 COMMENT '失败登录次数',
    locked_until DATETIME COMMENT '锁定至',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_client (client_id),
    INDEX idx_wechat (wechat_openid),
    INDEX idx_phone (phone),
    INDEX idx_status (status),
    FOREIGN KEY (client_id) REFERENCES clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户门户账户表';

-- 微信邀请记录表
DROP TABLE IF EXISTS wechat_invitations;
CREATE TABLE wechat_invitations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    invite_token VARCHAR(100) UNIQUE NOT NULL COMMENT '邀请token',
    client_id BIGINT UNSIGNED NOT NULL COMMENT '客户ID',
    invited_by BIGINT UNSIGNED NOT NULL COMMENT '邀请人ID(律师)',

    -- 授权范围
    authorized_cases JSON COMMENT '授权的案件列表',

    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/accepted/expired/cancelled',

    -- 时间
    expires_at DATETIME NOT NULL COMMENT '过期时间',
    accepted_at DATETIME COMMENT '接受时间',
    wechat_openid VARCHAR(100) COMMENT '绑定的OpenID',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_token (invite_token),
    INDEX idx_client (client_id),
    INDEX idx_status (status),
    INDEX idx_invited_by (invited_by),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (invited_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='微信邀请记录表';

-- ============================================================================
-- 8. 风控核心 (Risk Control Critical Scenarios)
-- ============================================================================

-- 8.1 隔离墙机制 (Ethical Wall)

-- 扩展 cases 表，添加隔离墙字段
DROP PROCEDURE IF EXISTS add_cases_ethical_wall_columns;
DELIMITER //
CREATE PROCEDURE add_cases_ethical_wall_columns()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'cases'
                   AND column_name = 'ethical_wall_enabled') THEN
        ALTER TABLE cases ADD COLUMN ethical_wall_enabled BOOLEAN DEFAULT FALSE COMMENT '是否启用隔离墙' AFTER status;
    END IF;

    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'cases'
                   AND column_name = 'ethical_wall_description') THEN
        ALTER TABLE cases ADD COLUMN ethical_wall_description TEXT COMMENT '隔离墙说明' AFTER ethical_wall_enabled;
    END IF;

    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'cases'
                   AND column_name = 'ethical_wall_enabled_by') THEN
        ALTER TABLE cases ADD COLUMN ethical_wall_enabled_by BIGINT UNSIGNED COMMENT '启用人ID' AFTER ethical_wall_description;
    END IF;

    IF NOT EXISTS (SELECT * FROM information_schema.columns
                   WHERE table_schema = DATABASE()
                   AND table_name = 'cases'
                   AND column_name = 'ethical_wall_enabled_at') THEN
        ALTER TABLE cases ADD COLUMN ethical_wall_enabled_at DATETIME COMMENT '启用时间' AFTER ethical_wall_enabled_by;
    END IF;
END //
DELIMITER ;

CALL add_cases_ethical_wall_columns();
DROP PROCEDURE add_cases_ethical_wall_columns;

-- 添加索引（如果不存在）
DROP PROCEDURE IF EXISTS add_cases_ethical_wall_indexes;
DELIMITER //
CREATE PROCEDURE add_cases_ethical_wall_indexes()
BEGIN
    IF NOT EXISTS (SELECT * FROM information_schema.statistics
                   WHERE table_schema = DATABASE()
                   AND table_name = 'cases'
                   AND index_name = 'idx_ethical_wall') THEN
        ALTER TABLE cases ADD INDEX idx_ethical_wall (ethical_wall_enabled);
    END IF;
END //
DELIMITER ;

CALL add_cases_ethical_wall_indexes();
DROP PROCEDURE add_cases_ethical_wall_indexes;

-- 隔离墙白名单表
DROP TABLE IF EXISTS case_ethical_wall_whitelist;
CREATE TABLE case_ethical_wall_whitelist (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    case_id BIGINT UNSIGNED NOT NULL COMMENT '案件ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    granted_by BIGINT UNSIGNED NOT NULL COMMENT '授权人ID',
    granted_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
    reason TEXT COMMENT '授权理由',

    UNIQUE KEY uk_case_user (case_id, user_id),
    INDEX idx_case (case_id),
    INDEX idx_user (user_id),
    INDEX idx_granted_by (granted_by),
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (granted_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='隔离墙白名单表';

-- 隔离墙访问日志表
DROP TABLE IF EXISTS ethical_wall_access_logs;
CREATE TABLE ethical_wall_access_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    case_id BIGINT UNSIGNED NOT NULL COMMENT '案件ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    access_type VARCHAR(20) NOT NULL COMMENT '访问类型: view/search/export',
    access_result VARCHAR(20) NOT NULL COMMENT '访问结果: allowed/denied',
    ip_address VARCHAR(45) COMMENT 'IP地址',
    user_agent TEXT COMMENT 'User-Agent',
    attempted_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '尝试时间',

    INDEX idx_case (case_id),
    INDEX idx_user (user_id),
    INDEX idx_result (access_result),
    INDEX idx_attempted_at (attempted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='隔离墙访问日志表';

-- 8.2 客户代管款管理 (Client Trust Funds)

-- 客户代管账户表
DROP TABLE IF EXISTS client_trust_accounts;
CREATE TABLE client_trust_accounts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    client_id BIGINT UNSIGNED NOT NULL COMMENT '客户ID',
    account_code VARCHAR(50) UNIQUE NOT NULL COMMENT '代管账户编号',

    -- 余额信息
    balance DECIMAL(15,2) DEFAULT 0 COMMENT '当前余额',
    currency VARCHAR(10) DEFAULT 'CNY' COMMENT '币种',
    frozen_amount DECIMAL(15,2) DEFAULT 0 COMMENT '冻结金额',

    -- 资金用途限制
    purpose_restriction VARCHAR(200) COMMENT '资金用途说明',
    authorized_uses JSON COMMENT '授权用途列表',

    -- 状态
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态: active/frozen/closed',
    opened_at DATETIME COMMENT '开户时间',
    closed_at DATETIME COMMENT '销户时间',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_client (client_id),
    INDEX idx_code (account_code),
    INDEX idx_status (status),
    FOREIGN KEY (client_id) REFERENCES clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户代管账户表';

-- 代管款交易记录表
DROP TABLE IF EXISTS client_trust_transactions;
CREATE TABLE client_trust_transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account_id BIGINT UNSIGNED NOT NULL COMMENT '账户ID',
    transaction_code VARCHAR(50) UNIQUE NOT NULL COMMENT '交易编号',

    -- 交易信息
    transaction_type VARCHAR(20) NOT NULL COMMENT '交易类型: deposit/deposit_refund/withdraw/transfer',
    amount DECIMAL(15,2) NOT NULL COMMENT '金额',
    description TEXT NOT NULL COMMENT '交易说明',

    -- 用途关联
    case_id BIGINT UNSIGNED COMMENT '关联案件ID',
    purpose_code VARCHAR(50) COMMENT '用途代码: court_fee/evidence_fee/investigation_fee等',

    -- 支出信息
    recipient_name VARCHAR(200) COMMENT '收款方名称',
    recipient_bank_account VARCHAR(50) COMMENT '收款账号',
    recipient_bank_name VARCHAR(100) COMMENT '收款银行',

    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/completed/cancelled',
    completed_at DATETIME COMMENT '完成时间',
    attachment_id BIGINT UNSIGNED COMMENT '凭证附件ID',

    -- 审计信息
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    approved_by BIGINT UNSIGNED COMMENT '审批人ID',
    approved_at DATETIME COMMENT '审批时间',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_account (account_id),
    INDEX idx_case (case_id),
    INDEX idx_status (status),
    INDEX idx_type (transaction_type),
    INDEX idx_created_by (created_by),
    INDEX idx_approved_by (approved_by),
    FOREIGN KEY (account_id) REFERENCES client_trust_accounts(id),
    FOREIGN KEY (case_id) REFERENCES cases(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代管款交易记录表';

-- 8.3 离职交接机制 (Offboarding)

-- 离职交接记录表
DROP TABLE IF EXISTS offboarding_records;
CREATE TABLE offboarding_records (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '离职用户ID',
    initiated_by BIGINT UNSIGNED NOT NULL COMMENT '发起人ID',
    initiated_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '发起时间',

    -- 案件移交
    original_cases JSON NOT NULL COMMENT '原主办案件列表',
    new_lawyer_id BIGINT UNSIGNED COMMENT '接收律师ID',
    case_transfer_completed_at DATETIME COMMENT '案件移交完成时间',

    -- 待办事项移交
    original_inbox_items JSON NOT NULL COMMENT '原待办事项列表',
    inbox_transfer_completed_at DATETIME COMMENT '待办移交完成时间',

    -- 文档处理
    document_disposal_method VARCHAR(50) COMMENT '文档处理方式: delete/transfer/revoke_access',
    document_disposal_completed_at DATETIME COMMENT '文档处理完成时间',

    -- 财务结算
    settlement_calculated BOOLEAN DEFAULT FALSE COMMENT '是否计算提成',
    settlement_amount DECIMAL(15,2) COMMENT '结算金额',
    settlement_paid BOOLEAN DEFAULT FALSE COMMENT '是否已支付',

    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/in_progress/completed/cancelled',
    notes TEXT COMMENT '备注',

    completed_at DATETIME COMMENT '完成时间',

    INDEX idx_user (user_id),
    INDEX idx_initiated_by (initiated_by),
    INDEX idx_new_lawyer (new_lawyer_id),
    INDEX idx_status (status),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (initiated_by) REFERENCES users(id),
    FOREIGN KEY (new_lawyer_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='离职交接记录表';

-- 离职交接详情表(记录每项移交)
DROP TABLE IF EXISTS offboarding_transfer_details;
CREATE TABLE offboarding_transfer_details (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    offboarding_id BIGINT UNSIGNED NOT NULL COMMENT '交接记录ID',
    transfer_type VARCHAR(50) NOT NULL COMMENT '移交类型: case/inbox/document',

    -- 原所有者
    original_owner_id BIGINT UNSIGNED NOT NULL COMMENT '原所有者ID',

    -- 新所有者
    new_owner_id BIGINT UNSIGNED COMMENT '新所有者ID',

    -- 移交内容
    item_id BIGINT UNSIGNED NOT NULL COMMENT '项目ID(案件ID/待办ID等)',
    item_name VARCHAR(200) NOT NULL COMMENT '项目名称',

    -- 状态
    transfer_status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/completed/failed',
    transferred_at DATETIME COMMENT '移交时间',
    error_message TEXT COMMENT '错误信息',

    INDEX idx_offboarding (offboarding_id),
    INDEX idx_status (transfer_status),
    INDEX idx_original_owner (original_owner_id),
    INDEX idx_new_owner (new_owner_id),
    FOREIGN KEY (offboarding_id) REFERENCES offboarding_records(id),
    FOREIGN KEY (original_owner_id) REFERENCES users(id),
    FOREIGN KEY (new_owner_id) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='离职交接详情表';

-- 令牌撤销记录表
DROP TABLE IF EXISTS token_revocation_logs;
CREATE TABLE token_revocation_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    revocation_type VARCHAR(50) NOT NULL COMMENT '撤销类型: offboarding/password_reset/manual',
    revoked_by BIGINT UNSIGNED COMMENT '撤销操作人ID',

    -- 撤销范围
    revoke_all BOOLEAN DEFAULT TRUE COMMENT '是否撤销所有令牌',
    revoked_tokens JSON COMMENT '撤销的令牌列表',

    -- 令牌信息
    token_type VARCHAR(50) COMMENT '令牌类型: access_token/refresh_token/api_key',
    revoked_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '撤销时间',
    ip_address VARCHAR(45) COMMENT '操作IP',

    INDEX idx_user (user_id),
    INDEX idx_type (revocation_type),
    INDEX idx_revoked_by (revoked_by),
    INDEX idx_revoked_at (revoked_at),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (revoked_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='令牌撤销记录表';

-- ============================================================================
-- 9. 数据迁移支持 (Migration Support)
-- ============================================================================

-- 数据导入任务表
DROP TABLE IF EXISTS data_import_tasks;
CREATE TABLE data_import_tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    task_code VARCHAR(50) UNIQUE NOT NULL COMMENT '任务编号',

    -- 导入信息
    import_type VARCHAR(50) NOT NULL COMMENT '导入类型: client/case/contract/document',
    file_path VARCHAR(500) COMMENT '导入文件路径',
    total_rows INT DEFAULT 0 COMMENT '总行数',

    -- 进度
    processed_rows INT DEFAULT 0 COMMENT '已处理行数',
    success_rows INT DEFAULT 0 COMMENT '成功行数',
    failed_rows INT DEFAULT 0 COMMENT '失败行数',

    -- 状态
    status VARCHAR(20) DEFAULT 'pending' COMMENT '状态: pending/processing/completed/failed',

    -- 结果
    error_summary JSON COMMENT '错误汇总',
    result_summary JSON COMMENT '结果汇总',

    -- 操作信息
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建者ID',
    started_at DATETIME COMMENT '开始时间',
    completed_at DATETIME COMMENT '完成时间',

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_status (status),
    INDEX idx_type (import_type),
    INDEX idx_created_by (created_by),
    FOREIGN KEY (created_by) REFERENCES users(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据导入任务表';

-- 数据导入错误明细表
DROP TABLE IF EXISTS data_import_errors;
CREATE TABLE data_import_errors (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    import_task_id BIGINT UNSIGNED NOT NULL COMMENT '导入任务ID',
    row_number INT NOT NULL COMMENT '行号',
    row_data JSON COMMENT '行数据',
    error_message TEXT NOT NULL COMMENT '错误信息',
    error_type VARCHAR(50) COMMENT '错误类型',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_task (import_task_id),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (import_task_id) REFERENCES data_import_tasks(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据导入错误明细表';

-- ============================================================================
-- 10. 系统配置与日志
-- ============================================================================

-- 系统配置表
DROP TABLE IF EXISTS system_configs;
CREATE TABLE system_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    config_type VARCHAR(20) DEFAULT 'string' COMMENT '值类型: string/json/number/boolean',
    description TEXT COMMENT '配置说明',
    is_system BOOLEAN DEFAULT FALSE COMMENT '是否系统配置',
    sort INT DEFAULT 0 COMMENT '排序',
    status VARCHAR(20) DEFAULT 'active' COMMENT '状态',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_key (config_key),
    INDEX idx_status (status),
    INDEX idx_is_system (is_system)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 初始化系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description, is_system, sort) VALUES
('conflict.api_provider', 'qichacha', 'string', '工商API提供商', TRUE, 1),
('conflict.auto_scan_enabled', 'true', 'boolean', '是否启用自动冲突扫描', TRUE, 2),
('conflict.scan_schedule', '0 2 * * *', 'string', '扫描任务Cron表达式', TRUE, 3),
('notification.wechat.app_id', '', 'string', '微信服务号AppID', TRUE, 4),
('notification.wechat.app_secret', '', 'string', '微信服务号AppSecret', TRUE, 5),
('document.onlyoffice_url', '', 'string', 'OnlyOffice服务地址', TRUE, 6),
('document.max_file_size', '104857600', 'number', '文档最大大小(字节)', TRUE, 7),
('inbox.escalation_enabled', 'true', 'boolean', '是否启用待办升级', TRUE, 8),
('inbox.escalation_days', '3', 'number', '升级触发天数', TRUE, 9),
('offboarding.auto_revoke_tokens', 'true', 'boolean', '离职自动撤销令牌', TRUE, 10),
('system.timezone', 'Asia/Shanghai', 'string', '系统时区', TRUE, 11),
('system.date_format', '2006-01-02', 'string', '日期格式', TRUE, 12),
('system.datetime_format', '2006-01-02 15:04:05', 'string', '日期时间格式', TRUE, 13);

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================================================
-- 结束
-- ============================================================================
