package services

import (
	"context"
	"testing"
	"time"

	"github.com/law-oa-go/document-service/internal/models"
	"github.com/law-oa-go/document-service/internal/repositories/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUserService 测试用户服务
func TestUserService(t *testing.T) {
	// 创建模拟仓库
	userRepo := &mocks.UserRepository{}
	roleRepo := &mocks.RoleRepository{}
	permissionRepo := &mocks.DocumentPermissionRepository{}
	auditRepo := &mocks.DocumentAuditRepository{}

	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // 测试时只记录错误

	// 创建用户服务
	userService := NewUserService(userRepo, roleRepo, permissionRepo, auditRepo, logger)

	ctx := context.Background()

	t.Run("CreateUser", func(t *testing.T) {
		// 准备测试数据
		req := &CreateUserRequest{
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			TenantID:  "tenant1",
		}

		// 模拟仓库行为
		userRepo.On("GetByEmail", ctx, req.Email).Return(nil, assert.AnError) // 用户不存在
		userRepo.On("GetByUsername", ctx, req.Username).Return(nil, assert.AnError) // 用户名不存在
		userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

		// 执行测试
		result, err := userService.CreateUser(ctx, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Username, result.Username)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, req.FirstName, result.FirstName)
		assert.Equal(t, req.LastName, result.LastName)

		// 验证模拟调用
		userRepo.AssertExpectations(t)
	})

	t.Run("CreateUser_InvalidEmail", func(t *testing.T) {
		// 准备测试数据
		req := &CreateUserRequest{
			Username:  "testuser",
			Email:     "invalid-email", // 无效邮箱
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
			TenantID:  "tenant1",
		}

		// 执行测试
		result, err := userService.CreateUser(ctx, req)

		// 验证结果
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("GetUser", func(t *testing.T) {
		// 准备测试数据
		userID := "1"
		expectedUser := &models.User{
			ID:        1,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			TenantID:  "tenant1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 模拟仓库行为
		userRepo.On("GetByID", ctx, uint(1)).Return(expectedUser, nil)

		// 执行测试
		result, err := userService.GetUser(ctx, userID)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedUser.Username, result.Username)
		assert.Equal(t, expectedUser.Email, result.Email)

		// 验证模拟调用
		userRepo.AssertExpectations(t)
	})
}

// TestPermissionService 测试权限服务
func TestPermissionService(t *testing.T) {
	// 创建模拟仓库
	permissionRepo := &mocks.DocumentPermissionRepository{}
	docRepo := &mocks.DocumentRepository{}
	userRepo := &mocks.UserRepository{}
	roleRepo := &mocks.RoleRepository{}
	auditRepo := &mocks.DocumentAuditRepository{}

	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建权限服务
	permissionService := NewPermissionService(permissionRepo, docRepo, userRepo, roleRepo, auditRepo, logger)

	ctx := context.Background()

	t.Run("GrantPermission", func(t *testing.T) {
		// 准备测试数据
		req := &GrantPermissionRequest{
			DocumentID: "1",
			UserID:     "1",
			Permission: "read",
			TenantID:   "tenant1",
		}

		document := &models.Document{
			ID:        1,
			Name:      "Test Document",
			TenantID:  "tenant1",
		}

		// 模拟仓库行为
		docRepo.On("GetByID", ctx, uint(1)).Return(document, nil)
		permissionRepo.On("CheckUserPermission", ctx, uint(1), uint(1), "read").Return(false, nil)
		permissionRepo.On("Create", ctx, mock.AnythingOfType("*models.DocumentPermission")).Return(nil)

		// 执行测试
		err := permissionService.GrantPermission(ctx, req)

		// 验证结果
		assert.NoError(t, err)

		// 验证模拟调用
		docRepo.AssertExpectations(t)
		permissionRepo.AssertExpectations(t)
	})

	t.Run("CheckPermission", func(t *testing.T) {
		// 准备测试数据
		userID := "1"
		documentID := "1"
		permission := "read"

		// 模拟仓库行为
		permissionRepo.On("CheckUserPermission", ctx, uint(1), uint(1), "read").Return(true, nil)

		// 执行测试
		hasPermission, err := permissionService.CheckPermission(ctx, userID, documentID, permission)

		// 验证结果
		assert.NoError(t, err)
		assert.True(t, hasPermission)

		// 验证模拟调用
		permissionRepo.AssertExpectations(t)
	})
}

// TestStorageService 测试存储服务
func TestStorageService(t *testing.T) {
	// 创建日志器
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// 创建内存存储服务
	storageService := NewMemoryStorageService(logger)

	ctx := context.Background()

	t.Run("UploadFile", func(t *testing.T) {
		// 准备测试数据
		fileData := []byte("test file content")
		req := &UploadFileRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Data:        fileData,
			Size:        int64(len(fileData)),
			TenantID:    "tenant1",
		}

		// 执行测试
		result, err := storageService.UploadFile(ctx, req)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.FileName, result.FileName)
		assert.Equal(t, req.ContentType, result.ContentType)
		assert.Equal(t, req.Size, result.Size)
		assert.Equal(t, req.TenantID, result.TenantID)
		assert.NotEmpty(t, result.ID)
	})

	t.Run("DownloadFile", func(t *testing.T) {
		// 先上传文件
		fileData := []byte("test file content")
		uploadReq := &UploadFileRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Data:        fileData,
			Size:        int64(len(fileData)),
			TenantID:    "tenant1",
		}

		uploadResult, err := storageService.UploadFile(ctx, uploadReq)
		assert.NoError(t, err)

		// 下载文件
		downloadData, metadata, err := storageService.DownloadFile(ctx, uploadResult.ID)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, fileData, downloadData)
		assert.NotNil(t, metadata)
		assert.Equal(t, uploadResult.FileName, metadata.FileName)
		assert.Equal(t, uploadResult.ContentType, metadata.ContentType)
	})

	t.Run("DeleteFile", func(t *testing.T) {
		// 先上传文件
		fileData := []byte("test file content")
		uploadReq := &UploadFileRequest{
			FileName:    "test.txt",
			ContentType: "text/plain",
			Data:        fileData,
			Size:        int64(len(fileData)),
			TenantID:    "tenant1",
		}

		uploadResult, err := storageService.UploadFile(ctx, uploadReq)
		assert.NoError(t, err)

		// 删除文件
		err = storageService.DeleteFile(ctx, uploadResult.ID)

		// 验证结果
		assert.NoError(t, err)

		// 验证文件已被删除
		_, _, err = storageService.DownloadFile(ctx, uploadResult.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})
}

