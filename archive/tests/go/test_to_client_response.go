package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
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

	// 测试toClientResponse方法
	var clientModel models.Client
	result := db.First(&clientModel, 13) // 张三
	if result.Error != nil {
		fmt.Println("查询客户失败:", result.Error)
		return
	}

	fmt.Printf("📋 原始模型数据:\n")
	fmt.Printf("- ID: %d\n", clientModel.ID)
	fmt.Printf("- Name: %s\n", clientModel.Name)
	fmt.Printf("- Type: %s\n", clientModel.Type)
	fmt.Printf("- Email: %s\n", clientModel.Email)

	// 手动构建ClientResponse（模拟toClientResponse方法）
	clientResponse := &services.ClientResponse{
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
	fmt.Printf("\n📋 转换后的响应数据:\n")
	fmt.Printf("- ID: %d\n", clientResponse.ID)
	fmt.Printf("- Name: %s\n", clientResponse.Name)
	fmt.Printf("- Type: %s\n", clientResponse.Type)
	fmt.Printf("- Email: %s\n", clientResponse.Email)

	// 测试JSON序列化
	fmt.Printf("\n🔍 JSON序列化测试:\n")
	jsonBytes, err := json.MarshalIndent(clientResponse, "", "  ")
	if err != nil {
		fmt.Println("JSON序列化失败:", err)
		return
	}

	fmt.Println("序列化结果:")
	fmt.Println(string(jsonBytes))

	// 测试通过ListClients方法获取的数据
	fmt.Printf("\n🔍 测试ListClients方法:\n")
	clients, _, err := clientService.ListClients(nil, &services.ClientListRequest{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		fmt.Println("ListClients失败:", err)
		return
	}

	if len(clients) > 0 {
		firstClient := clients[0]
		fmt.Printf("第一个客户数据:\n")
		fmt.Printf("- ID: %d\n", firstClient.ID)
		fmt.Printf("- Name: %s\n", firstClient.Name)
		fmt.Printf("- Type: %s\n", firstClient.Type)
		fmt.Printf("- Email: %s\n", firstClient.Email)

		// 测试JSON序列化
		jsonBytes2, _ := json.MarshalIndent(firstClient, "", "  ")
		fmt.Println("序列化结果:")
		fmt.Println(string(jsonBytes2))
	}
}