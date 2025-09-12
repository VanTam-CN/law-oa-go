package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/services"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers godoc
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持过滤和搜索
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param role query string false "用户角色" Enums(admin,lawyer,user)
// @Param status query string false "用户状态" Enums(active,inactive)
// @Param search query string false "搜索关键词"
// @Success 200 {object} common.PageResponse{data=[]services.UserProfile} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "内部错误"
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req services.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	users, total, err := h.userService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		common.InternalServerError(c, "Failed to list users", err.Error())
		return
	}

	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	response := common.PageResponse{
		Data:     users,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetUser godoc
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response{data=services.UserProfile} "获取成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "用户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	user, err := h.userService.GetUserProfile(c.Request.Context(), uint(id))
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "User not found", "The requested user does not exist")
			return
		}
		common.InternalServerError(c, "Failed to get user", err.Error())
		return
	}

	common.Success(c, user)
}

// CreateUser godoc
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateUserRequest true "创建用户请求"
// @Success 200 {object} common.Response{data=services.UserProfile} "创建成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 500 {object} common.Response "内部错误"
// @Router /admin/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		if common.IsValidationError(err) {
			common.BadRequest(c, "Validation failed", err.Error())
			return
		}
		common.InternalServerError(c, "Failed to create user", err.Error())
		return
	}

	common.Success(c, user)
}

// UpdateUser godoc
// @Summary 更新用户信息
// @Description 更新指定用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body services.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} common.Response{data=services.UserProfile} "更新成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "用户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format", err.Error())
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "User not found", "The requested user does not exist")
			return
		}
		if common.IsValidationError(err) {
			common.BadRequest(c, "Validation failed", err.Error())
			return
		}
		common.InternalServerError(c, "Failed to update user", err.Error())
		return
	}

	common.Success(c, user)
}

// DeleteUser godoc
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response "删除成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 403 {object} common.Response "权限不足"
// @Failure 404 {object} common.Response "用户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.BadRequest(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		if common.IsNotFoundError(err) {
			common.NotFound(c, "User not found", "The requested user does not exist")
			return
		}
		common.InternalServerError(c, "Failed to delete user", err.Error())
		return
	}

	common.Success(c, gin.H{"message": "User deleted successfully"})
}