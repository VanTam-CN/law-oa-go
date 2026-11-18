-- 客户信息管理增强数据模型
-- 第一阶段：基础架构完善

-- 1. 客户扩展信息表
CREATE TABLE IF NOT EXISTS client_profiles (
    id VARCHAR(36) PRIMARY KEY COMMENT '客户档案ID',
    client_id VARCHAR(36) NOT NULL COMMENT '关联客户ID',

    -- 法律实体信息
    legal_form ENUM('PERSON', 'COMPANY', 'GOVERNMENT', 'NGO', 'PARTNERSHIP') NOT NULL COMMENT '法律形式',
    registration_number VARCHAR(100) COMMENT '注册号',
    business_license VARCHAR(100) COMMENT '营业执照号',
    tax_id VARCHAR(50) COMMENT '税号',
    incorporation_date DATE COMMENT '成立日期',
    registered_capital DECIMAL(20,2) COMMENT '注册资本',

    -- 企业结构信息
    parent_companies JSON COMMENT '母公司信息',
    subsidiaries JSON COMMENT '子公司信息',
    sister_companies JSON COMMENT '姐妹公司信息',
    joint_ventures JSON COMMENT '合资企业信息',
    shareholders JSON COMMENT '主要股东信息',

    -- 关键人员信息
    directors JSON COMMENT '董事信息',
    officers JSON COMMENT '高管信息',
    legal_representatives JSON COMMENT '法定代表人信息',
    key_contacts JSON COMMENT '关键联系人信息',

    -- 家庭成员信息（个人客户）
    family_members JSON COMMENT '家庭成员信息',
    spouse_info JSON COMMENT '配偶信息',
    dependents JSON COMMENT '受抚养人信息',

    -- 多语言和别名支持
    name_variations JSON COMMENT '名称变体',
    alias_names JSON COMMENT '别名信息',
    translations JSON COMMENT '多语言翻译',

    -- 行业和业务信息
    industry_code VARCHAR(50) COMMENT '行业代码',
    business_scope TEXT COMMENT '经营范围',
    main_products_services JSON COMMENT '主要产品/服务',

    -- 风险评估信息
    risk_level ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM' COMMENT '风险等级',
    compliance_notes TEXT COMMENT '合规备注',
    special_considerations JSON COMMENT '特殊考虑事项',

    -- 状态管理
    status ENUM('ACTIVE', 'INACTIVE', 'UNDER_REVIEW', 'SUSPENDED') DEFAULT 'ACTIVE' COMMENT '状态',
    verification_status ENUM('VERIFIED', 'PENDING', 'UNVERIFIED') DEFAULT 'PENDING' COMMENT '验证状态',
    last_verified_date TIMESTAMP COMMENT '最后验证日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_client_id (client_id),
    INDEX idx_legal_form (legal_form),
    INDEX idx_registration_number (registration_number),
    INDEX idx_tax_id (tax_id),
    INDEX idx_industry_code (industry_code),
    INDEX idx_risk_level (risk_level),
    INDEX idx_status (status),
    INDEX idx_verification_status (verification_status),
    INDEX idx_created_at (created_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_client_profile (name_variations, alias_names, business_scope),

    -- 外键约束
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户档案扩展信息表';

-- 2. 客户关系表（增强版）
CREATE TABLE IF NOT EXISTS client_relationships (
    id VARCHAR(36) PRIMARY KEY COMMENT '关系ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',
    related_entity_id VARCHAR(36) NOT NULL COMMENT '关联实体ID',
    related_entity_name VARCHAR(255) NOT NULL COMMENT '关联实体名称',

    -- 关系分类
    relationship_type ENUM('PARENT', 'SUBSIDIARY', 'SISTER', 'COMPETITOR', 'ADVERSE',
                          'SUPPLIER', 'CUSTOMER', 'PARTNER', 'JOINT_VENTURE', 'AFFILIATE',
                          'FAMILY_MEMBER', 'BUSINESS_ASSOCIATE', 'LEGAL_REPRESENTATIVE') NOT NULL COMMENT '关系类型',
    relationship_subtype VARCHAR(100) COMMENT '关系子类型',

    -- 关系强度和置信度
    relationship_strength DECIMAL(3,2) DEFAULT 1.00 COMMENT '关系强度(0.00-1.00)',
    confidence_score DECIMAL(3,2) DEFAULT 1.00 COMMENT '置信度(0.00-1.00)',

    -- 关系详情
    relationship_description TEXT COMMENT '关系描述',
    nature_of_relationship TEXT COMMENT '关系性质',
    business_context TEXT COMMENT '业务背景',

    -- 时间信息
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',
    last_reviewed_date DATE COMMENT '最后审查日期',

    -- 验证和来源
    verification_status ENUM('VERIFIED', 'PENDING', 'UNVERIFIED', 'REJECTED') DEFAULT 'PENDING' COMMENT '验证状态',
    verification_method VARCHAR(100) COMMENT '验证方法',
    source_document VARCHAR(255) COMMENT '来源文件',
    source_url VARCHAR(500) COMMENT '来源URL',

    -- 风险评估
    risk_impact ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM' COMMENT '风险影响',
    risk_factors JSON COMMENT '风险因素',
    mitigation_measures JSON COMMENT '缓解措施',

    -- 状态管理
    status ENUM('ACTIVE', 'INACTIVE', 'UNDER_REVIEW', 'SUSPENDED') DEFAULT 'ACTIVE' COMMENT '状态',
    monitoring_required BOOLEAN DEFAULT TRUE COMMENT '是否需要监控',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    approved_by VARCHAR(36) COMMENT '审批人',
    approved_date TIMESTAMP COMMENT '审批日期',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_client_id (client_id),
    INDEX idx_related_entity_id (related_entity_id),
    INDEX idx_relationship_type (relationship_type),
    INDEX idx_relationship_strength (relationship_strength),
    INDEX idx_effective_date (effective_date),
    INDEX idx_expiry_date (expiry_date),
    INDEX idx_verification_status (verification_status),
    INDEX idx_risk_impact (risk_impact),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),

    -- 复合索引
    INDEX idx_client_relationship (client_id, relationship_type, status),
    INDEX idx_entity_relationship (related_entity_id, relationship_type, status),

    -- 唯一约束
    UNIQUE KEY uk_client_relationship (client_id, related_entity_id, relationship_type, effective_date),

    -- 全文搜索索引
    FULLTEXT INDEX ft_relationship (related_entity_name, relationship_description, business_context)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户关系表';

-- 3. 客户名称变体表（支持模糊匹配）
CREATE TABLE IF NOT EXISTS client_name_variants (
    id VARCHAR(36) PRIMARY KEY COMMENT '变体ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',

    -- 名称变体信息
    variant_name VARCHAR(500) NOT NULL COMMENT '名称变体',
    variant_type ENUM('ALIAS', 'ABBREVIATION', 'TRANSLATION', 'FORMER_NAME', 'LEGAL_NAME',
                      'TRADING_NAME', 'BRAND_NAME', 'INTERNATIONAL_NAME') NOT NULL COMMENT '变体类型',

    -- 语言和地区
    language_code VARCHAR(10) DEFAULT 'zh-CN' COMMENT '语言代码',
    country_code VARCHAR(10) COMMENT '国家代码',

    -- 使用频率和置信度
    usage_frequency ENUM('RARE', 'OCCASIONAL', 'FREQUENT', 'COMMON') DEFAULT 'OCCASIONAL' COMMENT '使用频率',
    confidence_score DECIMAL(3,2) DEFAULT 0.80 COMMENT '置信度',

    -- 时间信息
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 来源和验证
    source VARCHAR(100) COMMENT '来源',
    verification_status ENUM('VERIFIED', 'PENDING', 'UNVERIFIED') DEFAULT 'PENDING' COMMENT '验证状态',

    -- 状态
    status ENUM('ACTIVE', 'INACTIVE') DEFAULT 'ACTIVE' COMMENT '状态',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_client_id (client_id),
    INDEX idx_variant_name (variant_name),
    INDEX idx_variant_type (variant_type),
    INDEX idx_language_code (language_code),
    INDEX idx_usage_frequency (usage_frequency),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),

    -- 唯一约束
    UNIQUE KEY uk_client_variant (client_id, variant_name, variant_type, language_code),

    -- 全文搜索索引
    FULLTEXT INDEX ft_variant_name (variant_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户名称变体表';

-- 4. 客户行业分类表
CREATE TABLE IF NOT EXISTS client_industry_classifications (
    id VARCHAR(36) PRIMARY KEY COMMENT '分类ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',

    -- 行业分类
    primary_industry VARCHAR(100) NOT NULL COMMENT '主要行业',
    secondary_industry VARCHAR(100) COMMENT '次要行业',
    industry_code VARCHAR(50) COMMENT '行业代码',
    classification_system VARCHAR(50) DEFAULT 'NAICS' COMMENT '分类体系',

    -- 业务领域
    business_sectors JSON COMMENT '业务领域',
    market_segments JSON COMMENT '市场细分',

    -- 竞争信息
    main_competitors JSON COMMENT '主要竞争对手',
    market_position ENUM('LEADER', 'CHALLENGER', 'FOLLOWER', 'NICHE') COMMENT '市场地位',

    -- 风险评估
    regulatory_risk_level ENUM('LOW', 'MEDIUM', 'HIGH') DEFAULT 'MEDIUM' COMMENT '监管风险等级',
    compliance_requirements JSON COMMENT '合规要求',

    -- 时间信息
    effective_date DATE NOT NULL COMMENT '生效日期',
    last_reviewed_date DATE COMMENT '最后审查日期',

    -- 状态
    status ENUM('ACTIVE', 'INACTIVE', 'UNDER_REVIEW') DEFAULT 'ACTIVE' COMMENT '状态',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_client_id (client_id),
    INDEX idx_primary_industry (primary_industry),
    INDEX idx_industry_code (industry_code),
    INDEX idx_classification_system (classification_system),
    INDEX idx_regulatory_risk_level (regulatory_risk_level),
    INDEX idx_status (status),

    -- 外键约束
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户行业分类表';

-- 5. 客户合规信息表
CREATE TABLE IF NOT EXISTS client_compliance_records (
    id VARCHAR(36) PRIMARY KEY COMMENT '合规记录ID',
    client_id VARCHAR(36) NOT NULL COMMENT '客户ID',

    -- 合规检查类型
    compliance_type ENUM('ANTI_MONEY_LAUNDERING', 'SANCTIONS', 'PEP_SCREENING', 'DUE_DILIGENCE',
                         'REGULATORY_COMPLIANCE', 'ETHICS_REVIEW', 'CONFLICT_OF_INTEREST') NOT NULL COMMENT '合规检查类型',

    -- 检查结果
    compliance_status ENUM('COMPLIANT', 'NON_COMPLIANT', 'PENDING_REVIEW', 'EXEMPTION', 'REQUIRE_MONITORING') NOT NULL COMMENT '合规状态',
    risk_level ENUM('LOW', 'MEDIUM', 'HIGH', 'CRITICAL') DEFAULT 'MEDIUM' COMMENT '风险等级',

    -- 检查详情
    check_details JSON COMMENT '检查详情',
    findings TEXT COMMENT '发现的问题',
    recommendations TEXT COMMENT '建议措施',

    -- 监控要求
    monitoring_required BOOLEAN DEFAULT FALSE COMMENT '是否需要监控',
    monitoring_frequency ENUM('DAILY', 'WEEKLY', 'MONTHLY', 'QUARTERLY', 'ANNUALLY') COMMENT '监控频率',
    next_review_date DATE COMMENT '下次审查日期',

    -- 豁免和授权
    exemption_granted BOOLEAN DEFAULT FALSE COMMENT '是否获得豁免',
    exemption_details TEXT COMMENT '豁免详情',
    exemption_expiry_date DATE COMMENT '豁免到期日',

    -- 时间信息
    check_date DATE NOT NULL COMMENT '检查日期',
    effective_date DATE COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 文档和证据
    supporting_documents JSON COMMENT '支持文件',
    evidence_urls JSON COMMENT '证据链接',

    -- 审计信息
    checked_by VARCHAR(36) NOT NULL COMMENT '检查人',
    approved_by VARCHAR(36) COMMENT '审批人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_client_id (client_id),
    INDEX idx_compliance_type (compliance_type),
    INDEX idx_compliance_status (compliance_status),
    INDEX idx_risk_level (risk_level),
    INDEX idx_check_date (check_date),
    INDEX idx_next_review_date (next_review_date),
    INDEX idx_monitoring_required (monitoring_required),

    -- 外键约束
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户合规记录表';

-- 插入初始的行业分类数据
INSERT IGNORE INTO client_industry_classifications
(id, client_id, primary_industry, industry_code, classification_system, effective_date, created_by, created_at)
SELECT
    UUID() as id,
    c.id as client_id,
    '其他行业' as primary_industry,
    '999999' as industry_code,
    'NAICS' as classification_system,
    CURDATE() as effective_date,
    'system' as created_by,
    NOW() as created_at
FROM clients c
WHERE NOT EXISTS (
    SELECT 1 FROM client_industry_classifications cic WHERE cic.client_id = c.id
);

-- 创建触发器：自动创建客户档案记录
DELIMITER $$
CREATE TRIGGER IF NOT EXISTS trigger_client_profile_creation
    AFTER INSERT ON clients
    FOR EACH ROW
BEGIN
    INSERT INTO client_profiles (
        id, client_id, legal_form, status, verification_status, created_by, created_at, updated_at
    ) VALUES (
        UUID(), NEW.id, 'PERSON', 'ACTIVE', 'PENDING', NEW.created_by, NOW(), NOW()
    );

    INSERT INTO client_name_variants (
        id, client_id, variant_name, variant_type, language_code, usage_frequency,
        confidence_score, effective_date, created_by, created_at, updated_at
    ) VALUES (
        UUID(), NEW.id, NEW.name, 'LEGAL_NAME', 'zh-CN', 'COMMON', 1.00, CURDATE(),
        NEW.created_by, NOW(), NOW()
    );
END$$
DELIMITER ;

-- 创建视图：客户完整信息视图
CREATE OR REPLACE VIEW client_complete_info AS
SELECT
    c.id as client_id,
    c.name as client_name,
    c.type as client_type,
    c.contact_info,
    c.created_at as client_created_at,

    -- 档案信息
    cp.legal_form,
    cp.registration_number,
    cp.business_license,
    cp.tax_id,
    cp.incorporation_date,
    cp.registered_capital,
    cp.risk_level as profile_risk_level,
    cp.verification_status,

    -- 行业分类
    cic.primary_industry,
    cic.secondary_industry,
    cic.industry_code,
    cic.regulatory_risk_level,

    -- 关系统计
    (SELECT COUNT(*) FROM client_relationships cr
     WHERE cr.client_id = c.id AND cr.status = 'ACTIVE') as active_relationships_count,

    -- 名称变体统计
    (SELECT COUNT(*) FROM client_name_variants cnv
     WHERE cnv.client_id = c.id AND cnv.status = 'ACTIVE') as name_variants_count,

    -- 合规检查统计
    (SELECT COUNT(*) FROM client_compliance_records ccr
     WHERE ccr.client_id = c.id AND ccr.compliance_status = 'NON_COMPLIANT') as compliance_issues_count,

    -- 最后更新时间
    GREATEST(
        COALESCE(cp.updated_at, c.created_at),
        COALESCE(cic.updated_at, c.created_at)
    ) as last_updated_at

FROM clients c
LEFT JOIN client_profiles cp ON c.id = cp.client_id
LEFT JOIN client_industry_classifications cic ON c.id = cic.client_id AND cic.status = 'ACTIVE'
WHERE c.deleted_at IS NULL;