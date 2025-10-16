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

	fmt.Println("🔍 检查cases表的case_type约束")

	// 查询现有案件的类型
	rows, err := db.Query("SELECT DISTINCT case_type FROM cases LIMIT 10")
	if err != nil {
		log.Fatal("查询案件类型失败:", err)
	}
	defer rows.Close()

	fmt.Println("📋 现有的案件类型:")
	for rows.Next() {
		var caseType string
		if err := rows.Scan(&caseType); err != nil {
			continue
		}
		fmt.Printf("   - %s\n", caseType)
	}

	// 尝试插入不同类型来测试约束
	testTypes := []string{"民事诉讼", "民事", "诉讼", "案件", "合同纠纷", "其他"}

	fmt.Println("\n🧪 测试不同案件类型:")
	for _, testType := range testTypes {
		// 先删除临时测试记录
		db.Exec("DELETE FROM cases WHERE title LIKE '测试案件_%'")

		_, err = db.Exec(`
			INSERT INTO cases (title, case_type, client_id, lawyer_id, description, status, priority, created_at, updated_at)
			VALUES (?, ?, 1, 1, ?, 'pending', 'low', NOW(), NOW())
		`, "测试案件_"+testType, testType, "测试案件描述")

		if err != nil {
			fmt.Printf("   ❌ %s: %v\n", testType, err)
		} else {
			fmt.Printf("   ✅ %s: 有效\n", testType)
			// 清理测试记录
			db.Exec("DELETE FROM cases WHERE title = ?", "测试案件_"+testType)
		}
	}

	fmt.Println("\n🎯 约束检查完成")
}