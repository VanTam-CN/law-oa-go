package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 模拟查询参数
	opponent := "腾讯\n垄断纠纷案"
	mappedCaseType := "知识产权"
	userIDUint := uint(45)

	fmt.Println("🔍 简化SQL查询测试:")
	fmt.Println("=====================")
	fmt.Printf("对方当事人: '%s'\n", opponent)
	fmt.Printf("映射案件类型: '%s'\n", mappedCaseType)
	fmt.Printf("用户ID: %d\n", userIDUint)

	// 分别测试各个匹配条件
	fmt.Println("\n1. 测试案件类型匹配:")
	query1 := "SELECT COUNT(*) as count FROM cases WHERE deleted_at IS NULL AND case_type = ? AND lawyer_id != ?"
	var count1 int64
	db.Raw(query1, mappedCaseType, userIDUint).Scan(&count1)
	fmt.Printf("案件类型匹配数量: %d\n", count1)

	fmt.Println("\n2. 测试标题匹配:")
	query2 := "SELECT COUNT(*) as count FROM cases WHERE deleted_at IS NULL AND title ILIKE ? AND lawyer_id != ?"
	var count2 int64
	db.Raw(query2, "%"+opponent+"%", userIDUint).Scan(&count2)
	fmt.Printf("标题匹配数量: %d\n", count2)

	fmt.Println("\n3. 测试描述匹配:")
	query3 := "SELECT COUNT(*) as count FROM cases WHERE deleted_at IS NULL AND description ILIKE ? AND lawyer_id != ?"
	var count3 int64
	db.Raw(query3, "%"+opponent+"%", userIDUint).Scan(&count3)
	fmt.Printf("描述匹配数量: %d\n", count3)

	fmt.Println("\n4. 测试客户名称匹配:")
	query4 := `
		SELECT COUNT(*) as count
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.deleted_at IS NULL
		AND cl.name ILIKE ?
		AND c.lawyer_id != ?
	`
	var count4 int64
	db.Raw(query4, "%"+opponent+"%", userIDUint).Scan(&count4)
	fmt.Printf("客户名称匹配数量: %d\n", count4)

	fmt.Println("\n🎯 问题分析:")
	if count1 > 0 && count2 == 0 && count3 == 0 && count4 == 0 {
		fmt.Printf("❌ 严重问题: 查询返回了 %d 个案件，但都是仅凭案件类型匹配！\n", count1)
		fmt.Println("这些案件与对方当事人'腾讯\\n垄断纠纷案'毫无关系。")
		fmt.Println("SQL查询的OR逻辑导致所有'知识产权'类型案件都被错误匹配。")
	}

	// 查看具体被错误匹配的案件
	fmt.Println("\n5. 查看被错误匹配的案件样例:")
	query5 := `
		SELECT c.id, c.title, cl.name as client_name
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.deleted_at IS NULL
		AND c.case_type = ?
		AND c.lawyer_id != ?
		LIMIT 5
	`

	rows, err := db.Raw(query5, mappedCaseType, userIDUint).Rows()
	if err == nil {
		for rows.Next() {
			var id int
			var title, clientName string
			rows.Scan(&id, &title, &clientName)
			fmt.Printf("  - ID:%d, 标题:%s, 客户:%s\n", id, title, clientName)
		}
		rows.Close()
	}
}