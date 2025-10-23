package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SetupJWTExample 设置JWT中间件示例
func SetupJWTExample(db *gorm.DB, logger *logrus.Logger) (*JWTService, error) {
	// 1. 创建JWT工厂
	factory := CreateDevelopmentJWTFactory(db, logger)

	// 2. 获取中间件选项
	options := DevelopmentJWTOptions()

	// 3. 创建JWT服务
	jwtService, err := factory.CreateJWTService(options)
	if err != nil {
		return nil, err
	}

	return jwtService, nil
}

// SetupProductionJWT 设置生产环境JWT中间件示例
func SetupProductionJWT(db *gorm.DB, logger *logrus.Logger, issuer, audience string) (*JWTService, error) {
	// 1. 创建JWT工厂
	factory := CreateProductionJWTFactory(db, logger, issuer, audience)

	// 2. 获取生产环境中间件选项
	options := ProductionJWTOptions()

	// 3. 创建JWT服务
	jwtService, err := factory.CreateJWTService(options)
	if err != nil {
		return nil, err
	}

	return jwtService, nil
}

// CreateRouterWithJWT 创建带JWT认证的路由器示例
func CreateRouterWithJWT(jwtService *JWTService) *gin.Engine {
	// 创建Gin路由器
	router := gin.New()

	// 添加全局中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加请求时间头中间件
	router.Use(func(c *gin.Context) {
		c.Header("X-Request-Time", time.Now().Format(time.RFC3339))
		c.Next()
	})

	// 公开路由（不需要认证）
	public := router.Group("/api/v1/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "healthy",
				"time":   time.Now().Format(time.RFC3339),
			})
		})

		public.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"service": "document-service",
				"version": "1.0.0",
				"auth":    "JWT",
			})
		})
	}

	// 认证路由
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			// 登录逻辑
			// 这里应该验证用户凭据
			// 成功后生成令牌对

			// 示例：模拟登录成功
			userID := uint(1)
			username := "testuser"
			tenantID := "tenant-001"
			roles := []string{"user"}
			permissions := []string{"document:read", "document:write"}
			ipAddress := c.ClientIP()
			userAgent := c.GetHeader("User-Agent")

			// 生成令牌对
			tokenPair, err := jwtService.GenerateTokenPair(
				userID, username, tenantID, roles, permissions, ipAddress, userAgent,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate tokens",
					"code":  "TOKEN_GENERATION_FAILED",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"access_token":  tokenPair.AccessToken,
					"refresh_token": tokenPair.RefreshToken,
					"token_type":    "Bearer",
					"expires_in":    int(jwtService.GetConfig().AccessTokenDuration.Seconds()),
					"user": gin.H{
						"id":       userID,
						"username": username,
						"tenant_id": tenantID,
						"roles":     roles,
					},
				},
			})
		})

		auth.POST("/refresh", func(c *gin.Context) {
			var request struct {
				RefreshToken string `json:"refresh_token" binding:"required"`
			}

			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request format",
					"code":  "INVALID_REQUEST",
				})
				return
			}

			// 刷新令牌
			tokenPair, err := jwtService.RefreshToken(request.RefreshToken)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Failed to refresh token",
					"code":  "REFRESH_FAILED",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"access_token":  tokenPair.AccessToken,
					"refresh_token": tokenPair.RefreshToken,
					"token_type":    "Bearer",
					"expires_in":    int(jwtService.GetConfig().AccessTokenDuration.Seconds()),
				},
			})
		})

		auth.POST("/logout", jwtService.Middleware().Middleware(), func(c *gin.Context) {
			claims, exists := GetClaims(c)
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
					"code":  "INVALID_TOKEN",
				})
				return
			}

			// 撤销令牌
			err := jwtService.RevokeToken(claims.JTI)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to revoke token",
					"code":  "REVOKE_FAILED",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"message": "Successfully logged out",
				},
			})
		})
	}

	// 受保护的路由（需要JWT认证）
	protected := router.Group("/api/v1/protected")
	protected.Use(jwtService.Middleware().Middleware())
	{
		// 基本受保护路由
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := GetUserID(c)
			username, _ := GetUsername(c)
			tenantID, _ := GetTenantID(c)
			claims, _ := GetClaims(c)

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": gin.H{
					"user_id":    userID,
					"username":   username,
					"tenant_id":  tenantID,
					"roles":      claims.Roles,
					"permissions": claims.Permissions,
					"session_id": claims.SessionID,
					"issued_at":  claims.IssuedAt,
					"expires_at": claims.ExpiresAt,
				},
			})
		})

		// 需要特定权限的路由
		documents := protected.Group("/documents")
		documents.Use(jwtService.Middleware().RequirePermission("document", "read"))
		{
			documents.GET("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"documents": []gin.H{
							{"id": 1, "title": "Document 1"},
							{"id": 2, "title": "Document 2"},
						},
					},
				})
			})

			documents.POST("/", jwtService.Middleware().RequirePermission("document", "write"), func(c *gin.Context) {
				// 创建文档逻辑
				c.JSON(http.StatusCreated, gin.H{
					"success": true,
					"data": gin.H{
						"message": "Document created successfully",
						"document_id": 123,
					},
				})
			})

			documents.DELETE("/:id", jwtService.Middleware().RequirePermission("document", "delete"), func(c *gin.Context) {
				// 删除文档逻辑
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"message": "Document deleted successfully",
					},
				})
			})
		}

		// 需要特定角色的路由
		admin := protected.Group("/admin")
		admin.Use(jwtService.Middleware().RequireRole("admin"))
		{
			admin.GET("/users", func(c *gin.Context) {
				// 管理员用户列表
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"users": []gin.H{
							{"id": 1, "username": "user1", "role": "user"},
							{"id": 2, "username": "admin1", "role": "admin"},
						},
					},
				})
			})

			admin.GET("/stats", func(c *gin.Context) {
				// 系统统计信息
				stats, err := jwtService.GetTokenStats()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to get stats",
						"code":  "STATS_FAILED",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"token_stats": stats,
					},
				})
			})

			admin.POST("/cleanup", func(c *gin.Context) {
				// 清理过期令牌
				err := jwtService.CleanupExpiredTokens()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to cleanup tokens",
						"code":  "CLEANUP_FAILED",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"message": "Expired tokens cleaned up successfully",
					},
				})
			})

			admin.POST("/rotate-keys", func(c *gin.Context) {
				// 轮换密钥
				err := jwtService.RotateKeys()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "Failed to rotate keys",
						"code":  "KEY_ROTATION_FAILED",
					})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"message": "Keys rotated successfully",
					},
				})
			})
		}

		// 多权限验证示例
		sensitive := protected.Group("/sensitive")
		sensitive.Use(jwtService.Middleware().RequirePermission("sensitive", "read"))
		sensitive.Use(jwtService.Middleware().RequireRole("admin", "manager"))
		{
			sensitive.GET("/data", func(c *gin.Context) {
				// 敏感数据访问
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"data": gin.H{
						"message": "Access to sensitive data granted",
						"data": "This is sensitive information",
					},
				})
			})
		}
	}

	// 错误处理路由
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Route not found",
			"code":  "ROUTE_NOT_FOUND",
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
			"code":  "METHOD_NOT_ALLOWED",
		})
	})

	return router
}

