DELETE FROM inbox_items
WHERE title = '待办：示例科技证据材料补充'
  AND source_type = 'case';

DELETE FROM client_trust_transactions
WHERE transaction_code = 'TT-DEMO-2026-001';

DELETE FROM client_trust_accounts
WHERE account_code = 'TA-DEMO-2026-001';

DELETE FROM cases
WHERE case_number IN (
    'DEMO-MVP-2026-001',
    'DEMO-MVP-2026-002',
    'DEMO-MVP-2026-003'
);

DELETE FROM clients
WHERE email IN (
    'contact@demo-tech.test',
    'legal@yunling-build.example',
    'zhao.haining@example.com'
);

DELETE FROM users
WHERE email IN (
    'demo.admin@example.test',
    'demo.lawyer@example.test',
    'demo.assistant@example.test',
    'demo.finance@example.test'
);
