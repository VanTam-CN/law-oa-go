-- 增强冲突检测功能数据库迁移脚本
-- 创建时间: 2025-10-16
-- 版本: v2.1.0

-- 1. 创建行业分类表
CREATE TABLE IF NOT EXISTS industry_classifications (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    parent_id INTEGER REFERENCES industry_classifications(id),
    level INTEGER DEFAULT 1,
    description TEXT,
    keywords TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. 创建竞争关系表
CREATE TABLE IF NOT EXISTS competitive_relations (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industry_classifications(id),
    competitor_type VARCHAR(50) NOT NULL, -- direct, indirect, substitute
    competitor_name VARCHAR(200) NOT NULL,
    competitor_pattern TEXT NOT NULL,
    conflict_level VARCHAR(20) NOT NULL, -- HIGH, MEDIUM, LOW
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(industry_id, competitor_name)
);

-- 3. 创建冲突规则表
CREATE TABLE IF NOT EXISTS conflict_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    rule_type VARCHAR(50) NOT NULL, -- industry_competition, client_conflict, case_type_conflict, etc.
    trigger_pattern TEXT,
    action_type VARCHAR(50) NOT NULL, -- block, warn, notify
    risk_score INTEGER DEFAULT 50,
    conditions TEXT, -- JSON格式
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(name, rule_type)
);

-- 4. 创建冲突检测历史表
CREATE TABLE IF NOT EXISTS conflict_detection_history (
    id SERIAL PRIMARY KEY,
    lawyer_id INTEGER NOT NULL REFERENCES users(id),
    case_id INTEGER REFERENCES cases(id),
    client_name VARCHAR(200) NOT NULL,
    opposing_party VARCHAR(200) NOT NULL,
    case_type VARCHAR(50) NOT NULL,
    detection_result TEXT, -- JSON格式
    conflicts_found INTEGER DEFAULT 0,
    risk_level VARCHAR(20) NOT NULL,
    user_action VARCHAR(50), -- pending, approved, rejected
    created_at TIMESTAMP DEFAULT NOW()
);

-- 5. 创建冲突案件关联表
CREATE TABLE IF NOT EXISTS conflict_case_conflicts (
    id SERIAL PRIMARY KEY,
    case_id INTEGER NOT NULL REFERENCES cases(id),
    conflict_case_id INTEGER NOT NULL REFERENCES cases(id),
    conflict_type VARCHAR(50) NOT NULL, -- direct_client, industry_competition, name_similarity
    risk_score INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(case_id, conflict_case_id, conflict_type)
);

-- 6. 修改现有表结构

-- 为客户表添加行业字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'clients' AND column_name = 'industry') THEN
        ALTER TABLE clients ADD COLUMN industry VARCHAR(100);
        CREATE INDEX idx_clients_industry ON clients(industry);
    END IF;
END $$;

-- 为案件表添加对方当事人字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'cases' AND column_name = 'opposing_party') THEN
        ALTER TABLE cases ADD COLUMN opposing_party TEXT;
        CREATE INDEX idx_cases_opposing_party ON cases USING gin(to_tsvector('simple', opposing_party));
    END IF;
END $$;

-- 为案件表添加优先级字段（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'cases' AND column_name = 'priority') THEN
        ALTER TABLE cases ADD COLUMN priority VARCHAR(20) DEFAULT 'medium';
        CREATE INDEX idx_cases_priority ON cases(priority);
    END IF;
END $$;

-- 7. 创建索引以提高查询性能

-- 行业分类索引
CREATE INDEX IF NOT EXISTS idx_industry_classifications_code ON industry_classifications(code);
CREATE INDEX IF NOT EXISTS idx_industry_classifications_active ON industry_classifications(is_active);
CREATE INDEX IF NOT EXISTS idx_industry_classifications_level ON industry_classifications(level);

-- 竞争关系索引
CREATE INDEX IF NOT EXISTS idx_competitive_relations_industry ON competitive_relations(industry_id);
CREATE INDEX IF NOT EXISTS idx_competitive_relations_active ON competitive_relations(is_active);
CREATE INDEX IF NOT EXISTS idx_competitive_relations_level ON competitive_relations(conflict_level);

-- 冲突规则索引
CREATE INDEX IF NOT EXISTS idx_conflict_rules_active ON conflict_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_priority ON conflict_rules(priority);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_type ON conflict_rules(rule_type);

-- 冲突检测历史索引
CREATE INDEX IF NOT EXISTS idx_conflict_history_lawyer ON conflict_detection_history(lawyer_id);
CREATE INDEX IF NOT EXISTS idx_conflict_history_created ON conflict_detection_history(created_at);
CREATE INDEX IF NOT EXISTS idx_conflict_history_risk_level ON conflict_detection_history(risk_level);

