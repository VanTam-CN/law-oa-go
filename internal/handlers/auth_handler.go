package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/auth"
	"law-oa-go/internal/common"
	errs "law-oa-go/internal/errors"
	"law-oa-go/internal/middleware"
	"law-oa-go/internal/services"
)

type AuthHandler struct {
	userService             *services.UserService
	tokenRevocationService  *auth.TokenRevocationService
}

func NewAuthHandler(userService *services.UserService, tokenRevocationService *auth.TokenRevocationService) *AuthHandler {
	return &AuthHandler{
		userService:            userService,
		tokenRevocationService: tokenRevocationService,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Account  string `json:"account"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Phone    string `json:"phone,omitempty"`
}

// publicRegistrationRole 公开注册强制固定的角色，杜绝请求体提权。
const publicRegistrationRole = "user"

// usernameFromEmail 由规范化邮箱派生稳定 username：本地部分 + 完整邮箱 SHA-256 前 8 位，
// 截断到 50 字符；同一邮箱生成稳定结果，且避免常见本地部分冲突。
func usernameFromEmail(rawEmail string) string {
	email := strings.ToLower(strings.TrimSpace(rawEmail))
	if email == "" {
		return ""
	}
	localPart := email
	if idx := strings.Index(email, "@"); idx > 0 {
		localPart = email[:idx]
	}
	localPart = strings.TrimSpace(localPart)
	if localPart == "" {
		localPart = "user"
	}
	sum := sha256.Sum256([]byte(email))
	suffix := hex.EncodeToString(sum[:])[:8]
	username := localPart + "-" + suffix
	if len(username) > 50 {
		username = username[:50]
	}
	return username
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
		common.APIBadRequest(c, "请求参数错误", "账号和密码不能为空")
		return
	}

	email := normalizeLoginAccount(req.Email, req.Account)
	if email == "" {
		common.APIBadRequest(c, "请求参数错误", "请输入账号或邮箱")
		return
	}

	user, err := h.userService.AuthenticateUser(c.Request.Context(), email, req.Password)
	if err != nil {
		common.APIUnauthorized(c, "认证失败", "邮箱或密码错误")
		return
	}

	// 生成真实的JWT token
	token, expiresAt, err := middleware.GenerateToken(user.ID, email, user.Role)
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

func normalizeLoginAccount(email string, account string) string {
	identifier := strings.ToLower(strings.TrimSpace(email))
	if identifier == "" {
		identifier = strings.ToLower(strings.TrimSpace(account))
	}

	if identifier == "" {
		return ""
	}

	aliases := map[string]string{
		"admin":            "demo.admin@example.test",
		"demo.admin":     "demo.admin@example.test",
		"lawyer":           "demo.lawyer@example.test",
		"demo.lawyer":    "demo.lawyer@example.test",
		"assistant":        "demo.assistant@example.test",
		"demo.assistant": "demo.assistant@example.test",
		"finance":          "demo.finance@example.test",
		"demo.finance":   "demo.finance@example.test",
	}

	if resolved, ok := aliases[identifier]; ok {
		return resolved
	}

	return identifier
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

	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := h.userService.CreateUser(c.Request.Context(), &services.CreateUserRequest{
		Username: usernameFromEmail(email),
		Name:     req.Name,
		Email:    email,
		Password: req.Password,
		Role:     publicRegistrationRole,
		Phone:    req.Phone,
	})
	if err != nil {
		// 邮箱冲突属于客户端输入问题，返回 400 而非 500；其他错误按内部错误处理。
		var bizErr *errs.EnhancedError
		if errors.As(err, &bizErr) && bizErr.Code() == "BUSINESS_ERROR" {
			common.APIBadRequest(c, "注册失败", err.Error())
			return
		}
		common.APIInternalServerError(c, "注册失败", "用户创建失败")
		return
	}

	// 生成真实 JWT，禁止返回 dev 占位符
	token, expiresAt, err := middleware.GenerateToken(user.ID, user.Email, user.Role)
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

// ============================================================================
// 令牌撤销相关处理器 (Token Revocation Handlers)
// ============================================================================

// LogoutRequest 登出请求
type LogoutRequest struct {
	Token string `json:"token" binding:"required"`
}

// RevokeByUserRequest 撤销用户令牌请求
type RevokeByUserRequest struct {
	UserID uint `json:"user_id" binding:"required"`
}

// RevokeByDeviceRequest 撤销设备令牌请求
type RevokeByDeviceRequest struct {
	UserID   uint `json:"user_id" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"`
}

// OffboardingRevocationRequest 离职撤销请求
type OffboardingRevocationRequest struct {
	UserID                uint   `json:"user_id" binding:"required"`
	SuccessorID           uint   `json:"successor_id" binding:"required"`
	TransferredCaseCount  int    `json:"transferred_case_count"`
	TransferredCaseIDs    []uint `json:"transferred_case_ids"`
	HandoverNote          string `json:"handover_note"`
}

// Logout godoc
// @Summary 用户登出
// @Description 撤销当前访问令牌（用户登出）
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body LogoutRequest true "登出请求"
// @Success 200 {object} common.APIResponse{data=auth.RevocationResult} "登出成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "token不能为空")
		return
	}

	// 获取客户端IP
	ipAddress := c.ClientIP()

	result, err := h.tokenRevocationService.RevokeSingle(c.Request.Context(), req.Token, ipAddress)
	if err != nil {
		common.APIInternalServerError(c, "登出失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// RevokeUserTokens godoc
// @Summary 撤销用户所有令牌
// @Description 撤销指定用户的所有令牌（用于密码重置等场景）
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RevokeByUserRequest true "撤销请求"
// @Success 200 {object} common.APIResponse{data=auth.RevocationResult} "撤销成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/revoke/user [post]
func (h *AuthHandler) RevokeUserTokens(c *gin.Context) {
	var req RevokeByUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "user_id不能为空")
		return
	}

	// 获取当前用户ID（从JWT中提取）
	currentUserID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	// 只有管理员或用户本人可以撤销
	currentID := currentUserID.(uint)
	if currentID != req.UserID {
		role, ok := middleware.GetCurrentRole(c)
		if !ok || (role != "admin" && role != "super_admin") {
			common.APIForbidden(c, "权限不足", "只能撤销自己的令牌")
			return
		}
	}

	ipAddress := c.ClientIP()
	result, err := h.tokenRevocationService.RevokeByUser(c.Request.Context(), req.UserID, ipAddress)
	if err != nil {
		common.APIInternalServerError(c, "撤销失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// RevokeDeviceTokens godoc
// @Summary 撤销设备所有令牌
// @Description 撤销指定设备的所有令牌（用于安全事件等场景）
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body RevokeByDeviceRequest true "撤销请求"
// @Success 200 {object} common.APIResponse{data=auth.RevocationResult} "撤销成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/revoke/device [post]
func (h *AuthHandler) RevokeDeviceTokens(c *gin.Context) {
	var req RevokeByDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "user_id和device_id不能为空")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	// 只有管理员或用户本人可以撤销
	currentID := currentUserID.(uint)
	if currentID != req.UserID {
		role, ok := middleware.GetCurrentRole(c)
		if !ok || (role != "admin" && role != "super_admin") {
			common.APIForbidden(c, "权限不足", "只能撤销自己的设备令牌")
			return
		}
	}

	ipAddress := c.ClientIP()
	result, err := h.tokenRevocationService.RevokeByDevice(c.Request.Context(), req.UserID, req.DeviceID, ipAddress)
	if err != nil {
		common.APIInternalServerError(c, "撤销失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// RevokeAllTokens godoc
// @Summary 撤销用户所有令牌和设备（离职）
// @Description 撤销用户所有令牌和设备信息（用于离职场景）
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body OffboardingRevocationRequest true "离职撤销请求"
// @Success 200 {object} common.APIResponse{data=auth.RevocationResult} "撤销成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/revoke/all [post]
func (h *AuthHandler) RevokeAllTokens(c *gin.Context) {
	var req OffboardingRevocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", "请检查所有必填字段")
		return
	}

	// 只有管理员可以执行离职撤销
	role, ok := middleware.GetCurrentRole(c)
	if !ok || (role != "admin" && role != "super_admin") {
		common.APIForbidden(c, "权限不足", "只有管理员可以执行离职撤销")
		return
	}

	offboardingData := &auth.OffboardingData{
		SuccessorID:          req.SuccessorID,
		TransferredCaseCount: req.TransferredCaseCount,
		TransferredCaseIDs:   req.TransferredCaseIDs,
		HandoverNote:         req.HandoverNote,
	}

	ipAddress := c.ClientIP()
	result, err := h.tokenRevocationService.RevokeAll(c.Request.Context(), req.UserID, offboardingData, ipAddress)
	if err != nil {
		common.APIInternalServerError(c, "撤销失败", err.Error())
		return
	}

	common.APISuccess(c, result)
}

// GetActiveDevices godoc
// @Summary 获取用户活动设备
// @Description 获取指定用户的所有活动设备信息
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path int true "用户ID"
// @Success 200 {object} common.APIResponse{data=[]map[string]interface{}} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/devices/:user_id [get]
func (h *AuthHandler) GetActiveDevices(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		common.APIBadRequest(c, "请求参数错误", "user_id不能为空")
		return
	}

	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		common.APIBadRequest(c, "请求参数错误", "user_id格式无效")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	// 只有管理员或用户本人可以查看
	currentID := currentUserID.(uint)
	if currentID != id {
		role, ok := middleware.GetCurrentRole(c)
		if !ok || (role != "admin" && role != "super_admin") {
			common.APIForbidden(c, "权限不足", "只能查看自己的设备信息")
			return
		}
	}

	devices, err := h.tokenRevocationService.GetUserActiveDevices(c.Request.Context(), id)
	if err != nil {
		common.APIInternalServerError(c, "获取设备失败", err.Error())
		return
	}

	common.APISuccess(c, devices)
}

// GetRevocationHistory godoc
// @Summary 获取令牌撤销历史
// @Description 获取指定用户的令牌撤销历史记录
// @Tags 认证管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param user_id path int true "用户ID"
// @Param limit query int false "限制数量" default(50)
// @Success 200 {object} common.APIResponse{data=[]models.TokenRevocationLog} "获取成功"
// @Failure 400 {object} common.APIResponse "请求参数错误"
// @Failure 401 {object} common.APIResponse "未认证"
// @Failure 403 {object} common.APIResponse "权限不足"
// @Failure 500 {object} common.APIResponse "内部错误"
// @Router /auth/revocation-history/:user_id [get]
func (h *AuthHandler) GetRevocationHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		common.APIBadRequest(c, "请求参数错误", "user_id不能为空")
		return
	}

	var id uint
	if _, err := fmt.Sscanf(userID, "%d", &id); err != nil {
		common.APIBadRequest(c, "请求参数错误", "user_id格式无效")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := c.Get("user_id")
	if !exists {
		common.APIUnauthorized(c, "未认证", "用户信息无效")
		return
	}

	// 只有管理员或用户本人可以查看
	currentID := currentUserID.(uint)
	if currentID != id {
		role, ok := middleware.GetCurrentRole(c)
		if !ok || (role != "admin" && role != "super_admin") {
			common.APIForbidden(c, "权限不足", "只能查看自己的撤销历史")
			return
		}
	}

	// 获取limit参数
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
			limit = l
		}
	}

	history, err := h.tokenRevocationService.GetRevocationHistory(c.Request.Context(), id, limit)
	if err != nil {
		common.APIInternalServerError(c, "获取历史失败", err.Error())
		return
	}

	common.APISuccess(c, history)
}
