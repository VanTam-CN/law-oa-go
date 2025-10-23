package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// JWTMiddleware JWT中间件
type JWTMiddleware struct {
	manager    *JWTManager
	validator  *TokenValidator
	config     *JWTConfig
	logger     *logrus.Logger
	options    *MiddlewareOptions
}

// MiddlewareOptions 中间件选项
type MiddlewareOptions struct {
	// 跳过验证的路径
	SkipPaths []string
	// 跳过验证的方法
	SkipMethods []string
	// 是否验证刷新令牌
	ValidateRefresh bool
	// 是否检查令牌是否在黑名单中
	CheckBlacklist bool
	// 是否缓存验证结果
	EnableCache bool
	// 缓存过期时间
	CacheTTL time.Duration
	// 提取器函数
	Extractor TokenExtractor
	// 错误处理函数
	ErrorHandler ErrorHandler
}

// DefaultMiddlewareOptions 默认中间件选项
func DefaultMiddlewareOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		SkipPaths:    []string{"/health", "/metrics", "/api/v1/auth/login", "/api/v1/auth/refresh"},
		SkipMethods:  []string{"GET"},
		ValidateRefresh: false,
		CheckBlacklist: true,
		EnableCache:   true,
		CacheTTL:      5 * time.Minute,
		Extractor:     DefaultTokenExtractor,
		ErrorHandler:  DefaultErrorHandler,
	}
}

// NewJWTMiddleware 创建JWT中间件
func NewJWTMiddleware(manager *JWTManager, validator *TokenValidator, config *JWTConfig, logger *logrus.Logger, options *MiddlewareOptions) *JWTMiddleware {
	if options == nil {
		options = DefaultMiddlewareOptions()
	}

	return &JWTMiddleware{
		manager:   manager,
		validator: validator,
		config:    config,
		logger:    logger,
		options:   options,
	}
}

// Middleware 返回Gin中间件函数
func (m *JWTMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否需要跳过验证
		if m.shouldSkip(c) {
			c.Next()
			return
		}

		// 提取令牌
		tokenString, err := m.options.Extractor(c)
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
				"ip":    c.ClientIP(),
			}).Warn("Token extraction failed")

			m.options.ErrorHandler(c, ErrTokenMissing)
			c.Abort()
			return
		}

		// 验证令牌
		claims, err := m.validateToken(tokenString, c)
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"error": err.Error(),
				"path":  c.Request.URL.Path,
				"ip":    c.ClientIP(),
			}).Warn("Token validation failed")

			m.options.ErrorHandler(c, err)
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		m.setContextClaims(c, claims)

		// 记录访问日志
		m.logAccess(c, claims)

		c.Next()
	}
}

// shouldSkip 检查是否需要跳过验证
func (m *JWTMiddleware) shouldSkip(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	// 检查路径是否在跳过列表中
	for _, skipPath := range m.options.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	// 检查方法是否在跳过列表中
	for _, skipMethod := range m.options.SkipMethods {
		if method == skipMethod {
			return true
		}
	}

	return false
}

// validateToken 验证令牌
func (m *JWTMiddleware) validateToken(tokenString string, c *gin.Context) (*TokenClaims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if m.options.ValidateRefresh {
			return m.manager.refreshKeyPair.PublicKey, nil
		}
		return m.manager.accessKeyPair.PublicKey, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, ErrTokenExpired
		}
		if strings.Contains(err.Error(), "signature is invalid") {
			return nil, ErrTokenInvalid
		}
		return nil, ErrTokenMalformed
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	// 检查令牌类型
	if m.options.ValidateRefresh && claims.TokenType != "refresh" {
		return nil, ErrTokenTypeMismatch
	}

	if !m.options.ValidateRefresh && claims.TokenType != "access" {
		return nil, ErrTokenTypeMismatch
	}

	// 检查令牌是否在黑名单中
	if m.options.CheckBlacklist {
		if m.manager.IsBlacklisted(claims.JTI) {
			return nil, ErrTokenBlacklisted
		}
	}

	// 使用验证器验证声明
	if validationErrors := m.validator.ValidateClaims(claims); len(validationErrors) > 0 {
		m.logger.WithFields(logrus.Fields{
			"jti":    claims.JTI,
			"userID": claims.UserID,
			"errors": validationErrors,
		}).Warn("Token claims validation failed")
		return nil, ErrClaimsInvalid
	}

	// 验证请求上下文
	if err := m.validateRequestContext(claims, c); err != nil {
		return nil, err
	}

	return claims, nil
}

