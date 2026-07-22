-- P0 conflict-review role and the third account used by the A/B/reviewer trial.
-- All identifiers below are fictional acceptance data.

INSERT INTO roles (name, code, description, status, sort_order)
VALUES ('独立冲突核查人', 'conflict_officer', '仅负责利益冲突证据复核，不具备一般案件管理权限', 'active', 8)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO permissions (name, code, type, parent_id, path, component, sort_order, status)
VALUES ('提交冲突复核', 'conflict:review', 'button', NULL, NULL, NULL, 3, 'active')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW();

WITH role_permission_codes(permission_code) AS (
    VALUES ('dashboard'), ('conflict_check'), ('conflict:check'), ('conflict:review')
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM role_permission_codes rpc
JOIN roles r ON r.code = 'conflict_officer'
JOIN permissions p ON p.code = rpc.permission_code
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO users (
    username, name, email, password, role, phone, status, department, seniority, created_at, updated_at
)
VALUES (
    'demo_conflict_officer', '独立冲突核查人', 'demo.conflict.officer@example.test',
    '$2a$10$h3ezJo0AzoxXQySkfqneQ.WwxYOq3lV5rJhEzDobnZ38EVG9znY2O',
    'conflict_officer', '13000001005', 'active', '合规风控部', '合伙人', NOW(), NOW()
)
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

WITH officer AS (
    SELECT u.id AS user_id, r.id AS role_id
    FROM users u
    JOIN roles r ON r.code = 'conflict_officer'
    WHERE u.email = 'demo.conflict.officer@example.test'
)
DELETE FROM user_roles ur
USING officer
WHERE ur.user_id = officer.user_id
  AND ur.role_id <> officer.role_id;

INSERT INTO user_roles (user_id, role_id, created_at, updated_at)
SELECT u.id, r.id, NOW(), NOW()
FROM users u
JOIN roles r ON r.code = 'conflict_officer'
WHERE u.email = 'demo.conflict.officer@example.test'
ON CONFLICT (user_id, role_id) DO NOTHING;
