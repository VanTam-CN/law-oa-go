package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	// 数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	fmt.Println("=== 创建完整的利益冲突验证数据 ===")

	// 1. 首先检查现有数据
	fmt.Println("\n📊 检查现有数据...")
	var lawyerCount, clientCount, caseCount int

	db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'lawyer' OR role = 'LAWYER'").Scan(&lawyerCount)
	db.QueryRow("SELECT COUNT(*) FROM clients").Scan(&clientCount)
	db.QueryRow("SELECT COUNT(*) FROM cases").Scan(&caseCount)

	fmt.Printf("现有律师数量: %d\n", lawyerCount)
	fmt.Printf("现有客户数量: %d\n", clientCount)
	fmt.Printf("现有案件数量: %d\n", caseCount)

	if lawyerCount == 0 {
		fmt.Println("❌ 需要先创建律师数据")
		createLawyers(db)
	}

	if clientCount < 5 {
		fmt.Println("❌ 需要创建更多客户数据")
		createClients(db)
	}

	// 2. 创建核心冲突场景
	fmt.Println("\n🎯 场景1: 互联网行业竞争冲突验证")
	createInternetCompetitionConflict(db)

	fmt.Println("\n🎯 场景2: 同一客户不同案件冲突")
	createSameClientConflict(db)

	fmt.Println("\n🎯 场景3: 对立当事人冲突")
	createOpposingPartyConflict(db)

	// 3. 验证创建结果
	fmt.Println("\n📋 验证冲突场景...")
	verifyConflictScenarios(db)
}

func createLawyers(db *sql.DB) {
	fmt.Println("👥 创建律师数据...")

	lawyers := []struct {
		Username string
		Name     string
		Email    string
		Phone    string
		Specialty string
	}{
		{"zhangwei", "张伟", "zhangwei@lawfirm.com", "13800138001", "互联网法律"},
		{"lilei", "李雷", "lilei@lawfirm.com", "13800138002", "知识产权"},
		{"wangmei", "王美", "wangmei@lawfirm.com", "13800138003", "商业纠纷"},
		{"chenjun", "陈军", "chenjun@lawfirm.com", "13800138004", "劳动法"},
		{"zhaolin", "赵琳", "zhaolin@lawfirm.com", "13800138005", "公司法律"},
	}

	for _, lawyer := range lawyers {
		var userID uint
		err := db.QueryRow(`
			INSERT INTO users (username, password_hash, name, email, phone, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (username) DO UPDATE SET name = EXCLUDED.name
			RETURNING id`,
			lawyer.Username, "$2a$10$example.hash", lawyer.Name, lawyer.Email,
			lawyer.Phone, "LAWYER", "active", time.Now(), time.Now(),
		).Scan(&userID)

		if err != nil {
			log.Printf("创建律师 %s 失败: %v", lawyer.Name, err)
		} else {
			fmt.Printf("✅ 创建律师: %s (ID: %d)\n", lawyer.Name, userID)
		}
	}
}

func createClients(db *sql.DB) {
	fmt.Println("🏢 创建客户数据...")

	clients := []struct {
		Name        string
		Type        string
		Contact     string
		Phone       string
		Email       string
		Industry    string
	}{
		{"腾讯科技有限公司", "COMPANY", "李经理", "13800138101", "legal@tencent.com", "互联网"},
		{"阿里巴巴集团控股有限公司", "COMPANY", "张法务", "13800138102", "legal@alibaba.com", "互联网"},
		{"字节跳动科技有限公司", "COMPANY", "王总", "13800138103", "legal@bytedance.com", "互联网"},
		{"百度在线网络技术有限公司", "COMPANY", "赵法务", "13800138104", "legal@baidu.com", "互联网"},
		{"美团点评", "COMPANY", "刘经理", "13800138105", "legal@meituan.com", "互联网"},
		{"京东集团", "COMPANY", "陈法务", "13800138106", "legal@jd.com", "互联网"},
		{"网易公司", "COMPANY", "周法务", "13800138107", "legal@netease.com", "互联网"},
		{"小米科技", "COMPANY", "吴经理", "13800138108", "legal@xiaomi.com", "互联网"},
	}

	for _, client := range clients {
		var clientID uint
		err := db.QueryRow(`
			INSERT INTO clients (name, type, contact_person, phone, email, industry, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (name) DO UPDATE SET contact_person = EXCLUDED.contact_person
			RETURNING id`,
			client.Name, client.Type, client.Contact, client.Phone, client.Email,
			client.Industry, time.Now(), time.Now(),
		).Scan(&clientID)

		if err != nil {
			log.Printf("创建客户 %s 失败: %v", client.Name, err)
		} else {
			fmt.Printf("✅ 创建客户: %s (ID: %d)\n", client.Name, clientID)
		}
	}
}

