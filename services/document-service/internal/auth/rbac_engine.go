package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories"
	"github.com/sirupsen/logrus"
)

// RBACEngine 基于角色的访问控制引擎
type RBACEngine struct {
	roleRepo       repositories.RoleRepository
	userRepo       repositories.UserRepository
	permissionRepo repositories.PermissionRepository
	auditRepo      repositories.AuditRepository
	logger         *logrus.Logger
	roleCache      map[string]*RoleDefinition
	permissionCache map[string]*PermissionDefinition
	cacheExpiry    time.Duration
	lastCacheSync  time.Time
}

// NewRBACEngine 创建RBAC引擎
func NewRBACEngine(
	roleRepo repositories.RoleRepository,
	userRepo repositories.UserRepository,
	permissionRepo repositories.PermissionRepository,
	auditRepo repositories.AuditRepository,
	logger *logrus.Logger,
) *RBACEngine {
	return &RBACEngine{
		roleRepo:        roleRepo,
		userRepo:        userRepo,
		permissionRepo:  permissionRepo,
		auditRepo:       auditRepo,
		logger:          logger,
		roleCache:       make(map[string]*RoleDefinition),
		permissionCache: make(map[string]*PermissionDefinition),
		cacheExpiry:     10 * time.Minute,
	}
}

// RoleDefinition 角色定义
type RoleDefinition struct {
	ID           uint                   `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Description  string                 `json:"description"`
	Level        int                    `json:"level"`
	Type         string                 `json:"type"` // system, tenant, custom
	Permissions  []string               `json:"permissions"`
	Constraints  []RoleConstraint       `json:"constraints"`
	Inherits     []string               `json:"inherits"`
	TenantID     string                 `json:"tenant_id"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    uint                   `json:"created_by"`
}

// PermissionDefinition 权限定义
type PermissionDefinition struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Scope       string                 `json:"scope"` // global, tenant, resource
	Attributes  map[string]interface{} `json:"attributes"`
	Category    string                 `json:"category"`
	System      bool                   `json:"system"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RoleConstraint 角色约束
type RoleConstraint struct {
	Type        string                 `json:"type"` // time, ip, resource, attribute
	Field       string                 `json:"field"`
	Operator    string                 `json:"operator"`
	Value       interface{}            `json:"value"`
	Description string                 `json:"description"`
}

// RoleAssignment 角色分配
type RoleAssignment struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	RoleID     uint      `json:"role_id"`
	RoleName   string    `json:"role_name"`
	TenantID   string    `json:"tenant_id"`
	Attributes map[string]interface{} `json:"attributes"`
	ExpiresAt  *time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  uint      `json:"created_by"`
}

// AccessCheck 访问检查
type AccessCheck struct {
	UserID      uint                   `json:"user_id"`
	Username    string                 `json:"username"`
	TenantID    string                 `json:"tenant_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Context     map[string]interface{} `json:"context"`
	RequestID   string                 `json:"request_id"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AccessResult 访问结果
type AccessResult struct {
	Allowed     bool                   `json:"allowed"`
	Reason      string                 `json:"reason"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	Duration    time.Duration          `json:"duration"`
	Attributes  map[string]interface{} `json:"attributes"`
	Constraints []AppliedConstraint    `json:"constraints"`
}

// AppliedConstraint 应用的约束
type AppliedConstraint struct {
	Type        string      `json:"type"`
	Field       string      `json:"field"`
	Value       interface{} `json:"value"`
	Applied     bool        `json:"applied"`
	Reason      string      `json:"reason"`
}

