package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SecurityService JWT安全服务
type SecurityService struct {
	db               *gorm.DB
	logger           *logrus.Logger
	config           *SecurityConfig
	ipChecker        *IPChecker
	deviceFingerprinter *DeviceFingerprinter
	rateLimiter      *RateLimiter
	anomalyDetector  *AnomalyDetector
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// IP地址验证
	ValidateIP           bool          `json:"validate_ip"`
	AllowedIPRanges      []string      `json:"allowed_ip_ranges"`
	BlockedIPRanges      []string      `json:"blocked_ip_ranges"`
	IPValidationMode     IPValidationMode `json:"ip_validation_mode"`

	// 设备指纹验证
	ValidateDevice       bool          `json:"validate_device"`
	DeviceValidationMode DeviceValidationMode `json:"device_validation_mode"`
	MaxDevicesPerUser    int           `json:"max_devices_per_user"`

	// 速率限制
	EnableRateLimit      bool          `json:"enable_rate_limit"`
	GlobalRateLimit      int           `json:"global_rate_limit"`
	UserRateLimit        int           `json:"user_rate_limit"`
	IPLimit              int           `json:"ip_limit"`
	RateLimitWindow      time.Duration `json:"rate_limit_window"`

	// 异常检测
	EnableAnomalyDetection bool         `json:"enable_anomaly_detection"`
	AnomalyThreshold      float64       `json:"anomaly_threshold"`
	BehaviorTracking      bool          `json:"behavior_tracking"`

	// 令牌安全
	TokenReuseDetection   bool          `json:"token_reuse_detection"`
	ConcurrentSessionLimit int          `json:"concurrent_session_limit"`

	// 黑名单
	EnableBlacklist      bool          `json:"enable_blacklist"`
	BlacklistThreshold   int           `json:"blacklist_threshold"`
	BlacklistDuration    time.Duration `json:"blacklist_duration"`

	// 审计和安全日志
	EnableAuditLog       bool          `json:"enable_audit_log"`
	SecurityLogLevel     logrus.Level  `json:"security_log_level"`
}

// IPValidationMode IP验证模式
type IPValidationMode int

const (
	// IPModeStrict 严格模式，只允许白名单IP
	IPModeStrict IPValidationMode = iota
	// IPModePermissive 宽松模式，阻止黑名单IP
	IPModePermissive
	// IPModeAdaptive 自适应模式，基于风险评估
	IPModeAdaptive
)

// DeviceValidationMode 设备验证模式
type DeviceValidationMode int

const (
	// DeviceModeNone 不验证设备
	DeviceModeNone DeviceValidationMode = iota
	// DeviceModeBasic 基本设备验证
	DeviceModeBasic
	// DeviceModeStrict 严格设备验证
	DeviceModeStrict
)

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		ValidateIP:             true,
		AllowedIPRanges:        []string{},
		BlockedIPRanges:        []string{},
		IPValidationMode:       IPModePermissive,

		ValidateDevice:         true,
		DeviceValidationMode:   DeviceModeBasic,
		MaxDevicesPerUser:      5,

		EnableRateLimit:        true,
		GlobalRateLimit:        1000,
		UserRateLimit:          100,
		IPLimit:                50,
		RateLimitWindow:        time.Hour,

		EnableAnomalyDetection: true,
		AnomalyThreshold:       0.7,
		BehaviorTracking:       true,

		TokenReuseDetection:    true,
		ConcurrentSessionLimit: 3,

		EnableBlacklist:        true,
		BlacklistThreshold:     5,
		BlacklistDuration:      24 * time.Hour,

		EnableAuditLog:         true,
		SecurityLogLevel:       logrus.WarnLevel,
	}
}

// NewSecurityService 创建安全服务
func NewSecurityService(db *gorm.DB, logger *logrus.Logger, config *SecurityConfig) *SecurityService {
	service := &SecurityService{
		db:       db,
		logger:   logger,
		config:   config,
		ipChecker: NewIPChecker(config),
		deviceFingerprinter: NewDeviceFingerprinter(config),
		rateLimiter: NewRateLimiter(config),
		anomalyDetector: NewAnomalyDetector(config, logger),
	}

	// 启动后台任务
	go service.startBackgroundTasks()

	return service
}

