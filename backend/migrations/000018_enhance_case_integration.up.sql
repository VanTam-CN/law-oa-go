-- 增强案例集成迁移
-- 支持案例与增强客户档案系统的集成，以及冲突检测流程

-- 1. 扩展 cases 表，添加增强字段
ALTER TABLE cases
ADD COLUMN client_profile_ids JSONB DEFAULT '[]'::jsonb COMMENT '客户档案ID列表，支持多客户案件',
ADD COLUMN conflict_check_request_id VARCHAR(36) COMMENT '关联的冲突检测请求ID',
ADD COLUMN conflict_detection_status VARCHAR(20) DEFAULT 'PENDING' COMMENT '冲突检测状态: PENDING, IN_PROGRESS, COMPLETED, FAILED',
ADD COLUMN conflict_detection_result JSONB COMMENT '冲突检测结果摘要',
ADD COLUMN risk_level VARCHAR(20) COMMENT '整体风险等级: LOW, MEDIUM, HIGH, CRITICAL',
ADD COLUMN waiver_application_id VARCHAR(36) COMMENT '关联的豁免申请ID',
ADD COLUMN waiver_status VARCHAR(20) COMMENT '豁免状态: NONE, PENDING, APPROVED, REJECTED',
ADD COLUMN ethical_screen_established BOOLEAN DEFAULT FALSE COMMENT '是否建立信息壁垒',
ADD COLUMN assigned_by VARCHAR(36) COMMENT '分配者ID',
ADD COLUMN practice_area VARCHAR(50) COMMENT '业务领域',
ADD COLUMN estimated_duration VARCHAR(50) COMMENT '预估案件持续时间',
ADD COLUMN billing_method VARCHAR(50) COMMENT '计费方式',
ADD COLUMN team_assignment JSONB DEFAULT '{}'::jsonb COMMENT '团队分配信息',
ADD COLUMN conflict_metadata JSONB DEFAULT '{}'::jsonb COMMENT '冲突检测元数据',
ADD COLUMN created_via_conflict_check BOOLEAN DEFAULT FALSE COMMENT '是否通过冲突检测流程创建';

-- 2. 创建案例-客户档案关联表
CREATE TABLE case_client_profiles (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid_generate_v4()),
    case_id UINT NOT NULL,
    client_profile_id VARCHAR(36) NOT NULL,
    client_role VARCHAR(50) NOT NULL DEFAULT 'PRIMARY' COMMENT '客户在案件中的角色: PRIMARY, SECONDARY, OPPOSING, THIRD_PARTY',
    relationship_description TEXT COMMENT '客户关系描述',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 外键约束
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (client_profile_id) REFERENCES client_profiles(id) ON DELETE CASCADE,

    -- 索引
    INDEX idx_case_client_profiles_case_id (case_id),
    INDEX idx_case_client_profiles_client_id (client_profile_id),
    INDEX idx_case_client_profiles_role (client_role),
    INDEX idx_case_client_profiles_deleted_at (deleted_at),

    -- 唯一约束：同一案例中同一客户只能有一个角色
    UNIQUE KEY uk_case_client_role (case_id, client_profile_id, deleted_at)
) COMMENT '案例-客户档案关联表';

-- 3. 创建案例冲突检测记录表
CREATE TABLE case_conflict_records (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid_generate_v4()),
    case_id UINT NOT NULL,
    conflict_check_request_id VARCHAR(36) NOT NULL,
    detection_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    conflict_types_detected JSONB DEFAULT '[]'::jsonb COMMENT '检测到的冲突类型列表',
    risk_assessment JSONB NOT NULL COMMENT '风险评估结果',
    affected_parties JSONB DEFAULT '[]'::jsonb COMMENT '受影响的相关方',
    recommended_actions JSONB DEFAULT '[]'::jsonb COMMENT '建议的行动措施',
    detection_rules_applied JSONB DEFAULT '[]'::jsonb COMMENT '应用的检测规则',
    status VARCHAR(20) DEFAULT 'COMPLETED' COMMENT '检测状态',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 外键约束
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (conflict_check_request_id) REFERENCES professional_conflict_check_requests(id) ON DELETE CASCADE,

    -- 索引
    INDEX idx_case_conflict_case_id (case_id),
    INDEX idx_case_conflict_request_id (conflict_check_request_id),
    INDEX idx_case_conflict_date (detection_date),
    INDEX idx_case_conflict_status (status),
    INDEX idx_case_conflict_deleted_at (deleted_at)
) COMMENT '案例冲突检测记录表';

