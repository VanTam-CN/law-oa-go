package auth

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// RefreshTokenHandler 刷新令牌HTTP处理器
type RefreshTokenHandler struct {
	service *RefreshTokenService
	logger  *logrus.Logger
}

// NewRefreshTokenHandler 创建刷新令牌处理器
func NewRefreshTokenHandler(service *RefreshTokenService, logger *logrus.Logger) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		service: service,
		logger:  logger,
	}
}

// RefreshToken 刷新令牌端点
func (h *RefreshTokenHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	// 绑定请求数据
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		}).Warn("Invalid refresh token request")

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
				"details": err.Error(),
			},
		})
		return
	}

	// 获取客户端信息
	if req.ClientID == "" {
		req.ClientID = c.GetHeader("X-Client-ID")
	}
	if req.ClientSecret == "" {
		req.ClientSecret = c.GetHeader("X-Client-Secret")
	}

	// 获取设备信息
	req.DeviceInfo = h.extractDeviceInfo(c)

	// 执行令牌刷新
	result, err := h.service.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"ip":    c.ClientIP(),
		}).Warn("Token refresh failed")

		h.handleRefreshError(c, err, result)
		return
	}

	// 记录成功日志
	h.logger.WithFields(logrus.Fields{
		"user_id":      result.Response.UserInfo.UserID,
		"username":     result.Response.UserInfo.Username,
		"ip":           c.ClientIP(),
		"revoked":      result.TokenRevoked,
		"refresh_count": result.RefreshCount,
	}).Info("Token refresh successful")

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result.Response,
		"meta": gin.H{
			"timestamp":        result.Response.RefreshedAt,
			"token_revoked":    result.TokenRevoked,
			"refresh_count":    result.RefreshCount,
			"remaining_quota":  result.RemainingQuota,
		},
	})
}

// GetRefreshHistory 获取刷新历史端点
func (h *RefreshTokenHandler) GetRefreshHistory(c *gin.Context) {
	// 获取用户信息
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 解析查询参数
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50
	}

	// 获取刷新历史
	history, err := h.service.GetRefreshHistory(userID, limit)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to get refresh history")

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve refresh history",
			},
		})
		return
	}

	// 转换为响应格式
	historyItems := make([]gin.H, len(history))
	for i, item := range history {
		historyItems[i] = gin.H{
			"id":             item.ID,
			"refresh_reason": item.RefreshReason,
			"ip_address":     item.IPAddress,
			"user_agent":     item.UserAgent,
			"success":        item.Success,
			"error_message":  item.ErrorMessage,
			"created_at":     item.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"history": historyItems,
			"total":   len(historyItems),
		},
	})
}

// GetActiveRefreshTokens 获取活跃刷新令牌端点
func (h *RefreshTokenHandler) GetActiveRefreshTokens(c *gin.Context) {
	// 获取用户信息
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 获取活跃刷新令牌
	tokens, err := h.service.GetUserRefreshTokens(userID)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to get active refresh tokens")

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve active tokens",
			},
		})
		return
	}

	// 转换为响应格式
	tokenItems := make([]gin.H, len(tokens))
	for i, token := range tokens {
		tokenItems[i] = gin.H{
			"id":            token.ID,
			"device_id":     token.DeviceID,
			"ip_address":    token.IPAddress,
			"user_agent":    token.UserAgent,
			"created_at":    token.CreatedAt,
			"expires_at":    token.ExpiresAt,
			"last_used_at":  token.LastUsedAt,
			"revoked_at":    token.RevokedAt,
			"revoked_reason": token.RevokedReason,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"tokens": tokenItems,
			"total":  len(tokenItems),
		},
	})
}

