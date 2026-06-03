WITH seeded_permissions(code) AS (
    VALUES
        ('dashboard'), ('user_management'), ('role_management'), ('permission_management'), ('client_management'), ('case_management'), ('approval_center'), ('finance_management'), ('document_management'), ('tools_center'), ('system_settings'), ('statistics_reports'),
        ('user:view'), ('user:create'), ('user:edit'), ('user:delete'), ('role:view'), ('role:create'), ('role:edit'), ('role:delete'), ('client:view'), ('client:create'), ('client:edit'), ('client:delete'), ('case:view'), ('case:create'), ('case:edit'), ('case:delete'), ('case:assign'), ('finance:view'), ('finance:create'), ('finance:edit'), ('document:view'), ('document:upload'), ('document:edit'), ('document:delete')
),
seeded_roles(code) AS (
    VALUES ('super_admin'), ('admin'), ('lawyer'), ('assistant'), ('finance'), ('intern'), ('user')
)
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE code IN (SELECT code FROM seeded_roles))
   OR permission_id IN (SELECT id FROM permissions WHERE code IN (SELECT code FROM seeded_permissions));

WITH seeded_roles(code) AS (
    VALUES ('super_admin'), ('admin'), ('lawyer'), ('assistant'), ('finance'), ('intern'), ('user')
)
DELETE FROM user_roles
WHERE role_id IN (SELECT id FROM roles WHERE code IN (SELECT code FROM seeded_roles));

WITH seeded_permissions(code) AS (
    VALUES
        ('user:view'), ('user:create'), ('user:edit'), ('user:delete'), ('role:view'), ('role:create'), ('role:edit'), ('role:delete'), ('client:view'), ('client:create'), ('client:edit'), ('client:delete'), ('case:view'), ('case:create'), ('case:edit'), ('case:delete'), ('case:assign'), ('finance:view'), ('finance:create'), ('finance:edit'), ('document:view'), ('document:upload'), ('document:edit'), ('document:delete'),
        ('dashboard'), ('user_management'), ('role_management'), ('permission_management'), ('client_management'), ('case_management'), ('approval_center'), ('finance_management'), ('document_management'), ('tools_center'), ('system_settings'), ('statistics_reports')
)
DELETE FROM permissions
WHERE code IN (SELECT code FROM seeded_permissions);

WITH seeded_roles(code) AS (
    VALUES ('super_admin'), ('admin'), ('lawyer'), ('assistant'), ('finance'), ('intern'), ('user')
)
DELETE FROM roles
WHERE code IN (SELECT code FROM seeded_roles);
