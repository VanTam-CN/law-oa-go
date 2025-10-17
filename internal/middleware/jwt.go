package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"law-oa-go/internal/config"
)

// JWTManager 线程安全的JWT管理器
type JWTManager struct {
	secret []byte
	mu     sync.RWMutex
}

var jwtManager *JWTManager
var once sync.Once

// InitJWT 初始化 JWT 密钥管理器
func InitJWT(cfg *config.Config) {
	once.Do(func() {
		jwtManager = &JWTManager{
			secret: []byte(cfg.JWT.Secret),
		}
	})
}

// getJWTManager 获取JWT管理器单例
func getJWTManager() *JWTManager {
	if jwtManager == nil {
		panic("JWT manager not initialized. Call InitJWT first.")
	}
	return jwtManager
}

// getSecret 获取密钥（线程安全）
func (j *JWTManager) getSecret() []byte {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.secret
}

// updateSecret 更新密钥（线程安全）
func (j *JWTManager) updateSecret(newSecret string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.secret = []byte(newSecret)
}

// RotateSecret 轮换密钥（线程安全）
func RotateSecret(newSecret string) {
	manager := getJWTManager()
	manager.updateSecret(newSecret)
}

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT 令牌
func GenerateToken(userID uint, username, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Hour * 24) // 24小时过期
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	manager := getJWTManager()
	tokenString, err := token.SignedString(manager.getSecret())
	return tokenString, expiresAt, err
}

// ValidateToken 验证 JWT 令牌
func ValidateToken(tokenString string) (*JWTClaims, error) {
	return ParseToken(tokenString)
}

// ParseToken 解析 JWT 令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	manager := getJWTManager()
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return manager.getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "请提供认证令牌",
			})
			c.Abort()
			return
		}

		// 检查 Bearer 前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌格式",
			})
			c.Abort()
			return
		}

		tokenString := authHeader[7:] // 去掉 "Bearer " 前缀

		// 解析令牌
		claims, err := ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的令牌",
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GetUserID 从Gin上下文中获取用户ID
func GetUserID(c *gin.Context) int {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int); ok {
			return id
		}
	}
	return 0 // 默认返回0，表示未登录或无效用户
}

// RoleMiddleware 角色权限中间件
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "无访问权限",
			})
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限验证失败",
			})
			c.Abort()
			return
		}

		// 检查用户角色是否在允许的角色列表中
		for _, role := range roles {
			if role == roleStr {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "无访问权限",
		})
		c.Abort()
	}
}

// GetCurrentUserID 获取当前用户ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	if id, ok := userID.(uint); ok {
		return id, true
	}

	return 0, false
}

// GetCurrentUsername 获取当前用户名
func GetCurrentUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	if name, ok := username.(string); ok {
		return name, true
	}

	return "", false
}

// GetCurrentRole 获取当前用户角色
func GetCurrentRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}

	if r, ok := role.(string); ok {
		return r, true
	}

	return "", false
}