// SecurityCheckRequest 安全检查请求
type SecurityCheckRequest struct {
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	TenantID    string    `json:"tenant_id"`
	SessionID   string    `json:"session_id"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	DeviceID    string    `json:"device_id"`
	RequestPath string    `json:"request_path"`
	HTTPMethod  string    `json:"http_method"`
	Timestamp   time.Time `json:"timestamp"`
	TokenJTI    string    `json:"token_jti"`
}

// SecurityCheckResult 安全检查结果
type SecurityCheckResult struct {
	Allowed      bool                   `json:"allowed"`
	RiskScore    float64                `json:"risk_score"`
	Threats      []SecurityThreat       `json:"threats"`
	Recommendations []SecurityRecommendation `json:"recommendations"`
	BlockReason  string                 `json:"block_reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SecurityThreat 安全威胁
type SecurityThreat struct {
	Type        ThreatType `json:"type"`
	Severity    ThreatSeverity `json:"severity"`
	Description string     `json:"description"`
	Confidence  float64    `json:"confidence"`
	Source      string     `json:"source"`
	Timestamp   time.Time  `json:"timestamp"`
}

// SecurityRecommendation 安全建议
type SecurityRecommendation struct {
	Type        RecommendationType `json:"type"`
	Action      string             `json:"action"`
	Description string             `json:"description"`
	Priority    Priority           `json:"priority"`
}

// ThreatType 威胁类型
type ThreatType string

const (
	ThreatTypeIPBlacklist      ThreatType = "ip_blacklist"
	ThreatTypeDeviceMismatch   ThreatType = "device_mismatch"
	ThreatTypeRateLimit        ThreatType = "rate_limit"
	ThreatTypeAnomalousBehavior ThreatType = "anomalous_behavior"
	ThreatTypeTokenReuse       ThreatType = "token_reuse"
	ThreatTypeConcurrentSession ThreatType = "concurrent_session"
	ThreatTypeBruteForce       ThreatType = "brute_force"
	ThreatTypeReplayAttack     ThreatType = "replay_attack"
)

// ThreatSeverity 威胁严重程度
type ThreatSeverity string

const (
	SeverityLow    ThreatSeverity = "low"
	SeverityMedium ThreatSeverity = "medium"
	SeverityHigh   ThreatSeverity = "high"
	SeverityCritical ThreatSeverity = "critical"
)

// RecommendationType 建议类型
type RecommendationType string

const (
	RecommendationTypeBlock         RecommendationType = "block"
	RecommendationTypeRateLimit     RecommendationType = "rate_limit"
	RecommendationTypeMFA           RecommendationType = "mfa"
	RecommendationTypePasswordReset RecommendationType = "password_reset"
	RecommendationTypeDeviceVerification RecommendationType = "device_verification"
	RecommendationTypeIPWhitelist   RecommendationType = "ip_whitelist"
)

// Priority 优先级
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityCritical Priority = "critical"
)

// ValidateRequest 验证请求安全性
func (s *SecurityService) ValidateRequest(req *SecurityCheckRequest) *SecurityCheckResult {
	result := &SecurityCheckResult{
		Allowed:  true,
		RiskScore: 0.0,
		Threats:  []SecurityThreat{},
		Recommendations: []SecurityRecommendation{},
		Metadata: make(map[string]interface{}),
	}

	// 1. IP地址验证
	if s.config.ValidateIP {
		threat := s.ipChecker.ValidateIP(req.IPAddress)
		if threat != nil {
			result.Threats = append(result.Threats, *threat)
			result.RiskScore += threat.Confidence
			if threat.Severity == SeverityCritical || threat.Severity == SeverityHigh {
				result.Allowed = false
				result.BlockReason = threat.Description
			}
		}
	}

	// 2. 设备验证
	if s.config.ValidateDevice {
		threat := s.deviceFingerprinter.ValidateDevice(req)
		if threat != nil {
			result.Threats = append(result.Threats, *threat)
			result.RiskScore += threat.Confidence
			if threat.Severity == SeverityCritical {
				result.Allowed = false
				result.BlockReason = threat.Description
			}
		}
	}

	// 3. 速率限制检查
	if s.config.EnableRateLimit {
		threat := s.rateLimiter.CheckRateLimit(req)
		if threat != nil {
			result.Threats = append(result.Threats, *threat)
			result.RiskScore += threat.Confidence
			if threat.Severity == SeverityHigh || threat.Severity == SeverityCritical {
				result.Allowed = false
				result.BlockReason = threat.Description
			}
		}
	}

	// 4. 令牌重用检测
	if s.config.TokenReuseDetection {
		threat := s.detectTokenReuse(req)
		if threat != nil {
			result.Threats = append(result.Threats, *threat)
			result.RiskScore += threat.Confidence
			if threat.Severity == SeverityCritical {
				result.Allowed = false
				result.BlockReason = threat.Description
			}
		}
	}

	// 5. 并发会话检查
	if s.config.ConcurrentSessionLimit > 0 {
		threat := s.checkConcurrentSessions(req)
		if threat != nil {
			result.Threats = append(result.Threats, *threat)
			result.RiskScore += threat.Confidence
		}
	}

	// 6. 异常行为检测
	if s.config.EnableAnomalyDetection {
		threats := s.anomalyDetector.DetectAnomalies(req)
		result.Threats = append(result.Threats, threats...)
		for _, threat := range threats {
			result.RiskScore += threat.Confidence
		}
	}

	// 7. 生成安全建议
	result.Recommendations = s.generateRecommendations(result.Threats)

	// 8. 记录安全事件
	if s.config.EnableAuditLog && len(result.Threats) > 0 {
		s.logSecurityEvent(req, result)
	}

	// 9. 更新黑名单
	if s.config.EnableBlacklist && result.RiskScore > 0.8 {
		s.updateBlacklist(req, result.RiskScore)
	}

	return result
}

