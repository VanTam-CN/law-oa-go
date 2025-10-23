package security

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/casbin/casbin/v2"
	"gorm.io/gorm"
)

// NewPolicyManager 创建策略管理器
func NewPolicyManager(enforcer *casbin.Enforcer, logger *slog.Logger, db *gorm.DB) *PolicyManager {
	return &PolicyManager{
		enforcer: enforcer,
		logger:   logger,
		db:       db,
	}
}

// AddRole 添加角色
func (pm *PolicyManager) AddRole(role *Role) error {
	// 检查角色是否已存在
	var existingRole Role
	if err := pm.db.Where("id = ?", role.ID).First(&existingRole).Error; err == nil {
		return fmt.Errorf("角色已存在: %s", role.ID)
	}

	// 保存角色到数据库
	if err := pm.db.Create(role).Error; err != nil {
		return fmt.Errorf("保存角色失败: %w", err)
	}

	pm.logger.Info("角色添加成功", "role_id", role.ID, "role_name", role.Name)
	return nil
}

// UpdateRole 更新角色
func (pm *PolicyManager) UpdateRole(role *Role) error {
	// 检查角色是否存在
	var existingRole Role
	if err := pm.db.Where("id = ?", role.ID).First(&existingRole).Error; err != nil {
		return fmt.Errorf("角色不存在: %s", role.ID)
	}

	// 系统角色不允许修改
	if existingRole.IsSystem {
		return fmt.Errorf("系统角色不允许修改")
	}

	// 更新数据库
	if err := pm.db.Save(role).Error; err != nil {
		return fmt.Errorf("更新角色失败: %w", err)
	}

	pm.logger.Info("角色更新成功", "role_id", role.ID, "role_name", role.Name)
	return nil
}

// DeleteRole 删除角色
func (pm *PolicyManager) DeleteRole(roleID string) error {
	// 检查角色是否存在
	var role Role
	if err := pm.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return fmt.Errorf("角色不存在: %s", roleID)
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return fmt.Errorf("系统角色不允许删除")
	}

	// 检查是否有用户使用该角色
	var userCount int64
	if err := pm.db.Model(&User{}).Where("roles @> ?", fmt.Sprintf(`["%s"]`, roleID)).Count(&userCount).Error; err != nil {
		return fmt.Errorf("检查角色使用情况失败: %w", err)
	}
	if userCount > 0 {
		return fmt.Errorf("角色正在被使用，无法删除")
	}

	// 开始事务
	tx := pm.db.Begin()

	// 删除相关策略
	if err := tx.Exec("DELETE FROM casbin_rules WHERE v0 = ?", roleID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除角色策略失败: %w", err)
	}

	// 删除角色用户映射
	if err := tx.Exec("DELETE FROM casbin_rules WHERE v1 = ? AND ptype = 'g2'", roleID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除角色用户映射失败: %w", err)
	}

	// 删除角色记录
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除角色记录失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	pm.logger.Info("角色删除成功", "role_id", roleID)
	return nil
}

// GetRole 获取角色
func (pm *PolicyManager) GetRole(roleID string) (*Role, error) {
	var role Role
	if err := pm.db.Where("id = ?", roleID).First(&role).Error; err != nil {
		return nil, fmt.Errorf("角色不存在: %s", roleID)
	}
	return &role, nil
}

// ListRoles 列出所有角色
func (pm *PolicyManager) ListRoles() ([]*Role, error) {
	var roles []*Role
	if err := pm.db.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("查询角色列表失败: %w", err)
	}
	return roles, nil
}

// AssignRoleToUser 为用户分配角色
func (pm *PolicyManager) AssignRoleToUser(userID, roleID string) error {
	// 检查角色是否存在
	_, err := pm.GetRole(roleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 检查用户是否存在
	var user User
	if err := pm.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("用户不存在: %s", userID)
	}

	// 添加用户角色映射到Casbin
	if _, err := pm.enforcer.AddRoleForUser(userID, roleID); err != nil {
		return fmt.Errorf("添加用户角色映射失败: %w", err)
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("保存策略失败: %w", err)
	}

	pm.logger.Info("用户角色分配成功", "user_id", userID, "role_id", roleID)
	return nil
}

// RemoveRoleFromUser 移除用户角色
func (pm *PolicyManager) RemoveRoleFromUser(userID, roleID string) error {
	// 检查用户角色映射是否存在
	hasRole, err := pm.enforcer.HasRoleForUser(userID, roleID)
	if err != nil {
		return fmt.Errorf("检查用户角色失败: %w", err)
	}
	if !hasRole {
		return fmt.Errorf("用户没有该角色")
	}

	// 移除用户角色映射
	if _, err := pm.enforcer.RemoveRoleForUser(userID, roleID); err != nil {
		return fmt.Errorf("移除用户角色映射失败: %w", err)
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("保存策略失败: %w", err)
	}

	pm.logger.Info("用户角色移除成功", "user_id", userID, "role_id", roleID)
	return nil
}

// GetUserRoles 获取用户的所有角色
func (pm *PolicyManager) GetUserRoles(userID string) ([]string, error) {
	roles, err := pm.enforcer.GetRolesForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("获取用户角色失败: %w", err)
	}

	// 获取继承的角色
	implicitRoles, err := pm.enforcer.GetImplicitRolesForUser(userID)
	if err != nil {
		pm.logger.With("error", err, "user_id", userID).Warn("获取继承角色失败")
	} else {
		roles = append(roles, implicitRoles...)
	}

	return roles, nil
}

