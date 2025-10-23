package auth

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SecurityMiddleware JWT安全中间件
type SecurityMiddleware struct {
	securityService *SecurityService
	logger          *logrus.Logger
	config          *SecurityConfig
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware(db *gorm.DB, logger *logrus.Logger, config *SecurityConfig) *SecurityMiddleware {
	securityService := NewSecurityService(db, logger, config)

	return &SecurityMiddleware{
		securityService: securityService,
		logger:          logger,
		config:          config,
	}
}

// Middleware 返回安全检查中间件
func (m *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求信息
		req := m.buildSecurityRequest(c)

		// 执行安全检查
		result := m.securityService.ValidateRequest(req)

		// 添加安全响应头
		m.addSecurityHeaders(c, result)

		// 如果请求被阻止
		if !result.Allowed {
			m.logger.WithFields(logrus.Fields{
				"ip":         req.IPAddress,
				"user_id":    req.UserID,
				"risk_score": result.RiskScore,
				"block_reason": result.BlockReason,
				"path":       c.Request.URL.Path,
			}).Warn("Request blocked by security middleware")

			// 返回安全响应
			m.handleSecurityBlock(c, result)
			c.Abort()
			return
		}

		// 如果有威胁但允许通过，添加警告信息
		if len(result.Threats) > 0 {
			c.Set("security_threats", result.Threats)
			c.Set("risk_score", result.RiskScore)
			c.Set("security_recommendations", result.Recommendations)

			// 记录安全警告
			m.logger.WithFields(logrus.Fields{
				"ip":         req.IPAddress,
				"user_id":    req.UserID,
				"risk_score": result.RiskScore,
				"threat_count": len(result.Threats),
				"path":       c.Request.URL.Path,
			}).Warn("Security threats detected but request allowed")
		}

		// 将安全结果添加到上下文
		c.Set("security_result", result)
		c.Set("risk_score", result.RiskScore)

		c.Next()
	}
}

// buildSecurityRequest 构建安全检查请求
func (m *SecurityMiddleware) buildSecurityRequest(c *gin.Context) *SecurityCheckRequest {
	req := &SecurityCheckRequest{
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		RequestPath: c.Request.URL.Path,
		HTTPMethod:  c.Request.Method,
		Timestamp:   time.Now(),
	}

	// 尝试从JWT令牌获取用户信息
	if claims, exists := GetClaims(c); exists {
		req.UserID = claims.UserID
		req.Username = claims.Username
		req.TenantID = claims.TenantID
		req.SessionID = claims.SessionID
		req.TokenJTI = claims.JTI
		req.DeviceID = claims.DeviceID
	}

	// 从Header获取设备信息
	if req.DeviceID == "" {
		req.DeviceID = c.GetHeader("X-Device-ID")
	}

	return req
}

// addSecurityHeaders 添加安全响应头
func (m *SecurityMiddleware) addSecurityHeaders(c *gin.Context, result *SecurityCheckResult) {
	// 添加安全评分头
	c.Header("X-Risk-Score", strconv.FormatFloat(result.RiskScore, 'f', 2, 64))

	// 添加威胁数量头
	c.Header("X-Threat-Count", strconv.Itoa(len(result.Threats)))

	// 添加安全时间戳头
	c.Header("X-Security-Timestamp", time.Now().Format(time.RFC3339))

	// 添加安全检查状态头
	if result.Allowed {
		c.Header("X-Security-Status", "allowed")
	} else {
		c.Header("X-Security-Status", "blocked")
	}

	// 添加CSP头（内容安全策略）
	c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")

	// 添加其他安全头
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
}

