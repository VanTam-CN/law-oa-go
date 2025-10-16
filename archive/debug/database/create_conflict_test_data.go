package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// 主要测试场景设计：
// 1. 同一律师代理有利益冲突的多个客户
// 2. 同一客户在不同案件中处于对立位置
// 3. 公司客户与其关联公司之间的利益冲突
// 4. 律师与之前的对立当事人产生新的代理关系

func main() {
	// 数据库连接
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库ping失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("=== 开始创建利益冲突测试数据 ===")

	// 创建测试律师
	createConflictTestLawyers(db)

	// 创建测试客户
	createConflictTestClients(db)

	// 创建测试案件
	createConflictTestCases(db)

	fmt.Println("🎉 利益冲突测试数据创建完成！")
	fmt.Println("\n=== 测试场景说明 ===")
	fmt.Println("1. 直接利益冲突：律师张三同时代理了客户A和客户B，但两客户在某个案件中处于对立位置")
	fmt.Println("2. 间接利益冲突：律师李四的客户C与客户D是竞争对手关系")
	fmt.Println("3. 公司关联冲突：律师王五代理的公司E与之前代理的公司F是母子公司关系")
	fmt.Println("4. 时间重叠冲突：律师赵六在相近时间内代理了有利益冲突的案件")
}

// 创建利益冲突测试律师
func createConflictTestLawyers(db *sql.DB) {
	fmt.Println("\n--- 创建测试律师数据 ---")

	lawyers := []struct {
		name        string
		email       string
		phone       string
		specialty   string
		description string
	}{
		{"张三", "zhangsan@lawfirm.com", "13800138001", "民事诉讼", "专门处理合同纠纷案件"},
		{"李四", "lisi@lawfirm.com", "13800138002", "商业诉讼", "擅长处理公司间商业纠纷"},
		{"王五", "wangwu@lawfirm.com", "13800138003", "公司法务", "专注企业法律顾问服务"},
		{"赵六", "zhaoliu@lawfirm.com", "13800138004", "知识产权", "专业处理专利和商标案件"},
		{"孙七", "sunqi@lawfirm.com", "13800138005", "刑事辩护", "专门处理刑事案件"},
	}

	for _, lawyer := range lawyers {
		// 检查律师是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND role = 'lawyer'", lawyer.email).Scan(&count)
		if err != nil {
			log.Printf("查询律师失败 %s: %v", lawyer.name, err)
			continue
		}

		if count == 0 {
			// 创建新律师
			username := lawyer.email[:len(lawyer.email)-11] // 移除 @lawfirm.com
			_, err = db.Exec(`
				INSERT INTO users (name, email, phone, username, password, role, status, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, 'lawyer', 'active', NOW(), NOW())
			`, lawyer.name, lawyer.email, lawyer.phone, username, "default123")

			if err != nil {
				log.Printf("创建律师失败 %s: %v", lawyer.name, err)
			} else {
				fmt.Printf("✅ 创建律师: %s (%s)\n", lawyer.name, lawyer.specialty)
			}
		} else {
			fmt.Printf("ℹ️  律师已存在: %s\n", lawyer.name)
		}
	}
}

