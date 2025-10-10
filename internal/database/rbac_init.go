package database

import (
	"context"
	"fmt"
	"log"

	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// InitRBACData 初始化角色权限数据
func InitRBACData(db *gorm.DB) error {
	log.Println("正在初始化角色权限数据...")

	// 创建事务
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 检查是否已有数据
	var roleCount int64
	if err := tx.Model(&models.Role{}).Count(&roleCount).Error; err != nil {
		return fmt.Errorf("检查角色数量失败: %w", err)
	}

	if roleCount > 0 {
		log.Println("角色数据已存在，跳过初始化")
		return nil
	}

	// 创建RBAC服务
	rbacService := services.NewRBACService(db)

	// 创建系统角色
	roles := []*services.CreateRoleRequest{
		{Name: "超级管理员", Code: "super_admin", Description: "系统超级管理员，拥有所有权限", Status: "active", SortOrder: 1},
		{Name: "管理员", Code: "admin", Description: "系统管理员，拥有大部分管理权限", Status: "active", SortOrder: 2},
		{Name: "律师", Code: "lawyer", Description: "律师用户，可以管理案件和客户", Status: "active", SortOrder: 3},
		{Name: "助理", Code: "assistant", Description: "律师助理，协助律师处理案件", Status: "active", SortOrder: 4},
		{Name: "财务", Code: "finance", Description: "财务人员，负责财务管理", Status: "active", SortOrder: 5},
		{Name: "实习生", Code: "intern", Description: "实习人员，拥有基础查看权限", Status: "active", SortOrder: 6},
	}

	createdRoles := make(map[string]*models.Role)
	for _, roleReq := range roles {
		role, err := rbacService.CreateRole(context.Background(), roleReq)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("创建角色 %s 失败: %w", roleReq.Name, err)
		}
		createdRoles[roleReq.Code] = role
		log.Printf("✓ 创建角色: %s (%s)", role.Name, role.Code)
	}

	// 创建系统权限（菜单级）
	menuPermissions := []*services.CreatePermissionRequest{
		{Name: "仪表盘", Code: "dashboard", Type: "menu", Path: "/dashboard", Icon: "dashboard", Component: "Dashboard", SortOrder: 1},
		{Name: "用户管理", Code: "user_management", Type: "menu", Path: "/admin/users", Icon: "users", Component: "UserManagement", SortOrder: 2},
		{Name: "角色管理", Code: "role_management", Type: "menu", Path: "/admin/roles", Icon: "role", Component: "RoleManagement", SortOrder: 3},
		{Name: "权限管理", Code: "permission_management", Type: "menu", Path: "/admin/permissions", Icon: "permission", Component: "PermissionManagement", SortOrder: 4},
		{Name: "客户管理", Code: "client_management", Type: "menu", Path: "/clients", Icon: "clients", Component: "ClientManagement", SortOrder: 5},
		{Name: "案件管理", Code: "case_management", Type: "menu", Path: "/cases", Icon: "cases", Component: "CaseManagement", SortOrder: 6},
		{Name: "审批中心", Code: "approval_center", Type: "menu", Path: "/approvals", Icon: "approval", Component: "ApprovalCenter", SortOrder: 7},
		{Name: "财务管理", Code: "finance_management", Type: "menu", Path: "/finance", Icon: "finance", Component: "FinanceManagement", SortOrder: 8},
		{Name: "文档管理", Code: "document_management", Type: "menu", Path: "/documents", Icon: "documents", Component: "DocumentManagement", SortOrder: 9},
		{Name: "工具中心", Code: "tools_center", Type: "menu", Path: "/tools", Icon: "tools", Component: "ToolsCenter", SortOrder: 10},
		{Name: "系统设置", Code: "system_settings", Type: "menu", Path: "/settings", Icon: "settings", Component: "SystemSettings", SortOrder: 11},
		{Name: "统计报表", Code: "statistics_reports", Type: "menu", Path: "/statistics", Icon: "statistics", Component: "StatisticsReports", SortOrder: 12},
	}

	createdPermissions := make(map[string]*models.Permission)
	for _, permReq := range menuPermissions {
		permission, err := rbacService.CreatePermission(context.Background(), permReq)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("创建权限 %s 失败: %w", permReq.Name, err)
		}
		createdPermissions[permReq.Code] = permission
		log.Printf("✓ 创建权限: %s (%s)", permission.Name, permission.Code)
	}

	// 创建按钮级权限
	buttonPermissions := []*services.CreatePermissionRequest{
		// 用户管理按钮权限
		{Name: "查看用户", Code: "user:view", Type: "button", ParentID: &createdPermissions["user_management"].ID, SortOrder: 1},
		{Name: "创建用户", Code: "user:create", Type: "button", ParentID: &createdPermissions["user_management"].ID, SortOrder: 2},
		{Name: "编辑用户", Code: "user:edit", Type: "button", ParentID: &createdPermissions["user_management"].ID, SortOrder: 3},
		{Name: "删除用户", Code: "user:delete", Type: "button", ParentID: &createdPermissions["user_management"].ID, SortOrder: 4},

		// 角色管理按钮权限
		{Name: "查看角色", Code: "role:view", Type: "button", ParentID: &createdPermissions["role_management"].ID, SortOrder: 1},
		{Name: "创建角色", Code: "role:create", Type: "button", ParentID: &createdPermissions["role_management"].ID, SortOrder: 2},
		{Name: "编辑角色", Code: "role:edit", Type: "button", ParentID: &createdPermissions["role_management"].ID, SortOrder: 3},
		{Name: "删除角色", Code: "role:delete", Type: "button", ParentID: &createdPermissions["role_management"].ID, SortOrder: 4},

		// 客户管理按钮权限
		{Name: "查看客户", Code: "client:view", Type: "button", ParentID: &createdPermissions["client_management"].ID, SortOrder: 1},
		{Name: "创建客户", Code: "client:create", Type: "button", ParentID: &createdPermissions["client_management"].ID, SortOrder: 2},
		{Name: "编辑客户", Code: "client:edit", Type: "button", ParentID: &createdPermissions["client_management"].ID, SortOrder: 3},
		{Name: "删除客户", Code: "client:delete", Type: "button", ParentID: &createdPermissions["client_management"].ID, SortOrder: 4},

		// 案件管理按钮权限
		{Name: "查看案件", Code: "case:view", Type: "button", ParentID: &createdPermissions["case_management"].ID, SortOrder: 1},
		{Name: "创建案件", Code: "case:create", Type: "button", ParentID: &createdPermissions["case_management"].ID, SortOrder: 2},
		{Name: "编辑案件", Code: "case:edit", Type: "button", ParentID: &createdPermissions["case_management"].ID, SortOrder: 3},
		{Name: "删除案件", Code: "case:delete", Type: "button", ParentID: &createdPermissions["case_management"].ID, SortOrder: 4},
		{Name: "分配律师", Code: "case:assign", Type: "button", ParentID: &createdPermissions["case_management"].ID, SortOrder: 5},

		// 财务管理按钮权限
		{Name: "查看财务", Code: "finance:view", Type: "button", ParentID: &createdPermissions["finance_management"].ID, SortOrder: 1},
		{Name: "创建财务记录", Code: "finance:create", Type: "button", ParentID: &createdPermissions["finance_management"].ID, SortOrder: 2},
		{Name: "编辑财务记录", Code: "finance:edit", Type: "button", ParentID: &createdPermissions["finance_management"].ID, SortOrder: 3},

		// 文档管理按钮权限
		{Name: "查看文档", Code: "document:view", Type: "button", ParentID: &createdPermissions["document_management"].ID, SortOrder: 1},
		{Name: "上传文档", Code: "document:upload", Type: "button", ParentID: &createdPermissions["document_management"].ID, SortOrder: 2},
		{Name: "编辑文档", Code: "document:edit", Type: "button", ParentID: &createdPermissions["document_management"].ID, SortOrder: 3},
		{Name: "删除文档", Code: "document:delete", Type: "button", ParentID: &createdPermissions["document_management"].ID, SortOrder: 4},
	}

	for _, permReq := range buttonPermissions {
		permission, err := rbacService.CreatePermission(context.Background(), permReq)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("创建权限 %s 失败: %w", permReq.Name, err)
		}
		createdPermissions[permReq.Code] = permission
		log.Printf("✓ 创建权限: %s (%s)", permission.Name, permission.Code)
	}

	// 为超级管理员分配所有权限
	superAdminPerms := make([]uint, 0, len(createdPermissions))
	for _, perm := range createdPermissions {
		superAdminPerms = append(superAdminPerms, perm.ID)
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["super_admin"].ID, superAdminPerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为超级管理员分配权限失败: %w", err)
	}

	// 为管理员分配大部分权限（除了超级管理员专属权限）
	adminPerms := []uint{
		createdPermissions["dashboard"].ID,
		createdPermissions["user_management"].ID,
		createdPermissions["role_management"].ID,
		createdPermissions["client_management"].ID,
		createdPermissions["case_management"].ID,
		createdPermissions["approval_center"].ID,
		createdPermissions["finance_management"].ID,
		createdPermissions["document_management"].ID,
		createdPermissions["tools_center"].ID,
		createdPermissions["system_settings"].ID,
		createdPermissions["statistics_reports"].ID,
		// 按钮权限
		createdPermissions["user:view"].ID,
		createdPermissions["user:create"].ID,
		createdPermissions["user:edit"].ID,
		createdPermissions["user:delete"].ID,
		createdPermissions["role:view"].ID,
		createdPermissions["role:create"].ID,
		createdPermissions["role:edit"].ID,
		createdPermissions["role:delete"].ID,
		createdPermissions["client:view"].ID,
		createdPermissions["client:create"].ID,
		createdPermissions["client:edit"].ID,
		createdPermissions["client:delete"].ID,
		createdPermissions["case:view"].ID,
		createdPermissions["case:create"].ID,
		createdPermissions["case:edit"].ID,
		createdPermissions["case:delete"].ID,
		createdPermissions["case:assign"].ID,
		createdPermissions["finance:view"].ID,
		createdPermissions["finance:create"].ID,
		createdPermissions["finance:edit"].ID,
		createdPermissions["document:view"].ID,
		createdPermissions["document:upload"].ID,
		createdPermissions["document:edit"].ID,
		createdPermissions["document:delete"].ID,
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["admin"].ID, adminPerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为管理员分配权限失败: %w", err)
	}

	// 为律师分配相关权限
	lawyerPerms := []uint{
		createdPermissions["dashboard"].ID,
		createdPermissions["client_management"].ID,
		createdPermissions["case_management"].ID,
		createdPermissions["approval_center"].ID,
		createdPermissions["document_management"].ID,
		createdPermissions["tools_center"].ID,
		createdPermissions["statistics_reports"].ID,
		// 按钮权限
		createdPermissions["client:view"].ID,
		createdPermissions["client:create"].ID,
		createdPermissions["client:edit"].ID,
		createdPermissions["case:view"].ID,
		createdPermissions["case:create"].ID,
		createdPermissions["case:edit"].ID,
		createdPermissions["document:view"].ID,
		createdPermissions["document:upload"].ID,
		createdPermissions["document:edit"].ID,
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["lawyer"].ID, lawyerPerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为律师分配权限失败: %w", err)
	}

	// 为助理分配基础权限
	assistantPerms := []uint{
		createdPermissions["dashboard"].ID,
		createdPermissions["client_management"].ID,
		createdPermissions["case_management"].ID,
		createdPermissions["document_management"].ID,
		createdPermissions["tools_center"].ID,
		// 按钮权限
		createdPermissions["client:view"].ID,
		createdPermissions["case:view"].ID,
		createdPermissions["document:view"].ID,
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["assistant"].ID, assistantPerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为助理分配权限失败: %w", err)
	}

	// 为财务分配财务相关权限
	financePerms := []uint{
		createdPermissions["dashboard"].ID,
		createdPermissions["finance_management"].ID,
		createdPermissions["statistics_reports"].ID,
		// 按钮权限
		createdPermissions["finance:view"].ID,
		createdPermissions["finance:create"].ID,
		createdPermissions["finance:edit"].ID,
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["finance"].ID, financePerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为财务分配权限失败: %w", err)
	}

	// 为实习生分配只读权限
	internPerms := []uint{
		createdPermissions["dashboard"].ID,
		createdPermissions["tools_center"].ID,
	}

	if err := rbacService.AssignRolePermissions(context.Background(), createdRoles["intern"].ID, internPerms); err != nil {
		tx.Rollback()
		return fmt.Errorf("为实习生分配权限失败: %w", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Println("✓ 角色权限数据初始化完成")
	return nil
}