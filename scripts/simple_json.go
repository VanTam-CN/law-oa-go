package main

import (
	"encoding/json"
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

	// 查询客户
	var client models.Client
	result := db.First(&client, 13) // 张三
	if result.Error != nil {
		fmt.Println("查询客户失败:", result.Error)
		return
	}

	fmt.Printf("📋 原始数据:\n")
	fmt.Printf("- ID: %d\n", client.ID)
	fmt.Printf("- Name: %s\n", client.Name)
	fmt.Printf("- Type: %s\n", client.Type)
	fmt.Printf("- Email: %s\n", client.Email)

	// 测试原始模型的JSON序列化
	fmt.Printf("\n🔍 原始模型JSON序列化:\n")
	jsonBytes, err := json.MarshalIndent(client, "", "  ")
	if err != nil {
		fmt.Println("JSON序列化失败:", err)
		return
	}
	fmt.Println(string(jsonBytes))

	// 检查字段是否存在
	fmt.Printf("\n🔍 字段检查:\n")
	fmt.Printf("- Type字段值: '%s'\n", client.Type)
	if client.Type == "" {
		fmt.Println("⚠️ Type字段为空!")
	} else {
		fmt.Println("✅ Type字段有值")
	}
}