func createInternetCompetitionConflict(db *sql.DB) {
	fmt.Println("🌐 创建互联网行业竞争冲突场景...")

	// 获取张伟律师ID
	var zhangweiID uint
	db.QueryRow("SELECT id FROM users WHERE username = 'zhangwei' LIMIT 1").Scan(&zhangweiID)
	if zhangweiID == 0 {
		fmt.Println("❌ 未找到张伟律师")
		return
	}

	// 获取相关客户ID
	var tencentID, alibabaID, bytedanceID, baiduID uint
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%腾讯%' LIMIT 1").Scan(&tencentID)
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%阿里巴巴%' LIMIT 1").Scan(&alibabaID)
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%字节%' LIMIT 1").Scan(&bytedanceID)
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%百度%' LIMIT 1").Scan(&baiduID)

	// 为张伟律师创建阿里巴巴的案件（已有案件）
	if alibabaID > 0 {
		var alibabaCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", alibabaID, zhangweiID).Scan(&alibabaCaseCount)

		if alibabaCaseCount == 0 {
			_, err := db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"阿里巴巴诉字节跳动不正当竞争纠纷案", "commercial",
				"阿里巴巴公司指控字节跳动旗下产品涉嫌不正当竞争，包括抄袭产品设计、不正当获取商业机密等。案件涉及金额巨大，对互联网行业格局影响深远。",
				alibabaID, zhangweiID, "active", time.Now().AddDate(0, -6, 0), time.Now(),
			)
			if err == nil {
				fmt.Println("✅ 创建阿里巴巴案件（张伟律师代理）")
			}
		} else {
			fmt.Printf("✅ 阿里巴巴已有案件: %d个\n", alibabaCaseCount)
		}

		// 创建阿里巴巴的另一个案件
		var alibabaCase2Count int
		db.QueryRow(`
			SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2
			AND title LIKE '%知识产权%'
		`, alibabaID, zhangweiID).Scan(&alibabaCase2Count)

		if alibabaCase2Count == 0 {
			_, err := db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"阿里巴巴电商平台知识产权保护案", "intellectual_property",
				"阿里巴巴集团旗下电商平台知识产权保护系列案件，涉及商标侵权、专利纠纷等。",
				alibabaID, zhangweiID, "active", time.Now().AddDate(0, -3, 0), time.Now(),
			)
			if err == nil {
				fmt.Println("✅ 创建阿里巴巴知识产权案件（张伟律师代理）")
			}
		}
	}

	// 为张伟律师创建百度的案件（潜在冲突）
	if baiduID > 0 {
		var baiduCaseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", baiduID, zhangweiID).Scan(&baiduCaseCount)

		if baiduCaseCount == 0 {
			_, err := db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				"百度搜索引擎算法案", "technology",
				"百度公司搜索引擎算法相关的法律纠纷，涉及技术专利和算法保护。",
				baiduID, zhangweiID, "active", time.Now().AddDate(0, -2, 0), time.Now(),
			)
			if err == nil {
				fmt.Println("✅ 创建百度案件（张伟律师代理）")
			}
		}
	}
}

