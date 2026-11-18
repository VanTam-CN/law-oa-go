package auth

import (
	"github.com/gin-gonic/gin"
)

// GetUserID 从 Gin 上下文中获取用户ID
// 这是隔离墙中间件和其他业务代码使用的主要接口
func GetUserID(c *gin.Context) uint {
	if userID, exists := c.Get(ContextKeyUserID); exists {
		switch v := userID.(type) {
		case uint:
			return v
		case int:
			return uint(v)
		case int64:
			return uint(v)
		case float64:
			return uint(v)
		}
	}
	return 0
}

// GetUsername 从 Gin 上下文中获取用户名
func GetUsername(c *gin.Context) string {
	if username, exists := c.Get(ContextKeyUsername); exists {
		if s, ok := username.(string); ok {
			return s
		}
	}
	return ""
}

// GetEmail 从 Gin 上下文中获取邮箱
func GetEmail(c *gin.Context) string {
	if email, exists := c.Get(ContextKeyEmail); exists {
		if s, ok := email.(string); ok {
			return s
		}
	}
	return ""
}

// GetRole 从 Gin 上下文中获取角色
func GetRole(c *gin.Context) string {
	if role, exists := c.Get(ContextKeyRole); exists {
		if s, ok := role.(string); ok {
			return s
		}
	}
	return ""
}

// GetClaims 从 Gin 上下文中获取完整的令牌声明
func GetClaims(c *gin.Context) *TokenClaims {
	if claims, exists := c.Get(ContextKeyClaims); exists {
		if tc, ok := claims.(*TokenClaims); ok {
			return tc
		}
	}
	return nil
}

// GetUserIDAsString 从 Gin 上下文中获取用户ID（字符串格式）
// 用于需要字符串类型用户ID的场景
func GetUserIDAsString(c *gin.Context) string {
	userID := GetUserID(c)
	if userID == 0 {
		return ""
	}
	// 简单的 uint 转 string
	return string(rune(userID))
}

// IsAuthenticated 检查当前请求是否已认证
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) > 0
}

// HasRole 检查当前用户是否具有指定角色
func HasRole(c *gin.Context, role string) bool {
	return GetRole(c) == role
}

// HasAnyRole 检查当前用户是否具有任一指定角色
func HasAnyRole(c *gin.Context, roles ...string) bool {
	currentRole := GetRole(c)
	for _, role := range roles {
		if currentRole == role {
			return true
		}
	}
	return false
}

// SetAuthContext 将认证信息设置到上下文中
// 供认证中间件内部使用
func SetAuthContext(c *gin.Context, claims *TokenClaims) {
	c.Set(ContextKeyUserID, claims.UserID)
	c.Set(ContextKeyUsername, claims.Username)
	c.Set(ContextKeyEmail, claims.Email)
	c.Set(ContextKeyRole, claims.Role)
	c.Set(ContextKeyClaims, claims)
}
