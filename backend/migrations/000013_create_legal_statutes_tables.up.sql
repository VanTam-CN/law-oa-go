-- 创建法条相关表结构
-- 法条基本信息表
CREATE TABLE IF NOT EXISTS legal_statutes (
    id SERIAL PRIMARY KEY,
    statute_number VARCHAR(100) NOT NULL UNIQUE,  -- 法条编号，如"民法典第一千二百三十四条"
    title TEXT NOT NULL,                           -- 法条标题
    content TEXT NOT NULL,                         -- 法条内容
    category_id INTEGER REFERENCES legal_categories(id),  -- 法条分类ID
    law_name VARCHAR(200) NOT NULL,                -- 法律名称，如"中华人民共和国民法典"
    chapter VARCHAR(200),                          -- 章
    section VARCHAR(200),                          -- 节
    part VARCHAR(200),                             -- 篇
    effective_date DATE,                           -- 生效日期
    expiry_date DATE,                              -- 失效日期
    publishing_authority VARCHAR(200),             -- 发布机关
    status VARCHAR(20) DEFAULT 'active',           -- 状态: active, expired, repealed
    hierarchy_level INTEGER DEFAULT 1,            -- 层级深度
    parent_statute_id INTEGER REFERENCES legal_statutes(id),  -- 父法条ID（用于层级关系）
    order_in_hierarchy INTEGER,                   -- 在同级中的排序
    tags TEXT[],                                   -- 标签数组
    keywords TEXT[],                               -- 关键词数组
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建法条分类表
CREATE TABLE IF NOT EXISTS legal_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,            -- 分类名称，如"民法", "刑法", "行政法"
    code VARCHAR(50) NOT NULL UNIQUE,             -- 分类代码
    parent_id INTEGER REFERENCES legal_categories(id),  -- 父分类ID
    level INTEGER DEFAULT 1,                      -- 分类层级
    description TEXT,                             -- 分类描述
    is_active BOOLEAN DEFAULT true,              -- 是否启用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建法条层级关系表（用于复杂的层级关系管理）
CREATE TABLE IF NOT EXISTS legal_hierarchy (
    id SERIAL PRIMARY KEY,
    ancestor_id INTEGER NOT NULL REFERENCES legal_statutes(id),  -- 祖先法条ID
    descendant_id INTEGER NOT NULL REFERENCES legal_statutes(id), -- 后代法条ID
    depth INTEGER NOT NULL,                       -- 层级深度
    path TEXT,                                    -- 层级路径
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ancestor_id, descendant_id)
);

-- 创建法条版本历史表
CREATE TABLE IF NOT EXISTS legal_statute_versions (
    id SERIAL PRIMARY KEY,
    statute_id INTEGER NOT NULL REFERENCES legal_statutes(id),  -- 法条ID
    version_number INTEGER NOT NULL,              -- 版本号
    title TEXT NOT NULL,                          -- 历史标题
    content TEXT NOT NULL,                        -- 历史内容
    effective_date DATE,                          -- 生效日期
    expiry_date DATE,                             -- 失效日期
    change_description TEXT,                      -- 变更说明
    created_by INTEGER REFERENCES users(id),      -- 创建者
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(statute_id, version_number)
);

-- 创建用户法条收藏表
CREATE TABLE IF NOT EXISTS user_legal_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    statute_id INTEGER NOT NULL REFERENCES legal_statutes(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, statute_id)
);

