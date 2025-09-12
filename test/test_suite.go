package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

// TestSuite 测试套件
type TestSuite struct {
	DB           *gorm.DB
	Config       *config.TestConfig
	Router       *gin.Engine
	Helpers      *config.TestHelpers
	Cleanup      func()
	TestUser     *models.User
	TestLawyer   *models.Lawyer
	TestClient   *models.Client
	TestCase     *models.Case
	AuthToken    string
}

// NewTestSuite 创建新的测试套件
func NewTestSuite(t *testing.T) *TestSuite {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)
	
	// 加载测试配置
	cfg, err := config.LoadTestConfig()
	require.NoError(t, err, "Failed to load test config")
	
	// 设置测试数据库
	db, cleanup := config.SetupTestDB(t)
	
	// 创建测试路由
	router := setupTestRouter(db, cfg)
	
	// 创建测试辅助工具
	helpers := config.NewTestHelpers()
	
	// 创建测试套件
	suite := &TestSuite{
		DB:      db,
		Config:  cfg,
		Router:  router,
		Helpers: helpers,
		Cleanup: cleanup,
	}
	
	// 如果需要，创建测试数据
	if cfg.Test.SeedTestData {
		suite.SetupTestData(t)
	}
	
	return suite
}

// SetupTestData 设置测试数据
func (s *TestSuite) SetupTestData(t *testing.T) {
	// 创建测试用户
	s.TestUser = config.CreateTestUser(t, s.DB)
	
	// 创建测试律师
	s.TestLawyer = config.CreateTestLawyer(t, s.DB)
	
	// 创建测试客户
	s.TestClient = config.CreateTestClient(t, s.DB)
	
	// 创建测试案件
	s.TestCase = config.CreateTestCase(t, s.DB, s.TestLawyer.ID, s.TestClient.ID)
	
	// 生成认证令牌
	s.AuthToken = config.GenerateTestJWT(t, s.TestUser)
}

// Tearardown 拆卸测试套件
func (s *TestSuite) Tearardown() {
	if s.Cleanup != nil {
		s.Cleanup()
	}
}

// PerformRequest 执行HTTP请求
func (s *TestSuite) PerformRequest(method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyStr string
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyStr = string(bodyBytes)
	}
	
	req, _ := http.NewRequest(method, path, strings.NewReader(bodyStr))
	
	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// 默认设置Content-Type
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)
	
	return w
}

// PerformAuthRequest 执行需要认证的HTTP请求
func (s *TestSuite) PerformAuthRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + s.AuthToken,
	}
	return s.PerformRequest(method, path, body, headers)
}

// AssertSuccess 断言成功响应
func (s *TestSuite) AssertSuccess(t *testing.T, w *httptest.ResponseRecorder) interface{} {
	return s.Helpers.AssertSuccessResponse(t, w)
}

// AssertError 断言错误响应
func (s *TestSuite) AssertError(t *testing.T, w *httptest.ResponseRecorder, statusCode int) map[string]interface{} {
	return s.Helpers.AssertErrorResponse(t, w, statusCode)
}

// setupTestRouter 设置测试路由
func setupTestRouter(db *gorm.DB, cfg *config.TestConfig) *gin.Engine {
	router := gin.New()
	
	// 添加中间件
	router.Use(gin.Recovery())
	
	// 这里可以根据需要添加更多的中间件和路由
	// 简化的路由用于测试
	
	return router
}

// MockDatabase 模拟数据库
type MockDatabase struct {
	Calls    []MethodCall
	Error    error
	Response interface{}
}

// MethodCall 方法调用记录
type MethodCall struct {
	Method string
	Args   []interface{}
}

// NewMockDatabase 创建新的模拟数据库
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		Calls: make([]MethodCall, 0),
	}
}

// RecordCall 记录方法调用
func (m *MockDatabase) RecordCall(method string, args ...interface{}) {
	m.Calls = append(m.Calls, MethodCall{
		Method: method,
		Args:   args,
	})
}

