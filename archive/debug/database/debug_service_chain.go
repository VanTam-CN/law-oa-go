package main

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	// 数据库连接配置
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("数据库连接失败:", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 创建repository和service
	clientRepo := repositories.NewClientRepository(db)
	clientService := services.NewClientService(clientRepo)

	// 测试完整的调用链
	fmt.Println("\n🔍 测试Repository层:")
	clients, total, err := clientRepo.List(context.Background(), &repositories.ClientListParams{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		fmt.Println("Repository List失败:", err)
		return
	}

	fmt.Printf("Repository返回: 总数=%d, 客户数=%d\n", total, len(clients))
	if len(clients) > 0 {
		client := clients[0]
		fmt.Printf("第一个客户: ID=%d, Name=%s, Type='%s'\n", client.ID, client.Name, client.Type)
	}

	fmt.Println("\n🔍 测试Service层:")
	serviceResponses, _, err := clientService.ListClients(context.Background(), &services.ClientListRequest{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		fmt.Println("Service ListClients失败:", err)
		return
	}

	fmt.Printf("Service返回: 客户数=%d\n", len(serviceResponses))
	if len(serviceResponses) > 0 {
		response := serviceResponses[0]
		fmt.Printf("第一个客户响应: ID=%d, Name=%s, Type='%s'\n", response.ID, response.Name, response.Type)
	}

	fmt.Println("\n🔍 测试单个客户获取:")
	singleClient, err := clientService.GetClientByID(context.Background(), 13) // 张三
	if err != nil {
		fmt.Println("GetClientByID失败:", err)
		return
	}

	fmt.Printf("单个客户: ID=%d, Name=%s, Type='%s'\n", singleClient.ID, singleClient.Name, singleClient.Type)
}