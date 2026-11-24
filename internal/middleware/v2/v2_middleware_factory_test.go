package v2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"law-oa-go/internal/config"
)

// newFactory 构造仅含测试所需字段的中间件工厂。
func newFactory(env, secret string) *V2MiddlewareFactory {
	return &V2MiddlewareFactory{
		config: &config.Config{
			Environment: env,
			JWT:         config.JWTConfig{Secret: secret},
		},
	}
}

// newAuthContext 构造一个带可选请求头的 gin 测试上下文。
func newAuthContext(apiKey, bearer string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	if bearer != "" {
		req.Header.Set("Authorization", bearer)
	}
	c.Request = req
	return c
}

// TestCheckMetricsAuth_DevBypass 验证开发环境对 /metrics 与 /system/info 的访问
// 统一放行（这是环境门控，不依赖任何密钥）。
func TestCheckMetricsAuth_DevBypass(t *testing.T) {
	f := newFactory("development", "")
	if !f.checkMetricsAuth(newAuthContext("", "")) {
		t.Fatal("development environment must allow metrics access without any credential")
	}
}

// TestCheckMetricsAuth_ValidSecret 验证非开发环境下，携带与 JWT.Secret 一致的
// API 密钥才能访问受保护端点。
func TestCheckMetricsAuth_ValidSecret(t *testing.T) {
	f := newFactory("production", "prod-secret-0123456789abcdef")
	if !f.checkMetricsAuth(newAuthContext("prod-secret-0123456789abcdef", "")) {
		t.Fatal("valid API key equal to JWT.Secret must be authorized in non-dev environment")
	}
}

// TestCheckMetricsAuth_DevAPIKeyBackdoorRemoved 是 G101 后门修复的回归测试：
// 内置明文密钥 "dev-api-key" 在任何环境下都不再能访问受保护端点。
// 修复前 `apiKey == "dev-api-key"` 会放行；修复后该常量已从代码中移除。
func TestCheckMetricsAuth_DevAPIKeyBackdoorRemoved(t *testing.T) {
	f := newFactory("production", "prod-secret-0123456789abcdef")
	if f.checkMetricsAuth(newAuthContext("dev-api-key", "")) {
		t.Fatal("hardcoded dev-api-key credential must NOT authorize access (G101 backdoor removed)")
	}
}

// TestCheckMetricsAuth_WrongSecret 验证错误密钥被拒绝。
func TestCheckMetricsAuth_WrongSecret(t *testing.T) {
	f := newFactory("production", "prod-secret-0123456789abcdef")
	if f.checkMetricsAuth(newAuthContext("wrong-secret", "")) {
		t.Fatal("wrong API key must not be authorized")
	}
}

// TestCheckMetricsAuth_NoCredential 验证非开发环境无任何凭据时被拒绝。
func TestCheckMetricsAuth_NoCredential(t *testing.T) {
	f := newFactory("production", "prod-secret-0123456789abcdef")
	if f.checkMetricsAuth(newAuthContext("", "")) {
		t.Fatal("no credential must not be authorized in non-dev environment")
	}
}
