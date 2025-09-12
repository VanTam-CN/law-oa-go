package mock

import (
	"fmt"
	"time"
)

// TestDataFactory 测试数据工厂
type TestDataFactory struct{}

// NewTestDataFactory 创建测试数据工厂
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{}
}

// UserFactory 用户数据工厂
type UserFactory struct {
	factory *TestDataFactory
}

// Users 创建用户工厂
func (f *TestDataFactory) Users() *UserFactory {
	return &UserFactory{factory: f}
}

// CreateValidUser 创建有效用户数据
func (uf *UserFactory) CreateValidUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         uint(1),
		"name":       "测试用户",
		"email":      "test@example.com",
		"password":   "StrongPassword123!",
		"role":       "user",
		"status":     "active",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateAdminUser 创建管理员用户
func (uf *UserFactory) CreateAdminUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         uint(2),
		"name":       "管理员",
		"email":      "admin@example.com",
		"password":   "AdminPassword123!",
		"role":       "admin",
		"status":     "active",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateLawyerUser 创建律师用户
func (uf *UserFactory) CreateLawyerUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         uint(3),
		"name":       "张律师",
		"email":      "lawyer@example.com",
		"password":   "LawyerPassword123!",
		"role":       "lawyer",
		"status":     "active",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateInvalidUser 创建无效用户数据
func (uf *UserFactory) CreateInvalidUser() map[string]interface{} {
	return map[string]interface{}{
		"name":     "", // 空名称
		"email":    "invalid-email", // 无效邮箱
		"password": "123", // 弱密码
		"role":     "invalid_role", // 无效角色
	}
}

// CaseFactory 案件数据工厂
type CaseFactory struct {
	factory *TestDataFactory
}

// Cases 创建案件工厂
func (f *TestDataFactory) Cases() *CaseFactory {
	return &CaseFactory{factory: f}
}