-- 8. 启用扩展（PostgreSQL全文搜索和模糊匹配）
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS fuzzystrmatch;

-- 9. 创建全文搜索配置
CREATE TEXT SEARCH CONFIGURATION IF NOT EXISTS chinese (COPY = simple);

-- 10. 插入初始数据

-- 插入主要行业分类
INSERT INTO industry_classifications (code, name, level, keywords) VALUES
('TMT', '科技、媒体和通信', 1, '互联网,科技,通信,软件,游戏,电商,社交媒体,短视频,云计算,AI,人工智能,5G,芯片,半导体'),
('FINANCE', '金融', 1, '银行,保险,证券,基金,支付,借贷,投资,理财,金融科技,数字货币,区块链'),
('REAL_ESTATE', '房地产', 1, '房地产,建筑,物业,装修,家居,建材,城市规划,房地产开发'),
('MANUFACTURING', '制造业', 1, '制造,工厂,生产,加工,机械,电子,汽车,化工,纺织,食品加工'),
('RETAIL', '零售', 1, '零售,超市,商店,连锁,便利店,购物中心,百货商场,专业零售'),
('HEALTHCARE', '医疗健康', 1, '医院,诊所,医药,生物技术,医疗设备,健康,医疗器械,制药'),
('EDUCATION', '教育', 1, '教育,培训,学校,大学,在线教育,职业教育,基础教育'),
('TRANSPORTATION', '交通运输', 1, '交通,运输,物流,航空,铁路,公路,航运,快递'),
('ENERGY', '能源', 1, '能源,电力,石油,天然气,新能源,太阳能,风能,核能'),
('OTHER', '其他', 1, '其他行业,综合,多元化')
ON CONFLICT (code) DO NOTHING;

-- 插入TMT行业竞争关系
INSERT INTO competitive_relations (industry_id, competitor_type, competitor_name, competitor_pattern, conflict_level, description)
SELECT
    ic.id,
    'direct',
    unnest(ARRAY['阿里巴巴', '腾讯', '字节跳动', '百度', '京东', '美团', '拼多多', '网易']),
    unnest(ARRAY[
        '阿里巴巴|阿里|淘宝|天猫|支付宝|蚂蚁金服|蚂蚁集团|阿里云|菜鸟网络',
        '腾讯|微信|QQ|腾讯云|腾讯游戏|腾讯视频|腾讯音乐|微信支付',
        '字节跳动|抖音|TikTok|今日头条|西瓜视频|飞书|懂车帝',
        '百度|百度搜索|百度地图|百度网盘|百度贴吧|百度知道|百度AI',
        '京东|JD.com|京东商城|京东物流|京东数科|京东健康',
        '美团|美团外卖|大众点评|美团打车|美团买菜',
        '拼多多|PDD|Temu',
        '网易|网易游戏|网易云音乐|网易邮箱|网易新闻'
    ]),
    unnest(ARRAY['HIGH', 'HIGH', 'HIGH', 'MEDIUM', 'MEDIUM', 'MEDIUM', 'MEDIUM', 'MEDIUM']),
    unnest(ARRAY[
        '阿里巴巴集团及其关联公司，电商和云计算巨头',
        '腾讯公司及其关联公司，社交和游戏巨头',
        '字节跳动公司及其产品，短视频和资讯平台',
        '百度公司及其产品，搜索引擎和AI公司',
        '京东集团，电商和物流服务商',
        '美团公司，本地生活服务平台',
        '拼多多，社交电商平台',
        '网易公司，游戏和互联网服务'
    ])
FROM industry_classifications ic
WHERE ic.code = 'TMT'
ON CONFLICT (industry_id, competitor_name) DO NOTHING;

-- 插入基础冲突规则
INSERT INTO conflict_rules (name, rule_type, trigger_pattern, action_type, risk_score, conditions, priority) VALUES
('直接客户冲突检测', 'client_conflict', 'same_client', 'block', 100, '{"same_client": true, "different_case": true}', 1),
('对方当事人冲突检测', 'client_conflict', 'adverse_party', 'block', 85, '{"adverse_party": true}', 2),
('行业直接竞争检测', 'industry_competition', 'direct', 'warn', 90, '{"competitor_type": "direct", "same_lawyer": true}', 3),
('行业间接竞争检测', 'industry_competition', 'indirect', 'warn', 60, '{"competitor_type": "indirect", "same_lawyer": true}', 4),
('案件类型冲突检测', 'case_type_conflict', 'same_type', 'warn', 40, '{"same_case_type": true}', 5),
('时间接近冲突检测', 'time_proximity', 'recent', 'warn', 30, '{"time_proximity_days": 30}', 6),
('高风险阈值检测', 'risk_threshold', 'high', 'block', 80, '{"risk_score": 80}', 7),
('中风险阈值检测', 'risk_threshold', 'medium', 'warn', 50, '{"risk_score": 50}', 8)
ON CONFLICT (name, rule_type) DO NOTHING;