// RevokeRefreshToken 撤销刷新令牌端点
func (h *RefreshTokenHandler) RevokeRefreshToken(c *gin.Context) {
	// 获取用户信息
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 获取请求参数
	var req struct {
		TokenJTI string `json:"token_jti" binding:"required"`
		Reason   string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Invalid request format",
			},
		})
		return
	}

	// 设置默认撤销原因
	if req.Reason == "" {
		req.Reason = "user_revoked"
	}

	// 验证令牌属于当前用户
	record, err := h.service.getRefreshTokenRecord(req.TokenJTI)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_NOT_FOUND",
				"message": "Refresh token not found",
			},
		})
		return
	}

	if record.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Token does not belong to current user",
			},
		})
		return
	}

	// 撤销令牌
	err = h.service.revokeRefreshToken(req.TokenJTI, req.Reason)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
			"token_jti": req.TokenJTI,
		}).Error("Failed to revoke refresh token")

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to revoke token",
			},
		})
		return
	}

	// 记录成功日志
	h.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"token_jti": req.TokenJTI,
		"reason":   req.Reason,
		"ip":       c.ClientIP(),
	}).Info("Refresh token revoked")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Token revoked successfully",
		},
	})
}

// RevokeAllRefreshTokens 撤销所有刷新令牌端点
func (h *RefreshTokenHandler) RevokeAllRefreshTokens(c *gin.Context) {
	// 获取用户信息
	userID, exists := GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// 获取请求参数
	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "user_revoked_all"
	}

	// 设置默认撤销原因
	if req.Reason == "" {
		req.Reason = "user_revoked_all"
	}

	// 撤销用户所有刷新令牌
	err := h.service.RevokeUserRefreshTokens(userID, req.Reason)
	if err != nil {
		h.logger.WithFields(logrus.Fields{
			"error":   err.Error(),
			"user_id": userID,
		}).Error("Failed to revoke all refresh tokens")

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to revoke tokens",
			},
		})
		return
	}

	// 记录成功日志
	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"reason":  req.Reason,
		"ip":      c.ClientIP(),
	}).Info("All refresh tokens revoked")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "All tokens revoked successfully",
		},
	})
}

// extractDeviceInfo 提取设备信息
func (h *RefreshTokenHandler) extractDeviceInfo(c *gin.Context) *DeviceInfo {
	deviceInfo := &DeviceInfo{}

	// 从Header提取设备信息
	deviceInfo.DeviceID = c.GetHeader("X-Device-ID")
	deviceInfo.DeviceType = c.GetHeader("X-Device-Type")
	deviceInfo.Platform = c.GetHeader("X-Platform")
	deviceInfo.Version = c.GetHeader("X-App-Version")
	deviceInfo.AppVersion = c.GetHeader("X-App-Version")

	// 如果没有设备ID，尝试从User-Agent生成
	if deviceInfo.DeviceID == "" {
		userAgent := c.GetHeader("User-Agent")
		if userAgent != "" {
			deviceInfo.DeviceID = h.hashUserAgent(userAgent)
		}
	}

	// 如果没有设备类型，尝试推断
	if deviceInfo.DeviceType == "" {
		deviceInfo.DeviceType = h.inferDeviceType(c.GetHeader("User-Agent"))
	}

	return deviceInfo
}

// hashUserAgent 哈希用户代理
func (h *RefreshTokenHandler) hashUserAgent(userAgent string) string {
	// 简化实现，实际应该使用更安全的哈希算法
	return "device_" + strings.ToLower(userAgent[:10])
}

// inferDeviceType 推断设备类型
func (h *RefreshTokenHandler) inferDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		return "mobile"
	}
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}
	if strings.Contains(ua, "windows") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "linux") {
		return "desktop"
	}

	return "unknown"
}

