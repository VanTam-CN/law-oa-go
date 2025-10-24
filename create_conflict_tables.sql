-- 创建缺失的冲突检测相关表 - 简化版本

-- 创建冲突检测记录表
CREATE TABLE IF NOT EXISTS conflict_check_records (
    check_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    check_status VARCHAR(20) DEFAULT 'PROCESSING',
    has_conflict BOOLEAN DEFAULT FALSE,
    risk_level VARCHAR(20) DEFAULT 'LOW',
    search_parameters JSONB,
    check_result JSONB,
    user_id BIGINT,
    check_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建客户关联关系表
CREATE TABLE IF NOT EXISTS client_relations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    related_client_id UUID NOT NULL,
    relation_type VARCHAR(50) NOT NULL,
    relation_strength DECIMAL(3,2) DEFAULT 1.00,
    description TEXT,
    relation_detail VARCHAR(500),
    active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (client_id, related_client_id, relation_type)
);

-- 创建冲突案例表
CREATE TABLE IF NOT EXISTS conflict_cases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    conflict_type VARCHAR(100) NOT NULL,
    risk_level VARCHAR(20) DEFAULT 'LOW',
    description TEXT,
    opposing_parties JSONB,
    related_lawyers JSONB,
    case_no VARCHAR(100),
    case_status VARCHAR(50) DEFAULT 'ACTIVE',
    conflict_details TEXT,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建冲突检测规则表
CREATE TABLE IF NOT EXISTS conflict_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT '',
    description TEXT,
    priority INTEGER DEFAULT 1,
    version INTEGER DEFAULT 1,
    mcp_source VARCHAR(255),
    active BOOLEAN DEFAULT TRUE,
    conditions JSONB,
    actions JSONB,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_client_id ON conflict_check_records(client_id);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_check_status ON conflict_check_records(check_status);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_has_conflict ON conflict_check_records(has_conflict);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_risk_level ON conflict_check_records(risk_level);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_check_time ON conflict_check_records(check_time);
CREATE INDEX IF NOT EXISTS idx_conflict_check_records_user_id ON conflict_check_records(user_id);

CREATE INDEX IF NOT EXISTS idx_client_relations_client_id ON client_relations(client_id);
CREATE INDEX IF NOT EXISTS idx_client_relations_related_client_id ON client_relations(related_client_id);
CREATE INDEX IF NOT EXISTS idx_client_relations_relation_type ON client_relations(relation_type);

CREATE INDEX IF NOT EXISTS idx_conflict_cases_client_id ON conflict_cases(client_id);
CREATE INDEX IF NOT EXISTS idx_conflict_cases_case_type ON conflict_cases(case_type);
CREATE INDEX IF NOT EXISTS idx_conflict_cases_risk_level ON conflict_cases(risk_level);

CREATE INDEX IF NOT EXISTS idx_conflict_rules_type ON conflict_rules(type);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_active ON conflict_rules(active);
CREATE INDEX IF NOT EXISTS idx_conflict_rules_priority ON conflict_rules(priority);

-- 插入默认冲突检测规则
INSERT INTO conflict_rules (id, name, type, category, description, priority, active, conditions, actions) VALUES
(uuid_generate_v4(), '姓名相似性检测', 'NAME_SIMILARITY', 'GENERAL', '检测客户名称与历史案件的相似性', 5, TRUE, '{"threshold": 0.8, "algorithm": "levenshtein"}', '[]'),
(uuid_generate_v4(), '企业关联检测', 'CORPORATE_RELATION', 'GENERAL', '检测企业客户的关联关系冲突', 8, TRUE, '{"checkTypes": ["PARENT", "SUBSIDIARY", "SISTER"]}', '[]'),
(uuid_generate_v4(), '案件冲突检测', 'CASE_CONFLICT', 'GENERAL', '检测同一客户的案件冲突', 7, TRUE, '{"allowMultipleCases": false, "timeWindow": 365}', '[]'),
(uuid_generate_v4(), '对立当事人检测', 'ADVERSE_PARTY', 'GENERAL', '检测对立当事人冲突', 9, TRUE, '{"strictMode": true, "timeWindow": 1095}', '[]'),
(uuid_generate_v4(), '时间重叠检测', 'TIME_OVERLAP', 'GENERAL', '检测案件时间重叠', 3, TRUE, '{"overlapThreshold": 30, "unit": "days"}', '[]')
ON CONFLICT DO NOTHING;