-- 删除法条相关表结构（按依赖关系逆序删除）

-- 删除触发器
DROP TRIGGER IF EXISTS update_legal_statutes_updated_at ON legal_statutes;
DROP TRIGGER IF EXISTS update_legal_categories_updated_at ON legal_categories;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除表（按依赖关系逆序）
DROP TABLE IF EXISTS legal_statute_tags;
DROP TABLE IF EXISTS legal_tags;
DROP TABLE IF EXISTS legal_search_history;
DROP TABLE IF EXISTS user_legal_favorites;
DROP TABLE IF EXISTS legal_statute_versions;
DROP TABLE IF EXISTS legal_hierarchy;
DROP TABLE IF EXISTS legal_statutes;
DROP TABLE IF EXISTS legal_categories;