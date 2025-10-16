-- PostgreSQL全文搜索优化脚本
-- 为Law OA Go项目添加全文搜索索引和功能

-- 创建全文搜索配置
DO $$
BEGIN
    CREATE TEXT SEARCH CONFIGURATION law_oa_config (COPY = english);
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 为客户表添加全文搜索索引
-- 创建复合的搜索向量，包含客户姓名、邮箱、电话等字段
ALTER TABLE clients ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- 创建自动更新搜索向量的触发器函数
CREATE OR REPLACE FUNCTION update_client_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.email, '')), 'B') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.phone, '')), 'B') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.company, '')), 'C') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.industry, '')), 'D') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.notes, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器，自动更新搜索向量
DROP TRIGGER IF EXISTS update_client_search_vector_trigger ON clients;
CREATE TRIGGER update_client_search_vector_trigger
    BEFORE INSERT OR UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_client_search_vector();

-- 为案件表添加全文搜索索引
ALTER TABLE cases ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE OR REPLACE FUNCTION update_case_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.case_type, '')), 'C') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.status, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_case_search_vector_trigger ON cases;
CREATE TRIGGER update_case_search_vector_trigger
    BEFORE INSERT OR UPDATE ON cases
    FOR EACH ROW EXECUTE FUNCTION update_case_search_vector();

-- 为用户表添加全文搜索索引
ALTER TABLE users ADD COLUMN IF NOT EXISTS search_vector tsvector;