// CustomTokenExtractorExample 自定义令牌提取器示例
func CustomTokenExtractorExample() TokenExtractor {
	return func(c *gin.Context) (string, error) {
		// 尝试从Authorization header提取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				return parts[1], nil
			}
		}

		// 尝试从Cookie提取
		cookie, err := c.Cookie("auth_token")
		if err == nil && cookie != "" {
			return cookie, nil
		}

		// 尝试从查询参数提取（仅用于开发环境）
		if gin.Mode() == gin.DebugMode {
			token := c.Query("token")
			if token != "" {
				return token, nil
			}
		}

		return "", ErrTokenMissing
	}
}

// CustomErrorHandlerExample 自定义错误处理器示例
func CustomErrorHandlerExample() ErrorHandler {
	return func(c *gin.Context, err error) {
		// 记录错误日志
		logger := c.MustGet("logger").(*logrus.Logger)
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"path":  c.Request.URL.Path,
			"method": c.Request.Method,
			"ip":    c.ClientIP(),
		}).Error("JWT authentication error")

		// 返回统一格式的错误响应
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "AUTH_FAILED",
				"message": "Authentication failed",
				"details": err.Error(),
			},
			"meta": gin.H{
				"timestamp": time.Now().Format(time.RFC3339),
				"path":      c.Request.URL.Path,
			},
		})
	}
}