// detectTokenReuse 检测令牌重用
func (s *SecurityService) detectTokenReuse(req *SecurityCheckRequest) *SecurityThreat {
	// 检查令牌是否在短时间内被多个IP使用
	// 这里简化实现，实际应该查询数据库

	// 模拟实现：如果在过去5分钟内同一令牌被不同IP使用，则认为可能存在重用
	// 实际实现需要数据库查询和缓存支持

	return nil // 简化实现，暂不检测
}

// checkConcurrentSessions 检查并发会话
func (s *SecurityService) checkConcurrentSessions(req *SecurityCheckRequest) *SecurityThreat {
	// 查询用户当前活跃会话数
	// 这里简化实现，实际应该查询数据库

	return nil // 简化实现，暂不检测
}

// generateRecommendations 生成安全建议
func (s *SecurityService) generateRecommendations(threats []SecurityThreat) []SecurityRecommendation {
	recommendations := []SecurityRecommendation{}

	for _, threat := range threats {
		switch threat.Type {
		case ThreatTypeIPBlacklist:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypeBlock,
				Action:      "block_ip",
				Description: "Block malicious IP address",
				Priority:    PriorityHigh,
			})
		case ThreatTypeDeviceMismatch:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypeDeviceVerification,
				Action:      "verify_device",
				Description: "Require device verification",
				Priority:    PriorityMedium,
			})
		case ThreatTypeRateLimit:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypeRateLimit,
				Action:      "apply_stricter_rate_limit",
				Description: "Apply stricter rate limiting",
				Priority:    PriorityMedium,
			})
		case ThreatTypeAnomalousBehavior:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypeMFA,
				Action:      "require_mfa",
				Description: "Require multi-factor authentication",
				Priority:    PriorityHigh,
			})
		case ThreatTypeBruteForce:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypePasswordReset,
				Action:      "force_password_reset",
				Description: "Force password reset",
				Priority:    PriorityCritical,
			})
		case ThreatTypeReplayAttack:
			recommendations = append(recommendations, SecurityRecommendation{
				Type:        RecommendationTypeBlock,
				Action:      "block_session",
				Description: "Block and invalidate session",
				Priority:    PriorityCritical,
			})
		}
	}

	return recommendations
}

// logSecurityEvent 记录安全事件
func (s *SecurityService) logSecurityEvent(req *SecurityCheckRequest, result *SecurityCheckResult) {
	// 构建日志字段
	fields := logrus.Fields{
		"user_id":       req.UserID,
		"username":      req.Username,
		"tenant_id":     req.TenantID,
		"session_id":    req.SessionID,
		"ip_address":    req.IPAddress,
		"user_agent":    req.UserAgent,
		"request_path":  req.RequestPath,
		"http_method":   req.HTTPMethod,
		"risk_score":    result.RiskScore,
		"allowed":       result.Allowed,
		"threat_count":  len(result.Threats),
	}

	// 如果有威胁，添加威胁信息
	if len(result.Threats) > 0 {
		threatTypes := make([]string, len(result.Threats))
		for i, threat := range result.Threats {
			threatTypes[i] = string(threat.Type)
		}
		fields["threat_types"] = threatTypes
		fields["block_reason"] = result.BlockReason
	}

	// 根据风险评分决定日志级别
	var logLevel logrus.Level
	if result.RiskScore >= 0.8 {
		logLevel = logrus.ErrorLevel
	} else if result.RiskScore >= 0.6 {
		logLevel = logrus.WarnLevel
	} else if result.RiskScore >= 0.3 {
		logLevel = logrus.InfoLevel
	} else {
		logLevel = logrus.DebugLevel
	}

	s.logger.WithFields(fields).Log(logLevel, "Security event detected")
}