// AddPermission 添加权限
func (pm *PolicyManager) AddPermission(permission *Permission) error {
	// 检查权限是否已存在
	var existingPermission Permission
	if err := pm.db.Where("id = ?", permission.ID).First(&existingPermission).Error; err == nil {
		return fmt.Errorf("权限已存在: %s", permission.ID)
	}

	// 保存权限到数据库
	if err := pm.db.Create(permission).Error; err != nil {
		return fmt.Errorf("保存权限失败: %w", err)
	}

	pm.logger.Info("权限添加成功", "permission_id", permission.ID, "permission_name", permission.Name)
	return nil
}

// AssignPermissionToRole 为角色分配权限
func (pm *PolicyManager) AssignPermissionToRole(roleID, resource, action string) error {
	// 检查角色是否存在
	_, err := pm.GetRole(roleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 添加策略到Casbin
	if _, err := pm.enforcer.AddPolicy(roleID, resource, action); err != nil {
		return fmt.Errorf("添加策略失败: %w", err)
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("保存策略失败: %w", err)
	}

	pm.logger.Info("角色权限分配成功", "role_id", roleID, "resource", resource, "action", action)
	return nil
}

// RemovePermissionFromRole 移除角色权限
func (pm *PolicyManager) RemovePermissionFromRole(roleID, resource, action string) error {
	// 检查策略是否存在
	hasPolicy, err := pm.enforcer.HasPolicy(roleID, resource, action)
	if err != nil {
		return fmt.Errorf("检查策略失败: %w", err)
	}
	if !hasPolicy {
		return fmt.Errorf("策略不存在")
	}

	// 移除策略
	if _, err := pm.enforcer.RemovePolicy(roleID, resource, action); err != nil {
		return fmt.Errorf("移除策略失败: %w", err)
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		return fmt.Errorf("保存策略失败: %w", err)
	}

	pm.logger.Info("角色权限移除成功", "role_id", roleID, "resource", resource, "action", action)
	return nil
}

// GetRolePermissions 获取角色的所有权限
func (pm *PolicyManager) GetRolePermissions(roleID string) ([][]string, error) {
	policies := pm.enforcer.GetPermissionsForUser(roleID)
	return policies, nil
}

// BatchAssignRoles 批量分配角色
func (pm *PolicyManager) BatchAssignRoles(userID string, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return nil
	}

	// 开始事务
	tx := pm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查所有角色是否存在
	for _, roleID := range roleIDs {
		_, err := pm.GetRole(roleID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("角色不存在: %s", roleID)
		}
	}

	// 批量添加角色映射
	for _, roleID := range roleIDs {
		if _, err := pm.enforcer.AddRoleForUser(userID, roleID); err != nil {
			tx.Rollback()
			return fmt.Errorf("添加角色映射失败: %w", err)
		}
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		tx.Rollback()
		return fmt.Errorf("保存策略失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	pm.logger.Info("批量角色分配成功", "user_id", userID, "role_count", len(roleIDs))
	return nil
}

// BatchRemoveRoles 批量移除角色
func (pm *PolicyManager) BatchRemoveRoles(userID string, roleIDs []string) error {
	if len(roleIDs) == 0 {
		return nil
	}

	// 开始事务
	tx := pm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 批量移除角色映射
	for _, roleID := range roleIDs {
		if _, err := pm.enforcer.RemoveRoleForUser(userID, roleID); err != nil {
			tx.Rollback()
			return fmt.Errorf("移除角色映射失败: %w", err)
		}
	}

	// 保存策略
	if err := pm.enforcer.SavePolicy(); err != nil {
		tx.Rollback()
		return fmt.Errorf("保存策略失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	pm.logger.Info("批量角色移除成功", "user_id", userID, "role_count", len(roleIDs))
	return nil
}

// GetPolicyStats 获取策略统计信息
func (pm *PolicyManager) GetPolicyStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取所有策略
	allPolicies := pm.enforcer.GetPolicy()
	stats["total_policies"] = len(allPolicies)

	// 获取所有角色
	allRoles := pm.enforcer.GetAllRoles()
	stats["total_roles"] = len(allRoles)

	// 获取所有用户角色映射
	allSubjects := pm.enforcer.GetAllSubjects()
	stats["total_users"] = len(allSubjects)

	// 按角色统计用户数
	roleUserCount := make(map[string]int)
	for _, subject := range allSubjects {
		roles, err := pm.enforcer.GetRolesForUser(subject)
		if err != nil {
			continue
		}
		for _, role := range roles {
			roleUserCount[role]++
		}
	}
	stats["role_user_count"] = roleUserCount

	// 按资源统计权限数
	resourcePermissionCount := make(map[string]int)
	for _, policy := range allPolicies {
		if len(policy) >= 3 {
			resource := policy[1]
			resourcePermissionCount[resource]++
		}
	}
	stats["resource_permission_count"] = resourcePermissionCount

	return stats
}

// RefreshPolicy 刷新策略
func (pm *PolicyManager) RefreshPolicy() error {
	if err := pm.enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("刷新策略失败: %w", err)
	}

	pm.logger.Info("策略刷新完成")
	return nil
}

// ExportPolicies 导出策略
func (pm *PolicyManager) ExportPolicies() (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// 导出角色
	roles, err := pm.ListRoles()
	if err != nil {
		return nil, fmt.Errorf("导出角色失败: %w", err)
	}
	data["roles"] = roles

	// 导出策略
	policies := pm.enforcer.GetPolicy()
	data["policies"] = policies

	// 导出角色映射
	groupingPolicies := pm.enforcer.GetGroupingPolicy()
	data["grouping_policies"] = groupingPolicies

	// 添加导出时间
	data["exported_at"] = time.Now().Format(time.RFC3339)
	data["version"] = "1.0"

	return data, nil
}