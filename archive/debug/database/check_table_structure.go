package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔍 检查clients表结构")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer db.Close()

	// 查询表结构
	fmt.Println("📋 clients表结构:")
	query := "DESCRIBE clients"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("❌ 查询表结构失败: %v", err)
	}
	defer rows.Close()

	fmt.Println("   字段名          | 类型            | 是否为空 | 默认值")
	fmt.Println("   --------------- | --------------- | ------- | -------")

	for rows.Next() {
		var field, fieldType, null, key, extra string
		var defaultValue sql.NullString

		err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
		if err != nil {
			log.Printf("❌ 扫描行失败: %v", err)
			continue
		}

		defaultVal := "NULL"
		if defaultValue.Valid {
			defaultVal = defaultValue.String
		}

		fmt.Printf("   %-15s | %-15s | %-7s | %-10s\n", field, fieldType, null, defaultVal)
	}

	fmt.Println("\n🔍 检查name和type字段是否存在:")

	// 检查name字段
	var nameExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'law_oa'
		AND TABLE_NAME = 'clients'
		AND COLUMN_NAME = 'name'
	`).Scan(&nameExists)

	if err != nil {
		log.Printf("❌ 检查name字段失败: %v", err)
	} else {
		status := "❌ 不存在"
		if nameExists {
			status = "✅ 存在"
		}
		fmt.Printf("   name字段: %s\n", status)
	}

	// 检查type字段
	var typeExists bool
	err = db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = 'law_oa'
		AND TABLE_NAME = 'clients'
		AND COLUMN_NAME = 'type'
	`).Scan(&typeExists)

	if err != nil {
		log.Printf("❌ 检查type字段失败: %v", err)
	} else {
		status := "❌ 不存在"
		if typeExists {
			status = "✅ 存在"
		}
		fmt.Printf("   type字段: %s\n", status)
	}

	fmt.Println("\n🎯 表结构检查完成")
}