package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

func main() {
	fmt.Println("=== 案件数据检查和初始化 ===")
	fmt.Println("开始时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 连接数据库
	dsn := cfg.GetDatabaseDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 1. 检查现有数据
	fmt.Println("🔍 步骤1: 检查现有案件数据...")
	var caseCount int64
	db.Model(&models.Case{}).Count(&caseCount)
	fmt.Printf("现有案件总数: %d\n", caseCount)

	if caseCount < 10 {
		fmt.Println("⚠️  案件数据不足，补充更多案件数据...")
		additionalCaseData(db)
	} else {
		fmt.Println("✅ 案件数据已充足，显示现有案件:")
		showExistingCases(db)
	}

	fmt.Println()
	fmt.Println("🎉 案件数据初始化完成！")
	fmt.Println("完成时间:", time.Now().Format("2006-01-02 15:04:05"))
}

func showExistingCases(db *gorm.DB) {
	var cases []models.Case
	db.Find(&cases)

	for i, c := range cases {
		var clientName, lawyerName string
		if c.Client != nil {
			clientName = c.Client.Name
		}
		if c.Lawyer != nil {
			lawyerName = c.Lawyer.Name
		}

		fmt.Printf("%d. ID:%d - %s (客户:%s, 律师:%s)\n",
			i+1, c.ID, c.Title, clientName, lawyerName)
	}
}

func additionalCaseData(db *gorm.DB) {
	// 获取律师和客户数据
	var lawyers []models.User
	db.Where("role = ?", "lawyer").Find(&lawyers)

	var clients []models.Client
	db.Find(&clients)

	fmt.Printf("找到 %d 名律师, %d 个客户\n", len(lawyers), len(clients))

	if len(lawyers) == 0 || len(clients) == 0 {
		log.Fatal("律师或客户数据不足，无法创建案件")
	}

	// 创建案件数据 - 使用数据库允许的枚举值
	cases := []models.Case{
		{
			Title:       "阿里巴巴集团商标侵权纠纷案",
			Description: "阿里巴巴集团指控某电商平台商标侵权，要求停止侵权行为并赔偿损失",
			ClientID:    getClientIDByName(clients, "阿里巴巴集团控股有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "zhangwei@law.com"),
			CaseType:    "知识产权",
			Priority:    "high",
			Status:      "pending", // 使用pending而不是active
			StartDate:   &time.Time{},
		},
		{
			Title:       "腾讯控股股权争议案",
			Description: "腾讯控股与某投资公司之间的股权争议，涉及股权转让协议效力认定",
			ClientID:    getClientIDByName(clients, "腾讯控股有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "liming@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是公司业务
			Priority:    "high",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "字节跳动技术秘密保护案",
			Description: "字节跳动起诉前员工技术秘密侵权，要求停止违约行为并赔偿损失",
			ClientID:    getClientIDByName(clients, "字节跳动科技有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "zhangwei@law.com"),
			CaseType:    "知识产权",
			Priority:    "medium",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "中国建筑集团建设工程合同纠纷",
			Description: "中建集团与某房地产开发商之间的建设工程合同纠纷，涉及工程质量款支付问题",
			ClientID:    getClientIDByName(clients, "中国建筑集团有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "wangfang@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是建设工程
			Priority:    "high",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "中国中铁铁路工程纠纷案",
			Description: "中国中铁与某地方政府之间的铁路工程合同纠纷，涉及工程款结算和工期延误问题",
			ClientID:    getClientIDByName(clients, "中国中铁股份有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "wangfang@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是建设工程
			Priority:    "medium",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "万科地产股权转让仲裁案",
			Description: "万科企业股份有限公司与某投资公司之间的股权转让纠纷，通过仲裁程序解决",
			ClientID:    getClientIDByName(clients, "万科企业股份有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "zhangwei@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是公司业务
			Priority:    "high",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "宝能集团收购纠纷案",
			Description: "宝能集团与某上市公司之间的收购纠纷，涉及信息披露义务和收购程序合规性",
			ClientID:    getClientIDByName(clients, "宝能集团股份有限公司"),
			LawyerID:    getLawyerIDByEmail(lawyers, "liming@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是公司并购
			Priority:    "high",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "北京协和医院医疗纠纷案",
			Description: "患者诉北京协和医院医疗损害赔偿纠纷，涉及医疗行为过错认定和损害赔偿问题",
			ClientID:    getClientIDByName(clients, "北京协和医院"),
			LawyerID:    getLawyerIDByEmail(lawyers, "chenhao@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是医疗纠纷
			Priority:    "medium",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "刘德华演艺合同纠纷案",
			Description: "香港著名艺人刘德华与某演出公司之间的演艺合同纠纷，涉及演出费用和违约责任问题",
			ClientID:    getClientIDByName(clients, "刘德华"),
			LawyerID:    getLawyerIDByEmail(lawyers, "chenhao@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是演艺合同
			Priority:    "high",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
		{
			Title:       "王先生医疗损害赔偿案",
			Description: "患者因医疗事故诉某三甲医院要求损害赔偿，涉及医疗过错认定和伤残等级鉴定",
			ClientID:    getClientIDByName(clients, "王先生"),
			LawyerID:    getLawyerIDByEmail(lawyers, "sunlei@law.com"),
			CaseType:    "知识产权", // 使用知识产权而不是医疗纠纷
			Priority:    "low",
			Status:      "pending",
			StartDate:   &time.Time{},
		},
	}

	// 设置开始时间
	now := time.Now()
	for i := range cases {
		cases[i].StartDate = &now
	}

	// 创建案件
	for _, caseData := range cases {
		if err := db.Create(&caseData).Error; err != nil {
			fmt.Printf("❌ 创建案件失败: %s - %v\n", caseData.Title, err)
		} else {
			fmt.Printf("✅ 创建案件成功: %s (ID: %d)\n", caseData.Title, caseData.ID)
		}
	}

	fmt.Printf("\n🎉 成功创建案件数据！\n")
}

func getClientIDByName(clients []models.Client, name string) uint {
	for _, client := range clients {
		if client.Name == name {
			return client.ID
		}
	}
	return 0 // 如果找不到，返回0（应该避免这种情况）
}

func getLawyerIDByEmail(lawyers []models.User, email string) uint {
	for _, lawyer := range lawyers {
		if lawyer.Email == email {
			return lawyer.ID
		}
	}
	return 0 // 如果找不到，返回0（应该避免这种情况）
}