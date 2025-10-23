package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// JWTSecurityIntegration JWT安全集成
type JWTSecurityIntegration struct {
	jwtService      *JWTService
	securityService *SecurityService
	securityMiddleware *SecurityMiddleware
	logger          *logrus.Logger
	jwtConfig       *JWTConfig
	securityConfig  *SecurityConfig
}

// NewJWTSecurityIntegration 创建JWT安全集成
func NewJWTSecurityIntegration(
	db *gorm.DB,
	jwtConfig *JWTConfig,
	securityConfig *SecurityConfig,
	logger *logrus.Logger,
) (*JWTSecurityIntegration, error) {
	// 创建JWT工厂
	factory := NewJWTFactory(db, logger, jwtConfig)

	// 创建JWT服务
	jwtOptions := ProductionJWTOptions()
	jwtService, err := factory.CreateJWTService(jwtOptions)
	if err != nil {
		return nil, err
	}

	// 创建安全服务
	securityService := NewSecurityService(db, logger, securityConfig)

	// 创建安全中间件
	securityMiddleware := NewSecurityMiddleware(db, logger, securityConfig)

	return &JWTSecurityIntegration{
		jwtService:        jwtService,
		securityService:   securityService,
		securityMiddleware: securityMiddleware,
		logger:           logger,
		jwtConfig:        jwtConfig,
		securityConfig:   securityConfig,
	}, nil
}

// GetJWTService 获取JWT服务
func (i *JWTSecurityIntegration) GetJWTService() *JWTService {
	return i.jwtService
}

// GetSecurityService 获取安全服务
func (i *JWTSecurityIntegration) GetSecurityService() *SecurityService {
	return i.securityService
}

// GetSecurityMiddleware 获取安全中间件
func (i *JWTSecurityIntegration) GetSecurityMiddleware() *SecurityMiddleware {
	return i.securityMiddleware
}

// SetupSecureRoutes 设置安全路由
func (i *JWTSecurityIntegration) SetupSecureRoutes(router *gin.Engine) {
	// 获取中间件
	jwtMiddleware := i.jwtService.Middleware().Middleware()
	securityMiddleware := i.securityMiddleware.Middleware()

	// 应用安全中间件（在JWT中间件之前）
	router.Use(securityMiddleware)

	// 公开路由（只需要安全检查）
	public := router.Group("/api/v1/public")
	{
		public.GET("/health", i.HealthCheck)
		public.GET("/security-info", i.securityMiddleware.SecurityInfo())
	}

	// 需要JWT认证的路由
	protected := router.Group("/api/v1")
	protected.Use(jwtMiddleware)
	{
		// 基本认证路由
		auth := protected.Group("/auth")
		{
			auth.POST("/login", i.SecureLogin)
			auth.POST("/logout", i.SecureLogout)
			auth.POST("/refresh", i.SecureRefreshToken)
		}

		// 用户路由
		user := protected.Group("/user")
		{
			user.GET("/profile", i.GetUserProfile)
			user.PUT("/profile", i.UpdateUserProfile)
			user.POST("/change-password", i.ChangePassword)
			user.GET("/security-info", i.GetUserSecurityInfo)
			user.POST("/revoke-sessions", i.RevokeAllUserSessions)
		}

		// 管理员路由
		admin := protected.Group("/admin")
		admin.Use(i.RequireRole("admin"))
		{
			admin.GET("/security/metrics", i.securityMiddleware.getMetrics())
			admin.GET("/security/report", i.securityMiddleware.SecurityReport())
			admin.PUT("/security/config", i.securityMiddleware.updateConfig())
			admin.POST("/users/:user_id/block", i.BlockUser)
			admin.POST("/users/:user_id/unblock", i.UnblockUser)
		}
	}

	// 高安全路由（需要低风险评分）
	HighSecurity := protected.Group("/high-security")
	HighSecurity.Use(i.securityMiddleware.RequireMaxRiskScore(0.3))
	{
		HighSecurity.GET("/sensitive-data", i.GetSensitiveData)
		HighSecurity.POST("/admin-actions", i.PerformAdminAction)
	}

	// 注册安全相关路由
	i.securityMiddleware.RegisterSecurityRoutes(router, jwtMiddleware)
}

