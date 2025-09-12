package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/services"
	"law-oa-go/test"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserService_CreateUser(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	userService := services.NewUserService(testDB.DB)
	
	t.Run("successful user creation", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
			Role:     "lawyer",
		}
		
		// 模拟数据库查询
		testDB.Mock.ExpectBegin()
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnError(gorm.ErrRecordNotFound)
		
		testDB.Mock.ExpectExec("INSERT INTO `users`").
			WithArgs(
				sqlmock.AnyArg(), // ID
				req.Name,
				req.Email,
				sqlmock.AnyArg(), // Password (hashed)
				req.Role,
				"active",
				sqlmock.AnyArg(), // CreatedAt
				sqlmock.AnyArg(), // UpdatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		testDB.Mock.ExpectCommit()
		
		user, err := userService.CreateUser(context.Background(), req)
		
		require.NoError(t, err)
		assert.Equal(t, req.Name, user.Name)
		assert.Equal(t, req.Email, user.Email)
		assert.Equal(t, req.Role, user.Role)
		assert.Equal(t, "active", user.Status)
	})
	
	t.Run("duplicate email", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "Jane Doe",
			Email:    "jane@example.com",
			Password: "password123",
			Role:     "client",
		}
		
		// 模拟邮箱已存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		
		_, err := userService.CreateUser(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email already exists")
	})
	
	t.Run("invalid email format", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "Invalid User",
			Email:    "invalid-email",
			Password: "password123",
			Role:     "client",
		}
		
		_, err := userService.CreateUser(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
}

func TestUserService_AuthenticateUser(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	userService := services.NewUserService(testDB.DB)
	
	t.Run("successful authentication", func(t *testing.T) {
		email := "john@example.com"
		password := "correct_password"
		hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i" // bcrypt hash of "correct_password"
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "John Doe", email, hashedPassword, "lawyer", "active"))
		
		user, err := userService.AuthenticateUser(context.Background(), email, password)
		
		require.NoError(t, err)
		assert.Equal(t, uint(1), user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, email, user.Email)
	})
	
	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"
		password := "password123"
		
		// 模拟用户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(email).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := userService.AuthenticateUser(context.Background(), email, password)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
	
	t.Run("invalid password", func(t *testing.T) {
		email := "john@example.com"
		password := "wrong_password"
		hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i" // bcrypt hash of "correct_password"
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "John Doe", email, hashedPassword, "lawyer", "active"))
		
		_, err := userService.AuthenticateUser(context.Background(), email, password)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password")
	})
}

func TestUserService_GetUserProfile(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	userService := services.NewUserService(testDB.DB)
	
	t.Run("get existing user", func(t *testing.T) {
		userID := uint(1)
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role", "status", "created_at", "updated_at"}).
				AddRow(userID, "John Doe", "john@example.com", "lawyer", "active", test.TestTime(), test.TestTime()))
		
		user, err := userService.GetUserProfile(context.Background(), userID)
		
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Equal(t, "lawyer", user.Role)
		assert.Equal(t, "active", user.Status)
	})
	
	t.Run("user not found", func(t *testing.T) {
		userID := uint(999)
		
		// 模拟用户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := userService.GetUserProfile(context.Background(), userID)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	userService := services.NewUserService(testDB.DB)
	
	t.Run("update user successfully", func(t *testing.T) {
		userID := uint(1)
		req := &services.UpdateUserRequest{
			Name:  test.StringPtr("Updated Name"),
			Email: test.StringPtr("updated@example.com"),
			Role:  test.StringPtr("admin"),
		}
		
		// 模拟查询现有用户
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role", "status"}).
				AddRow(userID, "John Doe", "john@example.com", "lawyer", "active"))
		
		// 模拟邮箱唯一性检查
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ? AND id != ?").
			WithArgs(*req.Email, userID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		// 模拟更新
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(*req.Name, *req.Email, *req.Role, sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		user, err := userService.UpdateUser(context.Background(), userID, req)
		
		require.NoError(t, err)
		assert.Equal(t, *req.Name, user.Name)
		assert.Equal(t, *req.Email, user.Email)
		assert.Equal(t, *req.Role, user.Role)
	})
	
	t.Run("user not found", func(t *testing.T) {
		userID := uint(999)
		req := &services.UpdateUserRequest{
			Name: test.StringPtr("Updated Name"),
		}
		
		// 模拟用户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := userService.UpdateUser(context.Background(), userID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	userService := services.NewUserService(testDB.DB)
	
	t.Run("change password successfully", func(t *testing.T) {
		userID := uint(1)
		currentPassword := "old_password"
		newPassword := "new_password123"
		hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i" // bcrypt hash of "old_password"
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "John Doe", "john@example.com", hashedPassword, "lawyer", "active"))
		
		// 模拟更新密码
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		err := userService.ChangePassword(context.Background(), userID, currentPassword, newPassword)
		
		require.NoError(t, err)
	})
	
	t.Run("invalid current password", func(t *testing.T) {
		userID := uint(1)
		currentPassword := "wrong_password"
		newPassword := "new_password123"
		hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i" // bcrypt hash of "correct_password"
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "John Doe", "john@example.com", hashedPassword, "lawyer", "active"))
		
		err := userService.ChangePassword(context.Background(), userID, currentPassword, newPassword)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid current password")
	})
	
	t.Run("weak new password", func(t *testing.T) {
		userID := uint(1)
		currentPassword := "old_password"
		newPassword := "weak"
		hashedPassword := "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i"
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "John Doe", "john@example.com", hashedPassword, "lawyer", "active"))
		
		err := userService.ChangePassword(context.Background(), userID, currentPassword, newPassword)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password too weak")
	})
}

