package auth

import (
	"fmt"
	"strings"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

// TokenValidator 令牌验证器
type TokenValidator struct {
	config *JWTConfig
	logger *logrus.Logger
}

// NewTokenValidator 创建令牌验证器
func NewTokenValidator(config *JWTConfig, logger *logrus.Logger) *TokenValidator {
	return &TokenValidator{
		config: config,
		logger: logger,
	}
}

// ValidateClaims 验证声明
func (v *TokenValidator) ValidateClaims(claims *TokenClaims) []string {
	var errors []string

	// 验证标准声明
	errors = append(errors, v.validateRegisteredClaims(claims.RegisteredClaims)...)

	// 验证自定义声明
	errors = append(errors, v.validateCustomClaims(claims)...)

	// 验证业务规则
	errors = append(errors, v.validateBusinessRules(claims)...)

	return errors
}

// validateRegisteredClaims 验证标准声明
func (v *TokenValidator) validateRegisteredClaims(claims jwt.RegisteredClaims) []string {
	var errors []string

	// 验证发行者
	if claims.Issuer != v.config.Issuer {
		errors = append(errors, fmt.Sprintf("invalid issuer: expected %s, got %s", v.config.Issuer, claims.Issuer))
	}

	// 验证受众
	if claims.Audience != v.config.Audience && claims.Audience != v.config.Audience+":refresh" {
		errors = append(errors, fmt.Sprintf("invalid audience: expected %s or %s, got %s", v.config.Audience, v.config.Audience+":refresh", claims.Audience))
	}

	// 验证过期时间
	if claims.ExpiresAt == nil {
		errors = append(errors, "missing expiration time")
	} else {
		if claims.ExpiresAt.Time.Before(time.Now().Add(v.config.Leeway)) {
			errors = append(errors, "token is expired")
		}
	}

	// 验证生效时间
	if claims.NotBefore != nil && claims.NotBefore.Time.After(time.Now().Add(v.config.Leeway)) {
		errors = append(errors, "token is not yet valid")
	}

	// 验证签发时间
	if claims.IssuedAt == nil {
		errors = append(errors, "missing issued at time")
	} else {
		// 验证令牌年龄
		tokenAge := time.Since(claims.IssuedAt.Time)
		if tokenAge > v.config.MaxTokenAge {
			errors = append(errors, fmt.Sprintf("token is too old: %v, max allowed: %v", tokenAge, v.config.MaxTokenAge))
		}
	}

	// 验证JTI
	if claims.ID == "" {
		errors = append(errors, "missing JWT ID")
	}

	return errors
}

// validateCustomClaims 验证自定义声明
func (v *TokenValidator) validateCustomClaims(claims *TokenClaims) []string {
	var errors []string

	// 验证用户ID
	if claims.UserID == 0 {
		errors = append(errors, "missing user ID")
	}

	// 验证用户名
	if claims.Username == "" {
		errors = append(errors, "missing username")
	}

	// 验证租户ID
	if claims.TenantID == "" {
		errors = append(errors, "missing tenant ID")
	}

	// 验证设备ID
	if claims.DeviceID == "" {
		errors = append(errors, "missing device ID")
	}

	// 验证会话ID
	if claims.SessionID == "" {
		errors = append(errors, "missing session ID")
	}

	// 验证用户代理
	if claims.UserAgent == "" {
		errors = append(errors, "missing user agent")
	}

	// 验证IP地址
	if claims.IPAddress == "" {
		errors = append(errors, "missing IP address")
	}

	return errors
}

// validateBusinessRules 验证业务规则
func (v *TokenValidator) validateBusinessRules(claims *TokenClaims) []string {
	var errors []string

	// 验证令牌使用场景
	if claims.ResourceAccess != nil {
		if err := v.validateResourceAccess(claims.ResourceAccess, claims); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 验证约束
	if claims.Constraints != nil {
		if err := v.validateConstraints(claims.Constraints, claims); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 验证设备指纹
	if claims.Fingerprint != "" && claims.UserAgent != "" {
		if err := v.validateFingerprint(claims.Fingerprint, claims.UserAgent, claims.IPAddress); err != nil {
			errors = append(errors, err.Error())
		}
	}

	// 验证nonce防重放攻击
	if claims.Nonce != "" {
		if err := v.validateNonce(claims.Nonce, claims.SessionID); err != nil {
			errors = append(errors, err.Error())
		}
	}

	return errors
}

// validateResourceAccess 验证资源访问权限
func (v *TokenValidator) validateResourceAccess(resourceAccess map[string]interface{}, claims *TokenClaims) error {
	// 验证用户是否有访问特定资源的权限
	// 这里应该结合权限系统进行验证

	for resource, permissions := range resourceAccess {
		switch perms := permissions.(type) {
		case map[string]interface{}:
			// 验证细粒度权限
			for action, allowed := range perms {
				if allowed.(bool) {
					if err := v.checkResourcePermission(claims.UserID, resource, action, claims); err != nil {
						return err
					}
				}
			}
		case []interface{}:
			// 验证权限列表
			for _, perm := range perms {
				if action, ok := perm.(string); ok {
					if err := v.checkResourcePermission(claims.UserID, resource, action, claims); err != nil {
						return err
					}
				}
			}
		case bool:
			// 简单的布尔权限检查
			if !permissions.(bool) {
				return fmt.Errorf("access denied to resource %s", resource)
			}
		case string:
			// 字符串权限描述
			if err := v.checkResourcePermission(claims.UserID, resource, permissions.(string), claims); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateConstraints 验证约束条件
func (v *TokenValidator) validateConstraints(constraints map[string]interface{}, claims *TokenClaims) error {
	for constraintType, constraintValue := range constraints {
		switch constraintType {
		case "time_range":
			if err := v.validateTimeConstraint(constraintValue, claims); err != nil {
				return err
			}
		case "ip_range":
			if err := v.validateIPRangeConstraint(constraintValue, claims); err != nil {
				return err
			}
		case "device_type":
			if err := v.validateDeviceTypeConstraint(constraintValue, claims); err != nil {
				return err
			}
		case "location":
			if err := v.validateLocationConstraint(constraintValue, claims); err != nil {
				return err
			}
		case "rate_limit":
			if err := v.validateRateLimitConstraint(constraintValue, claims); err != nil {
				return err
			}
		default:
			v.logger.WithField("constraint_type", constraintType).Warn("Unknown constraint type")
		}
	}

	return nil
}

// validateFingerprint 验证设备指纹
func (v *TokenValidator) validateFingerprint(fingerprint, userAgent, ipAddress string) error {
	// 简化实现，实际应该验证指纹的一致性
	// 可以基于用户代理、IP地址等信息生成和验证指纹

	if fingerprint == "" {
		return fmt.Errorf("empty fingerprint")
	}

	// 检查指纹格式
	if len(fingerprint) < 16 {
		return fmt.Errorf("fingerprint too short, minimum 16 characters required")
	}

	// 验证指纹是否包含用户代理信息
	if userAgent != "" && !strings.Contains(fingerprint, v.hashString(userAgent)) {
		return fmt.Errorf("fingerprint does not match user agent")
	}

	return nil
}

// validateNonce 验证nonce防重放攻击
func (v *TokenValidator) validateNonce(nonce, sessionID string) error {
	// 简化实现，实际应该检查nonce的唯一性
	// 可以使用Redis缓存来存储已使用的nonce

	if nonce == "" {
		return fmt.Errorf("empty nonce")
	}

	if len(nonce) < 16 {
		return fmt.Errorf("nonce too short, minimum 16 characters required")
	}

	// 检查nonce格式（应该只包含字母数字和特殊字符）
	for _, char := range nonce {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.') {
			return fmt.Errorf("invalid nonce character: %c", char)
		}
	}

	return nil
}

// validateTimeConstraint 验证时间约束
func (v *TokenValidator) validateTimeConstraint(constraintValue interface{}, claims *TokenClaims) error {
	constraint, ok := constraintValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid time constraint format")
	}

	start, startExists := constraint["start"]
	end, endExists := constraint["end"]
	now := time.Now()

	if startExists && endExists {
		startTime, err := time.Parse(time.RFC3339, start.(string))
		if err != nil {
			return fmt.Errorf("invalid start time format: %w", err)
		}

		endTime, err := time.Parse(time.RFC3339, end.(string))
		if err != nil {
			return fmt.Errorf("invalid end time format: %w", err)
		}

		if now.Before(startTime) || now.After(endTime) {
			return fmt.Errorf("access outside allowed time range: %v - %v", startTime, endTime)
		}
	}

	return nil
}

// validateIPRangeConstraint 验证IP范围约束
func (v *TokenValidator) validateIPRangeConstraint(constraintValue interface{}, claims *TokenClaims) error {
	constraint, ok := constraintValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid IP range constraint format")
	}

	allowedIPs, exists := constraint["allowed_ips"]
	if !exists {
		return fmt.Errorf("missing allowed IPs in IP range constraint")
	}

	userIP := claims.IPAddress
	if userIP == "" {
		return fmt.Errorf("missing user IP address")
	}

	switch ips := allowedIPs.(type) {
	case []interface{}:
		for _, ip := range ips {
			if ipStr, ok := ip.(string); ok {
				if v.isIPInRange(userIP, ipStr) {
					return nil
				}
			}
		}
	case string:
		if v.isIPInRange(userIP, allowedIPs.(string)) {
			return nil
		}
	default:
		return fmt.Errorf("invalid allowed IPs format")
	}

	return fmt.Errorf("IP address %s not in allowed range", userIP)
}

// validateDeviceTypeConstraint 验证设备类型约束
func (v *TokenValidator) validateDeviceTypeConstraint(constraintValue interface{}, claims *TokenClaims) error {
	constraint, ok := constraintValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid device type constraint format")
	}

	allowedTypes, exists := constraint["allowed_types"]
	if !exists {
		return fmt.Errorf("missing allowed types in device type constraint")
	}

	// 从用户代理中提取设备类型
	deviceType := v.extractDeviceType(claims.UserAgent)
	if deviceType == "" {
		return fmt.Errorf("cannot determine device type from user agent")
	}

	switch types := allowedTypes.(type) {
	case []interface{}:
		for _, allowedType := range types {
			if typeStr, ok := allowedType.(string); ok && typeStr == deviceType {
				return nil
			}
		}
	case string:
		if allowedTypes.(string) == deviceType {
			return nil
		}
	default:
		return fmt.Errorf("invalid allowed types format")
	}

	return fmt.Errorf("device type %s is not allowed", deviceType)
}

// validateLocationConstraint 验证地理位置约束
func (v *TokenValidator) validateLocationConstraint(constraintValue interface{}, claims *TokenClaims) error {
	constraint, ok := constraintValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid location constraint format")
	}

	allowedLocations, exists := constraint["allowed_locations"]
	if !exists {
		return fmt.Errorf("missing allowed locations in location constraint")
	}

	// 简化实现，实际应该使用地理位置服务
	// 这里只检查国家或地区代码
	return nil
}

// validateRateLimitConstraint 验证速率限制约束
func (v *TokenValidator) validateRateLimitConstraint(constraintValue interface{}, claims *TokenClaims) error {
	constraint, ok := constraintValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid rate limit constraint format")
	}

	maxRequests, exists := constraint["max_requests"]
	window, windowExists := constraint["window"]

	if !exists || !windowExists {
		return fmt.Errorf("missing rate limit parameters")
	}

	// 简化实现，实际应该使用Redis或内存缓存来跟踪请求频率
	return nil
}