// SecureLogin 安全登录
func (i *JWTSecurityIntegration) SecureLogin(c *gin.Context) {
	var request struct {
		Username   string      `json:"username" binding:"required"`
		Password   string      `json:"password" binding:"required"`
		TenantID   string      `json:"tenant_id,omitempty"`
		RememberMe bool        `json:"remember_me"`
		DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
		Captcha    string      `json:"captcha,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		i.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		}).Warn("Invalid login request")

		c.JSON(400, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
			},
		})
		return
	}

	// 获取安全检查结果
	securityResult, exists := c.Get("security_result")
	if exists {
		result := securityResult.(*SecurityCheckResult)
		if result.RiskScore > 0.5 {
			// 高风险登录，要求额外验证
			i.handleHighRiskLogin(c, request, result)
			return
		}
	}

	// 执行用户认证
	user, err := i.authenticateUser(request.Username, request.Password)
	if err != nil {
		i.handleFailedLogin(c, request, err)
		return
	}

	// 检查用户是否被阻止
	if i.isUserBlocked(user.ID) {
		i.handleBlockedUser(c, user)
		return
	}

	// 生成令牌
	tokenPair, err := i.generateSecureTokenPair(user, request)
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate secure token pair")

		c.JSON(500, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_GENERATION_FAILED",
				"message": "Failed to generate authentication tokens",
			},
		})
		return
	}

	// 记录成功登录
	i.recordSuccessfulLogin(c, user, request)

	// 返回响应
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"token_type":    tokenPair.TokenType,
			"expires_in":    tokenPair.ExpiresIn,
			"user": gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"tenant_id":  user.TenantID,
				"roles":      user.Roles,
			},
			"security": gin.H{
				"risk_score":    i.getRiskScoreFromContext(c),
				"threats_count": i.getThreatsCountFromContext(c),
				"recommendations": i.getRecommendationsFromContext(c),
			},
		},
	})
}

// SecureLogout 安全登出
func (i *JWTSecurityIntegration) SecureLogout(c *gin.Context) {
	claims, exists := GetClaims(c)
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid token",
			},
		})
		return
	}

	// 撤销所有令牌
	err := i.revokeAllUserTokens(claims.UserID, "user_logout")
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"user_id": claims.UserID,
			"error":   err.Error(),
		}).Error("Failed to revoke user tokens")
	}

	// 记录安全登出
	i.recordSecurityLogout(c, claims)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}

// SecureRefreshToken 安全令牌刷新
func (i *JWTSecurityIntegration) SecureRefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
		DeviceInfo   *DeviceInfo `json:"device_info,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
			},
		})
		return
	}

	// 获取当前风险评分
	riskScore := i.getRiskScoreFromContext(c)
	if riskScore > 0.7 {
		// 高风险刷新，要求额外验证
		c.JSON(403, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "HIGH_RISK_REFRESH",
				"message": "High risk detected, additional verification required",
			},
		})
		return
	}

	// 执行令牌刷新
	tokenPair, err := i.jwtService.RefreshToken(request.RefreshToken)
	if err != nil {
		i.handleFailedTokenRefresh(c, request, err)
		return
	}

	// 记录令牌刷新
	i.recordTokenRefresh(c, request)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"token_type":    tokenPair.TokenType,
			"expires_in":    tokenPair.ExpiresIn,
		},
	})
}

// GetUserProfile 获取用户配置文件
func (i *JWTSecurityIntegration) GetUserProfile(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 获取用户信息
	user, err := i.getUserByID(userID)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "USER_NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	}

	// 获取安全信息
	securityInfo := i.getUserSecurityInfo(userID)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"user": gin.H{
				"id":         user.ID,
				"username":   user.Username,
				"tenant_id":  user.TenantID,
				"roles":      user.Roles,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
			"security": securityInfo,
		},
	})
}

// GetUserSecurityInfo 获取用户安全信息
func (i *JWTSecurityIntegration) GetUserSecurityInfo(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	securityInfo := i.getUserSecurityInfo(userID)

	c.JSON(200, gin.H{
		"success": true,
		"data":    securityInfo,
	})
}