// updateBlacklist 更新黑名单
func (s *SecurityService) updateBlacklist(req *SecurityCheckRequest, riskScore float64) {
	// 这里应该实现IP和设备的黑名单更新逻辑
	// 简化实现，实际应该写入数据库或缓存

	s.logger.WithFields(logrus.Fields{
		"ip_address": req.IPAddress,
		"user_id":    req.UserID,
		"risk_score": riskScore,
	}).Warn("Added to security blacklist")
}

// startBackgroundTasks 启动后台任务
func (s *SecurityService) startBackgroundTasks() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredData()
			s.updateSecurityMetrics()
		}
	}
}

// cleanupExpiredData 清理过期数据
func (s *SecurityService) cleanupExpiredData() {
	// 清理过期的安全事件记录
	// 清理过期的黑名单记录
	s.logger.Debug("Security cleanup completed")
}

// updateSecurityMetrics 更新安全指标
func (s *SecurityService) updateSecurityMetrics() {
	// 更新安全相关的监控指标
	s.logger.Debug("Security metrics updated")
}

// IPChecker IP地址检查器
type IPChecker struct {
	config       *SecurityConfig
	allowedCIDRs []*net.IPNet
	blockedCIDRs []*net.IPNet
}

// NewIPChecker 创建IP检查器
func NewIPChecker(config *SecurityConfig) *IPChecker {
	checker := &IPChecker{
		config: config,
	}

	// 解析允许的IP范围
	for _, ipRange := range config.AllowedIPRanges {
		_, cidr, err := net.ParseCIDR(ipRange)
		if err == nil {
			checker.allowedCIDRs = append(checker.allowedCIDRs, cidr)
		} else {
			// 尝试解析为单个IP
			if ip := net.ParseIP(ipRange); ip != nil {
				_, cidr, _ := net.ParseCIDR(ipRange + "/32")
				checker.allowedCIDRs = append(checker.allowedCIDRs, cidr)
			}
		}
	}

	// 解析阻止的IP范围
	for _, ipRange := range config.BlockedIPRanges {
		_, cidr, err := net.ParseCIDR(ipRange)
		if err == nil {
			checker.blockedCIDRs = append(checker.blockedCIDRs, cidr)
		} else {
			// 尝试解析为单个IP
			if ip := net.ParseIP(ipRange); ip != nil {
				_, cidr, _ := net.ParseCIDR(ipRange + "/32")
				checker.blockedCIDRs = append(checker.blockedCIDRs, cidr)
			}
		}
	}

	return checker
}

// ValidateIP 验证IP地址
func (c *IPChecker) ValidateIP(ipAddress string) *SecurityThreat {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return &SecurityThreat{
			Type:        ThreatTypeIPBlacklist,
			Severity:    SeverityHigh,
			Description: "Invalid IP address format",
			Confidence:  1.0,
			Source:      "ip_validator",
			Timestamp:   time.Now(),
		}
	}

	// 检查是否在阻止列表中
	for _, cidr := range c.blockedCIDRs {
		if cidr.Contains(ip) {
			return &SecurityThreat{
				Type:        ThreatTypeIPBlacklist,
				Severity:    SeverityCritical,
				Description: "IP address is in blacklist",
				Confidence:  1.0,
				Source:      "ip_blacklist",
				Timestamp:   time.Now(),
			}
		}
	}

	// 如果是严格模式，检查是否在允许列表中
	if c.config.IPValidationMode == IPModeStrict {
		allowed := false
		for _, cidr := range c.allowedCIDRs {
			if cidr.Contains(ip) {
				allowed = true
				break
			}
		}
		if !allowed && len(c.allowedCIDRs) > 0 {
			return &SecurityThreat{
				Type:        ThreatTypeIPBlacklist,
				Severity:    SeverityHigh,
				Description: "IP address not in whitelist (strict mode)",
				Confidence:  0.8,
				Source:      "ip_whitelist",
				Timestamp:   time.Now(),
			}
		}
	}

	return nil
}

