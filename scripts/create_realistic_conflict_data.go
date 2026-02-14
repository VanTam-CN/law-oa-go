//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("🏛️ 创建真实感利益冲突测试数据")
	fmt.Println("==============================")

	// 连接数据库
	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 创建真实的律师数据
	fmt.Println("\n👨‍⚖️ 创建真实律师数据...")
	createRealisticLawyers(db)

	// 创建多样化的客户数据
	fmt.Println("\n👥 创建多样化客户数据...")
	createRealisticClients(db)

	// 创建复杂的利益冲突案件
	fmt.Println("\n⚖️ 创建复杂利益冲突案件...")
	createConflictCases(db)

	fmt.Println("\n✅ 真实感利益冲突测试数据创建完成！")
	fmt.Println("=====================================")
}

func createRealisticLawyers(db *sql.DB) {
	// 真实感的律师数据
	lawyers := []struct {
		username  string
		email     string
		password  string
		fullName  string
		title     string
		department string
		phone     string
		status    string
		createdAt time.Time
	}{
		{
			username:  "zhangwei",
			email:     "zhangwei@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "张伟",
			title:     "高级合伙人",
			department: "公司业务部",
			phone:     "13800138001",
			status:    "active",
			createdAt: time.Now().AddDate(-3, 0, 0),
		},
		{
			username:  "liming",
			email:     "liming@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "李明",
			title:     "合伙人",
			department: "知识产权部",
			phone:     "13800138002",
			status:    "active",
			createdAt: time.Now().AddDate(-2, 0, 0),
		},
		{
			username:  "wangfang",
			email:     "wangfang@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "王芳",
			title:     "高级律师",
			department: "诉讼仲裁部",
			phone:     "13800138003",
			status:    "active",
			createdAt: time.Now().AddDate(-1, -6, 0),
		},
		{
			username:  "chenhao",
			email:     "chenhao@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "陈浩",
			title:     "律师",
			department: "公司业务部",
			phone:     "13800138004",
			status:    "active",
			createdAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			username:  "zhaojing",
			email:     "zhaojing@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "赵静",
			title:     "律师",
			department: "诉讼仲裁部",
			phone:     "13800138005",
			status:    "active",
			createdAt: time.Now().AddDate(-1, -3, 0),
		},
		{
			username:  "sunlei",
			email:     "sunlei@jinchenglaw.com",
			password:  hashPassword("law123456"),
			fullName:  "孙雷",
			title:     "律师助理",
			department: "知识产权部",
			phone:     "13800138006",
			status:    "active",
			createdAt: time.Now().AddDate(-0, -6, 0),
		},
	}

	for _, lawyer := range lawyers {
		// 检查律师是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", lawyer.username).Scan(&count)
		if err == nil && count > 0 {
			fmt.Printf("📋 律师 %s 已存在，跳过\n", lawyer.fullName)
			continue
		}

		// 插入律师数据
		insertQuery := `
			INSERT INTO users (username, email, password, name, phone, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'lawyer', $6, $7, $7)
			RETURNING id
		`

		var lawyerID int
		err = db.QueryRow(insertQuery,
			lawyer.username, lawyer.email, lawyer.password, lawyer.fullName,
			lawyer.phone, lawyer.status, lawyer.createdAt,
		).Scan(&lawyerID)

		if err != nil {
			log.Printf("创建律师 %s 失败: %v", lawyer.fullName, err)
			continue
		}

		fmt.Printf("✅ 创建律师: %s (%s) - ID: %d\n", lawyer.fullName, lawyer.title, lawyerID)
	}
}

