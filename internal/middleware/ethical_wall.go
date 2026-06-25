package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// DocumentResolver 文档→案件解析接口。
// 实现应通过 DocumentRepository.FindByID 读取文档，避免在中间件中复制 SQL。
// 返回值约定：
//   - applies=false, err=nil: 文档确实不属于案件，跳过隔离墙检查
//   - applies=true:  caseID 是文档关联的案件 ID
//   - err != nil:    解析失败，中间件必须 fail-closed 返回 503
type DocumentResolver interface {
	ResolveDocumentCase(ctx context.Context, documentID uint) (caseID uint, applies bool, err error)
}

// EthicalWallConfig 隔离墙中间件配置
type EthicalWallConfig struct {
	EthicalWallRepo  repositories.EthicalWallRepository
	DocumentResolver DocumentResolver
	SkipPaths        []string
	SkipPrefixes     []string
}

// EthicalWallMiddleware 隔离墙权限检查中间件
//
// 安全策略（fail-closed）：
//   - 仓储/文档解析器返回错误时，必须返回 503 并中止，禁止下游 handler 执行
//   - 文档详情路由（/documents/:id）必须通过 DocumentResolver 解析为 caseID
//   - 文档不属于案件（applies=false）时跳过检查，允许访问
//
// 此中间件必须在认证中间件之后执行，因为它需要从上下文中获取用户ID
func EthicalWallMiddleware(config EthicalWallConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		if shouldSkipEthicalWall(c, config.SkipPaths, config.SkipPrefixes) {
			c.Next()
			return
		}

		caseID, err := resolveCaseIDForEthicalWall(c, config)
		if err != nil {
			ethicalWallCheckDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("error").Inc()
			respondEthicalWallError(c, http.StatusServiceUnavailable, "资源解析失败，已拒绝访问")
			return
		}

		if caseID == 0 {
			c.Next()
			return
		}

		enabled, err := config.EthicalWallRepo.IsEthicalWallEnabled(c.Request.Context(), caseID)
		if err != nil {
			ethicalWallCheckDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("error").Inc()
			respondEthicalWallError(c, http.StatusServiceUnavailable, "隔离墙状态查询失败，已拒绝访问")
			return
		}

		if !enabled {
			ethicalWallCheckDuration.WithLabelValues("disabled").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("disabled").Inc()
			c.Next()
			return
		}

		userID := auth.GetUserID(c)
		if userID == 0 {
			ethicalWallCheckDuration.WithLabelValues("unauthenticated").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("unauthenticated").Inc()
			respondEthicalWallError(c, http.StatusUnauthorized, "未认证")
			return
		}

		whitelisted, err := config.EthicalWallRepo.IsUserWhitelisted(c.Request.Context(), caseID, userID)
		if err != nil {
			ethicalWallCheckDuration.WithLabelValues("error").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("error").Inc()
			respondEthicalWallError(c, http.StatusServiceUnavailable, "白名单查询失败，已拒绝访问")
			return
		}

		if !whitelisted {
			ethicalWallCheckDuration.WithLabelValues("denied").Observe(time.Since(start).Seconds())
			ethicalWallChecksTotal.WithLabelValues("denied").Inc()
			ethicalWallDenialsTotal.WithLabelValues(fmt.Sprintf("%d", caseID)).Inc()

			_ = config.EthicalWallRepo.LogAccessAttempt(
				c.Request.Context(),
				caseID,
				userID,
				getAccessType(c),
				"denied",
				c.ClientIP(),
				c.GetHeader("User-Agent"),
			)

			respondEthicalWallError(c, http.StatusForbidden, "该案件启用了隔离墙，您无权访问")
			return
		}

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

// respondEthicalWallError 统一构造错误响应并中止请求链
func respondEthicalWallError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
	c.Abort()
}

