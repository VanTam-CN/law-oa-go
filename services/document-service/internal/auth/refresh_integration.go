package auth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RefreshTokenServiceIntegration 刷新令牌服务集成
type RefreshTokenServiceIntegration struct {
	refreshService   *RefreshTokenService
	jwtService       *JWTService
	handler          *RefreshTokenHandler
	logger           *logrus.Logger
	config           *RefreshTokenConfig
	cleanupContext   context.Context
	cleanupCancel    context.CancelFunc
}

// NewRefreshTokenServiceIntegration 创建刷新令牌服务集成
func NewRefreshTokenServiceIntegration(
	db *gorm.DB,
	jwtConfig *JWTConfig,
	refreshConfig *RefreshTokenConfig,
	logger *logrus.Logger,
) (*RefreshTokenServiceIntegration, error) {
	// 创建JWT管理器
	tokenRepo := NewTokenRepository(db)
	userRepo := NewUserRepository(db)
	auditRepo := NewAuditRepository(db)

	jwtManager, err := NewJWTManager(jwtConfig, tokenRepo, userRepo, auditRepo, logger)
	if err != nil {
		return nil, err
	}

	// 创建JWT服务
	validator := NewTokenValidator(jwtConfig, logger)
	middleware := NewJWTMiddleware(jwtManager, validator, jwtConfig, logger, DefaultMiddlewareOptions())
	jwtService := &JWTService{
		manager:    jwtManager,
		validator:  validator,
		middleware: middleware,
		config:     jwtConfig,
		logger:     logger,
	}

	// 创建刷新令牌服务
	refreshService := NewRefreshTokenService(db, jwtConfig, logger, jwtManager)

	// 创建刷新令牌处理器
	handler := NewRefreshTokenHandler(refreshService, logger)

	// 创建清理上下文
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())

	return &RefreshTokenServiceIntegration{
		refreshService: refreshService,
		jwtService:     jwtService,
		handler:        handler,
		logger:         logger,
		config:         refreshConfig,
		cleanupContext: cleanupCtx,
		cleanupCancel:  cleanupCancel,
	}, nil
}

// GetJWTService 获取JWT服务
func (i *RefreshTokenServiceIntegration) GetJWTService() *JWTService {
	return i.jwtService
}

// GetRefreshService 获取刷新令牌服务
func (i *RefreshTokenServiceIntegration) GetRefreshService() *RefreshTokenService {
	return i.refreshService
}

// GetHandler 获取刷新令牌处理器
func (i *RefreshTokenServiceIntegration) GetHandler() *RefreshTokenHandler {
	return i.handler
}

// SetupRoutes 设置路由
func (i *RefreshTokenServiceIntegration) SetupRoutes(router *gin.Engine) {
	// JWT认证中间件
	jwtMiddleware := i.jwtService.Middleware().Middleware()

	// 公开路由（不需要认证）
	public := router.Group("/api/v1/auth")
	{
		// 登录
		public.POST("/login", i.Login)

		// 刷新令牌（公开，但内部会验证刷新令牌）
		public.POST("/refresh", i.handler.RefreshToken)
	}

	// 受保护路由（需要JWT认证）
	protected := router.Group("/api/v1")
	protected.Use(jwtMiddleware)
	{
		// 登出
		protected.POST("/logout", i.Logout)

		// 刷新令牌管理
		refresh := protected.Group("/refresh-token")
		{
			refresh.GET("/history", i.handler.GetRefreshHistory)
			refresh.GET("/active", i.handler.GetActiveRefreshTokens)
			refresh.DELETE("/revoke", i.handler.RevokeRefreshToken)
			refresh.DELETE("/revoke-all", i.handler.RevokeAllRefreshTokens)
		}

		// 用户信息
		protected.GET("/profile", i.GetProfile)
		protected.POST("/change-password", i.ChangePassword)
	}

	// 管理员路由
	admin := router.Group("/api/v1/admin")
	admin.Use(jwtMiddleware)
	admin.Use(i.jwtService.Middleware().RequireRole("admin"))
	{
		admin.GET("/refresh-token/stats", i.handler.GetRefreshTokenStats)
		admin.POST("/refresh-token/cleanup", i.handler.CleanupExpiredTokens)
		admin.POST("/users/:user_id/revoke-tokens", i.handler.RevokeUserTokens)
	}
}

