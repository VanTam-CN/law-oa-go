-- Supabase 兼容性扩展脚本
-- 创建与Supabase兼容的扩展和函数

-- 启用必要的PostgreSQL扩展
CREATE EXTENSION IF NOT EXISTS "pgcrypto";        -- 用于加密
CREATE EXTENSION IF NOT EXISTS "pgjwt";           -- 用于JWT处理
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";        -- UUID生成
CREATE EXTENSION IF NOT EXISTS "btree_gin";        -- 用于全文搜索索引
CREATE EXTENSION IF NOT EXISTS "btree_gist";       -- GiST索引
CREATE EXTENSION IF NOT EXISTS "fuzzystrmatch";    -- 模糊匹配
CREATE EXTENSION IF NOT EXISTS "pg_trgm";          -- 三元组匹配

-- 创建Supabase风格的函数

-- UUID生成函数
CREATE OR REPLACE FUNCTION gen_random_uuid() RETURNS UUID
AS $$
BEGIN
    RETURN uuid_generate_v4();
END;
$$ LANGUAGE plpgsql;

-- 创建审计触发器函数
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建软删除函数
CREATE OR REPLACE FUNCTION soft_delete()
RETURNS TRIGGER AS $$
BEGIN
    NEW.deleted_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建行级安全策略辅助函数
CREATE OR REPLACE FUNCTION current_user_id()
RETURNS UUID AS $$
BEGIN
    -- 这里应该从JWT token中提取用户ID
    -- 暂时返回NULL，实际使用时需要与Supabase auth集成
    RETURN NULL::UUID;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 创建用户权限检查函数
CREATE OR REPLACE FUNCTION user_is_admin()
RETURNS BOOLEAN AS $$
BEGIN
    -- 检查当前用户是否为管理员
    -- 这里需要根据实际的认证系统进行调整
    RETURN false;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 创建数据版本控制表
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(50) NOT NULL, -- INSERT, UPDATE, DELETE
    user_id UUID,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建审计日志触发器函数
CREATE OR REPLACE FUNCTION audit_trigger()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO audit_logs (
        table_name,
        operation,
        user_id,
        old_values,
        new_values
    ) VALUES (
        TG_TABLE_NAME,
        TG_OP,
        current_user_id(),
        CASE WHEN TG_OP IN ('UPDATE', 'DELETE') THEN row_to_json(OLD) ELSE NULL END,
        CASE WHEN TG_OP IN ('INSERT', 'UPDATE') THEN row_to_json(NEW) ELSE NULL END
    );
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- 为主要表创建审计触发器
CREATE TRIGGER audit_users
    AFTER INSERT OR UPDATE OR DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER audit_clients
    AFTER INSERT OR UPDATE OR DELETE ON clients
    FOR EACH ROW EXECUTE FUNCTION audit_trigger();

CREATE TRIGGER audit_cases
    AFTER INSERT OR UPDATE OR DELETE ON cases
    FOR EACH ROW EXECUTE FUNCTION audit_trigger();

-- 创建视图用于前端查询
CREATE OR REPLACE VIEW users_view AS
SELECT
    id,
    username,
    name,
    email,
    role,
    phone,
    avatar,
    status,
    created_at,
    updated_at
FROM users
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW clients_view AS
SELECT
    id,
    name,
    type,
    email,
    phone,
    address,
    company,
    id_card,
    industry,
    contact_person,
    contact_phone,
    source,
    notes,
    status,
    created_at,
    updated_at
FROM clients
WHERE deleted_at IS NULL;

CREATE OR REPLACE VIEW cases_view AS
SELECT
    c.id,
    c.title,
    c.description,
    c.client_id,
    cl.name as client_name,
    c.lawyer_id,
    u.name as lawyer_name,
    c.case_type,
    c.priority,
    c.status,
    c.start_date,
    c.end_date,
    c.created_at,
    c.updated_at
FROM cases c
LEFT JOIN clients cl ON c.client_id = cl.id
LEFT JOIN users u ON c.lawyer_id = u.id
WHERE c.deleted_at IS NULL
  AND cl.deleted_at IS NULL
  AND u.deleted_at IS NULL;

