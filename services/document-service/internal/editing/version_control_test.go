package editing

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService 模拟认证服务
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestNewVersionControlService 测试创建版本控制服务
func TestNewVersionControlService(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	service := NewVersionControlService("/tmp/test-version-control", logger, authService)

	assert.NotNil(t, service)
	assert.Implements(t, (*VersionControlService)(nil), service)
}

// TestInitializeRepository 测试初始化仓库
func TestInitializeRepository(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	// 模拟认证服务
	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-init-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	assert.NoError(t, err)

	// 验证仓库存在
	repoDir := filepath.Join("/tmp/test-init-repo", service.(*VersionControlImpl).hashDocumentID(docID))
	assert.DirExists(t, repoDir)
	assert.FileExists(t, filepath.Join(repoDir, ".git"))
	assert.FileExists(t, filepath.Join(repoDir, "README.md"))

	// 清理
	os.RemoveAll("/tmp/test-init-repo")
}

// TestGetRepository 测试获取仓库
func TestGetRepository(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-get-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-get"

	// 先初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取仓库
	repo, err := service.GetRepository(ctx, docID)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	// 清理
	os.RemoveAll("/tmp/test-get-repo")
}

// TestDeleteRepository 测试删除仓库
func TestDeleteRepository(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	// 模拟有删除权限
	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "admin",
	}, nil)

	service := NewVersionControlService("/tmp/test-delete-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-delete"

	// 先初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 验证仓库存在
	repoDir := filepath.Join("/tmp/test-delete-repo", service.(*VersionControlImpl).hashDocumentID(docID))
	assert.DirExists(t, repoDir)

	// 删除仓库
	err = service.DeleteRepository(ctx, docID)
	assert.NoError(t, err)

	// 验证仓库已删除
	assert.NoDirExists(t, repoDir)

	// 清理（如果还有残留）
	os.RemoveAll("/tmp/test-delete-repo")
}

// TestCreateCommit 测试创建提交
func TestCreateCommit(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-commit-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-commit"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建提交请求
	req := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Test commit",
		Files: []*FileChange{
			{
				Path:      "test.txt",
				Operation: "create",
				Content:   []byte("Hello, World!"),
			},
		},
		Metadata: map[string]string{
			"version": "1.0",
		},
	}

	// 创建提交
	commitInfo, err := service.CreateCommit(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, commitInfo)
	assert.Equal(t, "Test User", commitInfo.AuthorName)
	assert.Equal(t, "test@example.com", commitInfo.AuthorEmail)
	assert.Equal(t, "Test commit", commitInfo.Message)
	assert.Equal(t, "main", commitInfo.Branch)
	assert.Equal(t, "1.0", commitInfo.Metadata["version"])

	// 清理
	os.RemoveAll("/tmp/test-commit-repo")
}

// TestGetCommit 测试获取提交
func TestGetCommit(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-get-commit-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-get-commit"

	// 初始化仓库并创建提交
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	req := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Test commit for get",
		Files: []*FileChange{
			{
				Path:      "get_test.txt",
				Operation: "create",
				Content:   []byte("Get test content"),
			},
		},
	}

	commitInfo, err := service.CreateCommit(ctx, req)
	require.NoError(t, err)

	// 获取提交
	retrievedCommit, err := service.GetCommit(ctx, docID, commitInfo.Hash)
	assert.NoError(t, err)
	assert.NotNil(t, retrievedCommit)
	assert.Equal(t, commitInfo.Hash, retrievedCommit.Hash)
	assert.Equal(t, "Test User", retrievedCommit.AuthorName)
	assert.Equal(t, "Test commit for get", retrievedCommit.Message)

	// 清理
	os.RemoveAll("/tmp/test-get-commit-repo")
}