// resolveCaseIDForEthicalWall 解析当前请求关联的 caseID。
//
// 返回值：
//   - caseID > 0, err = nil: 需要进行隔离墙检查
//   - caseID == 0, err = nil: 跳过检查（资源不属于案件或无法判定）
//   - err != nil: 解析失败，调用方必须 fail-closed 返回 503
func resolveCaseIDForEthicalWall(c *gin.Context, config EthicalWallConfig) (uint, error) {
	fullPath := c.FullPath()

	// OnlyOffice 文档入口：通过路径或 body 的 document_id 解析为 caseID
	if isOnlyOfficePath(fullPath) {
		return resolveOnlyOfficeCaseID(c, config)
	}

	// 文档详情路由必须通过 DocumentResolver 解析
	if isDocumentDetailPath(fullPath) {
		if config.DocumentResolver == nil {
			// 向后兼容：未配置 resolver 时跳过（不触发 fail-closed）
			return 0, nil
		}
		docID := parseUintParam(c.Param("id"))
		if docID == 0 {
			return 0, nil
		}
		caseID, applies, err := config.DocumentResolver.ResolveDocumentCase(c.Request.Context(), docID)
		if err != nil {
			return 0, err
		}
		if !applies {
			return 0, nil
		}
		return caseID, nil
	}

	return extractCaseIDForPath(c, fullPath), nil
}

// resolveOnlyOfficeCaseID 解析 OnlyOffice 路由关联的 caseID。
//
// 支持的来源：
//   - 路径参数 :document_id（如 /documents/onlyoffice/:document_id/download/converted/:output_type）
//   - JSON body.document_id（如 /documents/onlyoffice/open|convert），body 会被重置供下游 handler 读取
//   - 无 document_id 可解析（如 /convert/status）时返回 (0, nil)，跳过检查
//
// DocumentResolver 未配置或解析失败时返回 err，触发 fail-closed。
func resolveOnlyOfficeCaseID(c *gin.Context, config EthicalWallConfig) (uint, error) {
	if config.DocumentResolver == nil {
		return 0, nil
	}

	docID := parseUintParam(c.Param("document_id"))
	if docID == 0 && shouldReadBody(c) {
		docID = readDocumentIDFromBodyPreserved(c)
	}
	if docID == 0 {
		return 0, nil
	}

	caseID, applies, err := config.DocumentResolver.ResolveDocumentCase(c.Request.Context(), docID)
	if err != nil {
		return 0, err
	}
	if !applies {
		return 0, nil
	}
	return caseID, nil
}

// readDocumentIDFromBodyPreserved 读取 body 中的 document_id，但保留 body 供下游 handler 重新读取
func readDocumentIDFromBodyPreserved(c *gin.Context) uint {
	if c.Request.Body == nil {
		return 0
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0
	}
	// 重置 body 供下游 handler 读取
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if len(bodyBytes) == 0 {
		return 0
	}
	var probe struct {
		DocumentID uint `json:"document_id"`
	}
	if err := json.Unmarshal(bodyBytes, &probe); err != nil {
		return 0
	}
	return probe.DocumentID
}

// extractCaseIDForPath 根据路由模板提取 caseID，不处理文档详情解析（由 DocumentResolver 负责）。
//
//   - /api/v1/cases/:id 系列 → 提取 :id 作为 caseID
//   - /api/v1/documents/:id → 返回 0（由 DocumentResolver 处理）
//   - /api/v1/documents/stats/:section → 返回 0（stats 不是案件资源）
//   - 其他路径 → 从查询参数 case_id 或 JSON body.case_id 提取
//
// 对 POST/PUT/PATCH 的 body 读取会保留原始 body 供下游 handler 使用。
func extractCaseIDForPath(c *gin.Context, fullPath string) uint {
	if isDocumentDetailPath(fullPath) || isDocumentStatsPath(fullPath) {
		return 0
	}

	if isCasesDetailPath(fullPath) {
		if id := parseUintParam(c.Param("id")); id > 0 {
			return id
		}
	}

	if id := parseUintParam(c.Query("case_id")); id > 0 {
		return id
	}

	if shouldReadBody(c) {
		if id := readCaseIDFromBodyPreserved(c); id > 0 {
			return id
		}
	}

	if shouldReadMultipart(c) {
		if id := readCaseIDFromMultipart(c); id > 0 {
			return id
		}
	}

	return 0
}

