//go:build ignore
// +build ignore

package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func TestBaseRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
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

func TestBaseRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	// 创建测试用户
	testUser := createTestUser(db)

	// 测试查找存在的用户
	foundUser, err := repo.GetByID(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, testUser.Name, foundUser.Name)

	// 测试查找不存在的用户
	_, err = repo.GetByID(ctx, 999999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "record not found")
}

func TestBaseRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 测试更新用户
	updates := map[string]interface{}{
		"name":  "更新后的用户名",
		"email": "updated@example.com",
	}

	err := repo.Update(ctx, testUser.ID, updates)
	assert.NoError(t, err)

	// 验证更新是否成功
	var updatedUser models.User
	err = db.First(&updatedUser, testUser.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "更新后的用户名", updatedUser.Name)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

func TestBaseRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
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
	assert.Contains(t, err.Error(), "record not found")
}

func TestBaseRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	// 创建多个测试用户
	for i := 0; i < 5; i++ {
		user := &models.User{
			Name:      "用户" + string(rune('A'+i)),
			Email:     "user" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	// 测试列表查询
	users, total, err := repo.List(ctx, 0, 3, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 3)

	// 测试带条件的列表查询
	conditions := map[string]interface{}{
		"role": "user",
	}
	users, total, err = repo.List(ctx, 0, 10, conditions)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, users, 5)
}

func TestBaseRepository_Count(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	// 创建测试用户
	for i := 0; i < 3; i++ {
		user := &models.User{
			Name:      "计数用户" + string(rune('A'+i)),
			Email:     "count" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	// 测试计数
	count, err := repo.Count(ctx, nil)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// 测试带条件的计数
	conditions := map[string]interface{}{
		"status": "active",
	}
	count, err = repo.Count(ctx, conditions)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestBaseRepository_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	// 创建多个用户用于批量插入
	var users []*models.User
	for i := 0; i < 5; i++ {
		user := &models.User{
			Name:      "批量用户" + string(rune('A'+i)),
			Email:     "batch" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		users = append(users, user)
	}

	// 测试批量创建
	err := repo.BatchCreate(ctx, users, 2)
	assert.NoError(t, err)

	// 验证批量创建是否成功
	var count int64
	err = db.Model(&models.User{}).Where("name LIKE ?", "批量用户%").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestBaseRepository_FindWithPreload(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	testUser := createTestUser(db)

	// 注意：这个测试假设 User 模型有预加载的关联
	// 如果没有，这个测试会跳过预加载部分
	foundUser, err := repo.FindWithPreload(ctx, testUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, testUser.ID, foundUser.ID)
}

func TestBaseRepository_Transaction(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	repo := NewBaseRepository[models.User](db)
	ctx := context.Background()

	// 测试事务操作
	err := repo.Transaction(ctx, func(tx *gorm.DB) error {
		// 在事务中创建用户
		user := &models.User{
			Name:      "事务用户",
			Email:     "transaction@example.com",
			Password:  "password",
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return tx.Create(user).Error
	})
	assert.NoError(t, err)

	// 验证事务中的操作是否成功
	var transactionUser models.User
	err = db.Where("email = ?", "transaction@example.com").First(&transactionUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "事务用户", transactionUser.Name)
}