// TestGetCommitHistory 测试获取提交历史
func TestGetCommitHistory(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-history-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-history"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建多个提交
	commits := make([]*CommitInfo, 0)
	for i := 0; i < 3; i++ {
		req := &CommitRequest{
			DocumentID:  docID,
			BranchName:  "main",
			AuthorName:  "Test User",
			AuthorEmail: "test@example.com",
			Message:     fmt.Sprintf("Commit %d", i+1),
			Files: []*FileChange{
				{
					Path:      fmt.Sprintf("file%d.txt", i+1),
					Operation: "create",
					Content:   []byte(fmt.Sprintf("Content %d", i+1)),
				},
			},
		}

		commitInfo, err := service.CreateCommit(ctx, req)
		require.NoError(t, err)
		commits = append(commits, commitInfo)

		// 等待一小段时间确保时间戳不同
		time.Sleep(10 * time.Millisecond)
	}

	// 获取提交历史
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{
		Limit: 10,
	})
	assert.NoError(t, err)
	assert.Len(t, history, 4) // 3个提交 + 1个初始提交

	// 验证历史顺序（最新的在前）
	assert.Equal(t, commits[2].Message, history[0].Message)
	assert.Equal(t, commits[1].Message, history[1].Message)
	assert.Equal(t, commits[0].Message, history[2].Message)

	// 清理
	os.RemoveAll("/tmp/test-history-repo")
}

// TestCreateBranch 测试创建分支
func TestCreateBranch(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-branch-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-branch"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取初始提交的哈希
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{Limit: 1})
	require.NoError(t, err)
	require.Len(t, history, 1)
	initialCommitHash := history[0].Hash

	// 创建分支
	err = service.CreateBranch(ctx, docID, "feature-branch", initialCommitHash)
	assert.NoError(t, err)

	// 列出分支验证
	branches, err := service.ListBranches(ctx, docID)
	assert.NoError(t, err)
	assert.Len(t, branches, 1) // 只有一个分支（main）

	// 清理
	os.RemoveAll("/tmp/test-branch-repo")
}

// TestDeleteBranch 测试删除分支
func TestDeleteBranch(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-delete-branch-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-delete-branch"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{Limit: 1})
	require.NoError(t, err)
	require.Len(t, history, 1)
	err = service.CreateBranch(ctx, docID, "test-branch", history[0].Hash)
	require.NoError(t, err)

	// 验证分支存在
	branches, err := service.ListBranches(ctx, docID)
	require.NoError(t, err)
	assert.Len(t, branches, 1)

	// 删除分支
	err = service.DeleteBranch(ctx, docID, "test-branch")
	assert.NoError(t, err)

	// 验证分支已删除
	branches, err = service.ListBranches(ctx, docID)
	require.NoError(t, err)
	assert.Len(t, branches, 0) // 只有main分支

	// 清理
	os.RemoveAll("/tmp/test-delete-branch-repo")
}

// TestListBranches 测试列出分支
func TestListBranches(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-list-branches-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-list-branches"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建测试分支
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{Limit: 1})
	require.NoError(t, err)
	require.Len(t, history, 1)
	err = service.CreateBranch(ctx, docID, "feature", history[0].Hash)
	require.NoError(t, err)

	// 列出分支
	branches, err := service.ListBranches(ctx, docID)
	assert.NoError(t, err)
	assert.Len(t, branches, 1)

	// 验证分支信息
	branch := branches[0]
	assert.Equal(t, "feature", branch.Name)
	assert.Equal(t, history[0].Hash, branch.CommitHash)
	assert.False(t, branch.IsDefault) // 不是默认分支

	// 清理
	os.RemoveAll("/tmp/test-list-branches-repo")
}

// TestGetFileContent 测试获取文件内容
func TestGetFileContent(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-file-content-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-file-content"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建包含文件的提交
	fileContent := []byte("This is test content for file reading")
	req := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Add test file",
		Files: []*FileChange{
			{
				Path:      "test_file.txt",
				Operation: "create",
				Content:   fileContent,
			},
		},
	}

	commitInfo, err := service.CreateCommit(ctx, req)
	require.NoError(t, err)

	// 获取文件内容
	content, err := service.GetFileContent(ctx, docID, commitInfo.Hash, "test_file.txt")
	assert.NoError(t, err)
	assert.Equal(t, fileContent, content)

	// 清理
	os.RemoveAll("/tmp/test-file-content-repo")
}