// isCasesDetailPath 判断路径模板是否是 case 详情路由
func isCasesDetailPath(fullPath string) bool {
	return strings.Contains(fullPath, "/cases/:id")
}

// isDocumentDetailPath 判断路径模板是否是文档详情路由（含 :id 参数）
func isDocumentDetailPath(fullPath string) bool {
	return strings.Contains(fullPath, "/documents/:id")
}

// isDocumentStatsPath 判断路径模板是否是文档统计路由（非案件资源）
func isDocumentStatsPath(fullPath string) bool {
	return strings.Contains(fullPath, "/documents/stats")
}

// isOnlyOfficePath 判断路径模板是否是 OnlyOffice 文档入口
// 包括 /documents/onlyoffice/open、/convert、/convert/status、/:document_id/download/...
func isOnlyOfficePath(fullPath string) bool {
	return strings.Contains(fullPath, "/documents/onlyoffice")
}

// parseUintParam 将字符串安全解析为 uint，非法或 0 返回 0
func parseUintParam(raw string) uint {
	if raw == "" {
		return 0
	}
	var id uint
	if _, err := fmt.Sscanf(raw, "%d", &id); err == nil && id > 0 {
		return id
	}
	return 0
}

// shouldReadBody 判断是否需要从 body 中读取 case_id
func shouldReadBody(c *gin.Context) bool {
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
		return false
	}
	return strings.HasPrefix(c.GetHeader("Content-Type"), "application/json")
}

func shouldReadMultipart(c *gin.Context) bool {
	if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
		return false
	}
	return strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data")
}

func readCaseIDFromMultipart(c *gin.Context) uint {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return 0
	}
	entityType := strings.TrimSpace(c.Request.FormValue("entity_type"))
	if entityType != "case" {
		return 0
	}
	return parseUintParam(c.Request.FormValue("entity_id"))
}

// readCaseIDFromBodyPreserved 读取 body 中的 case_id，但保留 body 供下游 handler 重新读取
func readCaseIDFromBodyPreserved(c *gin.Context) uint {
	if c.Request.Body == nil {
		return 0
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return 0
	}
	// 重置 body 供下游 handler 读取
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if len(bodyBytes) == 0 {
		return 0
	}
	var probe struct {
		CaseID uint `json:"case_id"`
	}
	if err := json.Unmarshal(bodyBytes, &probe); err != nil {
		return 0
	}
	return probe.CaseID
}

// getAccessType 根据请求方法和路径确定访问类型
func getAccessType(c *gin.Context) string {
	method := c.Request.Method
	path := c.Request.URL.Path

	if strings.Contains(path, "/export") || strings.Contains(path, "/download") {
		return "export"
	}

	if strings.Contains(path, "/search") {
		return "search"
	}

	switch method {
	case http.MethodGet, http.MethodHead:
		return "view"
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return "modify"
	default:
		return "unknown"
	}
}

// shouldSkipEthicalWall 检查是否跳过隔离墙检查
//
// 支持两种匹配方式：
//   - 精确匹配实际请求路径
//   - 匹配注册的路由模板（SkipPaths 可含 :param 占位符）
func shouldSkipEthicalWall(c *gin.Context, skipPaths, skipPrefixes []string) bool {
	path := c.Request.URL.Path
	fullPath := c.FullPath()

	for _, skipPath := range skipPaths {
		if path == skipPath || fullPath == skipPath {
			return true
		}
	}

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

// CaseIDExtractor 案件ID提取器接口（向后兼容）
type CaseIDExtractor interface {
	ExtractCaseID(c *gin.Context) uint
}

// DefaultCaseIDExtractor 默认案件ID提取器
type DefaultCaseIDExtractor struct{}

// ExtractCaseID 提取案件ID
func (e *DefaultCaseIDExtractor) ExtractCaseID(c *gin.Context) uint {
	return extractCaseIDForPath(c, c.FullPath())
}

// NewCaseIDExtractor 创建案件ID提取器
func NewCaseIDExtractor() CaseIDExtractor {
	return &DefaultCaseIDExtractor{}
}
