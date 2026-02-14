//go:build ignore

package main

import (
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 跳过自动迁移，直接使用现有表结构

	fmt.Println("🎯 开始创建基础测试数据")
	fmt.Println("=========================")

	
	// 1. 清理现有测试数据
	fmt.Println("🧹 清理现有测试数据...")
	cleanupTestData(db)

	// 2. 创建律师用户数据
	fmt.Println("\n👨‍⚖️ 创建律师用户数据...")
	lawyers := createLawyers(db)

	// 3. 创建客户数据
	fmt.Println("\n👥 创建客户数据...")
	clients := createClients(db)

	// 4. 创建案件数据
	fmt.Println("\n⚖️ 创建案件数据...")
	cases := createCases(db, lawyers, clients)

	// 输出测试数据总结
	fmt.Println("\n📊 基础测试数据创建完成！")
	fmt.Println("==========================")
	fmt.Printf("✅ 创建了 %d 名律师\n", len(lawyers))
	fmt.Printf("✅ 创建了 %d 个客户\n", len(clients))
	fmt.Printf("✅ 创建了 %d 个案件\n", len(cases))

	fmt.Println("\n🔑 测试账号信息：")
	fmt.Println("律师账号: zhangsan/lisi/wangchen/liusun (密码: 123456)")
	fmt.Println("管理员账号: admin/admin123")

	fmt.Println("\n🎯 利益冲突测试场景：")
	fmt.Println("1. 张三律师 同时代理 'ABC科技有限公司' 和 'XYZ软件公司'")
	fmt.Println("2. 李四律师 同时代理 '王五' 和 '赵六'（离婚案件对立）")
	fmt.Println("3. 王律师 代理 '华夏银行' 和 '中小企业投资公司'")
	fmt.Println("4. 陈律师 同时代理多个关联企业案件")

	fmt.Println("\n🌐 访问地址: http://localhost:3003")
	fmt.Println("🔗 API地址: http://localhost:8080/api/v1")
}

func cleanupTestData(db *gorm.DB) {
	// 按外键依赖顺序删除数据
	db.Unscoped().Where("username IN ?", []string{"zhangsan", "lisi", "wangchen", "liusun"}).Delete(&models.User{})
	db.Unscoped().Where("name LIKE ?", "测试-%").Delete(&models.Client{})
	db.Unscoped().Where("title LIKE ?", "测试-%").Delete(&models.Case{})
}

func createLawyers(db *gorm.DB) []*models.User {
	// 密码哈希（原始密码: 123456）
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	lawyers := []*models.User{
		{
			Username: "zhangsan",
			Name:     "张三",
			Email:    "zhangsan@lawfirm.com",
			Password: string(hashedPassword),
			Role:     "lawyer",
			Phone:    "13800138001",
			Status:   "active",
		},
		{
			Username: "lisi",
			Name:     "李四",
			Email:    "lisi@lawfirm.com",
			Password: string(hashedPassword),
			Role:     "lawyer",
			Phone:    "13800138002",
			Status:   "active",
		},
		{
			Username: "wangchen",
			Name:     "王律师",
			Email:    "wangchen@lawfirm.com",
			Password: string(hashedPassword),
			Role:     "lawyer",
			Phone:    "13800138003",
			Status:   "active",
		},
		{
			Username: "liusun",
			Name:     "陈律师",
			Email:    "liusun@lawfirm.com",
			Password: string(hashedPassword),
			Role:     "lawyer",
			Phone:    "13800138004",
			Status:   "active",
		},
	}

	createdLawyers := make([]*models.User, 0)
	for _, lawyer := range lawyers {
		if err := db.Create(lawyer).Error; err != nil {
			log.Printf("创建律师失败: %v", err)
		} else {
			createdLawyers = append(createdLawyers, lawyer)
			fmt.Printf("   ✅ 创建律师: %s (ID: %d)\n", lawyer.Name, lawyer.ID)
		}
	}
	return createdLawyers
}