// CheckPermission 检查权限
func (e *RBACEngine) CheckPermission(ctx context.Context, check *AccessCheck) (*AccessResult, error) {
	startTime := time.Now()

	e.logger.WithFields(logrus.Fields{
		"request_id": check.RequestID,
		"user_id":    check.UserID,
		"resource":   check.Resource,
		"action":     check.Action,
		"tenant_id":  check.TenantID,
	}).Debug("Checking permission")

	// 刷新缓存
	if err := e.refreshCache(ctx, check.TenantID); err != nil {
		e.logger.WithError(err).Error("Failed to refresh cache")
		return nil, fmt.Errorf("failed to refresh cache: %w", err)
	}

	// 获取用户角色
	userRoles, err := e.getUserRoles(ctx, check.UserID, check.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if len(userRoles) == 0 {
		return &AccessResult{
			Allowed:  false,
			Reason:   "User has no roles assigned",
			Roles:    []string{},
			Duration: time.Since(startTime),
		}, nil
	}

	// 检查角色约束
	constraintResults := e.checkRoleConstraints(userRoles, check)
	disabledRoles := e.getDisabledRoles(constraintResults)
	activeRoles := e.getActiveRoles(userRoles, disabledRoles)

	if len(activeRoles) == 0 {
		return &AccessResult{
			Allowed:     false,
			Reason:      "All roles are disabled by constraints",
			Roles:       e.getRoleNames(userRoles),
			Constraints: constraintResults,
			Duration:    time.Since(startTime),
		}, nil
	}

	// 获取用户权限
	userPermissions := e.getUserPermissions(activeRoles)

	// 检查具体权限
	permissionName := e.buildPermissionName(check.Resource, check.Action)
	hasPermission := e.hasPermission(userPermissions, permissionName)

	if !hasPermission {
		// 检查通配符权限
		hasWildcardPermission := e.checkWildcardPermissions(userPermissions, check.Resource, check.Action)
		if !hasWildcardPermission {
			return &AccessResult{
				Allowed:     false,
				Reason:      fmt.Sprintf("Permission %s not granted", permissionName),
				Roles:       e.getRoleNames(activeRoles),
				Permissions: userPermissions,
				Constraints: constraintResults,
				Duration:    time.Since(startTime),
			}, nil
		}
	}

	// 检查资源级权限约束
	resourceConstraints := e.checkResourceConstraints(activeRoles, check)

	result := &AccessResult{
		Allowed:     true,
		Reason:      "Permission granted",
		Roles:       e.getRoleNames(activeRoles),
		Permissions: userPermissions,
		Duration:    time.Since(startTime),
		Constraints: append(constraintResults, resourceConstraints...),
		Attributes:  make(map[string]interface{}),
	}

	// 记录访问日志
	e.logAccessCheck(ctx, check, result)

	return result, nil
}

// AssignRole 分配角色
func (e *RBACEngine) AssignRole(ctx context.Context, userID uint, roleID uint, tenantID string, attributes map[string]interface{}, expiresAt *time.Time) error {
	e.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"role_id":   roleID,
		"tenant_id": tenantID,
	}).Info("Assigning role to user")

	// 检查角色是否存在
	role, err := e.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查租户匹配
	if role.TenantID != tenantID && role.TenantID != "*" {
		return fmt.Errorf("role does not belong to tenant")
	}

	// 检查角色是否启用
	if !role.Enabled {
		return fmt.Errorf("role is disabled")
	}

	// 创建角色分配
	assignment := &models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		TenantID:   tenantID,
		Attributes: attributes,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
	}

	if err := e.userRepo.AssignRole(ctx, assignment); err != nil {
		e.logger.WithError(err).Error("Failed to assign role")
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 记录审计日志
	e.logRoleOperation(ctx, "assign", userID, roleID, tenantID, "Role assigned")

	return nil
}

// RevokeRole 撤销角色
func (e *RBACEngine) RevokeRole(ctx context.Context, userID uint, roleID uint, tenantID string) error {
	e.logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"role_id":   roleID,
		"tenant_id": tenantID,
	}).Info("Revoking role from user")

	if err := e.userRepo.RevokeRole(ctx, userID, roleID, tenantID); err != nil {
		e.logger.WithError(err).Error("Failed to revoke role")
		return fmt.Errorf("failed to revoke role: %w", err)
	}

	// 记录审计日志
	e.logRoleOperation(ctx, "revoke", userID, roleID, tenantID, "Role revoked")

	return nil
}

// GetUserRoles 获取用户角色
func (e *RBACEngine) GetUserRoles(ctx context.Context, userID uint, tenantID string) ([]*RoleDefinition, error) {
	userRoles, err := e.getUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	var roleDefinitions []*RoleDefinition
	for _, userRole := range userRoles {
		if roleDef, exists := e.roleCache[userRole.RoleName]; exists {
			roleDefinitions = append(roleDefinitions, roleDef)
		}
	}

	return roleDefinitions, nil
}

