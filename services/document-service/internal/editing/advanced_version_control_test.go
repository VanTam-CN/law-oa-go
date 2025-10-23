package editing

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAdvancedVersionControlService 测试创建高级版本控制服务
func TestAdvancedVersionControlService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-advanced-repos", logger, authService)

	assert.NotNil(t, service)
	assert.Implements(t, (*AdvancedVersionControlService)(nil), service)
}

// TestBranchInfo 测试获取分支信息
func TestBranchInfo(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-branch-info", logger, authService)

	ctx := context.Background()
	docID := "test-doc-branch-info"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 保存一些版本
	_, err = service.SaveVersion(ctx, docID, []byte("Initial content"), "test-user", "Initial commit")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Updated content"), "test-user", "Update commit")
	require.NoError(t, err)

	// 获取分支信息
	info, err := service.BranchInfo(ctx, docID, "main")
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "main", info.Name)
	assert.Equal(t, 2, info.CommitsCount)
	assert.False(t, info.Protected)
}

// TestCompareBranches 测试比较分支
func TestCompareBranches(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-compare-branches", logger, authService)

	ctx := context.Background()
	docID := "test-doc-compare"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 在main分支创建版本
	_, err = service.SaveVersion(ctx, docID, []byte("Main content v1"), "user1", "Main v1")
	require.NoError(t, err)

	// 创建feature分支
	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	// 在feature分支创建版本
	err = service.SwitchBranch(ctx, docID, "feature")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature content v1"), "user2", "Feature v1")
	require.NoError(t, err)

	// 回到main分支继续开发
	err = service.SwitchBranch(ctx, docID, "main")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Main content v2"), "user1", "Main v2")
	require.NoError(t, err)

	// 比较分支
	diff, err := service.CompareBranches(ctx, docID, "feature", "main")
	assert.NoError(t, err)
	assert.NotNil(t, diff)
	assert.Equal(t, "feature", diff.SourceBranch)
	assert.Equal(t, "main", diff.TargetBranch)
	assert.NotEmpty(t, diff.SourceCommits)
	assert.NotEmpty(t, diff.TargetCommits)
}

// TestProtectBranch 测试分支保护
func TestProtectBranch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-protect-branch", logger, authService)

	ctx := context.Background()
	docID := "test-doc-protect"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 保护main分支
	config := &BranchProtectionConfig{
		RequireReviews:      true,
		RequiredReviewers:   []string{"admin"},
		RequireStatusChecks:  false,
		AllowForcePushes:    false,
		RequireLinearHistory: true,
		RestrictPushes:      true,
		AllowedPushers:      []string{"admin"},
		EnforceAdmins:       true,
	}

	err = service.ProtectBranch(ctx, docID, "main", config)
	assert.NoError(t, err)

	// 验证分支受保护
	isProtected := service.IsBranchProtected(ctx, docID, "main")
	assert.True(t, isProtected)

	// 验证分支未受保护
	isProtected = service.IsBranchProtected(ctx, docID, "feature")
	assert.False(t, isProtected)

	// 取消保护
	err = service.UnprotectBranch(ctx, docID, "main")
	assert.NoError(t, err)

	// 验证分支不再受保护
	isProtected = service.IsBranchProtected(ctx, docID, "main")
	assert.False(t, isProtected)
}

// TestMergeBranch 测试分支合并
func TestMergeBranch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-merge-branch", logger, authService)

	ctx := context.Background()
	docID := "test-doc-merge"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 在main分支创建基础版本
	_, err = service.SaveVersion(ctx, docID, []byte("Base content"), "main-user", "Base commit")
	require.NoError(t, err)

	// 创建feature分支
	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	// 在feature分支创建版本
	err = service.SwitchBranch(ctx, docID, "feature")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature content"), "feature-user", "Feature commit")
	require.NoError(t, err)

	// 回到main分支
	err = service.SwitchBranch(ctx, docID, "main")
	require.NoError(t, err)

	// 合并分支（快进策略）
	result, err := service.MergeBranch(ctx, docID, "feature", "main", MergeStrategyFastForward)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, MergeStrategyFastForward, MergeStrategy(result.MergeType))

	// 测试三方合并
	err = service.CreateBranch(ctx, docID, "feature2")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "feature2")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature2 content"), "feature2-user", "Feature2 commit")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "main")
	require.NoError(t, err)

	result, err = service.MergeBranch(ctx, docID, "feature2", "main", MergeStrategyThreeWay)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, MergeStrategyThreeWay, MergeStrategy(result.MergeType))
}

// TestPreviewMerge 测试合并预览
func TestPreviewMerge(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-preview-merge", logger, authService)

	ctx := context.Background()
	docID := "test-doc-preview"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支并添加不同内容
	err = service.CreateBranch(ctx, docID, "source")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "source")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Source branch content"), "source-user", "Source commit")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "main")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Target branch content"), "target-user", "Target commit")
	require.NoError(t, err)

	// 预览合并
	preview, err := service.PreviewMerge(ctx, docID, "source", "main")
	assert.NoError(t, err)
	assert.NotNil(t, preview)
	assert.False(t, preview.CanMerge) // 简化实现总是认为有潜在冲突
	assert.NotEmpty(t, preview.Preview)
	assert.NotEmpty(t, preview.Strategy)
}

