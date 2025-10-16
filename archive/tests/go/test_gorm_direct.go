package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	fmt.Println("🔍 直接测试GORM查询")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ GORM数据库连接失败: %v", err)
	}

	// 测试1: 直接通过ID查询
	fmt.Println("📋 测试1: 直接查询ID=13")
	var client models.Client
	result := db.First(&client, 13)
	if result.Error != nil {
		log.Printf("❌ 查询失败: %v", result.Error)
	} else {
		fmt.Printf("✅ GORM查询结果:\n")
		fmt.Printf("   ID: %d\n", client.ID)
		fmt.Printf("   Name: '%s'\n", client.Name)
		fmt.Printf("   Type: '%s'\n", client.Type)
		fmt.Printf("   Email: '%s'\n", client.Email)
		fmt.Printf("   Phone: '%s'\n", client.Phone)
		fmt.Printf("   Status: '%s'\n", client.Status)
		fmt.Printf("   CreatedAt: %v\n", client.CreatedAt)
		fmt.Printf("   UpdatedAt: %v\n", client.UpdatedAt)
	}

	// 测试2: 搜索查询
	fmt.Println("\n📋 测试2: 搜索查询'张三'")
	var clients []models.Client
	result = db.Where("LOWER(name) LIKE ?", "%张三%").Find(&clients)
	if result.Error != nil {
		log.Printf("❌ 搜索查询失败: %v", result.Error)
	} else {
		fmt.Printf("✅ 搜索查询结果: 找到%d条记录\n", len(clients))
		for i, c := range clients {
			fmt.Printf("   %d. ID:%d Name:'%s' Type:'%s' Email:'%s'\n",
				i+1, c.ID, c.Name, c.Type, c.Email)
		}
	}

	// 测试3: 序列化测试
	fmt.Println("\n📋 测试3: 手动序列化为JSON")
	type ClientJSON struct {
		ID        uint      `json:"id"`
		Name      string    `json:"name"`
		Type      string    `json:"type"`
		Email     string    `json:"email"`
		Phone     string    `json:"phone"`
		Address   string    `json:"address"`
		Company   string    `json:"company"`
		Status    string    `json:"status"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	clientJSON := ClientJSON{
		ID:        client.ID,
		Name:      client.Name,
		Type:      client.Type,
		Email:     client.Email,
		Phone:     client.Phone,
		Address:   client.Address,
		Company:   client.Company,
		Status:    client.Status,
		CreatedAt: client.CreatedAt,
		UpdatedAt: client.UpdatedAt,
	}

	fmt.Printf("✅ 手动序列化结果:\n")
	fmt.Printf("   ID: %d\n", clientJSON.ID)
	fmt.Printf("   Name: '%s'\n", clientJSON.Name)
	fmt.Printf("   Type: '%s'\n", clientJSON.Type)
	fmt.Printf("   Email: '%s'\n", clientJSON.Email)
	fmt.Printf("   Phone: '%s'\n", clientJSON.Phone)

	// 测试4: 检查GORM模型字段映射
	fmt.Println("\n📋 测试4: 检查GORM模型字段")
	fmt.Printf("   models.Client字段 - Name: '%s', Type: '%s'\n", client.Name, client.Type)

	// 检查是否有字段映射问题
	if client.Name == "" {
		fmt.Println("   ❌ GORM查询结果中Name字段为空!")
	} else {
		fmt.Printf("   ✅ GORM查询结果中Name字段正确: '%s'\n", client.Name)
	}

	if client.Type == "" {
		fmt.Println("   ❌ GORM查询结果中Type字段为空!")
	} else {
		fmt.Printf("   ✅ GORM查询结果中Type字段正确: '%s'\n", client.Type)
	}

	fmt.Println("\n🎯 GORM测试结论:")
	if client.Name != "" && client.Type != "" {
		fmt.Println("   ✅ GORM查询正常，数据正确")
		fmt.Println("   🔍 问题可能在API响应序列化层")
	} else {
		fmt.Println("   ❌ GORM查询有问题，数据不正确")
		fmt.Println("   🔍 问题在数据库查询或模型映射层")
	}
}