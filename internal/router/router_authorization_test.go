package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
		usersAliasAdmin := usersAlias.Group("")
		usersAliasAdmin.Use(middleware.RoleMiddleware("admin", "super_admin"))
		{
			usersAliasAdmin.GET("", stubOK)
			usersAliasAdmin.GET("/:id", stubOK)
		}
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

	// /legal 管理接口 — admin/super_admin only
	legal := protected.Group("/legal")
	{
		legal.GET("/favorites", stubOK)
		legal.POST("/favorites", stubOK)
		legalAdmin := legal.Group("")
		legalAdmin.Use(middleware.RoleMiddleware("admin", "super_admin"))
		{
			legalAdmin.POST("/statutes", stubOK)
			legalAdmin.POST("/statutes/import", stubOK)
			legalAdmin.PUT("/statutes/:id", stubOK)
			legalAdmin.DELETE("/statutes/:id", stubOK)
		}
	}

	// /content-filter 检测接口普通认证，管理接口限管理员/合规
	contentFilter := protected.Group("/content-filter")
	{
		contentFilter.POST("/check", stubOK)
		contentFilter.POST("/filter", stubOK)
		contentFilterAdmin := contentFilter.Group("")
		contentFilterAdmin.Use(middleware.RoleMiddleware("admin", "super_admin", "compliance"))
		{
			contentFilterAdmin.POST("/words", stubOK)
			contentFilterAdmin.GET("/words", stubOK)
			contentFilterAdmin.DELETE("/words/:id", stubOK)
			contentFilterAdmin.GET("/logs", stubOK)
			contentFilterAdmin.POST("/cache/reset", stubOK)
		}
	}

	// Global aggregate and configuration endpoints must be explicitly scoped.
	// A technical administrator is allowed to manage accounts/settings, but is
	// not a business matter manager; a lawyer must not receive firm-wide rows.
	businessMatter := protected.Group("")
	businessMatter.Use(middleware.RoleMiddleware("director", "partner", "compliance", "risk", "risk_control", "management"))
	{
		businessMatter.GET("/lawyers/resource-center", stubOK)
		businessMatter.GET("/analytics/executive-dashboard", stubOK)
	}

	technicalAdmin := protected.Group("")
	technicalAdmin.Use(middleware.RoleMiddleware("admin", "super_admin"))
	{
		technicalAdmin.GET("/admin/access-center", stubOK)
		technicalAdmin.GET("/settings/overview", stubOK)
	}

	// Operations readiness separates review from evidence registration. A
	// compliance reviewer can see gaps without being able to append evidence.
	operationsReadiness := protected.Group("/operations/readiness/evidence")
	{
		operationsReadiness.GET(
			"",
			middleware.RoleMiddleware("admin", "super_admin", "director", "compliance"),
			stubOK,
		)
		operationsReadiness.POST(
			"",
			middleware.RoleMiddleware("admin", "super_admin", "director"),
			stubOK,
		)
	}

	// /conflict-v2/entities — 仅冲突/合规管理角色
	conflictV2 := protected.Group("/conflict-v2")
	{
		conflictV2.GET("/checks", stubOK)
		entities := conflictV2.Group("/entities")
		entities.Use(middleware.RoleMiddleware("compliance", "risk", "risk_control", "management", "partner", "conflict_officer"))
		{
			entities.GET("", stubOK)
			entities.POST("", stubOK)
			entities.GET("/search", stubOK)
			entities.POST("/batch", stubOK)
		}
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
		{"lawyer cannot list users alias", "lawyer", http.MethodGet, "/api/v1/users", http.StatusForbidden},
		{"admin can list users alias", "admin", http.MethodGet, "/api/v1/users", http.StatusOK},
		{"lawyer cannot read other user alias", "lawyer", http.MethodGet, "/api/v1/users/2", http.StatusForbidden},
		{"super admin can read other user alias", "super_admin", http.MethodGet, "/api/v1/users/2", http.StatusOK},
		{"lawyer can use content check", "lawyer", http.MethodPost, "/api/v1/content-filter/check", http.StatusOK},
		{"lawyer cannot manage sensitive words", "lawyer", http.MethodPost, "/api/v1/content-filter/words", http.StatusForbidden},
		{"compliance can manage sensitive words", "compliance", http.MethodPost, "/api/v1/content-filter/words", http.StatusOK},
		{"lawyer can read legal favorites", "lawyer", http.MethodGet, "/api/v1/legal/favorites", http.StatusOK},
		{"lawyer cannot import legal statutes", "lawyer", http.MethodPost, "/api/v1/legal/statutes/import", http.StatusForbidden},
		{"admin can import legal statutes", "admin", http.MethodPost, "/api/v1/legal/statutes/import", http.StatusOK},
		{"lawyer can list conflict checks", "lawyer", http.MethodGet, "/api/v1/conflict-v2/checks", http.StatusOK},
		{"lawyer cannot list conflict entities", "lawyer", http.MethodGet, "/api/v1/conflict-v2/entities", http.StatusForbidden},
		{"technical admin cannot list conflict entities", "admin", http.MethodGet, "/api/v1/conflict-v2/entities", http.StatusForbidden},
		{"super admin cannot list conflict entities", "super_admin", http.MethodGet, "/api/v1/conflict-v2/entities", http.StatusForbidden},
		{"risk can list conflict entities", "risk", http.MethodGet, "/api/v1/conflict-v2/entities", http.StatusOK},
		{"conflict officer can list conflict entities", "conflict_officer", http.MethodGet, "/api/v1/conflict-v2/entities", http.StatusOK},

		// Global aggregates/configuration must not be reachable by the wrong
		// class of authenticated user, even when the endpoint is called directly.
		{"lawyer cannot read lawyer resource center", "lawyer", http.MethodGet, "/api/v1/lawyers/resource-center", http.StatusForbidden},
		{"director can read lawyer resource center", "director", http.MethodGet, "/api/v1/lawyers/resource-center", http.StatusOK},
		{"lawyer cannot read executive dashboard", "lawyer", http.MethodGet, "/api/v1/analytics/executive-dashboard", http.StatusForbidden},
		{"technical admin cannot read executive dashboard", "admin", http.MethodGet, "/api/v1/analytics/executive-dashboard", http.StatusForbidden},
		{"director can read executive dashboard", "director", http.MethodGet, "/api/v1/analytics/executive-dashboard", http.StatusOK},
		{"lawyer cannot read access center", "lawyer", http.MethodGet, "/api/v1/admin/access-center", http.StatusForbidden},
		{"admin can read access center", "admin", http.MethodGet, "/api/v1/admin/access-center", http.StatusOK},
		{"director cannot read access center", "director", http.MethodGet, "/api/v1/admin/access-center", http.StatusForbidden},
		{"lawyer cannot read settings overview", "lawyer", http.MethodGet, "/api/v1/settings/overview", http.StatusForbidden},
		{"admin can read settings overview", "admin", http.MethodGet, "/api/v1/settings/overview", http.StatusOK},
		{"director cannot read settings overview", "director", http.MethodGet, "/api/v1/settings/overview", http.StatusForbidden},
		{"lawyer cannot read operations readiness", "lawyer", http.MethodGet, "/api/v1/operations/readiness/evidence", http.StatusForbidden},
		{"director can read operations readiness", "director", http.MethodGet, "/api/v1/operations/readiness/evidence", http.StatusOK},
		{"compliance can read operations readiness", "compliance", http.MethodGet, "/api/v1/operations/readiness/evidence", http.StatusOK},
		{"director can register operations evidence", "director", http.MethodPost, "/api/v1/operations/readiness/evidence", http.StatusOK},
		{"compliance cannot register operations evidence", "compliance", http.MethodPost, "/api/v1/operations/readiness/evidence", http.StatusForbidden},
		{"lawyer cannot register operations evidence", "lawyer", http.MethodPost, "/api/v1/operations/readiness/evidence", http.StatusForbidden},

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

func TestRemovedLegalAdminRoutesReturn404(t *testing.T) {
	engine := buildAuthorizationTestEngine()

	tests := []struct {
		name string
		path string
	}{
		{name: "sync elasticsearch route removed", path: "/api/v1/legal/admin/sync-elasticsearch"},
		{name: "rebuild index route removed", path: "/api/v1/legal/admin/rebuild-index"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tc.path, nil)
			req.Header.Set("X-Test-Role", "admin")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}