-- 4. 创建案例豁免关联表
CREATE TABLE case_waiver_associations (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid_generate_v4()),
    case_id UINT NOT NULL,
    waiver_application_id VARCHAR(36) NOT NULL,
    association_type VARCHAR(50) NOT NULL DEFAULT 'REQUIRED' COMMENT '关联类型: REQUIRED, ELECTIVE, PROACTIVE',
    conflict_summary TEXT NOT NULL COMMENT '冲突摘要',
    waiver_conditions JSONB DEFAULT '{}'::jsonb COMMENT '豁免条件',
    monitoring_requirements JSONB DEFAULT '{}'::jsonb COMMENT '监控要求',
    status VARCHAR(20) DEFAULT 'PENDING' COMMENT '关联状态',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 外键约束
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_application_id) REFERENCES waiver_applications(id) ON DELETE CASCADE,

    -- 索引
    INDEX idx_case_waiver_case_id (case_id),
    INDEX idx_case_waiver_waiver_id (waiver_application_id),
    INDEX idx_case_waiver_type (association_type),
    INDEX idx_case_waiver_status (status),
    INDEX idx_case_waiver_deleted_at (deleted_at)
) COMMENT '案例豁免关联表';

-- 5. 创建案例信息屏障表
CREATE TABLE case_ethical_screens (
    id VARCHAR(36) PRIMARY KEY DEFAULT (uuid_generate_v4()),
    case_id UINT NOT NULL,
    screen_type VARCHAR(50) NOT NULL COMMENT '信息屏障类型: INFORMATION_BARRIER, CHINESE_WALL, ETHICAL_WALL',
    screened_lawyers JSONB NOT NULL DEFAULT '[]'::jsonb COMMENT '被屏障的律师列表',
    screened_teams JSONB NOT NULL DEFAULT '[]'::jsonb COMMENT '被屏障的团队列表',
    restricted_information JSONB NOT NULL DEFAULT '{}'::jsonb COMMENT '受限信息范围',
    access_permissions JSONB NOT NULL DEFAULT '{}'::jsonb COMMENT '访问权限配置',
    monitoring_plan JSONB NOT NULL DEFAULT '{}'::jsonb COMMENT '监控计划',
    effective_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expiry_date TIMESTAMP WITH TIME ZONE COMMENT '到期日期，为空表示永久有效',
    status VARCHAR(20) DEFAULT 'ACTIVE' COMMENT '状态: ACTIVE, INACTIVE, EXPIRED',
    established_by VARCHAR(36) NOT NULL COMMENT '建立人ID',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- 外键约束
    FOREIGN KEY (case_id) REFERENCES cases(id) ON DELETE CASCADE,
    FOREIGN KEY (established_by) REFERENCES users(id),

    -- 索引
    INDEX idx_case_ethical_screen_case_id (case_id),
    INDEX idx_case_ethical_screen_type (screen_type),
    INDEX idx_case_ethical_screen_status (status),
    INDEX idx_case_ethical_screen_effective_date (effective_date),
    INDEX idx_case_ethical_screen_established_by (established_by),
    INDEX idx_case_ethical_screen_deleted_at (deleted_at)
) COMMENT '案例信息屏障表';

-- 6. 添加向后兼容的字段和约束
-- 保持原有的 ClientID 字段，但添加注释说明其用途
ALTER TABLE cases
ALTER COLUMN client_id COMMENT '原有客户ID，保持向后兼容，优先使用 client_profile_ids 字段';

-- 7. 创建触发器，确保数据一致性
-- 触发器：当 client_profile_ids 不为空时，自动更新 case_client_profiles 表
CREATE OR REPLACE FUNCTION sync_case_client_profiles()
RETURNS TRIGGER AS $$
BEGIN
    -- 当案例的 client_profile_ids 发生变化时
    IF TG_OP = 'UPDATE' AND (OLD.client_profile_ids IS DISTINCT FROM NEW.client_profile_ids) THEN
        -- 删除旧的关联记录（软删除）
        UPDATE case_client_profiles
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE case_id = NEW.id AND deleted_at IS NULL;

        -- 如果有新的客户档案ID，创建新的关联记录
        IF NEW.client_profile_ids IS NOT NULL AND jsonb_array_length(NEW.client_profile_ids) > 0 THEN
            INSERT INTO case_client_profiles (case_id, client_profile_id, client_role)
            SELECT
                NEW.id,
                value::text,
                'PRIMARY'
            FROM jsonb_array_elements_text(NEW.client_profile_ids);
        END IF;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_sync_case_client_profiles
    AFTER INSERT OR UPDATE ON cases
    FOR EACH ROW
    EXECUTE FUNCTION sync_case_client_profiles();

