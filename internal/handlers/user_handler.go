package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/middleware"
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
// @Success 200 {object} common.APIResponse{data=[]services.UserProfile} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	handler := APIListHandler(func(c *gin.Context, req *services.UserListRequest) ([]*services.UserProfile, int64, error) {
		return h.userService.ListUsers(c.Request.Context(), req)
	}, "users")
	handler(c)
}

// GetUser godoc
// @Summary 获取用户详情
// @Description 根据ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "用户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	handler := APIGetHandler(func(c *gin.Context, id uint) (*services.UserProfile, error) {
		return h.userService.GetUserProfile(c.Request.Context(), id)
	}, "user")
	handler(c)
}

// CreateUser godoc
// @Summary 创建用户
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.CreateUserRequest true "创建用户请求"
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "创建成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	handler := APICreateHandler(func(c *gin.Context, req *services.CreateUserRequest) (*services.UserProfile, error) {
		return h.userService.CreateUser(c.Request.Context(), req)
	}, "user")
	handler(c)
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
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "用户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	handler := APIUpdateHandler(func(c *gin.Context, id uint, req *services.UpdateUserRequest) (*services.UserProfile, error) {
		return h.userService.UpdateUser(c.Request.Context(), id, req)
	}, "user")
	handler(c)
}

// DeleteUser godoc
// @Summary 删除用户
// @Description 删除指定的用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} common.APIResponse "删除成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 404 {object} common.APIResponse "用户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	handler := APIDeleteHandler(func(c *gin.Context, id uint) error {
		return h.userService.DeleteUser(c.Request.Context(), id)
	}, "user")
	handler(c)
}

// GetCurrentUser godoc
// @Summary 获取当前用户信息
// @Description 从 JWT context 获取当前登录用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户信息无效")
		return
	}

	profile, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		common.APIInternalServerError(c, "获取用户信息失败", err.Error())
		return
	}

	common.APISuccess(c, profile)
}

// GetProfile godoc
// @Summary 获取当前用户资料
// @Description 从 JWT context 获取当前登录用户的详细资料
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户信息无效")
		return
	}

	profile, err := h.userService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		common.APIInternalServerError(c, "获取用户资料失败", err.Error())
		return
	}

	common.APISuccess(c, profile)
}

type updateProfileRequest struct {
	Name       *string `json:"name"`
	RealName   *string `json:"real_name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	Department *string `json:"department"`
	Seniority  *string `json:"seniority"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UpdateProfile 更新当前用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户信息无效")
		return
	}

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	name := req.Name
	if name == nil {
		name = req.RealName
	}
	profile, err := h.userService.UpdateUser(c.Request.Context(), userID, &services.UpdateUserRequest{
		Name:       name,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
		Seniority:  req.Seniority,
	})
	if err != nil {
		common.APIInternalServerError(c, "更新用户资料失败", err.Error())
		return
	}

	common.APISuccess(c, profile)
}

// ChangePassword 修改当前用户密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户信息无效")
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "当前密码和新密码不能为空，且新密码至少8位")
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		common.APIBadRequest(c, "修改密码失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{"message": "密码修改成功"})
}

// UploadAvatar 上传当前用户头像
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		common.APIUnauthorized(c, "未授权", "用户信息无效")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		common.APIBadRequest(c, "头像文件不能为空", err.Error())
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		common.APIBadRequest(c, "头像格式不支持", "仅支持 JPG、PNG、GIF、WEBP")
		return
	}
	if file.Size > 2*1024*1024 {
		common.APIBadRequest(c, "头像文件过大", "头像文件不能超过2MB")
		return
	}

	uploadDir := filepath.Join("uploads", "avatars")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		common.APIInternalServerError(c, "创建上传目录失败", err.Error())
		return
	}

	fileName := fmt.Sprintf("%d_%d%s", userID, timeNowUnixMilli(), ext)
	dst := filepath.Join(uploadDir, fileName)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		common.APIInternalServerError(c, "保存头像失败", err.Error())
		return
	}

	avatarPath := "/" + filepath.ToSlash(dst)
	profile, err := h.userService.UpdateUserAvatar(c.Request.Context(), userID, avatarPath)
	if err != nil {
		common.APIInternalServerError(c, "更新头像失败", err.Error())
		return
	}

	common.APISuccess(c, gin.H{
		"url":  avatarPath,
		"user": profile,
	})
}

func timeNowUnixMilli() int64 {
	return time.Now().UnixMilli()
}
