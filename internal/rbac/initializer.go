package rbac

import (
	"context"
	"fmt"
	"log"
	"law-oa-go/internal/models"
)

// DefaultPermissions 默认权限定义
var DefaultPermissions = []Permission{
	// 用户管理权限
	{Name: "查看用户", Key: "user:view", Description: "查看用户列表和详情", Resource: "/api/users", Action: "GET"},
	{Name: "创建用户", Key: "user:create", Description: "创建新用户", Resource: "/api/users", Action: "POST"},
	{Name: "更新用户", Key: "user:update", Description: "更新用户信息", Resource: "/api/users", Action: "PUT"},
	{Name: "删除用户", Key: "user:delete", Description: "删除用户", Resource: "/api/users", Action: "DELETE"},
	{Name: "分配角色", Key: "user:assign_role", Description: "为用户分配角色", Resource: "/api/users/role", Action: "POST"},
	
	// 角色管理权限
	{Name: "查看角色", Key: "role:view", Description: "查看角色列表和详情", Resource: "/api/roles", Action: "GET"},
	{Name: "创建角色", Key: "role:create", Description: "创建新角色", Resource: "/api/roles", Action: "POST"},
	{Name: "更新角色", Key: "role:update", Description: "更新角色信息", Resource: "/api/roles", Action: "PUT"},
	{Name: "删除角色", Key: "role:delete", Description: "删除角色", Resource: "/api/roles", Action: "DELETE"},
	{Name: "分配权限", Key: "role:assign_permission", Description: "为角色分配权限", Resource: "/api/roles/permissions", Action: "POST"},
	
	// 权限管理权限
	{Name: "查看权限", Key: "permission:view", Description: "查看权限列表", Resource: "/api/permissions", Action: "GET"},
	{Name: "创建权限", Key: "permission:create", Description: "创建新权限", Resource: "/api/permissions", Action: "POST"},
	{Name: "更新权限", Key: "permission:update", Description: "更新权限信息", Resource: "/api/permissions", Action: "PUT"},
	{Name: "删除权限", Key: "permission:delete", Description: "删除权限", Resource: "/api/permissions", Action: "DELETE"},
	
	// 案件管理权限
	{Name: "查看案件", Key: "case:view", Description: "查看案件列表和详情", Resource: "/api/cases", Action: "GET"},
	{Name: "创建案件", Key: "case:create", Description: "创建新案件", Resource: "/api/cases", Action: "POST"},
	{Name: "更新案件", Key: "case:update", Description: "更新案件信息", Resource: "/api/cases", Action: "PUT"},
	{Name: "删除案件", Key: "case:delete", Description: "删除案件", Resource: "/api/cases", Action: "DELETE"},
	{Name: "分配案件", Key: "case:assign", Description: "分配案件给律师", Resource: "/api/cases/assign", Action: "POST"},
	
	// 客户管理权限
	{Name: "查看客户", Key: "client:view", Description: "查看客户列表和详情", Resource: "/api/clients", Action: "GET"},
	{Name: "创建客户", Key: "client:create", Description: "创建新客户", Resource: "/api/clients", Action: "POST"},
	{Name: "更新客户", Key: "client:update", Description: "更新客户信息", Resource: "/api/clients", Action: "PUT"},
	{Name: "删除客户", Key: "client:delete", Description: "删除客户", Resource: "/api/clients", Action: "DELETE"},
	
	// 文档管理权限
	{Name: "查看文档", Key: "document:view", Description: "查看文档列表和内容", Resource: "/api/documents", Action: "GET"},
	{Name: "上传文档", Key: "document:upload", Description: "上传新文档", Resource: "/api/documents", Action: "POST"},
	{Name: "更新文档", Key: "document:update", Description: "更新文档信息", Resource: "/api/documents", Action: "PUT"},
	{Name: "删除文档", Key: "document:delete", Description: "删除文档", Resource: "/api/documents", Action: "DELETE"},
	{Name: "下载文档", Key: "document:download", Description: "下载文档", Resource: "/api/documents/download", Action: "GET"},
	{Name: "分享文档", Key: "document:share", Description: "分享文档", Resource: "/api/documents/share", Action: "POST"},
	
	// 系统管理权限
	{Name: "系统配置", Key: "system:config", Description: "系统配置管理", Resource: "/api/system/config", Action: "POST"},
	{Name: "查看日志", Key: "system:logs", Description: "查看系统日志", Resource: "/api/system/logs", Action: "GET"},
	{Name: "数据备份", Key: "system:backup", Description: "数据备份和恢复", Resource: "/api/system/backup", Action: "POST"},
	{Name: "系统监控", Key: "system:monitor", Description: "系统监控", Resource: "/api/system/monitor", Action: "GET"},
	
	// 财务管理权限
	{Name: "查看账单", Key: "finance:view", Description: "查看账单和财务信息", Resource: "/api/finance", Action: "GET"},
	{Name: "创建账单", Key: "finance:create", Description: "创建账单", Resource: "/api/finance", Action: "POST"},
	{Name: "更新账单", Key: "finance:update", Description: "更新账单信息", Resource: "/api/finance", Action: "PUT"},
	{Name: "删除账单", Key: "finance:delete", Description: "删除账单", Resource: "/api/finance", Action: "DELETE"},
	
	// 统计报表权限
	{Name: "查看报表", Key: "report:view", Description: "查看统计报表", Resource: "/api/reports", Action: "GET"},
	{Name: "导出报表", Key: "report:export", Description: "导出报表数据", Resource: "/api/reports/export", Action: "POST"},
	
	// 利益冲突检查权限
	{Name: "冲突检查", Key: "conflict:check", Description: "利益冲突检查", Resource: "/api/conflict", Action: "POST"},
	{Name: "查看冲突", Key: "conflict:view", Description: "查看冲突检查记录", Resource: "/api/conflict", Action: "GET"},
	
	// 个人权限
	{Name: "个人资料", Key: "profile:view", Description: "查看个人资料", Resource: "/api/profile", Action: "GET"},
	{Name: "修改密码", Key: "profile:password", Description: "修改个人密码", Resource: "/api/profile/password", Action: "POST"},
	{Name: "个人设置", Key: "profile:settings", Description: "个人设置", Resource: "/api/profile/settings", Action: "POST"},
}