// TestCreateTag 测试创建标签
func TestCreateTag(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-create-tag", logger, authService)

	ctx := context.Background()
	docID := "test-doc-tag"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建版本
	versionID, err := service.SaveVersion(ctx, docID, []byte("Release content"), "release-user", "Release v1.0")
	require.NoError(t, err)

	// 创建标签
	err = service.CreateTag(ctx, docID, "v1.0", versionID, "First stable release")
	assert.NoError(t, err)

	// 获取标签列表
	tags, err := service.GetTags(ctx, docID)
	assert.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, "v1.0", tags[0].Name)
	assert.Equal(t, versionID, tags[0].VersionID)
	assert.Equal(t, "First stable release", tags[0].Message)
}

// TestCherryPick 测试挑选提交
func TestCherryPick(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-cherry-pick", logger, authService)

	ctx := context.Background()
	docID := "test-doc-cherry-pick"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 在main分支创建版本
	_, err = service.SaveVersion(ctx, docID, []byte("Main v1"), "main-user", "Main v1")
	require.NoError(t, err)

	// 创建feature分支
	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	// 在feature分支创建版本
	err = service.SwitchBranch(ctx, docID, "feature")
	require.NoError(t, err)

	featureVersionID, err := service.SaveVersion(ctx, docID, []byte("Feature feature"), "feature-user", "Feature v1")
	require.NoError(t, err)

	// 回到main分支
	err = service.SwitchBranch(ctx, docID, "main")
	require.NoError(t, err)

	// 挑选提交
	err = service.CherryPick(ctx, docID, featureVersionID, "main")
	assert.NoError(t, err)

	// 验证提交已被挑选
	versions, err := service.GetVersions(ctx, docID)
	assert.NoError(t, err)
	assert.Greater(t, len(versions), 2)
}

// TestRebaseBranch 测试变基分支
func TestRebaseBranch(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-rebase", logger, authService)

	ctx := context.Background()
	docID := "test-doc-rebase"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 在main分支创建基础版本
	_, err = service.SaveVersion(ctx, docID, []byte("Base content"), "main-user", "Base commit")
	require.NoError(t, err)

	// 在main分支继续开发
	_, err = service.SaveVersion(ctx, docID, []byte("Base content v2"), "main-user", "Base v2")
	require.NoError(t, err)

	// 创建feature分支
	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	// 在feature分支创建版本
	err = service.SwitchBranch(ctx, docID, "feature")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature content"), "feature-user", "Feature commit")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature content v2"), "feature-user", "Feature v2")
	require.NoError(t, err)

	// 变基到main分支
	err = service.RebaseBranch(ctx, docID, "feature", "main")
	assert.NoError(t, err)

	// 验证变基结果
	versions, err := service.GetVersions(ctx, docID)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(versions), 3)
}

// TestDetectConflicts 测试冲突检测
func TestDetectConflicts(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-detect-conflicts", logger, authService)

	ctx := context.Background()
	docID := "test-doc-conflicts"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支
	err = service.CreateBranch(ctx, docID, "branch1")
	require.NoError(t, err)

	err = service.CreateBranch(ctx, docID, "branch2")
	require.NoError(t, err)

	// 在不同分支创建冲突内容
	err = service.SwitchBranch(ctx, docID, "branch1")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Branch 1 content"), "user1", "Branch 1 commit")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "branch2")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Branch 2 different content"), "user2", "Branch 2 commit")
	require.NoError(t, err)

	// 检测冲突
	conflicts, err := service.DetectConflicts(ctx, docID, "branch1", "branch2")
	assert.NoError(t, err)
	assert.NotEmpty(t, conflicts)

	// 验证冲突信息
	for _, conflict := range conflicts {
		assert.NotEmpty(t, conflict.ID)
		assert.NotEmpty(t, conflict.FilePath)
		assert.Equal(t, ConflictStatusPending, conflict.Status)
	}
}

// TestResolveConflict 测试冲突解决
func TestResolveConflict(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-resolve-conflict", logger, authService)

	ctx := context.Background()
	docID := "test-doc-resolve-conflict"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建冲突场景
	err = service.CreateBranch(ctx, docID, "branch1")
	require.NoError(t, err)

	err = service.CreateBranch(ctx, docID, "branch2")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "branch1")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Branch 1 content"), "user1", "Branch 1 commit")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "branch2")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Branch 2 different content"), "user2", "Branch 2 commit")
	require.NoError(t, err)

	// 检测冲突
	conflicts, err := service.DetectConflicts(ctx, docID, "branch1", "branch2")
	require.NoError(t, err)
	require.NotEmpty(t, conflicts)

	// 解决冲突
	conflict := conflicts[0]
	err = service.ResolveConflict(ctx, docID, conflict.ID, "Resolved content")
	assert.NoError(t, err)

	// 验证冲突已解决
	resolvedConflicts, err := service.GetConflicts(ctx, docID)
	assert.NoError(t, err)
	assert.Empty(t, resolvedConflicts)
}

