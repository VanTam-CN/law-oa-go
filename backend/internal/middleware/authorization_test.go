package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// stubOK 用于授权测试：只返回 200，避免依赖数据库或业务逻辑。
// 测试重点是中间件链是否在角色不匹配时阻止请求。
func stubOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// runWithRole 构造一个带预设 role 的上下文，跳过 AuthMiddleware，
// 直接运行 RoleMiddleware + stubOK，用于隔离测试角色判定本身。
func runWithRole(t *testing.T, allowed []string, role string, hasRole bool) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if hasRole {
			c.Set("role", role)
		}
		c.Next()
	})
	r.GET("/x", RoleMiddleware(allowed...), stubOK)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRoleMiddleware_AllowsMatchingRole(t *testing.T) {
	cases := []struct {
		name    string
		allowed []string
		role    string
	}{
		{"admin in admin-only", []string{"admin", "super_admin"}, "admin"},
		{"super_admin in admin-only", []string{"admin", "super_admin"}, "super_admin"},
		{"finance in finance group", []string{"admin", "super_admin", "finance"}, "finance"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if code := runWithRole(t, tc.allowed, tc.role, true); code != http.StatusOK {
				t.Fatalf("expected 200 for role=%s allowed=%v, got %d", tc.role, tc.allowed, code)
			}
		})
	}
}

func TestRoleMiddleware_RejectsNonMatchingRole(t *testing.T) {
	cases := []struct {
		name    string
		allowed []string
		role    string
	}{
		{"lawyer hitting admin-only", []string{"admin", "super_admin"}, "lawyer"},
		{"assistant hitting admin-only", []string{"admin", "super_admin"}, "assistant"},
		{"lawyer hitting finance group", []string{"admin", "super_admin", "finance"}, "lawyer"},
		{"empty role string", []string{"admin"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if code := runWithRole(t, tc.allowed, tc.role, true); code != http.StatusForbidden {
				t.Fatalf("expected 403 for role=%q allowed=%v, got %d", tc.role, tc.allowed, code)
			}
		})
	}
}

func TestRoleMiddleware_FailsClosed_WhenRoleMissing(t *testing.T) {
	// 上下文里没有 role 键（例如 JWT 被篡改或中间件顺序错误），
	// 必须 fail-closed 返回 403，不能让请求通过。
	if code := runWithRole(t, []string{"admin"}, "", false); code != http.StatusForbidden {
		t.Fatalf("expected 403 when role key missing, got %d", code)
	}
}

func TestRoleMiddleware_RejectsNonStringRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", 12345) // 类型错误
		c.Next()
	})
	r.GET("/x", RoleMiddleware("admin"), stubOK)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when role is non-string, got %d", w.Code)
	}
}