// CreateValidCase 创建有效案件数据
func (cf *CaseFactory) CreateValidCase() map[string]interface{} {
	return map[string]interface{}{
		"id":          uint(1),
		"title":       "测试案件",
		"description": "这是一个测试案件描述",
		"client_id":   uint(1),
		"lawyer_id":   uint(1),
		"case_type":   "civil",
		"priority":    "medium",
		"status":      "active",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
}

// CreateHighPriorityCase 创建高优先级案件
func (cf *CaseFactory) CreateHighPriorityCase() map[string]interface{} {
	return map[string]interface{}{
		"id":          uint(2),
		"title":       "紧急案件",
		"description": "这是一个紧急案件描述",
		"client_id":   uint(1),
		"lawyer_id":   uint(1),
		"case_type":   "criminal",
		"priority":    "high",
		"status":      "active",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
}

// CreateClosedCase 创建已关闭案件
func (cf *CaseFactory) CreateClosedCase() map[string]interface{} {
	return map[string]interface{}{
		"id":          uint(3),
		"title":       "已关闭案件",
		"description": "这是一个已关闭案件描述",
		"client_id":   uint(1),
		"lawyer_id":   uint(1),
		"case_type":   "civil",
		"priority":    "low",
		"status":      "closed",
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}
}

// CreateInvalidCase 创建无效案件数据
func (cf *CaseFactory) CreateInvalidCase() map[string]interface{} {
	return map[string]interface{}{
		"title":     "", // 空标题
		"client_id": uint(0), // 无效客户端ID
		"lawyer_id": uint(0), // 无效律师ID
		"case_type": "invalid_type", // 无效案件类型
		"priority":  "invalid_priority", // 无效优先级
	}
}

// ClientFactory 客户数据工厂
type ClientFactory struct {
	factory *TestDataFactory
}

// Clients 创建客户工厂
func (f *TestDataFactory) Clients() *ClientFactory {
	return &ClientFactory{factory: f}
}

// CreateValidClient 创建有效客户数据
func (cf *ClientFactory) CreateValidClient() map[string]interface{} {
	return map[string]interface{}{
		"id":         uint(1),
		"name":       "测试客户公司",
		"contact":    "张三",
		"phone":      "13800138000",
		"email":      "client@company.com",
		"address":    "北京市朝阳区",
		"type":       "company",
		"status":     "active",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateIndividualClient 创建个人客户
func (cf *ClientFactory) CreateIndividualClient() map[string]interface{} {
	return map[string]interface{}{
		"id":         uint(2),
		"name":       "李四",
		"contact":    "李四",
		"phone":      "13900139000",
		"email":      "personal@email.com",
		"address":    "上海市浦东新区",
		"type":       "individual",
		"status":     "active",
		"created_at": time.Now(),
		"updated_at": time.Now(),
	}
}

// CreateInvalidClient 创建无效客户数据
func (cf *ClientFactory) CreateInvalidClient() map[string]interface{} {
	return map[string]interface{}{
		"name":    "", // 空名称
		"contact": "", // 空联系人
		"phone":   "123", // 无效电话
		"email":   "invalid-email", // 无效邮箱
		"type":    "invalid_type", // 无效类型
	}
}

// AuthFactory 认证数据工厂
type AuthFactory struct {
	factory *TestDataFactory
}

// Auth 创建认证工厂
func (f *TestDataFactory) Auth() *AuthFactory {
	return &AuthFactory{factory: f}
}

// CreateValidLoginRequest 创建有效登录请求数据
func (af *AuthFactory) CreateValidLoginRequest() map[string]interface{} {
	return map[string]interface{}{
		"email":    "test@example.com",
		"password": "StrongPassword123!",
	}
}

// CreateInvalidLoginRequest 创建无效登录请求数据
func (af *AuthFactory) CreateInvalidLoginRequest() map[string]interface{} {
	return map[string]interface{}{
		"email":    "invalid-email",
		"password": "123",
	}
}

// CreateValidToken 创建有效令牌数据
func (af *AuthFactory) CreateValidToken() string {
	return "valid.jwt.token"
}

// CreateExpiredToken 创建过期令牌数据
func (af *AuthFactory) CreateExpiredToken() string {
	return "expired.jwt.token"
}

// CreateInvalidToken 创建无效令牌数据
func (af *AuthFactory) CreateInvalidToken() string {
	return "invalid.token"
}

// CacheFactory 缓存数据工厂
type CacheFactory struct {
	factory *TestDataFactory
}

// Cache 创建缓存工厂
func (f *TestDataFactory) Cache() *CacheFactory {
	return &CacheFactory{factory: f}
}

// CreateValidCacheData 创建有效缓存数据
func (cf *CacheFactory) CreateValidCacheData() map[string]interface{} {
	return map[string]interface{}{
		"user_id":    uint(1),
		"user_name":  "测试用户",
		"user_role":  "user",
		"cached_at":  time.Now(),
		"expires_at": time.Now().Add(time.Hour),
	}
}

// CreateCacheKey 创建缓存键
func (cf *CacheFactory) CreateCacheKey(prefix string, id uint) string {
	return prefix + ":test:" + fmt.Sprintf("%d", id)
}

// CreateInvalidCacheKey 创建无效缓存键
func (cf *CacheFactory) CreateInvalidCacheKey() string {
	return ""
}

// BatchFactory 批量数据工厂
type BatchFactory struct {
	factory *TestDataFactory
}

// Batch 创建批量工厂
func (f *TestDataFactory) Batch() *BatchFactory {
	return &BatchFactory{factory: f}
}

// CreateUsers 创建多个用户
func (bf *BatchFactory) CreateUsers(count int) []map[string]interface{} {
	users := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		users[i] = map[string]interface{}{
			"id":         uint(i + 1),
			"name":       fmt.Sprintf("测试用户%d", i+1),
			"email":      fmt.Sprintf("user%d@example.com", i+1),
			"password":   "Password123!",
			"role":       "user",
			"status":     "active",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
	}
	return users
}

// CreateCases 创建多个案件
func (bf *BatchFactory) CreateCases(count int) []map[string]interface{} {
	cases := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		cases[i] = map[string]interface{}{
			"id":          uint(i + 1),
			"title":       fmt.Sprintf("测试案件%d", i+1),
			"description": fmt.Sprintf("这是第%d个测试案件", i+1),
			"client_id":   uint(1),
			"lawyer_id":   uint(1),
			"case_type":   "civil",
			"priority":    "medium",
			"status":      "active",
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		}
	}
	return cases
}

// CreateClients 创建多个客户
func (bf *BatchFactory) CreateClients(count int) []map[string]interface{} {
	clients := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		clients[i] = map[string]interface{}{
			"id":         uint(i + 1),
			"name":       fmt.Sprintf("测试客户%d", i+1),
			"contact":    fmt.Sprintf("联系人%d", i+1),
			"phone":      fmt.Sprintf("1380013%d", i+1),
			"email":      fmt.Sprintf("client%d@example.com", i+1),
			"address":    fmt.Sprintf("测试地址%d", i+1),
			"type":       "company",
			"status":     "active",
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}
	}
	return clients
}