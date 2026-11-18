-- PostgreSQL-compatible RBAC schema and trial seed data.

CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_code ON roles(code);
CREATE INDEX IF NOT EXISTS idx_roles_status ON roles(status);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL DEFAULT 'menu',
    parent_id BIGINT REFERENCES permissions(id) ON DELETE SET NULL,
    path VARCHAR(255),
    icon VARCHAR(100),
    component VARCHAR(255),
    sort_order INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_permissions_code ON permissions(code);
CREATE INDEX IF NOT EXISTS idx_permissions_parent_id ON permissions(parent_id);
CREATE INDEX IF NOT EXISTS idx_permissions_type ON permissions(type);
CREATE INDEX IF NOT EXISTS idx_permissions_status ON permissions(status);
CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON permissions(deleted_at);

CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_role_permissions_role_permission ON role_permissions(role_id, permission_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_user_role ON user_roles(user_id, role_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

INSERT INTO roles (name, code, description, status, sort_order)
VALUES
    ('超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', 'active', 1),
    ('管理员', 'admin', '系统管理员，拥有大部分管理权限', 'active', 2),
    ('律师', 'lawyer', '律师用户，可以管理案件和客户', 'active', 3),
    ('助理', 'assistant', '律师助理，协助律师处理案件', 'active', 4),
    ('财务', 'finance', '财务人员，负责财务管理', 'active', 5),
    ('实习生', 'intern', '实习人员，拥有基础查看权限', 'active', 6),
    ('普通用户', 'user', '普通试用账号，拥有基础查看权限', 'active', 7)
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();

INSERT INTO permissions (name, code, type, parent_id, path, icon, component, sort_order, status)
VALUES
    ('仪表盘', 'dashboard', 'menu', NULL, '/dashboard', 'dashboard', 'Dashboard', 1, 'active'),
    ('用户管理', 'user_management', 'menu', NULL, '/user', 'users', 'UserManagement', 2, 'active'),
    ('角色管理', 'role_management', 'menu', NULL, '/admin/roles', 'role', 'RoleManagement', 3, 'active'),
    ('权限管理', 'permission_management', 'menu', NULL, '/admin/permissions', 'permission', 'PermissionManagement', 4, 'active'),
    ('客户管理', 'client_management', 'menu', NULL, '/client', 'clients', 'ClientManagement', 5, 'active'),
    ('案件管理', 'case_management', 'menu', NULL, '/case', 'cases', 'CaseManagement', 6, 'active'),
    ('审批中心', 'approval_center', 'menu', NULL, '/approval', 'approval', 'ApprovalCenter', 7, 'active'),
    ('财务管理', 'finance_management', 'menu', NULL, '/finance', 'finance', 'FinanceManagement', 8, 'active'),
    ('文档管理', 'document_management', 'menu', NULL, '/file', 'documents', 'DocumentManagement', 9, 'active'),
    ('工具中心', 'tools_center', 'menu', NULL, '/tools', 'tools', 'ToolsCenter', 10, 'active'),
    ('系统设置', 'system_settings', 'menu', NULL, '/settings', 'settings', 'SystemSettings', 11, 'active'),
    ('统计报表', 'statistics_reports', 'menu', NULL, '/statistics', 'statistics', 'StatisticsReports', 12, 'active')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    parent_id = EXCLUDED.parent_id,
    path = EXCLUDED.path,
    icon = EXCLUDED.icon,
    component = EXCLUDED.component,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW();

WITH permission_seed(name, code, parent_code, sort_order) AS (
    VALUES
        ('查看用户', 'user:view', 'user_management', 1),
        ('创建用户', 'user:create', 'user_management', 2),
        ('编辑用户', 'user:edit', 'user_management', 3),
        ('删除用户', 'user:delete', 'user_management', 4),
        ('查看角色', 'role:view', 'role_management', 1),
        ('创建角色', 'role:create', 'role_management', 2),
        ('编辑角色', 'role:edit', 'role_management', 3),
        ('删除角色', 'role:delete', 'role_management', 4),
        ('查看客户', 'client:view', 'client_management', 1),
        ('创建客户', 'client:create', 'client_management', 2),
        ('编辑客户', 'client:edit', 'client_management', 3),
        ('删除客户', 'client:delete', 'client_management', 4),
        ('查看案件', 'case:view', 'case_management', 1),
        ('创建案件', 'case:create', 'case_management', 2),
        ('编辑案件', 'case:edit', 'case_management', 3),
        ('删除案件', 'case:delete', 'case_management', 4),
        ('分配律师', 'case:assign', 'case_management', 5),
        ('查看财务', 'finance:view', 'finance_management', 1),
        ('创建财务记录', 'finance:create', 'finance_management', 2),
        ('编辑财务记录', 'finance:edit', 'finance_management', 3),
        ('查看文档', 'document:view', 'document_management', 1),
        ('上传文档', 'document:upload', 'document_management', 2),
        ('编辑文档', 'document:edit', 'document_management', 3),
        ('删除文档', 'document:delete', 'document_management', 4)
)
INSERT INTO permissions (name, code, type, parent_id, sort_order, status)
SELECT ps.name, ps.code, 'button', parent.id, ps.sort_order, 'active'
FROM permission_seed ps
JOIN permissions parent ON parent.code = ps.parent_code
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    type = EXCLUDED.type,
    parent_id = EXCLUDED.parent_id,
    sort_order = EXCLUDED.sort_order,
    status = EXCLUDED.status,
    updated_at = NOW();

WITH role_permission_codes(role_code, permission_code) AS (
    VALUES
        ('super_admin', 'dashboard'), ('super_admin', 'user_management'), ('super_admin', 'role_management'), ('super_admin', 'permission_management'), ('super_admin', 'client_management'), ('super_admin', 'case_management'), ('super_admin', 'approval_center'), ('super_admin', 'finance_management'), ('super_admin', 'document_management'), ('super_admin', 'tools_center'), ('super_admin', 'system_settings'), ('super_admin', 'statistics_reports'),
        ('super_admin', 'user:view'), ('super_admin', 'user:create'), ('super_admin', 'user:edit'), ('super_admin', 'user:delete'), ('super_admin', 'role:view'), ('super_admin', 'role:create'), ('super_admin', 'role:edit'), ('super_admin', 'role:delete'), ('super_admin', 'client:view'), ('super_admin', 'client:create'), ('super_admin', 'client:edit'), ('super_admin', 'client:delete'), ('super_admin', 'case:view'), ('super_admin', 'case:create'), ('super_admin', 'case:edit'), ('super_admin', 'case:delete'), ('super_admin', 'case:assign'), ('super_admin', 'finance:view'), ('super_admin', 'finance:create'), ('super_admin', 'finance:edit'), ('super_admin', 'document:view'), ('super_admin', 'document:upload'), ('super_admin', 'document:edit'), ('super_admin', 'document:delete'),
        ('admin', 'dashboard'), ('admin', 'user_management'), ('admin', 'role_management'), ('admin', 'client_management'), ('admin', 'case_management'), ('admin', 'approval_center'), ('admin', 'finance_management'), ('admin', 'document_management'), ('admin', 'tools_center'), ('admin', 'system_settings'), ('admin', 'statistics_reports'),
        ('admin', 'user:view'), ('admin', 'user:create'), ('admin', 'user:edit'), ('admin', 'user:delete'), ('admin', 'role:view'), ('admin', 'role:create'), ('admin', 'role:edit'), ('admin', 'role:delete'), ('admin', 'client:view'), ('admin', 'client:create'), ('admin', 'client:edit'), ('admin', 'client:delete'), ('admin', 'case:view'), ('admin', 'case:create'), ('admin', 'case:edit'), ('admin', 'case:delete'), ('admin', 'case:assign'), ('admin', 'finance:view'), ('admin', 'finance:create'), ('admin', 'finance:edit'), ('admin', 'document:view'), ('admin', 'document:upload'), ('admin', 'document:edit'), ('admin', 'document:delete'),
        ('lawyer', 'dashboard'), ('lawyer', 'client_management'), ('lawyer', 'case_management'), ('lawyer', 'approval_center'), ('lawyer', 'document_management'), ('lawyer', 'tools_center'), ('lawyer', 'statistics_reports'),
        ('lawyer', 'client:view'), ('lawyer', 'client:create'), ('lawyer', 'client:edit'), ('lawyer', 'case:view'), ('lawyer', 'case:create'), ('lawyer', 'case:edit'), ('lawyer', 'document:view'), ('lawyer', 'document:upload'), ('lawyer', 'document:edit'),
        ('assistant', 'dashboard'), ('assistant', 'client_management'), ('assistant', 'case_management'), ('assistant', 'document_management'), ('assistant', 'tools_center'), ('assistant', 'client:view'), ('assistant', 'case:view'), ('assistant', 'document:view'),
        ('finance', 'dashboard'), ('finance', 'finance_management'), ('finance', 'statistics_reports'), ('finance', 'finance:view'), ('finance', 'finance:create'), ('finance', 'finance:edit'),
        ('intern', 'dashboard'), ('intern', 'tools_center'),
        ('user', 'dashboard'), ('user', 'tools_center')
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
