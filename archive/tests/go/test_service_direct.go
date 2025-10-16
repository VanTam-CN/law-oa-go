package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	fmt.Println("🔍 直接测试Service层")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ GORM数据库连接失败: %v", err)
	}

	// 创建Repository和Service
	clientRepo := repositories.NewClientRepository(db)
	clientService := services.NewClientService(clientRepo)

	// 测试1: 通过ID获取客户
	fmt.Println("📋 测试1: 通过Service获取ID=13")
	client, err := clientService.GetClientByID(context.Background(), 13)
	if err != nil {
		log.Printf("❌ Service查询失败: %v", err)
	} else if client == nil {
		fmt.Println("❌ 客户不存在")
	} else {
		fmt.Printf("✅ Service查询结果:\n")
		fmt.Printf("   ID: %d\n", client.ID)
		fmt.Printf("   Name: '%s'\n", client.Name)
		fmt.Printf("   Type: '%s'\n", client.Type)
		fmt.Printf("   Email: '%s'\n", client.Email)
		fmt.Printf("   Phone: '%s'\n", client.Phone)
		fmt.Printf("   Status: '%s'\n", client.Status)
	}

	// 测试2: 列表查询
	fmt.Println("\n📋 测试2: 通过Service搜索客户")
	listReq := &services.ClientListRequest{
		Page:     1,
		PageSize: 10,
		Search:   "张三",
	}

	clients, total, err := clientService.ListClients(context.Background(), listReq)
	if err != nil {
		log.Printf("❌ Service列表查询失败: %v", err)
	} else {
		fmt.Printf("✅ Service列表查询结果: 找到%d条记录，总共%d条\n", len(clients), total)
		for i, c := range clients {
			fmt.Printf("   %d. ID:%d Name:'%s' Type:'%s' Email:'%s'\n",
				i+1, c.ID, c.Name, c.Type, c.Email)
		}
	}

	// 测试3: 直接调用Repository和Service转换
	fmt.Println("\n📋 测试3: 测试模型转换")
	var clientModel models.Client
	result := db.First(&clientModel, 13)
	if result.Error != nil {
		log.Printf("❌ 直接查询模型失败: %v", result.Error)
	} else {
		fmt.Printf("✅ 原始模型数据:\n")
		fmt.Printf("   ID: %d\n", clientModel.ID)
		fmt.Printf("   Name: '%s'\n", clientModel.Name)
		fmt.Printf("   Type: '%s'\n", clientModel.Type)
		fmt.Printf("   Email: '%s'\n", clientModel.Email)

		// 手动调用toClientResponse转换
		response := &services.ClientResponse{
			ID:            clientModel.ID,
			Name:          clientModel.Name,
			Type:          clientModel.Type,
			Email:         clientModel.Email,
			Phone:         clientModel.Phone,
			Address:       clientModel.Address,
			IDCard:        clientModel.IDCard,
			Company:       clientModel.Company,
			Industry:      clientModel.Industry,
			ContactPerson: clientModel.ContactPerson,
			ContactPhone:  clientModel.ContactPhone,
			Source:        clientModel.Source,
			Notes:         clientModel.Notes,
			Status:        clientModel.Status,
			CreatedAt:     clientModel.CreatedAt,
			UpdatedAt:     clientModel.UpdatedAt,
		}

		fmt.Printf("✅ 转换后的响应:\n")
		fmt.Printf("   ID: %d\n", response.ID)
		fmt.Printf("   Name: '%s'\n", response.Name)
		fmt.Printf("   Type: '%s'\n", response.Type)
		fmt.Printf("   Email: '%s'\n", response.Email)
	}

	fmt.Println("\n🎯 Service层测试结论:")
	if client != nil && client.Name != "" && client.Type != "" {
		fmt.Println("   ✅ Service层正常，数据正确")
		fmt.Println("   🔍 问题可能在API路由或中间件层")
	} else {
		fmt.Println("   ❌ Service层有问题")
	}
}