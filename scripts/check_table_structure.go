//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔍 检查数据库表结构")
	fmt.Println("===================")

	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 检查users表结构
	fmt.Println("\n📋 Users表结构:")
	checkTableStructure(db, "users")

	// 检查clients表结构
	fmt.Println("\n📋 Clients表结构:")
	checkTableStructure(db, "clients")

	// 检查cases表结构
	fmt.Println("\n📋 Cases表结构:")
	checkTableStructure(db, "cases")
}

func checkTableStructure(db *sql.DB, tableName string) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := db.Query(query, tableName)
	if err != nil {
		log.Printf("查询表 %s 结构失败: %v", tableName, err)
		return
	}
	defer rows.Close()

	fmt.Printf("表 %s:\n", tableName)
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault sql.NullString
		if err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault); err != nil {
			log.Printf("扫描列信息失败: %v", err)
			continue
		}

	 nullableStr := "NOT NULL"
		if isNullable.String == "YES" {
			nullableStr = "NULL"
		}

		defaultStr := ""
		if columnDefault.Valid {
			defaultStr = fmt.Sprintf(" DEFAULT %s", columnDefault.String)
		}

		fmt.Printf("  %s %s %s%s\n", columnName, dataType, nullableStr, defaultStr)
	}
}