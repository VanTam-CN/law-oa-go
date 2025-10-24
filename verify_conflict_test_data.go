package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// 连接数据库
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("🔍 验证冲突检测测试数据...")

	// 1. 检查用户数据
	fmt.Println("\n📋 检查律师用户数据:")
	userQuery := `
		SELECT id, name, email, role 
		FROM users 
		WHERE name IN ('张伟', '陈浩', '王芳') 
		ORDER BY name
	`
	rows, err := db.Query(userQuery)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name, email, role string
			if err := rows.Scan(&id, &name, &email, &role); err == nil {
				fmt.Printf("   👤 %s (ID: %d, Email: %s, Role: %s)\n", name, id, email, role)
			}
		}
	}

	// 2. 检查客户数据
	fmt.Println("\n📋 检查客户数据:")
	clientQuery := `
		SELECT id, name, type, phone, email 
		FROM clients 
		WHERE name IN ('阿里巴巴集团控股有限公司', '字节跳动科技有限公司', '腾讯控股有限公司', 
		               '刘德华', '朱丽倩', '万科企业股份有限公司', '宝能集团', 
		               '中国建筑股份有限公司', '中国中铁股份有限公司') 
		ORDER BY name
	`
	rows, err = db.Query(clientQuery)
	if err != nil {
		log.Printf("查询客户失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var id int
			var name, clientType, phone, email string
			if err := rows.Scan(&id, &name, &clientType, &phone, &email); err == nil {
				fmt.Printf("   🏢 %s (ID: %d, Type: %s)\n", name, id, clientType)
			}
		}
	}

	// 3. 检查案件数据
	fmt.Println("\n📋 检查案件数据:")
	caseQuery := `
		SELECT c.id, c.title, c.case_type, cl.name as client_name, u.name as lawyer_name, c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.title LIKE '%阿里巴巴%' OR c.title LIKE '%字节跳动%' OR c.title LIKE '%腾讯%' 
		   OR c.title LIKE '%刘德华%' OR c.title LIKE '%朱丽倩%' 
		   OR c.title LIKE '%万科%' OR c.title LIKE '%宝能%'
		   OR c.title LIKE '%中国建筑%' OR c.title LIKE '%中国中铁%'
		ORDER BY c.created_at DESC
	`
	rows, err = db.Query(caseQuery)
	if err != nil {
		log.Printf("查询案件失败: %v", err)
	} else {
		defer rows.Close()
		caseCount := 0
		for rows.Next() {
			var id int
			var title, caseType, clientName, lawyerName, createdAt string
			if err := rows.Scan(&id, &title, &caseType, &clientName, &lawyerName, &createdAt); err == nil {
				fmt.Printf("   📁 案件ID: %d\n", id)
				fmt.Printf("      标题: %s\n", title)
				fmt.Printf("      类型: %s\n", caseType)
				fmt.Printf("      委托人: %s\n", clientName)
				fmt.Printf("      律师: %s\n", lawyerName)
				fmt.Printf("      创建时间: %s\n", createdAt)
				fmt.Println("      ---")
				caseCount++
			}
		}
		fmt.Printf("   总计找到 %d 个相关案件\n", caseCount)
	}

	// 4. 检查律师代理的案件分布
	fmt.Println("\n📋 检查律师代理案件分布:")
	lawyerCaseQuery := `
		SELECT u.name as lawyer_name, COUNT(c.id) as case_count, 
		       STRING_AGG(DISTINCT cl.name, ', ') as clients
		FROM users u
		LEFT JOIN cases c ON u.id = c.lawyer_id
		LEFT JOIN clients cl ON c.client_id = cl.id
		WHERE u.name IN ('张伟', '陈浩', '王芳')
		GROUP BY u.id, u.name
		ORDER BY u.name
	`
	rows, err = db.Query(lawyerCaseQuery)
	if err != nil {
		log.Printf("查询律师案件分布失败: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var lawyerName string
			var caseCount int
			var clients sql.NullString
			if err := rows.Scan(&lawyerName, &caseCount, &clients); err == nil {
				fmt.Printf("   👨‍💼 %s: %d个案件\n", lawyerName, caseCount)
				if clients.Valid && clients.String != "" {
					fmt.Printf("      委托人: %s\n", clients.String)
				}
			}
		}
	}

	// 5. 检查具体的冲突场景
	fmt.Println("\n📋 验证具体冲突场景:")

	// 场景1: 张伟律师的案件
	fmt.Println("\n🔍 场景1 - 张伟律师代理的案件:")
	zhangweiQuery := `
		SELECT c.title, cl.name as client_name, c.case_type, c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE u.name = '张伟'
		ORDER BY c.created_at DESC
	`
	rows, err = db.Query(zhangweiQuery)
	if err != nil {
		log.Printf("查询张伟案件失败: %v", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var title, clientName, caseType, createdAt string
			if err := rows.Scan(&title, &clientName, &caseType, &createdAt); err == nil {
				fmt.Printf("   📁 %s (委托人: %s, 类型: %s)\n", title, clientName, caseType)
				count++
			}
		}
		if count == 0 {
			fmt.Println("   ⚠️ 张伟律师没有案件记录")
		}
	}

	// 场景2: 陈浩律师的案件
	fmt.Println("\n🔍 场景2 - 陈浩律师代理的案件:")
	chenhaoQuery := `
		SELECT c.title, cl.name as client_name, c.case_type, c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE u.name = '陈浩'
		ORDER BY c.created_at DESC
	`
	rows, err = db.Query(chenhaoQuery)
	if err != nil {
		log.Printf("查询陈浩案件失败: %v", err)
	} else {
		defer rows.Close()
		count := 0
		for rows.Next() {
			var title, clientName, caseType, createdAt string
			if err := rows.Scan(&title, &clientName, &caseType, &createdAt); err == nil {
				fmt.Printf("   📁 %s (委托人: %s, 类型: %s)\n", title, clientName, caseType)
				count++
			}
		}
		if count == 0 {
			fmt.Println("   ⚠️ 陈浩律师没有案件记录")
		}
	}

	// 6. 检查冲突检测相关表
	fmt.Println("\n📋 检查冲突检测相关表:")

	// 检查冲突检测记录表
	conflictRecordQuery := `SELECT COUNT(*) FROM conflict_check_records`
	var conflictRecordCount int
	if err := db.QueryRow(conflictRecordQuery).Scan(&conflictRecordCount); err == nil {
		fmt.Printf("   📊 冲突检测记录: %d 条\n", conflictRecordCount)
	}

	// 检查客户关系表
	clientRelationQuery := `SELECT COUNT(*) FROM client_relations`
	var clientRelationCount int
	if err := db.QueryRow(clientRelationQuery).Scan(&clientRelationCount); err == nil {
		fmt.Printf("   📊 客户关系记录: %d 条\n", clientRelationCount)
	}

	// 检查冲突案例表
	conflictCaseQuery := `SELECT COUNT(*) FROM conflict_cases`
	var conflictCaseCount int
	if err := db.QueryRow(conflictCaseQuery).Scan(&conflictCaseCount); err == nil {
		fmt.Printf("   📊 冲突案例记录: %d 条\n", conflictCaseCount)
	}

	fmt.Println("\n✅ 数据验证完成")
}