-- 11. 创建更新时间戳的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 12. 创建触发器自动更新 updated_at 字段
DROP TRIGGER IF EXISTS update_industry_classifications_updated_at ON industry_classifications;
CREATE TRIGGER update_industry_classifications_updated_at
    BEFORE UPDATE ON industry_classifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_competitive_relations_updated_at ON competitive_relations;
CREATE TRIGGER update_competitive_relations_updated_at
    BEFORE UPDATE ON competitive_relations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_conflict_rules_updated_at ON conflict_rules;
CREATE TRIGGER update_conflict_rules_updated_at
    BEFORE UPDATE ON conflict_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 13. 创建视图用于复杂查询
CREATE OR REPLACE VIEW conflict_summary AS
SELECT
    cdh.lawyer_id,
    cdh.client_name,
    cdh.opposing_party,
    cdh.case_type,
    cdh.conflicts_found,
    cdh.risk_level,
    cdh.user_action,
    cdh.created_at,
    u.name as lawyer_name,
    COUNT(*) OVER (PARTITION BY cdh.lawyer_id) as total_checks
FROM conflict_detection_history cdh
JOIN users u ON cdh.lawyer_id = u.id
WHERE cdh.created_at >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY cdh.created_at DESC;

-- 14. 创建统计信息函数
CREATE OR REPLACE FUNCTION get_conflict_statistics(
    p_lawyer_id INTEGER DEFAULT NULL,
    p_start_date DATE DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_end_date DATE DEFAULT CURRENT_DATE
)
RETURNS TABLE(
    total_checks BIGINT,
    conflicts_detected BIGINT,
    high_risk_cases BIGINT,
    medium_risk_cases BIGINT,
    low_risk_cases BIGINT,
    most_common_client TEXT,
    most_common_case_type TEXT
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COUNT(*) as total_checks,
        SUM(CASE WHEN cdh.conflicts_found > 0 THEN 1 ELSE 0 END) as conflicts_detected,
        SUM(CASE WHEN cdh.risk_level = 'HIGH' THEN 1 ELSE 0 END) as high_risk_cases,
        SUM(CASE WHEN cdh.risk_level = 'MEDIUM' THEN 1 ELSE 0 END) as medium_risk_cases,
        SUM(CASE WHEN cdh.risk_level = 'LOW' THEN 1 ELSE 0 END) as low_risk_cases,
        mode() WITHIN GROUP (ORDER BY cdh.client_name) as most_common_client,
        mode() WITHIN GROUP (ORDER BY cdh.case_type) as most_common_case_type
    FROM conflict_detection_history cdh
    WHERE (p_lawyer_id IS NULL OR cdh.lawyer_id = p_lawyer_id)
      AND cdh.created_at >= p_start_date
      AND cdh.created_at <= p_end_date;
END;
$$ LANGUAGE plpgsql;

-- 15. 添加注释
COMMENT ON TABLE industry_classifications IS '行业分类表，用于支持行业竞争冲突检测';
COMMENT ON TABLE competitive_relations IS '竞争关系表，定义企业间的竞争关系';
COMMENT ON TABLE conflict_rules IS '冲突检测规则表，定义各种冲突检测规则';
COMMENT ON TABLE conflict_detection_history IS '冲突检测历史表，记录每次检测的结果';
COMMENT ON TABLE conflict_case_conflicts IS '冲突案件关联表，记录案件间的冲突关系';

COMMENT ON COLUMN competitive_relations.competitor_type IS '竞争者类型：direct(直接竞争), indirect(间接竞争), substitute(替代品)';
COMMENT ON COLUMN competitive_relations.conflict_level IS '冲突等级：HIGH(高风险), MEDIUM(中风险), LOW(低风险)';
COMMENT ON COLUMN conflict_rules.action_type IS '动作类型：block(阻止), warn(警告), notify(通知)';
COMMENT ON COLUMN conflict_rules.conditions IS '规则条件，JSON格式存储';
COMMENT ON COLUMN conflict_detection_history.user_action IS '用户操作：pending(待处理), approved(已批准), rejected(已拒绝)';

-- 迁移完成标记
DO $$
BEGIN
    RAISE NOTICE '=== 增强冲突检测功能数据库迁移完成 ===';
    RAISE NOTICE '已创建表: industry_classifications, competitive_relations, conflict_rules, conflict_detection_history, conflict_case_conflicts';
    RAISE NOTICE '已添加索引和触发器';
    RAISE NOTICE '已插入初始数据';
    RAISE NOTICE '迁移完成时间: %', NOW();
END $$;