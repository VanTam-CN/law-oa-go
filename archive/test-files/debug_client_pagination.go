package main

import (
	"context"
	"fmt"
	"log"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	fmt.Println("=== 调试客户API分页问题 ===")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 连接数据库
	db, err := database.InitWithConfig(cfg)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 创建repository
	clientRepo := repositories.NewClientRepository(db)

	// 创建service
	clientService := services.NewClientService(clientRepo)

	// 测试不同的分页参数
	testCases := []struct {
		page     int
		pageSize int
		desc     string
	}{
		{1, 10, "第1页，每页10条"},
		{1, 20, "第1页，每页20条（默认）"},
		{1, 100, "第1页，每页100条（最大限制）"},
		{1, 9999, "第1页，每页9999条（前端请求）"},
	}

	for _, tc := range testCases {
		fmt.Printf("\n--- %s ---\n", tc.desc)

		req := &services.ClientListRequest{
			Page:     tc.page,
			PageSize: tc.pageSize,
		}

		clients, total, err := clientService.ListClients(context.Background(), req)
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			continue
		}

		fmt.Printf("✅ 返回 %d 条记录，总数: %d\n", len(clients), total)
		fmt.Printf("📋 客户列表:\n")
		for _, client := range clients {
			fmt.Printf("   - ID:%d, %s (%s)\n", client.ID, client.Name, client.Type)
		}
	}

	// 直接查询数据库总数
	fmt.Printf("\n--- 直接查询数据库 ---\n")
	var totalClients int64
	sqlDB, _ := db.DB()
	row := sqlDB.QueryRow("SELECT COUNT(*) FROM clients WHERE deleted_at IS NULL")
	row.Scan(&totalClients)
	fmt.Printf("✅ 数据库中实际客户总数: %d\n", totalClients)
}