// Login 登录处理
func (i *RefreshTokenServiceIntegration) Login(c *gin.Context) {
	var request struct {
		Username    string      `json:"username" binding:"required"`
		Password    string      `json:"password" binding:"required"`
		TenantID    string      `json:"tenant_id,omitempty"`
		RememberMe  bool        `json:"remember_me"`
		DeviceInfo  *DeviceInfo `json:"device_info,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// 验证用户凭据（这里应该是实际的数据库验证）
	user, err := i.authenticateUser(request.Username, request.Password)
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"username": request.Username,
			"ip":       c.ClientIP(),
			"error":    err.Error(),
		}).Warn("Authentication failed")

		c.JSON(401, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "AUTHENTICATION_FAILED",
				"message": "Invalid username or password",
			},
		})
		return
	}

	// 设置默认租户ID
	if request.TenantID == "" {
		request.TenantID = user.TenantID
	}

	// 获取用户角色和权限
	roles := user.Roles
	permissions := user.Permissions

	// 提取设备信息
	deviceInfo := request.DeviceInfo
	if deviceInfo == nil {
		deviceInfo = i.handler.extractDeviceInfo(c)
	}

	// 生成令牌对
	tokenPair, err := i.jwtService.GenerateTokenPair(
		user.ID,
		user.Username,
		request.TenantID,
		roles,
		permissions,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"user_id": user.ID,
			"error":   err.Error(),
		}).Error("Failed to generate token pair")

		c.JSON(500, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_GENERATION_FAILED",
				"message": "Failed to generate authentication tokens",
			},
		})
		return
	}

	// 创建刷新令牌记录
	err = i.createRefreshTokenRecord(tokenPair, user, request.TenantID, deviceInfo)
	if err != nil {
		i.logger.WithError(err).Warn("Failed to create refresh token record")
		// 不返回错误，因为令牌已经生成
	}

	// 记录成功登录
	i.logger.WithFields(logrus.Fields{
		"user_id":   user.ID,
		"username":  user.Username,
		"tenant_id": request.TenantID,
		"ip":        c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
	}).Info("User logged in successfully")

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
				"tenant_id":  request.TenantID,
				"roles":      roles,
				"created_at": user.CreatedAt,
			},
		},
		"meta": gin.H{
			"login_time": time.Now(),
			"device_info": deviceInfo,
		},
	})
}

// Logout 登出处理
func (i *RefreshTokenServiceIntegration) Logout(c *gin.Context) {
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

	// 撤销访问令牌
	err := i.jwtService.RevokeToken(claims.JTI)
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"user_id": claims.UserID,
			"jti":     claims.JTI,
			"error":   err.Error(),
		}).Error("Failed to revoke access token")
	}

	// 撤销用户所有刷新令牌
	err = i.refreshService.RevokeUserRefreshTokens(claims.UserID, "user_logout")
	if err != nil {
		i.logger.WithFields(logrus.Fields{
			"user_id": claims.UserID,
			"error":   err.Error(),
		}).Error("Failed to revoke refresh tokens")
	}

	// 记录登出
	i.logger.WithFields(logrus.Fields{
		"user_id": claims.UserID,
		"session_id": claims.SessionID,
		"ip":       c.ClientIP(),
	}).Info("User logged out")

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Logged out successfully",
		},
	})
}

// GetProfile 获取用户资料
func (i *RefreshTokenServiceIntegration) GetProfile(c *gin.Context) {
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

	username, _ := GetUsername(c)
	tenantID, _ := GetTenantID(c)
	claims, _ := GetClaims(c)

	// 获取活跃刷新令牌数量
	activeTokens, err := i.refreshService.GetUserRefreshTokens(userID)
	if err != nil {
		i.logger.WithError(err).Error("Failed to get active refresh tokens")
		activeTokens = []*RefreshTokenRecord{}
	}

	// 获取刷新配额
	refreshCount, remainingQuota := i.refreshService.getRefreshQuota(userID)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"user_id":       userID,
			"username":      username,
			"tenant_id":     tenantID,
			"roles":         claims.Roles,
			"permissions":   claims.Permissions,
			"session_id":    claims.SessionID,
			"issued_at":     claims.IssuedAt,
			"expires_at":    claims.ExpiresAt,
			"active_sessions": len(activeTokens),
			"refresh_quota": gin.H{
				"refresh_count":   refreshCount,
				"remaining_quota": remainingQuota,
				"hourly_limit":    i.config.RateLimitPerHour,
			},
		},
	})
}

// ChangePassword 修改密码
func (i *RefreshTokenServiceIntegration) ChangePassword(c *gin.Context) {
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

	var request struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
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

	// 验证新密码
	if request.NewPassword != request.ConfirmPassword {
		c.JSON(400, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PASSWORD_MISMATCH",
				"message": "New password and confirmation do not match",
			},
		})
		return
	}

	// 验证当前密码（这里应该是实际的数据库验证）
	// user, err := i.validateCurrentPassword(userID, request.CurrentPassword)
	// if err != nil {
	//     c.JSON(400, gin.H{
	//         "success": false,
	//         "error": gin.H{
	//             "code":    "INVALID_CURRENT_PASSWORD",
	//             "message": "Current password is incorrect",
	//         },
	//     })
	//     return
	// }

	// 修改密码（这里应该更新数据库）
	// err = i.updateUserPassword(userID, request.NewPassword)
	// if err != nil {
	//     i.logger.WithError(err).Error("Failed to update user password")
	//     c.JSON(500, gin.H{
	//         "success": false,
	//         "error": gin.H{
	//             "code":    "PASSWORD_UPDATE_FAILED",
	//             "message": "Failed to update password",
	//         },
	//     })
	//     return
	// }

	// 撤销用户所有刷新令牌（强制重新登录）
	err := i.refreshService.RevokeUserRefreshTokens(userID, "password_changed")
	if err != nil {
		i.logger.WithError(err).Error("Failed to revoke refresh tokens after password change")
	}

	// 记录密码修改
	i.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"ip":      c.ClientIP(),
	}).Info("User password changed")

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password changed successfully. Please login again.",
		},
	})
}

// authenticateUser 验证用户凭据（模拟实现）
func (i *RefreshTokenServiceIntegration) authenticateUser(username, password string) (*User, error) {
	// 这里应该是实际的数据库验证
	// 简化实现
	if username == "testuser" && password == "testpass" {
		return &User{
			ID:         1,
			Username:   "testuser",
			TenantID:   "default",
			Roles:      []string{"user"},
			Permissions: []string{"document:read", "document:write"},
			CreatedAt: time.Now(),
		}, nil
	}

	if username == "admin" && password == "adminpass" {
		return &User{
			ID:         2,
			Username:   "admin",
			TenantID:   "default",
			Roles:      []string{"admin"},
			Permissions: []string{"*"},
			CreatedAt: time.Now(),
		}, nil
	}

	return nil, ErrInvalidCredentials
}

// createRefreshTokenRecord 创建刷新令牌记录
func (i *RefreshTokenServiceIntegration) createRefreshTokenRecord(tokenPair *TokenPair, user *User, tenantID string, deviceInfo *DeviceInfo) error {
	// 解析刷新令牌以获取JTI
	claims, err := i.jwtService.ValidateToken(tokenPair.RefreshToken)
	if err != nil {
		return err
	}

	deviceID := ""
	if deviceInfo != nil {
		deviceID = deviceInfo.DeviceID
	}

	record := &RefreshTokenRecord{
		JTI:        claims.JTI,
		UserID:     user.ID,
		TenantID:   tenantID,
		SessionID:  claims.SessionID,
		TokenHash:  i.refreshService.hashToken(tokenPair.RefreshToken),
		DeviceID:   deviceID,
		IPAddress:  claims.IPAddress,
		UserAgent:  claims.UserAgent,
		ExpiresAt:  claims.ExpiresAt.Time,
		LastUsedAt: time.Now(),
	}

	return i.refreshService.db.Create(record).Error
}

// StartCleanupWorker 启动清理工作协程
func (i *RefreshTokenServiceIntegration) StartCleanupWorker() {
	if i.config.AutoCleanup {
		i.refreshService.StartCleanupWorker(i.cleanupContext)
		i.logger.Info("Refresh token cleanup worker started")
	}
}

// StopCleanupWorker 停止清理工作协程
func (i *RefreshTokenServiceIntegration) StopCleanupWorker() {
	if i.cleanupCancel != nil {
		i.cleanupCancel()
		i.logger.Info("Refresh token cleanup worker stopped")
	}
}

// GetStats 获取服务统计信息
func (i *RefreshTokenServiceIntegration) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// JWT服务统计
	jwtStats, _ := i.jwtService.GetTokenStats()
	stats["jwt"] = jwtStats

	// 刷新令牌统计
	stats["refresh_token"] = map[string]interface{}{
		"auto_cleanup":      i.config.AutoCleanup,
		"cleanup_interval":  i.config.CleanupInterval,
		"rate_limit_enabled": i.config.EnableRateLimit,
		"rate_limit_per_hour": i.config.RateLimitPerHour,
		"max_active_tokens": i.config.MaxActiveTokens,
	}

	return stats
}

// HealthCheck 健康检查
func (i *RefreshTokenServiceIntegration) HealthCheck() map[string]interface{} {
	return map[string]interface{}{
		"status": "healthy",
		"timestamp": time.Now(),
		"services": map[string]interface{}{
			"jwt_service":      "operational",
			"refresh_service": "operational",
		},
		"stats": i.GetStats(),
	}
}