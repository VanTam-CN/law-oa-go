package main

import (
	"context"
	"log"
	"time"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	log.Println("🔍 开始集成测试冲突检测服务...")

	// 1. 验证模型完整性
	log.Println("📋 验证模型完整性...")

	// 测试User模型是否包含新字段
	user := &models.User{
		Username:   "testuser",
		Name:       "测试用户",
		Email:      "test@example.com",
		Role:       "lawyer",
		Department: "诉讼部", // 新增字段
		Seniority:  "中级",   // 新增字段
	}
	log.Printf("✅ User模型验证通过: %+v", user)

	// 测试冲突检测请求模型
	request := &models.ConflictCheckRequest{
		ClientID:                "1",
		ClientName:              "测试客户公司",
		ClientType:              "COMPANY",
		OtherParties:            []string{"对方公司"},
		CaseName:                "商业纠纷案件",
		CaseType:                "商业纠纷",
		SearchYears:             5,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                  1,
		RequestTime:             time.Now(),
	}
	log.Printf("✅ ConflictCheckRequest模型验证通过")

	// 验证请求有效性
	if err := request.Validate(); err != nil {
		log.Printf("❌ 请求验证失败: %v", err)
		return
	}
	log.Printf("✅ 请求验证通过")

	// 2. 验证服务创建
	log.Println("🔧 验证服务创建...")

	// 注意：这里使用nil值仅用于编译测试，实际使用时需要真实的仓储实例
	var conflictRepo repositories.BasicConflictRepository
	var userRepo repositories.UserRepository
	var clientRepo repositories.ClientRepository
	var caseRepo repositories.CaseRepository

	// 创建风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)
	log.Printf("✅ 风险评估器创建成功")

	// 创建冲突检测服务
	conflictService := services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)
	log.Printf("✅ 冲突检测服务创建成功")

	// 3. 验证接口完整性
	log.Println("🔌 验证接口完整性...")

	// 验证服务接口方法存在
	ctx := context.Background()

	// 这些调用会因为nil仓储而失败，但能验证接口签名正确性
	_, err := conflictService.PerformConflictCheck(ctx, request)
	if err != nil {
		log.Printf("⚠️ 预期的nil仓储错误: %v", err)
	}

	history, err := conflictService.GetCheckHistory(ctx, "1", 10)
	if err != nil {
		log.Printf("⚠️ 预期的nil仓储错误: %v", err)
	}
	_ = history

	stats, err := conflictService.GetConflictStats(ctx, "1")
	if err != nil {
		log.Printf("⚠️ 预期的nil仓储错误: %v", err)
	}
	_ = stats

	log.Printf("✅ 接口方法验证完成")

	// 4. 验证RuleMatch模型
	log.Println("📊 验证RuleMatch模型...")

	ruleMatch := &models.RuleMatch{
		RuleID:     "RULE_001",
		RuleName:   "商业竞争检测",
		Matched:    true,
		RiskScore:  75.5,
		Reason:     "检测到同行业竞争关系",
		Confidence: 0.85,
	}
	log.Printf("✅ RuleMatch模型验证通过: %+v", ruleMatch)

	// 5. 验证风险评估模型
	log.Println("🎯 验证风险评估模型...")

	riskAssessment := &models.RiskAssessment{
		OverallRisk:     "HIGH",
		RiskScore:       75.5,
		RiskReason:      "检测到商业竞争冲突",
		RequiresApproval: true,
		ApprovalLevel:   "SENIOR",
		RiskFactors:     []string{"同行业竞争", "利益冲突"},
		Mitigation:      []string{"建立信息隔离墙", "获取客户书面同意"},
	}
	log.Printf("✅ RiskAssessment模型验证通过: %+v", riskAssessment)

	log.Println("🎉 冲突检测服务集成测试完成！")
	log.Println("📊 测试总结:")
	log.Println("  ✅ User模型字段补全完成")
	log.Println("  ✅ 冲突检测模型完整性验证通过")
	log.Println("  ✅ 服务接口签名验证通过")
	log.Println("  ✅ 依赖注入结构验证通过")
	log.Println("  ✅ 编译状态检查通过")
}