// handleSecurityBlock 处理安全阻止
func (m *SecurityMiddleware) handleSecurityBlock(c *gin.Context, result *SecurityCheckResult) {
	// 根据风险评分和威胁类型决定响应状态码
	statusCode := http.StatusForbidden
	errorCode := "SECURITY_BLOCK"
	message := "Request blocked due to security policy"

	if result.RiskScore >= 0.9 {
		statusCode = http.StatusTooManyRequests
		errorCode = "HIGH_RISK_BLOCK"
		message = "Request blocked due to high security risk"
	}

	// 根据主要威胁类型调整响应
	if len(result.Threats) > 0 {
		mainThreat := result.Threats[0]
		switch mainThreat.Type {
		case ThreatTypeRateLimit:
			statusCode = http.StatusTooManyRequests
			errorCode = "RATE_LIMITED"
			message = "Rate limit exceeded"
		case ThreatTypeIPBlacklist:
			statusCode = http.StatusForbidden
			errorCode = "IP_BLOCKED"
			message = "IP address blocked"
		case ThreatTypeDeviceMismatch:
			statusCode = http.StatusUnauthorized
			errorCode = "DEVICE_BLOCKED"
			message = "Device verification failed"
		}
	}

	c.JSON(statusCode, gin.H{
		"success": false,
		"error": gin.H{
			"code":    errorCode,
			"message": message,
			"details": result.BlockReason,
		},
		"meta": gin.H{
			"risk_score":   result.RiskScore,
			"threat_count": len(result.Threats),
			"timestamp":    time.Now().Format(time.RFC3339),
		},
	})
}

// RequireMaxRiskScore 要求最大风险评分
func (m *SecurityMiddleware) RequireMaxRiskScore(maxScore float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		riskScore, exists := c.Get("risk_score")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Security check not performed",
			})
			c.Abort()
			return
		}

		score, ok := riskScore.(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid risk score format",
			})
			c.Abort()
			return
		}

		if score > maxScore {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "RISK_SCORE_TOO_HIGH",
					"message": "Risk score exceeds allowed threshold",
					"details": gin.H{
						"current_score": score,
						"max_allowed":  maxScore,
					},
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireNoThreats 要求无威胁
func (m *SecurityMiddleware) RequireNoThreats(threatTypes ...ThreatType) gin.HandlerFunc {
	return func(c *gin.Context) {
		threats, exists := c.Get("security_threats")
		if !exists {
			c.Next()
			return
		}

		securityThreats, ok := threats.([]SecurityThreat)
		if !ok {
			c.Next()
			return
		}

		// 如果没有指定威胁类型，检查所有威胁
		if len(threatTypes) == 0 && len(securityThreats) > 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "THREATS_DETECTED",
					"message": "Security threats detected",
					"details": securityThreats,
				},
			})
			c.Abort()
			return
		}

		// 检查指定的威胁类型
		for _, threatType := range threatTypes {
			for _, threat := range securityThreats {
				if threat.Type == threatType {
					c.JSON(http.StatusForbidden, gin.H{
						"error": gin.H{
							"code":    "THREAT_TYPE_DETECTED",
							"message": "Specific security threat type detected",
							"details": gin.H{
								"threat_type": threatType,
								"threat":      threat,
							},
						},
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// SecurityInfo 安全信息中间件
func (m *SecurityMiddleware) SecurityInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		result, exists := c.Get("security_result")
		if !exists {
			c.JSON(http.StatusOK, gin.H{
				"message": "No security check performed",
			})
			return
		}

		securityResult, ok := result.(*SecurityCheckResult)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid security result format",
			})
			return
		}

		// 构建安全信息响应
		response := gin.H{
			"allowed":     securityResult.Allowed,
			"risk_score":  securityResult.RiskScore,
			"threat_count": len(securityResult.Threats),
			"timestamp":   time.Now().Format(time.RFC3339),
		}

		// 添加威胁信息（简化版）
		if len(securityResult.Threats) > 0 {
			threatSummaries := make([]gin.H, len(securityResult.Threats))
			for i, threat := range securityResult.Threats {
				threatSummaries[i] = gin.H{
					"type":        threat.Type,
					"severity":    threat.Severity,
					"description": threat.Description,
					"confidence":  threat.Confidence,
				}
			}
			response["threats"] = threatSummaries
		}

		// 添加建议信息
		if len(securityResult.Recommendations) > 0 {
			recommendations := make([]gin.H, len(securityResult.Recommendations))
			for i, rec := range securityResult.Recommendations {
				recommendations[i] = gin.H{
					"type":        rec.Type,
					"action":      rec.Action,
					"description": rec.Description,
					"priority":    rec.Priority,
				}
			}
			response["recommendations"] = recommendations
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    response,
		})
	}
}

// SecurityReport 安全报告中间件（管理员用）
func (m *SecurityMiddleware) SecurityReport() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查管理员权限
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			return
		}

		roles, _ := c.Get("roles")
		userRoles, ok := roles.([]string)
		if !ok || !m.hasRole(userRoles, "admin") {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
			})
			return
		}

		// 生成安全报告
		report := m.generateSecurityReport()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    report,
		})
	}
}