// DefaultRoles 默认角色定义
var DefaultRoles = []struct {
	Name           string
	Key            string
	Description    string
	PermissionKeys []string
}{
	{
		Name:           "超级管理员",
		Key:            "super_admin",
		Description:    "系统超级管理员，拥有所有权限",
		PermissionKeys: []string{"*"}, // 所有权限
	},
	{
		Name:           "系统管理员",
		Key:            "admin",
		Description:    "系统管理员，拥有大部分权限",
		PermissionKeys: []string{
			"user:view", "user:create", "user:update", "user:delete", "user:assign_role",
			"role:view", "role:create", "role:update", "role:delete", "role:assign_permission",
			"permission:view", "permission:create", "permission:update", "permission:delete",
			"case:view", "case:create", "case:update", "case:delete", "case:assign",
			"client:view", "client:create", "client:update", "client:delete",
			"document:view", "document:upload", "document:update", "document:delete", "document:download",
			"system:config", "system:logs", "system:backup", "system:monitor",
			"finance:view", "finance:create", "finance:update", "finance:delete",
			"report:view", "report:export",
			"conflict:check", "conflict:view",
		},
	},
	{
		Name:           "律师",
		Key:            "lawyer",
		Description:    "律师，处理案件相关事务",
		PermissionKeys: []string{
			"case:view", "case:update",
			"client:view", "client:create", "client:update",
			"document:view", "document:upload", "document:update", "document:download",
			"finance:view",
			"report:view",
			"conflict:check", "conflict:view",
			"profile:view", "profile:password", "profile:settings",
		},
	},
	{
		Name:           "律师助理",
		Key:            "assistant",
		Description:    "律师助理，协助律师处理案件",
		PermissionKeys: []string{
			"case:view",
			"client:view", "client:create", "client:update",
			"document:view", "document:upload", "document:update", "document:download",
			"report:view",
			"conflict:check", "conflict:view",
			"profile:view", "profile:password", "profile:settings",
		},
	},
	{
		Name:           "财务人员",
		Key:            "finance",
		Description:    "财务人员，处理财务相关事务",
		PermissionKeys: []string{
			"case:view",
			"client:view",
			"document:view", "document:download",
			"finance:view", "finance:create", "finance:update",
			"report:view", "report:export",
			"profile:view", "profile:password", "profile:settings",
		},
	},
	{
		Name:           "客户经理",
		Key:            "manager",
		Description:    "客户经理，负责客户关系管理",
		PermissionKeys: []string{
			"case:view",
			"client:view", "client:create", "client:update",
			"document:view", "document:download",
			"finance:view",
			"report:view",
			"profile:view", "profile:password", "profile:settings",
		},
	},
	{
		Name:           "访客",
		Key:            "guest",
		Description:    "访客，只有基本查看权限",
		PermissionKeys: []string{
			"profile:view", "profile:password", "profile:settings",
		},
	},
}

