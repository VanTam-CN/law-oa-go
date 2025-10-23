package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// PostgreSQL数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("=== 创建真实冲突检测测试数据 ===")

	// 1. 确认张伟律师存在
	var lawyerID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'zhangwei' AND role = 'lawyer'").Scan(&lawyerID)
	if err != nil {
		log.Fatal("张伟律师不存在:", err)
	}
	fmt.Printf("张伟律师ID: %d\n", lawyerID)

	// 2. 创建阿里巴巴客户
	var alibabaClientID int
	err = db.QueryRow("SELECT id FROM clients WHERE company = '阿里巴巴集团'").Scan(&alibabaClientID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("创建阿里巴巴客户...")
			err = db.QueryRow(`
				INSERT INTO clients (name, company, phone, email, type, address, industry, created_at, updated_at)
				VALUES ('阿里巴巴集团', '阿里巴巴集团', '400-800-1688', 'legal@alibaba.com', '企业', '杭州市余杭区文一西路969号', '电子商务', NOW(), NOW())
				RETURNING id
			`).Scan(&alibabaClientID)
			if err != nil {
				log.Fatal("创建阿里巴巴客户失败:", err)
			}
		} else {
			log.Fatal("查询阿里巴巴客户失败:", err)
		}
	}
	fmt.Printf("阿里巴巴客户ID: %d\n", alibabaClientID)

	// 3. 创建阿里巴巴相关的案件（张伟律师代理）
	var case1ID int
	err = db.QueryRow("SELECT id FROM cases WHERE title = '阿里巴巴诉字节跳动不正当竞争纠纷案'").Scan(&case1ID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("创建阿里巴巴诉字节跳动案件...")
			err = db.QueryRow(`
				INSERT INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date, created_at, updated_at)
				VALUES (
					'阿里巴巴诉字节跳动不正当竞争纠纷案', 
					'阿里巴巴诉字节跳动不正当竞争案件，涉及电商业务竞争，认为字节跳动在抖音平台存在不正当竞争行为', 
					$1, $2, '商事', 'high', 'in_progress', NOW() - INTERVAL '30 days', NOW(), NOW()
				)
				RETURNING id
			`, alibabaClientID, lawyerID).Scan(&case1ID)
			if err != nil {
				log.Fatal("创建案件1失败:", err)
			}
		} else {
			log.Fatal("查询案件1失败:", err)
		}
	}
	fmt.Printf("案件1 ID: %d\n", case1ID)

	// 4. 创建字节跳动客户
	var bytedanceClientID int
	err = db.QueryRow("SELECT id FROM clients WHERE company = '字节跳动科技有限公司'").Scan(&bytedanceClientID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("创建字节跳动客户...")
			err = db.QueryRow(`
				INSERT INTO clients (name, company, phone, email, type, address, industry, created_at, updated_at)
				VALUES ('字节跳动科技有限公司', '字节跳动科技有限公司', '400-690-0000', 'legal@bytedance.com', '企业', '北京市海淀区北三环西路甲18号院', '互联网科技', NOW(), NOW())
				RETURNING id
			`).Scan(&bytedanceClientID)
			if err != nil {
				log.Fatal("创建字节跳动客户失败:", err)
			}
		} else {
			log.Fatal("查询字节跳动客户失败:", err)
		}
	}
	fmt.Printf("字节跳动客户ID: %d\n", bytedanceClientID)

	// 5. 验证创建的数据
	fmt.Println("\n=== 验证创建的测试数据 ===")
	
	rows, err := db.Query(`
		SELECT c.title, cl.company as client_name, c.case_type, c.status
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		ORDER BY c.created_at DESC
	`, lawyerID)
	if err != nil {
		log.Fatal("查询验证数据失败:", err)
	}
	defer rows.Close()

	caseCount := 0
	fmt.Println("\n张伟律师当前代理的案件:")
	for rows.Next() {
		var title, clientName, caseType, status string
		err := rows.Scan(&title, &clientName, &caseType, &status)
		if err != nil {
			log.Printf("扫描验证数据失败: %v", err)
			continue
		}

		caseCount++
		fmt.Printf("%d. %s\n", caseCount, title)
		fmt.Printf("   - 客户: %s\n", clientName)
		fmt.Printf("   - 类型: %s\n", caseType)
		fmt.Printf("   - 状态: %s\n", status)
		fmt.Println()
	}

	fmt.Printf("\n=== 测试场景设置完成 ===\n")
	fmt.Printf("已创建 %d 个案件用于冲突检测测试\n", caseCount)
	fmt.Println("\n🎯 测试场景:")
	fmt.Println("1. 使用张伟律师账号登录 (zhangwei / law123456)")
	fmt.Println("2. 创建新案件:")
	fmt.Println("   - 案件名称: 字节跳动诉阿里巴巴垄断纠纷案")
	fmt.Println("   - 客户: 字节跳动科技有限公司")
	fmt.Println("   - 对方当事人: 阿里巴巴集团")
	fmt.Println("   - 案件类型: 商事案件")
	fmt.Println("3. 系统应该检测到商业竞争冲突")
	fmt.Println("4. 因为张伟已经代理了阿里巴巴诉字节跳动的案件")

	fmt.Printf("\n测试数据创建完成时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
}