-- 创建法条搜索历史表
CREATE TABLE IF NOT EXISTS legal_search_history (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    search_query TEXT NOT NULL,                   -- 搜索查询
    search_filters JSONB,                         -- 搜索过滤器（JSON格式）
    result_count INTEGER DEFAULT 0,              -- 结果数量
    search_duration INTEGER,                     -- 搜索耗时（毫秒）
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建法条标签表
CREATE TABLE IF NOT EXISTS legal_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,            -- 标签名称
    color VARCHAR(7) DEFAULT '#1890ff',          -- 标签颜色
    description TEXT,                             -- 标签描述
    usage_count INTEGER DEFAULT 0,              -- 使用次数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建法条标签关联表
CREATE TABLE IF NOT EXISTS legal_statute_tags (
    id SERIAL PRIMARY KEY,
    statute_id INTEGER NOT NULL REFERENCES legal_statutes(id),
    tag_id INTEGER NOT NULL REFERENCES legal_tags(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(statute_id, tag_id)
);

-- 创建索引
-- 法条表索引
CREATE INDEX IF NOT EXISTS idx_legal_statutes_number ON legal_statutes(statute_number);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_category ON legal_statutes(category_id);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_law_name ON legal_statutes(law_name);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_effective_date ON legal_statutes(effective_date);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_status ON legal_statutes(status);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_parent ON legal_statutes(parent_statute_id);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_hierarchy ON legal_statutes(hierarchy_level);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_tags ON legal_statutes USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_keywords ON legal_statutes USING GIN(keywords);
CREATE INDEX IF NOT EXISTS idx_legal_statutes_content_search ON legal_statutes USING GIN(to_tsvector('chinese', content));

-- 分类表索引
CREATE INDEX IF NOT EXISTS idx_legal_categories_parent ON legal_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_legal_categories_code ON legal_categories(code);
CREATE INDEX IF NOT EXISTS idx_legal_categories_level ON legal_categories(level);

-- 层级关系表索引
CREATE INDEX IF NOT EXISTS idx_legal_hierarchy_ancestor ON legal_hierarchy(ancestor_id);
CREATE INDEX IF NOT EXISTS idx_legal_hierarchy_descendant ON legal_hierarchy(descendant_id);
CREATE INDEX IF NOT EXISTS idx_legal_hierarchy_depth ON legal_hierarchy(depth);

-- 版本历史表索引
CREATE INDEX IF NOT EXISTS idx_legal_statute_versions_statute ON legal_statute_versions(statute_id);
CREATE INDEX IF NOT EXISTS idx_legal_statute_versions_date ON legal_statute_versions(effective_date);

-- 用户收藏表索引
CREATE INDEX IF NOT EXISTS idx_user_legal_favorites_user ON user_legal_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_user_legal_favorites_statute ON user_legal_favorites(statute_id);

-- 搜索历史表索引
CREATE INDEX IF NOT EXISTS idx_legal_search_history_user ON legal_search_history(user_id);
CREATE INDEX IF NOT EXISTS idx_legal_search_history_created ON legal_search_history(created_at);

-- 标签表索引
CREATE INDEX IF NOT EXISTS idx_legal_tags_name ON legal_tags(name);
CREATE INDEX IF NOT EXISTS idx_legal_statute_tags_statute ON legal_statute_tags(statute_id);
CREATE INDEX IF NOT EXISTS idx_legal_statute_tags_tag ON legal_statute_tags(tag_id);

-- 插入默认的法条分类数据
INSERT INTO legal_categories (name, code, level, description) VALUES
('宪法', 'CONSTITUTION', 1, '宪法相关法律法规'),
('民法', 'CIVIL_LAW', 1, '民法相关法律法规'),
('商法', 'COMMERCIAL_LAW', 1, '商法相关法律法规'),
('刑法', 'CRIMINAL_LAW', 1, '刑法相关法律法规'),
('行政法', 'ADMINISTRATIVE_LAW', 1, '行政法相关法律法规'),
('经济法', 'ECONOMIC_LAW', 1, '经济法相关法律法规'),
('社会法', 'SOCIAL_LAW', 1, '社会法相关法律法规'),
('诉讼法', 'PROCEDURAL_LAW', 1, '诉讼程序相关法律法规'),
('国际法', 'INTERNATIONAL_LAW', 1, '国际法相关法律法规'),
('其他', 'OTHER', 1, '其他法律法规')
ON CONFLICT (code) DO NOTHING;

-- 插入默认标签数据
INSERT INTO legal_tags (name, color, description) VALUES
('常用', '#52c41a', '常用法条'),
('重要', '#ff4d4f', '重要法条'),
('最新', '#1890ff', '最新法条'),
('基础', '#722ed1', '基础法条'),
('专业', '#fa8c16', '专业法条')
ON CONFLICT (name) DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表添加更新时间触发器
CREATE TRIGGER update_legal_statutes_updated_at BEFORE UPDATE ON legal_statutes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_legal_categories_updated_at BEFORE UPDATE ON legal_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 添加表注释
COMMENT ON TABLE legal_statutes IS '法条基本信息表';
COMMENT ON TABLE legal_categories IS '法条分类表';
COMMENT ON TABLE legal_hierarchy IS '法条层级关系表';
COMMENT ON TABLE legal_statute_versions IS '法条版本历史表';
COMMENT ON TABLE user_legal_favorites IS '用户法条收藏表';
COMMENT ON TABLE legal_search_history IS '法条搜索历史表';
COMMENT ON TABLE legal_tags IS '法条标签表';
COMMENT ON TABLE legal_statute_tags IS '法条标签关联表';