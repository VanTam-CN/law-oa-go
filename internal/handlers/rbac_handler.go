package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

// RBACHandler 角色权限管理处理器
type RBACHandler struct {
	rbacService *services.RBACService
}

// NewRBACHandler 创建RBAC处理器
func NewRBACHandler(rbacService *services.RBACService) *RBACHandler {
	return &RBACHandler{
		rbacService: rbacService,
	}
}

// ========== 角色管理 API ==========

// GetRoleList godoc
// @Summary 获取角色列表
// @Description 获取角色列表，支持分页和过滤
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "角色名称"
// @Param code query string false "角色代码"
// @Param status query string false "状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} common.APIResponse{data=services.RolePageResponse} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles [get]
func (h *RBACHandler) GetRoleList(c *gin.Context) {
	var params services.RoleQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIBadRequest(c, "Invalid query parameters")
		return
	}

	roles, err := h.rbacService.GetRoleList(c.Request.Context(), &params)
	if err != nil {
		common.APIInternalServerError(c, "Failed to get role list")
		return
	}

	common.APISuccess(c, roles)
}

// GetAllRoles godoc
// @Summary 获取所有角色
// @Description 获取所有可用角色（不分页）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]models.Role} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/all [get]
func (h *RBACHandler) GetAllRoles(c *gin.Context) {
	roles, err := h.rbacService.GetAllRoles(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "Failed to get all roles")
		return
	}

	common.APISuccess(c, roles)
}

// GetRoleById godoc
// @Summary 获取角色详情
// @Description 根据ID获取角色详情
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} common.APIResponse{data=models.Role} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "角色不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id} [get]
func (h *RBACHandler) GetRoleById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	role, err := h.rbacService.GetRoleById(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Role not found")
		return
	}

	common.APISuccess(c, role)
}

// CreateRole godoc
// @Summary 创建角色
// @Description 创建新角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateRoleRequest true "创建角色请求"
// @Success 200 {object} common.APIResponse{data=models.Role} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 409 {object} common.APIResponse "角色名称或代码已存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles [post]
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req services.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}

	role, err := h.rbacService.CreateRole(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "role name already exists" || err.Error() == "role code already exists" {
			common.NewAPIError(c, http.StatusConflict, "ROLE_EXISTS", err.Error())
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create role")
		return
	}

	common.APISuccess(c, role)
}

// UpdateRole godoc
// @Summary 更新角色
// @Description 更新角色信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body services.UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} common.APIResponse{data=models.Role} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "角色不存在"
// @Failure 409 {object} common.APIResponse "角色名称已存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id} [put]
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	var req services.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}

	role, err := h.rbacService.UpdateRole(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "role name already exists" {
			common.NewAPIError(c, http.StatusConflict, "ROLE_EXISTS", err.Error())
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update role")
		return
	}

	common.APISuccess(c, role)
}

// DeleteRole godoc
// @Summary 删除角色
// @Description 删除角色
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "角色已被用户使用，无法删除"
// @Failure 404 {object} common.APIResponse "角色不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id} [delete]
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	err = h.rbacService.DeleteRole(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "cannot delete role: users are still assigned to this role" {
			common.NewAPIError(c, http.StatusForbidden, "ROLE_IN_USE", "Role is in use and cannot be deleted")
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete role")
		return
	}

	common.APISuccess(c, gin.H{"message": "Role deleted successfully"})
}

// UpdateRoleStatus godoc
// @Summary 更新角色状态
// @Description 更新角色状态（启用/禁用）
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body object{status string} true "状态更新请求"
// @Success 200 {object} common.APIResponse "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "角色不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id}/status [put]
func (h *RBACHandler) UpdateRoleStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active inactive"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}

	err = h.rbacService.UpdateRoleStatus(c.Request.Context(), uint(id), req.Status)
	if err != nil {
		common.NewAPIError(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update role status")
		return
	}

	common.APISuccess(c, gin.H{"message": "Role status updated successfully"})
}

// ========== 权限管理 API ==========

