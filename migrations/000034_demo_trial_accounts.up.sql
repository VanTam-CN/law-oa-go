-- Explicit trial accounts and branded PostgreSQL seed data for 示例律师事务所OA.
-- Password for all trial accounts: Demo@2026

INSERT INTO roles (name, code, description, status, sort_order)
VALUES
    ('管理员', 'admin', '示例律师事务所OA管理员，负责用户、角色和权限配置', 'active', 2),
    ('律师', 'lawyer', '示例律师事务所执业律师，负责客户和案件办理', 'active', 3),
    ('助理', 'assistant', '示例律师事务所律师助理，负责案件协同和材料跟进', 'active', 4),
    ('财务', 'finance', '示例律师事务所财务人员，负责收费和代管款管理', 'active', 5)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

WITH trial_users(username, name, email, role, phone, department, seniority) AS (
    VALUES
        ('demo_admin', '示例管理员', 'demo.admin@example.test', 'admin', '13000001000', '综合管理部', '合伙人'),
        ('demo_lawyer', '陈示例', 'demo.lawyer@example.test', 'lawyer', '13000001001', '争议解决部', '合伙人'),
        ('demo_assistant', '李若海', 'demo.assistant@example.test', 'assistant', '13000001002', '诉讼支持部', '中级'),
        ('demo_finance', '周明珠', 'demo.finance@example.test', 'finance', '13000001003', '财务部', '高级')
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
    NOW(),
    NOW()
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
    updated_at = NOW();

WITH expected_user_roles AS (
    SELECT u.id AS user_id, r.id AS role_id
    FROM users u
    JOIN roles r ON r.code = u.role
    WHERE u.email IN (
        'demo.admin@example.test',
        'demo.lawyer@example.test',
        'demo.assistant@example.test',
        'demo.finance@example.test'
    )
)
DELETE FROM user_roles ur
USING expected_user_roles eur
WHERE ur.user_id = eur.user_id
  AND ur.role_id <> eur.role_id;

WITH expected_user_roles AS (
    SELECT u.id AS user_id, r.id AS role_id
    FROM users u
    JOIN roles r ON r.code = u.role
    WHERE u.email IN (
        'demo.admin@example.test',
        'demo.lawyer@example.test',
        'demo.assistant@example.test',
        'demo.finance@example.test'
    )
)
INSERT INTO user_roles (user_id, role_id, created_at)
SELECT user_id, role_id, NOW()
FROM expected_user_roles
ON CONFLICT (user_id, role_id) DO NOTHING;

INSERT INTO clients (
    name, type, email, phone, address, company, industry,
    contact_person, contact_phone, source, notes, status, version, created_at, updated_at
)
VALUES
    ('示例科技有限公司', '企业', 'contact@demo-tech.test', '0971-6001001', '青海省西宁市城西区五四西路88号', '示例科技有限公司', '软件与信息服务', '林晓东', '13000002001', 'trial_seed', '示例律师事务所OA试用客户：服务合同纠纷与常年顾问', 'active', 1, NOW(), NOW()),
    ('青海云岭建设集团有限公司', '企业', 'legal@yunling-build.example', '0971-6001002', '青海省西宁市城中区南关街56号', '青海云岭建设集团有限公司', '建筑工程', '马立峰', '13000002002', 'trial_seed', '示例律师事务所OA试用客户：建设工程款争议', 'active', 1, NOW(), NOW()),
    ('赵海宁', '个人', 'zhao.haining@example.com', '13000002003', '青海省示例市平安区平安大道18号', NULL, '个人劳动争议', '赵海宁', '13000002003', 'trial_seed', '示例律师事务所OA试用客户：劳动争议咨询', 'active', 1, NOW(), NOW())
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
    updated_at = NOW();

WITH trial_cases(case_number, title, description, client_email, case_type, priority, status) AS (
    VALUES
        ('HD-MVP-2026-001', '示例科技服务合同纠纷', '客户主张软件实施服务合同尾款及违约金，需完成证据归集、诉讼费测算和立案审批。', 'contact@demo-tech.test', 'commercial', 'high', 'active'),
        ('HD-MVP-2026-002', '云岭建设工程款争议', '建设工程分包款结算争议，需做利益冲突检查、付款节点梳理和庭审准备。', 'legal@yunling-build.example', 'construction', 'urgent', 'pending'),
        ('HD-MVP-2026-003', '赵海宁劳动争议咨询', '个人客户劳动合同解除补偿咨询，需形成咨询记录和后续办案计划。', 'zhao.haining@example.com', 'labor', 'medium', 'active')
),
trial_lawyer AS (
    SELECT id FROM users WHERE email = 'demo.lawyer@example.test'
)
UPDATE cases existing_case
SET title = tc.title,
    description = tc.description,
    client_id = c.id,
    lawyer_id = trial_lawyer.id,
    case_type = tc.case_type,
    priority = tc.priority,
    status = tc.status,
    updated_at = NOW()
FROM trial_cases tc
JOIN clients c ON c.email = tc.client_email
CROSS JOIN trial_lawyer
WHERE existing_case.case_number = tc.case_number;

WITH trial_cases(case_number, title, description, client_email, case_type, priority, status) AS (
    VALUES
        ('HD-MVP-2026-001', '示例科技服务合同纠纷', '客户主张软件实施服务合同尾款及违约金，需完成证据归集、诉讼费测算和立案审批。', 'contact@demo-tech.test', 'commercial', 'high', 'active'),
        ('HD-MVP-2026-002', '云岭建设工程款争议', '建设工程分包款结算争议，需做利益冲突检查、付款节点梳理和庭审准备。', 'legal@yunling-build.example', 'construction', 'urgent', 'pending'),
        ('HD-MVP-2026-003', '赵海宁劳动争议咨询', '个人客户劳动合同解除补偿咨询，需形成咨询记录和后续办案计划。', 'zhao.haining@example.com', 'labor', 'medium', 'active')
),
trial_lawyer AS (
    SELECT id FROM users WHERE email = 'demo.lawyer@example.test'
)
INSERT INTO cases (
    case_number, title, description, client_id, lawyer_id, case_type,
    priority, status, start_date, ethical_wall_enabled, created_by, created_at, updated_at
)
SELECT
    tc.case_number,
    tc.title,
    tc.description,
    c.id,
    trial_lawyer.id,
    tc.case_type,
    tc.priority,
    tc.status,
    NOW(),
    false,
    'migration:000034',
    NOW(),
    NOW()
FROM trial_cases tc
JOIN clients c ON c.email = tc.client_email
CROSS JOIN trial_lawyer
WHERE NOT EXISTS (
    SELECT 1 FROM cases existing_case
    WHERE existing_case.case_number = tc.case_number
);

WITH trial_client AS (
    SELECT id FROM clients WHERE email = 'contact@demo-tech.test'
)
INSERT INTO client_trust_accounts (
    client_id, account_code, balance, currency, frozen_amount,
    purpose_restriction, authorized_uses, status, opened_at, created_at, updated_at
)
SELECT
    trial_client.id,
    'TA-HD-2026-001',
    120000.00,
    'CNY',
    30000.00,
    '仅限示例科技服务合同纠纷案诉讼费、保全费、调查费支出',
    '["court_fee","preservation_fee","investigation_fee"]'::jsonb,
    'active',
    NOW(),
    NOW(),
    NOW()
FROM trial_client
ON CONFLICT (account_code) DO UPDATE SET
    client_id = EXCLUDED.client_id,
    balance = EXCLUDED.balance,
    currency = EXCLUDED.currency,
    frozen_amount = EXCLUDED.frozen_amount,
    purpose_restriction = EXCLUDED.purpose_restriction,
    authorized_uses = EXCLUDED.authorized_uses,
    status = EXCLUDED.status,
    updated_at = NOW();

WITH trial_account AS (
    SELECT id FROM client_trust_accounts WHERE account_code = 'TA-HD-2026-001'
),
trial_case AS (
    SELECT id FROM cases WHERE case_number = 'HD-MVP-2026-001'
),
finance_user AS (
    SELECT id FROM users WHERE email = 'demo.finance@example.test'
),
admin_user AS (
    SELECT id FROM users WHERE email = 'demo.admin@example.test'
)
INSERT INTO client_trust_transactions (
    account_id, transaction_code, transaction_type, amount, description,
    case_id, purpose_code, status, completed_at, created_by, approved_by, approved_at,
    created_at, updated_at
)
SELECT
    trial_account.id,
    'TT-HD-2026-001',
    'deposit',
    120000.00,
    '示例科技诉讼专项代管款入账',
    trial_case.id,
    'court_fee',
    'completed',
    NOW(),
    finance_user.id,
    admin_user.id,
    NOW(),
    NOW(),
    NOW()
FROM trial_account, trial_case, finance_user, admin_user
ON CONFLICT (transaction_code) DO UPDATE SET
    account_id = EXCLUDED.account_id,
    amount = EXCLUDED.amount,
    description = EXCLUDED.description,
    case_id = EXCLUDED.case_id,
    purpose_code = EXCLUDED.purpose_code,
    status = EXCLUDED.status,
    approved_by = EXCLUDED.approved_by,
    approved_at = EXCLUDED.approved_at,
    updated_at = NOW();

WITH lawyer_user AS (
    SELECT id FROM users WHERE email = 'demo.lawyer@example.test'
),
trial_case AS (
    SELECT id FROM cases WHERE case_number = 'HD-MVP-2026-001'
)
INSERT INTO inbox_items (
    user_id, source_type, source_id, title, content, priority, due_date, due_date_type, created_at, updated_at
)
SELECT
    lawyer_user.id,
    'case',
    trial_case.id,
    '待办：示例科技证据材料补充',
    '请补齐服务合同、验收记录、付款流水和往来函件。',
    'high',
    NOW() + INTERVAL '2 days',
    'evidence',
    NOW(),
    NOW()
FROM lawyer_user, trial_case
WHERE NOT EXISTS (
    SELECT 1 FROM inbox_items
    WHERE source_type = 'case'
      AND source_id = trial_case.id
      AND title = '待办：示例科技证据材料补充'
);
