-- 创建冲突检测案例表
CREATE TABLE IF NOT EXISTS conflict_cases (
    id VARCHAR(36) PRIMARY KEY COMMENT '冲突案例ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    case_name VARCHAR(255) NOT NULL COMMENT '案件名称',
    case_type VARCHAR(100) NOT NULL COMMENT '案件类型',
    conflict_type VARCHAR(100) NOT NULL COMMENT '冲突类型',
    risk_level ENUM('HIGH', 'MEDIUM', 'LOW', 'MINIMAL') DEFAULT 'LOW' COMMENT '风险等级',
    description TEXT COMMENT '冲突描述',
    opposing_parties JSON COMMENT '对立当事人信息',
    related_lawyers JSON COMMENT '相关律师信息',
    created_by VARCHAR(36) COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_client_id (client_id),
    INDEX idx_case_type (case_type),
    INDEX idx_conflict_type (conflict_type),
    INDEX idx_risk_level (risk_level),
    INDEX idx_created_at (created_at),
    INDEX idx_client_case (client_id, case_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测案例表';

-- 创建冲突检测规则表
CREATE TABLE IF NOT EXISTS conflict_rules (
    id VARCHAR(36) PRIMARY KEY COMMENT '规则ID',
    name VARCHAR(255) NOT NULL COMMENT '规则名称',
    type VARCHAR(100) NOT NULL COMMENT '规则类型',
    description TEXT COMMENT '规则描述',
    priority INT DEFAULT 1 COMMENT '优先级',
    active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    conditions JSON COMMENT '规则条件配置',
    created_by VARCHAR(36) COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_rule_type (type),
    INDEX idx_active (active),
    INDEX idx_priority (priority),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测规则表';

-- 创建冲突检测记录表
CREATE TABLE IF NOT EXISTS conflict_check_records (
    check_id VARCHAR(36) PRIMARY KEY COMMENT '检查记录ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    client_name VARCHAR(255) NOT NULL COMMENT '客户名称',
    case_name VARCHAR(255) NOT NULL COMMENT '案件名称',
    case_type VARCHAR(100) NOT NULL COMMENT '案件类型',
    check_status ENUM('PROCESSING', 'COMPLETED', 'FAILED') DEFAULT 'PROCESSING' COMMENT '检查状态',
    has_conflict BOOLEAN DEFAULT FALSE COMMENT '是否存在冲突',
    risk_level ENUM('HIGH', 'MEDIUM', 'LOW', 'MINIMAL') DEFAULT 'LOW' COMMENT '风险等级',
    search_parameters JSON COMMENT '搜索参数',
    check_result JSON COMMENT '检查结果',
    user_id BIGINT UNSIGNED COMMENT '用户ID',
    check_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '检查时间',
    duration BIGINT COMMENT '检查耗时(毫秒)',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_client_id (client_id),
    INDEX idx_check_status (check_status),
    INDEX idx_has_conflict (has_conflict),
    INDEX idx_risk_level (risk_level),
    INDEX idx_check_time (check_time),
    INDEX idx_user_id (user_id),
    INDEX idx_client_case (client_id, case_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突检测记录表';

-- 创建客户关联关系表
CREATE TABLE IF NOT EXISTS client_relations (
    id VARCHAR(36) PRIMARY KEY COMMENT '关联关系ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    related_client_id VARCHAR(36) NOT NULL COMMENT '关联客户ID',
    relation_type ENUM('PARENT', 'SUBSIDIARY', 'SISTER', 'COMPETITOR', 'ADVERSE', 'OTHER') NOT NULL COMMENT '关系类型',
    relation_strength DECIMAL(3,2) DEFAULT 1.00 COMMENT '关系强度',
    description TEXT COMMENT '关系描述',
    created_by VARCHAR(36) COMMENT '创建人ID',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_client_id (client_id),
    INDEX idx_related_client_id (related_client_id),
    INDEX idx_relation_type (relation_type),
    INDEX idx_created_at (created_at),
    UNIQUE KEY uk_client_relation (client_id, related_client_id, relation_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户关联关系表';

-- 创建MCP标准记录表
CREATE TABLE IF NOT EXISTS mcp_standards (
    id VARCHAR(36) PRIMARY KEY COMMENT '标准记录ID',
    version VARCHAR(50) NOT NULL COMMENT '标准版本',
    title VARCHAR(255) NOT NULL COMMENT '标准标题',
    description TEXT COMMENT '标准描述',
    content JSON COMMENT '标准内容',
    effective_date DATE COMMENT '生效日期',
    source_url VARCHAR(500) COMMENT '来源URL',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_version (version),
    INDEX idx_active (is_active),
    INDEX idx_effective_date (effective_date),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='MCP标准记录表';

-- 插入默认的冲突检测规则
INSERT INTO conflict_rules (id, name, type, description, priority, active, conditions, created_at, updated_at) VALUES
('RULE_NAME_SIMILARITY_001', '姓名相似性检测', 'NAME_SIMILARITY', '检测客户名称与历史案件的相似性', 5, TRUE, '{"threshold": 0.8, "algorithm": "levenshtein"}', NOW(), NOW()),
('RULE_CORPORATE_RELATION_001', '企业关联检测', 'CORPORATE_RELATION', '检测企业客户的关联关系冲突', 8, TRUE, '{"checkTypes": ["PARENT", "SUBSIDIARY", "SISTER"]}', NOW(), NOW()),
('RULE_CASE_CONFLICT_001', '案件冲突检测', 'CASE_CONFLICT', '检测同一客户的案件冲突', 7, TRUE, '{"allowMultipleCases": false, "timeWindow": 365}', NOW(), NOW()),
('RULE_ADVERSE_PARTY_001', '对立当事人检测', 'ADVERSE_PARTY', '检测对立当事人冲突', 9, TRUE, '{"strictMode": true, "timeWindow": 1095}', NOW(), NOW()),
('RULE_TIME_OVERLAP_001', '时间重叠检测', 'TIME_OVERLAP', '检测案件时间重叠', 3, TRUE, '{"overlapThreshold": 30, "unit": "days"}', NOW(), NOW());

-- 插入默认的MCP标准记录
INSERT INTO mcp_standards (id, version, title, description, effective_date, is_active, created_at, updated_at) VALUES
('MCP_STD_001', '2024.1', 'ABA Model Rules', '美国律师协会利益冲突标准', '2024-01-01', TRUE, NOW(), NOW()),
('MCP_STD_002', '2024.1', 'Chinese Bar Standards', '中国律师协会利益冲突规定', '2024-01-01', TRUE, NOW(), NOW());