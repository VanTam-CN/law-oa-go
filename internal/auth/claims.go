package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

const (
	// ContextKeyUserID 用户ID在上下文中的键
	ContextKeyUserID = "user_id"
	// ContextKeyUsername 用户名在上下文中的键
	ContextKeyUsername = "username"
	// ContextKeyEmail 邮箱在上下文中的键
	ContextKeyEmail = "email"
	// ContextKeyRole 角色在上下文中的键
	ContextKeyRole = "role"
	// ContextKeyClaims 完整声明在上下文中的键
	ContextKeyClaims = "auth_claims"
)

// TokenClaims 统一的令牌声明结构
// 用于向后兼容现有两种 JWT 实现
type TokenClaims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type,omitempty"` // access 或 refresh
	jwt.RegisteredClaims
}

// NewTokenClaims 创建新的令牌声明
func NewTokenClaims(userID uint, username, email, role, tokenType string) *TokenClaims {
	return &TokenClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
	}
}

// GetUserID 获取用户ID
func (c *TokenClaims) GetUserID() uint {
	return c.UserID
}

// GetUsername 获取用户名
func (c *TokenClaims) GetUsername() string {
	return c.Username
}

// GetEmail 获取邮箱
func (c *TokenClaims) GetEmail() string {
	return c.Email
}

// GetRole 获取角色
func (c *TokenClaims) GetRole() string {
	return c.Role
}

// IsAccessToken 是否为访问令牌
func (c *TokenClaims) IsAccessToken() bool {
	return c.TokenType == "access" || c.TokenType == ""
}

// IsRefreshToken 是否为刷新令牌
func (c *TokenClaims) IsRefreshToken() bool {
	return c.TokenType == "refresh"
}
