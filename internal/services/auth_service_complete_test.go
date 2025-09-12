package services_test

import (
	"context"
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

func TestAuthService_Login(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	authService := services.NewAuthService(testDB.DB)
	
	t.Run("successful login", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "correct_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", req.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新最后登录时间
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		resp, err := authService.Login(context.Background(), req)
		
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, uint(1), resp.User.ID)
		assert.Equal(t, "张律师", resp.User.Name)
		assert.Equal(t, "lawyer", resp.User.Role)
		assert.NotEmpty(t, resp.ExpiresAt)
		assert.True(t, resp.ExpiresAt.After(time.Now()))
	})
	
	t.Run("user not found", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}
		
		// 模拟用户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err := authService.Login(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
	
	t.Run("invalid password", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "wrong_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", req.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		_, err := authService.Login(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})
	
	t.Run("user inactive", func(t *testing.T) {
		req := &services.LoginRequest{
			Email:    "inactive@example.com",
			Password: "password123",
		}
		
		// 模拟用户状态为非活跃
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(req.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "未激活用户", req.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "inactive"))
		
		_, err := authService.Login(context.Background(), req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user account is inactive")
	})
}

func TestAuthService_ValidateToken(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	authService := services.NewAuthService(testDB.DB)
	
	t.Run("valid token", func(t *testing.T) {
		// 首先登录获取token
		loginReq := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "correct_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(loginReq.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", loginReq.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新最后登录时间
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		loginResp, err := authService.Login(context.Background(), loginReq)
		require.NoError(t, err)
		
		// 验证token
		claims, err := authService.ValidateToken(loginResp.Token)
		
		require.NoError(t, err)
		assert.Equal(t, uint(1), claims.UserID)
		assert.Equal(t, "lawyer@example.com", claims.Email)
		assert.Equal(t, "张律师", claims.Name)
		assert.Equal(t, "lawyer", claims.Role)
	})
	
	t.Run("invalid token format", func(t *testing.T) {
		invalidToken := "invalid.jwt.token"
		
		_, err := authService.ValidateToken(invalidToken)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token format")
	})
	
	t.Run("expired token", func(t *testing.T) {
		// 创建一个已经过期的token（这里模拟）
		expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE2MDAwMDAwMDB9.invalid_signature"
		
		_, err := authService.ValidateToken(expiredToken)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "token has expired")
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	authService := services.NewAuthService(testDB.DB)
	
	t.Run("successful token refresh", func(t *testing.T) {
		// 首先登录获取token
		loginReq := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "correct_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(loginReq.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", loginReq.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新最后登录时间
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		loginResp, err := authService.Login(context.Background(), loginReq)
		require.NoError(t, err)
		
		// 刷新token
		refreshReq := &services.RefreshTokenRequest{
			Token: loginResp.Token,
		}
		
		// 模拟用户查询（验证用户仍然存在且状态正常）
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(uint(1)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "role", "status"}).
				AddRow(1, "张律师", "lawyer@example.com", "lawyer", "active"))
		
		refreshResp, err := authService.RefreshToken(context.Background(), refreshReq)
		
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.Token)
		assert.NotEqual(t, loginResp.Token, refreshResp.Token)
		assert.Equal(t, uint(1), refreshResp.User.ID)
		assert.True(t, refreshResp.ExpiresAt.After(time.Now()))
	})
	
	t.Run("invalid token refresh", func(t *testing.T) {
		refreshReq := &services.RefreshTokenRequest{
			Token: "invalid.token.here",
		}
		
		_, err := authService.RefreshToken(context.Background(), refreshReq)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
	
	t.Run("user not found during refresh", func(t *testing.T) {
		// 首先登录获取token
		loginReq := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "correct_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(loginReq.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", loginReq.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新最后登录时间
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		loginResp, err := authService.Login(context.Background(), loginReq)
		require.NoError(t, err)
		
		// 刷新token，但用户已不存在
		refreshReq := &services.RefreshTokenRequest{
			Token: loginResp.Token,
		}
		
		// 模拟用户不存在
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(uint(1)).
			WillReturnError(gorm.ErrRecordNotFound)
		
		_, err = authService.RefreshToken(context.Background(), refreshReq)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestAuthService_Logout(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	authService := services.NewAuthService(testDB.DB)
	
	t.Run("successful logout", func(t *testing.T) {
		// 首先登录获取token
		loginReq := &services.LoginRequest{
			Email:    "lawyer@example.com",
			Password: "correct_password",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE email = ?").
			WithArgs(loginReq.Email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(1, "张律师", loginReq.Email, "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新最后登录时间
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(sqlmock.AnyArg(), 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		loginResp, err := authService.Login(context.Background(), loginReq)
		require.NoError(t, err)
		
		// 登出
		logoutReq := &services.LogoutRequest{
			Token: loginResp.Token,
		}
		
		// 模拟将token加入黑名单
		testDB.Mock.ExpectExec("INSERT INTO `token_blacklist`").
			WithArgs(
				sqlmock.AnyArg(), // ID
				logoutReq.Token,
				sqlmock.AnyArg(), // ExpiresAt
				sqlmock.AnyArg(), // CreatedAt
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		err = authService.Logout(context.Background(), logoutReq)
		
		require.NoError(t, err)
	})
	
	t.Run("logout with invalid token", func(t *testing.T) {
		logoutReq := &services.LogoutRequest{
			Token: "invalid.token.here",
		}
		
		// 即使token无效，登出操作也应该成功（或至少不报错）
		err := authService.Logout(context.Background(), logoutReq)
		
		// 根据具体实现，可能返回错误或成功
		// 这里假设即使token无效也认为是成功的登出
		if err != nil {
			assert.Contains(t, err.Error(), "invalid token")
		}
	})
}

func TestAuthService_ChangePassword(t *testing.T) {
	testDB := test.NewTestDB(t)
	defer testDB.Close()
	
	authService := services.NewAuthService(testDB.DB)
	
	t.Run("successful password change", func(t *testing.T) {
		userID := uint(1)
		req := &services.ChangePasswordRequest{
			CurrentPassword: "old_password",
			NewPassword:     "new_strong_password123",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "张律师", "lawyer@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		// 模拟更新密码
		testDB.Mock.ExpectExec("UPDATE `users`").
			WithArgs(
				sqlmock.AnyArg(), // New password hash
				sqlmock.AnyArg(), // UpdatedAt
				userID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
		
		err := authService.ChangePassword(context.Background(), userID, req)
		
		require.NoError(t, err)
	})
	
	t.Run("invalid current password", func(t *testing.T) {
		userID := uint(1)
		req := &services.ChangePasswordRequest{
			CurrentPassword: "wrong_password",
			NewPassword:     "new_strong_password123",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "张律师", "lawyer@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		err := authService.ChangePassword(context.Background(), userID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid current password")
	})
	
	t.Run("weak new password", func(t *testing.T) {
		userID := uint(1)
		req := &services.ChangePasswordRequest{
			CurrentPassword: "old_password",
			NewPassword:     "weak",
		}
		
		// 模拟用户查询
		testDB.Mock.ExpectQuery("SELECT (.+) FROM `users` WHERE id = ?").
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "status"}).
				AddRow(userID, "张律师", "lawyer@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqJ3f3Q6p2KJvKQeTJ8L5X9N5q6Z8i", "lawyer", "active"))
		
		err := authService.ChangePassword(context.Background(), userID, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "new password is too weak")
	})
}