// GetSensitiveData 获取敏感数据（高安全路由）
func (i *JWTSecurityIntegration) GetSensitiveData(c *gin.Context) {
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 检查用户权限
	roles, _ := c.Get("roles")
	userRoles, ok := roles.([]string)
	if !ok || !i.hasAdminRole(userRoles) {
		c.JSON(403, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INSUFFICIENT_PRIVILEGES",
				"message": "Admin privileges required",
			},
		})
		return
	}

	// 获取风险评分
	riskScore := i.getRiskScoreFromContext(c)
	if riskScore > 0.3 {
		c.JSON(403, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "RISK_TOO_HIGH",
				"message": "Risk score too high for sensitive operation",
				"details": gin.H{
					"current_score": riskScore,
					"max_allowed":  0.3,
				},
			},
		})
		return
	}

	// 返回敏感数据（模拟）
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"sensitive_data": "This is highly sensitive information",
			"access_granted": true,
			"security_level": "high",
		},
	})
}

// HealthCheck 健康检查
func (i *JWTSecurityIntegration) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"services": gin.H{
			"jwt":      "operational",
			"security": "operational",
		},
		"version": "1.0.0",
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    status,
	})
}

// 辅助方法

// authenticateUser 用户认证
func (i *JWTSecurityIntegration) authenticateUser(username, password string) (*User, error) {
	// 这里应该实现实际的数据库认证
	// 简化实现
	if username == "admin" && password == "admin123" {
		return &User{
			ID:       1,
			Username: "admin",
			TenantID: "default",
			Roles:    []string{"admin"},
		}, nil
	}
	if username == "user" && password == "user123" {
		return &User{
			ID:       2,
			Username: "user",
			TenantID: "default",
			Roles:    []string{"user"},
		}, nil
	}
	return nil, ErrInvalidCredentials
}

// isUserBlocked 检查用户是否被阻止
func (i *JWTSecurityIntegration) isUserBlocked(userID uint) bool {
	// 这里应该查询数据库或缓存
	return false
}

// generateSecureTokenPair 生成安全令牌对
func (i *JWTSecurityIntegration) generateSecureTokenPair(user *User, request struct {
	Username   string
	TenantID   string
	DeviceInfo *DeviceInfo
}) (*TokenPair, error) {
	return i.jwtService.GenerateTokenPair(
		user.ID,
		user.Username,
		request.TenantID,
		user.Roles,
		[]string{"*"}, // 简化权限
		"",            // IP地址将由中间件设置
		"",            // UserAgent将由中间件设置
	)
}

// recordSuccessfulLogin 记录成功登录
func (i *JWTSecurityIntegration) recordSuccessfulLogin(c *gin.Context, user *User, request struct {
	Username string
}) {
	i.logger.WithFields(logrus.Fields{
		"user_id":  user.ID,
		"username": user.Username,
		"ip":       c.ClientIP(),
		"event":    "login_success",
	}).Info("User logged in successfully")
}

// recordSecurityLogout 记录安全登出
func (i *JWTSecurityIntegration) recordSecurityLogout(c *gin.Context, claims *TokenClaims) {
	i.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"username":   claims.Username,
		"session_id": claims.SessionID,
		"ip":         c.ClientIP(),
		"event":      "logout_success",
	}).Info("User logged out successfully")
}

// handleHighRiskLogin 处理高风险登录
func (i *JWTSecurityIntegration) handleHighRiskLogin(c *gin.Context, request struct {
	Username string
}, result *SecurityCheckResult) {
	c.JSON(403, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "HIGH_RISK_LOGIN",
			"message": "High risk detected, additional verification required",
			"details": gin.H{
				"risk_score":   result.RiskScore,
				"threats":      result.Threats,
				"recommendations": result.Recommendations,
			},
		},
	})
}

// handleFailedLogin 处理失败登录
func (i *JWTSecurityIntegration) handleFailedLogin(c *gin.Context, request struct {
	Username string
}, err error) {
	i.logger.WithFields(logrus.Fields{
		"username": request.Username,
		"ip":       c.ClientIP(),
		"error":    err.Error(),
	}).Warn("Login failed")

	c.JSON(401, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "AUTHENTICATION_FAILED",
			"message": "Invalid username or password",
		},
	})
}

// handleBlockedUser 处理被阻止的用户
func (i *JWTSecurityIntegration) handleBlockedUser(c *gin.Context, user *User) {
	c.JSON(403, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "USER_BLOCKED",
			"message": "User account is blocked",
		},
	})
}

// revokeAllUserTokens 撤销用户所有令牌
func (i *JWTSecurityIntegration) revokeAllUserTokens(userID uint, reason string) error {
	return i.jwtService.RevokeUserTokens(userID)
}

