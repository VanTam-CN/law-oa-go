-- 利益冲突与审批系统集成数据库迁移
-- 支持审批申请与冲突检测结果的无缝集成和自动化案件创建
-- PostgreSQL版本

-- 1. 为审批申请表添加冲突检测关联字段
ALTER TABLE approval_requests
ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(36) NULL,
ADD COLUMN IF NOT EXISTS conflict_risk_level VARCHAR(20) NULL CHECK (conflict_risk_level IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'MINIMAL')),
ADD COLUMN IF NOT EXISTS conflict_check_time TIMESTAMP NULL,
ADD COLUMN IF NOT EXISTS conflict_result JSONB NULL;

-- 2. 添加案件创建关联字段
ALTER TABLE approval_requests
ADD COLUMN IF NOT EXISTS case_created BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS created_case_id VARCHAR(36) NULL,
ADD COLUMN IF NOT EXISTS case_creation_time TIMESTAMP NULL,
ADD COLUMN IF NOT EXISTS case_creation_status VARCHAR(20) NULL CHECK (case_creation_status IN ('pending', 'in_progress', 'completed', 'failed')),
ADD COLUMN IF NOT EXISTS case_creation_error TEXT NULL,
ADD COLUMN IF NOT EXISTS case_creation_retry_count INT DEFAULT 0;

-- 3. 添加集成元数据字段
ALTER TABLE approval_requests
ADD COLUMN IF NOT EXISTS integration_type VARCHAR(20) DEFAULT 'none' CHECK (integration_type IN ('none', 'conflict', 'case', 'both')),
ADD COLUMN IF NOT EXISTS integration_metadata JSONB NULL,
ADD COLUMN IF NOT EXISTS auto_submitted BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS trigger_source VARCHAR(20) DEFAULT 'manual' CHECK (trigger_source IN ('manual', 'auto'));

-- 4. 添加工作流覆盖配置字段
ALTER TABLE approval_requests
ADD COLUMN IF NOT EXISTS workflow_override JSONB NULL,
ADD COLUMN IF NOT EXISTS conditional_approval BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS approval_conditions JSONB NULL,
ADD COLUMN IF NOT EXISTS imposed_requirements JSONB NULL;

-- 5. 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_approval_requests_conflict_check_id ON approval_requests(conflict_check_id);
CREATE INDEX IF NOT EXISTS idx_approval_requests_conflict_risk_level ON approval_requests(conflict_risk_level);
CREATE INDEX IF NOT EXISTS idx_approval_requests_conflict_check_time ON approval_requests(conflict_check_time);
CREATE INDEX IF NOT EXISTS idx_approval_requests_case_created ON approval_requests(case_created);
CREATE INDEX IF NOT EXISTS idx_approval_requests_created_case_id ON approval_requests(created_case_id);
CREATE INDEX IF NOT EXISTS idx_approval_requests_case_creation_time ON approval_requests(case_creation_time);
CREATE INDEX IF NOT EXISTS idx_approval_requests_case_creation_status ON approval_requests(case_creation_status);
CREATE INDEX IF NOT EXISTS idx_approval_requests_integration_type ON approval_requests(integration_type);
CREATE INDEX IF NOT EXISTS idx_approval_requests_auto_submitted ON approval_requests(auto_submitted);
CREATE INDEX IF NOT EXISTS idx_approval_requests_trigger_source ON approval_requests(trigger_source);
CREATE INDEX IF NOT EXISTS idx_approval_requests_conditional_approval ON approval_requests(conditional_approval);

-- 6. 创建审批-冲突检测关联表
CREATE TABLE IF NOT EXISTS approval_conflict_associations (
    id VARCHAR(36) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,
    conflict_check_id VARCHAR(36) NOT NULL,

    -- 关联状态
    association_status VARCHAR(20) DEFAULT 'pending' CHECK (association_status IN ('pending', 'active', 'superseded', 'cancelled')),
    association_type VARCHAR(20) DEFAULT 'required' CHECK (association_type IN ('required', 'optional', 'conditional')),

    -- 关联详情
    risk_level VARCHAR(20) CHECK (risk_level IN ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'MINIMAL')),
    risk_score DECIMAL(5,2),
    conflict_count INT DEFAULT 0,
    requires_approval BOOLEAN DEFAULT FALSE,

    -- 集成配置
    auto_approval BOOLEAN DEFAULT FALSE,
    approval_conditions JSONB,
    mitigation_measures JSONB,

    -- 数据映射
    data_mapping JSONB,
    mapped_fields JSONB,
    validation_errors JSONB,

    -- 审计信息
    created_by VARCHAR(36) NOT NULL,
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 索引
    INDEX idx_approval_conflict_associations_approval_request_id (approval_request_id),
    INDEX idx_approval_conflict_associations_conflict_check_id (conflict_check_id),
    INDEX idx_approval_conflict_associations_association_status (association_status),
    INDEX idx_approval_conflict_associations_risk_level (risk_level),
    INDEX idx_approval_conflict_associations_association_type (association_type),
    INDEX idx_approval_conflict_associations_created_at (created_at)
);

