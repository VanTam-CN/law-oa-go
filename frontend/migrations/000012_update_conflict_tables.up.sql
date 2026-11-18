-- 更新冲突检测规则表，添加缺失字段
ALTER TABLE conflict_rules
ADD COLUMN IF NOT EXISTS category VARCHAR(100) NOT NULL DEFAULT '' COMMENT '规则分类' AFTER type,
ADD COLUMN IF NOT EXISTS actions JSON COMMENT '规则动作配置' AFTER conditions,
ADD COLUMN IF NOT EXISTS version INT DEFAULT 1 COMMENT '版本号' AFTER priority,
ADD COLUMN IF NOT EXISTS mcp_source VARCHAR(255) COMMENT 'MCP来源' AFTER version;

-- 更新冲突案例表，添加缺失字段
ALTER TABLE conflict_cases
ADD COLUMN IF NOT EXISTS case_no VARCHAR(100) COMMENT '案件编号' AFTER case_name,
ADD COLUMN IF NOT EXISTS risk_level ENUM('HIGH', 'MEDIUM', 'LOW', 'MINIMAL') DEFAULT 'LOW' COMMENT '风险等级' AFTER conflict_type,
ADD COLUMN IF NOT EXISTS case_status VARCHAR(50) DEFAULT 'ACTIVE' COMMENT '案件状态' AFTER risk_level,
ADD COLUMN IF NOT EXISTS client_id VARCHAR(36) COMMENT '客户ID' AFTER case_status,
ADD COLUMN IF NOT EXISTS conflict_details TEXT COMMENT '冲突详情' AFTER opposing_parties;

-- 更新MCP标准表，添加缺失字段
ALTER TABLE mcp_standards
CHANGE COLUMN id id VARCHAR(36) PRIMARY KEY COMMENT '标准记录ID',
CHANGE COLUMN is_active active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
ADD COLUMN IF NOT EXISTS last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '最后更新时间' AFTER version,
ADD COLUMN IF NOT EXISTS standards JSON COMMENT '标准内容配置' AFTER content,
ADD COLUMN IF NOT EXISTS best_practices JSON COMMENT '最佳实践' AFTER standards,
ADD COLUMN IF NOT EXISTS compliance JSON COMMENT '合规要求' AFTER best_practices,
ADD COLUMN IF NOT EXISTS risk_thresholds JSON COMMENT '风险阈值' AFTER compliance;

-- 更新冲突检测记录表，修正字段名和类型
ALTER TABLE conflict_check_records
CHANGE COLUMN id id VARCHAR(36) PRIMARY KEY COMMENT '检查记录ID',
CHANGE COLUMN user_id user_id BIGINT UNSIGNED COMMENT '用户ID',
ADD COLUMN IF NOT EXISTS risk_level ENUM('HIGH', 'MEDIUM', 'LOW', 'MINIMAL') DEFAULT 'LOW' COMMENT '风险等级' AFTER has_conflict,
CHANGE COLUMN check_status check_status ENUM('PROCESSING', 'COMPLETED', 'FAILED') DEFAULT 'PROCESSING' COMMENT '检查状态';

-- 创建客户关联关系表（如果不存在）
CREATE TABLE IF NOT EXISTS client_relations (
    id VARCHAR(36) PRIMARY KEY COMMENT '关联关系ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    related_client_id VARCHAR(36) NOT NULL COMMENT '关联客户ID',
    relation_type ENUM('PARENT', 'SUBSIDIARY', 'SISTER', 'COMPETITOR', 'ADVERSE', 'OTHER') NOT NULL COMMENT '关系类型',
    relation_detail VARCHAR(500) COMMENT '关系详情',
    active BOOLEAN DEFAULT TRUE COMMENT '是否活跃',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_client_id (client_id),
    INDEX idx_related_client_id (related_client_id),
    INDEX idx_relation_type (relation_type),
    INDEX idx_active (active),
    UNIQUE KEY uk_client_relation (client_id, related_client_id, relation_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户关联关系表';

-- 更新冲突规则表数据，为现有记录设置默认值
UPDATE conflict_rules SET
    category = IF(category IS NULL OR category = '', 'GENERAL', category),
    version = IF(version IS NULL, 1, version),
    actions = IF(actions IS NULL, '[]', actions)
WHERE category IS NULL OR category = '' OR version IS NULL OR actions IS NULL;

-- 更新冲突案例表数据，为现有记录设置默认值
UPDATE conflict_cases SET
    risk_level = IF(risk_level IS NULL, 'LOW', risk_level),
    case_status = IF(case_status IS NULL, 'ACTIVE', case_status)
WHERE risk_level IS NULL OR case_status IS NULL;

-- 更新MCP标准表数据，为现有记录设置默认值
UPDATE mcp_standards SET
    last_updated = IF(last_updated IS NULL, NOW(), last_updated),
    standards = IF(standards IS NULL, '{}', standards),
    best_practices = IF(best_practices IS NULL, '[]', best_practices),
    compliance = IF(compliance IS NULL, '[]', compliance),
    risk_thresholds = IF(risk_thresholds IS NULL, '{}', risk_thresholds)
WHERE standards IS NULL OR best_practices IS NULL OR compliance IS NULL OR risk_thresholds IS NULL;

-- 更新冲突检测记录表数据，为现有记录设置默认值
UPDATE conflict_check_records SET
    risk_level = IF(risk_level IS NULL, 'LOW', risk_level)
WHERE risk_level IS NULL;