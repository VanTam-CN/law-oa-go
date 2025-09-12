package mock

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MockDB 模拟数据库连接
type MockDB struct {
	DB    *gorm.DB
	Mock  sqlmock.Sqlmock
	Close func()
}

// NewMockDB 创建模拟数据库
func NewMockDB() (*MockDB, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &MockDB{
		DB:   gormDB,
		Mock: mock,
		Close: func() {
			db.Close()
		},
	}, nil
}

// MockUserService 模拟用户服务
type MockUserService struct {
	mock.Mock
}

// NewMockUserService 创建模拟用户服务
func NewMockUserService() *MockUserService {
	return &MockUserService{}
}

// CreateUser 模拟创建用户
func (m *MockUserService) CreateUser(ctx interface{}, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// GetUserByID 模拟获取用户
func (m *MockUserService) GetUserByID(ctx interface{}, id uint) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

// UpdateUser 模拟更新用户
func (m *MockUserService) UpdateUser(ctx interface{}, id uint, req interface{}) (interface{}, error) {
	args := m.Called(ctx, id, req)
	return args.Get(0), args.Error(1)
}

// DeleteUser 模拟删除用户
func (m *MockUserService) DeleteUser(ctx interface{}, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCaseService 模拟案件服务
type MockCaseService struct {
	mock.Mock
}

// NewMockCaseService 创建模拟案件服务
func NewMockCaseService() *MockCaseService {
	return &MockCaseService{}
}

// CreateCase 模拟创建案件
func (m *MockCaseService) CreateCase(ctx interface{}, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// GetCaseByID 模拟获取案件
func (m *MockCaseService) GetCaseByID(ctx interface{}, id uint) (interface{}, error) {
	args := m.Called(ctx, id)
	return args.Get(0), args.Error(1)
}

// UpdateCase 模拟更新案件
func (m *MockCaseService) UpdateCase(id uint, req interface{}) (interface{}, error) {
	args := m.Called(id, req)
	return args.Get(0), args.Error(1)
}

// DeleteCase 模拟删除案件
func (m *MockCaseService) DeleteCase(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ListCases 模拟列出案件
func (m *MockCaseService) ListCases(req interface{}) ([]interface{}, int64, error) {
	args := m.Called(req)
	return args.Get(0).([]interface{}), args.Get(1).(int64), args.Error(2)
}

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

// NewMockAuthService 创建模拟认证服务
func NewMockAuthService() *MockAuthService {
	return &MockAuthService{}
}

// GenerateToken 模拟生成令牌
func (m *MockAuthService) GenerateToken(email, role string) (string, error) {
	args := m.Called(email, role)
	return args.String(0), args.Error(1)
}

// ValidateToken 模拟验证令牌
func (m *MockAuthService) ValidateToken(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// Login 模拟登录
func (m *MockAuthService) Login(ctx interface{}, req interface{}) (interface{}, error) {
	args := m.Called(ctx, req)
	return args.Get(0), args.Error(1)
}

// MockCacheService 模拟缓存服务
type MockCacheService struct {
	mock.Mock
}

// NewMockCacheService 创建模拟缓存服务
func NewMockCacheService() *MockCacheService {
	service := &MockCacheService{}
	// 预设一些默认行为
	service.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	service.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("cache miss"))
	service.On("Delete", mock.Anything, mock.Anything).Return(nil)
	service.On("Exists", mock.Anything, mock.Anything).Return(false, nil)
	service.On("Increment", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	return service
}

// Set 模拟设置缓存
func (m *MockCacheService) Set(ctx interface{}, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// Get 模拟获取缓存
func (m *MockCacheService) Get(ctx interface{}, key string, value interface{}) error {
	args := m.Called(ctx, key, value)

	// 如果有预设的返回值，设置到value中
	if args.Get(0) != nil {
		val := reflect.ValueOf(value).Elem()
		retVal := reflect.ValueOf(args.Get(0))
		if retVal.Kind() == reflect.Ptr {
			val.Set(retVal.Elem())
		} else {
			val.Set(retVal)
		}
	}

	return args.Error(1)
}

// Delete 模拟删除缓存
func (m *MockCacheService) Delete(ctx interface{}, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Exists 模拟检查缓存是否存在
func (m *MockCacheService) Exists(ctx interface{}, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

// GetClient 获取Redis客户端（Mock版本）
func (m *MockCacheService) GetClient() *redis.Client {
	// 返回一个简单的Mock Redis客户端
	return redis.NewClient(&redis.Options{
		Addr: "mock-redis:6379",
	})
}

// MockLogger 模拟日志器
type MockLogger struct {
	mock.Mock
	Entries []LogEntry
}

// LogEntry 日志条目
type LogEntry struct {
	Level   string
	Message string
	Fields  map[string]interface{}
	Time    time.Time
}

// NewMockLogger 创建模拟日志器
func NewMockLogger() *MockLogger {
	return &MockLogger{
		Entries: make([]LogEntry, 0),
	}
}

// Info 模拟信息日志
func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   "info",
		Message: msg,
		Fields:  parseFields(fields),
		Time:    time.Now(),
	})
}

// Error 模拟错误日志
func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   "error",
		Message: msg,
		Fields:  parseFields(fields),
		Time:    time.Now(),
	})
}

// Debug 模拟调试日志
func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   "debug",
		Message: msg,
		Fields:  parseFields(fields),
		Time:    time.Now(),
	})
}

// Warn 模拟警告日志
func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Entries = append(m.Entries, LogEntry{
		Level:   "warn",
		Message: msg,
		Fields:  parseFields(fields),
		Time:    time.Now(),
	})
}

// GetEntries 获取所有日志条目
func (m *MockLogger) GetEntries() []LogEntry {
	return m.Entries
}

// ClearEntries 清空日志条目
func (m *MockLogger) ClearEntries() {
	m.Entries = make([]LogEntry, 0)
}

// parseFields 解析日志字段
func parseFields(fields []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fields[i].(string)
			value := fields[i+1]
			result[key] = value
		}
	}
	return result
}

// TestHelper 测试助手
type TestHelper struct{}

// NewTestHelper 创建测试助手
func NewTestHelper() *TestHelper {
	return &TestHelper{}
}

// CreateMockUser 创建模拟用户数据
func (h *TestHelper) CreateMockUser(id uint, name, email, role string) map[string]interface{} {
	return map[string]interface{}{
		"id":         id,
		"name":       name,
		"email":      email,
		"role":       role,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateMockCase 创建模拟案件数据
func (h *TestHelper) CreateMockCase(id uint, title, description string, clientID, lawyerID uint) map[string]interface{} {
	return map[string]interface{}{
		"id":          id,
		"title":       title,
		"description": description,
		"client_id":   clientID,
		"lawyer_id":   lawyerID,
		"status":      "active",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
}

// ToJSONSlice 将接口切片转换为JSON字节切片
func (h *TestHelper) ToJSONSlice(data []interface{}) []driver.Value {
	result := make([]driver.Value, len(data))
	for i, item := range data {
		jsonData, err := json.Marshal(item)
		if err != nil {
			result[i] = nil
		} else {
			result[i] = jsonData
		}
	}
	return result
}

// MockGinContext 创建模拟Gin上下文
func (h *TestHelper) MockGinContext() interface{} {
	// 这里返回一个简单的模拟对象，实际使用时可以根据需要扩展
	return map[string]interface{}{
		"request_id": "test-request-id",
		"user_id":    uint(1),
		"user_role":  "admin",
	}
}