func createClients(db *gorm.DB) []*models.Client {
	clients := []*models.Client{
		// 张三律师的客户 - 商业竞争对手
		{
			Name:     "测试-ABC科技有限公司",
			Type:     "企业",
			Email:    "contact@abc-tech.com",
			Phone:    "010-12345678",
			Address:  "北京市朝阳区科技园区A座1001室",
			Company:  "ABC科技有限公司",
			Industry: "软件开发",
			ContactPerson: "张总",
			ContactPhone: "13900138001",
			Status:   "active",
		},
		{
			Name:     "测试-XYZ软件公司",
			Type:     "企业",
			Email:    "contact@xyz-soft.com",
			Phone:    "010-87654321",
			Address:  "北京市海淀区软件园B座2001室",
			Company:  "XYZ软件公司",
			Industry: "软件开发",
			ContactPerson: "李总",
			ContactPhone: "13900138002",
			Status:   "active",
		},
		// 李四律师的客户 - 离婚案件对立双方
		{
			Name:    "测试-王五",
			Type:    "个人",
			Email:   "wangwu@email.com",
			Phone:   "13900138003",
			Address: "北京市朝阳区幸福小区1号楼301室",
			IDCard:  "110101199001011234",
			Status:  "active",
		},
		{
			Name:    "测试-赵六",
			Type:    "个人",
			Email:   "zhaoliu@email.com",
			Phone:   "13900138004",
			Address: "北京市朝阳区幸福小区2号楼301室",
			IDCard:  "110101199002022345",
			Status:  "active",
		},
		// 王律师的客户 - 银行和投资公司
		{
			Name:     "测试-华夏银行",
			Type:     "企业",
			Email:    "legal@huaxia-bank.com",
			Phone:    "010-11111111",
			Address:  "北京市金融街1号华夏银行大厦",
			Company:  "华夏银行股份有限公司",
			Industry: "金融服务",
			ContactPerson: "钱法务",
			ContactPhone: "13900138005",
			Status:   "active",
		},
		{
			Name:     "测试-中小企业投资公司",
			Type:     "企业",
			Email:    "info@sme-investment.com",
			Phone:    "010-22222222",
			Address:  "北京市朝阳区建国门外大街1号",
			Company:  "中小企业投资管理有限公司",
			Industry: "投资管理",
			ContactPerson: "孙经理",
			ContactPhone: "13900138006",
			Status:   "active",
		},
		// 陈律师的客户 - 关联企业
		{
			Name:     "测试-蓝天集团",
			Type:     "企业",
			Email:    "legal@blue-sky.com",
			Phone:    "021-33333333",
			Address:  "上海市浦东新区世纪大道100号蓝天大厦",
			Company:  "蓝天集团有限公司",
			Industry: "投资控股",
			ContactPerson: "周董事长",
			ContactPhone: "13900138007",
			Status:   "active",
		},
		{
			Name:     "测试-白云投资",
			Type:     "企业",
			Email:    "info@white-cloud.com",
			Phone:    "021-44444444",
			Address:  "上海市浦东新区陆家嘴200号白云金融中心",
			Company:  "白云投资有限公司",
			Industry: "投资管理",
			ContactPerson: "吴总经理",
			ContactPhone: "13900138008",
			Status:   "active",
		},
		// 普通客户 - 无冲突
		{
			Name:    "测试-周小明",
			Type:    "个人",
			Email:   "zhouxm@email.com",
			Phone:   "13900138009",
			Address: "北京市西城区新街口外大街19号",
			IDCard:  "110101199003033456",
			Status:  "active",
		},
	}

	createdClients := make([]*models.Client, 0)
	for _, client := range clients {
		if err := db.Create(client).Error; err != nil {
			log.Printf("创建客户失败: %v", err)
		} else {
			createdClients = append(createdClients, client)
			fmt.Printf("   ✅ 创建客户: %s (ID: %d)\n", client.Name, client.ID)
		}
	}
	return createdClients
}

