package editing

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAuthService 模拟认证服务
type MockAuthService struct{}

func (m *MockAuthService) AuthenticateUser(ctx context.Context, token string) (string, error) {
	return "user123", nil
}

func (m *MockAuthService) HasPermission(ctx context.Context, userID, resource, action string) error {
	return nil
}

// TestNewSimpleVersionControlService 测试创建简化版本控制服务
func TestNewSimpleVersionControlService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	assert.NotNil(t, service)
	assert.Implements(t, (*SimpleVersionControlService)(nil), service)
}

// TestInitializeRepository 测试初始化仓库
func TestInitializeRepository(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-123"

	// 测试初始化
	err := service.InitializeRepository(ctx, docID)
	assert.NoError(t, err)

	// 验证仓库存在
	repoInfo, err := service.GetRepository(ctx, docID)
	assert.NoError(t, err)
	assert.Equal(t, docID, repoInfo)

	// 重复初始化应该不报错
	err = service.InitializeRepository(ctx, docID)
	assert.NoError(t, err)
}

// TestSaveVersion 测试保存版本
func TestSaveVersion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-save"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 保存版本
	content := []byte("Hello, World!")
	author := "test-user"
	message := "Initial version"

	versionID, err := service.SaveVersion(ctx, docID, content, author, message)
	assert.NoError(t, err)
	assert.NotEmpty(t, versionID)

	// 保存第二个版本
	content2 := []byte("Hello, World! Updated!")
	message2 := "Updated version"

	versionID2, err := service.SaveVersion(ctx, docID, content2, author, message2)
	assert.NoError(t, err)
	assert.NotEmpty(t, versionID2)
	assert.NotEqual(t, versionID, versionID2)
}

// TestGetVersion 测试获取版本
func TestGetVersion(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-get"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 保存版本
	content := []byte("Test content for get")
	author := "test-user"
	message := "Test version"

	versionID, err := service.SaveVersion(ctx, docID, content, author, message)
	require.NoError(t, err)

	// 获取版本
	retrievedContent, err := service.GetVersion(ctx, docID, versionID)
	assert.NoError(t, err)
	assert.Equal(t, content, retrievedContent)

	// 获取不存在的版本
	_, err = service.GetVersion(ctx, docID, "non-existent-version")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "版本不存在")
}

// TestGetVersions 测试获取版本列表
func TestGetVersions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-list"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取空版本列表
	versions, err := service.GetVersions(ctx, docID)
	assert.NoError(t, err)
	assert.Empty(t, versions)

	// 保存多个版本
	contents := []string{"Version 1", "Version 2", "Version 3"}
	messages := []string{"First version", "Second version", "Third version"}

	for i, content := range contents {
		_, err := service.SaveVersion(ctx, docID, []byte(content), "test-user", messages[i])
		require.NoError(t, err)
	}

	// 获取版本列表
	versions, err = service.GetVersions(ctx, docID)
	assert.NoError(t, err)
	assert.Len(t, versions, 3)

	// 验证版本信息
	for i, version := range versions {
		assert.Equal(t, docID, version.DocID)
		assert.Equal(t, "test-user", version.Author)
		assert.Equal(t, messages[i], version.Message)
		assert.Equal(t, "main", version.Branch)
		assert.Greater(t, version.Size, int64(0))
	}
}

// TestCompareVersions 测试版本比较
func TestCompareVersions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-compare"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 保存第一个版本
	content1 := []byte("Line 1\nLine 2\nLine 3")
	versionID1, err := service.SaveVersion(ctx, docID, content1, "test-user", "First version")
	require.NoError(t, err)

	// 保存第二个版本
	content2 := []byte("Line 1\nLine 2 Modified\nLine 3\nLine 4")
	versionID2, err := service.SaveVersion(ctx, docID, content2, "test-user", "Second version")
	require.NoError(t, err)

	// 比较版本
	diffResult, err := service.CompareVersions(ctx, docID, versionID1, versionID2)
	assert.NoError(t, err)
	assert.NotNil(t, diffResult)
	assert.Equal(t, versionID1, diffResult.FromVersion)
	assert.Equal(t, versionID2, diffResult.ToVersion)
	assert.NotEmpty(t, diffResult.Changes)
	assert.NotNil(t, diffResult.Summary)
}

// TestCreateBranch 测试创建分支
func TestCreateBranch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-branch"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支
	err = service.CreateBranch(ctx, docID, "feature-branch")
	assert.NoError(t, err)

	// 重复创建应该报错
	err = service.CreateBranch(ctx, docID, "feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分支已存在")
}

// TestGetBranches 测试获取分支列表
func TestGetBranches(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-get-branches"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取初始分支列表
	branches, err := service.GetBranches(ctx, docID)
	assert.NoError(t, err)
	assert.Contains(t, branches, "main")

	// 创建新分支
	err = service.CreateBranch(ctx, docID, "feature-branch")
	require.NoError(t, err)

	// 获取更新后的分支列表
	branches, err = service.GetBranches(ctx, docID)
	assert.NoError(t, err)
	assert.Contains(t, branches, "main")
	assert.Contains(t, branches, "feature-branch")
	assert.Len(t, branches, 2)
}

// TestSwitchBranch 测试切换分支
func TestSwitchBranch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-switch"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支
	err = service.CreateBranch(ctx, docID, "feature-branch")
	require.NoError(t, err)

	// 切换分支
	err = service.SwitchBranch(ctx, docID, "feature-branch")
	assert.NoError(t, err)

	// 切换不存在的分支
	err = service.SwitchBranch(ctx, docID, "non-existent-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分支不存在")
}

// TestDeleteRepository 测试删除仓库
func TestDeleteRepository(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/test-repos", logger, authService)

	ctx := context.Background()
	docID := "test-doc-delete"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 验证仓库存在
	_, err = service.GetRepository(ctx, docID)
	assert.NoError(t, err)

	// 删除仓库
	err = service.DeleteRepository(ctx, docID)
	assert.NoError(t, err)

	// 验证仓库不存在
	_, err = service.GetRepository(ctx, docID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仓库不存在")
}

// BenchmarkSaveVersion 性能测试：保存版本
func BenchmarkSaveVersion(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/bench-repos", logger, authService)
	ctx := context.Background()
	docID := "bench-doc"

	// 初始化仓库
	service.InitializeRepository(ctx, docID)

	content := []byte("Benchmark content for performance testing")
	author := "bench-user"
	message := "Benchmark version"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SaveVersion(ctx, docID, content, author, message)
	}
}

// BenchmarkGetVersion 性能测试：获取版本
func BenchmarkGetVersion(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewSimpleVersionControlService("/tmp/bench-repos", logger, authService)
	ctx := context.Background()
	docID := "bench-doc-get"

	// 初始化仓库并保存一些版本
	service.InitializeRepository(ctx, docID)

	var versionIDs []string
	for i := 0; i < 100; i++ {
		versionID, _ := service.SaveVersion(ctx, docID, []byte("Benchmark content"), "bench-user", "Benchmark version")
		versionIDs = append(versionIDs, versionID)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		versionIndex := i % len(versionIDs)
		service.GetVersion(ctx, docID, versionIDs[versionIndex])
	}
}