-- 回滚审批系统数据库表结构

-- 删除视图
DROP VIEW IF EXISTS approval_requests_complete;

-- 删除触发器
DROP TRIGGER IF EXISTS increment_approval_template_usage;

-- 删除存储过程
DROP PROCEDURE IF EXISTS generate_approval_request_number;

-- 删除表（按照依赖关系顺序）
DROP TABLE IF EXISTS approval_notifications;
DROP TABLE IF EXISTS approval_records;
DROP TABLE IF EXISTS approval_templates;
DROP TABLE IF EXISTS approval_workflows;
DROP TABLE IF EXISTS approval_requests;