func createCases(db *gorm.DB, lawyers []*models.User, clients []*models.Client) []*models.Case {
	// 建立名称到ID的映射
	clientMap := make(map[string]uint)
	for _, client := range clients {
		clientMap[client.Name] = client.ID
	}

	lawyerMap := make(map[string]uint)
	for _, lawyer := range lawyers {
		lawyerMap[lawyer.Name] = lawyer.ID
	}

	now := time.Now()
	cases := []*models.Case{
		// 张三律师的案件 - 商业竞争对手
		{
			Title:       "测试-ABC科技公司诉XYZ软件公司软件著作权侵权案",
			Description: "ABC科技公司指控XYZ软件公司侵犯其软件著作权，要求停止侵权并赔偿损失。案件涉及复杂的源代码比对和技术分析。",
			ClientID:    clientMap["测试-ABC科技有限公司"],
			LawyerID:    lawyerMap["张三"],
			CaseType:    "知识产权",
			Priority:    "high",
			Status:      "in_progress",
			StartDate:   &now,
		},
		{
			Title:       "测试-XYZ软件公司诉ABC科技公司商业秘密案",
			Description: "XYZ软件公司反诉ABC科技公司存在商业秘密侵权行为，涉及技术机密和客户资源争议。",
			ClientID:    clientMap["测试-XYZ软件公司"],
			LawyerID:    lawyerMap["张三"],
			CaseType:    "商事",
			Priority:    "high",
			Status:      "in_progress",
			StartDate:   &now,
		},
		// 李四律师的案件 - 离婚对立
		{
			Title:       "测试-王五诉赵六离婚财产分割案",
			Description: "王五起诉赵六离婚，主要争议点包括：房产分割（北京3套房产）、子女抚养权（两个子女）、股权分割（公司股份）。",
			ClientID:    clientMap["测试-王五"],
			LawyerID:    lawyerMap["李四"],
			CaseType:    "民事",
			Priority:    "medium",
			Status:      "in_progress",
			StartDate:   &now,
		},
		{
			Title:       "测试-赵六诉王五离婚反诉案",
			Description: "赵六对王五提起的离婚诉讼提出反诉，主要争议：婚前财产认定、债务承担、损害赔偿等。",
			ClientID:    clientMap["测试-赵六"],
			LawyerID:    lawyerMap["李四"],
			CaseType:    "民事",
			Priority:    "medium",
			Status:      "in_progress",
			StartDate:   &now,
		},
		// 王律师的案件 - 银行相关
		{
			Title:       "测试-华夏银行诉中小企业投资公司贷款纠纷案",
			Description: "华夏银行起诉中小企业投资公司，要求偿还逾期贷款本息合计5000万元，涉及抵押物处置和保证责任。",
			ClientID:    clientMap["测试-华夏银行"],
			LawyerID:    lawyerMap["王律师"],
			CaseType:    "商事",
			Priority:    "high",
			Status:      "in_progress",
			StartDate:   &now,
		},
		{
			Title:       "测试-中小企业投资公司诉华夏银行理财纠纷案",
			Description: "中小企业投资公司起诉华夏银行，认为理财产品存在欺诈销售，要求赔偿投资损失3000万元。",
			ClientID:    clientMap["测试-中小企业投资公司"],
			LawyerID:    lawyerMap["王律师"],
			CaseType:    "商事",
			Priority:    "medium",
			Status:      "in_progress",
			StartDate:   &now,
		},
		// 陈律师的案件 - 关联企业
		{
			Title:       "测试-蓝天集团诉白云投资股权转让案",
			Description: "蓝天集团起诉白云投资，要求确认股权转让协议无效，涉及2亿元股权转让款和控股权争议。",
			ClientID:    clientMap["测试-蓝天集团"],
			LawyerID:    lawyerMap["陈律师"],
			CaseType:    "商事",
			Priority:    "high",
			Status:      "in_progress",
			StartDate:   &now,
		},
		{
			Title:       "测试-白云投资诉蓝天集团合作清算案",
			Description: "白云投资起诉蓝天集团，要求进行合作项目清算，涉及投资回报计算和资产分配争议。",
			ClientID:    clientMap["测试-白云投资"],
			LawyerID:    lawyerMap["陈律师"],
			CaseType:    "商事",
			Priority:    "medium",
			Status:      "in_progress",
			StartDate:   &now,
		},
		// 普通案件 - 无冲突
		{
			Title:       "测试-周小明诉前公司劳动争议案",
			Description: "周小明起诉前公司，要求支付违法解除劳动合同赔偿金、加班费等共计50万元。",
			ClientID:    clientMap["测试-周小明"],
			LawyerID:    lawyerMap["陈律师"],
			CaseType:    "劳动",
			Priority:    "low",
			Status:      "in_progress",
			StartDate:   &now,
		},
	}

	createdCases := make([]*models.Case, 0)
	for _, case_ := range cases {
		if err := db.Create(case_).Error; err != nil {
			log.Printf("创建案件失败: %v", err)
		} else {
			createdCases = append(createdCases, case_)
			fmt.Printf("   ✅ 创建案件: %s (ID: %d)\n", case_.Title, case_.ID)
		}
	}
	return createdCases
}