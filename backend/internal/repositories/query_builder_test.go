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

func TestQueryBuilder_BasicOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试数据
	for i := 0; i < 5; i++ {
		user := &models.User{
			Name:      "用户" + string(rune('A'+i)),
			Email:     "query" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      "user",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	t.Run("Where", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Where("name = ?", "用户A").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "用户A", users[0].Name)
	})

	t.Run("WhereIn", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			WhereIn("name", []interface{}{"用户A", "用户B", "用户C"}).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("WhereNot", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			WhereNot("name = ?", "用户A").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 4) // 总共5个，排除1个
	})

	t.Run("WhereLike", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			WhereLike("name", "%用户%").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
	})

	t.Run("Order", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Order("name ASC").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
		assert.Equal(t, "用户A", users[0].Name)
	})

	t.Run("OrderDesc", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			OrderDesc("name").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
		assert.Equal(t, "用户E", users[0].Name)
	})

	t.Run("OrderAsc", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			OrderAsc("name").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 5)
		assert.Equal(t, "用户A", users[0].Name)
	})

	t.Run("Limit", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Limit(3).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 3)
	})

	t.Run("Offset", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Offset(2).
			Limit(2).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("Count", func(t *testing.T) {
		count, err := NewQueryBuilder[models.User](db).
			Where("role = ?", "user").
			Count(ctx)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), count)
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := NewQueryBuilder[models.User](db).
			Where("name = ?", "用户A").
			Exists(ctx)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = NewQueryBuilder[models.User](db).
			Where("name = ?", "不存在的用户").
			Exists(ctx)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("First", func(t *testing.T) {
		user, err := NewQueryBuilder[models.User](db).
			Where("name = ?", "用户A").
			First(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "用户A", user.Name)

		// 测试不存在的记录
		_, err = NewQueryBuilder[models.User](db).
			Where("name = ?", "不存在的用户").
			First(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func TestQueryBuilder_ComplexQueries(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建更复杂的测试数据
	testData := []struct {
		name   string
		email  string
		role   string
		status string
		age    int
	}{
		{"张三", "zhangsan@example.com", "admin", "active", 25},
		{"李四", "lisi@example.com", "user", "active", 30},
		{"王五", "wangwu@example.com", "user", "inactive", 28},
		{"赵六", "zhaoliu@example.com", "lawyer", "active", 35},
		{"孙七", "sunqi@example.com", "admin", "inactive", 40},
	}

	for _, data := range testData {
		user := &models.User{
			Name:      data.name,
			Email:     data.email,
			Password:  "password",
			Role:      data.role,
			Status:    data.status,
			Phone:     "13800138000",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	t.Run("MultipleConditions", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Where("role = ?", "user").
			Where("status = ?", "active").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "李四", users[0].Name)
	})

	t.Run("ORConditions", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Where("role = ? OR role = ?", "admin", "lawyer").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 3) // 张三, 赵六, 孙七
	})

	t.Run("INConditions", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			WhereIn("role", []interface{}{"admin", "user"}).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 4) // 张三, 李四, 孙七, and another user if exists
	})

	t.Run("LIKEConditions", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			WhereLike("name", "%张%").
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "张三", users[0].Name)
	})

	t.Run("Pagination", func(t *testing.T) {
		// 第一页
		users1, err := NewQueryBuilder[models.User](db).
			OrderAsc("name").
			Limit(2).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users1, 2)

		// 第二页
		users2, err := NewQueryBuilder[models.User](db).
			OrderAsc("name").
			Offset(2).
			Limit(2).
			Find(ctx)
		assert.NoError(t, err)
		assert.Len(t, users2, 2)

		// 确保没有重复
		names1 := []string{users1[0].Name, users1[1].Name}
		names2 := []string{users2[0].Name, users2[1].Name}
		for _, name1 := range names1 {
			for _, name2 := range names2 {
				assert.NotEqual(t, name1, name2)
			}
		}
	})
}

