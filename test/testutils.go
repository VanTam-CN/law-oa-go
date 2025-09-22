package test

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
)

// TestDB 测试数据库
type TestDB struct {
	DB    *gorm.DB
	Mock  sqlmock.Sqlmock
	Close func()
}

// NewTestDB 创建测试数据库
func NewTestDB(t *testing.T) *TestDB {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to create gorm db: %v", err)
	}

	return &TestDB{
		DB:   gormDB,
		Mock: mock,
		Close: func() {
			db.Close()
		},
	}
}

// MockGinContext 创建模拟的Gin上下文
func MockGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := &testResponseWriter{}
	c, _ := gin.CreateTestContext(w)
	return c
}

type testResponseWriter struct {
	statusCode int
	body       []byte
	headers    map[string]string
}

func (w *testResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(map[string]string)
	}
	header := make(http.Header)
	for k, v := range w.headers {
		header[k] = []string{v}
	}
	return header
}

func (w *testResponseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return len(data), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *testResponseWriter) WriteString(s string) (int, error) {
	w.body = append(w.body, s...)
	return len(s), nil
}

// JSONDriver 用于模拟JSON字段的SQL驱动
type JSONDriver struct {
	data interface{}
}

func (j *JSONDriver) Value() (driver.Value, error) {
	return json.Marshal(j.data)
}

func (j *JSONDriver) Scan(src interface{}) error {
	bytes, ok := src.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, &j.data)
}

// AnyMatcher 匹配任何值
type AnyMatcher struct{}

func (a AnyMatcher) Match(v driver.Value) bool {
	return true
}

// TimeMatcher 匹配时间范围
type TimeMatcher struct {
	Start time.Time
	End   time.Time
}

func (t *TimeMatcher) Match(v driver.Value) bool {
	val, ok := v.(time.Time)
	if !ok {
		return false
	}
	return val.After(t.Start) && val.Before(t.End)
}

// NewTimeMatcher 创建新的时间匹配器
func NewTimeMatcher(duration time.Duration) *TimeMatcher {
	now := time.Now()
	return &TimeMatcher{
		Start: now.Add(-duration),
		End:   now.Add(duration),
	}
}

// Assert 包裹常用的断言函数
func Assert(t *testing.T) *assert.Assertions {
	return assert.New(t)
}

// Require 包裹常用的require函数
func Require(t *testing.T) *require.Assertions {
	return require.New(t)
}

// MockRows 创建模拟的数据库行
func MockRows(columns []string) *sqlmock.Rows {
	return sqlmock.NewRows(columns)
}

// TestTime 返回测试用的固定时间
func TestTime() time.Time {
	return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
}

// RandomString 生成随机字符串用于测试
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// RandomEmail 生成随机邮箱
func RandomEmail() string {
	return "test-" + RandomString(8) + "@example.com"
}

// StructToMap 将结构体转换为map
func StructToMap(obj interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

// AssertSuccessResponse 断言成功响应
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder) *require.Assertions {
	req := Require(t)
	req.Equal(http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	req.NoError(err)
	req.Equal(true, response["success"])

	return req
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T) *models.User {
	return &models.User{
		ID:        uint(1),
		Name:      "测试用户",
		Email:     "test@example.com",
		Password:  "$2a$10$hashedpassword", // bcrypt hash
		Role:      "user",
		Phone:     "1234567890",
		Status:    "active",
		CreatedAt: TestTime(),
		UpdatedAt: TestTime(),
	}
}

// CreateTestUserB 为基准测试创建测试用户
func CreateTestUserB(b *testing.B) *models.User {
	return &models.User{
		ID:        uint(1),
		Name:      "测试用户",
		Email:     "test@example.com",
		Password:  "$2a$10$hashedpassword", // bcrypt hash
		Role:      "user",
		Phone:     "1234567890",
		Status:    "active",
		CreatedAt: TestTime(),
		UpdatedAt: TestTime(),
	}
}

// AuthTokenCache 认证令牌缓存，避免重复数据库查询
var AuthTokenCache = make(map[string]string)

// GetAuthToken 获取认证令牌（带缓存）
func GetAuthToken(t *testing.T, userService *services.UserService, email, password string) string {
	// 检查缓存
	if token, exists := AuthTokenCache[email]; exists {
		return token
	}

	// 缓存未命中，执行认证
	_, err := userService.AuthenticateUser(context.Background(), email, password)
	Require(t).NoError(err)

	// 生成并缓存token
	token := "test-token-" + RandomString(16)
	AuthTokenCache[email] = token

	return token
}

// GetAuthTokenB 为基准测试获取认证令牌（带缓存）
func GetAuthTokenB(b *testing.B, userService *services.UserService, email, password string) string {
	// 检查缓存
	if token, exists := AuthTokenCache[email]; exists {
		return token
	}

	// 缓存未命中，执行认证
	_, err := userService.AuthenticateUser(context.Background(), email, password)
	if err != nil {
		b.Logf("Warning: Authentication failed: %v", err)
	}

	// 生成并缓存token
	token := "test-token-" + RandomString(16)
	AuthTokenCache[email] = token

	return token
}

// ClearAuthTokenCache 清理认证令牌缓存（测试清理用）
func ClearAuthTokenCache() {
	AuthTokenCache = make(map[string]string)
}

// PreloadTestAuthTokens 预加载测试认证令牌（性能测试专用）
func PreloadTestAuthTokens(t *testing.T, userService *services.UserService, emails []string) {
	for _, email := range emails {
		if _, exists := AuthTokenCache[email]; !exists {
			// 首先创建用户（如果不存在）
			createReq := &services.CreateUserRequest{
				Name:     "Test User " + email,
				Email:    email,
				Password: "Password123!",
				Role:     "user",
				Phone:    "1234567890",
			}

			// 尝试创建用户，忽略已存在错误
			_, err := userService.CreateUser(context.Background(), createReq)
			if err != nil && !strings.Contains(err.Error(), "already exists") {
				Require(t).NoError(err)
			}

			// 然后认证获取令牌
			_, err = userService.AuthenticateUser(context.Background(), email, "Password123!")
			Require(t).NoError(err)

			token := "test-token-" + email
			AuthTokenCache[email] = token
		}
	}
}

// CreateTestClient 创建测试客户
func CreateTestClient(t *testing.T) *models.Client {
	return &models.Client{
		ID:        uint(1),
		Name:      "测试客户",
		Email:     "client@example.com",
		Phone:     "9876543210",
		Address:   "测试地址",
		Company:   "测试公司",
		Status:    "active",
		CreatedAt: TestTime(),
		UpdatedAt: TestTime(),
	}
}