CREATE OR REPLACE FUNCTION update_user_search_vector() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.username, '')), 'B') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.email, '')), 'B') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.phone, '')), 'C') ||
        setweight(to_tsvector('law_oa_config', COALESCE(NEW.role, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_user_search_vector_trigger ON users;
CREATE TRIGGER update_user_search_vector_trigger
    BEFORE INSERT OR UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_user_search_vector();

-- 创建全文搜索索引（GIN索引用于全文搜索最高效）
CREATE INDEX IF NOT EXISTS idx_clients_search_vector ON clients USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_cases_search_vector ON cases USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_users_search_vector ON users USING GIN(search_vector);

-- 为现有数据初始化搜索向量
UPDATE clients SET search_vector =
    setweight(to_tsvector('law_oa_config', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('law_oa_config', COALESCE(email, '')), 'B') ||
    setweight(to_tsvector('law_oa_config', COALESCE(phone, '')), 'B') ||
    setweight(to_tsvector('law_oa_config', COALESCE(company, '')), 'C') ||
    setweight(to_tsvector('law_oa_config', COALESCE(industry, '')), 'D') ||
    setweight(to_tsvector('law_oa_config', COALESCE(notes, '')), 'D');

UPDATE cases SET search_vector =
    setweight(to_tsvector('law_oa_config', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('law_oa_config', COALESCE(description, '')), 'B') ||
    setweight(to_tsvector('law_oa_config', COALESCE(case_type, '')), 'C') ||
    setweight(to_tsvector('law_oa_config', COALESCE(status, '')), 'D');

UPDATE users SET search_vector =
    setweight(to_tsvector('law_oa_config', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('law_oa_config', COALESCE(username, '')), 'B') ||
    setweight(to_tsvector('law_oa_config', COALESCE(email, '')), 'B') ||
    setweight(to_tsvector('law_oa_config', COALESCE(phone, '')), 'C') ||
    setweight(to_tsvector('law_oa_config', COALESCE(role, '')), 'D');

-- 创建全文搜索函数
-- 客户搜索函数
CREATE OR REPLACE FUNCTION search_clients(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20,
    client_status TEXT DEFAULT NULL,
    client_type TEXT DEFAULT NULL
)
RETURNS TABLE (
    id BIGINT,
    name TEXT,
    email TEXT,
    phone TEXT,
    type TEXT,
    status TEXT,
    company TEXT,
    industry TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        c.id,
        c.name,
        c.email,
        c.phone,
        c.type,
        c.status,
        c.company,
        c.industry,
        c.created_at,
        c.updated_at,
        ts_rank(c.search_vector, websearch_to_tsquery('law_oa_config', search_query)) AS rank
    FROM clients c
    WHERE c.search_vector @@ websearch_to_tsquery('law_oa_config', search_query)
        AND (client_status IS NULL OR c.status = client_status)
        AND (client_type IS NULL OR c.type = client_type)
        AND c.deleted_at IS NULL
    ORDER BY rank DESC, c.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 案件搜索函数
CREATE OR REPLACE FUNCTION search_cases(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20,
    case_status TEXT DEFAULT NULL,
    case_type_param TEXT DEFAULT NULL,
    lawyer_id BIGINT DEFAULT NULL,
    client_id BIGINT DEFAULT NULL
)
RETURNS TABLE (
    id BIGINT,
    title TEXT,
    description TEXT,
    case_type TEXT,
    priority TEXT,
    status TEXT,
    client_id BIGINT,
    lawyer_id BIGINT,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    client_name TEXT,
    lawyer_name TEXT,
    rank REAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        ca.id,
        ca.title,
        ca.description,
        ca.case_type,
        ca.priority,
        ca.status,
        ca.client_id,
        ca.lawyer_id,
        ca.start_date,
        ca.end_date,
        ca.created_at,
        ca.updated_at,
        cl.name AS client_name,
        u.name AS lawyer_name,
        ts_rank(ca.search_vector, websearch_to_tsquery('law_oa_config', search_query)) AS rank
    FROM cases ca
    LEFT JOIN clients cl ON ca.client_id = cl.id
    LEFT JOIN users u ON ca.lawyer_id = u.id
    WHERE ca.search_vector @@ websearch_to_tsquery('law_oa_config', search_query)
        AND (case_status IS NULL OR ca.status = case_status)
        AND (case_type_param IS NULL OR ca.case_type = case_type_param)
        AND (lawyer_id IS NULL OR ca.lawyer_id = lawyer_id)
        AND (client_id IS NULL OR ca.client_id = client_id)
        AND ca.deleted_at IS NULL
    ORDER BY rank DESC, ca.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 用户搜索函数
CREATE OR REPLACE FUNCTION search_users(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20,
    user_status TEXT DEFAULT NULL,
    user_role TEXT DEFAULT NULL
)
RETURNS TABLE (
    id BIGINT,
    username TEXT,
    name TEXT,
    email TEXT,
    phone TEXT,
    role TEXT,
    status TEXT,
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
        u.phone,
        u.role,
        u.status,
        u.created_at,
        u.updated_at,
        ts_rank(u.search_vector, websearch_to_tsquery('law_oa_config', search_query)) AS rank
    FROM users u
    WHERE u.search_vector @@ websearch_to_tsquery('law_oa_config', search_query)
        AND (user_status IS NULL OR u.status = user_status)
        AND (user_role IS NULL OR u.role = user_role)
        AND u.deleted_at IS NULL
    ORDER BY rank DESC, u.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 创建搜索建议函数（自动完成）
CREATE OR REPLACE FUNCTION get_search_suggestions(
    prefix_param TEXT,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE (
    suggestion TEXT,
    suggestion_count BIGINT
) AS $$
BEGIN
    RETURN QUERY
    WITH search_data AS (
        SELECT name FROM users WHERE name ILIKE prefix_param || '%' AND deleted_at IS NULL
        UNION
        SELECT name FROM clients WHERE name ILIKE prefix_param || '%' AND deleted_at IS NULL
        UNION
        SELECT title FROM cases WHERE title ILIKE prefix_param || '%' AND deleted_at IS NULL
    )
    SELECT
        suggestion.name,
        COUNT(*) OVER (PARTITION BY suggestion.name) as suggestion_count
    FROM search_data suggestion
    GROUP BY suggestion.name
    ORDER BY
        CASE WHEN suggestion.name ILIKE prefix_param || '%' THEN 1 ELSE 2 END,
        LENGTH(suggestion.name),
        suggestion.name
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- 创建全文搜索统计函数
CREATE OR REPLACE FUNCTION get_search_stats()
RETURNS TABLE (
    table_name TEXT,
    total_records BIGINT,
    indexed_records BIGINT,
    last_updated TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        'clients'::TEXT,
        COUNT(*) AS total_records,
        COUNT(search_vector) AS indexed_records,
        MAX(updated_at) AS last_updated
    FROM clients

    UNION ALL

    SELECT
        'cases'::TEXT,
        COUNT(*) AS total_records,
        COUNT(search_vector) AS indexed_records,
        MAX(updated_at) AS last_updated
    FROM cases

    UNION ALL

    SELECT
        'users'::TEXT,
        COUNT(*) AS total_records,
        COUNT(search_vector) AS indexed_records,
        MAX(updated_at) AS last_updated
    FROM users;
END;
$$ LANGUAGE plpgsql;

-- 创建搜索性能分析视图
CREATE OR REPLACE VIEW search_performance AS
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE indexname LIKE '%search_vector%'
ORDER BY idx_scan DESC;

-- 创建全文搜索监控函数
CREATE OR REPLACE FUNCTION monitor_fulltext_search()
RETURNS TABLE (
    query_type TEXT,
    execution_time_ms INTEGER,
    rows_returned BIGINT,
    search_timestamp TIMESTAMP
) AS $$
BEGIN
    -- 这里可以添加实际的监控逻辑
    -- 为了演示，返回空结果
    RETURN;
END;
$$ LANGUAGE plpgsql;