// TestServiceManager 测试服务管理器
func TestServiceManager(t *testing.T) {
	// 注意：这个测试需要实际的仓库管理器，这里只测试基本结构
	// 在实际实现中，应该使用模拟的仓库管理器

	t.Run("Health", func(t *testing.T) {
		// 这里只测试健康检查的基本逻辑
		// 实际测试需要完整的服务管理器设置

		// 创建一个简单的服务管理器实例进行测试
		sm := &serviceManager{
			logger: logrus.New(),
		}

		ctx := context.Background()

		// 执行健康检查
		health, err := sm.Health(ctx)

		// 验证结果
		assert.NoError(t, err)
		assert.NotNil(t, health)

		// 由于服务未初始化，所有服务状态应该是错误
		for serviceName, serviceErr := range health {
			assert.Error(t, serviceErr, "Service %s should have error", serviceName)
			assert.Contains(t, serviceErr.Error(), "not initialized")
		}
	})
}

// BenchmarkStorageService 存储服务性能测试
func BenchmarkStorageService_UploadDownload(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	storageService := NewMemoryStorageService(logger)
	ctx := context.Background()

	// 准备测试数据
	fileData := make([]byte, 1024) // 1KB文件
	for i := range fileData {
		fileData[i] = byte(i % 256)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 上传文件
		uploadReq := &UploadFileRequest{
			FileName:    "benchmark.txt",
			ContentType: "application/octet-stream",
			Data:        fileData,
			Size:        int64(len(fileData)),
			TenantID:    "tenant1",
		}

		result, err := storageService.UploadFile(ctx, uploadReq)
		if err != nil {
			b.Fatal(err)
		}

		// 下载文件
		_, _, err = storageService.DownloadFile(ctx, result.ID)
		if err != nil {
			b.Fatal(err)
		}

		// 清理文件
		err = storageService.DeleteFile(ctx, result.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}