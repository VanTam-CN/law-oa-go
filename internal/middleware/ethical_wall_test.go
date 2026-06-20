package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"law-oa-go/internal/auth"
	"law-oa-go/internal/models"
)

// MockEthicalWallRepository 隔离墙仓储 mock
type MockEthicalWallRepository struct {
	mock.Mock
}

func (m *MockEthicalWallRepository) IsEthicalWallEnabled(ctx context.Context, caseID uint) (bool, error) {
	args := m.Called(ctx, caseID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEthicalWallRepository) EnableEthicalWall(ctx context.Context, caseID, userID uint, description string) error {
	return m.Called(ctx, caseID, userID, description).Error(0)
}

func (m *MockEthicalWallRepository) DisableEthicalWall(ctx context.Context, caseID uint) error {
	return m.Called(ctx, caseID).Error(0)
}

func (m *MockEthicalWallRepository) IsUserWhitelisted(ctx context.Context, caseID, userID uint) (bool, error) {
	args := m.Called(ctx, caseID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockEthicalWallRepository) AddToWhitelist(ctx context.Context, caseID, userID, grantedBy uint, reason string) error {
	return m.Called(ctx, caseID, userID, grantedBy, reason).Error(0)
}

func (m *MockEthicalWallRepository) RemoveFromWhitelist(ctx context.Context, caseID, userID uint) error {
	return m.Called(ctx, caseID, userID).Error(0)
}

func (m *MockEthicalWallRepository) GetWhitelistByCase(ctx context.Context, caseID uint) ([]*models.CaseEthicalWallWhitelist, error) {
	args := m.Called(ctx, caseID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CaseEthicalWallWhitelist), args.Error(1)
}

func (m *MockEthicalWallRepository) GetWhitelistByUser(ctx context.Context, userID uint) ([]*models.CaseEthicalWallWhitelist, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.CaseEthicalWallWhitelist), args.Error(1)
}

func (m *MockEthicalWallRepository) LogAccessAttempt(ctx context.Context, caseID, userID uint, accessType, accessResult, ipAddress, userAgent string) error {
	return m.Called(ctx, caseID, userID, accessType, accessResult, ipAddress, userAgent).Error(0)
}

func (m *MockEthicalWallRepository) GetAccessLogs(ctx context.Context, caseID uint, limit int) ([]*models.EthicalWallAccessLog, error) {
	args := m.Called(ctx, caseID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.EthicalWallAccessLog), args.Error(1)
}

func (m *MockEthicalWallRepository) ClearWhitelist(ctx context.Context, caseID uint) error {
	return m.Called(ctx, caseID).Error(0)
}

// MockDocumentResolver 文档→案件解析 mock，避免依赖数据库
type MockDocumentResolver struct {
	mock.Mock
}

func (m *MockDocumentResolver) ResolveDocumentCase(ctx context.Context, documentID uint) (caseID uint, applies bool, err error) {
	args := m.Called(ctx, documentID)
	return uint(args.Int(0)), args.Bool(1), args.Error(2)
}

// newTestEngine 构造测试引擎：预设认证 + 隔离墙中间件 + 用户自定义路由。
func newTestEngine(t *testing.T, cfg EthicalWallConfig, userID uint) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		auth.SetAuthContext(c, &auth.TokenClaims{
			UserID:   userID,
			Username: "tester",
			Email:    "tester@example.test",
			Role:     "lawyer",
		})
		c.Next()
	})
	r.Use(EthicalWallMiddleware(cfg))
	return r
}

// =====================================================================
// 计划 Step 1 明示要求覆盖的回归场景
// =====================================================================

// 计划: IsEthicalWallEnabled 返回错误 -> 503, next handler 未执行
func TestEthicalWall_FailClosed_OnEnabledCheckError(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(99)).Return(false, errors.New("db down"))

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	handlerCalled := false
	r.GET("/api/v1/cases/:id", func(c *gin.Context) { handlerCalled = true; c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code, "db 错误必须 fail-closed 返回 503")
	assert.False(t, handlerCalled, "fail-closed 时下游 handler 不得执行")
	repo.AssertExpectations(t)
}

