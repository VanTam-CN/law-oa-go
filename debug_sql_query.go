package main

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestCase struct {
	ID          uint   `gorm:"column:id"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
	ClientName  string `gorm:"column:clients>name"`
}

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

	fmt.Println("🔍 模拟有问题的SQL查询:")
	fmt.Println("=====================================")
	fmt.Printf("对方当事人: '%s'\n", opponent)
	fmt.Printf("映射案件类型: '%s'\n", mappedCaseType)
	fmt.Printf("用户ID: %d\n", userIDUint)

	// 构建原始的有问题的SQL查询
	query := `
		SELECT
			c.id,
			c.title as case_name,
			c.case_type,
			c.description,
			cl.name as client_name,
			u.name as lawyer_name,
			c.created_at,
			c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.deleted_at IS NULL
		AND (c.title ILIKE ? OR c.description ILIKE ? OR cl.name ILIKE ? OR c.case_type = ?)
		AND c.lawyer_id != ?
		ORDER BY c.created_at DESC
		LIMIT 10
	`

	// 执行查询
	rows, err := db.WithContext(db.Statement.Context).Raw(query,
		"%"+opponent+"%", "%"+opponent+"%", "%"+opponent+"%", mappedCaseType, userIDUint).Rows()
	if err != nil {
		log.Printf("❌ 查询失败: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("\n📋 SQL查询结果:\n")

	recordCount := 0
	for rows.Next() {
		recordCount++
		var case_ TestCase
		var caseType string
		if err := rows.Scan(&case_.ID, &case_.Title, &caseType, &case_.Description, &case_.ClientName, nil, nil, nil); err != nil {
			continue
		}

		fmt.Printf("\n案件ID: %d\n", case_.ID)
		fmt.Printf("标题: %s\n", case_.Title)
		fmt.Printf("案件类型: %s\n", caseType)
		fmt.Printf("描述: %s\n", case_.Description)
		fmt.Printf("客户: %s\n", case_.ClientName)

		// 分析匹配原因
		fmt.Println("匹配原因分析:")
		titleMatch := strings.Contains(strings.ToLower(case_.Title), strings.ToLower(opponent))
		descMatch := strings.Contains(strings.ToLower(case_.Description), strings.ToLower(opponent))
		clientMatch := strings.Contains(strings.ToLower(case_.ClientName), strings.ToLower(opponent))
		typeMatch := caseType == mappedCaseType

		fmt.Printf("  标题匹配: %t\n", titleMatch)
		fmt.Printf("  描述匹配: %t\n", descMatch)
		fmt.Printf("  客户匹配: %t\n", clientMatch)
		fmt.Printf("  案件类型匹配: %t (类型: %s = %s)\n", typeMatch, caseType, mappedCaseType)

		if !titleMatch && !descMatch && !clientMatch && typeMatch {
			fmt.Printf("  ⚠️ 问题: 仅凭案件类型匹配被返回！\n")
		}
	}

	fmt.Printf("\n📋 总共处理了 %d 条记录\n", recordCount)

	fmt.Println("\n🎯 问题总结:")
	fmt.Println("问题在于SQL查询使用了OR逻辑，导致所有'知识产权'类型案件都被返回，")
	fmt.Println("即使它们与对方当事人'腾讯\\n垄断纠纷案'毫无关系。")
}