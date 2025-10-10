-- 删除客户关联关系表
DROP TABLE IF EXISTS client_relations;

-- 删除MCP标准表中新增的字段
ALTER TABLE mcp_standards
DROP COLUMN IF EXISTS last_updated,
DROP COLUMN IF EXISTS standards,
DROP COLUMN IF EXISTS best_practices,
DROP COLUMN IF EXISTS compliance,
DROP COLUMN IF EXISTS risk_thresholds;

-- 恢复MCP标准表的原始字段名和类型
ALTER TABLE mcp_standards
CHANGE COLUMN active is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用';

-- 删除冲突检测记录表中新增的字段
ALTER TABLE conflict_check_records
DROP COLUMN IF EXISTS risk_level;

-- 恢复冲突检测记录表的原始字段名和类型
ALTER TABLE conflict_check_records
CHANGE COLUMN id id VARCHAR(36) PRIMARY KEY COMMENT '检查记录ID',
CHANGE COLUMN user_id user_id BIGINT UNSIGNED COMMENT '用户ID',
CHANGE COLUMN check_status check_status VARCHAR(50) DEFAULT 'PROCESSING' COMMENT '检查状态';

-- 删除冲突案例表中新增的字段
ALTER TABLE conflict_cases
DROP COLUMN IF EXISTS case_no,
DROP COLUMN IF EXISTS risk_level,
DROP COLUMN IF EXISTS case_status,
DROP COLUMN IF EXISTS client_id,
DROP COLUMN IF EXISTS conflict_details;

-- 删除冲突检测规则表中新增的字段
ALTER TABLE conflict_rules
DROP COLUMN IF EXISTS category,
DROP COLUMN IF EXISTS actions,
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS mcp_source;