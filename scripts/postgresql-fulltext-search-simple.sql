-- PostgreSQL全文搜索优化脚本（简化版）
-- 为Law OA Go项目添加全文搜索索引和功能

-- 为现有数据初始化搜索向量
UPDATE clients SET search_vector =
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(email, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(phone, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(company, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(industry, '')), 'D') ||
    setweight(to_tsvector('english', COALESCE(notes, '')), 'D')
WHERE search_vector IS NULL;

UPDATE cases SET search_vector =
    setweight(to_tsvector('english', COALESCE(title, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(description, '')), 'B')
WHERE search_vector IS NULL;

UPDATE users SET search_vector =
    setweight(to_tsvector('english', COALESCE(name, '')), 'A') ||
    setweight(to_tsvector('english', COALESCE(username, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(email, '')), 'B') ||
    setweight(to_tsvector('english', COALESCE(phone, '')), 'C') ||
    setweight(to_tsvector('english', COALESCE(role, '')), 'D')
WHERE search_vector IS NULL;

-- 简化的客户搜索函数
CREATE OR REPLACE FUNCTION search_clients_simple(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20
)
RETURNS TABLE (
    id BIGINT,
    name TEXT,
    email TEXT,
    phone TEXT,
    type TEXT,
    status TEXT,
    company TEXT,
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
        ts_rank(c.search_vector, plainto_tsquery('english', search_query)) AS rank
    FROM clients c
    WHERE c.search_vector @@ plainto_tsquery('english', search_query)
        AND c.deleted_at IS NULL
    ORDER BY rank DESC, c.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 简化的案件搜索函数
CREATE OR REPLACE FUNCTION search_cases_simple(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20
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
        ts_rank(ca.search_vector, plainto_tsquery('english', search_query)) AS rank
    FROM cases ca
    WHERE ca.search_vector @@ plainto_tsquery('english', search_query)
        AND ca.deleted_at IS NULL
    ORDER BY rank DESC, ca.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 简化的用户搜索函数
CREATE OR REPLACE FUNCTION search_users_simple(
    search_query TEXT,
    page_num INTEGER DEFAULT 1,
    page_size INTEGER DEFAULT 20
)
RETURNS TABLE (
    id BIGINT,
    username TEXT,
    name TEXT,
    email TEXT,
    phone TEXT,
    role TEXT,
    status TEXT,
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
        ts_rank(u.search_vector, plainto_tsquery('english', search_query)) AS rank
    FROM users u
    WHERE u.search_vector @@ plainto_tsquery('english', search_query)
        AND u.deleted_at IS NULL
    ORDER BY rank DESC, u.created_at DESC
    LIMIT page_size OFFSET (page_num - 1) * page_size;
END;
$$ LANGUAGE plpgsql;

-- 验证全文搜索配置
SELECT
    'Full-text Search Setup Verification' as status,
    'Search vectors created' as message,
    COUNT(*) as total_clients
FROM clients
WHERE search_vector IS NOT NULL;

SELECT
    'Full-text Search Setup Verification' as status,
    'Search vectors created' as message,
    COUNT(*) as total_cases
FROM cases
WHERE search_vector IS NOT NULL;

SELECT
    'Full-text Search Setup Verification' as status,
    'Search vectors created' as message,
    COUNT(*) as total_users
FROM users
WHERE search_vector IS NOT NULL;