package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
	"law-oa-go/internal/errors"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

type AuthHandler struct {
	userService *services.UserService
}

func NewAuthHandler(userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token     string                `json:"token"`
	ExpiresAt time.Time             `json:"expires_at"`
	User      *services.UserProfile `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// Login godoc
// @Summary 用户登录
// @Description 用户使用邮箱和密码登录系统
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} common.APIResponse{data=LoginResponse} "登录成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "认证失败"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format", "Invalid request format"))
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		_ = c.Error(errors.NewInternalError("token_generation", "Failed to generate token", err))
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.APISuccess(c, response)
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body services.CreateUserRequest true "注册请求"
// @Success 200 {object} common.APIResponse{data=LoginResponse} "注册成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format", "Invalid request format"))
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		_ = c.Error(errors.NewInternalError("token_generation", "Failed to generate token", err))
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.APISuccess(c, response)
}

// GetProfile godoc
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "获取成功"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "用户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(errors.NewAuthorizationError("authentication_required", "Authentication required: User ID not found in token", "authenticated", "none"))
		return
	}

	user, err := h.userService.GetUserProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, user)
}

// UpdateProfile godoc
// @Summary 更新用户资料
// @Description 更新当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body services.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} common.APIResponse{data=services.UserProfile} "更新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 404 {object} common.APIResponse "用户不存在"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(errors.NewAuthorizationError("authentication_required", "Authentication required: User ID not found in token", "authenticated", "none"))
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format", "Invalid request format"))
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, user)
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 修改当前用户的登录密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} common.APIResponse "修改成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未授权"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /users/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		_ = c.Error(errors.NewAuthorizationError("authentication_required", "Authentication required: User ID not found in token", "authenticated", "none"))
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format", "Invalid request format"))
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uint), req.CurrentPassword, req.NewPassword)
	if err != nil {
		_ = c.Error(err)
		return
	}

	common.APISuccess(c, gin.H{"message": "Password changed successfully"})
}

// RefreshToken godoc
// @Summary 刷新Token
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "刷新令牌请求"
// @Success 200 {object} common.APIResponse{data=LoginResponse} "刷新成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "无效的刷新令牌"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(errors.NewValidationError("request_binding", "request_binding", "Invalid request format", "Invalid request format"))
		return
	}

	claims, err := middleware.ValidateToken(req.RefreshToken)
	if err != nil {
		_ = c.Error(errors.NewAuthorizationError("invalid_refresh_token", "Invalid refresh token: "+err.Error(), "valid_token", "invalid_token"))
		return
	}

	user, err := h.userService.GetUserProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		_ = c.Error(errors.NewInternalError("token_generation", "Failed to generate new token", err))
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.APISuccess(c, response)
}

// Logout godoc
// @Summary 用户登出
// @Description 用户登出系统
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.APIResponse "登出成功"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	common.APISuccess(c, gin.H{"message": "Logged out successfully"})
}
