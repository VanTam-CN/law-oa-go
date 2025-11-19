-- 利益冲突与审批系统集成数据库迁移回滚脚本
-- 回滚之前添加的所有集成相关字段和表
-- PostgreSQL版本

-- 1. 删除触发器
DROP TRIGGER IF EXISTS auto_trigger_conflict_detection ON approval_requests;
DROP TRIGGER IF EXISTS update_approval_conflict_associations_updated_at ON approval_conflict_associations;
DROP TRIGGER IF EXISTS update_approval_case_creation_tracking_updated_at ON approval_case_creation_tracking;
DROP TRIGGER IF EXISTS update_approval_integration_configs_updated_at ON approval_integration_configs;

-- 2. 删除函数
DROP FUNCTION IF EXISTS auto_trigger_conflict_check(VARCHAR, VARCHAR, VARCHAR, VARCHAR, VARCHAR);
DROP FUNCTION IF EXISTS set_auto_conflict_integration();
DROP FUNCTION IF EXISTS get_integration_stats();
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 3. 删除视图
DROP VIEW IF EXISTS approval_requests_integrated;

-- 4. 删除数据表（按依赖关系倒序）
DROP TABLE IF EXISTS approval_integration_configs;
DROP TABLE IF EXISTS approval_case_creation_tracking;
DROP TABLE IF EXISTS approval_conflict_associations;

-- 5. 删除审批申请表的集成字段（按创建顺序倒序）
ALTER TABLE approval_requests
DROP COLUMN IF EXISTS imposed_requirements,
DROP COLUMN IF EXISTS approval_conditions,
DROP COLUMN IF EXISTS conditional_approval,
DROP COLUMN IF EXISTS workflow_override,
DROP COLUMN IF EXISTS trigger_source,
DROP COLUMN IF EXISTS auto_submitted,
DROP COLUMN IF EXISTS integration_metadata,
DROP COLUMN IF EXISTS integration_type,
DROP COLUMN IF EXISTS case_creation_retry_count,
DROP COLUMN IF EXISTS case_creation_error,
DROP COLUMN IF EXISTS case_creation_status,
DROP COLUMN IF EXISTS case_creation_time,
DROP COLUMN IF EXISTS created_case_id,
DROP COLUMN IF EXISTS case_created,
DROP COLUMN IF EXISTS conflict_result,
DROP COLUMN IF EXISTS conflict_check_time,
DROP COLUMN IF EXISTS conflict_risk_level,
DROP COLUMN IF EXISTS conflict_check_id;