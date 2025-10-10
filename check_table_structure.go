package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接配置
	dsn := "law_oa:law_oa_password@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("=== 案件表结构 ===")
	rows, err := db.Query("DESCRIBE cases")
	if err != nil {
		log.Fatal("查询案件表结构失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key, def, extra sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err != nil {
			log.Printf("扫描表结构失败: %v", err)
			continue
		}
		fmt.Printf("字段: %s, 类型: %s, 空值: %s, 键: %s, 默认值: %s, 额外: %s\n",
			field.String, typ.String, null.String, key.String, def.String, extra.String)
	}

	fmt.Println("\n=== 案件表数据（修正查询）===")
	rows, err = db.Query("SELECT * FROM cases LIMIT 10")
	if err != nil {
		log.Fatal("查询案件表数据失败:", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		log.Fatal("获取列名失败:", err)
	}
	fmt.Printf("列名: %v\n", columns)

	caseCount := 0
	for rows.Next() {
		caseCount++
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			log.Printf("扫描案件数据失败: %v", err)
			continue
		}

		fmt.Printf("记录 %d:\n", caseCount)
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			fmt.Printf("  %s: %v\n", col, v)
		}
		fmt.Println()
	}

	if caseCount == 0 {
		fmt.Println("案件表中没有数据")
	}

	fmt.Println("\n=== 客户表结构 ===")
	rows, err = db.Query("DESCRIBE clients")
	if err != nil {
		log.Fatal("查询客户表结构失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, typ, null, key, def, extra sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err != nil {
			log.Printf("扫描客户表结构失败: %v", err)
			continue
		}
		fmt.Printf("字段: %s, 类型: %s, 空值: %s, 键: %s, 默认值: %s, 额外: %s\n",
			field.String, typ.String, null.String, key.String, def.String, extra.String)
	}
}