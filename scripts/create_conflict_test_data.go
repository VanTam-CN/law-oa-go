package main

import (
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"

	"github.com/joho/godotenv"
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

	// 自动迁移所有模型
	if err := db.AutoMigrate(
		&models.User{},
		&models.Client{},
		&models.Case{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("🎯 开始创建利益冲突验证测试数据")
	fmt.Println("==================================")

	// 1. 清理现有测试数据
	fmt.Println("🧹 清理现有测试数据...")
	cleanupTestData(db)

	// 2. 创建律师数据
	fmt.Println("\n👨‍⚖️ 创建律师数据...")
	lawyers := createLawyers(db)

	// 3. 创建客户数据（包含潜在利益冲突的客户）
	fmt.Println("\n👥 创建客户数据...")
	clients := createClients(db)

	// 4. 创建案件数据（包含利益冲突案例）
	fmt.Println("\n⚖️ 创建案件数据...")
	createCases(db, lawyers, clients)

	// 输出测试数据总结
	fmt.Println("\n📊 测试数据创建完成！")
	fmt.Println("==========================")
	fmt.Printf("✅ 创建了 %d 名律师\n", len(lawyers))
	fmt.Printf("✅ 创建了 %d 个客户\n", len(clients))
	fmt.Println("\n🎯 利益冲突测试场景：")
	fmt.Println("1. 张三律师 同时代理 'ABC科技有限公司' 和 'XYZ软件公司'（商业竞争关系）")
	fmt.Println("2. 李四律师 同时代理 '王五' 和 '赵六'（离婚案件，双方对立）")
	fmt.Println("\n🔑 验证账号信息：")
	fmt.Println("律师账号: zhangsan/lisi (密码: 123456)")
	fmt.Println("管理员账号: admin/admin123")
	fmt.Println("\n🌐 访问地址: http://localhost:3003")
	fmt.Println("🔗 冲突检测API: http://localhost:8080/api/v1/conflict/check")
}

func cleanupTestData(db *gorm.DB) {
	// 按外键依赖顺序删除数据
	_ = db.Unscoped().Where("username LIKE 'test%'").Delete(&models.User{})
	_ = db.Unscoped().Where("name LIKE '测试%' OR company LIKE '测试%'").Delete(&models.Client{})
	_ = db.Unscoped().Where("title LIKE '测试%'").Delete(&models.Case{})
}

func createLawyers(db *gorm.DB) []*models.User {
	lawyers := []models.User{
		{
			Username: "zhangsan_test",
			Name:     "张三",
			Email:    "zhangsan@lawfirm.com",
			Password: "$2a$10$example_hashed_password_1",
			Role:     "lawyer",
			Phone:    "13800138001",
			Status:   "active",
		},
		{
			Username: "lisi_test",
			Name:     "李四",
			Email:    "lisi@lawfirm.com",
			Password: "$2a$10$example_hashed_password_2",
			Role:     "lawyer",
			Phone:    "13800138002",
			Status:   "active",
		},
	}

	createdLawyers := make([]*models.User, 0)
	for i := range lawyers {
		if err := db.Create(&lawyers[i]).Error; err != nil {
			log.Printf("创建律师失败: %v", err)
		} else {
			createdLawyers = append(createdLawyers, &lawyers[i])
			fmt.Printf("   ✅ 创建律师: %s (ID: %d)\n", lawyers[i].Name, lawyers[i].ID)
		}
	}
	return createdLawyers
}

func createClients(db *gorm.DB) []*models.Client {
	clients := []models.Client{
		// 张三律师的客户 - 潜在冲突1
		{
			Name:     "测试-ABC科技有限公司",
			Type:     "企业",
			Email:    "contact@abc-tech.com",
			Phone:    "010-12345678",
			Address:  "北京市朝阳区科技园区A座",
			Company:  "ABC科技有限公司",
			Industry: "软件开发",
			Status:   "active",
		},
		// 张三律师的客户 - 潜在冲突2
		{
			Name:     "测试-XYZ软件公司",
			Type:     "企业",
			Email:    "contact@xyz-soft.com",
			Phone:    "010-87654321",
			Address:  "北京市海淀区软件园B座",
			Company:  "XYZ软件公司",
			Industry: "软件开发",
			Status:   "active",
		},
		// 李四律师的客户 - 离婚案件冲突
		{
			Name:    "测试-王五",
			Type:    "个人",
			Email:   "wangwu@email.com",
			Phone:   "13900138003",
			Address: "北京市朝阳区幸福小区1号楼",
			IDCard:  "110101199001011234",
			Status:  "active",
		},
		// 李四律师的客户 - 离婚案件冲突
		{
			Name:    "测试-赵六",
			Type:    "个人",
			Email:   "zhaoliu@email.com",
			Phone:   "13900138004",
			Address: "北京市朝阳区幸福小区2号楼",
			IDCard:  "110101199002022345",
			Status:  "active",
		},
		// 普通客户 - 无冲突
		{
			Name:    "测试-周小明",
			Type:    "个人",
			Email:   "zhouxm@email.com",
			Phone:   "13900138005",
			Address: "北京市西城区新街口外大街",
			IDCard:  "110101199003033456",
			Status:  "active",
		},
	}

	createdClients := make([]*models.Client, 0)
	for i := range clients {
		if err := db.Create(&clients[i]).Error; err != nil {
			log.Printf("创建客户失败: %v", err)
		} else {
			createdClients = append(createdClients, &clients[i])
			fmt.Printf("   ✅ 创建客户: %s (ID: %d)\n", clients[i].Name, clients[i].ID)
		}
	}
	return createdClients
}

func createCases(db *gorm.DB, lawyers []*models.User, clients []*models.Client) {
	// 为了方便分配，建立映射
	clientMap := make(map[string]*models.Client)
	for _, client := range clients {
		clientMap[client.Name] = client
	}

	lawyerMap := make(map[string]*models.User)
	for _, lawyer := range lawyers {
		lawyerMap[lawyer.Name] = lawyer
	}

	now := time.Now()
	cases := []models.Case{
		// 张三律师的案件 - 潜在冲突1
		{
			Title:       "测试-ABC科技公司软件著作权纠纷案",
			Description: "ABC科技公司诉XYZ软件公司软件著作权侵权纠纷",
			ClientID:    clientMap["测试-ABC科技有限公司"].ID,
			LawyerID:    lawyerMap["张三"].ID,
			CaseType:    "知识产权纠纷",
			Priority:    "high",
			Status:      "active",
			StartDate:   &now,
		},
		// 张三律师的案件 - 潜在冲突2
		{
			Title:       "测试-XYZ软件公司商业秘密保护案",
			Description: "XYZ软件公司商业秘密被侵犯的保护案件",
			ClientID:    clientMap["测试-XYZ软件公司"].ID,
			LawyerID:    lawyerMap["张三"].ID,
			CaseType:    "商业秘密纠纷",
			Priority:    "high",
			Status:      "active",
			StartDate:   &now,
		},
		// 李四律师的案件 - 离婚冲突1
		{
			Title:       "测试-王五诉赵六离婚纠纷案",
			Description: "王五起诉赵六离婚，涉及财产分割和子女抚养权",
			ClientID:    clientMap["测试-王五"].ID,
			LawyerID:    lawyerMap["李四"].ID,
			CaseType:    "婚姻家庭纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &now,
		},
		// 李四律师的案件 - 离婚冲突2
		{
			Title:       "测试-赵六诉王五离婚反诉案",
			Description: "赵六对王五提起的离婚诉讼进行反诉",
			ClientID:    clientMap["测试-赵六"].ID,
			LawyerID:    lawyerMap["李四"].ID,
			CaseType:    "婚姻家庭纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &now,
		},
		// 普通案件 - 无冲突
		{
			Title:       "测试-周小明劳动合同纠纷案",
			Description: "周小明与前公司的劳动合同纠纷",
			ClientID:    clientMap["测试-周小明"].ID,
			LawyerID:    lawyerMap["李四"].ID,
			CaseType:    "劳动纠纷",
			Priority:    "low",
			Status:      "active",
			StartDate:   &now,
		},
	}

	for i := range cases {
		if err := db.Create(&cases[i]).Error; err != nil {
			log.Printf("创建案件失败: %v", err)
		} else {
			fmt.Printf("   ✅ 创建案件: %s (ID: %d)\n", cases[i].Title, cases[i].ID)
		}
	}
}