package main

import (
	"fmt"
	"log"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	fmt.Println("🔍 检查数据库枚举值...")

	// 检查case_type枚举
	var caseTypes []string
	rows, err := db.Raw("SELECT unnest(enum_range(NULL::case_type))").Rows()
	if err != nil {
		log.Printf("获取case_type枚举值失败: %v", err)
	} else {
		for rows.Next() {
			var caseType string
			rows.Scan(&caseType)
			caseTypes = append(caseTypes, caseType)
		}
		rows.Close()

		fmt.Println("\n📋 case_type 可用枚举值:")
		for _, ct := range caseTypes {
			fmt.Printf("   - %s\n", ct)
		}
	}

	// 检查其他枚举类型
	enumTypes := []string{"user_role", "user_status", "client_type", "case_priority", "case_status"}

	for _, enumType := range enumTypes {
		query := fmt.Sprintf("SELECT unnest(enum_range(NULL::%s))", enumType)
		rows, err := db.Raw(query).Rows()
		if err != nil {
			log.Printf("获取%s枚举值失败: %v", enumType, err)
			continue
		}

		var values []string
		for rows.Next() {
			var value string
			rows.Scan(&value)
			values = append(values, value)
		}
		rows.Close()

		fmt.Printf("\n📋 %s 可用枚举值:\n", enumType)
		for _, v := range values {
			fmt.Printf("   - %s\n", v)
		}
	}
}