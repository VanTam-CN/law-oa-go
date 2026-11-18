package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
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
		if !ok || (roleStr != "admin" && roleStr != "super_admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "需要管理员权限",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