// 计划: IsUserWhitelisted 返回错误 -> 503, next handler 未执行
func TestEthicalWall_FailClosed_OnWhitelistCheckError(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(99)).Return(true, nil)
	repo.On("IsUserWhitelisted", mock.Anything, uint(99), uint(123)).Return(false, errors.New("redis down"))

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	handlerCalled := false
	r.GET("/api/v1/cases/:id", func(c *gin.Context) { handlerCalled = true; c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.False(t, handlerCalled)
	repo.AssertExpectations(t)
}

// 计划: 无白名单 -> 403; 有白名单 -> 200
func TestEthicalWall_WhitelistDenies_Returns403(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(99)).Return(true, nil)
	repo.On("IsUserWhitelisted", mock.Anything, uint(99), uint(123)).Return(false, nil)
	repo.On("LogAccessAttempt", mock.Anything, uint(99), uint(123), "view", "denied", mock.Anything, mock.Anything).Return(nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/cases/:id", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestEthicalWall_WhitelistAllows_Returns200(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(99)).Return(true, nil)
	repo.On("IsUserWhitelisted", mock.Anything, uint(99), uint(123)).Return(true, nil)
	repo.On("LogAccessAttempt", mock.Anything, uint(99), uint(123), "view", "allowed", mock.Anything, mock.Anything).Return(nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/cases/:id", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/cases/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 计划: /documents/stats/overview 不得把 "overview" 或其他 :id 当案件 ID
func TestEthicalWall_StatsPath_NotTreatedAsCaseID(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/documents/stats/:section", func(c *gin.Context) { c.Status(200) })

	// "overview" 不能被传给 IsEthicalWallEnabled——它不是数字。
	// 即使被解析为 0（uint("overview")），中间件应当跳过。
	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/stats/overview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertNotCalled(t, "IsEthicalWallEnabled")
}

// 计划: 文档 ID=10，EntityType=case，EntityID=99；路由 /documents/10 必须检查 case 99
func TestEthicalWall_DocumentWithCaseEntity_ResolvesToCaseID(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(99)).Return(true, nil)
	repo.On("IsUserWhitelisted", mock.Anything, uint(99), uint(123)).Return(true, nil)
	repo.On("LogAccessAttempt", mock.Anything, uint(99), uint(123), "view", "allowed", mock.Anything, mock.Anything).Return(nil)

	docResolver := new(MockDocumentResolver)
	docResolver.On("ResolveDocumentCase", mock.Anything, uint(10)).Return(99, true, nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo, DocumentResolver: docResolver}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/documents/:id", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	docResolver.AssertExpectations(t)
	repo.AssertExpectations(t)
}

// 计划: 文档不关联案件（EntityType != "case"）-> applies=false，不调用白名单检查
func TestEthicalWall_DocumentWithoutCaseEntity_SkipsWhitelistCheck(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	docResolver := new(MockDocumentResolver)
	docResolver.On("ResolveDocumentCase", mock.Anything, uint(10)).Return(0, false, nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo, DocumentResolver: docResolver}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/documents/:id", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertNotCalled(t, "IsEthicalWallEnabled")
	docResolver.AssertExpectations(t)
}

// 计划: 文档→案件解析失败 -> 503
func TestEthicalWall_DocumentResolveError_Returns503(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	docResolver := new(MockDocumentResolver)
	docResolver.On("ResolveDocumentCase", mock.Anything, uint(10)).Return(0, false, errors.New("doc store unavailable"))

	cfg := EthicalWallConfig{EthicalWallRepo: repo, DocumentResolver: docResolver}
	r := newTestEngine(t, cfg, 123)
	handlerCalled := false
	r.GET("/api/v1/documents/:id", func(c *gin.Context) { handlerCalled = true; c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.False(t, handlerCalled)
}

// 计划: POST JSON 经中间件后 handler 仍可读取完整 body
func TestEthicalWall_PreservesRequestBody_ForDownstreamHandler(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	// case_id 在 body 中 -> enabled=false -> 通过
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(55)).Return(false, nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)

	var receivedBody map[string]interface{}
	r.POST("/api/v1/cases", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		_ = json.Unmarshal(body, &receivedBody)
		c.Status(200)
	})

	payload := `{"name":"新案件","case_id":55,"description":"隔离墙测试"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cases", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code, "handler 必须被调用")
	require.NotNil(t, receivedBody, "下游 handler 必须能读到 body")
	assert.Equal(t, "新案件", receivedBody["name"], "下游收到的 body 必须完整")
	assert.EqualValues(t, 55, receivedBody["case_id"])
}

// =====================================================================
// 边界场景
// =====================================================================

func TestEthicalWall_SkipPathShortCircuits(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	cfg := EthicalWallConfig{
		EthicalWallRepo: repo,
		SkipPaths:       []string{"/api/v1/ethical-wall/cases/:id/status"},
	}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/ethical-wall/cases/:id/status", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ethical-wall/cases/99/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertNotCalled(t, "IsEthicalWallEnabled")
}

func TestEthicalWall_QueryCaseID_TriggersCheck(t *testing.T) {
	repo := new(MockEthicalWallRepository)
	repo.On("IsEthicalWallEnabled", mock.Anything, uint(789)).Return(false, nil)

	cfg := EthicalWallConfig{EthicalWallRepo: repo}
	r := newTestEngine(t, cfg, 123)
	r.GET("/api/v1/documents", func(c *gin.Context) { c.Status(200) })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/documents?case_id=789", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

// TestExtractCaseID_PathAware 表驱动：不同 c.FullPath() 决定 :id 语义
func TestExtractCaseID_PathAware(t *testing.T) {
	tests := []struct {
		name     string
		fullPath string // c.FullPath() 模板
		route    string // 注册路由
		method   string
		target   string // 实际请求 URL
		wantCase uint
	}{
		{"cases/:id 直接是 case", "/api/v1/cases/:id", "/api/v1/cases/:id", http.MethodGet, "/api/v1/cases/456", 456},
		{"cases list 无 id", "/api/v1/cases", "/api/v1/cases", http.MethodGet, "/api/v1/cases", 0},
		{"documents/:id 不应直接当 case（需 resolver）", "/api/v1/documents/:id", "/api/v1/documents/:id", http.MethodGet, "/api/v1/documents/10", 0},
		{"documents/stats/:section 不应解析为 case", "/api/v1/documents/stats/:section", "/api/v1/documents/stats/:section", http.MethodGet, "/api/v1/documents/stats/overview", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			var gotCase uint
			r.Use(func(c *gin.Context) {
				gotCase = extractCaseIDForPath(c, tc.fullPath)
				c.Next()
			})
			r.Handle(tc.method, tc.route, func(c *gin.Context) { c.Status(200) })

			req := httptest.NewRequest(tc.method, tc.target, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tc.wantCase, gotCase)
		})
	}
}

// 保留对 getAccessType 的最小回归（不动业务逻辑）
func TestGetAccessType(t *testing.T) {
	cases := []struct {
		method, path, want string
	}{
		{http.MethodGet, "/api/v1/cases/123", "view"},
		{http.MethodPost, "/api/v1/cases", "modify"},
		{http.MethodGet, "/api/v1/cases/123/export", "export"},
		{http.MethodGet, "/api/v1/search", "search"},
	}
	for _, tc := range cases {
		t.Run(tc.want+"_"+tc.method, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.Use(func(c *gin.Context) {
				assert.Equal(t, tc.want, getAccessType(c))
				c.AbortWithStatus(200)
			})
			r.Handle(tc.method, "/*any", func(c *gin.Context) {})
			req := httptest.NewRequest(tc.method, tc.path, nil)
			r.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}

// 防止 lint 报未使用 import（当某些断言路径无 require 时仍保留 require 引用）
var _ = require.New
var _ = strings.HasPrefix
