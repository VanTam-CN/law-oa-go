package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 检查cases表结构
	fmt.Println("🔍 检查cases表结构")
	fmt.Println("====================================")

	rows, err := db.Query("DESCRIBE cases")
	if err != nil {
		log.Fatal("查询cases表结构失败:", err)
	}
	defer rows.Close()

	fmt.Println("📋 cases表结构:")
	fmt.Println("   字段名          | 类型            | 是否为空 | 默认值")
	fmt.Println("   --------------- | --------------- | ------- | -------")

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			continue
		}

		fmt.Printf("   %-15s | %-15s | %-7s | %s\n", field.String, fieldType.String, null.String, defaultVal.String)
	}

	// 检查具体字段是否存在
	fmt.Println("\n🔍 检查关键字段:")
	checkField(db, "name", "案件名称")
	checkField(db, "title", "案件标题")
	checkField(db, "case_name", "案件名称")
	checkField(db, "case_type", "案件类型")
	checkField(db, "client_id", "客户ID")
	checkField(db, "lawyer_id", "律师ID")
	checkField(db, "opposing_party", "对立方")
	checkField(db, "description", "描述")
	checkField(db, "status", "状态")

	fmt.Println("\n🎯 检查完成")
}

func checkField(db *sql.DB, fieldName, description string) {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = 'law_oa' AND table_name = 'cases' AND column_name = '%s'", fieldName)).Scan(&count)
	if err != nil {
		fmt.Printf("   %s (%s): ❌ 查询失败\n", fieldName, description)
		return
	}

	if count > 0 {
		fmt.Printf("   %s (%s): ✅ 存在\n", fieldName, description)
	} else {
		fmt.Printf("   %s (%s): ❌ 不存在\n", fieldName, description)
	}
}