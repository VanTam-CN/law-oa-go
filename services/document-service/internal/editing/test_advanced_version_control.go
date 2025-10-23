package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sirupsen/logrus"
)

// MockAuthService 模拟认证服务
type MockAuthService struct{}

func (m *MockAuthService) AuthenticateUser(ctx context.Context, token string) (string, error) {
	return "user123", nil
}

func (m *MockAuthService) HasPermission(ctx context.Context, userID, resource, action string) error {
	return nil
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	authService := &MockAuthService{}

	fmt.Println("🔧 开始测试高级版本控制系统...")

	// 创建高级版本控制服务
	service := NewAdvancedVersionControlService("/tmp/test-advanced-repos", logger, authService)
	if service == nil {
		log.Fatal("❌ 无法创建高级版本控制服务")
	}

	fmt.Println("✅ 高级版本控制服务创建成功")

	// 测试基本功能
	ctx := context.Background()
	docID := "test-doc-advanced"

	// 初始化仓库
	err := service.InitializeRepository(ctx, docID)
	if err != nil {
		log.Fatalf("❌ 初始化仓库失败: %v", err)
	}
	fmt.Println("✅ 仓库初始化成功")

	// 保存版本
	content := []byte("# 测试文档\n\n这是一个测试文档，用于验证高级版本控制功能。")
	versionID, err := service.SaveVersion(ctx, docID, content, "test-user", "初始版本")
	if err != nil {
		log.Fatalf("❌ 保存版本失败: %v", err)
	}
	fmt.Printf("✅ 版本保存成功，ID: %s\n", versionID)

	// 创建分支
	err = service.CreateBranch(ctx, docID, "feature")
	if err != nil {
		log.Fatalf("❌ 创建分支失败: %v", err)
	}
	fmt.Println("✅ 分支创建成功")

	// 切换分支
	err = service.SwitchBranch(ctx, docID, "feature")
	if err != nil {
		log.Fatalf("❌ 切换分支失败: %v", err)
	}
	fmt.Println("✅ 分支切换成功")

	// 在feature分支保存版本
	featureContent := []byte("# 测试文档\n\n这是一个测试文档。\n\n## 新功能\n\n添加了新功能内容。")
	featureVersionID, err := service.SaveVersion(ctx, docID, featureContent, "feature-user", "功能开发")
	if err != nil {
		log.Fatalf("❌ 在feature分支保存版本失败: %v", err)
	}
	fmt.Printf("✅ Feature分支版本保存成功，ID: %s\n", featureVersionID)

	// 切换回main分支
	err = service.SwitchBranch(ctx, docID, "main")
	if err != nil {
		log.Fatalf("❌ 切换回main分支失败: %v", err)
	}
	fmt.Println("✅ 已切换回main分支")

	// 在main分支保存版本
	mainContent := []byte("# 测试文档\n\n这是一个更新后的测试文档。\n\n## 主线开发\n\n主线分支的新内容。")
	mainVersionID, err := service.SaveVersion(ctx, docID, mainContent, "main-user", "主线更新")
	if err != nil {
		log.Fatalf("❌ 在main分支保存版本失败: %v", err)
	}
	fmt.Printf("✅ Main分支版本保存成功，ID: %s\n", mainVersionID)

	// 获取分支信息
	mainInfo, err := service.BranchInfo(ctx, docID, "main")
	if err != nil {
		log.Fatalf("❌ 获取main分支信息失败: %v", err)
	}
	fmt.Printf("✅ Main分支信息: %d个提交\n", mainInfo.CommitsCount)

	featureInfo, err := service.BranchInfo(ctx, docID, "feature")
	if err != nil {
		log.Fatalf("❌ 获取feature分支信息失败: %v", err)
	}
	fmt.Printf("✅ Feature分支信息: %d个提交\n", featureInfo.CommitsCount)

	// 比较分支
	diff, err := service.CompareBranches(ctx, docID, "feature", "main")
	if err != nil {
		log.Fatalf("❌ 比较分支失败: %v", err)
	}
	fmt.Printf("✅ 分支比较: Feature有%d个提交，Main有%d个提交\n",
		len(diff.SourceCommits), len(diff.TargetCommits))

	// 预览合并
	preview, err := service.PreviewMerge(ctx, docID, "feature", "main")
	if err != nil {
		log.Fatalf("❌ 预览合并失败: %v", err)
	}
	fmt.Printf("✅ 合并预览: %s\n", preview.Preview)

	// 执行合并
	result, err := service.MergeBranch(ctx, docID, "feature", "main", MergeStrategyThreeWay)
	if err != nil {
		log.Fatalf("❌ 合并分支失败: %v", err)
	}
	if result.Success {
		fmt.Printf("✅ 分支合并成功，策略: %s\n", result.MergeType)
	} else {
		fmt.Printf("❌ 分支合并失败: %s\n", result.Message)
	}

	// 分支保护测试
	protectionConfig := &BranchProtectionConfig{
		RequireReviews:      true,
		RequiredReviewers:   []string{"admin"},
		RequireStatusChecks:  false,
		AllowForcePushes:    false,
		RequireLinearHistory: true,
		RestrictPushes:      true,
		AllowedPushers:      []string{"admin"},
		EnforceAdmins:       true,
	}

	err = service.ProtectBranch(ctx, docID, "main", protectionConfig)
	if err != nil {
		log.Fatalf("❌ 保护分支失败: %v", err)
	}
	fmt.Println("✅ 分支保护设置成功")

	isProtected := service.IsBranchProtected(ctx, docID, "main")
	fmt.Printf("✅ Main分支保护状态: %v\n", isProtected)

	// 创建标签
	err = service.CreateTag(ctx, docID, "v1.0", versionID, "第一个稳定版本")
	if err != nil {
		log.Fatalf("❌ 创建标签失败: %v", err)
	}
	fmt.Println("✅ 标签创建成功")

	// 获取标签列表
	tags, err := service.GetTags(ctx, docID)
	if err != nil {
		log.Fatalf("❌ 获取标签列表失败: %v", err)
	}
	fmt.Printf("✅ 标签列表: %d个标签\n", len(tags))
	for _, tag := range tags {
		fmt.Printf("   - %s: %s\n", tag.Name, tag.Message)
	}

	// 获取版本列表
	versions, err := service.GetVersions(ctx, docID)
	if err != nil {
		log.Fatalf("❌ 获取版本列表失败: %v", err)
	}
	fmt.Printf("✅ 版本列表: %d个版本\n", len(versions))

	// 获取贡献者列表
	contributors, err := service.GetContributors(ctx, docID)
	if err != nil {
		log.Fatalf("❌ 获取贡献者列表失败: %v", err)
	}
	fmt.Printf("✅ 贡献者列表: %d个贡献者\n", len(contributors))
	for _, contributor := range contributors {
		fmt.Printf("   - %s: %d个提交\n", contributor.UserName, contributor.Commits)
	}

	// 获取活动日志
	activityOptions := &ActivityLogOptions{
		Limit: 10,
	}
	activities, err := service.GetActivityLog(ctx, docID, activityOptions)
	if err != nil {
		log.Fatalf("❌ 获取活动日志失败: %v", err)
	}
	fmt.Printf("✅ 活动日志: %d个活动\n", len(activities))

	// 获取版本图
	graph, err := service.GetVersionGraph(ctx, docID)
	if err != nil {
		log.Fatalf("❌ 获取版本图失败: %v", err)
	}
	fmt.Printf("✅ 版本图: %d个提交，%d个分支\n", graph.Commits, graph.BranchesCnt)

	fmt.Println("\n🎉 高级版本控制系统所有功能测试通过！")
}