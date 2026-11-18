-- 豁免管理数据模型
-- 知情同意与豁免管理系统

-- 1. 豁免申请表
CREATE TABLE IF NOT EXISTS waiver_applications (
    id VARCHAR(36) PRIMARY KEY COMMENT '申请ID',
    application_number VARCHAR(50) NOT NULL UNIQUE COMMENT '申请编号',

    -- 关联信息
    conflict_check_id VARCHAR(36) NOT NULL COMMENT '关联的冲突检查ID',
    case_id VARCHAR(36) COMMENT '关联的案件ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    lawyer_id VARCHAR(36) NOT NULL COMMENT '律师ID',

    -- 申请类型
    waiver_type ENUM('INFORMED_CONSENT', 'ETHICAL_BARRIER', 'INFORMATION_SCREEN', 'STRUCTURAL_BARRIER') NOT NULL COMMENT '豁免类型',
    waiver_category ENUM('CLIENT_CONSENT', 'BARRER_IMPLEMENTATION', 'MONITORING_ARRANGEMENT', 'SPECIAL_CIRCUMSTANCES') NOT NULL COMMENT '豁免类别',

    -- 冲突详情
    conflict_summary TEXT NOT NULL COMMENT '冲突情况摘要',
    conflicts JSON NOT NULL COMMENT '冲突详情列表',
    risk_assessment JSON NOT NULL COMMENT '风险评估结果',

    -- 豁免条件
    proposed_conditions JSON NOT NULL COMMENT '建议的豁免条件',
    limitations JSON COMMENT '限制条件',
    monitoring_requirements JSON COMMENT '监控要求',
    reporting_requirements JSON COMMENT '报告要求',

    -- 时间管理
    requested_effective_date DATE NOT NULL COMMENT '申请生效日期',
    requested_expiry_date DATE COMMENT '申请到期日期',
    duration_days INT COMMENT '豁免期限（天）',

    -- 申请理由
    rationale TEXT NOT NULL COMMENT '申请理由',
    supporting_evidence JSON COMMENT '支持证据',
    alternatives_considered JSON COMMENT '考虑的其他方案',

    -- 客户信息
    client_representative_name VARCHAR(255) NOT NULL COMMENT '客户代表姓名',
    client_representative_title VARCHAR(100) COMMENT '客户代表职位',
    client_representative_contact VARCHAR(255) COMMENT '客户代表联系方式',

    -- 律师信息
    requesting_lawyer_name VARCHAR(255) NOT NULL COMMENT '申请律师姓名',
    requesting_lawyer_title VARCHAR(100) COMMENT '申请律师职位',
    supervising_lawyer_name VARCHAR(255) COMMENT '监督律师姓名',

    -- 状态管理
    status ENUM('DRAFT', 'SUBMITTED', 'UNDER_REVIEW', 'REVIEW_COMPLETED', 'APPROVED', 'REJECTED', 'EXPIRED', 'REVOKED') DEFAULT 'DRAFT' COMMENT '状态',
    submission_date TIMESTAMP COMMENT '提交日期',
    review_priority ENUM('LOW', 'MEDIUM', 'HIGH', 'URGENT') DEFAULT 'MEDIUM' COMMENT '审核优先级',

    -- 审批流程
    current_stage ENUM('INITIAL_REVIEW', 'DEPARTMENT_REVIEW', 'COMPLIANCE_REVIEW', 'MANAGEMENT_APPROVAL', 'FINAL_APPROVAL') COMMENT '当前阶段',
    assigned_reviewer VARCHAR(36) COMMENT '分配的审核人',
    review_deadline TIMESTAMP COMMENT '审核截止日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_application_number (application_number),
    INDEX idx_conflict_check_id (conflict_check_id),
    INDEX idx_case_id (case_id),
    INDEX idx_client_id (client_id),
    INDEX idx_lawyer_id (lawyer_id),
    INDEX idx_waiver_type (waiver_type),
    INDEX idx_status (status),
    INDEX idx_submission_date (submission_date),
    INDEX idx_review_priority (review_priority),
    INDEX idx_current_stage (current_stage),
    INDEX idx_assigned_reviewer (assigned_reviewer),
    INDEX idx_created_at (created_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_waiver_application (conflict_summary, rationale, client_representative_name, requesting_lawyer_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='豁免申请表';

-- 2. 豁免审批记录表
CREATE TABLE IF NOT EXISTS waiver_approval_records (
    id VARCHAR(36) PRIMARY KEY COMMENT '审批记录ID',
    waiver_application_id VARCHAR(36) NOT NULL COMMENT '豁免申请ID',

    -- 审批基本信息
    approval_stage VARCHAR(100) NOT NULL COMMENT '审批阶段',
    approver_id VARCHAR(36) NOT NULL COMMENT '审批人ID',
    approver_name VARCHAR(255) NOT NULL COMMENT '审批人姓名',
    approver_title VARCHAR(100) COMMENT '审批人职位',
    approver_role ENUM('LAWYER', 'DEPARTMENT_HEAD', 'COMPLIANCE_OFFICER', 'MANAGING_PARTNER', 'ETHICS_COMMITTEE') NOT NULL COMMENT '审批人角色',

    -- 审批决定
    decision ENUM('APPROVE', 'REJECT', 'REQUEST_CHANGES', 'DEFER', 'ESCALATE') NOT NULL COMMENT '审批决定',
    decision_reason TEXT NOT NULL COMMENT '审批理由',
    decision_comments TEXT COMMENT '审批意见',

    -- 条件和限制
    approved_conditions JSON COMMENT '批准的条件',
    imposed_limitations JSON COMMENT '施加的限制',
    monitoring_requirements JSON COMMENT '监控要求',
    reporting_requirements JSON COMMENT '报告要求',

    -- 风险评估
    risk_assessment JSON COMMENT '风险评估结果',
    risk_mitigation_plan JSON COMMENT '风险缓解计划',
    follow_up_actions JSON COMMENT '后续行动',

    -- 时间信息
    approval_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '审批日期',
    effective_date DATE COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',
    next_review_date DATE COMMENT '下次审查日期',

    -- 附件和证据
    supporting_documents JSON COMMENT '支持文件',
    evidence_references JSON COMMENT '证据引用',

    -- 状态
    status ENUM('ACTIVE', 'SUPERSEDED', 'EXPIRED', 'REVOKED') DEFAULT 'ACTIVE' COMMENT '状态',

    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_waiver_application_id (waiver_application_id),
    INDEX idx_approver_id (approver_id),
    INDEX idx_approval_stage (approval_stage),
    INDEX idx_approver_role (approver_role),
    INDEX idx_decision (decision),
    INDEX idx_approval_date (approval_date),
    INDEX idx_effective_date (effective_date),
    INDEX idx_status (status),

    -- 外键约束
    FOREIGN KEY (waiver_application_id) REFERENCES waiver_applications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='豁免审批记录表';

-- 3. 电子签名记录表
CREATE TABLE IF NOT EXISTS waiver_signatures (
    id VARCHAR(36) PRIMARY KEY COMMENT '签名记录ID',
    waiver_application_id VARCHAR(36) NOT NULL COMMENT '豁免申请ID',

    -- 签名方信息
    signer_type ENUM('CLIENT', 'LAWYER', 'WITNESS', 'NOTARY', 'APPROVER') NOT NULL COMMENT '签名方类型',
    signer_name VARCHAR(255) NOT NULL COMMENT '签名方姓名',
    signer_title VARCHAR(100) COMMENT '签名方职位',
    signer_organization VARCHAR(255) COMMENT '签名方组织',
    signer_contact VARCHAR(255) COMMENT '签名方联系方式',

    -- 签名内容
    signature_content TEXT NOT NULL COMMENT '签名内容',
    signed_statement TEXT NOT NULL COMMENT '签署的声明',
    terms_accepted JSON COMMENT '接受的条款',

    -- 签名技术信息
    signature_method ENUM('ELECTRONIC_SIGNATURE', 'DIGITAL_SIGNATURE', 'WET_SIGNATURE', 'VERBAL_CONSENT') NOT NULL COMMENT '签名方式',
    signature_algorithm VARCHAR(100) COMMENT '签名算法',
    digital_signature_hash VARCHAR(500) COMMENT '数字签名哈希',
    signature_timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '签名时间戳',

    -- 验证信息
    verification_status ENUM('VERIFIED', 'PENDING', 'FAILED', 'EXPIRED') DEFAULT 'PENDING' COMMENT '验证状态',
    verification_method VARCHAR(100) COMMENT '验证方式',
    verification_result JSON COMMENT '验证结果',
    verified_at TIMESTAMP COMMENT '验证时间',

    -- IP地址和设备信息
    signer_ip_address VARCHAR(45) COMMENT '签名方IP地址',
    user_agent TEXT COMMENT '用户代理信息',
    device_fingerprint VARCHAR(255) COMMENT '设备指纹',
    location_info JSON COMMENT '位置信息',

    -- 附加文件
    signed_document_url VARCHAR(500) COMMENT '签名文件URL',
    backup_document_url VARCHAR(500) COMMENT '备份文件URL',
    certificate_url VARCHAR(500) COMMENT '证书URL',

    -- 状态
    status ENUM('ACTIVE', 'REVOKED', 'EXPIRED', 'INVALID') DEFAULT 'ACTIVE' COMMENT '状态',
    revocation_reason TEXT COMMENT '撤销原因',
    revocation_date TIMESTAMP COMMENT '撤销日期',

    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_waiver_application_id (waiver_application_id),
    INDEX idx_signer_type (signer_type),
    INDEX idx_signer_name (signer_name),
    INDEX idx_signature_method (signature_method),
    INDEX idx_verification_status (verification_status),
    INDEX idx_signature_timestamp (signature_timestamp),
    INDEX idx_status (status),
    INDEX idx_signer_ip_address (signer_ip_address),

    -- 外键约束
    FOREIGN KEY (waiver_application_id) REFERENCES waiver_applications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='电子签名记录表';

-- 4. 豁免监控记录表
CREATE TABLE IF NOT EXISTS waiver_monitoring_records (
    id VARCHAR(36) PRIMARY KEY COMMENT '监控记录ID',
    waiver_application_id VARCHAR(36) NOT NULL COMMENT '豁免申请ID',

    -- 监控基本信息
    monitoring_type ENUM('COMPLIANCE_CHECK', 'RISK_ASSESSMENT', 'BARRIER_EFFECTIVENESS', 'PERFORMANCE_REVIEW', 'CLIENT_FEEDBACK') NOT NULL COMMENT '监控类型',
    monitoring_date DATE NOT NULL COMMENT '监控日期',
    monitoring_period_start DATE COMMENT '监控期间开始',
    monitoring_period_end DATE COMMENT '监控期间结束',

    -- 监控内容
    monitoring_items JSON NOT NULL COMMENT '监控项目',
    check_results JSON NOT NULL COMMENT '检查结果',
    findings JSON COMMENT '发现的问题',
    observations TEXT COMMENT '观察记录',

    -- 风险评估
    current_risk_level ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') COMMENT '当前风险等级',
    risk_trend ENUM('IMPROVING', 'STABLE', 'DETERIORATING') COMMENT '风险趋势',
    risk_factors JSON COMMENT '风险因素',
    risk_mitigation_status JSON COMMENT '风险缓解状态',

    -- 合规状况
    compliance_status ENUM('COMPLIANT', 'PARTIALLY_COMPLIANT', 'NON_COMPLIANT', 'UNDER_REVIEW') DEFAULT 'UNDER_REVIEW' COMMENT '合规状况',
    compliance_issues JSON COMMENT '合规问题',
    corrective_actions JSON COMMENT '纠正措施',

    -- 豁免条件执行情况
    conditions_compliance JSON COMMENT '条件执行情况',
    limitation_adherence JSON COMMENT '限制遵守情况',
    barrier_effectiveness JSON COMMENT '屏障有效性',

    -- 后续行动
    recommended_actions JSON COMMENT '建议行动',
    required_follow_up JSON COMMENT '需要跟进事项',
    next_monitoring_date DATE COMMENT '下次监控日期',

    -- 报告信息
    report_generated BOOLEAN DEFAULT FALSE COMMENT '是否已生成报告',
    report_url VARCHAR(500) COMMENT '报告URL',
    report_recipients JSON COMMENT '报告接收人',

    -- 状态
    status ENUM('SCHEDULED', 'IN_PROGRESS', 'COMPLETED', 'OVERDUE', 'CANCELLED') DEFAULT 'SCHEDULED' COMMENT '状态',

    -- 审计信息
    monitored_by VARCHAR(36) NOT NULL COMMENT '监控人',
    reviewed_by VARCHAR(36) COMMENT '审核人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_waiver_application_id (waiver_application_id),
    INDEX idx_monitoring_type (monitoring_type),
    INDEX idx_monitoring_date (monitoring_date),
    INDEX idx_current_risk_level (current_risk_level),
    INDEX idx_compliance_status (compliance_status),
    INDEX idx_status (status),
    INDEX idx_next_monitoring_date (next_monitoring_date),
    INDEX idx_monitored_by (monitored_by),

    -- 外键约束
    FOREIGN KEY (waiver_application_id) REFERENCES waiver_applications(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='豁免监控记录表';

-- 5. 豁免模板表
CREATE TABLE IF NOT EXISTS waiver_templates (
    id VARCHAR(36) PRIMARY KEY COMMENT '模板ID',
    template_name VARCHAR(255) NOT NULL COMMENT '模板名称',
    template_code VARCHAR(100) NOT NULL UNIQUE COMMENT '模板代码',

    -- 模板分类
    template_type ENUM('INFORMED_CONSENT', 'ETHICAL_BARRIER', 'INFORMATION_SCREEN', 'MONITORING_PLAN') NOT NULL COMMENT '模板类型',
    template_category VARCHAR(100) COMMENT '模板分类',
    practice_area VARCHAR(100) COMMENT '执业领域',

    -- 模板内容
    template_content JSON NOT NULL COMMENT '模板内容',
    required_clauses JSON COMMENT '必需条款',
    optional_clauses JSON COMMENT '可选条款',
    placeholders JSON COMMENT '占位符说明',

    -- 适用条件
    applicable_scenarios JSON COMMENT '适用场景',
    conflict_types JSON COMMENT '适用的冲突类型',
    risk_levels JSON COMMENT '适用的风险等级',
    jurisdiction_rules JSON COMMENT '司法管辖区规则',

    -- 审批要求
    approval_requirements JSON COMMENT '审批要求',
    required_approvers JSON COMMENT '必需审批人',
    approval_workflow JSON COMMENT '审批流程',

    -- 状态和版本
    status ENUM('ACTIVE', 'INACTIVE', 'UNDER_REVIEW', 'DEPRECATED') DEFAULT 'ACTIVE' COMMENT '状态',
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
    INDEX idx_template_category (template_category),
    INDEX idx_practice_area (practice_area),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),
    INDEX idx_usage_count (usage_count)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='豁免模板表';

-- 插入默认豁免模板
INSERT INTO waiver_templates (
    id, template_name, template_code, template_type, template_category,
    template_content, required_clauses, optional_clauses,
    approval_requirements, effective_date, created_by
) VALUES
-- 知情同意模板
(UUID(), '标准知情同意书', 'STANDARD_INFORMED_CONSENT', 'INFORMED_CONSENT', '通用',
JSON_OBJECT(
    'title', '利益冲突知情同意书',
    'introduction', '本律师事务所在此向客户充分披露潜在的利益冲突情况',
    'conflict_disclosure', '【冲突情况披露】',
    'potential_risks', '【潜在风险说明】',
    'alternatives', '【替代方案说明】',
    'client_rights', '【客户权利声明】',
    'consent_statement', '【同意声明】',
    'contact_information', '【联系方式】'
),
JSON_ARRAY('冲突情况披露', '潜在风险说明', '同意声明'),
JSON_ARRAY('替代方案说明', '客户权利声明', '法律后果说明'),
JSON_OBJECT('required_approvers' => JSON_ARRAY('DEPARTMENT_HEAD', 'COMPLIANCE_OFFICER')),
CURDATE(), 'system'),

-- 伦理屏障模板
(UUID(), '伦理屏障设置同意书', 'ETHICAL_BARRIER_CONSENT', 'ETHICAL_BARRIER', '通用',
JSON_OBJECT(
    'title', '伦理屏障设置同意书',
    'barrier_description', '【屏障描述】',
    'restricted_information', '【受限信息说明】',
    'access_controls', '【访问控制措施】',
    'monitoring_plan', '【监控计划】',
    'violation_consequences', '【违规后果】',
    'agreement_statement', '【同意声明】'
),
JSON_ARRAY('屏障描述', '访问控制措施', '同意声明'),
JSON_ARRAY('受限信息说明', '监控计划', '违规后果'),
JSON_OBJECT('required_approvers' => JSON_ARRAY('COMPLIANCE_OFFICER', 'MANAGING_PARTNER')),
CURDATE(), 'system'),

-- 信息屏蔽模板
(UUID(), '信息屏蔽协议', 'INFORMATION_SCREEN_PROTOCOL', 'INFORMATION_SCREEN', '通用',
JSON_OBJECT(
    'title', '信息屏蔽协议',
    'screening_scope', '【屏蔽范围】',
    'screening_methods', '【屏蔽方法】',
    'personnel_restrictions', '【人员限制】',
    'communication_protocols', '【沟通协议】',
    'emergency_procedures', '【应急程序】',
    'compliance_monitoring', '【合规监控】'
),
JSON_ARRAY('屏蔽范围', '屏蔽方法', '应急程序'),
JSON_ARRAY('人员限制', '沟通协议', '合规监控'),
JSON_OBJECT('required_approvers' => JSON_ARRAY('COMPLIANCE_OFFICER')),
CURDATE(), 'system');

-- 6. 创建视图：豁免申请完整信息视图
CREATE OR REPLACE VIEW waiver_applications_complete AS
SELECT
    wa.*,
    -- 申请编号格式化
    CONCAT('WA-', DATE_FORMAT(wa.created_at, '%Y%m%d'), '-', LPAD(wa.id, 8, '0')) as formatted_application_number,

    -- 当前审批状态
    COALESCE(war.decision, 'PENDING') as current_decision,
    COALESCE(war.approver_name, '待审批') as current_approver,
    COALESCE(war.approval_date, NULL) as current_approval_date,

    -- 签名状态统计
    (SELECT COUNT(*) FROM waiver_signatures ws WHERE ws.waiver_application_id = wa.id AND ws.status = 'ACTIVE') as active_signatures_count,
    (SELECT COUNT(*) FROM waiver_signatures ws WHERE ws.waiver_application_id = wa.id AND ws.signer_type = 'CLIENT' AND ws.status = 'ACTIVE') as client_signature_count,
    (SELECT COUNT(*) FROM waiver_signatures ws WHERE ws.waiver_application_id = wa.id AND ws.signer_type = 'LAWYER' AND ws.status = 'ACTIVE') as lawyer_signature_count,

    -- 最近监控记录
    (SELECT monitoring_date FROM waiver_monitoring_records wmr
     WHERE wmr.waiver_application_id = wa.id
     ORDER BY wmr.monitoring_date DESC LIMIT 1) as last_monitoring_date,

    -- 监控状态
    (SELECT status FROM waiver_monitoring_records wmr
     WHERE wmr.waiver_application_id = wa.id
     ORDER BY wmr.monitoring_date DESC LIMIT 1) as current_monitoring_status,

    -- 风险等级
    JSON_UNQUOTE(JSON_EXTRACT(wa.risk_assessment, '$.overallRisk')) as risk_level,

    -- 豁免期限信息
    CASE
        WHEN wa.requested_expiry_date IS NOT NULL THEN DATEDIFF(wa.requested_expiry_date, wa.requested_effective_date)
        WHEN wa.duration_days IS NOT NULL THEN wa.duration_days
        ELSE NULL
    END as waiver_duration_days,

    -- 是否即将到期（30天内）
    CASE
        WHEN (
            (wa.requested_expiry_date IS NOT NULL AND DATEDIFF(wa.requested_expiry_date, CURDATE()) BETWEEN 0 AND 30) OR
            (wa.duration_days IS NOT NULL AND DATEDIFF(DATE_ADD(wa.requested_effective_date, INTERVAL wa.duration_days DAY), CURDATE()) BETWEEN 0 AND 30)
        ) THEN TRUE
        ELSE FALSE
    END as expiring_soon,

    -- 是否已过期
    CASE
        WHEN (
            (wa.requested_expiry_date IS NOT NULL AND wa.requested_expiry_date < CURDATE()) OR
            (wa.duration_days IS NOT NULL AND DATE_ADD(wa.requested_effective_date, INTERVAL wa.duration_days DAY) < CURDATE())
        ) THEN TRUE
        ELSE FALSE
    END as expired

FROM waiver_applications wa
LEFT JOIN waiver_approval_records war ON wa.id = war.waiver_application_id
    AND war.id = (
        SELECT MAX(id) FROM waiver_approval_records
        WHERE waiver_application_id = wa.id
    )
WHERE wa.deleted_at IS NULL;

-- 7. 创建存储过程：生成申请编号
DELIMITER $$
CREATE PROCEDURE IF NOT EXISTS generate_waiver_application_number()
BEGIN
    DECLARE current_date DATE;
    DECLARE date_prefix VARCHAR(8);
    DECLARE sequence_number INT;
    DECLARE application_number VARCHAR(50);

    SET current_date = CURDATE();
    SET date_prefix = DATE_FORMAT(current_date, '%Y%m%d');

    -- 获取当天的序列号
    SELECT COALESCE(MAX(CAST(SUBSTRING(application_number, 12) AS UNSIGNED)), 0) + 1
    INTO sequence_number
    FROM waiver_applications
    WHERE DATE(created_at) = current_date;

    -- 生成申请编号
    SET application_number = CONCAT('WA-', date_prefix, '-', LPAD(sequence_number, 6, '0'));

    -- 返回申请编号
    SELECT application_number as next_application_number;
END$$
DELIMITER ;

-- 8. 创建触发器：更新使用统计
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS increment_template_usage
    AFTER INSERT ON waiver_applications
    FOR EACH ROW
BEGIN
    UPDATE waiver_templates
    SET usage_count = usage_count + 1,
        last_used_date = NOW()
    WHERE id = (
        SELECT template_id FROM waiver_application_templates
        WHERE application_id = NEW.id
        LIMIT 1
    );
END$$
DELIMITER ;