func createRealisticClients(db *sql.DB) {
	// 真实感的客户数据，包含可能产生利益冲突的关系
	clients := []struct {
		name    string
		typeStr string
		contact string
		email   string
		phone   string
		address string
		status  string
		company string
		idCard  string
		industry string
		notes   string
		createdAt time.Time
		lawyerID int
	}{
		// 互联网科技公司 - 潜在竞争关系
		{
			name:      "阿里巴巴集团控股有限公司",
			typeStr:   "企业",
			contact:   "张三丰",
			email:     "legal@alibaba.com",
			phone:     "0571-88888888",
			address:   "杭州市余杭区文一西路969号",
			status:    "active",
			company:   "阿里巴巴集团控股有限公司",
			industry:  "电子商务",
			notes:     "电子商务巨头，涉及多起知识产权纠纷",
			createdAt: time.Now().AddDate(-2, 0, 0),
			lawyerID:  7, // 张伟
		},
		{
			name:      "腾讯控股有限公司",
			typeStr:   "企业",
			contact:   "李四平",
			email:     "legal@tencent.com",
			phone:     "0755-86013388",
			address:   "深圳市南山区科技园科技中一路腾讯大厦",
			status:    "active",
			company:   "腾讯控股有限公司",
			industry:  "互联网服务",
			notes:     "互联网巨头，与阿里巴巴存在竞争关系",
			createdAt: time.Now().AddDate(-1, -8, 0),
			lawyerID:  8, // 李明
		},
		{
			name:      "字节跳动科技有限公司",
			typeStr:   "企业",
			contact:   "王五军",
			email:     "legal@bytedance.com",
			phone:     "010-84358888",
			address:   "北京市海淀区北三环西路甲18号",
			status:    "active",
			company:   "字节跳动科技有限公司",
			industry:  "互联网技术",
			notes:     "新兴互联网公司，与腾讯、阿里存在业务竞争",
			createdAt: time.Now().AddDate(-0, -10, 0),
			lawyerID:  7, // 张伟 (潜在冲突)
		},

		// 建筑公司 - 合作与竞争关系
		{
			name:      "中国建筑集团有限公司",
			typeStr:   "企业",
			contact:   "赵六工",
			email:     "legal@cscec.com",
			phone:     "010-88080000",
			address:   "北京市海淀区三里河路15号",
			status:    "active",
			company:   "中国建筑集团有限公司",
			industry:  "建筑工程",
			notes:     "大型国有企业，承接重大基础设施项目",
			createdAt: time.Now().AddDate(-3, -2, 0),
			lawyerID:  9, // 王芳
		},
		{
			name:      "中国中铁股份有限公司",
			typeStr:   "企业",
			contact:   "钱七程",
			email:     "legal@crecg.com",
			phone:     "010-51870000",
			address:   "北京市海淀区复兴路69号",
			status:    "active",
			company:   "中国中铁股份有限公司",
			industry:  "铁路建设",
			notes:     "铁路建设巨头，与中国建筑存在项目竞争",
			createdAt: time.Now().AddDate(-1, -5, 0),
			lawyerID:  9, // 王芳 (潜在冲突)
		},

		// 个人客户 - 家庭纠纷
		{
			name:      "刘德华",
			typeStr:   "个人",
			contact:   "刘德华",
			email:     "liudehua@email.com",
			phone:     "13800138888",
			address:   "北京市朝阳区三里屯",
			status:    "active",
			company:   "",
			industry:  "",
			idCard:    "110101198001010001",
			notes:     "明星客户，离婚纠纷案",
			createdAt: time.Now().AddDate(-0, -8, 0),
			lawyerID:  10, // 陈浩
		},
		{
			name:      "朱丽倩",
			typeStr:   "个人",
			contact:   "朱丽倩",
			email:     "zhuliqian@email.com",
			phone:     "13800138889",
			address:   "香港特别行政区",
			status:    "active",
			company:   "",
			industry:  "",
			idCard:    "810000198002020002",
			notes:     "刘德华的妻子，离婚纠纷案另一方",
			createdAt: time.Now().AddDate(-0, -8, 0),
			lawyerID:  11, // 赵静 (直接冲突)
		},

		// 股权纠纷相关客户
		{
			name:      "万科企业股份有限公司",
			typeStr:   "企业",
			contact:   "王石",
			email:     "legal@vanke.com",
			phone:     "0755-25007777",
			address:   "深圳市盐田区大梅沙环梅路33号",
			status:    "active",
			company:   "万科企业股份有限公司",
			industry:  "房地产开发",
			notes:     "房地产龙头企业，涉及多起股权纠纷",
			createdAt: time.Now().AddDate(-0, -6, 0),
			lawyerID:  7, // 张伟
		},
		{
			name:      "宝能集团股份有限公司",
			typeStr:   "企业",
			contact:   "姚振华",
			email:     "legal@baoneng.com",
			phone:     "0755-82990000",
			address:   "深圳市福田区深南大道7028号时代科技大厦",
			status:    "active",
			company:   "宝能集团股份有限公司",
			industry:  "综合投资",
			notes:     "投资集团，与万科存在股权纠纷",
			createdAt: time.Now().AddDate(-0, -6, 0),
			lawyerID:  8, // 李明 (与张伟的客户存在冲突)
		},

		// 医疗行业客户
		{
			name:      "北京协和医院",
			typeStr:   "企业",
			contact:   "张院长",
			email:     "legal@pumch.cn",
			phone:     "010-69156114",
			address:   "北京市东城区东单帅府园1号",
			status:    "active",
			company:   "北京协和医院",
			industry:  "医疗服务",
			notes:     "顶级医院，医疗纠纷案例",
			createdAt: time.Now().AddDate(-0, -4, 0),
			lawyerID:  10, // 陈浩
		},
		{
			name:      "王先生",
			typeStr:   "个人",
			contact:   "王建国",
			email:     "wangjianguo@email.com",
			phone:     "13800139999",
			address:   "北京市西城区金融街",
			status:    "active",
			company:   "",
			industry:  "",
			idCard:    "110101197503030003",
			notes:     "患者，与北京协和医院存在医疗纠纷",
			createdAt: time.Now().AddDate(-0, -4, 0),
			lawyerID:  11, // 赵静 (与陈浩的客户存在冲突)
		},
	}

	for _, client := range clients {
		// 检查客户是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM clients WHERE name = $1", client.name).Scan(&count)
		if err == nil && count > 0 {
			fmt.Printf("📋 客户 %s 已存在，跳过\n", client.name)
			continue
		}

		// 插入客户数据
		insertQuery := `
			INSERT INTO clients (name, type, contact_person, email, phone, address, status, company, id_card, industry, notes, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
			RETURNING id
		`

		var clientID int
		err = db.QueryRow(insertQuery,
			client.name, client.typeStr, client.contact, client.email, client.phone,
			client.address, client.status, client.company, client.idCard,
			client.industry, client.notes, client.createdAt,
		).Scan(&clientID)

		if err != nil {
			log.Printf("创建客户 %s 失败: %v", client.name, err)
			continue
		}

		fmt.Printf("✅ 创建客户: %s (%s) - ID: %d\n", client.name, client.typeStr, clientID)
	}
}

