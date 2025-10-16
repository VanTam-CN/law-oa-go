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
	// 数据库连接字符串 - 使用调试工具中验证成功的连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("=== 客户类型修复工具 ===")

	// 获取所有客户数据
	var clients []models.Client
	if err := db.Find(&clients).Error; err != nil {
		log.Fatal("查询客户数据失败:", err)
	}

	fmt.Printf("找到 %d 个客户记录\n\n", len(clients))

	fixCount := 0
	for i, client := range clients {
		originalType := client.Type
		newType := determineCorrectClientType(client)

		if originalType != newType {
			fmt.Printf("修复客户 %d:\n", i+1)
			fmt.Printf("  ID: %d\n", client.ID)
			fmt.Printf("  名称: '%s'\n", client.Name)
			fmt.Printf("  原类型: '%s'\n", originalType)
			fmt.Printf("  新类型: '%s'\n", newType)
			fmt.Printf("  公司名称: '%s'\n", client.Company)
			fmt.Printf("  理由: %s\n", getFixReason(client, originalType, newType))
			fmt.Println("---")

			// 更新数据库记录
			if err := db.Model(&client).Update("type", newType).Error; err != nil {
				fmt.Printf("❌ 更新失败: %v\n", err)
			} else {
				fmt.Printf("✅ 更新成功\n")
				fixCount++
			}
			fmt.Println()
		}
	}

	fmt.Printf("=== 修复完成 ===\n")
	fmt.Printf("总共修复了 %d 个客户记录\n", fixCount)

	// 显示修复后的统计
	var fixedClients []models.Client
	if err := db.Find(&fixedClients).Error; err != nil {
		log.Printf("查询修复后数据失败: %v", err)
		return
	}

	typeStats := make(map[string]int)
	for _, client := range fixedClients {
		typeStats[client.Type]++
	}

	fmt.Println("\n=== 修复后类型统计 ===")
	for clientType, count := range typeStats {
		fmt.Printf("类型 '%s': %d 个客户\n", clientType, count)
	}
}

// determineCorrectClientType 根据客户信息智能判断正确的客户类型
func determineCorrectClientType(client models.Client) string {
	// 如果类型已经正确，直接返回
	if client.Type == "个人" || client.Type == "企业" {
		// 进行二次验证，确保类型与数据一致
		if client.Type == "企业" && (client.Company == "" || strings.Contains(client.Name, "先生") || strings.Contains(client.Name, "女士") || strings.Contains(client.Name, "总")) {
			// 明显是个人名称但被标记为企业，修正为个人
			return "个人"
		}
		if client.Type == "个人" && client.Company != "" && !strings.Contains(client.Name, "先生") && !strings.Contains(client.Name, "女士") && !strings.Contains(client.Name, "总") {
			// 可能是企业但被标记为个人，检查公司名称
			if strings.Contains(client.Name, "公司") || strings.Contains(client.Name, "集团") || strings.Contains(client.Name, "有限") || strings.Contains(client.Name, "科技") {
				return "企业"
			}
		}
		return client.Type
	}

	// 如果类型为空，根据其他信息判断
	if client.Company != "" && client.Company != " " {
		return "企业"
	}

	// 根据名称特征判断
	name := client.Name
	if strings.Contains(name, "先生") || strings.Contains(name, "女士") || strings.Contains(name, "经理") || strings.Contains(name, "总") || strings.Contains(name, "先生") {
		return "个人"
	}

	if strings.Contains(name, "公司") || strings.Contains(name, "集团") || strings.Contains(name, "有限") || strings.Contains(name, "企业") || strings.Contains(name, "科技") || strings.Contains(name, "贸易") {
		return "企业"
	}

	// 默认为个人
	return "个人"
}

// getFixReason 获取修复理由
func getFixReason(client models.Client, originalType, newType string) string {
	name := client.Name

	if originalType == "个人" && newType == "企业" {
		if strings.Contains(name, "公司") || strings.Contains(name, "集团") {
			return fmt.Sprintf("客户名称 '%s' 包含企业特征词汇", name)
		}
		return fmt.Sprintf("客户有公司名称 '%s'", client.Company)
	}

	if originalType == "企业" && newType == "个人" {
		if strings.Contains(name, "先生") || strings.Contains(name, "女士") || strings.Contains(name, "总") {
			return fmt.Sprintf("客户名称 '%s' 是明显的个人姓名", name)
		}
		return "客户名称不符合企业特征"
	}

	return "根据客户信息综合判断"
}