// DeviceFingerprinter 设备指纹器
type DeviceFingerprinter struct {
	config *SecurityConfig
}

// NewDeviceFingerprinter 创建设备指纹器
func NewDeviceFingerprinter(config *SecurityConfig) *DeviceFingerprinter {
	return &DeviceFingerprinter{
		config: config,
	}
}

// ValidateDevice 验证设备
func (d *DeviceFingerprinter) ValidateDevice(req *SecurityCheckRequest) *SecurityThreat {
	if d.config.DeviceValidationMode == DeviceModeNone {
		return nil
	}

	// 生成设备指纹
	fingerprint := d.GenerateFingerprint(req.UserAgent, req.IPAddress)

	// 这里应该与数据库中的设备指纹进行比较
	// 简化实现，暂不进行实际的设备验证

	return nil
}

// GenerateFingerprint 生成设备指纹
func (d *DeviceFingerprinter) GenerateFingerprint(userAgent, ipAddress string) string {
	// 简化的指纹生成算法
	data := userAgent + "|" + ipAddress
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// RateLimiter 速率限制器
type RateLimiter struct {
	config    *SecurityConfig
	counters  map[string]*Counter
	mutex     sync.RWMutex
}

// Counter 计数器
type Counter struct {
	Count    int
	Window   time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(config *SecurityConfig) *RateLimiter {
	return &RateLimiter{
		config:   config,
		counters: make(map[string]*Counter),
	}
}

// CheckRateLimit 检查速率限制
func (r *RateLimiter) CheckRateLimit(req *SecurityCheckRequest) *SecurityThreat {
	now := time.Now()
	windowStart := now.Add(-r.config.RateLimitWindow)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// 检查全局速率限制
	globalKey := "global"
	if r.checkAndIncrement(globalKey, windowStart, r.config.GlobalRateLimit) {
		return &SecurityThreat{
			Type:        ThreatTypeRateLimit,
			Severity:    SeverityHigh,
			Description: "Global rate limit exceeded",
			Confidence:  1.0,
			Source:      "rate_limiter",
			Timestamp:   now,
		}
	}

	// 检查用户速率限制
	userKey := fmt.Sprintf("user:%d", req.UserID)
	if r.checkAndIncrement(userKey, windowStart, r.config.UserRateLimit) {
		return &SecurityThreat{
			Type:        ThreatTypeRateLimit,
			Severity:    SeverityMedium,
			Description: "User rate limit exceeded",
			Confidence:  0.8,
			Source:      "rate_limiter",
			Timestamp:   now,
		}
	}

	// 检查IP速率限制
	ipKey := fmt.Sprintf("ip:%s", req.IPAddress)
	if r.checkAndIncrement(ipKey, windowStart, r.config.IPLimit) {
		return &SecurityThreat{
			Type:        ThreatTypeRateLimit,
			Severity:    SeverityMedium,
			Description: "IP rate limit exceeded",
			Confidence:  0.7,
			Source:      "rate_limiter",
			Timestamp:   now,
		}
	}

	return nil
}

// checkAndIncrement 检查并递增计数器
func (r *RateLimiter) checkAndIncrement(key string, windowStart time.Time, limit int) bool {
	counter, exists := r.counters[key]
	if !exists || counter.Window.Before(windowStart) {
		// 重置计数器
		r.counters[key] = &Counter{
			Count:  1,
			Window: time.Now(),
		}
		return false
	}

	if counter.Count >= limit {
		return true
	}

	counter.Count++
	return false
}

// AnomalyDetector 异常检测器
type AnomalyDetector struct {
	config      *SecurityConfig
	logger      *logrus.Logger
	behaviorMap map[uint]*UserBehavior
	mutex       sync.RWMutex
}

// UserBehavior 用户行为
type UserBehavior struct {
	UserID              uint
	LoginTimes          []time.Time
	IPAddresses         map[string]int
	UserAgents          map[string]int
	RequestPatterns     map[string]int
	TypicalWorkingHours []int
	LastUpdate          time.Time
}

// NewAnomalyDetector 创建异常检测器
func NewAnomalyDetector(config *SecurityConfig, logger *logrus.Logger) *AnomalyDetector {
	return &AnomalyDetector{
		config:      config,
		logger:      logger,
		behaviorMap: make(map[uint]*UserBehavior),
	}
}

// DetectAnomalies 检测异常
func (a *AnomalyDetector) DetectAnomalies(req *SecurityCheckRequest) []SecurityThreat {
	var threats []SecurityThreat

	a.mutex.Lock()
	defer a.mutex.Unlock()

	behavior, exists := a.behaviorMap[req.UserID]
	if !exists {
		// 首次访问，创建行为记录
		behavior = &UserBehavior{
			UserID:              req.UserID,
			LoginTimes:          []time.Time{req.Timestamp},
			IPAddresses:         make(map[string]int),
			UserAgents:          make(map[string]int),
			RequestPatterns:     make(map[string]int),
			TypicalWorkingHours: []int{},
			LastUpdate:          req.Timestamp,
		}
		a.behaviorMap[req.UserID] = behavior
		return threats
	}

	// 更新行为记录
	a.updateBehavior(behavior, req)

	// 检测异常
	threats = append(threats, a.detectAnomalousIP(behavior, req)...)
	threats = append(threats, a.detectAnomalousTime(behavior, req)...)
	threats = append(threats, a.detectAnomalousLocation(behavior, req)...)

	return threats
}

// updateBehavior 更新用户行为记录
func (a *AnomalyDetector) updateBehavior(behavior *UserBehavior, req *SecurityCheckRequest) {
	behavior.IPAddresses[req.IPAddress]++
	behavior.UserAgents[req.UserAgent]++

	requestKey := fmt.Sprintf("%s %s", req.HTTPMethod, req.RequestPath)
	behavior.RequestPatterns[requestKey]++

	hour := req.Timestamp.Hour()
	behavior.TypicalWorkingHours = append(behavior.TypicalWorkingHours, hour)

	// 保持最近100个登录时间
	if len(behavior.LoginTimes) > 100 {
		behavior.LoginTimes = behavior.LoginTimes[1:]
	}
	behavior.LoginTimes = append(behavior.LoginTimes, req.Timestamp)

	behavior.LastUpdate = req.Timestamp
}

// detectAnomalousIP 检测异常IP
func (a *AnomalyDetector) detectAnomalousIP(behavior *UserBehavior, req *SecurityCheckRequest) []SecurityThreat {
	var threats []SecurityThreat

	// 如果用户使用了新IP，且该IP不在常用IP列表中
	if behavior.IPAddresses[req.IPAddress] == 1 && len(behavior.IPAddresses) > 1 {
		threats = append(threats, SecurityThreat{
			Type:        ThreatTypeAnomalousBehavior,
			Severity:    SeverityMedium,
			Description: "New IP address detected",
			Confidence:  0.6,
			Source:      "anomaly_detector",
			Timestamp:   time.Now(),
		})
	}

	return threats
}

// detectAnomalousTime 检测异常时间
func (a *AnomalyDetector) detectAnomalousTime(behavior *UserBehavior, req *SecurityCheckRequest) []SecurityThreat {
	var threats []SecurityThreat

	// 如果用户在非工作时间访问
	hour := req.Timestamp.Hour()
	if !a.isTypicalWorkingHour(behavior, hour) {
		threats = append(threats, SecurityThreat{
			Type:        ThreatTypeAnomalousBehavior,
			Severity:    SeverityLow,
			Description: "Access outside typical working hours",
			Confidence:  0.4,
			Source:      "anomaly_detector",
			Timestamp:   time.Now(),
		})
	}

	return threats
}

// detectAnomalousLocation 检测异常位置
func (a *AnomalyDetector) detectAnomalousLocation(behavior *UserBehavior, req *SecurityCheckRequest) []SecurityThreat {
	// 这里可以实现地理位置异常检测
	// 简化实现，暂不检测
	return []SecurityThreat{}
}

// isTypicalWorkingHour 检查是否为典型工作时间
func (a *AnomalyDetector) isTypicalWorkingHour(behavior *UserBehavior, hour int) bool {
	if len(behavior.TypicalWorkingHours) < 10 {
		return true // 数据不足，不判断
	}

	// 计算每个小时的出现频率
	hourCount := make(map[int]int)
	for _, h := range behavior.TypicalWorkingHours {
		hourCount[h]++
	}

	// 如果当前小时的使用频率低于10%，认为是异常
	totalHours := len(behavior.TypicalWorkingHours)
	currentHourCount := hourCount[hour]
	percentage := float64(currentHourCount) / float64(totalHours)

	return percentage >= 0.1
}