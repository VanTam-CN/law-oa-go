package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/common"
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

	// 测试完整的API响应构建
	fmt.Println("\n🔍 测试API响应构建:")

	// 调用service方法
	serviceResponses, total, err := clientService.ListClients(context.Background(), &services.ClientListRequest{
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		fmt.Println("Service调用失败:", err)
		return
	}

	fmt.Printf("Service返回: 客户数=%d, 总数=%d\n", len(serviceResponses), total)
	if len(serviceResponses) > 0 {
		response := serviceResponses[0]
		fmt.Printf("第一个客户: ID=%d, Name=%s, Type='%s'\n", response.ID, response.Name, response.Type)

		// 测试JSON序列化（模拟API响应）
		fmt.Println("\n🔍 测试API响应JSON序列化:")
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			fmt.Println("JSON序列化失败:", err)
			return
		}
		fmt.Println("单个客户响应JSON:")
		fmt.Println(string(jsonData))

		// 测试列表JSON序列化
		fmt.Println("\n🔍 测试客户列表JSON序列化:")
		listData, err := json.MarshalIndent(serviceResponses, "", "  ")
		if err != nil {
			fmt.Println("列表JSON序列化失败:", err)
			return
		}
		fmt.Println("客户列表JSON:")
		fmt.Println(string(listData))

		// 模拟完整的API响应结构
		fmt.Println("\n🔍 模拟完整API响应结构:")
		apiResponse := common.APIResponse{
			Success: true,
			Data:    serviceResponses,
		}

		fullResponse, err := json.MarshalIndent(apiResponse, "", "  ")
		if err != nil {
			fmt.Println("完整响应JSON序列化失败:", err)
			return
		}
		fmt.Println("完整API响应:")
		fmt.Println(string(fullResponse))
	}

	// 测试分页响应构建
	fmt.Println("\n🔍 测试分页响应构建:")
	page := 1
	pageSize := 1
	pagination := &common.PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasNext:    page < int((total + int64(pageSize) - 1) / int64(pageSize)),
		HasPrev:    page > 1,
	}

	paginatedResponse := common.APIResponse{
		Success:    true,
		Data:       serviceResponses,
		Pagination: pagination,
	}

	paginatedJson, err := json.MarshalIndent(paginatedResponse, "", "  ")
	if err != nil {
		fmt.Println("分页响应JSON序列化失败:", err)
		return
	}
	fmt.Println("分页API响应:")
	fmt.Println(string(paginatedJson))
}