// GetPermissionList godoc
// @Summary 获取权限列表
// @Description 获取权限列表（树形结构）
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "权限名称"
// @Param code query string false "权限代码"
// @Param type query string false "权限类型"
// @Param status query string false "状态"
// @Success 200 {object} common.APIResponse{data=[]models.Permission} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions [get]
func (h *RBACHandler) GetPermissionList(c *gin.Context) {
	var params services.PermissionQueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		common.APIBadRequest(c, "Invalid query parameters")
		return
	}

	permissions, err := h.rbacService.GetPermissionList(c.Request.Context(), &params)
	if err != nil {
		common.APIInternalServerError(c, "Failed to get permission list")
		return
	}

	common.APISuccess(c, permissions)
}

// CreatePermission godoc
// @Summary 创建权限
// @Description 创建新权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreatePermissionRequest true "创建权限请求"
// @Success 200 {object} common.APIResponse{data=models.Permission} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 409 {object} common.APIResponse "权限代码已存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions [post]
func (h *RBACHandler) CreatePermission(c *gin.Context) {
	var req services.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}

	permission, err := h.rbacService.CreatePermission(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "permission code already exists" {
			common.NewAPIError(c, http.StatusConflict, "PERMISSION_EXISTS", err.Error())
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create permission")
		return
	}

	common.APISuccess(c, permission)
}

// GetPermissionById godoc
// @Summary 获取权限详情
// @Description 根据ID获取权限详情
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} common.APIResponse{data=models.Permission} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "权限不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions/{id} [get]
func (h *RBACHandler) GetPermissionById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid permission ID")
		return
	}

	permission, err := h.rbacService.GetPermissionById(c.Request.Context(), uint(id))
	if err != nil {
		common.APINotFound(c, "Permission not found")
		return
	}

	common.APISuccess(c, permission)
}

// UpdatePermission godoc
// @Summary 更新权限
// @Description 更新权限信息
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Param request body services.UpdatePermissionRequest true "更新权限请求"
// @Success 200 {object} common.APIResponse{data=models.Permission} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "权限不存在"
// @Failure 409 {object} common.APIResponse "权限代码已存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions/{id} [put]
func (h *RBACHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid permission ID")
		return
	}

	var req services.UpdatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}

	permission, err := h.rbacService.UpdatePermission(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "permission code already exists" {
			common.NewAPIError(c, http.StatusConflict, "PERMISSION_EXISTS", err.Error())
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update permission")
		return
	}

	common.APISuccess(c, permission)
}

// DeletePermission godoc
// @Summary 删除权限
// @Description 删除权限
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限有子权限，无法删除"
// @Failure 404 {object} common.APIResponse "权限不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions/{id} [delete]
func (h *RBACHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid permission ID")
		return
	}

	err = h.rbacService.DeletePermission(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "cannot delete permission: it has child permissions" {
			common.NewAPIError(c, http.StatusForbidden, "PERMISSION_HAS_CHILDREN", "Permission has child permissions and cannot be deleted")
			return
		}
		if err.Error() == "cannot delete permission: roles are still assigned to this permission" {
			common.NewAPIError(c, http.StatusForbidden, "PERMISSION_IN_USE", "Permission is in use and cannot be deleted")
			return
		}
		common.NewAPIError(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete permission")
		return
	}

	common.APISuccess(c, gin.H{"message": "Permission deleted successfully"})
}

// GetAllPermissions godoc
// @Summary 获取所有权限
// @Description 获取所有权限（扁平结构）
// @Tags 权限管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]models.Permission} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/permissions/all [get]
func (h *RBACHandler) GetAllPermissions(c *gin.Context) {
	permissions, err := h.rbacService.GetAllPermissions(c.Request.Context())
	if err != nil {
		common.APIInternalServerError(c, "Failed to get all permissions")
		return
	}

	common.APISuccess(c, permissions)
}

// ========== 角色权限关联 API ==========

// GetRolePermissions godoc
// @Summary 获取角色的权限列表
// @Description 获取指定角色的权限ID列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} common.APIResponse{data=[]uint} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id}/permissions [get]
func (h *RBACHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	permissionIDs, err := h.rbacService.GetRolePermissions(c.Request.Context(), uint(id))
	if err != nil {
		common.APIInternalServerError(c, "Failed to get role permissions")
		return
	}

	common.APISuccess(c, permissionIDs)
}