// hasRole 检查是否有指定角色
func (m *SecurityMiddleware) hasRole(roles []string, targetRole string) bool {
	for _, role := range roles {
		if role == targetRole {
			return true
		}
	}
	return false
}

// generateSecurityReport 生成安全报告
func (m *SecurityMiddleware) generateSecurityReport() gin.H {
	// 这里应该从数据库或监控系统获取实际的统计数据
	// 简化实现，返回模拟数据

	return gin.H{
		"report_time": time.Now().Format(time.RFC3339),
		"period": "24h",
		"statistics": gin.H{
			"total_requests":       10000,
			"blocked_requests":     150,
			"high_risk_requests":   45,
			"unique_ips":          500,
			"blocked_ips":         25,
			"threats_detected": gin.H{
				"ip_blacklist":        50,
				"rate_limit":          30,
				"anomalous_behavior":  20,
				"device_mismatch":     15,
				"token_reuse":         10,
			},
		},
		"top_threats": []gin.H{
			{
				"type":        "ip_blacklist",
				"count":       50,
				"severity":    "high",
				"description": "Requests from blacklisted IP addresses",
			},
			{
				"type":        "rate_limit",
				"count":       30,
				"severity":    "medium",
				"description": "Requests exceeding rate limits",
			},
		},
		"recommendations": []string{
			"Consider tightening IP blacklist rules",
			"Review rate limiting thresholds",
			"Monitor anomalous behavior patterns",
		},
	}
}

// GetSecurityService 获取安全服务
func (m *SecurityMiddleware) GetSecurityService() *SecurityService {
	return m.securityService
}

// GetSecurityConfig 获取安全配置
func (m *SecurityMiddleware) GetSecurityConfig() *SecurityConfig {
	return m.config
}

// UpdateSecurityConfig 更新安全配置
func (m *SecurityMiddleware) UpdateSecurityConfig(newConfig *SecurityConfig) {
	m.config = newConfig
	// 这里可能需要重新初始化相关组件
	m.logger.Info("Security configuration updated")
}

// SecurityMetrics 安全指标
type SecurityMetrics struct {
	TotalRequests    int64   `json:"total_requests"`
	BlockedRequests int64   `json:"blocked_requests"`
	HighRiskRequests int64  `json:"high_risk_requests"`
	AverageRiskScore float64 `json:"average_risk_score"`
	TopThreats      []gin.H `json:"top_threats"`
	LastUpdateTime   time.Time `json:"last_update_time"`
}

// GetMetrics 获取安全指标
func (m *SecurityMiddleware) GetMetrics() *SecurityMetrics {
	// 这里应该从实际的监控系统获取数据
	// 简化实现

	return &SecurityMetrics{
		TotalRequests:     10000,
		BlockedRequests:   150,
		HighRiskRequests:  45,
		AverageRiskScore:  0.15,
		TopThreats: []gin.H{
			{"type": "ip_blacklist", "count": 50},
			{"type": "rate_limit", "count": 30},
		},
		LastUpdateTime: time.Now(),
	}
}

// RegisterSecurityRoutes 注册安全相关路由
func (m *SecurityMiddleware) RegisterSecurityRoutes(router *gin.RouterGroup, jwtMiddleware gin.HandlerFunc) {
	security := router.Group("/security")
	security.Use(jwtMiddleware)
	{
		security.GET("/info", m.SecurityInfo())
		security.GET("/metrics", m.RequireRole("admin"), m.getMetrics())
		security.POST("/config", m.RequireRole("admin"), m.updateConfig())
		security.GET("/report", m.RequireRole("admin"), m.SecurityReport())
	}
}

// getMetrics 获取安全指标
func (m *SecurityMiddleware) getMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := m.GetMetrics()
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    metrics,
		})
	}
}

// updateConfig 更新配置
func (m *SecurityMiddleware) updateConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newConfig SecurityConfig
		if err := c.ShouldBindJSON(&newConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":  "Invalid configuration format",
			})
			return
		}

		m.UpdateSecurityConfig(&newConfig)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": "Security configuration updated successfully",
			},
		})
	}
}

// RequireRole 要求指定角色
func (m *SecurityMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User roles not found",
			})
			c.Abort()
			return
		}

		userRoles, ok := roles.([]string)
		if !ok || !m.hasRole(userRoles, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient privileges",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}