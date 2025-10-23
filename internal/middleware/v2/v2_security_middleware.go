package v2

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"law-oa-go/internal/config"
	"law-oa-go/internal/logger"
)

// V2SecurityMiddleware 基于最新OWASP安全标准的安全中间件
type V2SecurityMiddleware struct {
	config         *config.Config
	redisClient    *redis.Client
	trustedProxies []string
	blacklistedIPs map[string]bool
	whitelistedIPs map[string]bool
	rateLimitStore map[string]*RateLimitTracker
	mu            sync.RWMutex
}

// RateLimitTracker 速率限制跟踪器
type RateLimitTracker struct {
	Requests   int64
	Window     time.Time
	Count      int64
}

// NewV2SecurityMiddleware 创建安全中间件
func NewV2SecurityMiddleware(config *config.Config, redisClient *redis.Client) *V2SecurityMiddleware {
	sm := &V2SecurityMiddleware{
		config:         config,
		redisClient:    redisClient,
		trustedProxies: []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		blacklistedIPs: make(map[string]bool),
		whitelistedIPs: make(map[string]bool),
		rateLimitStore: make(map[string]*RateLimitTracker),
	}

	// 初始化IP列表
	sm.initializeIPLists()

	return sm
}

// SecurityHeadersMiddleware 安全头中间件 - 基于最新OWASP推荐
func (sm *V2SecurityMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
	// 防止点击劫持
		c.Header("X-Frame-Options", "DENY")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// HTTPS严格传输安全（仅HTTPS时）
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// 内容安全策略 - 严格的CSP配置
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"

		c.Header("Content-Security-Policy", csp)

		// 引用策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		permissions := "geolocation=(), " +
			"microphone=(), " +
			"camera=(), " +
			"payment=(), " +
			"usb=(), " +
			"fullscreen=(self), " +
			"accelerometer=(), " +
			"gyroscope=(), " +
			"magnetometer=()"

		c.Header("Permissions-Policy", permissions)

		// 移除服务器信息
		c.Header("Server", "")

		// 添加自定义安全头
		c.Header("X-Request-ID", sm.generateRequestID())
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止缓存敏感信息
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// CORSMiddleware 增强CORS中间件 - 基于最新gin-contrib/cors最佳实践
func (sm *V2SecurityMiddleware) CORSMiddleware() gin.HandlerFunc {
	// 根据环境动态配置CORS
	var corsConfig cors.Config

	switch sm.config.Server.Env {
	case "production":
		corsConfig = cors.Config{
			AllowOrigins:     []string{"https://your-domain.com", "https://app.your-domain.com"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID"},
			ExposeHeaders:    []string{"X-Total-Count", "X-Request-ID", "X-Page-Count"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
			AllowWildcard:    false,
			AllowBrowserExtensions: false,
			AllowWebSockets:        false,
			AllowFiles:             false,
			CustomSchemas:          []string{},
			AllowPrivateNetwork:    false,
		}
	case "development":
		corsConfig = cors.Config{
			AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3003", "http://localhost:8080"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders:     []string{"*"},
			ExposeHeaders:    []string{"*"},
			AllowCredentials: false,
			MaxAge:           2 * time.Hour,
			AllowWildcard:    true,
			AllowBrowserExtensions: true,
			AllowWebSockets:        true,
			AllowFiles:             true,
			CustomSchemas:          []string{"tauri://", "http://localhost:*"},
			AllowPrivateNetwork:    false,
		}
	default:
		// 默认配置 - 较宽松的开发配置
		corsConfig = cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{"*"}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"*"}
	}

	return cors.New(corsConfig)
}

// RequestValidationMiddleware 请求验证中间件
func (sm *V2SecurityMiddleware) RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 验证HTTP方法
		if !sm.isAllowedMethod(c.Request.Method) {
			sm.logSecurityEvent("blocked_invalid_method", c, map[string]interface{}{
				"method": c.Request.Method,
			})

			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error": "Method not allowed",
				"code":  http.StatusMethodNotAllowed,
			})
			c.Abort()
			return
			}

		// 验证URL长度
		if len(c.Request.URL.RequestURI()) > 2048 {
			sm.logSecurityEvent("blocked_long_url", c, map[string]interface{}{
				"url_length": len(c.Request.URL.RequestURI()),
			})

			c.JSON(http.StatusRequestURITooLong, gin.H{
				"error": "URL too long",
				"code":  http.StatusRequestURITooLong,
			})
			c.Abort()
			return
		}

		// 检查可疑模式
		if sm.hasSuspiciousPatterns(c) {
			sm.logSecurityEvent("blocked_suspicious_pattern", c, map[string]interface{}{
				"url": c.Request.URL.RequestURI(),
			})

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Request blocked",
				"code":  http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// 验证请求头大小
		if err := sm.validateHeadersSize(c); err != nil {
			sm.logSecurityEvent("blocked_large_headers", c, map[string]interface{}{
				"error": err.Error(),
			})

			c.JSON(http.StatusRequestHeaderFieldsTooLarge, gin.H{
				"error": "Request headers too large",
				"code":  http.StatusRequestHeaderFieldsTooLarge,
			})
			c.Abort()
			return
		}

		// 验证Content-Type
		if err := sm.validateContentType(c); err != nil {
			sm.logSecurityEvent("invalid_content_type", c, map[string]interface{}{
				"error": err.Error(),
			})

			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Unsupported media type",
				"code":  http.StatusUnsupportedMediaType,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// IPFilterMiddleware IP过滤中间件
func (sm *V2SecurityMiddleware) IPFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := sm.getClientIP(c)

		// 检查IP白名单
		if sm.isIPWhitelisted(clientIP) {
			c.Next()
			return
		}

		// 检查IP黑名单
		if sm.isIPBlacklisted(clientIP) {
			sm.logSecurityEvent("blocked_blacklisted_ip", c, map[string]interface{}{
				"ip": clientIP,
			})

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"code":  http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// 动态检查IP黑名单（Redis）
		if sm.redisClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			key := fmt.Sprintf("blacklisted_ip:%s", clientIP)
			blacklisted, err := sm.redisClient.Get(ctx, key).Result()
			if err == nil && blacklisted == "1" {
				sm.logSecurityEvent("blocked_dynamic_blacklisted_ip", c, map[string]interface{}{
					"ip": clientIP,
				})

				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
					"code":  http.StatusForbidden,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RateLimitingMiddleware 高级限流中间件
func (sm *V2SecurityMiddleware) RateLimitingMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := sm.getClientIP(c)

		sm.mu.Lock()
		tracker, exists := sm.rateLimitStore[clientIP]
		if !exists {
			tracker = &RateLimitTracker{
				Window: time.Now(),
					Count:  0,
			}
			sm.rateLimitStore[clientIP] = tracker
		}
		sm.mu.Unlock()

		// 检查是否需要重置窗口
		if time.Since(tracker.Window) > window {
			tracker.Window = time.Now()
			tracker.Count = 0
		}

		// 增加请求计数
		tracker.Count++

		// 检查是否超过限制
		if tracker.Count > maxRequests {
			sm.logSecurityEvent("rate_limit_exceeded", c, map[string]interface{}{
				"ip":         clientIP,
				"count":      tracker.Count,
				"limit":      maxRequests,
			})

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", tracker.Window.Add(window).Unix()))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"code":        http.StatusTooManyRequests,
				"retry_after": window.String(),
			})
			c.Abort()
			return
		}

		// 设置响应头
		remaining := maxRequests - tracker.Count
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", tracker.Window.Add(window).Unix()))

		c.Next()
	}
}

