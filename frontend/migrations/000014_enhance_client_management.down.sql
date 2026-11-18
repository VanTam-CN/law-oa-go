-- 回滚客户信息管理增强数据模型

-- 删除视图
DROP VIEW IF EXISTS client_complete_info;

-- 删除触发器
DROP TRIGGER IF EXISTS trigger_client_profile_creation;

-- 删除表（按依赖关系倒序）
DROP TABLE IF EXISTS client_compliance_records;
DROP TABLE IF EXISTS client_industry_classifications;
DROP TABLE IF EXISTS client_name_variants;
DROP TABLE IF EXISTS client_relationships;
DROP TABLE IF EXISTS client_profiles;