-- 创建全文搜索索引
CREATE INDEX idx_users_fulltext ON users USING gin(to_tsvector('chinese', name || ' ' || email));
CREATE INDEX idx_clients_fulltext ON clients USING gin(to_tsvector('chinese', name || ' ' || email || ' ' || company));
CREATE INDEX idx_cases_fulltext ON cases USING gin(to_tsvector('chinese', title || ' ' || description));

-- 创建搜索函数
CREATE OR REPLACE FUNCTION search_users(query_text TEXT)
RETURNS TABLE (
    id INTEGER,
    username VARCHAR(50),
    name VARCHAR(100),
    email VARCHAR(100),
    role VARCHAR(50),
    phone VARCHAR(20),
    avatar VARCHAR(255),
    status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        u.id,
        u.username,
        u.name,
        u.email,
        u.role,
        u.phone,
        u.avatar,
        u.status,
        u.created_at,
        u.updated_at,
        ts_rank(to_tsvector('chinese', u.name || ' ' || u.email), plainto_tsquery('chinese', query_text)) as rank
    FROM users u
    WHERE u.deleted_at IS NULL
      AND to_tsvector('chinese', u.name || ' ' || u.email) @@ plainto_tsquery('chinese', query_text)
    ORDER BY rank DESC;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION search_clients(query_text TEXT)
RETURNS TABLE (
    id INTEGER,
    name VARCHAR(100),
    type VARCHAR(20),
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    company VARCHAR(100),
    status VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.name,
        c.type,
        c.email,
        c.phone,
        c.address,
        c.company,
        c.status,
        c.created_at,
        c.updated_at,
        ts_rank(to_tsvector('chinese', c.name || ' ' || c.email || ' ' || c.company), plainto_tsquery('chinese', query_text)) as rank
    FROM clients c
    WHERE c.deleted_at IS NULL
      AND to_tsvector('chinese', c.name || ' ' || c.email || ' ' || c.company) @@ plainto_tsquery('chinese', query_text)
    ORDER BY rank DESC;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION search_cases(query_text TEXT)
RETURNS TABLE (
    id INTEGER,
    title VARCHAR(200),
    description TEXT,
    client_id INTEGER,
    client_name VARCHAR(100),
    lawyer_id INTEGER,
    lawyer_name VARCHAR(100),
    case_type VARCHAR(50),
    priority VARCHAR(20),
    status VARCHAR(20),
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.title,
        c.description,
        c.client_id,
        cl.name as client_name,
        c.lawyer_id,
        u.name as lawyer_name,
        c.case_type,
        c.priority,
        c.status,
        c.start_date,
        c.end_date,
        c.created_at,
        c.updated_at,
        ts_rank(to_tsvector('chinese', c.title || ' ' || c.description), plainto_tsquery('chinese', query_text)) as rank
    FROM cases c
    LEFT JOIN clients cl ON c.client_id = cl.id
    LEFT JOIN users u ON c.lawyer_id = u.id
    WHERE c.deleted_at IS NULL
      AND cl.deleted_at IS NULL
      AND u.deleted_at IS NULL
      AND to_tsvector('chinese', c.title || ' ' || c.description) @@ plainto_tsquery('chinese', query_text)
    ORDER BY rank DESC;
END;
$$ LANGUAGE plpgsql;

-- 插入初始版本信息
INSERT INTO schema_migrations (version) VALUES ('2024-01-01-initial-setup')
ON CONFLICT (version) DO NOTHING;

-- 设置注释
COMMENT ON FUNCTION search_users IS '用户全文搜索函数';
COMMENT ON FUNCTION search_clients IS '客户全文搜索函数';
COMMENT ON FUNCTION search_cases IS '案件全文搜索函数';
COMMENT ON VIEW users_view IS '用户查询视图（已过滤删除数据）';
COMMENT ON VIEW clients_view IS '客户查询视图（已过滤删除数据）';
COMMENT ON VIEW cases_view IS '案件查询视图（已过滤删除数据，包含关联信息）';