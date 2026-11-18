-- 冲突分类标准数据模型
-- 符合IBA、ABA、中国律师协会标准

-- 1. 冲突分类标准表
CREATE TABLE IF NOT EXISTS conflict_classifications (
    id VARCHAR(36) PRIMARY KEY COMMENT '分类ID',
    code VARCHAR(50) NOT NULL UNIQUE COMMENT '分类代码',
    name VARCHAR(100) NOT NULL COMMENT '分类名称',
    display_name VARCHAR(150) NOT NULL COMMENT '显示名称',

    -- 标准依据
    category ENUM('DIRECT_OPPOSITION', 'CONCURRENT', 'FORMER_CLIENT', 'RELATIONSHIP',
                  'FINANCIAL', 'OFFICIAL', 'RELATED_ENTITY', 'INFORMATION') NOT NULL COMMENT '分类大类',
    subcategory VARCHAR(100) COMMENT '子分类',

    -- 法律标准依据
    mcp_standard VARCHAR(50) NOT NULL COMMENT '遵循标准(IBA/ABA/CHINESE_BAR)',
    legal_basis TEXT COMMENT '法律依据',
    regulatory_reference VARCHAR(255) COMMENT '法规引用',

    -- 处理属性
    waivable BOOLEAN DEFAULT FALSE COMMENT '是否可豁免',
    waiver_conditions JSON COMMENT '豁免条件',
    approval_required BOOLEAN DEFAULT TRUE COMMENT '是否需要审批',
    approval_level ENUM('DEPARTMENT_HEAD', 'COMMITTEE', 'MANAGEMENT_PARTNER', 'ETHICS_OFFICER') DEFAULT 'DEPARTMENT_HEAD' COMMENT '审批级别',

    -- 风险评估
    base_risk_level ENUM('CRITICAL', 'HIGH', 'MEDIUM', 'LOW') NOT NULL COMMENT '基础风险等级',
    risk_factors JSON COMMENT '风险因素列表',
    risk_score DECIMAL(5,2) DEFAULT 50.00 COMMENT '风险评分(0-100)',

    -- 处理措施
    default_action ENUM('REJECT_REPRESENTATION', 'OBTAIN_WAIVER', 'IMPLEMENT_BARRIER', 'MONITOR', 'DOCUMENT') NOT NULL COMMENT '默认处理措施',
    mitigation_measures JSON COMMENT '缓解措施',
    monitoring_requirements JSON COMMENT '监控要求',

    -- 时间相关规则
    time_limitation JSON COMMENT '时间限制规则',
    cooling_period_days INT DEFAULT 0 COMMENT '冷却期天数',

    -- 专业领域特殊规则
    practice_area_rules JSON COMMENT '执业领域特殊规则',
    jurisdiction_specific_rules JSON COMMENT '司法管辖区特殊规则',

    -- 示例场景
    example_scenarios JSON COMMENT '示例场景',
    case_examples JSON COMMENT '案例示例',

    -- 状态和版本管理
    status ENUM('ACTIVE', 'INACTIVE', 'UNDER_REVIEW', 'DEPRECATED') DEFAULT 'ACTIVE' COMMENT '状态',
    version INT DEFAULT 1 COMMENT '版本号',
    effective_date DATE NOT NULL COMMENT '生效日期',
    expiry_date DATE COMMENT '到期日期',

    -- 审计信息
    created_by VARCHAR(36) NOT NULL COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    approved_by VARCHAR(36) COMMENT '审批人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    -- 索引
    INDEX idx_code (code),
    INDEX idx_category (category),
    INDEX idx_mcp_standard (mcp_standard),
    INDEX idx_base_risk_level (base_risk_level),
    INDEX idx_waivable (waivable),
    INDEX idx_approval_required (approval_required),
    INDEX idx_status (status),
    INDEX idx_effective_date (effective_date),
    INDEX idx_created_at (created_at),

    -- 全文搜索索引
    FULLTEXT INDEX ft_classification (name, display_name, legal_basis, example_scenarios)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='冲突分类标准表';

-- 2. IBA国际仲裁利益冲突指南标准
INSERT INTO conflict_classifications (
    id, code, name, display_name, category, mcp_standard, legal_basis,
    waivable, waiver_conditions, approval_required, approval_level,
    base_risk_level, risk_factors, default_action, mitigation_measures,
    effective_date, created_by
) VALUES
-- IBA 红色清单（不可豁免）
(UUID(), 'IBA_RED_DIRECT_FINANCIAL', '直接财务利益冲突', '仲裁员与当事人存在直接财务利益',
 'FINANCIAL', 'IBA_2024', 'IBA Guidelines on Conflicts of Interest in International Arbitration 2024 - Red List',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('直接财务利益', '独立性受损', '公正性质疑'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('立即拒绝代理', '不得接受委托', '全面披露关系'), CURDATE(), 'system'),

(UUID(), 'IBA_RED_REPRESENTATION', '近期代理关系冲突', '仲裁员最近担任某方当事人的代理人',
 'RELATIONSHIP', 'IBA_2024', 'IBA Guidelines on Conflicts of Interest in International Arbitration 2024 - Red List',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('近期代理关系', '保密义务冲突', '忠诚义务冲突'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('拒绝接受委托', '建议更换仲裁员', '记录冲突原因'), CURDATE(), 'system'),

-- IBA 橙色清单（可豁免，需披露）
(UUID(), 'IBA_ORANGE_BUSINESS_RELATIONSHIP', '商业关系冲突', '仲裁员与当事人存在商业关系',
 'RELATIONSHIP', 'IBA_2024', 'IBA Guidelines on Conflicts of Interest in International Arbitration 2024 - Orange List',
 TRUE, JSON_ARRAY('充分披露', '当事人明确同意'), TRUE, 'DEPARTMENT_HEAD', 'HIGH',
 JSON_ARRAY('商业关系', '可能影响公正性', '合理怀疑'), 'OBTAIN_WAIVER',
 JSON_ARRAY('全面披露商业关系', '获取当事人知情同意书', '建立信息屏障'), CURDATE(), 'system'),

(UUID(), 'IBA_ORANGE_LEGAL_ADVISOR', '法律顾问关系冲突', '仲裁员近期为当事人提供法律咨询',
 'INFORMATION', 'IBA_2024', 'IBA Guidelines on Conflicts of Interest in International Arbitration 2024 - Orange List',
 TRUE, JSON_ARRAY('信息披露', '同意豁免'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('法律顾问关系', '信息保密义务', '专业判断可能受影响'), 'OBTAIN_WAIVER',
 JSON_ARRAY('披露咨询关系', '获取豁免同意', '监控利益冲突'), CURDATE(), 'system');

-- 3. ABA职业行为模范规则标准
INSERT INTO conflict_classifications (
    id, code, name, display_name, category, mcp_standard, legal_basis,
    waivable, waiver_conditions, approval_required, approval_level,
    base_risk_level, risk_factors, default_action, mitigation_measures,
    cooling_period_days, effective_date, created_by
) VALUES
-- ABA Rule 1.7 - 并发冲突
(UUID(), 'ABA_RULE_1_7_CONCURRENT', '并发利益冲突', '律师代理存在重大利益冲突的客户',
 'CONCURRENT', 'ABA_MODEL_RULES', 'ABA Model Rules of Professional Conduct Rule 1.7 - Concurrent Conflicts of Interest',
 TRUE, JSON_ARRAY('充分告知', '明确同意', '合理相信能够胜任'), TRUE, 'DEPARTMENT_HEAD', 'HIGH',
 JSON_ARRAY('利益直接对立', '重大利益冲突风险', '代理能力受损'), 'OBTAIN_WAIVER',
 JSON_ARRAY('全面披露冲突', '获取书面同意', '实施信息屏障', '定期评估'), 0, CURDATE(), 'system'),

(UUID(), 'ABA_RULE_1_7_MATERIAL_LIMITATION', '重大限制冲突', '律师代理能力受到重大限制',
 'CONCURRENT', 'ABA_MODEL_RULES', 'ABA Model Rules of Professional Conduct Rule 1.7 - Material Limitation',
 TRUE, JSON_ARRAY('充分披露', '合理相信能够胜任'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('代理能力受限', '服务质量可能受影响', '专业判断可能受影响'), 'OBTAIN_WAIVER',
 JSON_ARRAY('披露限制情况', '评估代理能力', '获取客户同意'), 0, CURDATE(), 'system'),

-- ABA Rule 1.9 - 前任客户冲突
(UUID(), 'ABA_RULE_1_9_FORMER_CLIENT_SAME', '相同事项前任客户冲突', '律师代理前任客户的相对方，涉及相同或实质相关事项',
 'FORMER_CLIENT', 'ABA_MODEL_RULES', 'ABA Model Rules of Professional Conduct Rule 1.9 - Duties to Former Clients',
 FALSE, JSON_ARRAY(), TRUE, 'COMMITTEE', 'HIGH',
 JSON_ARRAY('相同事项', '前任客户保密信息', '忠诚义务冲突'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('拒绝代理', '建议另寻律师', '记录拒绝原因'), 365, CURDATE(), 'system'),

(UUID(), 'ABA_RULE_1_9_FORMER_CLIENT_RELATED', '相关事项前任客户冲突', '律师代理前任客户的相对方，涉及实质相关事项',
 'FORMER_CLIENT', 'ABA_MODEL_RULES', 'ABA Model Rules of Professional Conduct Rule 1.9 - Substantially Related Matters',
 TRUE, JSON_ARRAY('无保密信息', '充分披露'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('实质相关事项', '可能使用前任客户信息'), 'OBTAIN_WAIVER',
 JSON_ARRAY('检查保密义务', '充分披露关系', '获取前任客户同意'), 180, CURDATE(), 'system');

-- 4. 中国律师执业规范标准
INSERT INTO conflict_classifications (
    id, code, name, display_name, category, mcp_standard, legal_basis,
    waivable, waiver_conditions, approval_required, approval_level,
    base_risk_level, risk_factors, default_action, mitigation_measures,
    effective_date, created_by
) VALUES
-- 中国律师执业规范第49条 - 禁止情况
(UUID(), 'CHINESE_BAR_ARTICLE_49_SAME_CASE', '同一案件双方代理', '律师代理同一案件双方当事人',
 'DIRECT_OPPOSITION', 'CHINESE_BAR', '中国律师执业规范第49条第1项',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('直接对立', '严重利益冲突', '违反执业道德'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('立即拒绝代理', '不得接受委托', '向律协报告'), CURDATE(), 'system'),

(UUID(), 'CHINESE_BAR_ARTICLE_49_FAMILY_CONFLICT', '近亲属代理冲突', '律师与其代理的当事人存在近亲属关系',
 'RELATIONSHIP', 'CHINESE_BAR', '中国律师执业规范第49条第3项',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('近亲属关系', '可能影响公正性', '利益冲突明显'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('拒绝代理', '更换律师', '记录冲突原因'), CURDATE(), 'system'),

(UUID(), 'CHINESE_BAR_ARTICLE_49_CRIMINAL_CONFLICT', '刑事案件代理冲突', '同一律师事务所不同律师同时代理同一刑事案件的被害人和被告人',
 'DIRECT_OPPOSITION', 'CHINESE_BAR', '中国律师执业规范第49条第4项',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('刑事案件对立', '律所内部冲突', '严重利益冲突'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('律所拒绝代理', '分配不同律所', '向当事人说明'), CURDATE(), 'system'),

-- 中国律师执业规范第50条 - 应当告知情况
(UUID(), 'CHINESE_BAR_ARTICLE_50_FAMILY_MEMBER', '律所其他律师近亲属冲突', '同一所其他律师是案件对方当事人的近亲属',
 'RELATIONSHIP', 'CHINESE_BAR', '中国律师执业规范第50条第1项',
 TRUE, JSON_ARRAY('客户明确同意', '书面记录'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('同事近亲属关系', '可能影响协作', '内部协调需求'), 'OBTAIN_WAIVER',
 JSON_ARRAY('向客户披露', '获取书面同意', '建立内部监督'), CURDATE(), 'system'),

(UUID(), 'CHINESE_BAR_ARTICLE_50_BUSINESS_CONFLICT', '同一律所业务冲突', '同一律师事务所接受对方当事人其他法律业务委托',
 'CONCURRENT', 'CHINESE_BAR', '中国律师执业规范第50条第3项',
 TRUE, JSON_ARRAY('书面同意', '信息屏障'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('律所内部业务冲突', '信息共享风险', '协作可能受影响'), 'OBTAIN_WAIVER',
 JSON_ARRAY('实施信息屏障', '获取客户同意', '定期合规检查'), CURDATE(), 'system');

-- 5. 通用商业利益冲突
INSERT INTO conflict_classifications (
    id, code, name, display_name, category, subcategory, mcp_standard, legal_basis,
    waivable, waiver_conditions, approval_required, approval_level,
    base_risk_level, risk_factors, default_action, mitigation_measures,
    example_scenarios, effective_date, created_by
) VALUES
-- 商业竞争冲突
(UUID(), 'BUSINESS_COMPETITION', '商业竞争冲突', '律师同时代理存在竞争关系的客户',
 'RELATED_ENTITY', 'COMPETITOR', 'GENERAL', '一般商业利益冲突原则',
 TRUE, JSON_ARRAY('全面披露', '客户同意', '竞争风险评估'), TRUE, 'DEPARTMENT_HEAD', 'HIGH',
 JSON_ARRAY('直接竞争关系', '商业秘密保护', '代理利益冲突'), 'OBTAIN_WAIVER',
 JSON_ARRAY('签署保密协议', '建立信息屏障', '定期评估冲突', '客户书面同意'),
 JSON_ARRAY('代理两个竞争对手的诉讼案件', '同时为上下游企业提供服务', '代理同行业公司的并购业务'),
 CURDATE(), 'system'),

-- 关联企业冲突
(UUID(), 'RELATED_ENTITY_CONFLICT', '关联实体冲突', '律师代理的客户与另一个客户存在关联关系',
 'RELATED_ENTITY', 'CORPORATE', 'GENERAL', '企业关联关系利益冲突原则',
 TRUE, JSON_ARRAY('披露关联关系', '评估冲突程度', '获取同意'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('母子公司关系', '关联企业交易', '利益输送风险'), 'OBTAIN_WAIVER',
 JSON_ARRAY('披露关联关系', '审查交易性质', '实施适当隔离措施', '定期审查'),
 JSON_ARRAY('代理母公司同时代理子公司', '代理集团内多家关联公司', '代理存在股权关系的公司'),
 CURDATE(), 'system'),

-- 时间序列冲突
(UUID(), 'TIME_BASED_CONFLICT', '时间序列冲突', '律师在前任代理中获取的保密信息与当前代理冲突',
 'INFORMATION', 'TIME_SEQUENCE', 'GENERAL', '保密义务时效原则',
 FALSE, JSON_ARRAY(), TRUE, 'COMMITTEE', 'HIGH',
 JSON_ARRAY('保密信息冲突', '忠诚义务冲突', '专业判断受影响'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('检查保密义务范围', '评估信息敏感性', '如无冲突可继续代理'),
 JSON_ARRAY('前客户商业机密与当前客户相关', '前客户案件信息与当前案件重叠', '使用前客户策略信息'),
 CURDATE(), 'system');

-- 6. 特殊领域冲突
INSERT INTO conflict_classifications (
    id, code, name, display_name, category, mcp_standard, legal_basis,
    waivable, waiver_conditions, approval_required, approval_level,
    base_risk_level, risk_factors, default_action, mitigation_measures,
    practice_area_rules, effective_date, created_by
) VALUES
-- 知识产权冲突
(UUID(), 'IP_CONFLICT', '知识产权冲突', '涉及专利、商标、著作权等知识产权的利益冲突',
 'RELATED_ENTITY', 'GENERAL', '知识产权代理特殊规则',
 TRUE, JSON_ARRAY('技术领域分离', '客户同意'), TRUE, 'DEPARTMENT_HEAD', 'MEDIUM',
 JSON_ARRAY('技术信息冲突', '知识产权保护', '商业秘密保护'), 'OBTAIN_WAIVER',
 JSON_ARRAY('技术领域审查', '签署保密协议', '建立信息屏障', '定期合规检查'),
 JSON_OBJECT('technology_separation' => true, 'confidentiality_agreement' => true, 'information_barrier' => true),
 CURDATE(), 'system'),

-- 证券发行冲突
(UUID(), 'SECURITY_OFFERING_CONFLICT', '证券发行冲突', '涉及证券发行、上市的特别利益冲突',
 'FINANCIAL', 'SECURITIES_REGULATION', '证券法相关规定',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'HIGH',
 JSON_ARRAY('发行人利益冲突', '投资者保护', '监管合规要求'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('检查监管禁止规定', '评估市场影响', '确保投资者保护'),
 JSON_OBJECT('regulatory_compliance' => true, 'investor_protection' => true, 'disclosure_requirements' => true),
 CURDATE(), 'system'),

-- 政府官员冲突
(UUID(), 'GOVERNMENT_OFFICIAL_CONFLICT', '政府官员冲突', '涉及政府官员、人大代表的特殊利益冲突',
 'OFFICIAL', 'GOVERNMENT_ETHICS', '公务员法、人大代表法相关规定',
 FALSE, JSON_ARRAY(), TRUE, 'MANAGEMENT_PARTNER', 'CRITICAL',
 JSON_ARRAY('职务影响', '公共利益冲突', '权力寻租风险'), 'REJECT_REPRESENTATION',
 JSON_ARRAY('核实官员身份', '评估职务影响', '拒绝代理或等待离职'),
 JSON_OBJECT('identity_verification' => true, 'conflict_assessment' => true, 'waiting_period' => true),
 CURDATE(), 'system');

-- 7. 创建分类视图
CREATE OR REPLACE VIEW conflict_classifications_active AS
SELECT
    cc.*,
    -- 计算动态风险评分
    CASE
        WHEN cc.base_risk_level = 'CRITICAL' THEN 90 + cc.risk_score * 0.1
        WHEN cc.base_risk_level = 'HIGH' THEN 70 + cc.risk_score * 0.2
        WHEN cc.base_risk_level = 'MEDIUM' THEN 50 + cc.risk_score * 0.3
        ELSE 30 + cc.risk_score * 0.4
    END as calculated_risk_score,

    -- 分类路径
    CONCAT(
        CASE cc.mcp_standard
            WHEN 'IBA_2024' THEN 'IBA 2024'
            WHEN 'ABA_MODEL_RULES' THEN 'ABA Model Rules'
            WHEN 'CHINESE_BAR' THEN '中国律师执业规范'
            ELSE 'General'
        END,
        ' > ',
        CASE cc.category
            WHEN 'DIRECT_OPPOSITION' THEN '直接对立'
            WHEN 'CONCURRENT' THEN '并发冲突'
            WHEN 'FORMER_CLIENT' THEN '前任客户'
            WHEN 'RELATIONSHIP' THEN '关系冲突'
            WHEN 'FINANCIAL' THEN '财务利益'
            WHEN 'OFFICIAL' THEN '职务冲突'
            WHEN 'RELATED_ENTITY' THEN '关联实体'
            WHEN 'INFORMATION' THEN '信息冲突'
        END,
        COALESCE(CONCAT(' > ', cc.subcategory), '')
    ) as classification_path,

    -- 可操作性评估
    CASE
        WHEN NOT cc.waivable AND cc.base_risk_level = 'CRITICAL' THEN '禁止代理'
        WHEN NOT cc.waivable AND cc.base_risk_level = 'HIGH' THEN '通常禁止'
        WHEN cc.waivable AND cc.approval_required = true THEN '需审批豁免'
        WHEN cc.waivable AND cc.approval_required = false THEN '可豁免'
        ELSE '需评估'
    END as action_guidance

FROM conflict_classifications cc
WHERE cc.status = 'ACTIVE'
  AND (cc.expiry_date IS NULL OR cc.expiry_date >= CURDATE());

-- 8. 创建冲突类型统计视图
CREATE OR REPLACE VIEW conflict_classification_stats AS
SELECT
    category,
    mcp_standard,
    COUNT(*) as total_classifications,
    SUM(CASE WHEN base_risk_level = 'CRITICAL' THEN 1 ELSE 0 END) as critical_count,
    SUM(CASE WHEN base_risk_level = 'HIGH' THEN 1 ELSE 0 END) as high_count,
    SUM(CASE WHEN base_risk_level = 'MEDIUM' THEN 1 ELSE 0 END) as medium_count,
    SUM(CASE WHEN base_risk_level = 'LOW' THEN 1 ELSE 0 END) as low_count,
    SUM(CASE WHEN waivable = true THEN 1 ELSE 0 END) as waivable_count,
    SUM(CASE WHEN approval_required = true THEN 1 ELSE 0 END) as approval_required_count,
    AVG(risk_score) as avg_risk_score,
    MIN(effective_date) as earliest_effective_date,
    MAX(effective_date) as latest_effective_date
FROM conflict_classifications
WHERE status = 'ACTIVE'
  AND (expiry_date IS NULL OR expiry_date >= CURDATE())
GROUP BY category, mcp_standard
ORDER BY category, mcp_standard;