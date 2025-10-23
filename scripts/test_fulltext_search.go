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

	// 创建测试数据
	testClients := []models.Client{
		{Name: "张三律师事务所", Email: "zhang@law.com", Phone: "13800138001", Type: "律师事务所", Company: "张三律师事务所", Status: "active"},
		{Name: "李四法务咨询", Email: "li@law.com", Phone: "13800138002", Type: "企业法务", Company: "李四法务咨询", Status: "active"},
		{Name: "王五知识产权", Email: "wang@law.com", Phone: "13800138003", Type: "知识产权", Company: "王五知识产权服务", Status: "active"},
	}

	// 清理并插入测试数据
	db.Where("name LIKE ?", "测试%").Delete(&models.Client{})
	for _, client := range testClients {
		client.Name = "测试_" + client.Name
		if err := db.Create(&client).Error; err != nil {
			log.Printf("创建测试客户失败: %v", err)
		}
	}

	// 测试搜索功能
	clientRepo := repositories.NewClientRepository(db)
	ctx := context.Background()

	searchTerms := []string{"律师事务所", "法务", "知识产权", "测试", "不存在的关键词"}

	fmt.Println("📋 测试2: PostgreSQL 全文搜索功能")
	fmt.Println("==================================")

	for _, term := range searchTerms {
		fmt.Printf("🔍 搜索关键词: '%s'\n", term)

		params := &repositories.ClientListParams{
			Search:   term,
			Page:     1,
			PageSize: 10,
		}

		results, total, err := clientRepo.List(ctx, params)
		if err != nil {
			fmt.Printf("   ❌ 搜索失败: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 找到 %d 条结果 (总计: %d)\n", len(results), total)
		for i, client := range results {
			if i >= 3 { // 只显示前3条
				fmt.Printf("      ... (还有 %d 条)\n", len(results)-i)
				break
			}
			fmt.Printf("      - %s (%s - %s)\n", client.Name, client.Company, client.Type)
		}
		fmt.Println()
	}

	// 清理测试数据
	db.Where("name LIKE ?", "测试%").Delete(&models.Client{})

	fmt.Println("✅ 全文搜索功能测试完成!")
}