-- 7. 创建案件创建跟踪表
CREATE TABLE IF NOT EXISTS approval_case_creation_tracking (
    id VARCHAR(36) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,

    -- 案件创建信息
    case_id VARCHAR(36) NULL,
    case_number VARCHAR(50) NULL,
    case_type VARCHAR(100) NULL,

    -- 创建状态
    creation_status VARCHAR(20) DEFAULT 'pending' CHECK (creation_status IN ('pending', 'processing', 'completed', 'failed', 'retrying')),
    creation_step VARCHAR(100) NULL,
    progress_percentage DECIMAL(5,2) DEFAULT 0.00,

    -- 错误处理
    error_code VARCHAR(50) NULL,
    error_message TEXT NULL,
    error_details JSONB NULL,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,

    -- 数据映射
    data_mapping JSONB NULL,
    mapped_fields JSONB NULL,
    unmapped_fields JSONB NULL,

    -- 条件和限制
    applied_conditions JSONB NULL,
    imposed_requirements JSONB NULL,
    workflow_actions JSONB NULL,

    -- 审计信息
    created_by VARCHAR(36) NOT NULL,
    processed_by VARCHAR(36) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,

    -- 索引
    INDEX idx_approval_case_creation_approval_request_id (approval_request_id),
    INDEX idx_approval_case_creation_case_id (case_id),
    INDEX idx_approval_case_creation_creation_status (creation_status),
    INDEX idx_approval_case_creation_case_number (case_number),
    INDEX idx_approval_case_creation_created_at (created_at),
    INDEX idx_approval_case_creation_processed_at (processed_at)
);

-- 8. 创建集成配置表
CREATE TABLE IF NOT EXISTS approval_integration_configs (
    id VARCHAR(36) PRIMARY KEY,
    config_name VARCHAR(100) NOT NULL UNIQUE,
    config_type VARCHAR(50) NOT NULL CHECK (config_type IN ('conflict_approval', 'case_creation', 'workflow_override')),

    -- 适用范围
    applicable_approval_types JSONB,
    applicable_workflows JSONB,
    applicable_departments JSONB,
    applicable_roles JSONB,

    -- 配置规则
    trigger_rules JSONB NOT NULL,
    processing_rules JSONB NOT NULL,
    validation_rules JSONB,

    -- 工作流配置
    workflow_config JSONB NOT NULL,
    approval_config JSONB,
    notification_config JSONB,

    -- 数据映射配置
    field_mapping JSONB,
    data_transformation JSONB,
    validation_mapping JSONB,

    -- 状态和版本
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'testing')),
    version INT DEFAULT 1,
    priority INT DEFAULT 0,

    -- 条件和限制
    conditions JSONB,
    limitations JSONB,

    -- 审计信息
    created_by VARCHAR(36) NOT NULL,
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    effective_date DATE,
    expiry_date DATE,

    -- 使用统计
    usage_count INT DEFAULT 0,
    last_used_date TIMESTAMP NULL,

    -- 索引
    INDEX idx_approval_integration_configs_config_name (config_name),
    INDEX idx_approval_integration_configs_config_type (config_type),
    INDEX idx_approval_integration_configs_status (status),
    INDEX idx_approval_integration_configs_priority (priority),
    INDEX idx_approval_integration_configs_effective_date (effective_date),
    INDEX idx_approval_integration_configs_created_at (created_at),
    INDEX idx_approval_integration_configs_usage_count (usage_count)
);