// PermissionInitializer 权限初始化器
type PermissionInitializer struct {
	rbacService *RBACService
}

// NewPermissionInitializer 创建权限初始化器
func NewPermissionInitializer(rbacService *RBACService) *PermissionInitializer {
	return &PermissionInitializer{
		rbacService: rbacService,
	}
}

// Initialize 初始化默认权限和角色
func (pi *PermissionInitializer) Initialize(ctx context.Context) error {
	log.Println("开始初始化权限系统...")
	
	// 创建默认权限
	if err := pi.createDefaultPermissions(ctx); err != nil {
		return fmt.Errorf("创建默认权限失败: %w", err)
	}
	
	// 创建默认角色
	if err := pi.createDefaultRoles(ctx); err != nil {
		return fmt.Errorf("创建默认角色失败: %w", err)
	}
	
	log.Println("权限系统初始化完成")
	return nil
}

// createDefaultPermissions 创建默认权限
func (pi *PermissionInitializer) createDefaultPermissions(ctx context.Context) error {
	for _, perm := range DefaultPermissions {
		// 检查权限是否已存在
		var existingPerm models.Permission
		if err := pi.rbacService.db.Where("permission_key = ?", perm.Key).First(&existingPerm).Error; err == nil {
			// 权限已存在，跳过
			continue
		}
		
		// 创建权限
		if err := pi.rbacService.CreatePermission(ctx, &perm); err != nil {
			return fmt.Errorf("创建权限 %s 失败: %w", perm.Key, err)
		}
		
		log.Printf("创建权限: %s (%s)", perm.Name, perm.Key)
	}
	
	return nil
}

