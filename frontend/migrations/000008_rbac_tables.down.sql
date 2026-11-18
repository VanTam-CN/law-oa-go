-- rbac_tables down migration

-- 删除用户角色关联表
DROP TABLE IF EXISTS `user_roles`;

-- 删除角色权限关联表
DROP TABLE IF EXISTS `role_permissions`;

-- 删除权限表
DROP TABLE IF EXISTS `permissions`;

-- 删除角色表
DROP TABLE IF EXISTS `roles`;
