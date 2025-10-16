package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接
	db, err := sql.Open("mysql", "law_user:law_password@tcp(localhost:3307)/law_oa_system?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("=== 查询律师张伟的代理关系 ===")

	// 查询张伟律师的信息
	var lawyerID int
	err = db.QueryRow("SELECT id FROM lawyers WHERE name = '张伟' AND username = 'zhangwei'").Scan(&lawyerID)
	if err != nil {
		log.Fatal("查询张伟律师失败:", err)
	}

	fmt.Printf("张伟律师ID: %d\n", lawyerID)

	// 查询张伟律师代理的所有案件
	fmt.Println("\n=== 张伟律师代理的案件 ===")
	rows, err := db.Query(`
		SELECT c.id, c.case_name, c.case_type, c.opposing_party, cl.name as client_name, cl.company as client_company
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = ? AND c.status != 'closed'
		ORDER BY c.created_at DESC
	`, lawyerID)
	if err != nil {
		log.Fatal("查询案件失败:", err)
	}
	defer rows.Close()

	caseCount := 0
	for rows.Next() {
		var caseID int
		var caseName, caseType, opposingParty, clientName, clientCompany sql.NullString

		err := rows.Scan(&caseID, &caseName, &caseType, &opposingParty, &clientName, &clientCompany)
		if err != nil {
			log.Printf("扫描案件数据失败: %v", err)
			continue
		}

		caseCount++
		clientDisplayName := "未知客户"
		if clientName.Valid {
			clientDisplayName = clientName.String
		} else if clientCompany.Valid {
			clientDisplayName = clientCompany.String
		}

		opposingPartyName := "未填写"
		if opposingParty.Valid && opposingParty.String != "" {
			opposingPartyName = opposingParty.String
		}

		fmt.Printf("案件%d: %s\n", caseCount, caseName.String)
		fmt.Printf("  - 客户: %s\n", clientDisplayName)
		fmt.Printf("  - 对方: %s\n", opposingPartyName)
		fmt.Printf("  - 类型: %s\n", caseType.String)
		fmt.Printf("  - ID: %d\n", caseID)
		fmt.Println()
	}

	if caseCount == 0 {
		fmt.Println("张伟律师暂无代理案件")
	}

	// 检查是否有阿里巴巴相关的案件
	fmt.Println("=== 检查互联网行业相关案件 ===")
	rows2, err := db.Query(`
		SELECT c.id, c.case_name, cl.name as client_name, cl.company as client_company, c.opposing_party
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = ? AND (
			cl.company LIKE '%阿里%' OR cl.company LIKE '%淘宝%' OR cl.company LIKE '%天猫%' OR
			cl.name LIKE '%阿里%' OR
			c.opposing_party LIKE '%阿里%' OR c.opposing_party LIKE '%淘宝%' OR c.opposing_party LIKE '%天猫%' OR
			c.case_name LIKE '%阿里%' OR c.case_name LIKE '%淘宝%' OR c.case_name LIKE '%天猫%'
		)
	`, lawyerID)
	if err != nil {
		log.Fatal("查询互联网案件失败:", err)
	}
	defer rows2.Close()

	techCaseCount := 0
	for rows2.Next() {
		var caseID int
		var caseName, clientName, clientCompany, opposingParty sql.NullString

		err := rows2.Scan(&caseID, &caseName, &clientName, &clientCompany, &opposingParty)
		if err != nil {
			log.Printf("扫描互联网案件数据失败: %v", err)
			continue
		}

		techCaseCount++
		clientDisplayName := "未知客户"
		if clientName.Valid {
			clientDisplayName = clientName.String
		} else if clientCompany.Valid {
			clientDisplayName = clientCompany.String
		}

		fmt.Printf("互联网案件%d: %s\n", techCaseCount, caseName.String)
		fmt.Printf("  - 客户: %s\n", clientDisplayName)
		if opposingParty.Valid && opposingParty.String != "" {
			fmt.Printf("  - 对方: %s\n", opposingParty.String)
		}
		fmt.Println()
	}

	if techCaseCount == 0 {
		fmt.Println("未找到阿里巴巴相关的案件")
		fmt.Println("这解释了为什么冲突检测没有发现行业竞争冲突")
	}

	fmt.Printf("\n=== 总结 ===\n")
	fmt.Printf("张伟律师当前代理案件数: %d\n", caseCount)
	fmt.Printf("其中互联网相关案件数: %d\n", techCaseCount)

	if techCaseCount == 0 {
		fmt.Println("\n💡 建议: 需要创建一些测试数据来模拟冲突检测场景")
	}
}