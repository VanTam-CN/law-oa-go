//go:build ignore
// +build ignore

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"law-oa-go/internal/models"
)

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		Name:      "测试用户",
		Email:     "test@example.com",
		Password:  "hashed_password",
		Role:      "user",
		Phone:     "13800138000",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	// 验证用户确实被创建
	var foundUser models.User
	err = db.First(&foundUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Name, foundUser.Name)
	assert.Equal(t, user.Email, foundUser.Email)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// 创建第一个用户
	user1 := &models.User{
		Name:      "用户1",
		Email:     "same@example.com",
		Password:  "password1",
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, user1)
	assert.NoError(t, err)

	// 尝试创建相同邮箱的用户
	user2 := &models.User{
		Name:      "用户2",
		Email:     "same@example.com", // 相同邮箱
		Password:  "password2",
		Role:      "user",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, user2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already exists")
}

func TestUserRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 测试查找存在的用户
	foundUser, err := repo.FindByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, testUser.Name, foundUser.Name)
	assert.Equal(t, testUser.Email, foundUser.Email)

	// 测试查找不存在的用户
	_, err = repo.FindByID(ctx, 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 测试查找存在的邮箱
	foundUser, err := repo.FindByEmail(ctx, testUser.Email)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, testUser.ID, foundUser.ID)
	assert.Equal(t, testUser.Email, foundUser.Email)

	// 测试查找不存在的邮箱
	_, err = repo.FindByEmail(ctx, "nonexistent@example.com")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 更新用户信息
	updatedUser := &models.User{
		ID:        testUser.ID,
		Name:      "更新后的用户名",
		Email:     "updated@example.com",
		Password:  "new_hashed_password",
		Role:      "admin",
		Phone:     "13900139000",
		Avatar:    "new_avatar.jpg",
		Status:    "inactive",
		CreatedAt: testUser.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err := repo.Update(ctx, updatedUser)
	assert.NoError(t, err)

	// 验证更新是否成功
	var foundUser models.User
	err = db.First(&foundUser, testUser.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的用户名", foundUser.Name)
	assert.Equal(t, "updated@example.com", foundUser.Email)
	assert.Equal(t, "admin", foundUser.Role)
	assert.Equal(t, "inactive", foundUser.Status)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 测试删除用户
	err := repo.Delete(ctx, testUser.ID)
	assert.NoError(t, err)

	// 验证用户是否被删除
	var deletedUser models.User
	err = db.First(&deletedUser, testUser.ID).Error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")

	// 测试删除不存在的用户
	err = repo.Delete(ctx, 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// 创建多个测试用户
	testUsers := []*models.User{
		{Name: "用户A", Email: "usera@example.com", Role: "admin", Status: "active"},
		{Name: "用户B", Email: "userb@example.com", Role: "user", Status: "active"},
		{Name: "用户C", Email: "userc@example.com", Role: "user", Status: "inactive"},
		{Name: "用户D", Email: "userd@example.com", Role: "lawyer", Status: "active"},
	}

	for _, user := range testUsers {
		user.Password = "password"
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		db.Create(user)
	}

	// 测试分页列表
	params := &UserListParams{
		Page:     1,
		PageSize: 2,
	}

	users, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, users, 2)

	// 测试按角色筛选
	params = &UserListParams{
		Page:     1,
		PageSize: 10,
		Role:     "user",
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // userb 和 userc
	assert.Len(t, users, 2)

	// 测试按状态筛选
	params = &UserListParams{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total) // usera, userb, userd
	assert.Len(t, users, 3)

	// 测试搜索功能
	params = &UserListParams{
		Page:     1,
		PageSize: 10,
		Search:   "用户A",
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "用户A", users[0].Name)

	// 测试组合筛选
	params = &UserListParams{
		Page:     1,
		PageSize: 10,
		Role:     "user",
		Status:   "active",
		Search:   "user",
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total) // 只有 userb 符合条件
	assert.Len(t, users, 1)
	assert.Equal(t, "用户B", users[0].Name)
}

func TestUserRepository_List_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	params := &UserListParams{
		Page:     1,
		PageSize: 10,
	}

	users, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, users)
}

func TestUserRepository_List_InvalidPagination(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// 创建一个测试用户
	createTestUser(db)

	// 测试页码为0的情况
	params := &UserListParams{
		Page:     0, // 应该默认为1
		PageSize: 10,
	}

	users, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)

	// 测试页面大小为0的情况
	params = &UserListParams{
		Page:     1,
		PageSize: 0, // 应该默认为20
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
}

func TestUserRepository_List_SearchCaseInsensitive(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// 创建测试用户
	testUsers := []*models.User{
		{Name: "张三", Email: "zhangsan@example.com", Role: "user"},
		{Name: "李四", Email: "lisi@example.com", Role: "user"},
		{Name: "Wang Wu", Email: "wangwu@example.com", Role: "user"},
	}

	for _, user := range testUsers {
		user.Password = "password"
		user.Status = "active"
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		db.Create(user)
	}

	// 测试小写搜索
	params := &UserListParams{
		Page:     1,
		PageSize: 10,
		Search:   "zhang",
	}

	users, total, err := repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "张三", users[0].Name)

	// 测试大写搜索
	params = &UserListParams{
		Page:     1,
		PageSize: 10,
		Search:   "WANG",
	}

	users, total, err = repo.List(ctx, params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, "Wang Wu", users[0].Name)
}