// TestCheckout 测试检出
func TestCheckout(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-checkout-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-checkout"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建测试提交
	req := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Initial commit",
		Files: []*FileChange{
			{
				Path:      "main_file.txt",
				Operation: "create",
				Content:   []byte("Main branch content"),
			},
		},
	}

	commitInfo, err := service.CreateCommit(ctx, req)
	require.NoError(t, err)

	// 创建新分支
	err = service.CreateBranch(ctx, docID, "feature", commitInfo.Hash)
	require.NoError(t, err)

	// 创建feature分支的提交
	featureReq := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "feature",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Feature branch commit",
		Files: []*FileChange{
			{
				Path:      "feature_file.txt",
				Operation: "create",
				Content:   []byte("Feature branch content"),
			},
		},
	}

	featureCommitInfo, err := service.CreateCommit(ctx, featureReq)
	require.NoError(t, err)

	// 检出到feature分支
	err = service.Checkout(ctx, docID, &CheckoutOptions{
		BranchName: "feature",
	})
	assert.NoError(t, err)

	// 验证文件存在
	repoDir := service.(*VersionControlImpl).getRepoPath(docID)
	featureFile := filepath.Join(repoDir, "feature_file.txt")
	assert.FileExists(t, featureFile)

	// 清理
	os.RemoveAll("/tmp/test-checkout-repo")
}

// TestRevertToCommit 测试回滚到提交
func TestRevertToCommit(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "admin",
	}, nil)

	service := NewVersionControlService("/tmp/test-revert-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-revert"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取初始提交
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{Limit: 1})
	require.NoError(t, err)
	require.Len(t, history, 1)
	initialCommitHash := history[0].Hash

	// 创建第二个提交
	secondReq := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Second commit",
		Files: []*FileChange{
			{
				Path:      "second.txt",
				Operation: "create",
				Content:   []byte("Second commit content"),
			},
		},
	}

	secondCommitInfo, err := service.CreateCommit(ctx, secondReq)
	require.NoError(t, err)

	// 验证第二个提交的文件存在
	repoDir := service.(*VersionControlImpl).getRepoPath(docID)
	secondFile := filepath.Join(repoDir, "second.txt")
	assert.FileExists(t, secondFile)

	// 回滚到初始提交
	err = service.RevertToCommit(ctx, docID, initialCommitHash)
	assert.NoError(t, err)

	// 验证第二个提交的文件已被删除
	assert.NoFileExists(t, secondFile)

	// 清理
	os.RemoveAll("/tmp/test-revert-repo")
}

// TestCherryPick 测试应用提交
func TestCherryPick(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-cherry-pick-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-cherry-pick"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 获取初始提交
	history, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{Limit: 1})
	require.NoError(t, err)
	require.Len(t, history, 1)
	initialCommitHash := history[0].Hash

	// 创建feature分支和提交
	err = service.CreateBranch(ctx, docID, "feature", initialCommitHash)
	require.NoError(t, err)

	featureReq := &CommitRequest{
		DocumentID:  docID,
	BranchName:  "feature",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Feature commit",
		Files: []*FileChange{
			{
				Path:      "feature.txt",
				Operation: "create",
				Content:   []byte("Feature content"),
			},
		},
	}

	featureCommitInfo, err := service.CreateCommit(ctx, featureReq)
	require.NoError(t, err)

	// 切换回main分支
	err = service.Checkout(ctx, docID, &CheckoutOptions{
		BranchName: "main",
	})
	require.NoError(t, err)

	// 应用feature提交到main分支
	err = service.CherryPick(ctx, docID, featureCommitInfo.Hash, "main")
	assert.NoError(t, err)

	// 验证feature分支的文件现在也在main分支中
	repoDir := service.(*VersionControlImpl).getRepoPath(docID)
	featureFile := filepath.Join(repoDir, "feature.txt")
	assert.FileExists(t, featureFile)

	// 清理
	os.RemoveAll("/tmp/test-cherry-pick-repo")
}

// TestCompareCommits 测试比较提交
func TestCompareCommits(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-compare-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-compare"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建第一个提交
	firstReq := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "First commit",
		Files: []*FileChange{
			{
				Path:      "compare.txt",
				Operation: "create",
				Content:   []byte("First content"),
			},
		},
	}

	firstCommitInfo, err := service.CreateCommit(ctx, firstReq)
	require.NoError(t, err)

	// 创建第二个提交
	secondReq := &CommitRequest{
		DocumentID:  docID,
		BranchName:  "main",
		AuthorName:  "Test User",
		AuthorEmail: "test@example.com",
		Message:     "Second commit",
		Files: []*FileChange{
			{
				Path:      "compare.txt",
				Operation: "update",
				Content:   []byte("Second content"),
			},
		},
	}

	secondCommitInfo, err := service.CreateCommit(ctx, secondReq)
	require.NoError(t, err)

	// 比较两个提交
	diffResult, err := service.CompareCommits(ctx, docID, firstCommitInfo.Hash, secondCommitInfo.Hash)
	assert.NoError(t, err)
	assert.NotNil(t, diffResult)
	assert.Equal(t, firstCommitInfo.Hash, diffResult.FromCommit)
	assert.Equal(t, secondCommitInfo.Hash, diffResult.ToCommit)
	assert.NotNil(t, diffResult.Summary)
	assert.Equal(t, 1, diffResult.Summary.FilesChanged)
	assert.Equal(t, 1, diffResult.Summary.LinesAdded)
	assert.Equal(t, 1, diffResult.Summary.LinesRemoved)

	// 清理
	os.RemoveAll("/tmp/test-compare-repo")
}

