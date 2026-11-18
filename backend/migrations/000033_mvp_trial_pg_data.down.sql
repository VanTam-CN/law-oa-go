DELETE FROM inbox_items
WHERE (source_type = 'approval' AND source_id = 1)
   OR (source_type = 'deadline' AND source_id = 1);

DELETE FROM client_trust_transactions
WHERE transaction_code = 'TT-MVP-001';

DELETE FROM client_trust_accounts
WHERE account_code = 'TA-MVP-001';

DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE code IN (
        'dashboard',
        'user_management',
        'role_management',
        'permission_management',
        'client_management',
        'case_management',
        'approval_center',
        'finance_management',
        'document_management',
        'tools_center',
        'system_settings',
        'statistics_reports',
        'lawyer_management',
        'conflict_check',
        'trust_management',
        'user:view',
        'user:create',
        'user:edit',
        'user:delete',
        'role:view',
        'role:create',
        'role:edit',
        'role:delete',
        'client:view',
        'client:create',
        'client:edit',
        'client:delete',
        'case:view',
        'case:create',
        'case:edit',
        'case:delete',
        'case:assign',
        'finance:view',
        'finance:create',
        'finance:edit',
        'document:view',
        'document:upload',
        'document:edit',
        'document:delete',
        'approval:view',
        'approval:manage',
        'conflict:check',
        'trust:view',
        'trust:manage',
        'tools:view',
        'reports:view'
    )
);

DELETE FROM permissions
WHERE code IN (
    'dashboard',
    'user_management',
    'role_management',
    'permission_management',
    'client_management',
    'case_management',
    'approval_center',
    'finance_management',
    'document_management',
    'tools_center',
    'system_settings',
    'statistics_reports',
    'lawyer_management',
    'conflict_check',
    'trust_management',
    'user:view',
    'user:create',
    'user:edit',
    'user:delete',
    'role:view',
    'role:create',
    'role:edit',
    'role:delete',
    'client:view',
    'client:create',
    'client:edit',
    'client:delete',
    'case:view',
    'case:create',
    'case:edit',
    'case:delete',
    'case:assign',
    'finance:view',
    'finance:create',
    'finance:edit',
    'document:view',
    'document:upload',
    'document:edit',
    'document:delete',
    'approval:view',
    'approval:manage',
    'conflict:check',
    'trust:view',
    'trust:manage',
    'tools:view',
    'reports:view'
);