// recordTokenRefresh 记录令牌刷新
func (i *JWTSecurityIntegration) recordTokenRefresh(c *gin.Context, request struct {
	RefreshToken string
}) {
	i.logger.WithFields(logrus.Fields{
		"ip":    c.ClientIP(),
		"event": "token_refresh",
	}).Info("Token refreshed")
}

// handleFailedTokenRefresh 处理失败的令牌刷新
func (i *JWTSecurityIntegration) handleFailedTokenRefresh(c *gin.Context, request struct {
	RefreshToken string
}, err error) {
	i.logger.WithFields(logrus.Fields{
		"ip":    c.ClientIP(),
		"error": err.Error(),
	}).Warn("Token refresh failed")

	c.JSON(401, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "REFRESH_FAILED",
			"message": "Failed to refresh token",
		},
	})
}

// getUserByID 根据ID获取用户
func (i *JWTSecurityIntegration) getUserByID(userID uint) (*User, error) {
	// 这里应该查询数据库
	return &User{
		ID:       userID,
		Username: "testuser",
		TenantID: "default",
		Roles:    []string{"user"},
	}, nil
}

// getUserSecurityInfo 获取用户安全信息
func (i *JWTSecurityIntegration) getUserSecurityInfo(userID uint) gin.H {
	// 这里应该查询数据库获取实际的安全信息
	return gin.H{
		"last_login":       time.Now().Add(-2 * time.Hour),
		"failed_attempts":  0,
		"blocked_until":    nil,
		"active_sessions":  1,
		"security_events": []gin.H{
			{
				"type":      "login",
				"timestamp": time.Now().Add(-2 * time.Hour),
				"ip":        "127.0.0.1",
				"success":   true,
			},
		},
	}
}

// getRiskScoreFromContext 从上下文获取风险评分
func (i *JWTSecurityIntegration) getRiskScoreFromContext(c *gin.Context) float64 {
	if riskScore, exists := c.Get("risk_score"); exists {
		if score, ok := riskScore.(float64); ok {
			return score
		}
	}
	return 0.0
}

// getThreatsCountFromContext 从上下文获取威胁数量
func (i *JWTSecurityIntegration) getThreatsCountFromContext(c *gin.Context) int {
	if threats, exists := c.Get("security_threats"); exists {
		if threatList, ok := threats.([]SecurityThreat); ok {
			return len(threatList)
		}
	}
	return 0
}

// getRecommendationsFromContext 从上下文获取建议
func (i *JWTSecurityIntegration) getRecommendationsFromContext(c *gin.Context) []SecurityRecommendation {
	if recommendations, exists := c.Get("security_recommendations"); exists {
		if recList, ok := recommendations.([]SecurityRecommendation); ok {
			return recList
		}
	}
	return []SecurityRecommendation{}
}

// hasRole 检查是否有指定角色
func (i *JWTSecurityIntegration) hasRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

// hasAdminRole 检查是否有管理员角色
func (i *JWTSecurityIntegration) hasAdminRole(roles []string) bool {
	return i.hasRole(roles, "admin")
}

// RequireRole 要求指定角色的中间件
func (i *JWTSecurityIntegration) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(401, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User roles not found",
				},
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok || !i.hasRole(userRoles, role) {
			c.JSON(403, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient privileges",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 其他方法的占位符实现
func (i *JWTSecurityIntegration) UpdateUserProfile(c *gin.Context)    { c.JSON(501, gin.H{"message": "Not implemented"}) }
func (i *JWTSecurityIntegration) ChangePassword(c *gin.Context)       { c.JSON(501, gin.H{"message": "Not implemented"}) }
func (i *JWTSecurityIntegration) RevokeAllUserSessions(c *gin.Context) { c.JSON(501, gin.H{"message": "Not implemented"}) }
func (i *JWTSecurityIntegration) BlockUser(c *gin.Context)              { c.JSON(501, gin.H{"message": "Not implemented"}) }
func (i *JWTSecurityIntegration) UnblockUser(c *gin.Context)            { c.JSON(501, gin.H{"message": "Not implemented"}) }
func (i *JWTSecurityIntegration) PerformAdminAction(c *gin.Context)      { c.JSON(501, gin.H{"message": "Not implemented"}) }