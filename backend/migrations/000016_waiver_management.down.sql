-- 回滚豁免管理数据模型

-- 删除存储过程和触发器
DROP PROCEDURE IF EXISTS generate_waiver_application_number;
DROP TRIGGER IF EXISTS increment_template_usage;

-- 删除视图
DROP VIEW IF EXISTS waiver_applications_complete;

-- 删除表（按依赖关系倒序）
DROP TABLE IF EXISTS waiver_monitoring_records;
DROP TABLE IF EXISTS waiver_signatures;
DROP TABLE IF EXISTS waiver_approval_records;
DROP TABLE IF EXISTS waiver_applications;
DROP TABLE IF EXISTS waiver_templates;