-- 8. 创建视图，简化查询
CREATE VIEW v_case_enhanced AS
SELECT
    c.*,
    -- 客户信息
    cp.client_names,
    cp.client_types,
    cp.primary_contacts,
    -- 冲突检测信息
    ccr.detection_date as last_conflict_check_date,
    ccr.conflict_types_detected,
    ccr.risk_assessment,
    -- 豁免信息
    cwa.waiver_status,
    cwa.association_type as waiver_type,
    -- 信息屏障信息
    ces.screen_type as ethical_screen_type,
    ces.status as ethical_screen_status
FROM cases c
LEFT JOIN LATERAL (
    SELECT
        array_agg(DISTINCT cp.name) as client_names,
        array_agg(DISTINCT cp.client_type) as client_types,
        array_agg(DISTINCT jsonb_build_object('name', cp.primary_contact_name, 'phone', cp.primary_contact_phone, 'email', cp.primary_contact_email)) as primary_contacts
    FROM case_client_profiles ccp
    JOIN client_profiles cp ON ccp.client_profile_id = cp.id
    WHERE ccp.case_id = c.id AND ccp.deleted_at IS NULL
) cp ON true
LEFT JOIN LATERAL (
    SELECT
        MAX(detection_date) as detection_date,
        jsonb_agg(DISTINCT conflict_types_detected) as conflict_types_detected,
        jsonb_agg(DISTINCT risk_assessment) as risk_assessment
    FROM case_conflict_records
    WHERE case_id = c.id AND deleted_at IS NULL
) ccr ON true
LEFT JOIN LATERAL (
    SELECT
        MIN(status) as waiver_status,
        MIN(association_type) as association_type
    FROM case_waiver_associations
    WHERE case_id = c.id AND deleted_at IS NULL
) cwa ON true
LEFT JOIN LATERAL (
    SELECT
        screen_type,
        status
    FROM case_ethical_screens
    WHERE case_id = c.id AND deleted_at IS NULL AND status = 'ACTIVE'
    LIMIT 1
) ces ON true
WHERE c.deleted_at IS NULL;

-- 9. 添加初始数据约束检查
-- 确保冲突检测状态的有效性
ALTER TABLE cases
ADD CONSTRAINT chk_conflict_detection_status
CHECK (conflict_detection_status IN ('PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'NOT_REQUIRED'));

-- 确保豁免状态的有效性
ALTER TABLE cases
ADD CONSTRAINT chk_waiver_status
CHECK (waiver_status IN ('NONE', 'PENDING', 'APPROVED', 'REJECTED', 'EXPIRED'));

-- 确保风险等级的有效性
ALTER TABLE cases
ADD CONSTRAINT chk_risk_level
CHECK (risk_level IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL', 'NOT_ASSESSED'));

-- 10. 创建索引以优化查询性能
-- 复合索引：支持冲突检测状态查询
CREATE INDEX idx_cases_conflict_status ON cases(conflict_detection_status, created_at) WHERE deleted_at IS NULL;

-- 复合索引：支持豁免状态查询
CREATE INDEX idx_cases_waiver_status ON cases(waiver_status, created_at) WHERE deleted_at IS NULL;

-- 复合索引：支持风险等级查询
CREATE INDEX idx_cases_risk_level ON cases(risk_level, created_at) WHERE deleted_at IS NULL;

-- 复合索引：支持多维度查询
CREATE INDEX idx_cases_composite ON cases(status, priority, conflict_detection_status, created_at) WHERE deleted_at IS NULL;

-- 11. 添加注释
COMMENT ON TABLE cases IS '案件表 - 增强版本，支持多客户、冲突检测、豁免管理等高级功能';
COMMENT ON TABLE case_client_profiles IS '案例-客户档案关联表，支持多客户案件和角色管理';
COMMENT ON TABLE case_conflict_records IS '案例冲突检测记录表，记录每次冲突检测的详细结果';
COMMENT ON TABLE case_waiver_associations IS '案例豁免关联表，管理案例相关的豁免申请和条件';
COMMENT ON TABLE case_ethical_screens IS '案例信息屏障表，管理信息屏障和访问控制';
COMMENT ON VIEW v_case_enhanced IS '案例增强视图，提供包含所有关联信息的统一查询接口';