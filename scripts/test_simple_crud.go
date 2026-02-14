//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.User{}, &models.Lawyer{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("📋 测试3: PostgreSQL CRUD 操作 (简化版)")
	fmt.Println("=====================================")

	// 初始化仓库
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	lawyerRepo := repositories.NewLawyerRepository(db)
	ctx := context.Background()

	// 1. 创建测试律师（用于外键约束）
	testLawyer := &models.Lawyer{
		Name:     "测试律师",
		Email:    "test_lawyer@example.com",
		Phone:    "13900139001",
		Specialty: "民事案件",
		Status:   "active",
	}

	err = lawyerRepo.Create(ctx, testLawyer)
	if err != nil {
		log.Fatalf("创建律师失败: %v", err)
	}
	fmt.Printf("   ✅ 创建律师成功，ID: %d\n", testLawyer.ID)

	// 2. 创建测试客户
	testClient := &models.Client{
		Name:    "测试客户_简化",
		Email:   "test_simple@example.com",
		Phone:   "13900139002",
		Type:    "个人",
		Company: "测试公司",
		Status:  "active",
	}

	err = clientRepo.Create(ctx, testClient)
	if err != nil {
		log.Fatalf("创建客户失败: %v", err)
	}
	fmt.Printf("   ✅ 创建客户成功，ID: %d\n", testClient.ID)

	// 3. 创建测试案件（关联律师和客户）
	testCase := &models.Case{
		Title:       "测试案件_简化",
		Description: "这是一个简化测试案件",
		Status:      "active",
		ClientID:    testClient.ID,
		LawyerID:    testLawyer.ID, // 设置律师ID
		Priority:    "medium",
	}

	err = caseRepo.Create(ctx, testCase)
	if err != nil {
		log.Fatalf("创建案件失败: %v", err)
	}
	fmt.Printf("   ✅ 创建案件成功，ID: %d\n", testCase.ID)

	// 4. 测试关联查询
	caseWithClient, err := caseRepo.FindByID(ctx, testCase.ID)
	if err != nil || caseWithClient == nil {
		log.Fatalf("关联查询失败: %v", err)
	}
	fmt.Printf("   ✅ 案件关联查询成功: %s\n", caseWithClient.Title)

	// 5. 测试列表查询和搜索
	params := &repositories.CaseListParams{
		Page:     1,
		PageSize: 10,
		Search:   "简化",
	}
	_, total, err := caseRepo.List(ctx, params)
	if err != nil {
		log.Fatalf("案件列表查询失败: %v", err)
	}
	fmt.Printf("   ✅ 案件搜索功能正常，找到 %d 条记录\n", total)

	// 清理测试数据
	fmt.Println("\n🧹 清理测试数据...")
	db.Unscoped().Where("name LIKE ?", "%测试%").Delete(&models.Client{})
	db.Unscoped().Where("name LIKE ?", "%测试%").Delete(&models.Lawyer{})
	db.Unscoped().Where("title LIKE ?", "%测试%").Delete(&models.Case{})

	fmt.Println("✅ CRUD 操作测试完成!")
}