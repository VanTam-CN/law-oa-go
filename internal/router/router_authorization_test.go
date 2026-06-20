package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/middleware"
)

// stubOK 仅返回 200，用于把"中间件是否阻止"与"处理器业务逻辑"解耦。
// 所有授权测试只关心中间件链是否阻止请求；不关心数据库或业务结果。
func stubOK(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// buildAuthorizationTestEngine 构造一个与 router.go 中授权矩阵结构一致的
// 测试引擎：相同的路由组挂载相同的 RoleMiddleware，但处理器全部是 stubOK。
//
// 这里是 mirror 而不是调用 router.Init，因为 Init 依赖 db/redis/es 等
// 重型基础设施；授权矩阵的语义不依赖这些。
//
// 当 router.go 中的路由组挂载变化时，本函数必须同步更新——这是测试的契约。
func buildAuthorizationTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	protected := api.Group("")
	// 用桩 AuthMiddleware：直接把 header 中的 X-Test-Role 写入 context，
	// 模拟 JWT 已解析后的状态。X-Test-Role 为空时视为未认证。
	protected.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Test-Role")
		if role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			return
		}
		c.Set("role", role)
		c.Next()
	})

	// /admin/users — admin/super_admin only
	users := protected.Group("/admin/users")
	users.Use(middleware.RoleMiddleware("admin", "super_admin"))
	{
		users.GET("", stubOK)
		users.POST("", stubOK)
		users.GET("/:id", stubOK)
		users.PUT("/:id", stubOK)
		users.DELETE("/:id", stubOK)
		users.GET("/:id/roles", stubOK)
		users.POST("/:id/roles", stubOK)
	}

	// /users/me, /users/profile — 任意已认证用户
	usersAlias := protected.Group("/users")
	{
		usersAlias.GET("/me", stubOK)
		usersAlias.GET("/profile", stubOK)
		usersAlias.PUT("/profile", stubOK)
		usersAlias.POST("/change-password", stubOK)
		usersAlias.POST("/avatar", stubOK)
	}

	// /admin/roles — admin/super_admin only
	roles := protected.Group("/admin/roles")
	roles.Use(middleware.RoleMiddleware("admin", "super_admin"))
	{
		roles.GET("", stubOK)
		roles.POST("", stubOK)
		roles.PUT("/:id", stubOK)
		roles.DELETE("/:id", stubOK)
		roles.GET("/:id/permissions", stubOK)
		roles.POST("/:id/permissions", stubOK)
	}

	// /admin/permissions — admin/super_admin only
	permissions := protected.Group("/admin/permissions")
	permissions.Use(middleware.RoleMiddleware("admin", "super_admin"))
	{
		permissions.GET("", stubOK)
		permissions.POST("", stubOK)
		permissions.PUT("/:id", stubOK)
		permissions.DELETE("/:id", stubOK)
	}

	// /finance — admin/super_admin/finance
	finance := protected.Group("/finance")
	finance.Use(middleware.RoleMiddleware("admin", "super_admin", "finance"))
	{
		finance.GET("/overview", stubOK)
		finance.POST("/contracts", stubOK)
		finance.POST("/payments", stubOK)
		finance.POST("/invoices/:id/approve", stubOK)
	}

	// /trust — admin/super_admin/finance
	trust := protected.Group("/trust")
	trust.Use(middleware.RoleMiddleware("admin", "super_admin", "finance"))
	{
		trust.GET("/accounts", stubOK)
		trust.POST("/accounts", stubOK)
		trust.GET("/transactions", stubOK)
		trust.POST("/transactions", stubOK)
		trust.POST("/transactions/:id/approve", stubOK)
		trust.POST("/transactions/:id/reject", stubOK)
	}

	return r
}

// TestAuthorizationMatrix 表驱动：覆盖计划中明确列出的全部矩阵用例，
// 并补充 owner-or-self 路径（/users/me, /users/profile）允许任意已认证角色。
func TestAuthorizationMatrix(t *testing.T) {
	tests := []struct {
		name   string
		role   string // 空字符串模拟未认证
		method string
		path   string
		want   int
	}{
		// 计划明示用例
		{"lawyer cannot create users", "lawyer", http.MethodPost, "/api/v1/admin/users", http.StatusForbidden},
		{"assistant cannot assign roles", "assistant", http.MethodPost, "/api/v1/admin/users/2/roles", http.StatusForbidden},
		{"lawyer cannot approve trust tx", "lawyer", http.MethodPost, "/api/v1/trust/transactions/1/approve", http.StatusForbidden},
		{"finance can approve trust tx", "finance", http.MethodPost, "/api/v1/trust/transactions/1/approve", http.StatusOK},
		{"finance can read overview", "finance", http.MethodGet, "/api/v1/finance/overview", http.StatusOK},
		{"admin can delete users", "admin", http.MethodDelete, "/api/v1/admin/users/2", http.StatusOK},
		{"super_admin can delete users", "super_admin", http.MethodDelete, "/api/v1/admin/users/2", http.StatusOK},

		// 财务组补充：律师不能审批发票
		{"lawyer cannot approve invoices", "lawyer", http.MethodPost, "/api/v1/finance/invoices/1/approve", http.StatusForbidden},
		{"finance can approve invoices", "finance", http.MethodPost, "/api/v1/finance/invoices/1/approve", http.StatusOK},

		// RBAC 管理：仅管理员
		{"lawyer cannot create roles", "lawyer", http.MethodPost, "/api/v1/admin/roles", http.StatusForbidden},
		{"admin can create roles", "admin", http.MethodPost, "/api/v1/admin/roles", http.StatusOK},
		{"lawyer cannot assign permissions", "lawyer", http.MethodPost, "/api/v1/admin/permissions", http.StatusForbidden},

		// Owner-or-self：任意已认证用户
		{"lawyer can read own profile", "lawyer", http.MethodGet, "/api/v1/users/me", http.StatusOK},
		{"assistant can read own profile", "assistant", http.MethodGet, "/api/v1/users/profile", http.StatusOK},
		{"finance can change own password", "finance", http.MethodPost, "/api/v1/users/change-password", http.StatusOK},

		// 未认证：所有受保护资源都应 401
		{"anonymous cannot create users", "", http.MethodPost, "/api/v1/admin/users", http.StatusUnauthorized},
		{"anonymous cannot read finance overview", "", http.MethodGet, "/api/v1/finance/overview", http.StatusUnauthorized},
		{"anonymous cannot approve trust tx", "", http.MethodPost, "/api/v1/trust/transactions/1/approve", http.StatusUnauthorized},
		{"anonymous cannot read own profile", "", http.MethodGet, "/api/v1/users/me", http.StatusUnauthorized},
	}

	engine := buildAuthorizationTestEngine()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			if tc.role != "" {
				req.Header.Set("X-Test-Role", tc.role)
			}
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)

			if w.Code != tc.want {
				t.Fatalf("role=%q %s %s: expected %d, got %d (body=%s)",
					tc.role, tc.method, tc.path, tc.want, w.Code, w.Body.String())
			}
		})
	}
}
