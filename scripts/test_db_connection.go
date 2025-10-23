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

	// 测试PostgreSQL连接
	fmt.Printf("数据库类型: %s\n", cfg.Database.Driver)
	fmt.Printf("数据库地址: %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("数据库名称: %s\n", cfg.Database.Database)

	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 测试基本查询
	var result int
	if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Fatalf("数据库查询失败: %v", err)
	}

	fmt.Println("✅ PostgreSQL 连接测试成功!")
}