package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔧 修复 conflict_cases 表结构")
	fmt.Println("=============================")

	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 检查并添加缺少的列
	fmt.Println("\n📋 检查 conflict_cases 表结构...")
	checkAndAddMissingColumns(db)

	fmt.Println("\n✅ conflict_cases 表结构修复完成！")
	fmt.Println("==================================")
}

func checkAndAddMissingColumns(db *sql.DB) {
	// 定义需要检查的列
	columns := map[string]string{
		"case_type":      "VARCHAR(100)",
		"case_no":        "VARCHAR(255)",
		"case_status":    "VARCHAR(50)",
		"client_id":      "VARCHAR(255)",
		"risk_level":     "VARCHAR(50)",
		"conflict_type":  "VARCHAR(100)",
		"check_id":       "VARCHAR(255)",
		"case_id":        "VARCHAR(255)",
		"case_name":      "VARCHAR(500)",
		"description":    "TEXT",
		"opposing_parties": "TEXT[]",
		"conflict_details": "TEXT",
		"created_at":     "TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP",
	}

	tableName := "conflict_cases"

	for columnName, columnDef := range columns {
		var columnExists bool
		checkColumnSQL := fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = '%s' AND column_name = '%s')", tableName, columnName)
		err := db.QueryRow(checkColumnSQL).Scan(&columnExists)
		if err != nil {
			log.Printf("检查列 %s 失败: %v", columnName, err)
			continue
		}

		if !columnExists {
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDef)
			_, err := db.Exec(alterSQL)
			if err != nil {
				log.Printf("添加列 %s 失败: %v", columnName, err)
			} else {
				fmt.Printf("✅ 添加列 %s 成功\n", columnName)
			}
		} else {
			fmt.Printf("📋 列 %s 已存在\n", columnName)
		}
	}
}