// TestGetFileHistory 测试获取文件历史
func TestGetFileHistory(t *testing.T) {
	logger := logrus.New()
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/test-file-history-repo", logger, authService)

	ctx := context.Background()
	docID := "test-doc-file-history"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建多个提交修改同一个文件
	filePath := "history_test.txt"
	for i := 0; i < 3; i++ {
		req := &CommitRequest{
			DocumentID:  docID,
			BranchName:  "main",
			AuthorName:  "Test User",
			AuthorEmail: "test@example.com",
			Message:     fmt.Sprintf("Update file %d", i+1),
			Files: []*FileChange{
				{
					Path:      filePath,
					Operation: "update",
					Content:   []byte(fmt.Sprintf("Content version %d", i+1)),
				},
			},
		}

		_, err := service.CreateCommit(ctx, req)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	}

	// 获取文件历史
	fileHistory, err := service.GetFileHistory(ctx, docID, filePath)
	assert.NoError(t, err)
	assert.Len(t, fileHistory, 3)

	// 验证历史顺序（最新的在前）
	assert.Contains(t, fileHistory[0].Message, "Update file 3")
	assert.Contains(t, fileHistory[1].Message, "Update file 2")
	assert.Contains(t, fileHistory[2].Message, "Update file 1")

	// 清理
	os.RemoveAll("/tmp/test-file-history-repo")
}

// BenchmarkCreateCommit 性能测试：创建提交
func BenchmarkCreateCommit(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/bench-create-commit", logger, authService)

	ctx := context.Background()
	docID := "bench-doc"

	// 预先初始化仓库
	err := service.InitializeRepository(ctx, docID)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := &CommitRequest{
			DocumentID:  docID,
			BranchName:  "main",
			AuthorName:  "Bench User",
			AuthorEmail: "bench@example.com",
			Message:     fmt.Sprintf("Bench commit %d", i),
			Files: []*FileChange{
				{
					Path:      fmt.Sprintf("bench_file_%d.txt", i),
					Operation: "create",
					Content:   []byte(fmt.Sprintf("Bench content %d", i)),
				},
			},
		}

		_, err := service.CreateCommit(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}

	// 清理
	os.RemoveAll("/tmp/bench-create-commit")
}

// BenchmarkGetCommitHistory 性能测试：获取提交历史
func BenchmarkGetCommitHistory(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	authService.On("ValidateToken", mock.Anything, mock.Anything).Return(map[string]interface{}{
		"user_id": "test-user",
		"role":    "editor",
	}, nil)

	service := NewVersionControlService("/tmp/bench-history", logger, authService)

	ctx := context.Background()
	docID := "bench-history-doc"

	// 预先初始化仓库并创建一些提交
	err := service.InitializeRepository(ctx, docID)
	if err != nil {
		b.Fatal(err)
	}

	// 创建100个提交
	for i := 0; i < 100; i++ {
		req := &CommitRequest{
			DocumentID:  docID,
			BranchName:  "main",
			AuthorName:  "Bench User",
			AuthorEmail: "bench@example.com",
			Message:     fmt.Sprintf("Bench commit %d", i),
			Files: []*FileChange{
				{
					Path:      fmt.Sprintf("bench_file_%d.txt", i),
					Operation: "create",
					Content:   []byte(fmt.Sprintf("Bench content %d", i)),
				},
			},
		}

		_, err := service.CreateCommit(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := service.GetCommitHistory(ctx, docID, &HistoryOptions{
			Limit: 10,
		})
		if err != nil {
			b.Fatal(err)
		}
	}

	// 清理
	os.RemoveAll("/tmp/bench-history")
}