// handleRefreshError 处理刷新错误
func (h *RefreshTokenHandler) handleRefreshError(c *gin.Context, err error, result *RefreshResult) {
	statusCode := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	message := "Internal server error"

	switch err {
	case ErrTokenMissing:
		statusCode = http.StatusBadRequest
		errorCode = "TOKEN_MISSING"
		message = "Refresh token is required"
	case ErrTokenMalformed:
		statusCode = http.StatusBadRequest
		errorCode = "TOKEN_MALFORMED"
		message = "Refresh token format is invalid"
	case ErrRefreshTokenExpired:
		statusCode = http.StatusUnauthorized
		errorCode = "REFRESH_TOKEN_EXPIRED"
		message = "Refresh token has expired"
	case ErrRefreshTokenUsed:
		statusCode = http.StatusUnauthorized
		errorCode = "REFRESH_TOKEN_USED"
		message = "Refresh token has already been used"
	case ErrRefreshTokenInvalid:
		statusCode = http.StatusUnauthorized
		errorCode = "REFRESH_TOKEN_INVALID"
		message = "Refresh token is invalid"
	case ErrDeviceMismatch:
		statusCode = http.StatusUnauthorized
		errorCode = "DEVICE_MISMATCH"
		message = "Device verification failed"
	case ErrUserAgentMismatch:
		statusCode = http.StatusUnauthorized
		errorCode = "USER_AGENT_MISMATCH"
		message = "User agent verification failed"
	case ErrRateLimitExceeded:
		statusCode = http.StatusTooManyRequests
		errorCode = "RATE_LIMIT_EXCEEDED"
		message = "Too many refresh requests, please try again later"
	}

	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    errorCode,
			"message": message,
		},
	}

	// 如果有额外的结果信息，添加到响应中
	if result != nil {
		response["meta"] = gin.H{
			"token_revoked": result.TokenRevoked,
		}
	}

	c.JSON(statusCode, response)
}

// RegisterRoutes 注册路由
func (h *RefreshTokenHandler) RegisterRoutes(router *gin.RouterGroup, middleware ...gin.HandlerFunc) {
	group := router.Group("/refresh-token")
	if len(middleware) > 0 {
		group.Use(middleware...)
	}

	{
		group.POST("/refresh", h.RefreshToken)
		group.GET("/history", h.GetRefreshHistory)
		group.GET("/active", h.GetActiveRefreshTokens)
		group.DELETE("/revoke", h.RevokeRefreshToken)
		group.DELETE("/revoke-all", h.RevokeAllRefreshTokens)
	}
}

// RegisterAdminRoutes 注册管理员路由
func (h *RefreshTokenHandler) RegisterAdminRoutes(router *gin.RouterGroup, adminMiddleware ...gin.HandlerFunc) {
	group := router.Group("/admin/refresh-token")
	if len(adminMiddleware) > 0 {
		group.Use(adminMiddleware...)
	}

	{
		group.GET("/cleanup", h.CleanupExpiredTokens)
		group.GET("/stats", h.GetRefreshTokenStats)
		group.POST("/revoke-user/:user_id", h.RevokeUserTokens)
	}
}

// CleanupExpiredTokens 清理过期令牌端点（管理员）
func (h *RefreshTokenHandler) CleanupExpiredTokens(c *gin.Context) {
	err := h.service.CleanupExpiredTokens()
	if err != nil {
		h.logger.WithError(err).Error("Failed to cleanup expired tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "CLEANUP_FAILED",
				"message": "Failed to cleanup expired tokens",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Expired tokens cleaned up successfully",
		},
	})
}

// GetRefreshTokenStats 获取刷新令牌统计信息端点（管理员）
func (h *RefreshTokenHandler) GetRefreshTokenStats(c *gin.Context) {
	// 这里应该实现统计信息查询
	// 简化实现
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_tokens":        0,
			"active_tokens":       0,
			"expired_tokens":      0,
			"revoked_tokens":      0,
			"last_cleanup_time":   nil,
			"message": "Stats endpoint not implemented yet",
		},
	})
}

// RevokeUserTokens 撤销用户令牌端点（管理员）
func (h *RefreshTokenHandler) RevokeUserTokens(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_USER_ID",
				"message": "Invalid user ID",
			},
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Reason = "admin_revoked"
	}

	err = h.service.RevokeUserRefreshTokens(uint(userID), req.Reason)
	if err != nil {
		h.logger.WithError(err).Error("Failed to revoke user tokens")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "REVOKE_FAILED",
				"message": "Failed to revoke user tokens",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "User tokens revoked successfully",
		},
	})
}