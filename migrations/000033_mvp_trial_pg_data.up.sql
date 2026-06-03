-- MVP trial data and tables for PostgreSQL.
-- Idempotent by design: it can be applied to an already seeded trial database.

CREATE TABLE IF NOT EXISTS inbox_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type VARCHAR(50) NOT NULL,
    source_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    due_date TIMESTAMPTZ,
    due_date_type VARCHAR(50),
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    is_completed BOOLEAN DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    reminder_sent BOOLEAN DEFAULT FALSE,
    reminder_count INT DEFAULT 0,
    escalated BOOLEAN DEFAULT FALSE,
    escalated_at TIMESTAMPTZ,
    snoozed_until TIMESTAMPTZ,
    snoozed_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS client_trust_accounts (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    client_id BIGINT NOT NULL REFERENCES clients(id) ON DELETE RESTRICT,
    account_code VARCHAR(50) NOT NULL UNIQUE,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    frozen_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    purpose_restriction VARCHAR(200),
    authorized_uses JSONB DEFAULT '[]'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    opened_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS client_trust_transactions (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    account_id BIGINT NOT NULL REFERENCES client_trust_accounts(id) ON DELETE CASCADE,
    transaction_code VARCHAR(50) NOT NULL UNIQUE,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    description TEXT NOT NULL,
    case_id BIGINT REFERENCES cases(id) ON DELETE SET NULL,
    purpose_code VARCHAR(50),
    recipient_name VARCHAR(200),
    recipient_bank_account VARCHAR(50),
    recipient_bank_name VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    completed_at TIMESTAMPTZ,
    attachment_id BIGINT,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    approved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_inbox_items_user_status ON inbox_items(user_id, is_completed, is_read);
CREATE INDEX IF NOT EXISTS idx_client_trust_accounts_client ON client_trust_accounts(client_id);
CREATE INDEX IF NOT EXISTS idx_client_trust_transactions_account ON client_trust_transactions(account_id);

UPDATE roles
SET code = 'super_admin', status = 'active', updated_at = NOW()
WHERE name = '超级管理员' AND code = 'admin';

UPDATE roles
SET name = '管理员', code = 'admin', status = 'active', updated_at = NOW()
WHERE name = 'admin' AND COALESCE(code, '') = '';

UPDATE roles
SET name = '助理', code = 'assistant', status = 'active', updated_at = NOW()
WHERE name = 'assistant' AND COALESCE(code, '') = '';

INSERT INTO roles (name, code, description, status, sort_order)
SELECT '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', 'active', 1
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'super_admin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO roles (name, code, description, status, sort_order)
SELECT '管理员', 'admin', '系统管理员，拥有管理权限', 'active', 2
WHERE NOT EXISTS (SELECT 1 FROM roles WHERE code = 'admin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO roles (name, code, description, status, sort_order)
VALUES
    ('财务', 'finance', '财务人员，负责财务和代管款管理', 'active', 5),
    ('实习生', 'intern', '实习人员，拥有基础查看权限', 'active', 6)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO permissions (name, code, type, path, component, sort_order, status)
VALUES
    ('仪表盘', 'dashboard', 'menu', '/dashboard', 'Dashboard', 1, 'active'),
    ('用户管理', 'user_management', 'menu', '/admin/users', 'UserManagement', 2, 'active'),
    ('角色管理', 'role_management', 'menu', '/admin/roles', 'RoleManagement', 3, 'active'),
    ('权限管理', 'permission_management', 'menu', '/admin/permissions', 'PermissionManagement', 4, 'active'),
    ('客户管理', 'client_management', 'menu', '/clients', 'ClientManagement', 5, 'active'),
    ('案件管理', 'case_management', 'menu', '/cases', 'CaseManagement', 6, 'active'),
    ('审批中心', 'approval_center', 'menu', '/approvals', 'ApprovalCenter', 7, 'active'),
    ('财务管理', 'finance_management', 'menu', '/finance', 'FinanceManagement', 8, 'active'),
    ('文档管理', 'document_management', 'menu', '/documents', 'DocumentManagement', 9, 'active'),
    ('工具中心', 'tools_center', 'menu', '/tools', 'ToolsCenter', 10, 'active'),
    ('系统设置', 'system_settings', 'menu', '/settings', 'SystemSettings', 11, 'active'),
    ('统计报表', 'statistics_reports', 'menu', '/statistics', 'StatisticsReports', 12, 'active'),
    ('律师管理', 'lawyer_management', 'menu', '/lawyers', 'LawyerManagement', 13, 'active'),
    ('利益冲突检查', 'conflict_check', 'menu', '/conflict/v2', 'ConflictCheckEnhanced', 14, 'active'),
    ('代管款管理', 'trust_management', 'menu', '/trust', 'TrustManagement', 15, 'active'),
    ('查看用户', 'user:view', 'button', NULL, NULL, 1, 'active'),
    ('创建用户', 'user:create', 'button', NULL, NULL, 2, 'active'),
    ('编辑用户', 'user:edit', 'button', NULL, NULL, 3, 'active'),
    ('删除用户', 'user:delete', 'button', NULL, NULL, 4, 'active'),
    ('查看角色', 'role:view', 'button', NULL, NULL, 1, 'active'),
    ('创建角色', 'role:create', 'button', NULL, NULL, 2, 'active'),
    ('编辑角色', 'role:edit', 'button', NULL, NULL, 3, 'active'),
    ('删除角色', 'role:delete', 'button', NULL, NULL, 4, 'active'),
    ('查看客户', 'client:view', 'button', NULL, NULL, 1, 'active'),
    ('创建客户', 'client:create', 'button', NULL, NULL, 2, 'active'),
    ('编辑客户', 'client:edit', 'button', NULL, NULL, 3, 'active'),
    ('删除客户', 'client:delete', 'button', NULL, NULL, 4, 'active'),
    ('查看案件', 'case:view', 'button', NULL, NULL, 1, 'active'),
    ('创建案件', 'case:create', 'button', NULL, NULL, 2, 'active'),
    ('编辑案件', 'case:edit', 'button', NULL, NULL, 3, 'active'),
    ('删除案件', 'case:delete', 'button', NULL, NULL, 4, 'active'),
    ('分配律师', 'case:assign', 'button', NULL, NULL, 5, 'active'),
    ('查看财务', 'finance:view', 'button', NULL, NULL, 1, 'active'),
    ('创建财务记录', 'finance:create', 'button', NULL, NULL, 2, 'active'),
    ('编辑财务记录', 'finance:edit', 'button', NULL, NULL, 3, 'active'),
    ('查看文档', 'document:view', 'button', NULL, NULL, 1, 'active'),
    ('上传文档', 'document:upload', 'button', NULL, NULL, 2, 'active'),
    ('编辑文档', 'document:edit', 'button', NULL, NULL, 3, 'active'),
    ('删除文档', 'document:delete', 'button', NULL, NULL, 4, 'active'),
    ('查看审批', 'approval:view', 'button', NULL, NULL, 1, 'active'),
    ('管理审批', 'approval:manage', 'button', NULL, NULL, 2, 'active'),
    ('执行冲突检查', 'conflict:check', 'button', NULL, NULL, 1, 'active'),
    ('查看代管款', 'trust:view', 'button', NULL, NULL, 1, 'active'),
    ('管理代管款', 'trust:manage', 'button', NULL, NULL, 2, 'active'),
    ('查看工具', 'tools:view', 'button', NULL, NULL, 1, 'active'),
    ('查看报表', 'reports:view', 'button', NULL, NULL, 1, 'active')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    path = EXCLUDED.path,
    component = EXCLUDED.component,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW();

WITH role_permission_codes(role_code, permission_code) AS (
    VALUES
        ('super_admin', 'dashboard'), ('super_admin', 'user_management'), ('super_admin', 'role_management'), ('super_admin', 'permission_management'), ('super_admin', 'client_management'), ('super_admin', 'case_management'), ('super_admin', 'approval_center'), ('super_admin', 'finance_management'), ('super_admin', 'document_management'), ('super_admin', 'tools_center'), ('super_admin', 'system_settings'), ('super_admin', 'statistics_reports'),
        ('super_admin', 'user:view'), ('super_admin', 'user:create'), ('super_admin', 'user:edit'), ('super_admin', 'user:delete'), ('super_admin', 'role:view'), ('super_admin', 'role:create'), ('super_admin', 'role:edit'), ('super_admin', 'role:delete'), ('super_admin', 'client:view'), ('super_admin', 'client:create'), ('super_admin', 'client:edit'), ('super_admin', 'client:delete'), ('super_admin', 'case:view'), ('super_admin', 'case:create'), ('super_admin', 'case:edit'), ('super_admin', 'case:delete'), ('super_admin', 'case:assign'), ('super_admin', 'finance:view'), ('super_admin', 'finance:create'), ('super_admin', 'finance:edit'), ('super_admin', 'document:view'), ('super_admin', 'document:upload'), ('super_admin', 'document:edit'), ('super_admin', 'document:delete'),
        ('super_admin', 'lawyer_management'), ('super_admin', 'conflict_check'), ('super_admin', 'trust_management'),
        ('super_admin', 'approval:view'), ('super_admin', 'approval:manage'), ('super_admin', 'conflict:check'),
        ('super_admin', 'trust:view'), ('super_admin', 'trust:manage'), ('super_admin', 'tools:view'), ('super_admin', 'reports:view'),
        ('admin', 'dashboard'), ('admin', 'user_management'), ('admin', 'role_management'), ('admin', 'permission_management'), ('admin', 'client_management'), ('admin', 'case_management'), ('admin', 'approval_center'), ('admin', 'finance_management'), ('admin', 'document_management'), ('admin', 'tools_center'), ('admin', 'system_settings'), ('admin', 'statistics_reports'),
        ('admin', 'user:view'), ('admin', 'user:create'), ('admin', 'user:edit'), ('admin', 'user:delete'), ('admin', 'role:view'), ('admin', 'role:create'), ('admin', 'role:edit'), ('admin', 'role:delete'), ('admin', 'client:view'), ('admin', 'client:create'), ('admin', 'client:edit'), ('admin', 'client:delete'), ('admin', 'case:view'), ('admin', 'case:create'), ('admin', 'case:edit'), ('admin', 'case:delete'), ('admin', 'case:assign'), ('admin', 'finance:view'), ('admin', 'finance:create'), ('admin', 'finance:edit'), ('admin', 'document:view'), ('admin', 'document:upload'), ('admin', 'document:edit'), ('admin', 'document:delete'),
        ('admin', 'lawyer_management'), ('admin', 'conflict_check'), ('admin', 'trust_management'),
        ('admin', 'approval:view'), ('admin', 'approval:manage'), ('admin', 'conflict:check'),
        ('admin', 'trust:view'), ('admin', 'trust:manage'), ('admin', 'tools:view'), ('admin', 'reports:view'),
        ('lawyer', 'dashboard'), ('lawyer', 'client_management'), ('lawyer', 'case_management'), ('lawyer', 'approval_center'), ('lawyer', 'document_management'), ('lawyer', 'tools_center'), ('lawyer', 'statistics_reports'),
        ('lawyer', 'client:view'), ('lawyer', 'client:create'), ('lawyer', 'client:edit'), ('lawyer', 'case:view'), ('lawyer', 'case:create'), ('lawyer', 'case:edit'), ('lawyer', 'document:view'), ('lawyer', 'document:upload'), ('lawyer', 'document:edit'),
        ('lawyer', 'conflict_check'), ('lawyer', 'approval:view'), ('lawyer', 'approval:manage'), ('lawyer', 'conflict:check'), ('lawyer', 'tools:view'), ('lawyer', 'reports:view'),
        ('assistant', 'dashboard'), ('assistant', 'client_management'), ('assistant', 'case_management'), ('assistant', 'document_management'), ('assistant', 'tools_center'), ('assistant', 'client:view'), ('assistant', 'case:view'), ('assistant', 'document:view'), ('assistant', 'tools:view'),
        ('finance', 'dashboard'), ('finance', 'finance_management'), ('finance', 'statistics_reports'), ('finance', 'finance:view'), ('finance', 'finance:create'), ('finance', 'finance:edit'),
        ('finance', 'trust_management'), ('finance', 'trust:view'), ('finance', 'trust:manage'), ('finance', 'reports:view'),
        ('intern', 'dashboard'), ('intern', 'tools_center'), ('intern', 'tools:view'),
        ('user', 'dashboard'), ('user', 'tools_center'), ('user', 'tools:view')
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_codes rpc
JOIN roles r ON r.code = rpc.role_code
JOIN permissions p ON p.code = rpc.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, COALESCE(mapped_role.id, fallback_role.id)
FROM users u
LEFT JOIN roles mapped_role ON mapped_role.code = u.role
JOIN roles fallback_role ON fallback_role.code = 'user'
WHERE u.deleted_at IS NULL
ON CONFLICT (user_id, role_id) DO NOTHING;

WITH seed_user AS (
    SELECT id FROM users ORDER BY id LIMIT 1
),
seed_client AS (
    SELECT id FROM clients ORDER BY id LIMIT 1
)
INSERT INTO client_trust_accounts (
    client_id, account_code, balance, currency, frozen_amount,
    purpose_restriction, authorized_uses, status, opened_at
)
SELECT
    seed_client.id,
    'TA-MVP-001',
    50000.00,
    'CNY',
    0,
    '仅限本案诉讼费、保全费、调查费支出',
    '["court_fee","preservation_fee","investigation_fee"]'::jsonb,
    'active',
    now()
FROM seed_client
ON CONFLICT (account_code) DO NOTHING;

WITH seed_user AS (
    SELECT id FROM users ORDER BY id LIMIT 1
),
seed_case AS (
    SELECT id FROM cases ORDER BY id LIMIT 1
),
seed_account AS (
    SELECT id FROM client_trust_accounts WHERE account_code = 'TA-MVP-001'
)
INSERT INTO client_trust_transactions (
    account_id, transaction_code, transaction_type, amount, description,
    case_id, purpose_code, status, completed_at, created_by, approved_by, approved_at
)
SELECT
    seed_account.id,
    'TT-MVP-001',
    'deposit',
    50000.00,
    '客户代管款初始入账',
    seed_case.id,
    'court_fee',
    'completed',
    now(),
    seed_user.id,
    seed_user.id,
    now()
FROM seed_account, seed_user, seed_case
ON CONFLICT (transaction_code) DO NOTHING;

WITH seed_user AS (
    SELECT id FROM users ORDER BY id LIMIT 1
)
INSERT INTO inbox_items (
    user_id, source_type, source_id, title, content, priority, due_date, due_date_type
)
SELECT seed_user.id, 'approval', 1, '待处理：立案审批', '请审核客户立案申请及利益冲突检查结果。', 'high', now() + interval '1 day', 'approval'
FROM seed_user
WHERE NOT EXISTS (SELECT 1 FROM inbox_items WHERE source_type = 'approval' AND source_id = 1)
UNION ALL
SELECT seed_user.id, 'deadline', 1, '提醒：证据材料补充', '请在期限前补齐案件证据材料。', 'medium', now() + interval '3 days', 'evidence'
FROM seed_user
WHERE NOT EXISTS (SELECT 1 FROM inbox_items WHERE source_type = 'deadline' AND source_id = 1);
