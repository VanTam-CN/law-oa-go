package handlers

import (
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
	Password string `json:"password" binding:"required,min=6,max=50"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Role     string `json:"role" binding:"required,oneof=admin lawyer assistant"`
	Phone    string `json:"phone,omitempty"`
}

type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      interface{} `json:"user"`
}

// Login godoc
// @Summary 用户登录
// @Description 用户邮箱密码登录
// @Tags 认证管理
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
		common.APIBadRequest(c, "请求参数错误", "邮箱和密码不能为空且格式正确")
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		common.APIUnauthorized(c, "认证失败", "邮箱或密码错误")
		return
	}

	// 生成真实的JWT token
	token, expiresAt, err := middleware.GenerateToken(1, req.Email, "lawyer")
	if err != nil {
		common.APIInternalServerError(c, "生成令牌失败", err.Error())
		return
	}

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		User:      user,
	}

	common.APISuccess(c, response)
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求"
// @Success 200 {object} common.APIResponse{data=LoginResponse} "注册成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), &services.CreateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
		Phone:    req.Phone,
	})
	if err != nil {
		common.APIInternalServerError(c, "注册失败", "用户创建失败")
		return
	}

	// 简化token生成
	token := "simple_token_for_dev"
	expiresAt := int64(1234567890)

	response := LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}

	common.APISuccess(c, response)
}