// 创建利益冲突测试客户
func createConflictTestClients(db *sql.DB) {
	fmt.Println("\n--- 创建测试客户数据 ---")

	clients := []struct {
		name         string
		clientType   string
		contact      string
		description  string
		conflictInfo string
	}{
		// 场景1：直接利益冲突的客户
		{"阿里巴巴集团", "COMPANY", "0571-85012345", "大型互联网公司，电子商务业务", "场景1：与腾讯在电商领域竞争"},
		{"腾讯控股", "COMPANY", "0755-86013388", "大型互联网公司，社交和游戏业务", "场景1：与阿里巴巴在多个领域竞争"},

		// 场景2：间接利益冲突的客户
		{"华为技术有限公司", "COMPANY", "0755-28780808", "通信设备和智能终端制造商", "场景2：与小米在手机市场竞争"},
		{"小米科技", "COMPANY", "010-56408888", "智能手机和IoT设备制造商", "场景2：与华为在手机市场竞争"},

		// 场景3：公司关联关系冲突
		{"字节跳动", "COMPANY", "010-84388866", "互联网内容平台", "场景3：母公司"},
		{"抖音公司", "COMPANY", "010-12345678", "短视频平台", "场景3：字节跳动子公司"},

		// 场景4：个人客户冲突
		{"张伟", "PERSON", "13901390001", "企业家，投资人", "场景4：与李明有商业纠纷"},
		{"李明", "PERSON", "13901390002", "企业家，创业者", "场景4：与张伟有商业纠纷"},

		// 测试用客户
		{"微软中国", "COMPANY", "010-58568888", "软件和云服务提供商", "测试客户，无明显冲突"},
		{"谷歌中国", "COMPANY", "010-82538888", "搜索引擎和广告服务", "测试客户，无明显冲突"},
	}

	for _, client := range clients {
		// 检查客户是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM clients WHERE name = ?", client.name).Scan(&count)
		if err != nil {
			log.Printf("查询客户失败 %s: %v", client.name, err)
			continue
		}

		if count == 0 {
			// 创建新客户
			_, err = db.Exec(`
				INSERT INTO clients (name, type, phone, email, industry, notes, status, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, 'active', NOW(), NOW())
			`, client.name, client.clientType, client.contact, client.contact+"@example.com", "其他", client.description)

			if err != nil {
				log.Printf("创建客户失败 %s: %v", client.name, err)
			} else {
				fmt.Printf("✅ 创建客户: %s (%s) - %s\n", client.name, client.clientType, client.conflictInfo)
			}
		} else {
			fmt.Printf("ℹ️  客户已存在: %s\n", client.name)
		}
	}
}