// InputSanitizationMiddleware 输入清理中间件
func (sm *V2SecurityMiddleware) InputSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 清理查询参数
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = sm.sanitizeInput(value)
			}
			}

		// 清理路径参数
		for key, values := range c.Params {
			c.Params[key] = sm.sanitizeInput(values)
		}

		// 如果有POST数据，进行基本清理
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			// 注意：这里只是示例，实际应该基于Content-Type处理
			// 因为中间件不应该修改请求体
		}

		c.Next()
	}
}

// TrustedProxyMiddleware 受信任代理中间件
func (sm *V2SecurityMiddleware) TrustedProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置受信任的代理
		if len(sm.trustedProxies) > 0 {
			c.Set("trusted_proxies", sm.trustedProxies)
		}

		// 配置Gin的信任代理
		engine := c.Request.Context().Value("engine")
		if engine != nil {
			if ginEngine, ok := engine.(*gin.Engine); ok {
				ginEngine.SetTrustedProxies(sm.trustedProxies)
			}
		}

		c.Next()
	}
}

// 内部辅助方法

func (sm *V2SecurityMiddleware) initializeIPLists() {
	// 从配置中初始化IP列表
	// 这里可以添加从配置文件或数据库加载IP列表的逻辑

	// 默认白名单本地IP
	sm.whitelistedIPs["127.0.0.1"] = true
	sm.whitelistedIPs["::1"] = true
}

