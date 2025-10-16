package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
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

	// 查询ID为13的客户（张三）
	var client models.Client
	result := db.First(&client, 13)
	if result.Error != nil {
		fmt.Println("查询客户失败:", result.Error)
		return
	}

	fmt.Printf("📋 客户ID: %d 的详细信息:\n", client.ID)
	fmt.Printf("- Name: '%s'\n", client.Name)
	fmt.Printf("- Type: '%s'\n", client.Type)
	fmt.Printf("- Email: '%s'\n", client.Email)
	fmt.Printf("- Phone: '%s'\n", client.Phone)
	fmt.Printf("- Address: '%s'\n", client.Address)
	fmt.Printf("- IDCard: '%s'\n", client.IDCard)
	fmt.Printf("- Company: '%s'\n", client.Company)
	fmt.Printf("- Industry: '%s'\n", client.Industry)
	fmt.Printf("- ContactPerson: '%s'\n", client.ContactPerson)
	fmt.Printf("- ContactPhone: '%s'\n", client.ContactPhone)
	fmt.Printf("- Source: '%s'\n", client.Source)
	fmt.Printf("- Notes: '%s'\n", client.Notes)
	fmt.Printf("- Status: '%s'\n", client.Status)

	// 检查字段是否为空
	fmt.Println("\n🔍 字段检查:")
	if client.Type == "" {
		fmt.Println("⚠️ Type 字段为空!")
	} else {
		fmt.Printf("✅ Type 字段正常: %s\n", client.Type)
	}

	// 查询多个客户测试
	var clients []models.Client
	db.Limit(3).Order("created_at DESC").Find(&clients)

	fmt.Printf("\n📊 查询到 %d 个客户:\n", len(clients))
	for i, c := range clients {
		fmt.Printf("%d. ID:%d Name:%s Type:'%s' Email:%s\n", i+1, c.ID, c.Name, c.Type, c.Email)
	}
}