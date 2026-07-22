DELETE FROM approval_snapshots WHERE approval_request_id = 'LAWYER-TRIAL-APPROVAL-001';
DELETE FROM approval_requests WHERE id = 'LAWYER-TRIAL-APPROVAL-001';
DELETE FROM conflict_cases WHERE id = 'LAWYER-TRIAL-HIGH-001-case-37';
DELETE FROM conflict_check_records WHERE check_id = 'LAWYER-TRIAL-HIGH-001';
DELETE FROM cases WHERE case_number = 'DEMO-HIGH-2026-001';
DELETE FROM cases WHERE case_number = 'DEMO-ISO-B-2026-001';
DELETE FROM clients WHERE email = 'lawyer-b-client@example.test';
DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM users WHERE email = 'demo.lawyer.b@example.test');
DELETE FROM users WHERE email = 'demo.lawyer.b@example.test';

UPDATE clients
SET
    name = '示例科技有限公司',
    company = '示例科技有限公司',
    updated_at = CURRENT_TIMESTAMP
WHERE email = 'contact@demo-tech.test'
  AND name = '上海示例科技有限公司';