// validateRequestContext 验证请求上下文
func (m *JWTMiddleware) validateRequestContext(claims *TokenClaims, c *gin.Context) error {
	// 验证IP地址
	if claims.IPAddress != "" && claims.IPAddress != c.ClientIP() {
		m.logger.WithFields(logrus.Fields{
			"expectedIP": claims.IPAddress,
			"actualIP":   c.ClientIP(),
			"userID":     claims.UserID,
		}).Warn("IP address mismatch")
		return ErrIPMismatch
	}

	// 验证用户代理
	if claims.UserAgent != "" && claims.UserAgent != c.GetHeader("User-Agent") {
		m.logger.WithFields(logrus.Fields{
			"expectedUA": claims.UserAgent,
			"actualUA":   c.GetHeader("User-Agent"),
			"userID":     claims.UserID,
		}).Warn("User agent mismatch")
		return ErrUserAgentMismatch
	}

	return nil
}

// setContextClaims 将用户信息存储到上下文中
func (m *JWTMiddleware) setContextClaims(c *gin.Context, claims *TokenClaims) {
	// 存储完整的声明
	c.Set("claims", claims)

	// 存储常用字段以便快速访问
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("tenant_id", claims.TenantID)
	c.Set("session_id", claims.SessionID)
	c.Set("roles", claims.Roles)
	c.Set("permissions", claims.Permissions)

	// 存储上下钥值对
	ctx := context.WithValue(c.Request.Context(), "claims", claims)
	ctx = context.WithValue(ctx, "user_id", claims.UserID)
	ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
	c.Request = c.Request.WithContext(ctx)
}

// logAccess 记录访问日志
func (m *JWTMiddleware) logAccess(c *gin.Context, claims *TokenClaims) {
	m.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"tenant_id":  claims.TenantID,
		"session_id": claims.SessionID,
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"ip":         c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"jti":        claims.JTI,
	}).Info("API access")
}

// RequirePermission 权限检查中间件
func (m *JWTMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			m.options.ErrorHandler(c, ErrTokenMissing)
			c.Abort()
			return
		}

		tokenClaims := claims.(*TokenClaims)

		// 检查权限
		if !m.hasPermission(tokenClaims, resource, action) {
			m.logger.WithFields(logrus.Fields{
				"user_id":  tokenClaims.UserID,
				"resource": resource,
				"action":   action,
				"path":     c.Request.URL.Path,
			}).Warn("Permission denied")

			m.options.ErrorHandler(c, ErrPermissionDenied)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 角色检查中间件
func (m *JWTMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			m.options.ErrorHandler(c, ErrTokenMissing)
			c.Abort()
			return
		}

		tokenClaims := claims.(*TokenClaims)

		// 检查角色
		if !m.hasRole(tokenClaims, roles...) {
			m.logger.WithFields(logrus.Fields{
				"user_id":   tokenClaims.UserID,
				"required":  roles,
				"actual":    tokenClaims.Roles,
				"path":      c.Request.URL.Path,
			}).Warn("Role check failed")

			m.options.ErrorHandler(c, ErrRoleRequired)
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasPermission 检查是否有指定权限
func (m *JWTMiddleware) hasPermission(claims *TokenClaims, resource, action string) bool {
	// 管理员有所有权限
	for _, role := range claims.Roles {
		if role == "admin" {
			return true
		}
	}

	// 检查资源访问权限
	if claims.ResourceAccess != nil {
		if resourcePerms, exists := claims.ResourceAccess[resource]; exists {
			switch perms := resourcePerms.(type) {
			case map[string]interface{}:
				if allowed, exists := perms[action]; exists {
					return allowed.(bool)
				}
			case bool:
				return perms
			case string:
				return perms == action
			}
		}
	}

	// 检查全局权限
	for _, perm := range claims.Permissions {
		if perm == resource+":"+action || perm == "*" {
			return true
		}
	}

	return false
}

// hasRole 检查是否有指定角色
func (m *JWTMiddleware) hasRole(claims *TokenClaims, roles ...string) bool {
	for _, requiredRole := range roles {
		for _, userRole := range claims.Roles {
			if userRole == requiredRole {
				return true
			}
		}
	}
	return false
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetTenantID 从上下文获取租户ID
func GetTenantID(c *gin.Context) (string, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return "", false
	}
	return tenantID.(string), true
}

// GetClaims 从上下文获取完整声明
func GetClaims(c *gin.Context) (*TokenClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	return claims.(*TokenClaims), true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}