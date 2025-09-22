package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// MockUserRepository 模拟用户仓库
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, params *repositories.UserListParams) ([]*models.User, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Count(ctx context.Context, params *repositories.UserListParams) (int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(int64), args.Error(1)
}

// MockClientRepository 模拟客户仓库
type MockClientRepository struct {
	mock.Mock
}

func (m *MockClientRepository) Create(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) FindByID(ctx context.Context, id uint) (*models.Client, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) FindByEmail(ctx context.Context, email string) (*models.Client, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Client), args.Error(1)
}

func (m *MockClientRepository) Update(ctx context.Context, client *models.Client) error {
	args := m.Called(ctx, client)
	return args.Error(0)
}

func (m *MockClientRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockClientRepository) List(ctx context.Context, params *repositories.ClientListParams) ([]*models.Client, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.Client), args.Get(1).(int64), args.Error(2)
}

func (m *MockClientRepository) Count(ctx context.Context, params *repositories.ClientListParams) (int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(int64), args.Error(1)
}

// MockCaseRepository 模拟案件仓库
type MockCaseRepository struct {
	mock.Mock
}

func (m *MockCaseRepository) Create(ctx context.Context, caseModel *models.Case) error {
	args := m.Called(ctx, caseModel)
	return args.Error(0)
}

func (m *MockCaseRepository) FindByID(ctx context.Context, id uint) (*models.Case, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Case), args.Error(1)
}

func (m *MockCaseRepository) Update(ctx context.Context, caseModel *models.Case) error {
	args := m.Called(ctx, caseModel)
	return args.Error(0)
}

func (m *MockCaseRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCaseRepository) List(ctx context.Context, params *repositories.CaseListParams) ([]*models.Case, int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]*models.Case), args.Get(1).(int64), args.Error(2)
}

func (m *MockCaseRepository) Count(ctx context.Context, params *repositories.CaseListParams) (int64, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(int64), args.Error(1)
}

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	require.NoError(t, err)

	return db
}

// CreateTestUser 创建测试用户
func CreateTestUser(t *testing.T, db *gorm.DB, email, name, role string) *models.User {
	user := &models.User{
		Name:     name,
		Email:    email,
		Password: "hashedPassword123!",
		Role:     role,
		Phone:    "1234567890",
		Status:   "active",
	}

	err := db.Create(user).Error
	require.NoError(t, err)

	return user
}

// CreateTestClient 创建测试客户
func CreateTestClient(t *testing.T, db *gorm.DB, name, email string) *models.Client {
	client := &models.Client{
		Name:    name,
		Email:   email,
		Phone:   "1234567890",
		Address: "123 Test St",
		Company: "Test Company",
		Notes:   "Test client notes",
		Status:  "active",
	}

	err := db.Create(client).Error
	require.NoError(t, err)

	return client
}

// CreateTestCase 创建测试案件
func CreateTestCase(t *testing.T, db *gorm.DB, title string, clientID, lawyerID uint) *models.Case {
	caseModel := &models.Case{
		Title:       title,
		Description: "Test case description",
		ClientID:    clientID,
		LawyerID:    lawyerID,
		CaseType:    "civil",
		Priority:    "medium",
		Status:      "pending",
	}

	err := db.Create(caseModel).Error
	require.NoError(t, err)

	return caseModel
}

// SetupTestRouter 设置测试路由器
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// MakeRequest 创建HTTP请求
func MakeRequest(method, path string, body interface{}, headers map[string]string) *http.Request {
	var reqBody bytes.Buffer
	if body != nil {
		json.NewEncoder(&reqBody).Encode(body)
	}

	req, _ := http.NewRequest(method, path, &reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// ExecuteRequest 执行HTTP请求并返回响应
func ExecuteRequest(router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// AssertSuccessResponse 断言成功响应
func AssertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatusCode int) {
	assert.Equal(t, expectedStatusCode, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["error"])
	assert.NotNil(t, response["data"])
}

// AssertErrorResponse 断言错误响应
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatusCode int) {
	assert.Equal(t, expectedStatusCode, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["error"])
	assert.Nil(t, response["data"])
}

// ParseResponseData 解析响应数据
func ParseResponseData(t *testing.T, w *httptest.ResponseRecorder, target interface{}) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	dataBytes, err := json.Marshal(response["data"])
	require.NoError(t, err)

	err = json.Unmarshal(dataBytes, target)
	require.NoError(t, err)
}

// GenerateTestToken 生成测试JWT token
func GenerateTestToken(userID uint, email, role string) string {
	// 这里使用简单的模拟token，实际测试中应该使用真实的JWT生成逻辑
	return fmt.Sprintf("fake_token_%d_%s_%s", userID, email, role)
}

// Wait 等待指定时间
func Wait(duration time.Duration) {
	time.Sleep(duration)
}

// Retry 重试函数
func Retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return err
}

// BenchmarkHelper 性能测试辅助函数
type BenchmarkHelper struct {
	Timer time.Time
}

func NewBenchmarkHelper() *BenchmarkHelper {
	return &BenchmarkHelper{
		Timer: time.Now(),
	}
}

func (bh *BenchmarkHelper) Start() {
	bh.Timer = time.Now()
}

func (bh *BenchmarkHelper) Duration() time.Duration {
	return time.Since(bh.Timer)
}

func (bh *BenchmarkHelper) LogDuration(operation string) {
	fmt.Printf("%s took %v\n", operation, bh.Duration())
}

// TestLogger 测试日志记录器
type TestLogger struct {
	logs []string
}

func NewTestLogger() *TestLogger {
	return &TestLogger{
		logs: make([]string, 0),
	}
}

func (l *TestLogger) Log(format string, args ...interface{}) {
	l.logs = append(l.logs, fmt.Sprintf(format, args...))
}

func (l *TestLogger) GetLogs() []string {
	return l.logs
}

func (l *TestLogger) Clear() {
	l.logs = make([]string, 0)
}

func (l *TestLogger) Contains(log string) bool {
	for _, l := range l.logs {
		if l == log {
			return true
		}
	}
	return false
}
