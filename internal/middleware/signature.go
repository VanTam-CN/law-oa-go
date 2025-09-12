package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
)

// RequestSignatureMiddleware 请求签名验证中间件
func RequestSignatureMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查和swagger等路径
		skipPaths := []string{"/health", "/metrics", "/swagger", "/docs"}
		for _, path := range skipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// 验证必要header
		timestamp := c.GetHeader("X-Timestamp")
		signature := c.GetHeader("X-Signature")
		apiKey := c.GetHeader("X-API-Key")

		if timestamp == "" || signature == "" || apiKey == "" {
			common.Unauthorized(c, "Missing required headers: X-Timestamp, X-Signature, X-API-Key")
			c.Abort()
			return
		}

		// 验证时间戳（防止重放攻击）
		requestTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			common.BadRequest(c, "Invalid timestamp format")
			c.Abort()
			return
		}

		// 时间戳有效期5分钟
		if time.Since(requestTime) > 5*time.Minute || time.Until(requestTime) > 5*time.Minute {
			common.BadRequest(c, "Timestamp expired or invalid")
			c.Abort()
			return
		}

		// 验证API Key
		if !validateAPIKey(apiKey) {
			common.Unauthorized(c, "Invalid API key")
			c.Abort()
			return
		}

		// 验证签名
		body, err := c.GetRawData()
		if err != nil {
			common.BadRequest(c, "Failed to read request body")
			c.Abort()
			return
		}

		// 重新设置body供后续使用
		c.Request.Body = common.NewRequestBodyBuffer(body)

		// 计算期望签名
		expectedSignature := calculateSignature(body, timestamp, apiKey)

		// 验证签名
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			common.Unauthorized(c, "Invalid signature")
			c.Abort()
			return
		}

		// 记录验证成功
		logger := c.MustGet("logger").(*slog.Logger)
		logger.Info("Request signature verified",
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
			"api_key", apiKey[:8]+"...", // 只显示前8位
			"timestamp", timestamp,
		)

		c.Next()
	}
}

// validateAPIKey 验证API Key
func validateAPIKey(apiKey string) bool {
	// 这里应该从数据库或配置中验证API Key
	// 为了示例，我们简单检查长度和格式
	if len(apiKey) < 32 {
		return false
	}

	// 检查是否包含非法字符
	for _, c := range apiKey {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}

	return true
}

// calculateSignature 计算请求签名
func calculateSignature(body []byte, timestamp, apiKey string) string {
	// 构建签名字符串
	stringToSign := timestamp + ":" + string(body)

	// 使用HMAC-SHA256计算签名
	h := hmac.New(sha256.New, []byte(apiKey))
	h.Write([]byte(stringToSign))

	return hex.EncodeToString(h.Sum(nil))
}

// WebhookSignatureMiddleware Webhook签名验证
func WebhookSignatureMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Webhook-Signature")
		if signature == "" {
			common.Unauthorized(c, "Missing webhook signature")
			c.Abort()
			return
		}

		body, err := c.GetRawData()
		if err != nil {
			common.BadRequest(c, "Failed to read webhook body")
			c.Abort()
			return
		}

		// 重新设置body
		c.Request.Body = common.NewRequestBodyBuffer(body)

		// 验证签名
		expectedSignature := calculateWebhookSignature(body, secret)
		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			common.Unauthorized(c, "Invalid webhook signature")
			c.Abort()
			return
		}

		c.Next()
	}
}

// calculateWebhookSignature 计算Webhook签名
func calculateWebhookSignature(body []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}

// RateLimitByAPIKey 基于API Key的限流
func RateLimitByAPIKey() gin.HandlerFunc {
	// 这里可以实现基于API Key的限流逻辑
	// 每个API Key可以有不同的限流策略
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		// 这里可以根据API Key查询用户的限流配置
		// 然后应用相应的限流策略
		// 为了示例，我们使用默认的限流

		c.Next()
	}
}
