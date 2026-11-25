package performance

// api_fixture_test.go 提供 BenchmarkAPIRequest / BenchmarkConcurrentRequests
// 的真实入口：真实 router.Init + 隔离 PG + 真实登录 token。
// DB/Redis 限定 loopback 合成服务。所有迭代严格断言 ==200，失败只报状态码。

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// BenchmarkAPIRequest 驱动 4 个 API 子项：3 个带 token 的 GET 与 1 个真实登录
// POST。dashboard 等路由已在 setupBenchRouter 内完成计时前预验证，这里只
// 在计时前确认一次各子项可达。
func BenchmarkAPIRequest(b *testing.B) {
	b.Helper()
	fx := setupBenchRouter(b)
	app := fx.app
	token := fx.token

	loginBody := []byte(npBenchLoginJSON)
	testCases := []struct {
		name   string
		path   string
		method string
		auth   bool
	}{
		{"GetDashboardStats", "/api/v1/dashboard/statistics", "GET", true},
		{"GetCases", "/api/v1/cases", "GET", true},
		{"GetClients", "/api/v1/clients", "GET", true},
		{"LoginRequest", "/api/v1/auth/login", "POST", false},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// 计时前确认一次真实可达；断言失败即终止该子项。
			var reqToken string
			var reqBody []byte
			if tc.auth {
				reqToken = token
			} else {
				reqBody = loginBody
			}
			if code, _ := benchRequest(app, reqToken, tc.method, tc.path, reqBody); code != 200 {
				b.Fatalf("precheck %s status=%d", tc.path, code)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				code, _ := benchRequest(app, reqToken, tc.method, tc.path, reqBody)
				if code != 200 {
					b.Errorf("%s status=%d", tc.path, code)
					return
				}
			}
		})
	}
}

// BenchmarkConcurrentRequests 并发请求真实 dashboard。固定 4 个并发档，
// RunParallel 内每次迭代新建 request（httptest.Request 不可复用于并发）。
func BenchmarkConcurrentRequests(b *testing.B) {
	b.Helper()
	fx := setupBenchRouter(b)
	app := fx.app
	token := fx.token

	for _, level := range []int{1, 10, 50, 100} {
		b.Run(fmt.Sprintf("Concurrency_%d", level), func(b *testing.B) {
			if code, _ := benchRequest(app, token, "GET", "/api/v1/dashboard/statistics", nil); code != 200 {
				b.Fatalf("precheck status=%d", code)
			}

			b.ResetTimer()
			b.SetParallelism(level)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					req := httptest.NewRequest("GET", "/api/v1/dashboard/statistics", nil)
					req.Header.Set("Authorization", "Bearer "+token)
					w := httptest.NewRecorder()
					app.ServeHTTP(w, req)
					if w.Code != 200 {
						b.Errorf("status=%d", w.Code)
						return
					}
				}
			})
		})
	}
}

// loginBenchUser 真实登录 POST /api/v1/auth/login，读取 data.token 并核对用户。
func loginBenchUser(app *gin.Engine, wantUserID uint) (string, error) {
	code, body := benchRequest(app, "", "POST", "/api/v1/auth/login", []byte(npBenchLoginJSON))
	if code != 200 {
		return "", fmt.Errorf("login failed status=%d", code)
	}
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				ID    uint   `json:"id"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("login response parse failed: %w", err)
	}
	if !resp.Success || resp.Data.Token == "" {
		return "", fmt.Errorf("login response missing token")
	}
	if resp.Data.User.ID != wantUserID || resp.Data.User.Email != npBenchEmail {
		return "", fmt.Errorf("login identity mismatch id=%d", resp.Data.User.ID)
	}
	return resp.Data.Token, nil
}

// benchRequest 真实 app 请求；失败只报状态码，不输出正文或 token。
func benchRequest(app *gin.Engine, token, method, path string, body []byte) (int, []byte) {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	return w.Code, w.Body.Bytes()
}

// verifyBenchRoutes 计时前预验证：3 个 GET 实体/统计 + 无 token 401。
func verifyBenchRoutes(app *gin.Engine, token string) error {
	for _, path := range []string{"/api/v1/cases", "/api/v1/clients", "/api/v1/dashboard/statistics"} {
		if code, _ := benchRequest(app, "", "GET", path, nil); code != 401 {
			return fmt.Errorf("no-token %s expect 401 got %d", path, code)
		}
	}
	checks := []struct{ path, marker string }{
		{"/api/v1/cases", npBenchTag + "case"},
		{"/api/v1/clients", npBenchTag + "client"},
	}
	for _, check := range checks {
		code, body := benchRequest(app, token, "GET", check.path, nil)
		if code != 200 || !bytes.Contains(body, []byte(check.marker)) {
			return fmt.Errorf("GET %s missing synthetic content status=%d", check.path, code)
		}
	}
	code, body := benchRequest(app, token, "GET", "/api/v1/dashboard/statistics", nil)
	if code != 200 {
		return fmt.Errorf("GET dashboard status=%d", code)
	}
	var stats struct {
		Data struct {
			TotalClients int `json:"total_clients"`
			TotalCases   int `json:"total_cases"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &stats); err != nil {
		return fmt.Errorf("dashboard parse failed: %w", err)
	}
	if stats.Data.TotalClients != 1 || stats.Data.TotalCases != 1 {
		return fmt.Errorf("unexpected dashboard counts clients=%d cases=%d", stats.Data.TotalClients, stats.Data.TotalCases)
	}
	return nil
}
