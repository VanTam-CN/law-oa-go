DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM users WHERE email = 'demo.conflict.officer@example.test');

DELETE FROM users WHERE email = 'demo.conflict.officer@example.test';

DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE code = 'conflict_officer');

DELETE FROM roles WHERE code = 'conflict_officer';

DELETE FROM permissions WHERE code = 'conflict:review';