-- 9. 创建视图：审批申请集成完整信息视图
CREATE OR REPLACE VIEW approval_requests_integrated AS
SELECT
    ar.*,

    -- 冲突检测关联信息
    COALESCE(aca.risk_level, 'NONE') as conflict_risk_level_display,
    COALESCE(aca.risk_score, 0.00) as conflict_risk_score,
    COALESCE(aca.conflict_count, 0) as conflict_count,
    COALESCE(aca.requires_approval, FALSE) as requires_approval_for_conflict,

    -- 案件创建状态
    COALESCE(ar.case_created, FALSE) as case_created_display,
    COALESCE(ar.case_creation_status, 'pending') as case_creation_status_display,
    COALESCE(ar.case_creation_retry_count, 0) as case_creation_retry_count_display,

    -- 集成状态
    COALESCE(ar.integration_type, 'none') as integration_type_display,
    ar.conflict_check_time as last_conflict_check,
    ar.case_creation_time as last_case_creation,

    -- 自动化状态
    COALESCE(ar.auto_submitted, FALSE) as auto_submitted_display,
    ar.trigger_source as submission_source_display,
    COALESCE(ar.conditional_approval, FALSE) as conditional_approval_display,

    -- 关联统计
    (SELECT COUNT(*) FROM approval_conflict_associations aca2
     WHERE aca2.approval_request_id = ar.id
     AND aca2.association_status = 'active') as active_conflict_associations,

    (SELECT COUNT(*) FROM approval_case_creation_tracking act
     WHERE act.approval_request_id = ar.id
     AND act.creation_status = 'completed') as completed_case_creations,

    -- 进度计算
    CASE
        WHEN ar.integration_type = 'both' THEN
            CASE
                WHEN ar.conflict_check_id IS NOT NULL AND ar.case_created THEN 100
                WHEN ar.conflict_check_id IS NOT NULL THEN 50
                ELSE 0
            END
        WHEN ar.integration_type = 'conflict' THEN
            CASE WHEN ar.conflict_check_id IS NOT NULL THEN 100 ELSE 0 END
        WHEN ar.integration_type = 'case' THEN
            CASE WHEN ar.case_created THEN 100 ELSE 0 END
        ELSE 0
    END as integration_progress_percentage,

    -- 状态汇总
    CASE
        WHEN ar.status = 'approved' AND ar.case_created THEN 'approved_with_case'
        WHEN ar.status = 'approved' AND NOT ar.case_created THEN 'approved_pending_case'
        WHEN ar.status = 'rejected' THEN 'rejected'
        WHEN ar.status = 'submitted' THEN 'in_review'
        ELSE ar.status
    END as comprehensive_status

FROM approval_requests ar
LEFT JOIN approval_conflict_associations aca ON ar.conflict_check_id = aca.conflict_check_id
    AND aca.association_status = 'active'
WHERE ar.deleted_at IS NULL;

-- 10. 创建函数：自动触发冲突检测
CREATE OR REPLACE FUNCTION auto_trigger_conflict_check(
    p_approval_request_id VARCHAR(36),
    p_client_id VARCHAR(36),
    p_case_name VARCHAR(255),
    p_case_type VARCHAR(100),
    p_user_id VARCHAR(36)
) RETURNS VARCHAR(36) AS $$
DECLARE
    v_check_id VARCHAR(36);
    v_risk_level VARCHAR(20);
    v_has_conflict BOOLEAN;
BEGIN
    -- 生成检测ID
    v_check_id := 'AUTO_' || EXTRACT(EPOCH FROM NOW())::TEXT || '_' || FLOOR(RANDOM() * 1000)::TEXT;
    v_risk_level := 'MEDIUM';
    v_has_conflict := FALSE;

    -- 更新审批申请的冲突检测关联信息
    UPDATE approval_requests
    SET
        conflict_check_id = v_check_id,
        conflict_risk_level = v_risk_level,
        conflict_check_time = NOW(),
        conflict_result = jsonb_build_object(
            'check_id', v_check_id,
            'risk_level', v_risk_level,
            'has_conflict', v_has_conflict,
            'client_id', p_client_id,
            'case_name', p_case_name,
            'case_type', p_case_type,
            'triggered_by', 'auto',
            'check_time', NOW()
        ),
        updated_at = NOW()
    WHERE id = p_approval_request_id;

    -- 创建关联记录
    INSERT INTO approval_conflict_associations (
        id, approval_request_id, conflict_check_id,
        association_status, association_type,
        risk_level, requires_approval,
        created_by, created_at
    ) VALUES (
        gen_random_uuid(), p_approval_request_id, v_check_id,
        'active', 'required',
        v_risk_level, v_has_conflict,
        p_user_id, NOW()
    );

    -- 返回检测ID
    RETURN v_check_id;