// AssignRolePermissions godoc
// @Summary 为角色分配权限
// @Description 为指定角色分配权限
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Param request body object{permission_ids []uint} true "权限分配请求"
// @Success 200 {object} common.APIResponse "分配成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/roles/{id}/permissions [post]
func (h *RBACHandler) AssignRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		common.APIBadRequest(c, "Invalid role ID")
		return
	}

	var req struct {
		PermissionIDs      []uint `json:"permission_ids"`
		PermissionIDsCamel []uint `json:"permissionIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}
	if len(req.PermissionIDs) == 0 {
		req.PermissionIDs = req.PermissionIDsCamel
	}
	if req.PermissionIDs == nil {
		common.APIBadRequest(c, "permission_ids is required")
		return
	}

	err = h.rbacService.AssignRolePermissions(c.Request.Context(), uint(id), req.PermissionIDs)
	if err != nil {
		common.NewAPIError(c, http.StatusInternalServerError, "ASSIGN_FAILED", "Failed to assign role permissions")
		return
	}

	common.APISuccess(c, gin.H{"message": "Role permissions assigned successfully"})
}

// ========== 用户角色关联 API ==========

// GetUserRoles godoc
// @Summary 获取用户的角色列表
// @Description 获取指定用户的角色列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Success 200 {object} common.APIResponse{data=[]models.Role} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users/{user_id}/roles [get]
func (h *RBACHandler) GetUserRoles(c *gin.Context) {
	userIDParam := c.Param("user_id")
	if userIDParam == "" {
		userIDParam = c.Param("id")
	}
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		common.NewAPIError(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	roles, err := h.rbacService.GetUserRoles(c.Request.Context(), uint(userID))
	if err != nil {
		common.APIInternalServerError(c, "Failed to get user roles")
		return
	}

	common.APISuccess(c, roles)
}

// AssignUserRoles godoc
// @Summary 为用户分配角色
// @Description 为指定用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "用户ID"
// @Param request body object{role_ids []uint} true "角色分配请求"
// @Success 200 {object} common.APIResponse "分配成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users/{user_id}/roles [post]
func (h *RBACHandler) AssignUserRoles(c *gin.Context) {
	userIDParam := c.Param("user_id")
	if userIDParam == "" {
		userIDParam = c.Param("id")
	}
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		common.NewAPIError(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req struct {
		RoleIDs      []uint `json:"role_ids"`
		RoleIDsCamel []uint `json:"roleIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "Invalid request parameters")
		return
	}
	if len(req.RoleIDs) == 0 {
		req.RoleIDs = req.RoleIDsCamel
	}
	if req.RoleIDs == nil {
		common.APIBadRequest(c, "role_ids is required")
		return
	}

	err = h.rbacService.AssignUserRoles(c.Request.Context(), uint(userID), req.RoleIDs)
	if err != nil {
		common.NewAPIError(c, http.StatusInternalServerError, "ASSIGN_FAILED", "Failed to assign user roles")
		return
	}

	common.APISuccess(c, gin.H{"message": "User roles assigned successfully"})
}

// ========== 当前用户权限 API ==========

// GetCurrentUserRoles godoc
// @Summary 获取当前用户的角色列表
// @Description 获取当前登录用户的角色列表
// @Tags 用户权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]models.Role} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/current-user/roles [get]
func (h *RBACHandler) GetCurrentUserRoles(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.NewAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	roles, err := h.rbacService.GetUserRoles(c.Request.Context(), userID.(uint))
	if err != nil {
		common.APIInternalServerError(c, "Failed to get current user roles")
		return
	}

	common.APISuccess(c, roles)
}

// GetCurrentUserPermissions godoc
// @Summary 获取当前用户的权限列表
// @Description 获取当前登录用户的权限列表
// @Tags 用户权限
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=[]models.Permission} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/current-user/permissions [get]
func (h *RBACHandler) GetCurrentUserPermissions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.NewAPIError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	permissions, err := h.rbacService.GetUserPermissions(c.Request.Context(), userID.(uint))
	if err != nil {
		common.APIInternalServerError(c, "Failed to get current user permissions")
		return
	}

	common.APISuccess(c, permissions)
}
