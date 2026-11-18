-- Correct the historical trial fixture: a short-name candidate is evidence for
-- human review, not proof of a high-risk conflict.
UPDATE conflict_check_records
SET risk_level = 'MEDIUM',
    has_conflict = FALSE,
    search_parameters = COALESCE(search_parameters, '{}'::jsonb) || jsonb_build_object(
        'matchMode', 'NAME_CANDIDATE',
        'automaticConclusion', FALSE,
        'verificationRequired', TRUE
    ),
    check_result = jsonb_set(
        jsonb_set(
            COALESCE(check_result, '{}'::jsonb),
            '{riskAssessment}',
            jsonb_build_object(
                'overallRisk', 'MEDIUM',
                'riskScore', 58,
                'riskReason', '简称“示例科技”与“上海示例科技有限公司”形成名称候选，尚不能确认是同一主体，需人工核实。',
                'requiresApproval', TRUE,
                'riskFactors', jsonb_build_array('名称候选匹配', '缺少统一社会信用代码或主体关系证据'),
                'matchEvidence', jsonb_build_object(
                    'queryName', '示例科技',
                    'candidateName', '上海示例科技有限公司',
                    'matchType', 'NAME_CANDIDATE',
                    'algorithm', 'NORMALIZED_CONTAINS',
                    'automaticConclusion', FALSE
                )
            ),
            TRUE
        ),
        '{isConflict}',
        'false'::jsonb,
        TRUE
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE check_id = 'LAWYER-TRIAL-HIGH-001';

UPDATE conflict_cases
SET conflict_type = '名称相似待核实',
    risk_level = 'MEDIUM',
    description = '简称“示例科技”与“上海示例科技有限公司”形成名称候选，需核实统一社会信用代码或主体关系。',
    conflict_details = '名称候选只用于人工核实，不自动认定高风险冲突。'
WHERE check_id = 'LAWYER-TRIAL-HIGH-001';

UPDATE approval_requests
SET content = '名称候选检测结果需要人工核实，请主任复核。',
    priority = 'medium',
    conflict_risk_level = 'MEDIUM',
    conflict_result = jsonb_set(
        COALESCE(conflict_result, '{}'::jsonb),
        '{riskAssessment}',
        jsonb_build_object('overallRisk', 'MEDIUM', 'riskScore', 58, 'requiresApproval', TRUE),
        TRUE
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE conflict_check_id = 'LAWYER-TRIAL-HIGH-001';
