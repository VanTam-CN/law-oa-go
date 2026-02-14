//go:build ignore

package main

import (
	"fmt"
	"log"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"

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

	fmt.Println("🔍 检查数据库表结构...")

	// 检查表是否存在
	tables := []string{
		"users", "clients", "cases", "conflict_rules",
		"conflict_cases", "conflict_check_records", "client_relations",
	}

	for _, table := range tables {
		if db.Migrator().HasTable(table) {
			fmt.Printf("✅ 表 '%s' 存在\n", table)
		} else {
			fmt.Printf("❌ 表 '%s' 不存在\n", table)
		}
	}

	// 检查用户表字段
	if db.Migrator().HasTable("users") {
		fmt.Println("\n📋 users表字段:")
		columns, err := db.Migrator().ColumnTypes("users")
		if err != nil {
			log.Printf("获取users表字段失败: %v", err)
		} else {
			for _, col := range columns {
				nullable, _ := col.Nullable()
				fmt.Printf("   - %s: %s (nullable: %v)\n", col.Name(), col.DatabaseTypeName(), nullable)
			}
		}
	}

	// 检查客户表字段
	if db.Migrator().HasTable("clients") {
		fmt.Println("\n📋 clients表字段:")
		columns, err := db.Migrator().ColumnTypes("clients")
		if err != nil {
			log.Printf("获取clients表字段失败: %v", err)
		} else {
			for _, col := range columns {
				nullable, _ := col.Nullable()
				fmt.Printf("   - %s: %s (nullable: %v)\n", col.Name(), col.DatabaseTypeName(), nullable)
			}
		}
	}

	// 检查案件表字段
	if db.Migrator().HasTable("cases") {
		fmt.Println("\n📋 cases表字段:")
		columns, err := db.Migrator().ColumnTypes("cases")
		if err != nil {
			log.Printf("获取cases表字段失败: %v", err)
		} else {
			for _, col := range columns {
				nullable, _ := col.Nullable()
				fmt.Printf("   - %s: %s (nullable: %v)\n", col.Name(), col.DatabaseTypeName(), nullable)
			}
		}
	}

	// 检查现有数据
	fmt.Println("\n📊 现有数据统计:")
	var userCount, clientCount, caseCount int64
	db.Model(&models.User{}).Count(&userCount)
	db.Model(&models.Client{}).Count(&clientCount)
	db.Model(&models.Case{}).Count(&caseCount)

	fmt.Printf("   用户数量: %d\n", userCount)
	fmt.Printf("   客户数量: %d\n", clientCount)
	fmt.Printf("   案件数量: %d\n", caseCount)
}