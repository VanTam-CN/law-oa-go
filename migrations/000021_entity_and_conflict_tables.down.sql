-- ============================================================
-- Law OA Go 数据库迁移文件
-- 版本: 000021 (DOWN)
-- 描述: 回滚实体管理、实体关系、案件当事人及冲突检测核心表
-- 注意: 按照外键依赖关系的逆序删除
-- ============================================================

-- 1. 删除冲突详情表（依赖 conflict_checks, entities, cases, users）
DROP TABLE IF EXISTS conflict_details;

-- 2. 删除冲突检查表（依赖 cases, users）
DROP TABLE IF EXISTS conflict_checks;

-- 3. 删除案件当事人表（依赖 cases, entities）
DROP TABLE IF EXISTS case_parties;

-- 4. 删除实体名称变更历史表（依赖 entities）
DROP TABLE IF EXISTS entity_name_history;

-- 5. 删除实体关系表（依赖 entities）
DROP TABLE IF EXISTS entity_relations;

-- 6. 删除实体表（被其他表引用的基础表，必须最后删除）
DROP TABLE IF EXISTS entities;
