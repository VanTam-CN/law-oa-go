package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接信息
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"

	// 连接数据库
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 读取SQL文件
	sqlFile := "scripts/postgresql-complete-schema.sql"
	content, err := ioutil.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("读取SQL文件失败:", err)
	}

	// 分割SQL语句
	sqlStatements := strings.Split(string(content), ";")

	// 执行SQL语句
	successCount := 0
	errorCount := 0
	for i, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}

		// 跳过扩展创建语句，因为可能已存在
		if strings.Contains(stmt, "CREATE EXTENSION") {
			continue
		}

		// 执行SQL语句
		_, err := db.Exec(stmt)
		if err != nil {
			// 忽略已存在的错误
			if strings.Contains(err.Error(), "already exists") ||
			   strings.Contains(err.Error(), "duplicate") ||
			   strings.Contains(err.Error(), "does not exist") {
				fmt.Printf("⚠️  语句 %d: 已存在或忽略 - %s\n", i+1, err.Error())
				successCount++
			} else {
				fmt.Printf("❌ 语句 %d 执行失败: %s\n", i+1, err.Error())
				fmt.Printf("   SQL: %s\n", stmt[:min(100, len(stmt))])
				errorCount++
			}
		} else {
			if !strings.HasPrefix(stmt, "DO $$") && !strings.HasPrefix(stmt, "CREATE OR REPLACE") {
				fmt.Printf("✅ 语句 %d 执行成功\n", i+1)
			}
			successCount++
		}
	}

	fmt.Printf("\n📊 执行完成统计:\n")
	fmt.Printf("   成功: %d 个语句\n", successCount)
	fmt.Printf("   失败: %d 个语句\n", errorCount)

	// 验证关键表是否存在
	tables := []string{"conflict_check_records", "client_relations", "conflict_cases", "conflict_rules"}
	fmt.Println("\n🔍 验证关键表是否存在:")

	for _, table := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '%s'", table)).Scan(&count)
		if err != nil {
			fmt.Printf("   ❌ %s: 查询失败 - %s\n", table, err.Error())
		} else if count > 0 {
			fmt.Printf("   ✅ %s: 存在\n", table)
		} else {
			fmt.Printf("   ❌ %s: 不存在\n", table)
		}
	}

	fmt.Println("\n🎉 数据库表创建完成！")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}