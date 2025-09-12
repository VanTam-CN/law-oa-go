package test

import (
	"database/sql/driver"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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