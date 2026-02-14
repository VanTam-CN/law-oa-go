//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	fmt.Println("=== 创建利益冲突测试数据 ===")

	// 首先检查现有数据
	fmt.Println("\n📊 检查现有数据...")
	var lawyerCount, clientCount, caseCount int

	db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'lawyer' OR role = 'LAWYER'").Scan(&lawyerCount)
	db.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clientCount)
	db.QueryRow("SELECT COUNT(*) FROM cases").Scan(&caseCount)

	fmt.Printf("现有律师数量: %d\n", lawyerCount)
	fmt.Printf("现有客户数量: %d\n", clientCount)
	fmt.Printf("现有案件数量: %d\n", caseCount)

	if lawyerCount == 0 || clientCount < 3 {
		fmt.Println("❌ 数据不足，需要至少1个律师和3个客户来创建冲突场景")
		return
	}

	// 获取第一个律师ID
	var lawyerID uint
	db.QueryRow("SELECT id FROM users WHERE role = 'lawyer' OR role = 'LAWYER' LIMIT 1").Scan(&lawyerID)
	fmt.Printf("使用律师ID: %d\n", lawyerID)

	// 创建冲突场景1：同一律师代理对立客户
	fmt.Println("\n🎯 场景1：同一律师代理对立客户（腾讯 vs 字节跳动）")

	// 获取或创建客户
	var tencentClientID, bytedanceClientID uint

	// 查找现有客户
	err = db.QueryRow("SELECT id FROM clients WHERE name LIKE '%腾讯%' LIMIT 1").Scan(&tencentClientID)
	if err != nil {
		// 创建腾讯客户
		err = db.QueryRow(`
			INSERT INTO clients (name, type, contact_person, phone, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			"腾讯科技有限公司", "COMPANY", "李经理", "13800138001", "legal@tencent.com", time.Now(), time.Now(),
		).Scan(&tencentClientID)
		if err != nil {
			log.Printf("创建腾讯客户失败: %v", err)
		} else {
			fmt.Printf("✅ 创建腾讯客户，ID: %d\n", tencentClientID)
		}
	} else {
		fmt.Printf("✅ 找到腾讯客户，ID: %d\n", tencentClientID)
	}

	err = db.QueryRow("SELECT id FROM clients WHERE name LIKE '%字节%' LIMIT 1").Scan(&bytedanceClientID)
	if err != nil {
		// 创建字节跳动客户
		err = db.QueryRow(`
			INSERT INTO clients (name, type, contact_person, phone, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			"字节跳动科技有限公司", "COMPANY", "王总", "13800138002", "legal@bytedance.com", time.Now(), time.Now(),
		).Scan(&bytedanceClientID)
		if err != nil {
			log.Printf("创建字节跳动客户失败: %v", err)
		} else {
			fmt.Printf("✅ 创建字节跳动客户，ID: %d\n", bytedanceClientID)
		}
	} else {
		fmt.Printf("✅ 找到字节跳动客户，ID: %d\n", bytedanceClientID)
	}

	// 创建冲突案件
	if tencentClientID > 0 && bytedanceClientID > 0 {
		// 为腾讯创建案件（已存在，跳过）
		var tencentCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", tencentClientID, lawyerID).Scan(&tencentCaseCount)

		if tencentCaseCount == 0 {
			_, err = db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"腾讯诉字节跳动不正当竞争案", "commercial",
				"腾讯公司起诉字节跳动公司涉嫌不正当竞争，要求停止侵权并赔偿损失。对方当事人：字节跳动科技有限公司",
				tencentClientID, lawyerID, "active", time.Now().AddDate(0, -6, 0), time.Now(),
			)
			if err != nil {
				log.Printf("创建腾讯案件失败: %v", err)
			} else {
				fmt.Println("✅ 创建腾讯案件")
			}
		} else {
			fmt.Printf("✅ 腾讯已有案件: %d个\n", tencentCaseCount)
		}

		// 为字节跳动创建新案件（这将触发冲突）
		var bytedanceCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", bytedanceClientID, lawyerID).Scan(&bytedanceCaseCount)

		if bytedanceCaseCount == 0 {
			_, err = db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"字节跳动诉腾讯滥用市场支配地位案", "commercial",
				"字节跳动公司起诉腾讯公司涉嫌滥用市场支配地位，垄断互联网市场。对方当事人：腾讯科技有限公司",
				bytedanceClientID, lawyerID, "active", time.Now(), time.Now(),
			)
			if err != nil {
				log.Printf("创建字节跳动案件失败: %v", err)
			} else {
				fmt.Println("✅ 创建字节跳动案件（这将触发冲突！)")
			}
		} else {
			fmt.Printf("✅ 字节跳动已有案件: %d个\n", bytedanceCaseCount)
		}
	}

	// 场景2：同一律师代理有利益关联的客户
	fmt.Println("\n🎯 场景2：同一律师代理关联公司客户")

	// 创建阿里巴巴和蚂蚁金服（关联公司）
	var alibabaClientID, antClientID uint

	err = db.QueryRow("SELECT id FROM clients WHERE name LIKE '%阿里巴巴%' LIMIT 1").Scan(&alibabaClientID)
	if err != nil {
		err = db.QueryRow(`
			INSERT INTO clients (name, type, contact_person, phone, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			"阿里巴巴集团控股有限公司", "COMPANY", "张法务", "13800138003", "legal@alibaba.com", time.Now(), time.Now(),
		).Scan(&alibabaClientID)
		if err == nil {
			fmt.Printf("✅ 创建阿里巴巴客户，ID: %d\n", alibabaClientID)
		}
	} else {
		fmt.Printf("✅ 找到阿里巴巴客户，ID: %d\n", alibabaClientID)
	}

	err = db.QueryRow("SELECT id FROM clients WHERE name LIKE '%蚂蚁%' LIMIT 1").Scan(&antClientID)
	if err != nil {
		err = db.QueryRow(`
			INSERT INTO clients (name, type, contact_person, phone, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			"蚂蚁科技集团股份有限公司", "COMPANY", "赵法务", "13800138004", "legal@antgroup.com", time.Now(), time.Now(),
		).Scan(&antClientID)
		if err == nil {
			fmt.Printf("✅ 创建蚂蚁金服客户，ID: %d\n", antClientID)
		}
	} else {
		fmt.Printf("✅ 找到蚂蚁金服客户，ID: %d\n", antClientID)
	}

	if alibabaClientID > 0 && antClientID > 0 {
		var alibabaCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", alibabaClientID, lawyerID).Scan(&alibabaCaseCount)

		if alibabaCaseCount == 0 {
			_, err = db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"阿里巴巴电商平台合作协议纠纷", "commercial",
				"阿里巴巴与某电商平台合作协议纠纷案",
				alibabaClientID, lawyerID, "active", time.Now().AddDate(0, -3, 0), time.Now(),
			)
			if err == nil {
				fmt.Println("✅ 创建阿里巴巴案件")
			}
		}

		var antCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", antClientID, lawyerID).Scan(&antCaseCount)

		if antCaseCount == 0 {
			_, err = db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"蚂蚁金服金融牌照申请", "financial",
				"蚂蚁金服申请金融牌照相关法律事务",
				antClientID, lawyerID, "active", time.Now(), time.Now(),
			)
			if err == nil {
				fmt.Println("✅ 创建蚂蚁金服案件（可能触发关联冲突）")
			}
		}
	}

	// 验证创建结果
	fmt.Println("\n📋 验证冲突场景...")

	rows, err := db.Query(`
		SELECT
			c.id, c.title, c.case_type, cl.name as client_name, cl.type as client_type,
			c.created_at, c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		WHERE c.lawyer_id = $1
		ORDER BY c.created_at DESC
		LIMIT 10`, lawyerID)
	if err != nil {
		log.Printf("查询失败: %v", err)
		return
	}
	defer rows.Close()

	var cases []struct {
		ID         int
		Title      string
		CaseType   string
		ClientName string
		ClientType string
		CreatedAt  time.Time
		LawyerID   int
	}

	for rows.Next() {
		var c struct {
			ID         int
			Title      string
			CaseType   string
			ClientName string
			ClientType string
			CreatedAt  time.Time
			LawyerID   int
		}
		err := rows.Scan(&c.ID, &c.Title, &c.CaseType, &c.ClientName, &c.ClientType, &c.CreatedAt, &c.LawyerID)
		if err != nil {
			log.Printf("扫描失败: %v", err)
			continue
		}
		cases = append(cases, c)
	}

	fmt.Printf("\n律师ID %d 代理的案件列表:\n", lawyerID)
	for i, c := range cases {
		fmt.Printf("%d. %s (%s) - 客户: %s (%s)\n", i+1, c.Title, c.CaseType, c.ClientName, c.ClientType)
	}

	if len(cases) >= 2 {
		fmt.Printf("\n🎉 成功创建冲突场景！律师ID %d 代理了 %d 个案件，将触发利益冲突检测\n", lawyerID, len(cases))
		fmt.Println("\n📝 测试步骤:")
		fmt.Println("1. 在前端创建新案件")
		fmt.Println("2. 选择上述律师代理新案件")
		fmt.Println("3. 选择任意现有客户作为新案件的客户")
		fmt.Println("4. 进入利益冲突检查步骤")
		fmt.Println("5. 应该能看到具体的冲突案例详情！")
	} else {
		fmt.Println("❌ 冲突场景创建失败，需要至少2个案件")
	}
}
