-- Lawyer trial acceptance seed data.
-- Covers: fuzzy conflict seed, A/B role fixtures, and a director-review approval.

UPDATE clients
SET
    name = '上海示例科技有限公司',
    company = '上海示例科技有限公司',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'contact@demo-tech.test'
   OR name = '示例科技有限公司';

WITH trial_users(username, name, email, role, phone, department, seniority) AS (
    VALUES
        ('demo_lawyer_b', '王海平', 'demo.lawyer.b@example.test', 'lawyer', '13000001004', '争议解决部', '律师')
)
INSERT INTO users (username, name, email, password, role, phone, status, department, seniority, created_at, updated_at)
SELECT
    username,
    name,
    email,
    '$2a$10$h3ezJo0AzoxXQySkfqneQ.WwxYOq3lV5rJhEzDobnZ38EVG9znY2O',
    role,
    phone,
    'active',
    department,
    seniority,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM trial_users
ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    name = EXCLUDED.name,
    password = EXCLUDED.password,
    role = EXCLUDED.role,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    department = EXCLUDED.department,
    seniority = EXCLUDED.seniority,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO user_roles (user_id, role_id, created_at)
SELECT u.id, r.id, CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.code = 'lawyer'
WHERE u.email = 'demo.lawyer.b@example.test'
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO clients (
    name, type, email, phone, address, company, industry,
    contact_person, contact_phone, source, notes, status, version, created_at, updated_at
)
VALUES (
    '示例隔离测试客户B', '企业', 'lawyer-b-client@example.test', '0971-6001099',
    '青海省示例市隔离测试路2号', '示例隔离测试客户B', '验收测试',
    '王海平', '13000001004', 'lawyer_trial_acceptance_seed',
    '用于验证 Lawyer A/B 案件数据隔离。', 'active', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
ON CONFLICT (email) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    phone = EXCLUDED.phone,
    address = EXCLUDED.address,
    company = EXCLUDED.company,
    industry = EXCLUDED.industry,
    contact_person = EXCLUDED.contact_person,
    contact_phone = EXCLUDED.contact_phone,
    source = EXCLUDED.source,
    notes = EXCLUDED.notes,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP;

DELETE FROM cases WHERE case_number = 'HD-ISO-B-2026-001';
DELETE FROM cases WHERE case_number = 'HD-HIGH-2026-001';

WITH seed AS (
    SELECT
        c.id AS client_id,
        u.id AS lawyer_id
    FROM clients c
    CROSS JOIN users u
    WHERE c.email = 'lawyer-b-client@example.test'
      AND u.email = 'demo.lawyer.b@example.test'
)
INSERT INTO cases (
    case_number, title, description, client_id, lawyer_id, case_type,
    priority, status, start_date, created_at, updated_at
)
SELECT
    'HD-ISO-B-2026-001',
    'Lawyer B 独立隔离验收案件',
    '用于验证律师 A 不能访问律师 B 的案件。',
    seed.client_id,
    seed.lawyer_id,
    'commercial',
    'medium',
    'active',
    CURRENT_DATE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM seed;

WITH seed AS (
    SELECT
        c.id AS client_id,
        u.id AS lawyer_id
    FROM clients c
    CROSS JOIN users u
    WHERE c.email = 'lawyer-b-client@example.test'
      AND u.email = 'demo.lawyer.b@example.test'
)
INSERT INTO cases (
    case_number, title, description, client_id, lawyer_id, case_type,
    priority, status, start_date, created_at, updated_at
)
SELECT
    'HD-HIGH-2026-001',
    '上海示例科技高风险历史事项',
    '示例科技与上海示例科技有限公司存在历史高风险关联，用于验证高风险冲突硬阻断。',
    seed.client_id,
    seed.lawyer_id,
    'commercial',
    'high',
    'active',
    CURRENT_DATE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM seed;

WITH seed AS (
    SELECT
        c.id::text AS client_id,
        c.name AS client_name,
        u.id AS lawyer_id
    FROM clients c
    CROSS JOIN users u
    WHERE c.email = 'contact@demo-tech.test'
      AND u.email = 'demo.lawyer@example.test'
    LIMIT 1
)
INSERT INTO conflict_check_records (
    check_id, client_id, client_name, case_name, case_type, check_status,
    has_conflict, risk_level, search_parameters, check_result, user_id,
    duration, check_time, created_at, updated_at
)
SELECT
    'LAWYER-TRIAL-HIGH-001',
    seed.client_id,
    seed.client_name,
    '新增案件联调测试',
    'commercial',
    'COMPLETED',
    TRUE,
    'HIGH',
    jsonb_build_object(
        'query', '示例科技',
        'searchDepth', 'STANDARD',
        'searchYears', 5,
        'includeCorporateRelations', TRUE,
        'matchedClientName', '上海示例科技有限公司'
    ),
    jsonb_build_object(
        'checkId', 'LAWYER-TRIAL-HIGH-001',
        'riskAssessment', jsonb_build_object(
            'overallRisk', 'HIGH',
            'riskScore', 92,
            'riskReason', '示例科技模糊命中上海示例科技有限公司高风险关联，需暂停承办并发起冲突审批。',
            'requiresApproval', TRUE
        ),
        'checkStatistics', jsonb_build_object(
            'totalCasesChecked', 1,
            'relatedPartiesChecked', 1
        )
    ),
    seed.lawyer_id,
    358,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM seed
ON CONFLICT (check_id) DO UPDATE SET
    client_id = EXCLUDED.client_id,
    client_name = EXCLUDED.client_name,
    case_name = EXCLUDED.case_name,
    case_type = EXCLUDED.case_type,
    check_status = EXCLUDED.check_status,
    has_conflict = EXCLUDED.has_conflict,
    risk_level = EXCLUDED.risk_level,
    search_parameters = EXCLUDED.search_parameters,
    check_result = EXCLUDED.check_result,
    user_id = EXCLUDED.user_id,
    duration = EXCLUDED.duration,
    check_time = EXCLUDED.check_time,
    updated_at = CURRENT_TIMESTAMP;

WITH seed AS (
    SELECT
        c.id::text AS client_id,
        cs.id::text AS case_id
    FROM clients c
    LEFT JOIN cases cs ON cs.case_number = 'CASE-20260513173242'
    WHERE c.email = 'contact@demo-tech.test'
    LIMIT 1
)
INSERT INTO conflict_cases (
    id, check_id, case_id, case_name, case_no, conflict_type, risk_level,
    description, case_status, client_id, opposing_parties, conflict_details,
    created_at, case_type
)
SELECT
    'LAWYER-TRIAL-HIGH-001-case-37',
    'LAWYER-TRIAL-HIGH-001',
    COALESCE(seed.case_id, '37'),
    '新增案件联调测试',
    'CASE-20260513173242',
    '现有客户高风险关联',
    'HIGH',
    '示例科技模糊命中上海示例科技有限公司历史高风险事项，需冲突复核。',
    'pending',
    seed.client_id,
    '["上海示例科技有限公司", "示例科技"]'::jsonb,
    '验收种子：用于验证律师从案件详情进入本案冲突复核时可以看到高风险检测结果。',
    CURRENT_TIMESTAMP,
    'commercial'
FROM seed
ON CONFLICT (id) DO UPDATE SET
    check_id = EXCLUDED.check_id,
    case_id = EXCLUDED.case_id,
    case_name = EXCLUDED.case_name,
    case_no = EXCLUDED.case_no,
    conflict_type = EXCLUDED.conflict_type,
    risk_level = EXCLUDED.risk_level,
    description = EXCLUDED.description,
    case_status = EXCLUDED.case_status,
    client_id = EXCLUDED.client_id,
    opposing_parties = EXCLUDED.opposing_parties,
    conflict_details = EXCLUDED.conflict_details,
    created_at = EXCLUDED.created_at,
    case_type = EXCLUDED.case_type;

WITH users_seed AS (
    SELECT
        (SELECT id::text FROM users WHERE email = 'demo.lawyer@example.test' LIMIT 1) AS lawyer_id,
        (SELECT name FROM users WHERE email = 'demo.lawyer@example.test' LIMIT 1) AS lawyer_name,
        (SELECT id::text FROM users WHERE email = 'demo.admin@example.test' LIMIT 1) AS director_id,
        (SELECT name FROM users WHERE email = 'demo.admin@example.test' LIMIT 1) AS director_name
)
INSERT INTO approval_requests (
    id, request_number, title, type, category, content, applicant_id,
    applicant_name, applicant_title, department_id, department_name, urgency,
    priority, status, submission_date, current_stage, current_approver_id,
    current_approver_name, workflow_type, workflow_config, attachments,
    metadata, created_by, created_at, updated_at, conflict_check_id,
    conflict_risk_level, conflict_check_time, conflict_result
)
SELECT
    'LAWYER-TRIAL-APPROVAL-001',
    'APR-LAWYER-TRIAL-001',
    '冲突审查审批 - 新增案件联调测试',
    'conflict_approval',
    'conflict_review',
    '上海示例科技有限公司高风险冲突检测结果，请主任复核。',
    users_seed.lawyer_id,
    users_seed.lawyer_name,
    '律师',
    'risk',
    '合规风控部',
    'urgent',
    'high',
    'submitted',
    CURRENT_TIMESTAMP,
    '主任复核',
    users_seed.director_id,
    users_seed.director_name,
    'CONFLICT_APPROVAL',
    '{}'::jsonb,
    '[]'::jsonb,
    jsonb_build_object(
        'source', 'lawyer_trial_acceptance_seed',
        'conflict_task_id', 'LAWYER-TRIAL-HIGH-001',
        'conflict_result', jsonb_build_object(
            'checkId', 'LAWYER-TRIAL-HIGH-001',
            'riskAssessment', jsonb_build_object('overallRisk', 'HIGH', 'riskScore', 92, 'requiresApproval', TRUE)
        )
    ),
    users_seed.lawyer_id,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    'LAWYER-TRIAL-HIGH-001',
    'HIGH',
    CURRENT_TIMESTAMP,
    jsonb_build_object(
        'checkId', 'LAWYER-TRIAL-HIGH-001',
        'riskAssessment', jsonb_build_object('overallRisk', 'HIGH', 'riskScore', 92, 'requiresApproval', TRUE)
    )
FROM users_seed
WHERE users_seed.lawyer_id IS NOT NULL
  AND users_seed.director_id IS NOT NULL
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    status = EXCLUDED.status,
    current_stage = EXCLUDED.current_stage,
    current_approver_id = EXCLUDED.current_approver_id,
    current_approver_name = EXCLUDED.current_approver_name,
    metadata = EXCLUDED.metadata,
    conflict_check_id = EXCLUDED.conflict_check_id,
    conflict_risk_level = EXCLUDED.conflict_risk_level,
    conflict_check_time = EXCLUDED.conflict_check_time,
    conflict_result = EXCLUDED.conflict_result,
    updated_at = CURRENT_TIMESTAMP;
