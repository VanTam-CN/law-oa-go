package main

import (
	"context"
	"log"

	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	log.Println("测试冲突检测服务编译状态...")

	// 模拟创建仓储（仅用于编译测试）
	var conflictRepo repositories.BasicConflictRepository
	var userRepo repositories.UserRepository
	var clientRepo repositories.ClientRepository
	var caseRepo repositories.CaseRepository

	// 创建风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)

	// 创建冲突检测服务
	conflictService := services.NewConflictDetectionService(
		conflictRepo,
		riskAssessor,
		userRepo,
		clientRepo,
		caseRepo,
	)

	// 模拟请求
	request := &models.ConflictCheckRequest{
		ClientID: "1",
		ClientName: "测试客户",
		CaseName: "测试案件",
		CaseType: "商业纠纷",
		UserID: 1,
	}

	// 测试方法调用（仅验证编译，不实际执行）
	_ = conflictService
	_ = request

	log.Println("✅ 冲突检测服务编译测试通过")
}