// TestVersionGraph 测试版本图
func TestVersionGraph(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-version-graph", logger, authService)

	ctx := context.Background()
	docID := "test-doc-graph"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建多个分支和版本
	_, err = service.SaveVersion(ctx, docID, []byte("Initial content"), "user1", "Initial")
	require.NoError(t, err)

	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "feature")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Feature content"), "user2", "Feature commit")
	require.NoError(t, err)

	err = service.CreateBranch(ctx, docID, "hotfix")
	require.NoError(t, err)

	err = service.SwitchBranch(ctx, docID, "hotfix")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("Hotfix content"), "user3", "Hotfix commit")
	require.NoError(t, err)

	// 获取版本图
	graph, err := service.GetVersionGraph(ctx, docID)
	assert.NoError(t, err)
	assert.NotNil(t, graph)
	assert.Equal(t, docID, graph.DocID)
	assert.Equal(t, 3, graph.Commits)
	assert.Equal(t, 3, graph.BranchesCnt)
	assert.NotEmpty(t, graph.Nodes)
	assert.NotEmpty(t, graph.Edges)
	assert.NotEmpty(t, graph.Branches)
}

// TestContributors 测试贡献者统计
func TestContributors(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-contributors", logger, authService)

	ctx := context.Background()
	docID := "test-doc-contributors"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 多个用户贡献
	_, err = service.SaveVersion(ctx, docID, []byte("User 1 contribution"), "user1", "User 1 commit")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("User 2 contribution"), "user2", "User 2 commit")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("User 1 another contribution"), "user1", "User 1 second commit")
	require.NoError(t, err)

	_, err = service.SaveVersion(ctx, docID, []byte("User 3 contribution"), "user3", "User 3 commit")
	require.NoError(t, err)

	// 获取贡献者列表
	contributors, err := service.GetContributors(ctx, docID)
	assert.NoError(t, err)
	assert.Len(t, contributors, 3)

	// 验证贡献统计
	user1Found := false
	user2Found := false
	user3Found := false

	for _, contributor := range contributors {
		switch contributor.UserName {
		case "user1":
			user1Found = true
			assert.Equal(t, 2, contributor.Commits)
		case "user2":
			user2Found = true
			assert.Equal(t, 1, contributor.Commits)
		case "user3":
			user3Found = true
			assert.Equal(t, 1, contributor.Commits)
		}
	}

	assert.True(t, user1Found)
	assert.True(t, user2Found)
	assert.True(t, user3Found)
}

// TestActivityLog 测试活动日志
func TestActivityLog(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/test-activity-log", logger, authService)

	ctx := context.Background()
	docID := "test-doc-activity-log"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	require.NoError(t, err)

	// 创建分支
	err = service.CreateBranch(ctx, docID, "feature")
	require.NoError(t, err)

	// 保存版本（会生成活动）
	_, err = service.SaveVersion(ctx, docID, []byte("Activity test"), "test-user", "Test commit")
	require.NoError(t, err)

	// 合并分支
	err = service.MergeBranch(ctx, docID, "feature", "main", MergeStrategyFastForward)
	require.NoError(t, err)

	// 获取活动日志
	options := &ActivityLogOptions{
		Limit: 10,
	}

	activities, err := service.GetActivityLog(ctx, docID, options)
	assert.NoError(t, err)
	assert.NotEmpty(t, activities)

	// 验证活动类型
	actions := make(map[string]bool)
	for _, activity := range activities {
		actions[activity.Action] = true
	}

	assert.True(t, actions["create_branch"])
	assert.True(t, actions["save_version"])
	assert.True(t, actions["merge"])
}

// BenchmarkMergeBranch 性能测试：合并分支
func BenchmarkMergeBranch(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/bench-merge", logger, authService)
	ctx := context.Background()
	docID := "bench-doc"

	// 初始化仓库
	service.InitializeRepository(ctx, docID)

	// 创建源分支
	service.CreateBranch(ctx, docID, "source")
	service.SwitchBranch(ctx, docID, "source")
	service.SaveVersion(ctx, docID, []byte("Source content"), "bench-user", "Source commit")

	// 切换到目标分支
	service.SwitchBranch(ctx, docID, "main")
	service.SaveVersion(ctx, docID, []byte("Target content"), "bench-user", "Target commit")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.MergeBranch(ctx, docID, "source", "main", MergeStrategyFastForward)
	}
}

// BenchmarkDetectConflicts 性能测试：检测冲突
func BenchmarkDetectConflicts(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	authService := &MockAuthService{}

	service := NewAdvancedVersionControlService("/tmp/bench-conflicts", logger, authService)
	ctx := context.Background()
	docID := "bench-doc-conflicts"

	// 初始化仓库和创建冲突场景
	service.InitializeRepository(ctx, docID)
	service.CreateBranch(ctx, docID, "branch1")
	service.CreateBranch(ctx, docID, "branch2")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.DetectConflicts(ctx, docID, "branch1", "branch2")
	}
}