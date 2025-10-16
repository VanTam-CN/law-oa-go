package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	// 数据库连接配置
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 查询客户数据并验证类型字段
	var clients []models.Client
	result := db.Select("id", "name", "type", "phone", "email", "status").Find(&clients)
	if result.Error != nil {
		log.Fatal("查询客户失败:", result.Error)
	}

	fmt.Printf("📊 客户总数: %d\n", len(clients))

	if len(clients) == 0 {
		fmt.Println("⚠️ 没有找到客户数据")
		return
	}

	fmt.Println("\n📋 客户类型验证:")

	personalCount := 0
	enterpriseCount := 0

	for _, client := range clients {
		fmt.Printf("- ID:%d %s (类型: %s) - %s - %s\n",
			client.ID, client.Name, client.Type, client.Phone, client.Email)

		if client.Type == "个人" {
			personalCount++
		} else if client.Type == "企业" {
			enterpriseCount++
		}
	}

	fmt.Printf("\n📈 类型统计:\n")
	fmt.Printf("- 个人客户: %d\n", personalCount)
	fmt.Printf("- 企业客户: %d\n", enterpriseCount)

	// 检查是否有无效类型
	invalidTypes := db.Where("type NOT IN ('个人', '企业') AND type IS NOT NULL AND type != ''").Find(&clients)
	if invalidTypes.Error == nil && len(clients) > 0 {
		fmt.Printf("\n⚠️ 发现无效类型的客户:\n")
		for _, client := range clients {
			fmt.Printf("- ID:%d %s (类型: %s)\n", client.ID, client.Name, client.Type)
		}
	}

	// 检查空类型
	emptyTypes := db.Where("type IS NULL OR type = ''").Find(&clients)
	if emptyTypes.Error == nil && len(clients) > 0 {
		fmt.Printf("\n⚠️ 发现空类型的客户:\n")
		for _, client := range clients {
			fmt.Printf("- ID:%d %s (类型: '%s')\n", client.ID, client.Name, client.Type)
		}
	}

	fmt.Println("\n🎉 客户类型验证完成！")

	if personalCount > 0 || enterpriseCount > 0 {
		fmt.Println("✅ 客户类型字段工作正常，前端应该能正确显示客户类型")
	} else {
		fmt.Println("❌ 客户类型数据有问题，需要进一步检查")
	}
}