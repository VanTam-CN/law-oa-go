package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/handlers"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 数据库连接
	dsn := "host=localhost user=postgres password=admin dbname=law_oa_go port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// Redis连接
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 初始化仓库
	repos := repositories.NewRepositories(db, rdb, nil)

	// 创建风险评估器
	riskAssessor := services.NewRiskAssessor(nil, nil)

	// 创建冲突检测服务
	conflictService := services.NewConflictDetectionService(
		repos.ConflictRepo,
		riskAssessor,
		repos.UserRepo,
		repos.ClientRepo,
		repos.CaseRepo,
	)

	// 创建处理器
	conflictHandler := handlers.NewConflictHandlerSimple(conflictService)

	// 创建测试请求
	testRequest := &models.ConflictCheckRequest{
		ClientID:                "55",  // 阿里巴巴集团
		ClientName:              "阿里巴巴集团控股有限公司",
		ClientType:              "COMPANY",
		OtherParties:            []string{"字节跳动"},
		CaseName:                "测试案件",
		CaseType:                "civil",    // 使用前端发送的英文类型
		SearchYears:             5,
		IncludeCorporateRelations: true,
		SearchDepth:             "STANDARD",
		UserID:                  "45",     // 张伟律师的ID
		RequestTime:             time.Now(),
	}

	log.Println("🧪 开始测试冲突检测修复...")
	log.Printf("测试请求: %+v", testRequest)

	// 执行冲突检测
	ctx := context.Background()
	response, err := conflictService.PerformConflictCheck(ctx, testRequest)
	if err != nil {
		log.Fatalf("❌ 冲突检测失败: %v", err)
	}

	// 输出结果
	log.Println("✅ 冲突检测成功!")
	log.Printf("检查ID: %s", response.CheckID)
	log.Printf("是否有冲突: %t", response.HasConflict)
	log.Printf("冲突案例数量: %d", len(response.ConflictCases))

	if response.HasConflict {
		log.Println("🔥 发现的冲突案例:")
		for i, conflict := range response.ConflictCases {
			log.Printf("  %d. 案件ID: %s, 名称: %s, 风险: %s",
				i+1, conflict.CaseID, conflict.CaseName, conflict.RiskLevel)
			log.Printf("     描述: %s", conflict.Description)
		}
	}

	log.Printf("风险评估: %s", response.RiskAssessment.OverallRisk)
	log.Printf("建议数量: %d", len(response.Recommendations))
	for i, rec := range response.Recommendations {
		log.Printf("  %d. %s", i+1, rec)
	}

	// 序列化响应为JSON用于调试
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err == nil {
		log.Println("📄 完整响应数据:")
		fmt.Println(string(jsonData))
	}
}