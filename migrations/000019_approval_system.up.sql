-- 通用审批系统数据库表结构
-- 支持各种类型的审批申请（请假、报销、立项等）

-- 1. 审批申请表
CREATE TABLE IF NOT EXISTS approval_requests (
    id VARCHAR(36) PRIMARY KEY COMMENT '申请ID',
    request_number VARCHAR(50) NOT NULL UNIQUE COMMENT '申请编号',

    -- 申请基本信息
    title VARCHAR(255) NOT NULL COMMENT '申请标题',
    type VARCHAR(50) NOT NULL COMMENT '申请类型',
    category VARCHAR(100) COMMENT '申请分类',
    content TEXT NOT NULL COMMENT '申请内容',

    -- 申请人信息
    applicant_id VARCHAR(36) NOT NULL COMMENT '申请人ID',
    applicant_name VARCHAR(255) NOT NULL COMMENT '申请人姓名',
    applicant_title VARCHAR(100) COMMENT '申请人职位',
    department_id VARCHAR(36) COMMENT '部门ID',
    department_name VARCHAR(255) COMMENT '部门名称',

    -- 紧急程度
    urgency ENUM('normal', 'urgent', 'very_urgent') DEFAULT 'normal' COMMENT '紧急程度',
    priority ENUM('low', 'medium', 'high', 'critical') DEFAULT 'medium' COMMENT '优先级',

    -- 时间相关
    expected_effective_date TIMESTAMP COMMENT '期望生效时间',
    expected_expiry_date TIMESTAMP COMMENT '期望到期时间',
    duration_days INT COMMENT '持续天数',

    -- 申请状态
    status ENUM('draft', 'submitted', 'under_review', 'approved', 'rejected', 'cancelled', 'expired') DEFAULT 'draft' COMMENT '状态',
    submission_date TIMESTAMP COMMENT '提交时间',

    -- 当前审批信息
    current_stage VARCHAR(100) COMMENT '当前审批阶段',
    current_approver_id VARCHAR(36) COMMENT '当前审批人ID',
    current_approver_name VARCHAR(255) COMMENT '当前审批人姓名',

    -- 审批流程配置
    workflow_type VARCHAR(100) NOT NULL COMMENT '工作流类型',
    workflow_config JSON COMMENT '工作流配置',

    -- 附加信息
    attachments JSON COMMENT '附件列表',
    metadata JSON COMMENT '扩展元数据',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP COMMENT '删除时间',

    -- 索引
    INDEX idx_request_number (request_number),
    INDEX idx_applicant_id (applicant_id),
    INDEX idx_department_id (department_id),
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_submission_date (submission_date),
    INDEX idx_current_stage (current_stage),
    INDEX idx_current_approver_id (current_approver_id),
    INDEX idx_workflow_type (workflow_type),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_approval_request (title, content, applicant_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批申请表';

-- 2. 审批工作流定义表
CREATE TABLE IF NOT EXISTS approval_workflows (
    id VARCHAR(36) PRIMARY KEY COMMENT '工作流ID',
    workflow_code VARCHAR(100) NOT NULL UNIQUE COMMENT '工作流代码',
    workflow_name VARCHAR(255) NOT NULL COMMENT '工作流名称',
    workflow_type VARCHAR(100) NOT NULL COMMENT '工作流类型',

    -- 适用范围
    applicable_types JSON COMMENT '适用的申请类型',
    applicable_departments JSON COMMENT '适用的部门',
    applicable_roles JSON COMMENT '适用的角色',

    -- 工作流配置
    stages JSON NOT NULL COMMENT '审批阶段配置',
    conditions JSON COMMENT '触发条件',
    timeouts JSON COMMENT '超时配置',

    -- 权限配置
    permissions JSON COMMENT '权限配置',
    notifications JSON COMMENT '通知配置',

    -- 状态和版本
    status ENUM('active', 'inactive', 'deprecated') DEFAULT 'active' COMMENT '状态',
    version INT DEFAULT 1 COMMENT '版本号',
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_workflow_code (workflow_code),
    INDEX idx_workflow_type (workflow_type),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批工作流定义表';

-- 3. 审批记录表
CREATE TABLE IF NOT EXISTS approval_records (
    id VARCHAR(36) PRIMARY KEY COMMENT '审批记录ID',
    approval_request_id VARCHAR(36) NOT NULL COMMENT '审批申请ID',

    -- 审批基本信息
    stage VARCHAR(100) NOT NULL COMMENT '审批阶段',
    stage_order INT NOT NULL COMMENT '阶段顺序',
    approver_id VARCHAR(36) NOT NULL COMMENT '审批人ID',
    approver_name VARCHAR(255) NOT NULL COMMENT '审批人姓名',
    approver_title VARCHAR(100) COMMENT '审批人职位',
    approver_role VARCHAR(100) COMMENT '审批人角色',

    -- 审批决定
    decision ENUM('approve', 'reject', 'request_changes', 'defer', 'escalate', 'reassign') NOT NULL COMMENT '审批决定',
    decision_reason TEXT NOT NULL COMMENT '审批理由',
    decision_comments TEXT COMMENT '审批意见',

    -- 审批条件
    approved_conditions JSON COMMENT '批准的条件',
    imposed_requirements JSON COMMENT '施加的要求',
    follow_up_actions JSON COMMENT '后续行动',

    -- 时间信息
    approval_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '审批时间',
    effective_date TIMESTAMP COMMENT '生效日期',
    next_review_date TIMESTAMP COMMENT '下次审查日期',

    -- 附件和证据
    supporting_documents JSON COMMENT '支持文件',
    evidence_references JSON COMMENT '证据引用',

    -- 状态
    status ENUM('active', 'superseded', 'cancelled') DEFAULT 'active' COMMENT '状态',

    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_approval_request_id (approval_request_id),
    INDEX idx_approver_id (approver_id),
    INDEX idx_stage (stage),
    INDEX idx_stage_order (stage_order),
    INDEX idx_decision (decision),
    INDEX idx_approval_date (approval_date),
    INDEX idx_status (status),

    -- 外键约束
    FOREIGN KEY (approval_request_id) REFERENCES approval_requests(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批记录表';

-- 4. 审批模板表
CREATE TABLE IF NOT EXISTS approval_templates (
    id VARCHAR(36) PRIMARY KEY COMMENT '模板ID',
    template_code VARCHAR(100) NOT NULL UNIQUE COMMENT '模板代码',
    template_name VARCHAR(255) NOT NULL COMMENT '模板名称',
    template_type VARCHAR(100) NOT NULL COMMENT '模板类型',

    -- 模板分类
    category VARCHAR(100) COMMENT '模板分类',
    workflow_type VARCHAR(100) NOT NULL COMMENT '对应工作流类型',

    -- 模板内容
    template_content JSON NOT NULL COMMENT '模板内容',
    form_schema JSON COMMENT '表单结构',
    validation_rules JSON COMMENT '验证规则',

    -- 默认配置
    default_values JSON COMMENT '默认值',
    required_fields JSON COMMENT '必填字段',
    optional_fields JSON COMMENT '可选字段',

    -- 适用条件
    applicable_scenarios JSON COMMENT '适用场景',
    applicable_roles JSON COMMENT '适用角色',
    applicable_departments JSON COMMENT '适用部门',

    -- 状态和版本
    status ENUM('active', 'inactive', 'under_review') DEFAULT 'active' COMMENT '状态',
    version INT DEFAULT 1 COMMENT '版本号',
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 使用统计
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    last_used_date TIMESTAMP COMMENT '最后使用日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    approved_by VARCHAR(36) COMMENT '审批人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_template_code (template_code),
    INDEX idx_template_type (template_type),
    INDEX idx_workflow_type (workflow_type),
    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),
    INDEX idx_usage_count (usage_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批模板表';

-- 5. 审批通知记录表
CREATE TABLE IF NOT EXISTS approval_notifications (
    id VARCHAR(36) PRIMARY KEY COMMENT '通知记录ID',
    approval_request_id VARCHAR(36) NOT NULL COMMENT '审批申请ID',
    approval_record_id VARCHAR(36) COMMENT '审批记录ID',

    -- 通知信息
    notification_type ENUM('submission', 'approval', 'rejection', 'reminder', 'escalation', 'completion') NOT NULL COMMENT '通知类型',
    recipient_id VARCHAR(36) NOT NULL COMMENT '接收人ID',
    recipient_name VARCHAR(255) NOT NULL COMMENT '接收人姓名',
    recipient_email VARCHAR(255) COMMENT '接收人邮箱',

    -- 通知内容
    subject VARCHAR(500) NOT NULL COMMENT '通知主题',
    content TEXT NOT NULL COMMENT '通知内容',

    -- 发送信息
    send_method ENUM('email', 'sms', 'system', 'wechat') NOT NULL COMMENT '发送方式',
    send_status ENUM('pending', 'sent', 'failed', 'cancelled') DEFAULT 'pending' COMMENT '发送状态',
    send_attempts INT DEFAULT 0 COMMENT '发送尝试次数',

    -- 时间信息
    scheduled_at TIMESTAMP COMMENT '计划发送时间',
    sent_at TIMESTAMP COMMENT '实际发送时间',

    -- 响应信息
    read_at TIMESTAMP COMMENT '阅读时间',
    response_action VARCHAR(100) COMMENT '响应动作',

    -- 错误信息
    error_message TEXT COMMENT '错误信息',

    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_approval_request_id (approval_request_id),
    INDEX idx_approval_record_id (approval_record_id),
    INDEX idx_recipient_id (recipient_id),
    INDEX idx_notification_type (notification_type),
    INDEX idx_send_status (send_status),
    INDEX idx_scheduled_at (scheduled_at),
    INDEX idx_sent_at (sent_at),
    INDEX idx_created_at (created_at),

    -- 外键约束
    FOREIGN KEY (approval_request_id) REFERENCES approval_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (approval_record_id) REFERENCES approval_records(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审批通知记录表';

-- 6. 创建视图：审批申请完整信息视图
CREATE OR REPLACE VIEW approval_requests_complete AS
SELECT
    ar.*,
    -- 申请编号格式化
    CONCAT('AP-', DATE_FORMAT(ar.created_at, '%Y%m%d'), '-', LPAD(ar.id, 8, '0')) as formatted_request_number,

    -- 当前审批状态
    COALESCE(approver.decision, 'PENDING') as current_decision,
    COALESCE(approver.decision_reason, '等待审批') as current_decision_reason,
    COALESCE(approver.approval_date, NULL) as current_approval_date,

    -- 审批进度统计
    (SELECT COUNT(*) FROM approval_records ar2 WHERE ar2.approval_request_id = ar.id) as total_records,
    (SELECT COUNT(*) FROM approval_records ar2 WHERE ar2.approval_request_id = ar.id AND ar2.decision = 'approve') as approve_count,
    (SELECT COUNT(*) FROM approval_records ar2 WHERE ar2.approval_request_id = ar.id AND ar2.decision = 'reject') as reject_count,

    -- 最近通知状态
    (SELECT send_status FROM approval_notifications an
     WHERE an.approval_request_id = ar.id
     ORDER BY an.created_at DESC LIMIT 1) as latest_notification_status,

    -- 附件数量
    COALESCE(JSON_LENGTH(ar.attachments), 0) as attachment_count,

    -- 是否即将到期（7天内）
    CASE
        WHEN ar.expected_expiry_date IS NOT NULL AND DATEDIFF(ar.expected_expiry_date, NOW()) BETWEEN 0 AND 7 THEN TRUE
        ELSE FALSE
    END as expiring_soon,

    -- 是否已过期
    CASE
        WHEN ar.expected_expiry_date IS NOT NULL AND ar.expected_expiry_date < NOW() THEN TRUE
        ELSE FALSE
    END as expired,

    -- 处理时长
    CASE
        WHEN ar.submission_date IS NOT NULL THEN
            TIMESTAMPDIFF(HOUR, ar.submission_date, COALESCE(approver.approval_date, NOW()))
        ELSE NULL
    END as processing_hours

FROM approval_requests ar
LEFT JOIN approval_records approver ON ar.id = approver.approval_request_id
    AND approver.id = (
        SELECT MAX(id) FROM approval_records
        WHERE approval_request_id = ar.id
    )
WHERE ar.deleted_at IS NULL;

-- 7. 创建存储过程：生成申请编号
DELIMITER $$
CREATE PROCEDURE IF NOT EXISTS generate_approval_request_number()
BEGIN
    DECLARE current_date DATE;
    DECLARE date_prefix VARCHAR(8);
    DECLARE sequence_number INT;
    DECLARE request_number VARCHAR(50);

    SET current_date = CURDATE();
    SET date_prefix = DATE_FORMAT(current_date, '%Y%m%d');

    -- 获取当天的序列号
    SELECT COALESCE(MAX(CAST(SUBSTRING(request_number, 12) AS UNSIGNED)), 0) + 1
    INTO sequence_number
    FROM approval_requests
    WHERE DATE(created_at) = current_date;

    -- 生成申请编号
    SET request_number = CONCAT('AP-', date_prefix, '-', LPAD(sequence_number, 6, '0'));

    -- 返回申请编号
    SELECT request_number as next_request_number;
END$$
DELIMITER ;

-- 8. 创建触发器：更新模板使用统计
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS increment_approval_template_usage
    AFTER INSERT ON approval_requests
    FOR EACH ROW
BEGIN
    -- 如果申请使用了模板，更新使用统计
    IF NEW.metadata IS NOT NULL AND JSON_EXTRACT(NEW.metadata, '$.template_id') IS NOT NULL THEN
        UPDATE approval_templates
        SET usage_count = usage_count + 1,
            last_used_date = NOW()
        WHERE id = JSON_UNQUOTE(JSON_EXTRACT(NEW.metadata, '$.template_id'));
    END IF;
END$$
DELIMITER ;

-- 9. 插入默认工作流模板
INSERT INTO approval_workflows (
    id, workflow_code, workflow_name, workflow_type,
    applicable_types, stages, status, effective_date, created_by
) VALUES
-- 标准审批工作流
(UUID(), 'STANDARD_APPROVAL', '标准审批流程', 'standard',
JSON_ARRAY('请假申请', '报销申请', '采购申请', '立项申请'),
JSON_ARRAY(
    JSON_OBJECT(
        'stage_name', '部门主管审批',
        'stage_order', 1,
        'approver_role', 'DEPARTMENT_HEAD',
        'required', true,
        'timeout_hours', 48
    ),
    JSON_OBJECT(
        'stage_name', '分管领导审批',
        'stage_order', 2,
        'approver_role', 'MANAGEMENT',
        'required', true,
        'timeout_hours', 72
    )
),
'active', CURDATE(), 'system'),

-- 快速审批工作流
(UUID(), 'QUICK_APPROVAL', '快速审批流程', 'quick',
JSON_ARRAY('用章申请', '开票申请'),
JSON_ARRAY(
    JSON_OBJECT(
        'stage_name', '直接主管审批',
        'stage_order', 1,
        'approver_role', 'SUPERVISOR',
        'required', true,
        'timeout_hours', 24
    )
),
'active', CURDATE(), 'system'),

-- 紧急审批工作流
(UUID(), 'URGENT_APPROVAL', '紧急审批流程', 'urgent',
JSON_ARRAY('紧急采购', '紧急请假'),
JSON_ARRAY(
    JSON_OBJECT(
        'stage_name', '紧急审批',
        'stage_order', 1,
        'approver_role', 'EMERGENCY_APPROVER',
        'required', true,
        'timeout_hours', 4
    )
),
'active', CURDATE(), 'system');

-- 10. 插入默认审批模板
INSERT INTO approval_templates (
    id, template_code, template_name, template_type, category, workflow_type,
    template_content, form_schema, default_values, status, effective_date, created_by
) VALUES
-- 请假申请模板
(UUID(), 'LEAVE_REQUEST', '请假申请模板', 'leave', '人事行政', 'STANDARD_APPROVAL',
JSON_OBJECT(
    'title_template', '{{applicant_name}}的{{leave_type}}申请',
    'content_template', '因{{reason}}需请假{{duration}}天，时间从{{start_date}}到{{end_date}}',
    'required_attachments', JSON_ARRAY()
),
JSON_OBJECT(
    'fields', JSON_ARRAY(
        JSON_OBJECT('name', 'leave_type', 'type', 'select', 'label', '请假类型', 'required', true, 'options', JSON_ARRAY('年假', '病假', '事假', '婚假', '产假')),
        JSON_OBJECT('name', 'reason', 'type', 'textarea', 'label', '请假原因', 'required', true),
        JSON_OBJECT('name', 'start_date', 'type', 'date', 'label', '开始日期', 'required', true),
        JSON_OBJECT('name', 'end_date', 'type', 'date', 'label', '结束日期', 'required', true)
    )
),
JSON_OBJECT('urgency', 'normal'),
'active', CURDATE(), 'system'),

-- 报销申请模板
(UUID(), 'EXPENSE_REIMBURSEMENT', '费用报销模板', 'expense', '财务', 'STANDARD_APPROVAL',
JSON_OBJECT(
    'title_template', '{{applicant_name}}的费用报销申请',
    'content_template', '报销项目：{{expense_items}}，总金额：{{total_amount}}元',
    'required_attachments', JSON_ARRAY('发票', '报销单')
),
JSON_OBJECT(
    'fields', JSON_ARRAY(
        JSON_OBJECT('name', 'expense_type', 'type', 'select', 'label', '费用类型', 'required', true, 'options', JSON_ARRAY('差旅费', '办公费', '招待费', '培训费', '其他')),
        JSON_OBJECT('name', 'amount', 'type', 'number', 'label', '报销金额', 'required', true),
        JSON_OBJECT('name', 'expense_description', 'type', 'textarea', 'label', '费用说明', 'required', true)
    )
),
JSON_OBJECT('urgency', 'normal'),
'active', CURDATE(), 'system'),

-- 立项申请模板
(UUID(), 'PROJECT_APPROVAL', '立项申请模板', 'project', '业务', 'STANDARD_APPROVAL',
JSON_OBJECT(
    'title_template', '{{project_name}}立项申请',
    'content_template', '项目描述：{{project_description}}，预算：{{budget}}元，周期：{{duration}}天',
    'required_attachments', JSON_ARRAY('项目计划书', '预算表')
),
JSON_OBJECT(
    'fields', JSON_ARRAY(
        JSON_OBJECT('name', 'project_name', 'type', 'text', 'label', '项目名称', 'required', true),
        JSON_OBJECT('name', 'project_description', 'type', 'textarea', 'label', '项目描述', 'required', true),
        JSON_OBJECT('name', 'budget', 'type', 'number', 'label', '项目预算', 'required', true),
        JSON_OBJECT('name', 'duration', 'type', 'number', 'label', '项目周期（天）', 'required', true)
    )
),
JSON_OBJECT('urgency', 'normal'),
'active', CURDATE(), 'system');