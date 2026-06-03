package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/auth"
	"law-oa-go/internal/repositories"
)

var (
	ethicalWallCheckDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "ethical_wall_check_duration_seconds",
		Help:    "Duration of ethical wall permission checks",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
	}, []string{"result"})

	ethicalWallChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ethical_wall_checks_total",
		Help: "Total number of ethical wall checks",
	}, []string{"result"})

	ethicalWallDenialsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ethical_wall_denials_total",
		Help: "Total number of ethical wall access denials",
	}, []string{"case_id"})
)

// EthicalWallConfig 隔离墙中间件配置
type EthicalWallConfig struct {
	EthicalWallRepo repositories.EthicalWallRepository
	SkipPaths       []string
	SkipPrefixes    []string
}

// EthicalWallMiddleware 隔离墙权限检查中间件
//
// 此中间件必须在认证中间件之后执行，因为它需要从上下文中获取用户ID
//
// 执行流程:
//  1. 从请求路径中提取案件ID
//  2. 检查案件是否启用隔离墙
//  3. 如果启用，检查用户是否在白名单中
//  4. 如果不在白名单，记录拒绝日志并返回403
//
// 性能要求: < 10ms (已通过 Prometheus 指标监控)
func EthicalWallMiddleware(config EthicalWallConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 检查是否跳过隔离墙检查
		if shouldSkipEthicalWall(c, config.SkipPaths, config.SkipPrefixes) {
			c.Next()
			return
		}

		// 提取案件ID
		caseID := extractCaseID(c)
		if caseID == 0 {
			// 无法提取案件ID，跳过检查
			c.Next()
			return
		}

		// 检查案件是否启用隔离墙
		enabled, err := config.EthicalWallRepo.IsEthicalWallEnabled(c.Request.Context(), caseID)
		if err != nil {
			// 查询失败，记录错误但允许通过（避免因系统错误导致正常访问受阻）
			ethicalWallCheckDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("error").Inc()
			c.Next()
			return
		}

		if !enabled {
			// 未启用隔离墙，正常通过
			ethicalWallCheckDuration.WithLabelValues("disabled").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("disabled").Inc()
			c.Next()
			return
		}

		// 隔离墙已启用，检查白名单
		userID := auth.GetUserID(c)
		if userID == 0 {
			// 用户未认证，返回401
			ethicalWallCheckDuration.WithLabelValues("unauthenticated").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("unauthenticated").Inc()
			c.JSON(401, gin.H{
				"code":    401,
				"message": "未认证",
				"error":   "Authentication required",
			})
			c.Abort()
			return
		}

		// 检查是否在白名单中
		whitelisted, err := config.EthicalWallRepo.IsUserWhitelisted(c.Request.Context(), caseID, userID)
		if err != nil {
			// 查询失败，记录错误但允许通过
			ethicalWallCheckDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("error").Inc()
			c.Next()
			return
		}

		if !whitelisted {
			// 不在白名单，拒绝访问
			ethicalWallCheckDuration.WithLabelValues("denied").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("denied").Inc()
			ethicalWallDenialsTotal.WithLabelValues(fmt.Sprintf("%d", caseID)).Inc()

			// 记录拒绝日志
			_ = config.EthicalWallRepo.LogAccessAttempt(
				c.Request.Context(),
				caseID,
				userID,
				getAccessType(c),
				"denied",
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)

			c.JSON(403, gin.H{
				"code":    403,
				"message": "该案件启用了隔离墙，您无权访问",
				"error":   "Access denied by ethical wall",
			})
			c.Abort()
			return
		}

		// 在白名单中，记录允许访问并继续
		ethicalWallCheckDuration.WithLabelValues("allowed").Observe(time.Since(start).Seconds())
		ethicalWallChecksTotal.WithLabelValues("allowed").Inc()

		_ = config.EthicalWallRepo.LogAccessAttempt(
			c.Request.Context(),
			caseID,
			userID,
			getAccessType(c),
			"allowed",
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)

		c.Next()
	}
}

// extractCaseID 从请求路径中提取案件ID
// 支持的路径格式:
//   - /api/v1/cases/{id}
//   - /api/v1/cases/{id}/*
//   - /api/v1/documents?case_id={id}
func extractCaseID(c *gin.Context) uint {
	// 从路径参数中提取
	if caseIDParam := c.Param("id"); caseIDParam != "" {
		var id uint
		if _, err := fmt.Sscanf(caseIDParam, "%d", &id); err == nil && id > 0 {
			return id
		}
	}

	// 从查询参数中提取
	if caseIDQuery := c.Query("case_id"); caseIDQuery != "" {
		var id uint
		if _, err := fmt.Sscanf(caseIDQuery, "%d", &id); err == nil && id > 0 {
			return id
		}
	}

	// 从请求体中提取 (对于 POST/PUT 请求)
	if (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") &&
		strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
		// 尝试从 JSON body 中读取 case_id
		type CaseIDRequest struct {
			CaseID uint `json:"case_id"`
		}
		var req CaseIDRequest
		if err := c.ShouldBindJSON(&req); err == nil && req.CaseID > 0 {
			return req.CaseID
		}
	}

	return 0
}

// getAccessType 根据请求方法和路径确定访问类型
func getAccessType(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	// 导出操作
	if strings.Contains(path, "/export") || strings.Contains(path, "/download") {
		return "export"
	}

	// 搜索操作
	if strings.Contains(path, "/search") {
		return "search"
	}

	// 查看操作
	switch method {
	case "GET", "HEAD":
		return "view"
	case "POST", "PUT", "PATCH", "DELETE":
		return "modify"
	default:
		return "unknown"
	}
}

// shouldSkipEthicalWall 检查是否跳过隔离墙检查
func shouldSkipEthicalWall(c *gin.Context, skipPaths, skipPrefixes []string) bool {
	path := c.Request.URL.Path

	// 检查精确匹配
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	// 检查前缀匹配
	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// GetEthicalWallSkipPaths 返回默认的跳过路径列表
func GetEthicalWallSkipPaths() []string {
	return []string{
		"/health",
		"/ping",
		"/api/v1/auth/login",
		"/api/v1/auth/logout",
		"/api/v1/ethical-wall", // 隔离墙管理API本身不受隔离墙限制
	}
}

// GetEthicalWallSkipPrefixes 返回默认的跳过前缀列表
func GetEthicalWallSkipPrefixes() []string {
	return []string{
		"/static",
		"/assets",
		"/favicon",
		"/swagger",
		"/api/v1/notifications", // 通知系统不受隔离墙限制
	}
}

// CaseIDExtractor 案件ID提取器接口
// 允许自定义案件ID提取逻辑
type CaseIDExtractor interface {
	ExtractCaseID(c *gin.Context) uint
}

// DefaultCaseIDExtractor 默认案件ID提取器
type DefaultCaseIDExtractor struct{}

// ExtractCaseID 提取案件ID
func (e *DefaultCaseIDExtractor) ExtractCaseID(c *gin.Context) uint {
	return extractCaseID(c)
}

// NewCaseIDExtractor 创建案件ID提取器
func NewCaseIDExtractor() CaseIDExtractor {
	return &DefaultCaseIDExtractor{}
}