func createConflictCases(db *sql.DB) {
	// 复杂的利益冲突案件数据
	cases := []struct {
		title       string
		clientID    int
		lawyerID    int
		caseType    string
		status      string
		description string
		value       float64
		createdAt   time.Time
		updatedAt   time.Time
		conflictType string
		riskLevel   string
	}{
		// 互联网竞争案件
		{
			title:       "阿里巴巴诉字节跳动不正当竞争纠纷案",
			clientID:    12, // 阿里巴巴
			lawyerID:    7,  // 张伟
			caseType:    "知识产权纠纷",
			status:      "active",
			description: "指控字节跳动抄袭淘宝商品信息，构成不正当竞争",
			value:       5000000.00,
			createdAt:   time.Now().AddDate(-0, -2, 0),
			updatedAt:   time.Now().AddDate(-0, -1, 0),
			conflictType: "商业竞争",
			riskLevel:   "HIGH",
		},
		{
			title:       "字节跳动诉腾讯垄断纠纷案",
			clientID:    14, // 字节跳动
			lawyerID:    7,  // 张伟 (冲突！)
			caseType:    "反垄断纠纷",
			status:      "active",
			description: "指控腾讯滥用市场支配地位，阻止抖音在微信平台传播",
			value:       8000000.00,
			createdAt:   time.Now().AddDate(-0, -1, 0),
			updatedAt:   time.Now().AddDate(-0, -0, -15),
			conflictType: "商业竞争",
			riskLevel:   "CRITICAL", // 张伟同时代理阿里巴巴和字节跳动
		},

		// 离婚纠纷案件 - 直接冲突
		{
			title:       "刘德华诉朱丽倩离婚纠纷案",
			clientID:    17, // 刘德华
			lawyerID:    10, // 陈浩
			caseType:    "婚姻家庭纠纷",
			status:      "active",
			description: "涉及财产分割、子女抚养权等问题",
			value:       100000000.00,
			createdAt:   time.Now().AddDate(-0, -7, 0),
			updatedAt:   time.Now().AddDate(-0, -6, 0),
			conflictType: "法律对立",
			riskLevel:   "CRITICAL",
		},
		{
			title:       "朱丽倩诉刘德华离婚反诉案",
			clientID:    18, // 朱丽倩
			lawyerID:    11, // 赵静 (冲突！)
			caseType:    "婚姻家庭纠纷",
			status:      "active",
			description: "反诉要求获得更多财产份额和子女抚养权",
			value:       100000000.00,
			createdAt:   time.Now().AddDate(-0, -7, 0),
			updatedAt:   time.Now().AddDate(-0, -6, 0),
			conflictType: "法律对立",
			riskLevel:   "CRITICAL", // 同一律师事务所代理离婚双方
		},

		// 建筑工程竞争案件
		{
			title:       "中国建筑与中国中铁项目竞标纠纷案",
			clientID:    15, // 中国建筑
			lawyerID:    9,  // 王芳
			caseType:    "建设工程纠纷",
			status:      "active",
			description: "就某高铁项目竞标过程中是否存在违规行为产生纠纷",
			value:       30000000.00,
			createdAt:   time.Now().AddDate(-0, -4, 0),
			updatedAt:   time.Now().AddDate(-0, -3, 0),
			conflictType: "商业竞争",
			riskLevel:   "HIGH",
		},

		// 万科宝能股权大战
		{
			title:       "万科诉宝能恶意收购纠纷案",
			clientID:    19, // 万科
			lawyerID:    7,  // 张伟
			caseType:    "公司并购纠纷",
			status:      "active",
			description: "指控宝能恶意收购，损害公司和股东利益",
			value:       500000000.00,
			createdAt:   time.Now().AddDate(-0, -5, 0),
			updatedAt:   time.Now().AddDate(-0, -4, 0),
			conflictType: "股权纠纷",
			riskLevel:   "HIGH",
		},
		{
			title:       "宝能反诉万科损害股东利益案",
			clientID:    20, // 宝能
			lawyerID:    8,  // 李明 (潜在冲突)
			caseType:    "股东权益纠纷",
			status:      "active",
			description: "反诉万科管理层损害股东利益，要求赔偿",
			value:       200000000.00,
			createdAt:   time.Now().AddDate(-0, -5, 0),
			updatedAt:   time.Now().AddDate(-0, -4, 0),
			conflictType: "股权纠纷",
			riskLevel:   "HIGH", // 同一所律师事务所代理对立双方
		},

		// 医疗纠纷案件
		{
			title:       "王先生诉北京协和医院医疗事故纠纷案",
			clientID:    22, // 王先生
			lawyerID:    11, // 赵静
			caseType:    "医疗纠纷",
			status:      "active",
			description: "患者认为医院存在医疗过失，造成身体损害",
			value:       500000.00,
			createdAt:   time.Now().AddDate(-0, -3, 0),
			updatedAt:   time.Now().AddDate(-0, -2, 0),
			conflictType: "服务纠纷",
			riskLevel:   "MEDIUM",
		},

		// 知识产权案件
		{
			title:       "腾讯诉抖音短视频版权侵权案",
			clientID:    13, // 腾讯
			lawyerID:    8,  // 李明
			caseType:    "知识产权纠纷",
			status:      "active",
			description: "指控抖音平台存在大量腾讯独家版权视频内容",
			value:       10000000.00,
			createdAt:   time.Now().AddDate(-0, -1, 0),
			updatedAt:   time.Now().AddDate(-0, -0, -20),
			conflictType: "知识产权冲突",
			riskLevel:   "MEDIUM",
		},

		// 劳动纠纷案件
		{
			title:       "李工程师诉字节跳动劳动合同纠纷案",
			clientID:    23, // 额外创建的工程师客户
			lawyerID:    6,  // 孙雷
			caseType:    "劳动纠纷",
			status:      "active",
			description: "前员工起诉公司违法解除劳动合同",
			value:       500000.00,
			createdAt:   time.Now().AddDate(-0, -0, -10),
			updatedAt:   time.Now().AddDate(-0, -0, -5),
			conflictType: "劳动关系",
			riskLevel:   "LOW",
		},
	}

	for _, caseData := range cases {
		// 检查案件是否已存在
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM cases WHERE title = $1", caseData.title).Scan(&count)
		if err == nil && count > 0 {
			fmt.Printf("📋 案件 %s 已存在，跳过\n", caseData.title)
			continue
		}

		// 插入案件数据
		insertQuery := `
			INSERT INTO cases (title, client_id, lawyer_id, case_type, status, description, start_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			RETURNING id
		`

		var caseID int
		err = db.QueryRow(insertQuery,
			caseData.title, caseData.clientID, caseData.lawyerID,
			caseData.caseType, caseData.status, caseData.description,
			caseData.createdAt, caseData.createdAt,
		).Scan(&caseID)

		if err != nil {
			log.Printf("创建案件 %s 失败: %v", caseData.title, err)
			continue
		}

		riskIndicator := ""
		if caseData.riskLevel == "CRITICAL" {
			riskIndicator = " 🚨 极高风险"
		} else if caseData.riskLevel == "HIGH" {
			riskIndicator = " ⚠️ 高风险"
		}

		fmt.Printf("✅ 创建案件: %s (%s) - ID: %d%s\n", caseData.title, caseData.caseType, caseID, riskIndicator)
	}
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}