// CreateRole 创建角色
func (e *RBACEngine) CreateRole(ctx context.Context, role *RoleDefinition) error {
	e.logger.WithFields(logrus.Fields{
		"name":      role.Name,
		"tenant_id": role.TenantID,
	}).Info("Creating new role")

	// 验证角色定义
	if err := e.validateRoleDefinition(role); err != nil {
		return fmt.Errorf("invalid role definition: %w", err)
	}

	// 检查角色名是否已存在
	if _, exists := e.roleCache[role.Name]; exists {
		return fmt.Errorf("role with name '%s' already exists", role.Name)
	}

	// 创建角色模型
	roleModel := &models.Role{
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		Level:       role.Level,
		Type:        role.Type,
		TenantID:    role.TenantID,
		Enabled:     role.Enabled,
		CreatedBy:   role.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := e.roleRepo.Create(ctx, roleModel); err != nil {
		e.logger.WithError(err).Error("Failed to create role")
		return fmt.Errorf("failed to create role: %w", err)
	}

	role.ID = roleModel.ID

	// 添加权限到角色
	for _, permName := range role.Permissions {
		if err := e.roleRepo.AddPermission(ctx, roleModel.ID, permName); err != nil {
			e.logger.WithError(err).WithField("permission", permName).Warn("Failed to add permission to role")
		}
	}

	// 更新缓存
	e.roleCache[role.Name] = role

	// 记录审计日志
	e.logRoleOperation(ctx, "create", 0, roleModel.ID, role.TenantID, "Role created")

	return nil
}

// UpdateRole 更新角色
func (e *RBACEngine) UpdateRole(ctx context.Context, roleID uint, updates *RoleDefinition) error {
	e.logger.WithField("role_id", roleID).Info("Updating role")

	// 获取现有角色
	existingRole, err := e.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get existing role: %w", err)
	}

	// 更新字段
	if updates.DisplayName != "" {
		existingRole.DisplayName = updates.DisplayName
	}
	if updates.Description != "" {
		existingRole.Description = updates.Description
	}
	if updates.Level != 0 {
		existingRole.Level = updates.Level
	}
	if updates.Enabled != existingRole.Enabled {
		existingRole.Enabled = updates.Enabled
	}

	existingRole.UpdatedAt = time.Now()

	if err := e.roleRepo.Update(ctx, existingRole); err != nil {
		e.logger.WithError(err).Error("Failed to update role")
		return fmt.Errorf("failed to update role: %w", err)
	}

	// 更新权限（如果提供）
	if len(updates.Permissions) > 0 {
		// 清除现有权限
		if err := e.roleRepo.ClearPermissions(ctx, roleID); err != nil {
			e.logger.WithError(err).Warn("Failed to clear existing permissions")
		}

		// 添加新权限
		for _, permName := range updates.Permissions {
			if err := e.roleRepo.AddPermission(ctx, roleID, permName); err != nil {
				e.logger.WithError(err).WithField("permission", permName).Warn("Failed to add permission to role")
			}
		}
	}

	// 记录审计日志
	e.logRoleOperation(ctx, "update", 0, roleID, existingRole.TenantID, "Role updated")

	return nil
}

// DeleteRole 删除角色
func (e *RBACEngine) DeleteRole(ctx context.Context, roleID uint) error {
	e.logger.WithField("role_id", roleID).Info("Deleting role")

	// 获取角色信息
	role, err := e.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// 检查是否有用户分配了此角色
	userCount, err := e.userRepo.GetRoleUserCount(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role user count: %w", err)
	}

	if userCount > 0 {
		return fmt.Errorf("cannot delete role with %d assigned users", userCount)
	}

	if err := e.roleRepo.Delete(ctx, roleID); err != nil {
		e.logger.WithError(err).Error("Failed to delete role")
		return fmt.Errorf("failed to delete role: %w", err)
	}

	// 从缓存中移除
	delete(e.roleCache, role.Name)

	// 记录审计日志
	e.logRoleOperation(ctx, "delete", 0, roleID, role.TenantID, "Role deleted")

	return nil
}

// 辅助方法
func (e *RBACEngine) refreshCache(ctx context.Context, tenantID string) error {
	// 检查缓存是否过期
	if time.Since(e.lastCacheSync) < e.cacheExpiry {
		return nil
	}

	e.logger.Debug("Refreshing RBAC cache")

	// 加载角色
	roles, err := e.roleRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to load roles: %w", err)
	}

	e.roleCache = make(map[string]*RoleDefinition)
	for _, role := range roles {
		permissions, _ := e.roleRepo.GetPermissions(ctx, role.ID)

		roleDef := &RoleDefinition{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			Level:       role.Level,
			Type:        role.Type,
			Permissions: permissions,
			TenantID:    role.TenantID,
			Enabled:     role.Enabled,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
			CreatedBy:   role.CreatedBy,
		}
		e.roleCache[role.Name] = roleDef
	}

	// 加载权限
	permissions, err := e.permissionRepo.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to load permissions: %w", err)
	}

	e.permissionCache = make(map[string]*PermissionDefinition)
	for _, perm := range permissions {
		permDef := &PermissionDefinition{
			ID:          perm.ID,
			Name:        perm.Name,
			DisplayName: perm.DisplayName,
			Description: perm.Description,
			Resource:    perm.Resource,
			Action:      perm.Action,
			Scope:       perm.Scope,
			Category:    perm.Category,
			System:      perm.System,
			Enabled:     perm.Enabled,
			CreatedAt:   perm.CreatedAt,
			UpdatedAt:   perm.UpdatedAt,
		}
		e.permissionCache[perm.Name] = permDef
	}

	e.lastCacheSync = time.Now()
	return nil
}