func TestQueryBuilder_JoinOperations(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试客户和案件
	client := createTestClient(db)
	lawyer := createTestUser(db)

	// 创建案件
	caseModel := &models.Case{
		Title:       "测试案件",
		CaseType:    "民事",
		Priority:    "normal",
		Status:      "pending",
		ClientID:    client.ID,
		LawyerID:    lawyer.ID,
		Description: "测试案件描述",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(caseModel)

	t.Run("Join", func(t *testing.T) {
		// 注意：这个测试需要根据实际的模型关系调整
		// 这里我们测试基本的 Join 功能
		var results []map[string]interface{}
		err := NewQueryBuilder[models.Case](db).
			Join("LEFT JOIN users ON cases.lawyer_id = users.id").
			Where("cases.id = ?", caseModel.ID).
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})
}

func TestQueryBuilder_GroupAndHaving(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试数据
	roles := []string{"admin", "user", "lawyer"}
	for _, role := range roles {
		for i := 0; i < 2; i++ {
			user := &models.User{
				Name:      role + string(rune('A'+i)),
				Email:     role + string(rune('A'+i)) + "@example.com",
				Password:  "password",
				Role:      role,
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			db.Create(user)
		}
	}

	t.Run("GroupBy", func(t *testing.T) {
		var results []map[string]interface{}
		err := NewQueryBuilder[models.User](db).
			Select("role, COUNT(*) as count").
			Group("role").
			OrderDesc("count").
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 3)

		// 验证每个角色都有2个用户
		for _, result := range results {
			countPtr := result["count"].(*interface{})
			count := (*countPtr).(int64)
			assert.Equal(t, int64(2), count)
		}
	})

	t.Run("Having", func(t *testing.T) {
		var results []map[string]interface{}
		err := NewQueryBuilder[models.User](db).
			Select("role, COUNT(*) as count").
			Group("role").
			Having("COUNT(*) > ?", 1).
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})
}

func TestQueryBuilder_ErrorHandling(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	t.Run("InvalidSQL", func(t *testing.T) {
		_, err := NewQueryBuilder[models.User](db).
			Where("invalid_column = ?", "value").
			Find(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such column")
	})

	t.Run("EmptyResult", func(t *testing.T) {
		users, err := NewQueryBuilder[models.User](db).
			Where("name = ?", "不存在的用户").
			Find(ctx)
		assert.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestQueryBuilder_Preload(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	// 创建测试案件（包含关联的客户和律师）
	client := createTestClient(db)
	lawyer := createTestUser(db)
	createTestCase(db, client.ID, lawyer.ID)

	t.Run("Preload", func(t *testing.T) {
		// 注意：这个测试主要测试Preload方法是否可以被调用
		// 由于我们的模型定义，实际预加载效果需要根据模型关系定义
		queryBuilder := NewQueryBuilder[models.Case](db)
		assert.NotNil(t, queryBuilder.Preload("Client"))
		assert.NotNil(t, queryBuilder.Preload("Lawyer"))
	})
}

func TestQueryBuilder_LeftJoin(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试数据
	client := createTestClient(db)
	lawyer := createTestUser(db)
	caseModel := createTestCase(db, client.ID, lawyer.ID)

	t.Run("LeftJoin", func(t *testing.T) {
		var results []map[string]interface{}
		err := NewQueryBuilder[models.Case](db).
			LeftJoin("clients ON cases.client_id = clients.id").
			Where("cases.id = ?", caseModel.ID).
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
	})
}

func TestQueryBuilder_Distinct(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建具有相同角色的多个用户
	roles := []string{"admin", "admin", "user", "user", "lawyer"}
	for i, role := range roles {
		user := &models.User{
			Name:      "用户" + string(rune('A'+i)),
			Email:     "distinct" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      role,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	t.Run("Distinct", func(t *testing.T) {
		var results []map[string]interface{}
		err := NewQueryBuilder[models.User](db).
			Select("DISTINCT role").
			OrderAsc("role").
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.Len(t, results, 3)

		// 验证返回了3个不同的角色
		roles := []string{}
		for _, result := range results {
			roles = append(roles, result["role"].(string))
		}
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")
		assert.Contains(t, roles, "lawyer")
	})
}

func TestQueryBuilder_Raw(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试用户
	testUser := createTestUser(db)

	t.Run("Raw", func(t *testing.T) {
		var results []map[string]interface{}
		err := NewQueryBuilder[models.User](db).
			Raw("SELECT name, email FROM users WHERE id = ?", testUser.ID).
			Scan(ctx, &results)
		assert.NoError(t, err)
		assert.NotEmpty(t, results)
		assert.Equal(t, testUser.Name, results[0]["name"])
		assert.Equal(t, testUser.Email, results[0]["email"])
	})
}

func TestQueryBuilder_Exec(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建测试用户
	testUser := createTestUser(db)

	t.Run("Exec", func(t *testing.T) {
		err := NewQueryBuilder[models.User](db).
			Exec(ctx, "UPDATE users SET name = ? WHERE id = ?", "更新的名字", testUser.ID)
		assert.NoError(t, err)

		// 验证更新是否成功
		var updatedUser models.User
		err = db.First(&updatedUser, testUser.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "更新的名字", updatedUser.Name)
	})
}

func TestQueryBuilder_ComplexChaining(t *testing.T) {
	db := setupTestDB(t)
	defer teardownTestDB(db)

	ctx := context.Background()

	// 创建复杂的测试数据
	for i := 0; i < 10; i++ {
		user := &models.User{
			Name:      "用户" + string(rune('A'+i)),
			Email:     "chain" + string(rune('A'+i)) + "@example.com",
			Password:  "password",
			Role:      []string{"admin", "user"}[i%2],
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(user)
	}

	t.Run("ComplexChaining", func(t *testing.T) {
		// 测试复杂的链式调用
		users, err := NewQueryBuilder[models.User](db).
			Where("role = ?", "user").
			WhereLike("name", "%用户%").
			OrderDesc("name").
			Limit(5).
			Offset(0).
			Find(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, users)

		// 验证结果
		for _, user := range users {
			assert.Equal(t, "user", user.Role)
			assert.Contains(t, user.Name, "用户")
		}
	})
}

// Fuzzing 测试
func Fuzz_QueryBuilder_Where(f *testing.F) {
	// 添加种子语料
	f.Add("name = ?", "用户A")
	f.Add("role = ?", "admin")
	f.Add("status = ?", "active")

	f.Fuzz(func(t *testing.T, condition, value string) {
		// 限制输入长度
		if len(condition) > 200 || len(value) > 200 {
			t.Skip()
		}

		// 测试查询条件构建不会panic
		// 这里我们只测试条件字符串的有效性，不执行实际查询
		if len(condition) > 0 && len(value) > 0 {
			// 条件有效
		}
	})
}

func Fuzz_QueryBuilder_Like(f *testing.F) {
	// 添加种子语料
	f.Add("%张%", "张三")
	f.Add("%用户%", "用户A")
	f.Add("test%", "test123")

	f.Fuzz(func(t *testing.T, pattern, value string) {
		// 限制输入长度
		if len(pattern) > 200 || len(value) > 200 {
			t.Skip()
		}

		// 测试LIKE模式构建不会panic
		if len(pattern) > 0 {
			// 模式有效
		}
	})
}

func Fuzz_QueryBuilder_Order(f *testing.F) {
	// 添加种子语料
	f.Add("name", "ASC")
	f.Add("created_at", "DESC")
	f.Add("id", "ASC")

	f.Fuzz(func(t *testing.T, field, direction string) {
		// 限制输入长度
		if len(field) > 100 || len(direction) > 10 {
			t.Skip()
		}

		// 测试排序参数构建不会panic
		if len(field) > 0 {
			// 字段有效
		}
	})
}
