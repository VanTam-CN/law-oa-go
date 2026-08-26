package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	services "law-oa-go/internal/services"
	testmock "law-oa-go/test/mock"
)

func TestUserService_CreateUser(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("Create User Success", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
			Role:     "user",
			Phone:    "1234567890",
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, repositories.ErrUserNotFound)
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

		// 执行测试
		user, err := userService.CreateUser(context.Background(), req)

		// 断言结果
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "Test User", user.Name)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "user", user.Role)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Create User Email Already Exists", func(t *testing.T) {
		existingUser := &models.User{
			ID:     2,
			Email:  "existing@example.com",
			Status: "active",
		}

		req := &services.CreateUserRequest{
			Name:     "New User",
			Email:    "existing@example.com",
			Password: "Password123!",
			Role:     "user",
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

		// 执行测试
		user, err := userService.CreateUser(context.Background(), req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Email address already exists")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Create User Invalid Role", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
			Role:     "invalid_role",
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, repositories.ErrUserNotFound)

		// 执行测试
		user, err := userService.CreateUser(context.Background(), req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Invalid role")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Create User Weak Password", func(t *testing.T) {
		req := &services.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "weak",
			Role:     "user",
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, repositories.ErrUserNotFound)

		// 执行测试
		user, err := userService.CreateUser(context.Background(), req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Password too short")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Create User Database Error", func(t *testing.T) {
		mockUserRepo = new(testmock.MockUserRepository)
		userService = services.NewUserService(mockUserRepo)
		req := &services.CreateUserRequest{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "Password123!",
			Role:     "user",
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(nil, repositories.ErrUserNotFound)
		mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(assert.AnError)

		// 执行测试
		user, err := userService.CreateUser(context.Background(), req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Failed to create user")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_AuthenticateUser(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("Authenticate User Success", func(t *testing.T) {
		// 准备密码哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Username:  "test-account",
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(hashedPassword),
			Role:      "lawyer",
			Phone:     "1234567890",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		// 执行测试
		profile, err := userService.AuthenticateUser(context.Background(), "test@example.com", "Password123!")

		// 断言结果
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "Test User", profile.Name)
		assert.Equal(t, "test@example.com", profile.Email)
		assert.Equal(t, "test-account", profile.Username)
		assert.Equal(t, "lawyer", profile.Role)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Authenticate User Not Found", func(t *testing.T) {
		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, repositories.ErrUserNotFound)
		mockUserRepo.On("FindByUsername", mock.Anything, "nonexistent@example.com").Return(nil, repositories.ErrUserNotFound)

		// 执行测试
		profile, err := userService.AuthenticateUser(context.Background(), "nonexistent@example.com", "Password123!")

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "User not found")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Authenticate User Wrong Password", func(t *testing.T) {
		// 准备密码哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("CorrectPassword123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Username:  "test-account",
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(hashedPassword),
			Role:      "lawyer",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(user, nil)

		// 执行测试 - 错误密码
		profile, err := userService.AuthenticateUser(context.Background(), "test@example.com", "WrongPassword123!")

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "Invalid password")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Authenticate User By Username", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Username:  "lawyer.wang",
			Name:      "Wang Lawyer",
			Email:     "wang@example.com",
			Password:  string(hashedPassword),
			Role:      "lawyer",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockUserRepo.On("FindByEmail", mock.Anything, "Lawyer.Wang").Return(nil, repositories.ErrUserNotFound)
		mockUserRepo.On("FindByUsername", mock.Anything, "Lawyer.Wang").Return(user, nil)

		profile, err := userService.AuthenticateUser(context.Background(), "Lawyer.Wang", "Password123!")

		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "lawyer.wang", profile.Username)
		assert.Equal(t, "wang@example.com", profile.Email)
		assert.Equal(t, "lawyer", profile.Role)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Authenticate Inactive User", func(t *testing.T) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		inactiveUser := &models.User{
			ID:        1,
			Name:      "Inactive User",
			Email:     "inactive@example.com",
			Password:  string(hashedPassword),
			Role:      "user",
			Status:    "inactive",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByEmail", mock.Anything, "inactive@example.com").Return(inactiveUser, nil)

		// 执行测试
		profile, err := userService.AuthenticateUser(context.Background(), "inactive@example.com", "Password123!")

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "User not found")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserProfile(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("Get User Profile Success", func(t *testing.T) {
		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Role:      "lawyer",
			Phone:     "1234567890",
			Avatar:    "avatar.jpg",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil)

		// 执行测试
		profile, err := userService.GetUserProfile(context.Background(), 1)

		// 断言结果
		require.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "Test User", profile.Name)
		assert.Equal(t, "test@example.com", profile.Email)
		assert.Equal(t, "lawyer", profile.Role)
		assert.Equal(t, "1234567890", profile.Phone)
		assert.Equal(t, "avatar.jpg", profile.Avatar)
		assert.Equal(t, "active", profile.Status)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Get User Profile Not Found", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, repositories.ErrUserNotFound)

		// 执行测试
		profile, err := userService.GetUserProfile(context.Background(), 999)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("Update User Success", func(t *testing.T) {
		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Role:      "lawyer",
			Phone:     "1234567890",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		req := &services.UpdateUserRequest{
			Name:  stringPtr("Updated Name"),
			Email: stringPtr("newemail@example.com"),
			Phone: stringPtr("9876543210"),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil).Once()
		mockUserRepo.On("FindByEmail", mock.Anything, "newemail@example.com").Return(nil, repositories.ErrUserNotFound)
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil).Once()

		// 执行测试
		profile, err := userService.UpdateUser(context.Background(), 1, req)

		// 断言结果
		require.NoError(t, err)
		assert.NotNil(t, profile)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Update User Not Found", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		req := &services.UpdateUserRequest{
			Name: stringPtr("Updated Name"),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, repositories.ErrUserNotFound)

		// 执行测试
		profile, err := userService.UpdateUser(context.Background(), 999, req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "user not found")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Update User Email Conflict", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		existingUser := &models.User{
			ID:     2,
			Email:  "existing@example.com",
			Status: "active",
		}

		req := &services.UpdateUserRequest{
			Email: stringPtr("existing@example.com"),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
		mockUserRepo.On("FindByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)

		// 执行测试
		profile, err := userService.UpdateUser(context.Background(), 1, req)

		// 断言结果
		assert.Error(t, err)
		assert.Nil(t, profile)
		assert.Contains(t, err.Error(), "Email already exists")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_ChangePassword(t *testing.T) {
	t.Run("Change Password Success", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		// 准备当前密码哈希
		currentHashedPassword, err := bcrypt.GenerateFromPassword([]byte("CurrentPassword123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(currentHashedPassword),
			Role:      "lawyer",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil)
		mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

		// 执行测试
		err = userService.ChangePassword(context.Background(), 1, "CurrentPassword123!", "NewPassword123!")

		// 断言结果
		require.NoError(t, err)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Change Password Wrong Current Password", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		// 准备当前密码哈希
		currentHashedPassword, err := bcrypt.GenerateFromPassword([]byte("CurrentPassword123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(currentHashedPassword),
			Role:      "lawyer",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil)

		// 执行测试 - 错误的当前密码
		err = userService.ChangePassword(context.Background(), 1, "WrongPassword123!", "NewPassword123!")

		// 断言结果
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Current password is incorrect")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Change Password Weak New Password", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		currentHashedPassword, err := bcrypt.GenerateFromPassword([]byte("CurrentPassword123!"), bcrypt.DefaultCost)
		require.NoError(t, err)

		user := &models.User{
			ID:        1,
			Name:      "Test User",
			Email:     "test@example.com",
			Password:  string(currentHashedPassword),
			Role:      "lawyer",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 设置模拟期望
		mockUserRepo.On("FindByID", mock.Anything, uint(1)).Return(user, nil)

		// 执行测试 - 弱新密码
		err = userService.ChangePassword(context.Background(), 1, "CurrentPassword123!", "weak")

		// 断言结果
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Password too weak")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_ListUsers(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("List Users Success", func(t *testing.T) {
		// 准备测试用户列表
		users := []*models.User{
			{
				ID:        1,
				Name:      "User 1",
				Email:     "user1@example.com",
				Role:      "lawyer",
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        2,
				Name:      "User 2",
				Email:     "user2@example.com",
				Role:      "user",
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		req := &services.UserListRequest{
			Page:     1,
			PageSize: 10,
			Role:     "lawyer",
			Status:   "active",
			Search:   "user",
		}

		// 设置模拟期望
		mockUserRepo.On("List", mock.Anything, mock.AnythingOfType("*repositories.UserListParams")).
			Return(users, int64(2), nil)

		// 执行测试
		profiles, total, err := userService.ListUsers(context.Background(), req)

		// 断言结果
		require.NoError(t, err)
		assert.Len(t, profiles, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, "User 1", profiles[0].Name)
		assert.Equal(t, "User 2", profiles[1].Name)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("List Users Default Parameters", func(t *testing.T) {
		mockUserRepo := new(testmock.MockUserRepository)
		userService := services.NewUserService(mockUserRepo)

		// 准备测试用户列表
		users := []*models.User{
			{
				ID:        1,
				Name:      "User 1",
				Email:     "user1@example.com",
				Role:      "user",
				Status:    "active",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		req := &services.UserListRequest{} // 使用默认参数

		// 设置模拟期望
		mockUserRepo.On("List", mock.Anything, mock.AnythingOfType("*repositories.UserListParams")).
			Return(users, int64(1), nil)

		// 执行测试
		profiles, total, err := userService.ListUsers(context.Background(), req)

		// 断言结果
		require.NoError(t, err)
		assert.Len(t, profiles, 1)
		assert.Equal(t, int64(1), total)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	mockUserRepo := new(testmock.MockUserRepository)
	userService := services.NewUserService(mockUserRepo)

	t.Run("Delete User Success", func(t *testing.T) {
		// 设置模拟期望
		mockUserRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

		// 执行测试
		err := userService.DeleteUser(context.Background(), 1)

		// 断言结果
		require.NoError(t, err)

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Delete User Not Found", func(t *testing.T) {
		// 设置模拟期望
		mockUserRepo.On("Delete", mock.Anything, uint(999)).Return(repositories.ErrUserNotFound)

		// 执行测试
		err := userService.DeleteUser(context.Background(), 999)

		// 断言结果
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "User not found")

		// 验证模拟调用
		mockUserRepo.AssertExpectations(t)
	})
}

// 辅助函数：字符串指针
func stringPtr(s string) *string {
	return &s
}
