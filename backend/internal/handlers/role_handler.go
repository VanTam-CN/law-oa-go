package handlers

import (
	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
)

// RoleHandler 角色处理器
type RoleHandler struct{}

// NewRoleHandler 创建角色处理器
func NewRoleHandler() *RoleHandler {
	return &RoleHandler{}
}

// ListRoles 获取角色列表（分页）
// @Summary 获取角色列表
// @Description 分页获取角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} common.APIResponse "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Router /admin/roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	common.APISuccess(c, gin.H{
		"data": []gin.H{},
		"pagination": gin.H{
			"page":        1,
			"page_size":   20,
			"total":       0,
			"total_pages": 0,
		},
	})
}

// GetAllRoles 获取所有角色（不分页）
// @Summary 获取所有角色
// @Description 获取所有可用角色列表
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Router /admin/roles/all [get]
func (h *RoleHandler) GetAllRoles(c *gin.Context) {
	common.APISuccess(c, []gin.H{
		{"id": 1, "name": "管理员", "code": "admin"},
		{"id": 2, "name": "律师", "code": "lawyer"},
		{"id": 3, "name": "普通用户", "code": "user"},
	})
}

// GetRole 获取角色详情
// @Summary 获取角色详情
// @Description 根据ID获取角色详细信息
// @Tags 角色管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "角色ID"
// @Success 200 {object} common.APIResponse "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "角色不存在"
// @Router /admin/roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	common.APISuccess(c, gin.H{"id": 1, "name": "管理员", "code": "admin"})
}