// checkResourcePermission 检查资源权限
func (v *TokenValidator) checkResourcePermission(userID uint, resource, action string, claims *TokenClaims) error {
	// 简化实现，实际应该调用权限服务
	// 这里应该检查用户是否有对特定资源的特定操作权限

	// 管单的角色检查
	if len(claims.Roles) > 0 {
		for _, role := range claims.Roles {
			if role == "admin" {
				return nil // 管理员有所有权限
			}
		}
	}

	// 简单的权限检查逻辑
	if resource == "document" && action == "read" {
		return nil // 默认允许读取文档
	}

	return fmt.Errorf("access denied: insufficient permissions for resource %s action %s", resource, action)
}

// isIPInRange 检查IP是否在范围内
func (v *TokenValidator) isIPInRange(ip, rangeStr string) bool {
	// 简化实现，支持单个IP和CIDR格式
	if ip == rangeStr {
		return true
	}

	// 支持CIDR格式
	if strings.Contains(rangeStr, "/") {
		// 解析CIDR并检查IP是否在范围内
		// 简化实现
		return false
	}

	// 支持IP范围
	if strings.Contains(rangeStr, "-") {
		// 解析IP范围并检查IP是否在范围内
		// 简化实现
		return false
	}

	return false
}

// extractDeviceType 从用户代理中提取设备类型
func (v *TokenValidator) extractDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	// 移动设备检测
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") {
		return "mobile"
	}

	// 平板检测
	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}

	// 桌面端检测
	if strings.Contains(ua, "windows") || strings.Contains(ua, "macintosh") || strings.Contains(ua, "linux") || strings.Contains(ua, "x11") {
		return "desktop"
	}

	return "unknown"
}

// hashString 哈希字符串
func (v *TokenValidator) hashString(input string) string {
	// 简化实现，实际应该使用更安全的哈希算法
	return fmt.Sprintf("%x", len(input)*31 + int(input[0]))
}