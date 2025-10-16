package main

import (
	"fmt"
	"log"
	"strings"

	"law-oa-go/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// 尝试多个常见的数据库连接配置
	dsns := []string{
		"root:password@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local",
		"root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local",
		"root:root@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local",
		"root:123456@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local",
	}

	var db *gorm.DB
	var err error
	var successfulDSN string

	for _, dsn := range dsns {
		fmt.Printf("尝试连接数据库: %s\n", maskPassword(dsn))
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			fmt.Println("✅ 数据库连接成功!")
			successfulDSN = dsn
			break
		}
		fmt.Printf("❌ 连接失败: %v\n", err)
	}

	if err != nil {
		log.Fatal("所有数据库连接尝试均失败")
	}

	// 隐藏密码显示连接信息
	fmt.Printf("使用连接: %s\n", maskPassword(successfulDSN))

	// 检查客户数据结构
	fmt.Println("=== 检查客户数据 ===")

	var clients []models.Client
	if err := db.Find(&clients).Error; err != nil {
		log.Fatal("查询客户数据失败:", err)
	}

	fmt.Printf("找到 %d 个客户:\n\n", len(clients))

	for i, client := range clients {
		fmt.Printf("客户 %d:\n", i+1)
		fmt.Printf("  ID: %d\n", client.ID)
		fmt.Printf("  姓名: '%s'\n", client.Name)
		fmt.Printf("  类型: '%s'\n", client.Type)
		fmt.Printf("  邮箱: '%s'\n", client.Email)
		fmt.Printf("  电话: '%s'\n", client.Phone)
		fmt.Printf("  地址: '%s'\n", client.Address)
		fmt.Printf("  公司名称: '%s'\n", client.Company)
		fmt.Printf("  身份证号: '%s'\n", client.IDCard)
		fmt.Printf("  状态: '%s'\n", client.Status)
		fmt.Printf("  创建时间: %s\n", client.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("---")
	}

	// 检查类型统计
	fmt.Println("\n=== 类型统计 ===")
	typeStats := make(map[string]int)
	for _, client := range clients {
		typeStats[client.Type]++
	}

	for clientType, count := range typeStats {
		fmt.Printf("类型 '%s': %d 个客户\n", clientType, count)
	}

	// 检查是否有空名称的客户
	fmt.Println("\n=== 检查异常数据 ===")
	var emptyNameClients []models.Client
	if err := db.Where("name = '' OR name IS NULL").Find(&emptyNameClients).Error; err != nil {
		log.Printf("查询空名称客户失败: %v", err)
	} else {
		fmt.Printf("找到 %d 个空名称的客户\n", len(emptyNameClients))
	}

	// 检查类型为空或异常的客户
	var invalidTypeClients []models.Client
	if err := db.Where("type = '' OR type IS NULL OR type NOT IN ('个人', '企业')").Find(&invalidTypeClients).Error; err != nil {
		log.Printf("查询异常类型客户失败: %v", err)
	} else {
		fmt.Printf("找到 %d 个类型异常的客户\n", len(invalidTypeClients))
		for _, client := range invalidTypeClients {
			fmt.Printf("  ID: %d, 名称: '%s', 类型: '%s'\n", client.ID, client.Name, client.Type)
		}
	}
}

// maskPassword 隐藏密码显示
func maskPassword(dsn string) string {
	if strings.Contains(dsn, ":password@") {
		return strings.Replace(dsn, "password", "***", 1)
	} else if strings.Contains(dsn, ":root@") {
		return strings.Replace(dsn, "root", "***", 1)
	} else if strings.Contains(dsn, ":123456@") {
		return strings.Replace(dsn, "123456", "***", 1)
	}
	return dsn
}