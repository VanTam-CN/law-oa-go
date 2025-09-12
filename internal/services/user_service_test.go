package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

// MockDB 创建内存数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.Client{}, &models.Case{})
	assert.NoError(t, err)

	return db
}

func TestUserService_CreateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		req     *CreateUserRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid user creation",
			req: &CreateUserRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
				Role:     "lawyer",
			},
			wantErr: false,
		},
		{
			name: "invalid email format",
			req: &CreateUserRequest{
				Name:     "Test User",
				Email:    "invalid-email",
				Password: "password123",
				Role:     "lawyer",
			},
			wantErr: true,
			errMsg:  "invalid email format",
		},
		{
			name: "weak password",
			req: &CreateUserRequest{
				Name:     "Test User",
				Email:    "test2@example.com",
				Password: "123",
				Role:     "lawyer",
			},
			wantErr: true,
			errMsg:  "password too weak",
		},
		{
			name: "invalid role",
			req: &CreateUserRequest{
				Name:     "Test User",
				Email:    "test3@example.com",
				Password: "password123",
				Role:     "invalid",
			},
			wantErr: true,
			errMsg:  "invalid role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.CreateUser(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Name, user.Name)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, tt.req.Role, user.Role)
				assert.NotEmpty(t, user.ID)
			}
		})
	}
}

func TestUserService_AuthenticateUser(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)
	ctx := context.Background()

	// 创建测试用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "lawyer",
	}
	db.Create(testUser)

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid credentials",
			email:    "test@example.com",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "user not found",
			email:    "notfound@example.com",
			password: "password123",
			wantErr:  true,
			errMsg:   "user not found",
		},
		{
			name:     "invalid password",
			email:    "test@example.com",
			password: "wrongpassword",
			wantErr:  true,
			errMsg:   "invalid password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.AuthenticateUser(ctx, tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}
		})
	}
}

func TestUserService_GetUserProfile(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)
	ctx := context.Background()

	// 创建测试用户
	testUser := &models.User{
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "lawyer",
	}
	db.Create(testUser)

	tests := []struct {
		name    string
		userID  uint
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid user ID",
			userID:  testUser.ID,
			wantErr: false,
		},
		{
			name:    "user not found",
			userID:  999,
			wantErr: true,
			errMsg:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUserProfile(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}
		})
	}
}

func TestUserService_ChangePassword(t *testing.T) {
	db := setupTestDB(t)
	service := NewUserService(db)
	ctx := context.Background()

	// 创建测试用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	testUser := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "lawyer",
	}
	db.Create(testUser)

	tests := []struct {
		name            string
		userID          uint
		currentPassword string
		newPassword     string
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "valid password change",
			userID:          testUser.ID,
			currentPassword: "oldpassword",
			newPassword:     "newpassword123",
			wantErr:         false,
		},
		{
			name:            "user not found",
			userID:          999,
			currentPassword: "oldpassword",
			newPassword:     "newpassword123",
			wantErr:         true,
			errMsg:          "user not found",
		},
		{
			name:            "invalid current password",
			userID:          testUser.ID,
			currentPassword: "wrongpassword",
			newPassword:     "newpassword123",
			wantErr:         true,
			errMsg:          "invalid current password",
		},
		{
			name:            "weak new password",
			userID:          testUser.ID,
			currentPassword: "oldpassword",
			newPassword:     "123",
			wantErr:         true,
			errMsg:          "password too weak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ChangePassword(ctx, tt.userID, tt.currentPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				
				// 验证密码已更改
				var user models.User
				db.First(&user, tt.userID)
				err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.newPassword))
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkUserService_CreateUser(b *testing.B) {
	db := setupTestDB(&testing.T{})
	service := NewUserService(db)
	ctx := context.Background()

	req := &CreateUserRequest{
		Name:     "Benchmark User",
		Email:    "benchmark@example.com",
		Password: "password123",
		Role:     "lawyer",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Email = fmt.Sprintf("benchmark%d@example.com", i)
		_, _ = service.CreateUser(ctx, req)
	}
}

func BenchmarkUserService_AuthenticateUser(b *testing.B) {
	db := setupTestDB(&testing.T{})
	service := NewUserService(db)
	ctx := context.Background()

	// 创建测试用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := &models.User{
		Name:     "Benchmark User",
		Email:    "benchmark@example.com",
		Password: string(hashedPassword),
		Role:     "lawyer",
	}
	db.Create(testUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.AuthenticateUser(ctx, "benchmark@example.com", "password123")
	}
}