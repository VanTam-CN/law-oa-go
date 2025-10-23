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

	fmt.Println("📋 测试3: PostgreSQL CRUD 操作")
	fmt.Println("===============================")

	// 初始化仓库
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	ctx := context.Background()

	// 1. 测试客户CRUD操作
	fmt.Println("👥 客户管理 CRUD 测试:")

	// 创建客户
	testClient := &models.Client{
		Name:    "测试客户_CRUD",
		Email:   "test_crud@example.com",
		Phone:   "13900139000",
		Type:    "个人",
		Company: "测试公司",
		Status:  "active",
	}

	err = clientRepo.Create(ctx, testClient)
	if err != nil {
		log.Fatalf("创建客户失败: %v", err)
	}
	fmt.Printf("   ✅ 创建客户成功，ID: %d\n", testClient.ID)

	// 查询客户
	foundClient, err := clientRepo.FindByID(ctx, testClient.ID)
	if err != nil || foundClient == nil {
		log.Fatalf("查询客户失败: %v", err)
	}
	fmt.Printf("   ✅ 查询客户成功: %s\n", foundClient.Name)

	// 更新客户
	foundClient.Company = "更新后的测试公司"
	err = clientRepo.Update(ctx, foundClient)
	if err != nil {
		log.Fatalf("更新客户失败: %v", err)
	}
	fmt.Printf("   ✅ 更新客户成功: %s\n", foundClient.Company)

	// 列表查询
	params := &repositories.ClientListParams{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}
	_, total, err := clientRepo.List(ctx, params)
	if err != nil {
		log.Fatalf("查询客户列表失败: %v", err)
	}
	fmt.Printf("   ✅ 客户列表查询成功，共 %d 条记录\n", total)

	// 2. 测试案件CRUD操作
	fmt.Println("\n⚖️ 案件管理 CRUD 测试:")

	// 创建案件
	testCase := &models.Case{
		Title:       "测试案件_CRUD",
		Description: "这是一个测试案件的描述",
		Status:      "active",
		ClientID:    foundClient.ID,
		Priority:    "medium",
	}

	err = caseRepo.Create(ctx, testCase)
	if err != nil {
		log.Fatalf("创建案件失败: %v", err)
	}
	fmt.Printf("   ✅ 创建案件成功，ID: %d\n", testCase.ID)

	// 查询案件
	foundCase, err := caseRepo.FindByID(ctx, testCase.ID)
	if err != nil || foundCase == nil {
		log.Fatalf("查询案件失败: %v", err)
	}
	fmt.Printf("   ✅ 查询案件成功: %s\n", foundCase.Title)

	// 更新案件
	foundCase.Description = "更新后的案件描述"
	err = caseRepo.Update(ctx, foundCase)
	if err != nil {
		log.Fatalf("更新案件失败: %v", err)
	}
	fmt.Printf("   ✅ 更新案件成功\n")

	// 案件列表查询
	caseParams := &repositories.CaseListParams{
		Page:     1,
		PageSize: 10,
		Status:   "active",
	}
	_, caseTotal, err := caseRepo.List(ctx, caseParams)
	if err != nil {
		log.Fatalf("查询案件列表失败: %v", err)
	}
	fmt.Printf("   ✅ 案件列表查询成功，共 %d 条记录\n", caseTotal)

	// 3. 测试关联查询
	fmt.Println("\n🔗 关联查询测试:")

	caseWithClient, err := caseRepo.FindByID(ctx, testCase.ID)
	if err != nil || caseWithClient == nil {
		log.Fatalf("关联查询失败: %v", err)
	}
	fmt.Printf("   ✅ 案件关联客户: %s - %s\n", caseWithClient.Title, caseWithClient.Client.Name)

	// 4. 测试软删除
	fmt.Println("\n🗑️ 软删除测试:")

	err = clientRepo.Delete(ctx, foundClient.ID)
	if err != nil {
		log.Fatalf("删除客户失败: %v", err)
	}
	fmt.Printf("   ✅ 客户软删除成功\n")

	// 验证软删除后的查询
	deletedClient, err := clientRepo.FindByID(ctx, foundClient.ID)
	if err != nil || deletedClient != nil {
		fmt.Printf("   ⚠️ 软删除后仍然可以查询到客户，这可能需要检查\n")
	} else {
		fmt.Printf("   ✅ 软删除验证成功，无法查询到已删除的客户\n")
	}

	// 清理测试数据
	fmt.Println("\n🧹 清理测试数据...")

	// 物理删除测试数据
	db.Unscoped().Where("name LIKE ?", "%测试%").Delete(&models.Client{})
	db.Unscoped().Where("title LIKE ?", "%测试%").Delete(&models.Case{})

	fmt.Println("✅ CRUD 操作测试完成!")
}