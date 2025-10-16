package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// PostgreSQL数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	fmt.Println("=== 检查现有数据状态 ===")

	// 1. 确认张伟律师
	var lawyerID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'zhangwei' AND role = 'lawyer'").Scan(&lawyerID)
	if err != nil {
		log.Fatal("张伟律师不存在:", err)
	}
	fmt.Printf("张伟律师ID: %d\n", lawyerID)

	// 2. 查询现有客户
	fmt.Println("\n=== 现有客户 ===")
	clientRows, err := db.Query(`
		SELECT id, name, company, type, industry
		FROM clients 
		WHERE company ILIKE '%阿里%' OR company ILIKE '%字节%' OR company ILIKE '%腾讯%'
		ORDER BY company
	`)
	if err != nil {
		log.Fatal("查询客户失败:", err)
	}
	defer clientRows.Close()

	clients := make(map[string]int)
	for clientRows.Next() {
		var id int
		var name, company, clientType, industry sql.NullString
		err := clientRows.Scan(&id, &name, &company, &clientType, &industry)
		if err != nil {
			log.Printf("扫描客户数据失败: %v", err)
			continue
		}

		clientName := company.String
		if clientName == "" {
			clientName = name.String
		}
		clients[clientName] = id
		
		fmt.Printf("客户: %s (ID: %d, 类型: %s, 行业: %s)\n", 
			clientName, id, clientType.String, industry.String)
	}

	// 3. 查询张伟律师的案件
	fmt.Println("\n=== 张伟律师现有案件 ===")
	caseRows, err := db.Query(`
		SELECT c.id, c.title, cl.company as client_name, c.case_type, c.status, c.description
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		ORDER BY c.created_at DESC
	`, lawyerID)
	if err != nil {
		log.Fatal("查询案件失败:", err)
	}
	defer caseRows.Close()

	caseCount := 0
	for caseRows.Next() {
		var id int
		var title, clientName, caseType, status, description sql.NullString
		err := caseRows.Scan(&id, &title, &clientName, &caseType, &status, &description)
		if err != nil {
			log.Printf("扫描案件数据失败: %v", err)
			continue
		}

		caseCount++
		fmt.Printf("案件%d: %s (ID: %d)\n", caseCount, title.String, id)
		fmt.Printf("  - 客户: %s\n", clientName.String)
		fmt.Printf("  - 类型: %s\n", caseType.String)
		fmt.Printf("  - 状态: %s\n", status.String)
		if description.Valid {
			fmt.Printf("  - 描述: %s\n", description.String)
		}
		fmt.Println()
	}

	// 4. 分析潜在冲突场景
	fmt.Println("=== 潜在冲突场景分析 ===")
	
	// 检查是否有互联网行业的案件
	hasInternetCases := false
	caseRows2, err := db.Query(`
		SELECT c.title, cl.company as client_name
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1 AND (
			cl.industry ILIKE '%互联网%' OR 
			cl.industry ILIKE '%电子商务%' OR 
			cl.industry ILIKE '%科技%' OR
			cl.company ILIKE '%阿里%' OR 
			cl.company ILIKE '%字节%' OR 
			cl.company ILIKE '%腾讯%'
		)
	`, lawyerID)
	if err == nil {
		for caseRows2.Next() {
			var title, clientName sql.NullString
			err := caseRows2.Scan(&title, &clientName)
			if err == nil {
				hasInternetCases = true
				fmt.Printf("✓ 发现互联网相关案件: %s (客户: %s)\n", 
					title.String, clientName.String)
			}
		}
		caseRows2.Close()
	}

	if hasInternetCases {
		fmt.Println("\n🎯 推荐测试场景:")
		fmt.Println("1. 为腾讯创建新案件，对方填写 '阿里巴巴'")
		fmt.Println("2. 为阿里巴巴创建新案件，对方填写 '腾讯'")
		fmt.Println("3. 为字节跳动创建新案件，对方填写 '阿里巴巴'")
		fmt.Println("4. 应该都能检测到商业竞争冲突")
	} else {
		fmt.Println("\n⚠️ 未发现互联网相关案件，需要创建测试数据")
		fmt.Println("建议创建互联网行业的案件来测试冲突检测功能")
	}

	fmt.Printf("\n数据检查完成时间: %s\n", "2025-10-16 12:53:30")
}