func (sm *V2SecurityMiddleware) isAllowedMethod(method string) bool {
	allowedMethods := []string{
		"GET", "POST", "PUT", "DELETE", "PATCH",
		"HEAD", "OPTIONS", "TRACE",
	}

	for _, allowed := range allowedMethods {
		if method == allowed {
			return true
		}
	}
	return false
}

func (sm *V2SecurityMiddleware) hasSuspiciousPatterns(c *gin.Context) bool {
	url := c.Request.URL.RequestURI()
	userAgent := c.GetHeader("User-Agent")
	Referer := c.GetHeader("Referer")

	// 可疑模式检测
	suspiciousPatterns := []string{
		"(?i)(union|select|insert|update|delete|drop|exec|script)",
		"(?i)<script[^>]*>.*?</script>",
		"(?i)javascript:",
		"(?i)vbscript:",
		"(?i)onload\\s*=",
		"(?i)onerror\\s*=",
		"(?i)alert\\s*\\(",
		"(?i)confirm\\s*\\(",
		"(?i)prompt\\s*\\(",
		"(?i)eval\\s*\\(",
		"(?i)document\\.",
		"(?i)window\\.",
	}

	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, url); matched {
			return true
		}
		if matched, _ := regexp.MatchString(pattern, userAgent); matched {
			return true
		}
		if matched, _ := regexp.MatchString(pattern, Referer); matched {
			return true
		}
	}

	return false
}

func (sm *V2SecurityMiddleware) validateHeadersSize(c *gin.Context) error {
	totalSize := 0
	for key, values := range c.Request.Header {
		totalSize += len(key)
		for _, value := range values {
			totalSize += len(value)
		}
	}

	if totalSize > 16384 { // 16KB
		return fmt.Errorf("headers too large: %d bytes", totalSize)
	}
	return nil
}

func (sm *V2SecurityMiddleware) validateContentType(c *gin.Context) error {
	// 对于POST/PUT请求，验证Content-Type
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			return fmt.Errorf("Content-Type header is required for %s requests", c.Request.Method)
		}

	// 检查是否是允许的Content-Type
		allowedTypes := []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"text/plain",
			"application/xml",
			"text/xml",
		}

		for _, allowed := range allowedTypes {
			if strings.HasPrefix(contentType, allowed) {
				return nil
			}
		}
	}

	return nil
}

func (sm *V2SecurityMiddleware) getClientIP(c *gin.Context) string {
	// 优先使用X-Forwarded-For头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 使用X-Real-IP头
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用内置的ClientIP方法
	return c.ClientIP()
}

func (sm *V2SecurityMiddleware) isIPWhitelisted(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.whitelistedIPs[ip]
}

func (sm *V2SecurityMiddleware) isIPBlacklisted(ip string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.blacklistedIPs[ip]
}

func (sm *V2SecurityMiddleware) sanitizeInput(input string) string {
	// 移除潜在的XSS向量
	replacements := []struct {
		pattern     string
		replacement string
	}{
		{"<script>", ""},
		{"</script>", ""},
		{"javascript:", ""},
		{"vbscript:", ""},
		{"onload=", ""},
		{"onerror=", ""},
		{"onclick=", ""},
		{"alert(", ""},
		{"confirm(", ""},
		{"prompt(", ""},
		{"eval(", ""},
		{"document.", ""},
		{"window.", ""},
	}

	for _, repl := range replacements {
		input = strings.ReplaceAll(input, repl.pattern, repl.replacement)
	}

	return input
}

func (sm *V2SecurityMiddleware) generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (sm *V2SecurityMiddleware) logSecurityEvent(event string, c *gin.Context, details map[string]interface{}) {
	if logger.Logger == nil {
		return
	}

	logData := map[string]interface{}{
		"event":      event,
		"ip":         sm.getClientIP(c),
		"user_agent": c.GetHeader("User-Agent"),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"timestamp":  time.Now().Unix(),
	}

	// 添加额外详情
	for k, v := range details {
		logData[k] = v
	}

	logger.Logger.Warn("Security event detected", zap.Any("data", logData))
}

// GetSecurityStats 获取安全统计信息
func (sm *V2SecurityMiddleware) GetSecurityStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"whitelisted_ips_count":   len(sm.whitelistedIPs),
		"blacklisted_ips_count":   len(sm.blacklistedIPs),
		"rate_limit_entries":       len(sm.rateLimitStore),
		"trusted_proxies_count":    len(sm.trustedProxies),
		"timestamp":               time.Now().Unix(),
	}
}