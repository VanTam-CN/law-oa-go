package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
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
// @Success 200 {object} common.Response{data=LoginResponse} "登录成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "认证失败"
// @Failure 500 {object} common.Response "内部错误"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if err.Error() == "user not found" || err.Error() == "invalid password" {
			common.Unauthorized(c, "Invalid credentials")
			return
		}
		common.InternalServerError(c, "Authentication failed: "+err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		common.InternalServerError(c, "Failed to generate token: "+err.Error())
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.Success(c, response)
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户账户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body services.CreateUserRequest true "注册请求"
// @Success 200 {object} common.Response{data=LoginResponse} "注册成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 500 {object} common.Response "内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "email already exists" {
			common.BadRequest(c, "Registration failed: Email address is already registered")
			return
		}
		if err.Error() == "invalid email format" ||
			err.Error() == "password too weak" ||
			err.Error() == "invalid role" {
			common.BadRequest(c, "Validation failed: "+err.Error())
			return
		}
		common.InternalServerError(c, "Registration failed: "+err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		common.InternalServerError(c, "Failed to generate token: "+err.Error())
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.Success(c, response)
}

// GetProfile godoc
// @Summary 获取用户资料
// @Description 获取当前登录用户的资料信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.Response{data=services.UserProfile} "获取成功"
// @Failure 401 {object} common.Response "未授权"
// @Failure 404 {object} common.Response "用户不存在"
// @Failure 500 {object} common.Response "内部错误"
// @Router /users/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "Authentication required: User ID not found in token")
		return
	}

	user, err := h.userService.GetUserProfile(c.Request.Context(), userID.(uint))
	if err != nil {
		if err.Error() == "user not found" {
			common.NotFound(c, "User not found: The requested user does not exist")
			return
		}
		common.InternalServerError(c, "Failed to get user profile: "+err.Error())
		return
	}

	common.Success(c, user)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "Authentication required: User ID not found in token")
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userID.(uint), &req)
	if err != nil {
		if err.Error() == "user not found" {
			common.NotFound(c, "User not found: The requested user does not exist")
			return
		}
		if err.Error() == "email already exists" {
			common.BadRequest(c, "Update failed: Email address is already in use")
			return
		}
		if err.Error() == "invalid email format" {
			common.BadRequest(c, "Validation failed: "+err.Error())
			return
		}
		common.InternalServerError(c, "Failed to update profile: "+err.Error())
		return
	}

	common.Success(c, user)
}

// ChangePassword godoc
// @Summary 修改密码
// @Description 修改当前用户的登录密码
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "修改密码请求"
// @Success 200 {object} common.Response "修改成功"
// @Failure 400 {object} common.Response "请求参数错误"
// @Failure 401 {object} common.Response "未授权"
// @Failure 500 {object} common.Response "内部错误"
// @Router /users/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		common.Unauthorized(c, "Authentication required: User ID not found in token")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userID.(uint), req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "user not found" {
			common.NotFound(c, "User not found: The requested user does not exist")
			return
		}
		if err.Error() == "invalid current password" {
			common.BadRequest(c, "Password change failed: Current password is incorrect")
			return
		}
		if err.Error() == "password too weak" {
			common.BadRequest(c, "Validation failed: New password does not meet security requirements")
			return
		}
		common.InternalServerError(c, "Failed to change password: "+err.Error())
		return
	}

	common.Success(c, gin.H{"message": "Password changed successfully"})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, "Invalid request format: "+err.Error())
		return
	}

	claims, err := middleware.ValidateToken(req.RefreshToken)
	if err != nil {
		common.Unauthorized(c, "Invalid refresh token: "+err.Error())
		return
	}

	user, err := h.userService.GetUserProfile(c.Request.Context(), claims.UserID)
	if err != nil {
		if err.Error() == "user not found" {
			common.Unauthorized(c, "User not found: The user associated with this token no longer exists")
			return
		}
		common.InternalServerError(c, "Failed to refresh token: "+err.Error())
		return
	}

	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		common.InternalServerError(c, "Failed to generate new token: "+err.Error())
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.Success(c, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	common.Success(c, gin.H{"message": "Logged out successfully"})
}
