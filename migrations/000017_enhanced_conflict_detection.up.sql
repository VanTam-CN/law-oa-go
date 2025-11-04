-- 增强冲突检测引擎数据模型
-- 专业级多维度冲突检测系统

-- 1. 专业冲突检查请求表
CREATE TABLE IF NOT EXISTS professional_conflict_check_requests (
    id VARCHAR(36) PRIMARY KEY COMMENT '请求ID',
    check_number VARCHAR(50) NOT NULL UNIQUE COMMENT '检查编号',

    -- 基本请求信息
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    client_name VARCHAR(255) NOT NULL COMMENT '客户名称',
    client_type ENUM('PERSON', 'COMPANY', 'GOVERNMENT', 'NGO') NOT NULL COMMENT '客户类型',
    client_legal_identifier VARCHAR(100) COMMENT '客户法律标识（注册号/身份证号）',

    -- 案件信息
    case_name VARCHAR(500) NOT NULL COMMENT '案件名称',
    case_number VARCHAR(100) COMMENT '案件编号',
    case_type ENUM('CIVIL', 'COMMERCIAL', 'CRIMINAL', 'ADMINISTRATIVE', 'LABOR', 'INTELLECTUAL_PROPERTY', 'FAMILY', 'ARBITRATION', 'OTHER') NOT NULL COMMENT '案件类型',
    case_subtype VARCHAR(100) COMMENT '案件子类型',
    practice_area VARCHAR(100) NOT NULL COMMENT '执业领域',
    jurisdiction VARCHAR(100) COMMENT '司法管辖区',

    -- 对方当事人信息
    opposing_parties JSON NOT NULL COMMENT '对方当事人列表',
    related_parties JSON COMMENT '相关方列表',
    third_parties JSON COMMENT '第三方信息',

    -- 律师团队信息
    primary_lawyer_id VARCHAR(36) NOT NULL COMMENT '主办律师ID',
    primary_lawyer_name VARCHAR(255) NOT NULL COMMENT '主办律师姓名',
    team_members JSON NOT NULL COMMENT '团队成员列表',
    supervising_lawyer_id VARCHAR(36) COMMENT '监督律师ID',

    -- 检查参数
    search_depth ENUM('BASIC', 'STANDARD', 'COMPREHENSIVE', 'EXHAUSTIVE') DEFAULT 'STANDARD' COMMENT '搜索深度',
    search_years INT DEFAULT 5 COMMENT '搜索年限',
    include_corporate_relations BOOLEAN DEFAULT TRUE COMMENT '是否包含企业关系',
    include_family_relations BOOLEAN DEFAULT TRUE COMMENT '是否包含家庭关系',
    include_financial_relations BOOLEAN DEFAULT TRUE COMMENT '是否包含财务关系',
    include_time_series_analysis BOOLEAN DEFAULT TRUE COMMENT '是否包含时间序列分析',

    -- 特殊检查要求
    special_requirements JSON COMMENT '特殊检查要求',
    exclusion_criteria JSON COMMENT '排除标准',
    custom_search_rules JSON COMMENT '自定义搜索规则',

    -- 客户指示和限制
    client_instructions TEXT COMMENT '客户指示',
    confidentiality_level ENUM('STANDARD', 'CONFIDENTIAL', 'HIGHLY_CONFIDENTIAL') DEFAULT 'STANDARD' COMMENT '保密级别',
    disclosure_restrictions JSON COMMENT '披露限制',

    -- 风险评估参数
    risk_tolerance_level ENUM('LOW', 'MEDIUM', 'HIGH') DEFAULT 'MEDIUM' COMMENT '风险承受水平',
    acceptable_risk_score DECIMAL(5,2) DEFAULT 70.00 COMMENT '可接受风险评分',
    mandatory_approvals JSON COMMENT '必需的审批',

    -- 紧急程度和优先级
    urgency_level ENUM('LOW', 'MEDIUM', 'HIGH', 'URGENT') DEFAULT 'MEDIUM' COMMENT '紧急程度',
    business_priority ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM' COMMENT '业务优先级',
    deadline TIMESTAMP COMMENT '截止时间',

    -- 状态管理
    status ENUM('DRAFT', 'SUBMITTED', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED') DEFAULT 'DRAFT' COMMENT '状态',
    submission_date TIMESTAMP COMMENT '提交时间',
    completion_date TIMESTAMP COMMENT '完成时间',

    -- 关联信息
    related_checks JSON COMMENT '相关检查记录',
    reference_checks JSON COMMENT '参考检查记录',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_check_number (check_number),
    INDEX idx_client_id (client_id),
    INDEX idx_client_name (client_name),
    INDEX idx_case_name (case_name),
    INDEX idx_case_type (case_type),
    INDEX idx_practice_area (practice_area),
    INDEX idx_primary_lawyer_id (primary_lawyer_id),
    INDEX idx_primary_lawyer_name (primary_lawyer_name),
    INDEX idx_search_depth (search_depth),
    INDEX idx_status (status),
    INDEX idx_submission_date (submission_date),
    INDEX idx_urgency_level (urgency_level),
    INDEX idx_business_priority (business_priority),
    INDEX idx_created_at (created_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_conflict_check (client_name, case_name, opposing_parties, related_parties, client_instructions)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='专业冲突检查请求表';

-- 2. 多维度冲突检测结果表
CREATE TABLE IF NOT EXISTS multi_dimensional_conflict_results (
    id VARCHAR(36) PRIMARY KEY COMMENT '结果ID',
    check_request_id VARCHAR(36) NOT NULL COMMENT '检查请求ID',

    -- 冲突基本信息
    conflict_id VARCHAR(36) NOT NULL UNIQUE COMMENT '冲突ID',
    conflict_title VARCHAR(500) NOT NULL COMMENT '冲突标题',
    conflict_description TEXT NOT NULL COMMENT '冲突描述',

    -- 冲突分类
    conflict_type ENUM('DIRECT_OPPOSITION', 'CONCURRENT', 'FORMER_CLIENT', 'RELATIONSHIP',
                      'FINANCIAL', 'OFFICIAL', 'RELATED_ENTITY', 'INFORMATION', 'TEMPORAL', 'PROFESSIONAL') NOT NULL COMMENT '冲突类型',
    conflict_subtype VARCHAR(100) COMMENT '冲突子类型',
    classification_standard VARCHAR(50) NOT NULL COMMENT '分类标准',
    classification_code VARCHAR(50) NOT NULL COMMENT '分类代码',
    category VARCHAR(100) NOT NULL COMMENT '冲突类别',

    -- 风险评估
    risk_level ENUM('CRITICAL', 'HIGH', 'MEDIUM', 'LOW') NOT NULL COMMENT '风险等级',
    risk_score DECIMAL(5,2) NOT NULL COMMENT '风险评分(0-100)',
    risk_factors JSON NOT NULL COMMENT '风险因素',
    impact_assessment JSON COMMENT '影响评估',

    -- 可处理性评估
    waivable BOOLEAN DEFAULT FALSE COMMENT '是否可豁免',
    waiver_conditions JSON COMMENT '豁免条件',
    approval_required BOOLEAN DEFAULT TRUE COMMENT '是否需要审批',
    approval_level ENUM('SELF', 'DEPARTMENT_HEAD', 'COMMITTEE', 'MANAGING_PARTNER', 'ETHICS_OFFICER') DEFAULT 'DEPARTMENT_HEAD' COMMENT '审批级别',

    -- 冲突来源信息
    source_entity_type ENUM('CLIENT', 'LAWYER', 'CASE', 'THIRD_PARTY', 'ORGANIZATION') NOT NULL COMMENT '来源实体类型',
    source_entity_id VARCHAR(36) NOT NULL COMMENT '来源实体ID',
    source_entity_name VARCHAR(255) NOT NULL COMMENT '来源实体名称',
    related_entity_id VARCHAR(36) COMMENT '关联实体ID',
    related_entity_name VARCHAR(255) COMMENT '关联实体名称',

    -- 时间维度信息
    conflict_time_dimension ENUM('CURRENT', 'HISTORICAL', 'FUTURE', 'ONGOING') NOT NULL COMMENT '时间维度',
    conflict_date DATE COMMENT '冲突日期',
    time_relevance ENUM('RECENT', 'MEDIUM_TERM', 'LONG_TERM') COMMENT '时间相关性',
    cooling_period_days INT DEFAULT 0 COMMENT '冷却期天数',
    expiry_date DATE COMMENT '到期日期',

    -- 业务维度信息
    business_dimension ENUM('CASE_RELATED', 'CLIENT_RELATED', 'LAWYER_RELATED', 'ORGANIZATIONAL', 'REGULATORY') NOT NULL COMMENT '业务维度',
    practice_area VARCHAR(100) COMMENT '执业领域',
    jurisdiction VARCHAR(100) COMMENT '司法管辖区',
    regulatory_framework JSON COMMENT '监管框架',

    -- 检测方法信息
    detection_method ENUM('AUTOMATED', 'MANUAL_REVIEW', 'RULE_BASED', 'MACHINE_LEARNING', 'HUMAN_JUDGMENT') NOT NULL COMMENT '检测方法',
    detection_confidence DECIMAL(3,2) DEFAULT 0.80 COMMENT '检测置信度',
    detection_rules_applied JSON COMMENT '应用的检测规则',
    algorithm_version VARCHAR(50) COMMENT '算法版本',

    -- 处理建议
    recommended_actions JSON NOT NULL COMMENT '建议行动',
    mitigation_measures JSON COMMENT '缓解措施',
    monitoring_requirements JSON COMMENT '监控要求',
    follow_up_actions JSON COMMENT '后续行动',

    -- 证据和文档
    supporting_evidence JSON COMMENT '支持证据',
    related_documents JSON COMMENT '相关文档',
    case_references JSON COMMENT '案例引用',
    regulatory_references JSON COMMENT '法规引用',

    -- 状态和验证
    verification_status ENUM('UNVERIFIED', 'VERIFIED', 'DISPUTED', 'SUPERSEDED') DEFAULT 'UNVERIFIED' COMMENT '验证状态',
    verification_comments TEXT COMMENT '验证意见',
    verified_by VARCHAR(36) COMMENT '验证人',
    verified_at TIMESTAMP COMMENT '验证时间',

    -- 关联分析
    related_conflicts JSON COMMENT '相关冲突',
    conflict_clusters JSON COMMENT '冲突集群',
    systemic_issues JSON COMMENT '系统性问题',

    -- 状态
    status ENUM('ACTIVE', 'RESOLVED', 'MONITORED', 'ESCALATED', 'CLOSED') DEFAULT 'ACTIVE' COMMENT '状态',
    resolution_method VARCHAR(100) COMMENT '解决方式',
    resolution_date TIMESTAMP COMMENT '解决日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_check_request_id (check_request_id),
    INDEX idx_conflict_id (conflict_id),
    INDEX idx_conflict_type (conflict_type),
    INDEX idx_classification_code (classification_code),
    INDEX idx_risk_level (risk_level),
    INDEX idx_risk_score (risk_score),
    INDEX idx_waivable (waivable),
    INDEX idx_approval_required (approval_required),
    INDEX idx_source_entity_type (source_entity_type),
    INDEX idx_source_entity_name (source_entity_name),
    INDEX idx_conflict_time_dimension (conflict_time_dimension),
    INDEX idx_business_dimension (business_dimension),
    INDEX idx_detection_method (detection_method),
    INDEX idx_verification_status (verification_status),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),

    -- 复合索引
    INDEX idx_request_type_risk (check_request_id, conflict_type, risk_level),
    INDEX idx_source_risk_status (source_entity_name, risk_level, status),

    -- 外键约束
    FOREIGN KEY (check_request_id) REFERENCES professional_conflict_check_requests(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='多维度冲突检测结果表';

-- 3. 冲突检测规则引擎表
CREATE TABLE IF NOT EXISTS conflict_detection_rules (
    id VARCHAR(36) PRIMARY KEY COMMENT '规则ID',
    rule_name VARCHAR(255) NOT NULL COMMENT '规则名称',
    rule_code VARCHAR(100) NOT NULL UNIQUE COMMENT '规则代码',
    rule_description TEXT NOT NULL COMMENT '规则描述',

    -- 规则分类
    rule_category ENUM('ENTITY_MATCHING', 'RELATIONSHIP_ANALYSIS', 'TIME_SERIES', 'RISK_ASSESSMENT',
                       'COMPLIANCE_CHECK', 'PROFESSIONAL_STANDARDS', 'CUSTOM_RULES') NOT NULL COMMENT '规则类别',
    rule_type ENUM('HARD_RULE', 'SOFT_RULE', 'ADVISORY', 'WARNING') NOT NULL COMMENT '规则类型',
    severity_level ENUM('CRITICAL', 'HIGH', 'MEDIUM', 'LOW') NOT NULL COMMENT '严重程度',

    -- 规则条件
    trigger_conditions JSON NOT NULL COMMENT '触发条件',
    matching_criteria JSON NOT NULL COMMENT '匹配条件',
    exclusion_conditions JSON COMMENT '排除条件',
    validation_rules JSON COMMENT '验证规则',

    -- 规则逻辑
    rule_logic JSON NOT NULL COMMENT '规则逻辑',
    decision_tree JSON COMMENT '决策树',
    algorithm JSON COMMENT '算法配置',
    machine_learning_model JSON COMMENT '机器学习模型',

    -- 输出配置
    output_template JSON NOT NULL COMMENT '输出模板',
    result_actions JSON NOT NULL COMMENT '结果行动',
    risk_score_formula VARCHAR(500) COMMENT '风险评分公式',
    recommendation_rules JSON COMMENT '建议规则',

    -- 适用范围
    applicable_jurisdictions JSON COMMENT '适用司法管辖区',
    applicable_practice_areas JSON COMMENT '适用执业领域',
    applicable_entity_types JSON COMMENT '适用实体类型',
    applicable_case_types JSON COMMENT '适用案件类型',

    -- 性能配置
    execution_priority INT DEFAULT 5 COMMENT '执行优先级',
    timeout_seconds INT DEFAULT 30 COMMENT '超时时间',
    max_results INT DEFAULT 100 COMMENT '最大结果数',
    caching_enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用缓存',
    cache_ttl_seconds INT DEFAULT 3600 COMMENT '缓存有效期',

    -- 版本和状态
    version INT DEFAULT 1 COMMENT '版本号',
    status ENUM('ACTIVE', 'INACTIVE', 'DEPRECATED', 'UNDER_REVIEW') DEFAULT 'ACTIVE' COMMENT '状态',
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 测试和验证
    test_cases JSON COMMENT '测试用例',
    validation_results JSON COMMENT '验证结果',
    last_test_date TIMESTAMP COMMENT '最后测试日期',
    accuracy_rate DECIMAL(5,2) COMMENT '准确率',

    -- 使用统计
    usage_count INT DEFAULT 0 COMMENT '使用次数',
    success_rate DECIMAL(5,2) COMMENT '成功率',
    avg_execution_time DECIMAL(10,2) COMMENT '平均执行时间（毫秒）',
    last_used_date TIMESTAMP COMMENT '最后使用日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    approved_by VARCHAR(36) COMMENT '审批人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_rule_code (rule_code),
    INDEX idx_rule_category (rule_category),
    INDEX idx_rule_type (rule_type),
    INDEX idx_severity_level (severity_level),
    INDEX idx_execution_priority (execution_priority),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),
    INDEX idx_usage_count (usage_count),
    INDEX idx_success_rate (success_rate),
    INDEX idx_created_at (created_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_rule (rule_name, rule_description)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测规则引擎表';

-- 4. 冲突检测规则执行记录表
CREATE TABLE IF NOT EXISTS conflict_rule_executions (
    id VARCHAR(36) PRIMARY KEY COMMENT '执行记录ID',
    check_request_id VARCHAR(36) NOT NULL COMMENT '检查请求ID',
    rule_id VARCHAR(36) NOT NULL COMMENT '规则ID',

    -- 执行信息
    execution_start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '执行开始时间',
    execution_end_time TIMESTAMP COMMENT '执行结束时间',
    execution_duration_ms INT COMMENT '执行时长（毫秒）',
    execution_status ENUM('STARTED', 'RUNNING', 'COMPLETED', 'FAILED', 'TIMEOUT', 'CANCELLED') NOT NULL COMMENT '执行状态',

    -- 输入数据
    input_data JSON NOT NULL COMMENT '输入数据',
    processed_data JSON COMMENT '处理数据',

    -- 执行结果
    rule_matched BOOLEAN DEFAULT FALSE COMMENT '规则是否匹配',
    match_confidence DECIMAL(3,2) COMMENT '匹配置信度',
    match_details JSON COMMENT '匹配详情',
    rule_output JSON COMMENT '规则输出',

    -- 错误信息
    error_message TEXT COMMENT '错误信息',
    error_code VARCHAR(100) COMMENT '错误代码',
    error_stack_trace TEXT COMMENT '错误堆栈',
    retry_count INT DEFAULT 0 COMMENT '重试次数',

    -- 性能信息
    memory_usage_mb DECIMAL(10,2) COMMENT '内存使用量（MB）',
    cpu_usage_percent DECIMAL(5,2) COMMENT 'CPU使用率',
    database_queries_count INT COMMENT '数据库查询次数',
    cache_hit_rate DECIMAL(5,2) COMMENT '缓存命中率',

    -- 环境信息
    execution_environment VARCHAR(100) COMMENT '执行环境',
    algorithm_version VARCHAR(50) COMMENT '算法版本',
    model_version VARCHAR(50) COMMENT '模型版本',

    -- 状态
    status ENUM('PENDING', 'COMPLETED', 'FAILED', 'CANCELLED') DEFAULT 'PENDING' COMMENT '状态',

    -- 审计信息
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_check_request_id (check_request_id),
    INDEX idx_rule_id (rule_id),
    INDEX idx_execution_start_time (execution_start_time),
    INDEX idx_execution_status (execution_status),
    INDEX idx_rule_matched (rule_matched),
    INDEX idx_execution_duration_ms (execution_duration_ms),
    INDEX idx_status (status),

    -- 外键约束
    FOREIGN KEY (check_request_id) REFERENCES professional_conflict_check_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (rule_id) REFERENCES conflict_detection_rules(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测规则执行记录表';

-- 5. 插入默认的冲突检测规则
INSERT INTO conflict_detection_rules (
    id, rule_name, rule_code, rule_description,
    rule_category, rule_type, severity_level,
    trigger_conditions, matching_criteria, rule_logic,
    output_template, result_actions, risk_score_formula,
    applicable_jurisdictions, effective_date, created_by
) VALUES
-- 客户名称直接匹配规则
(UUID(), '客户名称直接匹配', 'CLIENT_NAME_EXACT_MATCH', '检查客户名称是否与历史案件客户名称完全匹配',
'ENTITY_MATCHING', 'HARD_RULE', 'HIGH',
JSON_OBJECT('comparison_type' => 'EXACT', 'case_sensitive' => false, 'fuzzy_threshold' => 0.9),
JSON_OBJECT('fields' => JSON_ARRAY('client_name', 'opposing_parties')),
JSON_OBJECT('algorithm' => 'string_matching', 'parameters' => JSON_OBJECT('similarity_threshold' => 0.9)),
JSON_OBJECT('conflict_type' => 'DIRECT_OPPOSITION', 'risk_level' => 'HIGH', 'description_template' => '客户名称与历史案件客户名称匹配：{client_name}'),
JSON_ARRAY('标记为直接对立冲突', '建议拒绝代理', '记录详细冲突信息'),
'85.0 + (case_similarity * 15)',
JSON_ARRAY('CHINA', 'US', 'UK', 'EU'),
CURDATE(), 'system'),

-- 关联企业检测规则
(UUID(), '关联企业检测', 'RELATED_ENTITY_DETECTION', '检测客户与企业历史案件客户是否存在关联关系',
'RELATIONSHIP_ANALYSIS', 'SOFT_RULE', 'MEDIUM',
JSON_OBJECT('relationship_types' => JSON_ARRAY('PARENT', 'SUBSIDIARY', 'SISTER', 'JOINT_VENTURE'), 'confidence_threshold' => 0.7),
JSON_OBJECT('data_sources' => JSON_ARRAY('client_profiles', 'client_relationships', 'corporate_registry')),
JSON_OBJECT('algorithm' => 'relationship_graph', 'depth' => 3, 'min_confidence' => 0.6),
JSON_OBJECT('conflict_type' => 'RELATED_ENTITY', 'risk_level' => 'MEDIUM', 'description_template' => '检测到关联企业关系：{entity1} - {entity2} ({relationship})'),
JSON_ARRAY('标记为关联实体冲突', '建议信息屏障', '加强监控'),
'60.0 + (relationship_strength * 25) + (confidence_score * 15)',
JSON_ARRAY('CHINA', 'US', 'UK', 'EU'),
CURDATE(), 'system'),

-- 时间序列冲突规则
(UUID(), '时间序列冲突检测', 'TIME_SERIES_CONFLICT_DETECTION', '检测律师在特定时间范围内的代理冲突',
'TIME_SERIES', 'HARD_RULE', 'HIGH',
JSON_OBJECT('time_window_years' => 3, 'conflict_types' => JSON_ARRAY('SAME_CASE', 'RELATED_CASE', 'OPPOSING_PARTY')),
JSON_OBJECT('check_criteria' => JSON_ARRAY('same_lawyer', 'same_client', 'related_case')),
JSON_OBJECT('algorithm' => 'temporal_analysis', 'cooling_period_days' => 365),
JSON_OBJECT('conflict_type' => 'CONCURRENT', 'risk_level' => 'HIGH', 'description_template' => '律师在{time_window}年内代理相关案件：{case_details}'),
JSON_ARRAY('标记为并发冲突', '评估风险等级', '考虑豁免条件'),
'70.0 + (time_relevance_score * 20) + (case_similarity * 10)',
JSON_ARRAY('CHINA', 'US', 'UK', 'EU'),
CURDATE(), 'system'),

-- 财务利益冲突规则
(UUID(), '财务利益冲突检测', 'FINANCIAL_INTEREST_CONFLICT', '检测律师与客户是否存在财务利益关系',
'COMPLIANCE_CHECK', 'HARD_RULE', 'CRITICAL',
JSON_OBJECT('financial_relationships' => JSON_ARRAY('HOLDINGS', 'INVESTMENTS', 'LOANS', 'GUARANTEES'), 'materiality_threshold' => 10000),
JSON_OBJECT('screening_criteria' => JSON_ARRAY('direct_holdings', 'indirect_holdings', 'family_holdings')),
JSON_OBJECT('algorithm' => 'financial_interest_analysis', 'reporting_threshold' => 5000),
JSON_OBJECT('conflict_type' => 'FINANCIAL', 'risk_level' => 'CRITICAL', 'description_template' => '检测到财务利益关系：{relationship_details}'),
JSON_ARRAY('标记为财务利益冲突', '要求完全披露', '考虑放弃代理'),
'90.0 + (financial_amount_score * 10)',
JSON_ARRAY('CHINA', 'US', 'UK', 'EU'),
CURDATE(), 'system');

-- 6. 创建视图：专业冲突检查统计视图
CREATE OR REPLACE VIEW professional_conflict_check_stats AS
SELECT
    -- 按日期统计
    DATE(pccr.created_at) as check_date,
    COUNT(*) as total_checks,
    SUM(CASE WHEN pccr.status = 'COMPLETED' THEN 1 ELSE 0 END) as completed_checks,
    SUM(CASE WHEN pccr.status = 'FAILED' THEN 1 ELSE 0 END) as failed_checks,
    AVG(TIMESTAMPDIFF(SECOND, pccr.submission_date, pccr.completion_date)) as avg_processing_time_seconds,

    -- 按案件类型统计
    pccr.case_type,
    COUNT(*) as checks_by_case_type,

    -- 按风险等级统计
    COUNT(CASE WHEN mdcr.risk_level = 'CRITICAL' THEN 1 END) as critical_conflicts,
    COUNT(CASE WHEN mdcr.risk_level = 'HIGH' THEN 1 END) as high_conflicts,
    COUNT(CASE WHEN mdcr.risk_level = 'MEDIUM' THEN 1 END) as medium_conflicts,
    COUNT(CASE WHEN mdcr.risk_level = 'LOW' THEN 1 END) as low_conflicts,
    AVG(mdcr.risk_score) as avg_risk_score,

    -- 按冲突类型统计
    mdcr.conflict_type,
    COUNT(*) as conflicts_by_type,

    -- 按检测方法统计
    mdcr.detection_method,
    AVG(mdcr.detection_confidence) as avg_detection_confidence,

    -- 按可处理性统计
    COUNT(CASE WHEN mdcr.waivable = TRUE THEN 1 END) as waivable_conflicts,
    COUNT(CASE WHEN mdcr.approval_required = TRUE THEN 1 END) as approval_required_conflicts

FROM professional_conflict_check_requests pccr
LEFT JOIN multi_dimensional_conflict_results mdcr ON pccr.id = mdcr.check_request_id
WHERE pccr.deleted_at IS NULL
GROUP BY DATE(pccr.created_at), pccr.case_type, mdcr.conflict_type, mdcr.detection_method
ORDER BY check_date DESC;

-- 7. 创建存储过程：生成专业冲突检查编号
DELIMITER $$
CREATE PROCEDURE IF NOT EXISTS generate_professional_check_number()
BEGIN
    DECLARE current_date DATE;
    DECLARE date_prefix VARCHAR(8);
    DECLARE sequence_number INT;
    DECLARE check_number VARCHAR(50);

    SET current_date = CURDATE();
    SET date_prefix = DATE_FORMAT(current_date, '%Y%m%d');

    -- 获取当天的序列号
    SELECT COALESCE(MAX(CAST(SUBSTRING(check_number, 12) AS UNSIGNED)), 0) + 1
    INTO sequence_number
    FROM professional_conflict_check_requests
    WHERE DATE(created_at) = current_date;

    -- 生成检查编号
    SET check_number = CONCAT('CC-', date_prefix, '-', LPAD(sequence_number, 6, '0'));

    -- 返回检查编号
    SELECT check_number as next_check_number;
END$$
DELIMITER ;

-- 8. 创建触发器：自动执行冲突检测规则
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS auto_execute_conflict_detection
    AFTER INSERT ON professional_conflict_check_requests
    FOR EACH ROW
BEGIN
        DECLARE rule_count INT;

        -- 获取活跃规则数量
        SELECT COUNT(*) INTO rule_count
        FROM conflict_detection_rules
        WHERE status = 'ACTIVE'
          AND (expiry_date IS NULL OR expiry_date >= CURDATE());

        -- 为每个请求创建规则执行记录
        INSERT INTO conflict_rule_executions (
            id, check_request_id, rule_id, execution_start_time, execution_status, input_data, status
        )
        SELECT
            UUID(),
            NEW.id,
            cdr.id,
            NOW(),
            'STARTED',
            JSON_OBJECT(
                'client_name' => NEW.client_name,
                'case_name' => NEW.case_name,
                'case_type' => NEW.case_type,
                'opposing_parties' => NEW.opposing_parties,
                'search_depth' => NEW.search_depth,
                'search_years' => NEW.search_years
            ),
            'PENDING'
        FROM conflict_detection_rules cdr
        WHERE cdr.status = 'ACTIVE'
          AND (cdr.expiry_date IS NULL OR cdr.expiry_date >= CURDATE());
END$$
DELIMITER ;