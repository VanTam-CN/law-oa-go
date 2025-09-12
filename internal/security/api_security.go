package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"law-oa-go/internal/cache"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	apiSecurityDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "api_security_duration_seconds",
		Help:    "Duration of API security operations",
		Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
	}, []string{"operation"})
	
	apiSecurityErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_security_errors_total",
		Help: "Total number of API security errors",
	}, []string{"operation", "type"})
	
	throttledRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_throttled_requests_total",
		Help: "Total number of throttled API requests",
	})
	
	blockedRequests = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_blocked_requests_total",
		Help: "Total number of blocked API requests",
	})
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableRateLimit      bool
	EnableIPWhitelist    bool
	EnableIPBlacklist    bool
	EnableRequestSigning bool
	EnableAPIThrottling  bool
	EnableWAFProtection  bool
	EnableDDoSProtection bool
	EnableRequestValidation bool
	EnableCORS           bool
	EnableCSRF           bool
	
	RateLimitWindow      time.Duration
	RateLimitMaxRequests int
	
	WhitelistedIPs      []string
	BlacklistedIPs      []string
	
	APIThrottlingConfig map[string]ThrottlingConfig
	
	WAFRules []WAFRule
	
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials  bool
	MaxAge             int
}

// ThrottlingConfig API限流配置
type ThrottlingConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	BurstSize         int
}

// WAFRule Web应用防火墙规则
type WAFRule struct {
	ID          string
	Name        string
	Description string
	Type        string // "sql_injection", "xss", "path_traversal", "command_injection", "file_inclusion"
	Pattern     string
	Action      string // "block", "log", "allow"
	Severity    string // "low", "medium", "high", "critical"
	Enabled     bool
}

// SecurityMetrics 安全指标
type SecurityMetrics struct {
	TotalRequests      int64
	BlockedRequests   int64
	ThrottledRequests int64
	FailedAuths        int64
	SuspiciousIPs      map[string]int
	AttackAttempts     map[string]int
}

// APISecurityService API安全服务
type APISecurityService struct {
	config        *SecurityConfig
	cacheService  *cache.CacheService
	metrics       *SecurityMetrics
	blockedIPs    map[string]time.Time
	throttledIPs  map[string][]time.Time
}

// NewAPISecurityService 创建API安全服务
func NewAPISecurityService(config *SecurityConfig, cacheService *cache.CacheService) *APISecurityService {
	return &APISecurityService{
		config:        config,
		cacheService:  cacheService,
		metrics:       &SecurityMetrics{SuspiciousIPs: make(map[string]int), AttackAttempts: make(map[string]int)},
		blockedIPs:    make(map[string]time.Time),
		throttledIPs:  make(map[string][]time.Time),
	}
}

// RateLimiterMiddleware 速率限制中间件
func (s *APISecurityService) RateLimiterMiddleware() gin.HandlerFunc {
	if !s.config.EnableRateLimit {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("rate_limit").Observe(time.Since(start).Seconds())
		}()
		
		ip := c.ClientIP()
		key := fmt.Sprintf("rate_limit:%s", ip)
		
		// 使用滑动窗口算法进行速率限制
		now := time.Now()
		windowStart := now.Add(-s.config.RateLimitWindow)
		
		var timestamps []time.Time
		err := s.cacheService.Get(c.Request.Context(), key, &timestamps)
		if err != nil {
			timestamps = []time.Time{}
		}
		
		// 清理过期的请求记录
		validTimestamps := []time.Time{}
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		
		// 检查是否超过限制
		if len(validTimestamps) >= s.config.RateLimitMaxRequests {
			throttledRequests.Inc()
			s.metrics.ThrottledRequests++
			s.metrics.SuspiciousIPs[ip]++
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁",
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}
		
		// 记录当前请求
		validTimestamps = append(validTimestamps, now)
		s.cacheService.Set(c.Request.Context(), key, validTimestamps, s.config.RateLimitWindow)
		
		c.Next()
	}
}