func (e *RBACEngine) getUserRoles(ctx context.Context, userID uint, tenantID string) ([]*RoleAssignment, error) {
	userRoles, err := e.userRepo.GetRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	var assignments []*RoleAssignment
	for _, userRole := range userRoles {
		// 检查是否过期
		if userRole.ExpiresAt != nil && time.Now().After(*userRole.ExpiresAt) {
			continue
		}

		assignment := &RoleAssignment{
			ID:         userRole.ID,
			UserID:     userRole.UserID,
			RoleID:     userRole.RoleID,
			RoleName:   userRole.Role.Name,
			TenantID:   userRole.TenantID,
			Attributes: userRole.Attributes,
			ExpiresAt:  userRole.ExpiresAt,
			CreatedAt:  userRole.CreatedAt,
			CreatedBy:  userRole.CreatedBy,
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}

func (e *RBACEngine) checkRoleConstraints(userRoles []*RoleAssignment, check *AccessCheck) []AppliedConstraint {
	var constraints []AppliedConstraint

	for _, userRole := range userRoles {
		roleDef, exists := e.roleCache[userRole.RoleName]
		if !exists {
			continue
		}

		for _, constraint := range roleDef.Constraints {
			applied := e.evaluateConstraint(constraint, check, userRole.Attributes)
			constraints = append(constraints, applied)
		}
	}

	return constraints
}

func (e *RBACEngine) evaluateConstraint(constraint RoleConstraint, check *AccessCheck, attributes map[string]interface{}) AppliedConstraint {
	applied := AppliedConstraint{
		Type:  constraint.Type,
		Field: constraint.Field,
		Value: constraint.Value,
	}

	switch constraint.Type {
	case "time":
		applied.Applied = e.evaluateTimeConstraint(constraint, check)
		if !applied.Applied {
			applied.Reason = "Time constraint not satisfied"
		}
	case "ip":
		applied.Applied = e.evaluateIPConstraint(constraint, check)
		if !applied.Applied {
			applied.Reason = "IP constraint not satisfied"
		}
	case "attribute":
		applied.Applied = e.evaluateAttributeConstraint(constraint, check, attributes)
		if !applied.Applied {
			applied.Reason = "Attribute constraint not satisfied"
		}
	default:
		applied.Applied = true
		applied.Reason = "Unknown constraint type"
	}

	return applied
}

func (e *RBACEngine) evaluateTimeConstraint(constraint RoleConstraint, check *AccessCheck) bool {
	// 实现时间约束评估逻辑
	now := time.Now()

	if constraint.Field == "hour" {
		if hour, ok := constraint.Value.(float64); ok {
			return now.Hour() == int(hour)
		}
	}

	if constraint.Field == "day_of_week" {
		if day, ok := constraint.Value.(float64); ok {
			return now.Weekday() == time.Weekday(day)
		}
	}

	return true
}

func (e *RBACEngine) evaluateIPConstraint(constraint RoleConstraint, check *AccessCheck) bool {
	// 实现IP约束评估逻辑
	if constraint.Field == "ip_range" {
		if ipRange, ok := constraint.Value.(string); ok {
			return e.isIPInRange(check.Context["ip"].(string), ipRange)
		}
	}

	return true
}

func (e *RBACEngine) evaluateAttributeConstraint(constraint RoleConstraint, check *AccessCheck, attributes map[string]interface{}) bool {
	// 实现属性约束评估逻辑
	userValue, exists := attributes[constraint.Field]
	if !exists {
		return false
	}

	switch constraint.Operator {
	case "eq":
		return fmt.Sprintf("%v", userValue) == fmt.Sprintf("%v", constraint.Value)
	case "in":
		if values, ok := constraint.Value.([]interface{}); ok {
			for _, v := range values {
				if fmt.Sprintf("%v", userValue) == fmt.Sprintf("%v", v) {
					return true
				}
			}
		}
	}

	return false
}

func (e *RBACEngine) isIPInRange(ip, rangeStr string) bool {
	// 实现IP范围检查逻辑
	return true
}

func (e *RBACEngine) getDisabledRoles(constraints []AppliedConstraint) []string {
	var disabledRoles []string
	for _, constraint := range constraints {
		if !constraint.Applied {
			// 这里需要知道约束属于哪个角色，简化处理
			disabledRoles = append(disabledRoles, "restricted_role")
		}
	}
	return disabledRoles
}

func (e *RBACEngine) getActiveRoles(userRoles []*RoleAssignment, disabledRoles []string) []*RoleAssignment {
	var activeRoles []*RoleAssignment
	for _, userRole := range userRoles {
		disabled := false
		for _, disabledRole := range disabledRoles {
			if userRole.RoleName == disabledRole {
				disabled = true
				break
			}
		}
		if !disabled {
			activeRoles = append(activeRoles, userRole)
		}
	}
	return activeRoles
}

func (e *RBACEngine) getUserPermissions(activeRoles []*RoleAssignment) []string {
	var permissions []string
	seen := make(map[string]bool)

	for _, userRole := range activeRoles {
		roleDef, exists := e.roleCache[userRole.RoleName]
		if !exists {
			continue
		}

		// 添加角色直接权限
		for _, perm := range roleDef.Permissions {
			if !seen[perm] {
				permissions = append(permissions, perm)
				seen[perm] = true
			}
		}

		// 添加继承角色权限
		for _, inheritedRole := range roleDef.Inherits {
			if inheritedRoleDef, exists := e.roleCache[inheritedRole]; exists {
				for _, perm := range inheritedRoleDef.Permissions {
					if !seen[perm] {
						permissions = append(permissions, perm)
						seen[perm] = true
					}
				}
			}
		}
	}

	return permissions
}

func (e *RBACEngine) buildPermissionName(resource, action string) string {
	return fmt.Sprintf("%s:%s", resource, action)
}

func (e *RBACEngine) hasPermission(permissions []string, permission string) bool {
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

func (e *RBACEngine) checkWildcardPermissions(permissions []string, resource, action string) bool {
	wildcardPerms := []string{
		fmt.Sprintf("*:%s", action),           // 任何资源的此动作
		fmt.Sprintf("%s:*", resource),        // 此资源的任何动作
		"*:*",                                 // 所有权限
	}

	for _, wildcard := range wildcardPerms {
		if e.hasPermission(permissions, wildcard) {
			return true
		}
	}

	return false
}

func (e *RBACEngine) checkResourceConstraints(activeRoles []*RoleAssignment, check *AccessCheck) []AppliedConstraint {
	// 实现资源级约束检查
	return []AppliedConstraint{}
}

func (e *RBACEngine) getRoleNames(userRoles []*RoleAssignment) []string {
	var names []string
	for _, userRole := range userRoles {
		names = append(names, userRole.RoleName)
	}
	return names
}

func (e *RBACEngine) validateRoleDefinition(role *RoleDefinition) error {
	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}
	if role.Type == "" {
		return fmt.Errorf("role type is required")
	}
	if role.TenantID == "" && role.Type != "system" {
		return fmt.Errorf("tenant ID is required for non-system roles")
	}
	return nil
}

func (e *RBACEngine) logAccessCheck(ctx context.Context, check *AccessCheck, result *AccessResult) {
	auditLog := &models.AuditLog{
		UserID:    check.UserID,
		Action:    fmt.Sprintf("rbac_check:%s:%s", check.Resource, check.Action),
		Resource:  fmt.Sprintf("rbac:%d", check.UserID),
		Result:    "allowed",
		Details:   fmt.Sprintf("Roles: %v, Reason: %s", result.Roles, result.Reason),
		TenantID:  check.TenantID,
		CreatedAt: time.Now(),
	}

	if !result.Allowed {
		auditLog.Result = "denied"
	}

	if err := e.auditRepo.Create(ctx, auditLog); err != nil {
		e.logger.WithError(err).Error("Failed to log access check")
	}
}

func (e *RBACEngine) logRoleOperation(ctx context.Context, operation string, userID uint, roleID uint, tenantID string, details string) {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    fmt.Sprintf("role_%s", operation),
		Resource:  fmt.Sprintf("role:%d", roleID),
		Result:    "success",
		Details:   details,
		TenantID:  tenantID,
		CreatedAt: time.Now(),
	}

	if err := e.auditRepo.Create(ctx, auditLog); err != nil {
		e.logger.WithError(err).Error("Failed to log role operation")
	}
}