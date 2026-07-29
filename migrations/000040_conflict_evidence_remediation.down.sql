UPDATE conflict_check_records
SET risk_level = 'HIGH',
    has_conflict = TRUE,
    check_result = jsonb_set(
        COALESCE(check_result, '{}'::jsonb),
        '{riskAssessment}',
        jsonb_build_object(
            'overallRisk', 'HIGH',
            'riskScore', 92,
            'riskReason', '示例科技模糊命中上海示例科技有限公司高风险关联，需暂停承办并发起冲突审批。',
            'requiresApproval', TRUE
        ),
        TRUE
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE check_id = 'LAWYER-TRIAL-HIGH-001';

UPDATE conflict_cases
SET conflict_type = '现有客户高风险关联',
    risk_level = 'HIGH',
    description = '示例科技模糊命中上海示例科技有限公司历史高风险事项，需冲突复核。',
    conflict_details = '验收种子：用于验证高风险检测结果。'
WHERE check_id = 'LAWYER-TRIAL-HIGH-001';

UPDATE approval_requests
SET content = '上海示例科技有限公司高风险冲突检测结果，请主任复核。',
    priority = 'high',
    conflict_risk_level = 'HIGH',
    conflict_result = jsonb_set(
        COALESCE(conflict_result, '{}'::jsonb),
        '{riskAssessment}',
        jsonb_build_object('overallRisk', 'HIGH', 'riskScore', 92, 'requiresApproval', TRUE),
        TRUE
    ),
    updated_at = CURRENT_TIMESTAMP
WHERE conflict_check_id = 'LAWYER-TRIAL-HIGH-001';