// IPWhitelistMiddleware IP白名单中间件
func (s *APISecurityService) IPWhitelistMiddleware() gin.HandlerFunc {
	if !s.config.EnableIPWhitelist {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("ip_whitelist").Observe(time.Since(start).Seconds())
		}()
		
		ip := c.ClientIP()
		
		// 检查IP是否在白名单中
		if !s.isIPWhitelisted(ip) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			s.metrics.SuspiciousIPs[ip]++
			
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "IP地址不在允许列表中",
				"error":   "IP not whitelisted",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// IPBlacklistMiddleware IP黑名单中间件
func (s *APISecurityService) IPBlacklistMiddleware() gin.HandlerFunc {
	if !s.config.EnableIPBlacklist {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("ip_blacklist").Observe(time.Since(start).Seconds())
		}()
		
		ip := c.ClientIP()
		
		// 检查IP是否在黑名单中
		if s.isIPBlacklisted(ip) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			s.metrics.SuspiciousIPs[ip]++
			
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "IP地址已被封禁",
				"error":   "IP blacklisted",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// WAFMiddleware Web应用防火墙中间件
func (s *APISecurityService) WAFMiddleware() gin.HandlerFunc {
	if !s.config.EnableWAFProtection {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("waf").Observe(time.Since(start).Seconds())
		}()
		
		// 检查URL
		if s.detectWAFThreat(c.Request.URL.String(), "url") {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			s.metrics.AttackAttempts["waf_url"]++
			
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "请求包含恶意内容",
				"error":   "WAF blocked malicious request",
			})
			c.Abort()
			return
		}
		
		// 检查请求头
		for key, values := range c.Request.Header {
			for _, value := range values {
				if s.detectWAFThreat(value, "header") {
					blockedRequests.Inc()
					s.metrics.BlockedRequests++
					s.metrics.AttackAttempts["waf_header"]++
					
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": "请求头包含恶意内容",
						"error":   "WAF blocked malicious header",
					})
					c.Abort()
					return
				}
			}
		}
		
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if s.detectWAFThreat(value, "query") {
					blockedRequests.Inc()
					s.metrics.BlockedRequests++
					s.metrics.AttackAttempts["waf_query"]++
					
					c.JSON(http.StatusForbidden, gin.H{
						"code":    403,
						"message": "查询参数包含恶意内容",
						"error":   "WAF blocked malicious query parameter",
					})
					c.Abort()
					return
				}
			}
		}
		
		// 检查POST数据（如果是POST请求）
		if c.Request.Method == "POST" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if s.detectWAFThreat(value, "post") {
							blockedRequests.Inc()
							s.metrics.BlockedRequests++
							s.metrics.AttackAttempts["waf_post"]++
							
							c.JSON(http.StatusForbidden, gin.H{
								"code":    403,
								"message": "POST数据包含恶意内容",
								"error":   "WAF blocked malicious POST data",
							})
							c.Abort()
							return
						}
					}
				}
			}
		}
		
		c.Next()
	}
}

// RequestValidationMiddleware 请求验证中间件
func (s *APISecurityService) RequestValidationMiddleware() gin.HandlerFunc {
	if !s.config.EnableRequestValidation {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("request_validation").Observe(time.Since(start).Seconds())
		}()
		
		// 验证请求方法
		if !s.isValidMethod(c.Request.Method) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"code":    405,
				"message": "不支持的请求方法",
				"error":   "Method not allowed",
			})
			c.Abort()
			return
		}
		
		// 验证Content-Type
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			contentType := c.GetHeader("Content-Type")
			if !s.isValidContentType(contentType) {
				blockedRequests.Inc()
				s.metrics.BlockedRequests++
				
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"code":    415,
					"message": "不支持的媒体类型",
					"error":   "Unsupported media type",
				})
				c.Abort()
				return
			}
		}
		
		// 验证请求大小
		if c.Request.ContentLength > 10*1024*1024 { // 10MB限制
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"code":    413,
				"message": "请求体过大",
				"error":   "Request entity too large",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// DDoSProtectionMiddleware DDoS保护中间件
func (s *APISecurityService) DDoSProtectionMiddleware() gin.HandlerFunc {
	if !s.config.EnableDDoSProtection {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("ddos_protection").Observe(time.Since(start).Seconds())
		}()
		
		ip := c.ClientIP()
		
		// 检查IP是否在黑名单中
		if s.isIPBlacklisted(ip) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "IP地址已被封禁",
				"error":   "IP blacklisted",
			})
			c.Abort()
			return
		}
		
		// 检查请求频率
		if s.isDDoSAttack(ip) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			s.metrics.AttackAttempts["ddos"]++
			
			// 将IP加入临时黑名单
			s.blockIP(ip, time.Hour)
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "检测到异常请求",
				"error":   "DDoS protection activated",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// RequestSigningMiddleware 请求签名中间件
