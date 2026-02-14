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
	if err := db.AutoMigrate(&models.Client{}, &models.Case{}, &models.User{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("📋 测试3: PostgreSQL CRUD 操作 (最终版)")
	fmt.Println("=====================================")

	// 初始化仓库
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	userRepo := repositories.NewUserRepository(db)
	ctx := context.Background()

	// 1. 创建测试用户（律师角色，用于外键约束）
	testUser := &models.User{
		Username: "test_lawyer",
		Name:     "测试律师",
		Email:    "test_lawyer@example.com",
		Password: "hashed_password",
		Role:     "lawyer",
		Phone:    "13900139001",
		Status:   "active",
	}

	err = userRepo.Create(ctx, testUser)
	if err != nil {
		log.Fatalf("创建用户(律师)失败: %v", err)
	}
	fmt.Printf("   ✅ 创建用户(律师)成功，ID: %d\n", testUser.ID)

	// 2. 创建测试客户
	testClient := &models.Client{
		Name:    "测试客户_最终版",
		Email:   "test_final@example.com",
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
		Title:       "测试案件_最终版",
		Description: "这是一个最终版测试案件",
		Status:      "active",
		ClientID:    testClient.ID,
		LawyerID:    testUser.ID, // 设置律师ID
		Priority:    "medium",
		CaseType:    "民事案件",
	}

	err = caseRepo.Create(ctx, testCase)
	if err != nil {
		log.Fatalf("创建案件失败: %v", err)
	}
	fmt.Printf("   ✅ 创建案件成功，ID: %d\n", testCase.ID)

	// 4. 测试关联查询
	caseWithDetails, err := caseRepo.FindByID(ctx, testCase.ID)
	if err != nil || caseWithDetails == nil {
		log.Fatalf("关联查询失败: %v", err)
	}
	fmt.Printf("   ✅ 案件关联查询成功: %s (律师: %s)\n", caseWithDetails.Title, caseWithDetails.Lawyer.Name)

	// 5. 测试更新操作
	caseWithDetails.Status = "completed"
	caseWithDetails.Description = "已更新的案件描述"
	err = caseRepo.Update(ctx, caseWithDetails)
	if err != nil {
		log.Fatalf("更新案件失败: %v", err)
	}
	fmt.Printf("   ✅ 案件更新成功\n")

	// 6. 测试列表查询和搜索
	params := &repositories.CaseListParams{
		Page:     1,
		PageSize: 10,
		Search:   "最终版",
	}
	_, total, err := caseRepo.List(ctx, params)
	if err != nil {
		log.Fatalf("案件列表查询失败: %v", err)
	}
	fmt.Printf("   ✅ 案件搜索功能正常，找到 %d 条记录\n", total)

	// 7. 测试分页查询
	clientParams := &repositories.ClientListParams{
		Page:     1,
		PageSize: 5,
	}
	_, clientTotal, err := clientRepo.List(ctx, clientParams)
	if err != nil {
		log.Fatalf("客户列表查询失败: %v", err)
	}
	fmt.Printf("   ✅ 客户分页查询正常，总计 %d 条记录\n", clientTotal)

	// 清理测试数据
	fmt.Println("\n🧹 清理测试数据...")
	db.Unscoped().Where("username LIKE ?", "%test_%").Delete(&models.User{})
	db.Unscoped().Where("name LIKE ?", "%测试%").Delete(&models.Client{})
	db.Unscoped().Where("title LIKE ?", "%测试%").Delete(&models.Case{})

	fmt.Println("✅ CRUD 操作测试完成!")
}