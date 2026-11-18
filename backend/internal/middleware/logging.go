package middleware

import (
	"crypto/rand"
	"log/slog"
	"math/big"
	"time"

	"github.com/gin-gonic/gin"
)

type responseLogger struct {
	gin.ResponseWriter
	body *[]byte
}

func (r *responseLogger) Write(b []byte) (int, error) {
	*r.body = append(*r.body, b...)
	return r.ResponseWriter.Write(b)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 记录请求开始
		logger.Info("Request started",
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"request_id", c.GetString("request_id"),
		)

		// 处理请求
		c.Next()

		// 记录请求结束
		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		logger.Info("Request completed",
			"method", c.Request.Method,
			"path", path,
			"query", raw,
			"status", status,
			"duration", duration,
			"size", size,
			"ip", c.ClientIP(),
			"request_id", c.GetString("request_id"),
			"errors", c.Errors.String(),
		)
	}
}

// RequestIDMiddleware 生成请求ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRandomString(16)
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
