//go:build ignore
// +build ignore

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// MockDB 用于测试的模拟数据库
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Create(value interface{}) *gorm.DB {
	ret := m.Called(value)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) First(dest interface{}, conds ...interface{}) *gorm.DB {
	ret := m.Called(dest, conds)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Where(query interface{}, args ...interface{}) *gorm.DB {
	ret := m.Called(query, args)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Save(value interface{}) *gorm.DB {
	ret := m.Called(value)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Delete(value interface{}, conds ...interface{}) *gorm.DB {
	ret := m.Called(value, conds)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Model(value interface{}) *gorm.DB {
	ret := m.Called(value)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Count(count *int64) *gorm.DB {
	ret := m.Called(count)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Offset(offset int) *gorm.DB {
	ret := m.Called(offset)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Limit(limit int) *gorm.DB {
	ret := m.Called(limit)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Order(value interface{}) *gorm.DB {
	ret := m.Called(value)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Find(dest interface{}, conds ...interface{}) *gorm.DB {
	ret := m.Called(dest, conds)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) Updates(values interface{}) *gorm.DB {
	ret := m.Called(values)
	return ret.Get(0).(*gorm.DB)
}

func (m *MockDB) WithContext(ctx context.Context) *gorm.DB {
	ret := m.Called(ctx)
	return ret.Get(0).(*gorm.DB)
}

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移模型
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	if err != nil {
		t.Fatalf("Failed to migrate models: %v", err)
	}

	return db
}

// teardownTestDB 清理测试数据库
func teardownTestDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// createTestUser 创建测试用户
func createTestUser(db *gorm.DB) *models.User {
	return createTestUserWithSuffix(db, "")
}

// createTestUserWithSuffix 创建带后缀的测试用户
func createTestUserWithSuffix(db *gorm.DB, suffix string) *models.User {
	email := "test@example.com"
	if suffix != "" {
		email = "test" + suffix + "@example.com"
	}

	user := &models.User{
		Name:      "测试用户" + suffix,
		Email:     email,
		Password:  "hashed_password",
		Role:      "user",
		Phone:     "13800138000",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(user).Error; err != nil {
		panic(err)
	}

	return user
}

// createTestClient 创建测试客户
func createTestClient(db *gorm.DB) *models.Client {
	return createTestClientWithSuffix(db, "")
}

// createTestClientWithSuffix 创建带后缀的测试客户
func createTestClientWithSuffix(db *gorm.DB, suffix string) *models.Client {
	email := "client@example.com"
	if suffix != "" {
		email = "client" + suffix + "@example.com"
	}

	client := &models.Client{
		Name:      "测试客户" + suffix,
		Email:     email,
		Phone:     "13900139000",
		Address:   "测试地址" + suffix,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(client).Error; err != nil {
		panic(err)
	}

	return client
}

// createTestCase 创建测试案件
func createTestCase(db *gorm.DB, clientID, lawyerID uint) *models.Case {
	caseModel := &models.Case{
		Title:       "测试案件",
		CaseType:    "民事",
		Priority:    "normal",
		Status:      "pending",
		ClientID:    clientID,
		LawyerID:    lawyerID,
		Description: "测试案件描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Create(caseModel).Error; err != nil {
		panic(err)
	}

	return caseModel
}