// 创建利益冲突测试案件
func createConflictTestCases(db *sql.DB) {
	fmt.Println("\n--- 创建测试案件数据 ---")

	// 首先获取律师和客户ID
	lawyers := getLawyerIDs(db)
	clients := getClientIDs(db)

	if len(lawyers) == 0 || len(clients) == 0 {
		log.Fatal("没有找到足够的律师或客户数据")
	}

	cases := []struct {
		caseName      string
		caseType      string
		clientID      string
		lawyerID      string
		description   string
		opposingParty string
		status        string
		conflictType  string
		testScenario  string
	}{
		// 场景1：直接利益冲突 - 同一律师代理竞争双方
		{
			"阿里巴巴诉腾讯不正当竞争案",
			"民事诉讼",
			clients["阿里巴巴集团"],
			lawyers["张三"],
			"阿里巴巴指控腾讯在电商直播领域存在不正当竞争行为",
			"腾讯控股",
			"进行中",
			"direct_conflict",
			"张三律师同时代理了竞争双方",
		},
		{
			"腾讯诉阿里巴巴垄断案",
			"民事诉讼",
			clients["腾讯控股"],
			lawyers["张三"],
			"腾讯指控阿里巴巴滥用市场支配地位",
			"阿里巴巴集团",
			"进行中",
			"direct_conflict",
			"与上一案件形成直接利益冲突",
		},

		// 场景2：间接利益冲突 - 竞争对手关系
		{
			"华为诉小米专利侵权案",
			"知识产权",
			clients["华为技术有限公司"],
			lawyers["李四"],
			"华为指控小米侵犯其通信技术专利",
			"小米科技",
			"进行中",
			"indirect_conflict",
			"同一律师代理竞争对手",
		},
		{
			"小米诉华为商业诋毁案",
			"民事诉讼",
			clients["小米科技"],
			lawyers["李四"],
			"小米指控华为进行商业诋毁",
			"华为技术有限公司",
			"进行中",
			"indirect_conflict",
			"与上一案件形成间接利益冲突",
		},

		// 场景3：公司关联关系冲突
		{
			"字节跳动投资纠纷案",
			"商业诉讼",
			clients["字节跳动"],
			lawyers["王五"],
			"字节跳动与其他公司的投资纠纷",
			"某科技公司",
			"进行中",
			"corporate_relation_conflict",
			"母公司案件",
		},
		{
			"抖音公司股权争议案",
			"公司诉讼",
			clients["抖音公司"],
			lawyers["王五"],
			"抖音公司的股权结构争议",
			"某投资公司",
			"进行中",
			"corporate_relation_conflict",
			"子公司案件，与母公司可能存在关联冲突",
		},

		// 场景4：个人客户冲突
		{
			"张伟诉李明合同纠纷案",
			"合同纠纷",
			clients["张伟"],
			lawyers["赵六"],
			"两人之间的商业合同纠纷",
			"李明",
			"进行中",
			"personal_conflict",
			"个人客户间的直接冲突",
		},
		{
			"李明诉张伟违约案",
			"合同纠纷",
			clients["李明"],
			lawyers["赵六"],
			"另一项违约纠纷",
			"张伟",
			"进行中",
			"personal_conflict",
			"同一律师代理对立的双方",
		},

		// 无冲突的测试案件
		{
			"微软中国软件许可案",
			"知识产权",
			clients["微软中国"],
			lawyers["孙七"],
			"微软中国的软件许可业务",
			"某软件公司",
			"进行中",
			"no_conflict",
			"无明显利益冲突",
		},
		{
			"谷歌中国广告纠纷案",
			"广告纠纷",
			clients["谷歌中国"],
			lawyers["孙七"],
			"谷歌中国的广告业务纠纷",
			"某广告公司",
			"进行中",
			"no_conflict",
			"无明显利益冲突",
		},
	}

	for _, caseData := range cases {
		if caseData.clientID == "" || caseData.lawyerID == "" {
			fmt.Printf("⚠️  跳过案件（缺少客户或律师）: %s\n", caseData.caseName)
			continue
		}

		// 检查案件是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM cases WHERE title = ?", caseData.caseName).Scan(&count)
		if err != nil {
			log.Printf("查询案件失败 %s: %v", caseData.caseName, err)
			continue
		}

		if count == 0 {
			// 将对立方信息包含在描述中
			fullDescription := fmt.Sprintf("%s\n\n对立方: %s\n\n冲突类型: %s\n测试场景: %s",
				caseData.description, caseData.opposingParty, caseData.conflictType, caseData.testScenario)

			// 创建新案件
			_, err = db.Exec(`
				INSERT INTO cases (title, case_type, client_id, lawyer_id, description, status, priority, created_at, updated_at)
				VALUES (?, 'civil', ?, ?, ?, 'pending', 'high', NOW(), NOW())
			`, caseData.caseName, caseData.clientID, caseData.lawyerID, fullDescription)

			if err != nil {
				log.Printf("创建案件失败 %s: %v", caseData.caseName, err)
			} else {
				fmt.Printf("✅ 创建案件: %s - %s\n", caseData.caseName, caseData.testScenario)
			}
		} else {
			fmt.Printf("ℹ️  案件已存在: %s\n", caseData.caseName)
		}
	}
}

// 获取律师ID映射
func getLawyerIDs(db *sql.DB) map[string]string {
	lawyers := make(map[string]string)

	rows, err := db.Query("SELECT id, name FROM users WHERE role = 'lawyer'")
	if err != nil {
		log.Printf("查询律师失败: %v", err)
		return lawyers
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		lawyers[name] = id
	}

	return lawyers
}

// 获取客户ID映射
func getClientIDs(db *sql.DB) map[string]string {
	clients := make(map[string]string)

	rows, err := db.Query("SELECT id, name FROM clients")
	if err != nil {
		log.Printf("查询客户失败: %v", err)
		return clients
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			continue
		}
		clients[name] = id
	}

	return clients
}