func (s *APISecurityService) RequestSigningMiddleware() gin.HandlerFunc {
	if !s.config.EnableRequestSigning {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			apiSecurityDuration.WithLabelValues("request_signing").Observe(time.Since(start).Seconds())
		}()
		
		// 获取签名
		signature := c.GetHeader("X-API-Signature")
		if signature == "" {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少API签名",
				"error":   "Missing API signature",
			})
			c.Abort()
			return
		}
		
		// 验证签名
		if !s.validateRequestSignature(c.Request, signature) {
			blockedRequests.Inc()
			s.metrics.BlockedRequests++
			s.metrics.AttackAttempts["invalid_signature"]++
			
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的API签名",
				"error":   "Invalid API signature",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// SecurityHeadersMiddleware 安全头部中间件
func (s *APISecurityService) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 安全相关HTTP头部
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline';")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("X-Download-Options", "noopen")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("X-API-Version", "v1")
		
		// 移除敏感头部
		c.Header("Server", "")
		c.Header("X-Powered-By", "")
		
		// 添加安全相关头部
		c.Header("X-Request-ID", generateRequestID())
		c.Header("X-Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// 辅助方法
func (s *APISecurityService) isIPWhitelisted(ip string) bool {
	if len(s.config.WhitelistedIPs) == 0 {
		return true // 没有白名单限制
	}
	
	for _, whitelistedIP := range s.config.WhitelistedIPs {
		if ip == whitelistedIP {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) isIPBlacklisted(ip string) bool {
	// 检查配置中的黑名单
	for _, blacklistedIP := range s.config.BlacklistedIPs {
		if ip == blacklistedIP {
			return true
		}
	}
	
	// 检查动态黑名单
	if blockTime, exists := s.blockedIPs[ip]; exists {
		if time.Now().Before(blockTime) {
			return true
		}
		// 过期了，删除
		delete(s.blockedIPs, ip)
	}
	
	return false
}

func (s *APISecurityService) blockIP(ip string, duration time.Duration) {
	s.blockedIPs[ip] = time.Now().Add(duration)
}

func (s *APISecurityService) isDDoSAttack(ip string) bool {
	// 检查最近1分钟内的请求次数
	recentRequests := 0
	now := time.Now()
	
	if timestamps, exists := s.throttledIPs[ip]; exists {
		for _, ts := range timestamps {
			if now.Sub(ts) < time.Minute {
				recentRequests++
			}
		}
		
		// 记录当前请求
		s.throttledIPs[ip] = append(timestamps, now)
		
		// 清理过期的记录
		validTimestamps := []time.Time{}
		for _, ts := range timestamps {
			if now.Sub(ts) < time.Hour {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		s.throttledIPs[ip] = validTimestamps
	} else {
		s.throttledIPs[ip] = []time.Time{now}
	}
	
	// 如果1分钟内超过100个请求，认为是DDoS攻击
	return recentRequests > 100
}

func (s *APISecurityService) detectWAFThreat(input, inputType string) bool {
	for _, rule := range s.config.WAFRules {
		if !rule.Enabled {
			continue
		}
		
		// 根据规则类型进行不同的检测
		switch rule.Type {
		case "sql_injection":
			if s.detectSQLInjection(input) {
				return true
			}
		case "xss":
			if s.detectXSS(input) {
				return true
			}
		case "path_traversal":
			if s.detectPathTraversal(input) {
				return true
			}
		case "command_injection":
			if s.detectCommandInjection(input) {
				return true
			}
		case "file_inclusion":
			if s.detectFileInclusion(input) {
				return true
			}
		}
	}
	
	return false
}

func (s *APISecurityService) detectSQLInjection(input string) bool {
	// SQL注入检测模式
	sqlPatterns := []string{
		`(?i)(union\s+select|insert\s+into|delete\s+from|update\s+set|drop\s+table|alter\s+table)`,
		`(?i)(or\s+1\s*=\s*1|or\s+1\s*=\s*'1'|or\s+1\s*=\s*"1")`,
		`(?i)(;|--|\/\*|\*\/|@@|@)`,
		`(?i)(exec|execute|sp_|xp_|select|insert|update|delete|drop)`,
		`(?i)(waitfor\s+delay|sleep\()`,
	}
	
	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) detectXSS(input string) bool {
	// XSS检测模式
	xssPatterns := []string{
		`(?i)(<script|javascript:|onload=|onerror=|onclick=)`,
		`(?i)(<iframe|<object|<embed|<applet)`,
		`(?i)(eval\(|alert\(|confirm\(|prompt\()`,
		`(?i)(document\.|window\.|location\.|self\.)`,
		`(?i)(expression\(|script:|data:text/html)`,
	}
	
	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) detectPathTraversal(input string) bool {
	// 路径遍历检测模式
	pathPatterns := []string{
		`(\.\./|\.\.\\)`,
		`(/etc/passwd|/etc/shadow|/etc/hosts)`,
		`(c:\\|d:\\|e:\\|f:\\)`,
		`(%2e%2e%2f|%2e%2e%5c)`,
	}
	
	for _, pattern := range pathPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) detectCommandInjection(input string) bool {
	// 命令注入检测模式
	cmdPatterns := []string{
		`(?i)(;|\||&|&&|\|\||>|\>>|<|<<)`,
		`(?i)(rm\s+-rf|del\s+/f|format\s+c:)`,
		`(?i)(nc\s+-l|netcat|telnet|wget|curl)`,
		`(?i)(\/bin\/sh|\/bin\/bash|cmd\.exe|powershell)`,
		`(?i)(` + "`" + `|\$\(|\${)`,
	}
	
	for _, pattern := range cmdPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) detectFileInclusion(input string) bool {
	// 文件包含检测模式
	filePatterns := []string{
		`(?i)(php:\/\/|zip:\/\/|phar:\/\/|data:\/\/)`,
		`(?i)(include|require|include_once|require_once)`,
		`(?i)(file_get_contents|file_put_contents|fopen|readfile)`,
		`(?i)(\.\./|\.\.\\|\.\.\/)`,
	}
	
	for _, pattern := range filePatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) isValidMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return true
		}
	}
	return false
}

func (s *APISecurityService) isValidContentType(contentType string) bool {
	if contentType == "" {
		return true // 没有Content-Type也可以接受
	}
	
	validTypes := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
		"application/xml",
		"text/xml",
	}
	
	for _, validType := range validTypes {
		if strings.Contains(contentType, validType) {
			return true
		}
	}
	
	return false
}

func (s *APISecurityService) validateRequestSignature(req *http.Request, signature string) bool {
	// 这里应该实现具体的签名验证逻辑
	// 1. 获取API密钥
	// 2. 构造签名字符串
	// 3. 验证签名
	// 简化实现，实际项目中应该使用HMAC或其他签名算法
	
	expectedSignature := s.generateRequestSignature(req)
	return signature == expectedSignature
}

func (s *APISecurityService) generateRequestSignature(req *http.Request) string {
	// 生成请求签名的简化实现
	// 实际项目中应该使用更复杂的签名算法
	
	data := fmt.Sprintf("%s:%s:%s", req.Method, req.URL.Path, req.URL.Query().Encode())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetSecurityMetrics 获取安全指标
func (s *APISecurityService) GetSecurityMetrics() *SecurityMetrics {
	return s.metrics
}

// ClearMetrics 清除安全指标
func (s *APISecurityService) ClearMetrics() {
	s.metrics = &SecurityMetrics{SuspiciousIPs: make(map[string]int), AttackAttempts: make(map[string]int)}
}

// GetDefaultWAFRules 获取默认WAF规则
func GetDefaultWAFRules() []WAFRule {
	return []WAFRule{
		{
			ID:          "sql_injection_1",
			Name:        "SQL注入检测",
			Description: "检测常见的SQL注入攻击",
			Type:        "sql_injection",
			Pattern:     "(?i)(union\\s+select|insert\\s+into|delete\\s+from|update\\s+set)",
			Action:      "block",
			Severity:    "high",
			Enabled:     true,
		},
		{
			ID:          "xss_1",
			Name:        "XSS攻击检测",
			Description: "检测跨站脚本攻击",
			Type:        "xss",
			Pattern:     "(?i)(<script|javascript:|onload=|onerror=)",
			Action:      "block",
			Severity:    "high",
			Enabled:     true,
		},
		{
			ID:          "path_traversal_1",
			Name:        "路径遍历检测",
			Description: "检测路径遍历攻击",
			Type:        "path_traversal",
			Pattern:     "(\\.\\./|\\.\\.\\\\)",
			Action:      "block",
			Severity:    "medium",
			Enabled:     true,
		},
		{
			ID:          "command_injection_1",
			Name:        "命令注入检测",
			Description: "检测命令注入攻击",
			Type:        "command_injection",
			Pattern:     "(?i)(;|\\||&|&&|\\|\\||>|>>)",
			Action:      "block",
			Severity:    "high",
			Enabled:     true,
		},
	}
}