// createDefaultRoles 创建默认角色
func (pi *PermissionInitializer) createDefaultRoles(ctx context.Context) error {
	for _, roleDef := range DefaultRoles {
		// 检查角色是否已存在
		var existingRole models.Role
		if err := pi.rbacService.db.Where("role_key = ?", roleDef.Key).First(&existingRole).Error; err == nil {
			// 角色已存在，跳过
			continue
		}
		
		// 创建角色
		role := &Role{
			Name:        roleDef.Name,
			Key:         roleDef.Key,
			Description: roleDef.Description,
		}
		
		if err := pi.rbacService.CreateRole(ctx, role, []uint{}); err != nil {
			return fmt.Errorf("创建角色 %s 失败: %w", roleDef.Key, err)
		}
		
		// 如果不是超级管理员（特殊处理），需要分配权限
		if roleDef.Key != "super_admin" {
			// 获取权限ID
			var permissionIDs []uint
			for _, permKey := range roleDef.PermissionKeys {
				if permKey == "*" {
					// 超级管理员特殊处理
					continue
				}
				
				var perm models.Permission
				if err := pi.rbacService.db.Where("permission_key = ?", permKey).First(&perm).Error; err != nil {
					log.Printf("警告: 权限 %s 不存在", permKey)
					continue
				}
				permissionIDs = append(permissionIDs, perm.ID)
			}
			
			// 分配权限
			if len(permissionIDs) > 0 {
				if err := pi.rbacService.AssignPermissionsToRole(ctx, role.ID, permissionIDs); err != nil {
					return fmt.Errorf("为角色 %s 分配权限失败: %w", roleDef.Key, err)
				}
			}
		}
		
		log.Printf("创建角色: %s (%s)", roleDef.Name, roleDef.Key)
	}
	
	return nil
}

// ResetPermissions 重置权限系统
func (pi *PermissionInitializer) ResetPermissions(ctx context.Context) error {
	log.Println("开始重置权限系统...")
	
	// 清除现有权限和角色
	if err := pi.rbacService.db.Exec("DELETE FROM role_permissions").Error; err != nil {
		return fmt.Errorf("清除角色权限关联失败: %w", err)
	}
	
	if err := pi.rbacService.db.Exec("DELETE FROM user_roles").Error; err != nil {
		return fmt.Errorf("清除用户角色关联失败: %w", err)
	}
	
	if err := pi.rbacService.db.Exec("DELETE FROM permissions").Error; err != nil {
		return fmt.Errorf("清除权限失败: %w", err)
	}
	
	if err := pi.rbacService.db.Exec("DELETE FROM roles").Error; err != nil {
		return fmt.Errorf("清除角色失败: %w", err)
	}
	
	// 清除缓存
	pi.rbacService.clearPermissionCache(ctx)
	
	// 重新创建默认权限和角色
	if err := pi.Initialize(ctx); err != nil {
		return fmt.Errorf("重新初始化权限系统失败: %w", err)
	}
	
	log.Println("权限系统重置完成")
	return nil
}

// GetPermissionKeyByResourceAction 根据资源和动作获取权限键
func GetPermissionKeyByResourceAction(resource, action string) string {
	// 从资源路径提取资源类型
	resourceType := extractResourceType(resource)
	
	// 特殊处理
	if resourceType == "profile" {
		return fmt.Sprintf("profile:%s", action)
	}
	
	return fmt.Sprintf("%s:%s", resourceType, action)
}

// extractResourceType 从资源路径提取资源类型
func extractResourceType(resource string) string {
	// 移除前缀和后缀
	resource = resource[4:] // 移除 "/api"
	
	// 按 / 分割
	parts := strings.Split(resource, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	
	return "unknown"
}

// ValidatePermissionKey 验证权限键格式
func ValidatePermissionKey(key string) bool {
	// 格式: resource:action
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return false
	}
	
	// 检查资源名和动作名是否有效
	resource := parts[0]
	action := parts[1]
	
	if resource == "" || action == "" {
		return false
	}
	
	// 只允许字母、数字和下划线
	for _, char := range key {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' || char == ':') {
			return false
		}
	}
	
	return true
}

// GetPermissionKeyMapping 获取权限键映射表
func GetPermissionKeyMapping() map[string]string {
	mapping := make(map[string]string)
	
	for _, perm := range DefaultPermissions {
		mapping[perm.Key] = perm.Description
	}
	
	return mapping
}

// GetRolePermissionMapping 获取角色权限映射表
func GetRolePermissionMapping() map[string][]string {
	mapping := make(map[string][]string)
	
	for _, roleDef := range DefaultRoles {
		mapping[roleDef.Key] = roleDef.PermissionKeys
	}
	
	return mapping
}