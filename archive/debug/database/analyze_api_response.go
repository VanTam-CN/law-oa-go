package main

import (
	"encoding/json"
	"fmt"
)

// 根据HTTP API响应推断的结构体
type HTTPClientResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Company   string    `json:"company"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	// 注意：缺少 Type, IDCard, Industry, ContactPerson, ContactPhone, Source 字段
}

func main() {
	// 模拟从HTTP API返回的数据
	apiData := `{
		"id": 20,
		"name": "吴十电商",
		"email": "service@wushi-ecom.com",
		"phone": "023-77777777",
		"address": "重庆市渝中区解放碑步行街168号",
		"company": "吴十电商",
		"notes": "电商企业客户",
		"status": "active",
		"created_at": "2025-10-13T10:27:49+08:00",
		"updated_at": "2025-10-13T10:27:49+08:00"
	}`

	var httpClient HTTPClientResponse
	err := json.Unmarshal([]byte(apiData), &httpClient)
	if err != nil {
		fmt.Println("解析失败:", err)
		return
	}

	fmt.Println("📋 HTTP API响应结构分析:")
	fmt.Printf("- ID: %d\n", httpClient.ID)
	fmt.Printf("- Name: %s\n", httpClient.Name)
	fmt.Printf("- Email: %s\n", httpClient.Email)
	fmt.Printf("- Phone: %s\n", httpClient.Phone)
	fmt.Printf("- Address: %s\n", httpClient.Address)
	fmt.Printf("- Company: %s\n", httpClient.Company)
	fmt.Printf("- Notes: %s\n", httpClient.Notes)
	fmt.Printf("- Status: %s\n", httpClient.Status)
	fmt.Printf("- CreatedAt: %s\n", httpClient.CreatedAt)
	fmt.Printf("- UpdatedAt: %s\n", httpClient.UpdatedAt)

	fmt.Println("\n⚠️ 缺失的字段:")
	fmt.Println("- Type (客户类型)")
	fmt.Println("- IDCard (身份证号)")
	fmt.Println("- Industry (行业)")
	fmt.Println("- ContactPerson (联系人)")
	fmt.Println("- ContactPhone (联系电话)")
	fmt.Println("- Source (客户来源)")

	fmt.Println("\n🔍 推测:")
	fmt.Println("HTTP API返回的是一个不完整的客户结构体")
	fmt.Println("可能存在两个问题:")
	fmt.Println("1. 某个地方使用了错误的Client结构体定义")
	fmt.Println("2. JSON序列化过程中过滤掉了某些字段")
	fmt.Println("3. 存在字段映射或覆盖的问题")
}