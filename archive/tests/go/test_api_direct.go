package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"
)

func main() {
	fmt.Println("🔍 直接测试API完整流程")
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

	// 开始测试

	// 测试1: 模拟单个客户API响应
	fmt.Println("📋 测试1: 模拟GetClient API响应")
	client, err := clientService.GetClientByID(context.Background(), 13)
	if err != nil {
		log.Printf("❌ Service查询失败: %v", err)
	} else {
		fmt.Printf("✅ Service层结果: Name='%s', Type='%s'\n", client.Name, client.Type)

		// 直接JSON序列化测试
		jsonData, err := json.MarshalIndent(client, "", "  ")
		if err != nil {
			log.Printf("❌ JSON序列化失败: %v", err)
		} else {
			fmt.Printf("✅ JSON序列化结果:\n%s\n", string(jsonData))
		}

		// 创建模拟的gin.Context
		c, _ := gin.CreateTestContext(nil)
		c.Request = &http.Request{}
		c.Request.Header = make(http.Header)

		// 模拟API响应包装
		apiResponse := common.APIResponse{
			Success: true,
			Data:    client,
			Meta: common.ResponseMeta{
				Timestamp:   time.Now(),
				Version:     "v1",
				Server:      "law-oa-go",
				Environment: "development",
			},
		}

		apiJsonData, err := json.MarshalIndent(apiResponse, "", "  ")
		if err != nil {
			log.Printf("❌ API响应序列化失败: %v", err)
		} else {
			fmt.Printf("✅ 完整API响应:\n%s\n", string(apiJsonData))
		}
	}

	// 测试2: 模拟列表API响应
	fmt.Println("\n📋 测试2: 模拟ListClients API响应")
	listReq := &services.ClientListRequest{
		Page:     1,
		PageSize: 1,
		Search:   "张三",
	}

	clients, total, err := clientService.ListClients(context.Background(), listReq)
	if err != nil {
		log.Printf("❌ Service列表查询失败: %v", err)
	} else {
		fmt.Printf("✅ Service列表结果: 找到%d条记录\n", len(clients))
		for i, c := range clients {
			fmt.Printf("   %d. Name:'%s', Type:'%s'\n", i+1, c.Name, c.Type)
		}

		// 检查每个客户的字段
		for i, client := range clients {
			fmt.Printf("\n🔍 客户%d字段检查:\n", i+1)
			fmt.Printf("   ID: %d\n", client.ID)
			fmt.Printf("   Name: '%s' (长度:%d)\n", client.Name, len(client.Name))
			fmt.Printf("   Type: '%s' (长度:%d)\n", client.Type, len(client.Type))
			fmt.Printf("   Email: '%s' (长度:%d)\n", client.Email, len(client.Email))
			fmt.Printf("   Phone: '%s' (长度:%d)\n", client.Phone, len(client.Phone))

			// 单独序列化这个客户
			singleJson, _ := json.MarshalIndent(client, "", "  ")
			fmt.Printf("   单独JSON:\n%s\n", string(singleJson))
		}

		// 模拟完整API响应
		pagination := &common.PaginationInfo{
			Page:       1,
			PageSize:   1,
			Total:      total,
			TotalPages: int((total + 1 - 1) / 1),
			HasNext:    total > 1,
			HasPrev:    false,
		}

		listApiResponse := common.APIResponse{
			Success:    true,
			Data:       clients,
			Pagination: pagination,
			Meta: common.ResponseMeta{
				Timestamp:   time.Now(),
				Version:     "v1",
				Server:      "law-oa-go",
				Environment: "development",
			},
		}

		listJsonData, err := json.MarshalIndent(listApiResponse, "", "  ")
		if err != nil {
			log.Printf("❌ 列表API响应序列化失败: %v", err)
		} else {
			fmt.Printf("✅ 完整列表API响应:\n%s\n", string(listJsonData))
		}
	}

	fmt.Println("\n🎯 API序列化测试结论:")
	fmt.Println("   如果JSON序列化显示正确的name，")
	fmt.Println("   但API响应显示空name，")
	fmt.Println("   问题在于Gin框架或中间件处理")
}