func createSameClientConflict(db *sql.DB) {
	fmt.Println("🔄 创建同一客户不同案件冲突场景...")

	// 获取李雷律师ID
	var lileiID uint
	db.QueryRow("SELECT id FROM users WHERE username = 'lilei' LIMIT 1").Scan(&lileiID)
	if lileiID == 0 {
		fmt.Println("❌ 未找到李雷律师")
		return
	}

	// 获取腾讯客户ID
	var tencentID uint
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%腾讯%' LIMIT 1").Scan(&tencentID)
	if tencentID == 0 {
		fmt.Println("❌ 未找到腾讯客户")
		return
	}

	// 为腾讯创建多个案件
	cases := []struct {
		Title string
		Type  string
		Desc  string
	}{
		{
			"腾讯游戏著作权纠纷案",
			"intellectual_property",
			"腾讯公司游戏产品著作权保护纠纷，涉及游戏玩法、界面设计等知识产权问题。",
		},
		{
			"腾讯微信支付合规案",
			"financial",
			"腾讯微信支付相关合规审查和法律事务处理。",
		},
		{
			"腾讯云服务合同纠纷",
			"commercial",
			"腾讯云服务与企业客户之间的合同纠纷处理。",
		},
	}

	for i, caseData := range cases {
		var caseCount int
		db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2 AND title = $3",
			tencentID, lileiID, caseData.Title).Scan(&caseCount)

		if caseCount == 0 {
			_, err := db.Exec(`
				INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				caseData.Title, caseData.Type, caseData.Desc,
				tencentID, lileiID, "active",
				time.Now().AddDate(0, -int(i+1), 0), time.Now(),
			)
			if err == nil {
				fmt.Printf("✅ 创建腾讯案件 %d: %s\n", i+1, caseData.Title)
			}
		}
	}
}

func createOpposingPartyConflict(db *sql.DB) {
	fmt.Println("⚔️ 创建对立当事人冲突场景...")

	// 获取王美律师ID
	var wangmeiID uint
	db.QueryRow("SELECT id FROM users WHERE username = 'wangmei' LIMIT 1").Scan(&wangmeiID)
	if wangmeiID == 0 {
		fmt.Println("❌ 未找到王美律师")
		return
	}

	// 获取字节跳动和腾讯客户ID
	var bytedanceID, tencentID uint
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%字节%' LIMIT 1").Scan(&bytedanceID)
	db.QueryRow("SELECT id FROM clients WHERE name LIKE '%腾讯%' LIMIT 1").Scan(&tencentID)

	if bytedanceID == 0 || tencentID == 0 {
		fmt.Println("❌ 未找到字节跳动或腾讯客户")
		return
	}

	// 王美律师代理字节跳动起诉腾讯的案件
	var byteVsTencentCount int
	db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2", bytedanceID, wangmeiID).Scan(&byteVsTencentCount)

	if byteVsTencentCount == 0 {
		_, err := db.Exec(`
			INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			"字节跳动诉腾讯滥用市场支配地位案", "antitrust",
			"字节跳动公司起诉腾讯公司滥用市场支配地位，垄断互联网社交和内容分发市场。案件涉及反垄断法、不正当竞争等复杂法律问题，索赔金额高达数十亿元。",
			bytedanceID, wangmeiID, "active", time.Now().AddDate(0, -1, 0), time.Now(),
		)
		if err == nil {
			fmt.Println("✅ 创建字节跳动诉腾讯案件（王美律师代理）")
		}
	}

	// 王美律师也代理腾讯的其他案件（冲突！）
	var tencentOtherCaseCount int
	db.QueryRow("SELECT COUNT(*) FROM cases WHERE client_id = $1 AND lawyer_id = $2 AND title NOT LIKE '%字节%'",
		tencentID, wangmeiID).Scan(&tencentOtherCaseCount)

	if tencentOtherCaseCount == 0 {
		_, err := db.Exec(`
			INSERT INTO cases (title, case_type, description, client_id, lawyer_id, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			"腾讯投资并购合规审查", "investment",
			"腾讯公司投资并购相关的法律合规审查和风险评估工作。",
			tencentID, wangmeiID, "active", time.Now().AddDate(0, -4, 0), time.Now(),
		)
		if err == nil {
			fmt.Println("✅ 创建腾讯投资案件（王美律师代理 - 冲突！)")
		}
	}
}

func verifyConflictScenarios(db *sql.DB) {
	fmt.Println("\n🔍 验证冲突场景数据...")

	// 查询张伟律师的案件（互联网竞争冲突）
	var zhangweiCases []struct {
		ID         int
		Title      string
		ClientName string
		CaseType   string
		CreatedAt  time.Time
	}

	rows, err := db.Query(`
		SELECT c.id, c.title, cl.name as client_name, c.case_type, c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE u.username = 'zhangwei'
		ORDER BY c.created_at DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c struct {
				ID         int
				Title      string
				ClientName string
				CaseType   string
				CreatedAt  time.Time
			}
			rows.Scan(&c.ID, &c.Title, &c.ClientName, &c.CaseType, &c.CreatedAt)
			zhangweiCases = append(zhangweiCases, c)
		}
	}

	fmt.Printf("\n📋 张伟律师代理的案件（潜在冲突）:\n")
	for i, c := range zhangweiCases {
		fmt.Printf("%d. %s - 客户: %s (%s)\n", i+1, c.Title, c.ClientName, c.CaseType)
	}

	// 查询王美律师的案件（对立当事人冲突）
	var wangmeiCases []struct {
		ID         int
		Title      string
		ClientName string
		CaseType   string
		CreatedAt  time.Time
	}

	rows, err = db.Query(`
		SELECT c.id, c.title, cl.name as client_name, c.case_type, c.created_at
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE u.username = 'wangmei'
		ORDER BY c.created_at DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var c struct {
				ID         int
				Title      string
				ClientName string
				CaseType   string
				CreatedAt  time.Time
			}
			rows.Scan(&c.ID, &c.Title, &c.ClientName, &c.CaseType, &c.CreatedAt)
			wangmeiCases = append(wangmeiCases, c)
		}
	}

	fmt.Printf("\n⚔️ 王美律师代理的案件（对立当事人冲突）:\n")
	for i, c := range wangmeiCases {
		fmt.Printf("%d. %s - 客户: %s (%s)\n", i+1, c.Title, c.ClientName, c.CaseType)
	}

	// 输出测试指导
	fmt.Println("\n🧪 冲突检测测试指导:")
	fmt.Println(strings.Repeat("=", 50))

	if len(zhangweiCases) >= 2 {
		fmt.Println("✅ 场景1: 互联网行业竞争冲突")
		fmt.Println("   - 账号: zhangwei / law123456")
		fmt.Println("   - 操作: 尝试为美团或京东创建新案件")
		fmt.Println("   - 预期: 检测到与阿里巴巴、百度的行业竞争冲突")
	}

	if len(wangmeiCases) >= 2 {
		fmt.Println("✅ 场景2: 对立当事人冲突")
		fmt.Println("   - 账号: wangmei / law123456")
		fmt.Println("   - 操作: 尝试为字节跳动或腾讯创建新案件")
		fmt.Println("   - 预期: 检测到代理对立当事人的严重冲突")
	}

	fmt.Println("\n📝 具体测试步骤:")
	fmt.Println("1. 使用指定律师账号登录系统")
	fmt.Println("2. 进入'新建案件'流程")
	fmt.Println("3. 填写案件基本信息")
	fmt.Println("4. 选择相关律师和客户")
	fmt.Println("5. 在第3步（利益冲突检查）中填写对方当事人信息")
	fmt.Println("6. 点击'下一步'触发冲突检查")
	fmt.Println("7. 查看详细的冲突检查结果")

	fmt.Println("\n🎯 预期冲突类型:")
	fmt.Println("- 代理冲突: 同一律师代理对立客户的案件")
	fmt.Println("- 商业竞争冲突: 同一律师代理同行业竞争客户的案件")
	fmt.Println("- 当事人冲突: 新案件的对方当事人与历史案件相关联")
}