END;
$$ LANGUAGE plpgsql;

-- 11. 创建触发器函数：审批申请类型为冲突时自动触发检测
CREATE OR REPLACE FUNCTION set_auto_conflict_integration() RETURNS TRIGGER AS $$
BEGIN
    -- 如果申请类型为conflict或包含冲突相关关键词，自动设置集成类型
    IF NEW.type = 'conflict' OR NEW.category = 'conflict' THEN
        NEW.integration_type := 'conflict';
        NEW.trigger_source := 'auto';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 12. 创建触发器
DROP TRIGGER IF EXISTS auto_trigger_conflict_detection ON approval_requests;
CREATE TRIGGER auto_trigger_conflict_detection
    BEFORE INSERT ON approval_requests
    FOR EACH ROW
    EXECUTE FUNCTION set_auto_conflict_integration();

-- 13. 插入默认集成配置
INSERT INTO approval_integration_configs (
    id, config_name, config_type,
    applicable_approval_types, trigger_rules, processing_rules,
    validation_rules, workflow_config, status, created_by
) VALUES
-- 冲突检测审批集成配置
(gen_random_uuid(), 'conflict_approval_integration', 'conflict_approval',
ARRAY['conflict', 'risk_assessment']::JSONB,
jsonb_build_object('trigger_condition', jsonb_build_object('field', 'type', 'operator', 'equals', 'value', 'conflict')),
jsonb_build_object('auto_conflict_check', true, 'validate_client_history', true),
jsonb_build_object('min_risk_threshold', 0.3, 'max_conflict_cases', 10),
jsonb_build_object('approval_workflow', 'STANDARD_APPROVAL', 'escalation_rules', ARRAY['risk_level_critical', 'conflict_count_gt_5']::JSONB),
'active', 'system'),

-- 案件创建集成配置
(gen_random_uuid(), 'case_creation_integration', 'case_creation',
ARRAY['conflict', 'case_creation']::JSONB,
jsonb_build_object('trigger_condition', jsonb_build_object('field', 'status', 'operator', 'equals', 'value', 'approved')),
jsonb_build_object('auto_case_creation', true, 'data_validation', true, 'error_handling', 'retry'),
jsonb_build_object('required_fields', ARRAY['client_id', 'case_name', 'case_type']::JSONB),
jsonb_build_object('case_creation_template', jsonb_build_object('auto_populate', true, 'approval_link', true)),
'active', 'system')
ON CONFLICT (config_name) DO NOTHING;

-- 14. 创建函数：获取集成状态统计
CREATE OR REPLACE FUNCTION get_integration_stats()
RETURNS JSONB AS $$
BEGIN
    RETURN (
        SELECT jsonb_build_object(
            'total_approvals', COUNT(*),
            'conflict_integrated', SUM(CASE WHEN conflict_check_id IS NOT NULL THEN 1 ELSE 0 END),
            'case_created', SUM(CASE WHEN case_created = TRUE THEN 1 ELSE 0 END),
            'auto_submitted', SUM(CASE WHEN auto_submitted = TRUE THEN 1 ELSE 0 END),
            'integration_complete', SUM(CASE WHEN integration_type = 'both' AND conflict_check_id IS NOT NULL AND case_created = TRUE THEN 1 ELSE 0 END)
        )
        FROM approval_requests
        WHERE deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql;

-- 15. 创建更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 16. 为需要的表添加更新时间戳触发器
DROP TRIGGER IF EXISTS update_approval_conflict_associations_updated_at ON approval_conflict_associations;
CREATE TRIGGER update_approval_conflict_associations_updated_at
    BEFORE UPDATE ON approval_conflict_associations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_approval_case_creation_tracking_updated_at ON approval_case_creation_tracking;
CREATE TRIGGER update_approval_case_creation_tracking_updated_at
    BEFORE UPDATE ON approval_case_creation_tracking
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_approval_integration_configs_updated_at ON approval_integration_configs;
CREATE TRIGGER update_approval_integration_configs_updated_at
    BEFORE UPDATE ON approval_integration_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();