// AssertCall 断言方法调用
func (m *MockDatabase) AssertCall(t *testing.T, method string, args ...interface{}) {
	found := false
	for _, call := range m.Calls {
		if call.Method == method {
			if len(call.Args) == len(args) {
				match := true
				for i, arg := range args {
					if call.Args[i] != arg {
						match = false
						break
					}
				}
				if match {
					found = true
					break
				}
			}
		}
	}
	
	assert.True(t, found, "Expected method call %s with args %v not found", method, args)
}

// ClearCalls 清除调用记录
func (m *MockDatabase) ClearCalls() {
	m.Calls = make([]MethodCall, 0)
}

// TestLogger 测试日志记录器
type TestLogger struct {
	Entries []LogEntry
}

// LogEntry 日志条目
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
	Time    time.Time
}

// NewTestLogger 创建新的测试日志记录器
func NewTestLogger() *TestLogger {
	return &TestLogger{
		Entries: make([]LogEntry, 0),
	}
}

// Log 记录日志
func (l *TestLogger) Log(level, message string, fields map[string]interface{}) {
	l.Entries = append(l.Entries, LogEntry{
		Level:   level,
		Message: message,
		Fields:  fields,
		Time:    time.Now(),
	})
}

// AssertLogEntry 断言日志条目
func (l *TestLogger) AssertLogEntry(t *testing.T, level, message string, fields map[string]interface{}) {
	found := false
	for _, entry := range l.Entries {
		if entry.Level == level && entry.Message == message {
			if fields == nil || len(fields) == 0 {
				found = true
				break
			}
			
			// 检查字段
			match := true
			for key, value := range fields {
				if entry.Fields[key] != value {
					match = false
					break
				}
			}
			if match {
				found = true
				break
			}
		}
	}
	
	assert.True(t, found, "Expected log entry with level %s and message %s not found", level, message)
}

// ClearEntries 清除日志条目
func (l *TestLogger) ClearEntries() {
	l.Entries = make([]LogEntry, 0)
}

// BenchmarkSuite 基准测试套件
type BenchmarkSuite struct {
	DB     *gorm.DB
	Config *config.TestConfig
}

// NewBenchmarkSuite 创建新的基准测试套件
func NewBenchmarkSuite(b *testing.B) *BenchmarkSuite {
	cfg, err := config.LoadTestConfig()
	require.NoError(b, err, "Failed to load test config")
	
	db, cleanup := config.SetupTestDB(b)
	
	b.Cleanup(cleanup)
	
	return &BenchmarkSuite{
		DB:     db,
		Config: cfg,
	}
}

// BenchmarkedOperation 基准测试操作
type BenchmarkedOperation struct {
	Name string
	Fn   func(b *testing.B)
}

// RunBenchmarks 运行基准测试
func (s *BenchmarkSuite) RunBenchmarks(b *testing.B, operations []BenchmarkedOperation) {
	for _, op := range operations {
		b.Run(op.Name, op.Fn)
	}
}

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	TestSuite
	RedisClient  *redis.Client
	ESClient     *elasticsearch.Client
	TestServer   *httptest.Server
	BaseURL      string
}

// NewIntegrationTestSuite 创建新的集成测试套件
func NewIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{}
	suite.TestSuite = *NewTestSuite(t)
	
	// 设置Redis客户端
	suite.RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
	
	// 设置Elasticsearch客户端
	suite.ESClient, _ = elasticsearch.NewDefaultClient()
	
	// 创建测试服务器
	suite.TestServer = httptest.NewServer(suite.Router)
	suite.BaseURL = suite.TestServer.URL
	
	// 设置清理函数
	originalCleanup := suite.Cleanup
	suite.Cleanup = func() {
		suite.TestServer.Close()
		suite.RedisClient.Close()
		suite.ESClient.Close()
		if originalCleanup != nil {
			originalCleanup()
		}
	}
	
	return suite
}

// PerformIntegrationRequest 执行集成测试请求
func (s *IntegrationTestSuite) PerformIntegrationRequest(method, path string, body interface{}, headers map[string]string) *http.Response {
	var bodyStr string
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyStr = string(bodyBytes)
	}
	
	req, _ := http.NewRequest(method, s.BaseURL+path, strings.NewReader(bodyStr))
	
	// 设置请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	// 默认设置Content-Type
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NotNil(t, resp, "Failed to perform integration request")
	
	return resp
}