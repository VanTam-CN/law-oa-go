-- 回滚增强冲突检测引擎数据模型

-- 删除存储过程和触发器
DROP PROCEDURE IF EXISTS generate_professional_check_number;
DROP TRIGGER IF EXISTS auto_execute_conflict_detection;

-- 删除视图
DROP VIEW IF EXISTS professional_conflict_check_stats;

-- 删除表（按依赖关系倒序）
DROP TABLE IF EXISTS conflict_rule_executions;
DROP TABLE IF EXISTS conflict_detection_rules;
DROP TABLE IF EXISTS multi_dimensional_conflict_results;
DROP TABLE IF EXISTS professional_conflict_check_requests;