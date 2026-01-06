package services

import (
	"context"
	"fmt"
)

// ContextKey 上下文键类型
type ContextKey string

const (
	// UserIDKey 用户ID键
	UserIDKey ContextKey = "user_id"
	// UsernameKey 用户名键
	UsernameKey ContextKey = "username"
	// RoleKey 角色键
	RoleKey ContextKey = "role"
	// TenantIDKey 租户ID键
	TenantIDKey ContextKey = "tenant_id"
)

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) (userID uint, username string, role string, err error) {
	var ok bool

	// 获取用户ID
	userIDInterface := ctx.Value(UserIDKey)
	if userIDInterface == nil {
		return 0, "", "", fmt.Errorf("user ID not found in context")
	}

	// 尝试转换为不同格式
	switch v := userIDInterface.(type) {
	case uint:
		userID = v
	case int:
		userID = uint(v)
	case int64:
		userID = uint(v)
	case float64:
		userID = uint(v)
	case string:
		var parsedID uint
		if _, err := fmt.Sscanf(v, "%d", &parsedID); err != nil {
			return 0, "", "", fmt.Errorf("invalid user ID format: %v", v)
		}
		userID = parsedID
	default:
		return 0, "", "", fmt.Errorf("unsupported user ID type: %T", userIDInterface)
	}

	// 获取用户名
	usernameInterface := ctx.Value(UsernameKey)
	if usernameInterface != nil {
		username, ok = usernameInterface.(string)
		if !ok {
			return 0, "", "", fmt.Errorf("invalid username format: %T", usernameInterface)
		}
	}

	// 获取角色
	roleInterface := ctx.Value(RoleKey)
	if roleInterface != nil {
		role, ok = roleInterface.(string)
		if !ok {
			return 0, "", "", fmt.Errorf("invalid role format: %T", roleInterface)
		}
	}

	// 获取租户ID
	tenantIDInterface := ctx.Value(TenantIDKey)
	if tenantIDInterface == nil {
		return userID, username, role, fmt.Errorf("tenant ID not found in context")
	}

	return userID, username, role, nil
}

// GetUserIDFromContext 从上下文获取用户ID（简化版）
func GetUserIDFromContext(ctx context.Context) (uint, error) {
	userID, _, _, err := GetUserFromContext(ctx)
	return userID, err
}

// WithUserContext 为用户信息创建上下文
func WithUserContext(ctx context.Context, userID uint, username, role, tenantID string) context.Context {
	ctx = context.WithValue(ctx, UserIDKey, userID)
	ctx = context.WithValue(ctx, UsernameKey, username)
	ctx = context.WithValue(ctx, RoleKey, role)
	ctx = context.WithValue(ctx, TenantIDKey, tenantID)
	return ctx
}