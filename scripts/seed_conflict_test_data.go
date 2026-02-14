//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"

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
		&models.ConflictRule{},
		&models.ConflictCase{},
		&models.ConflictCheckRecord{},
		&models.ClientRelation{},
	); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	fmt.Println("🎯 开始创建利益冲突验证测试数据")
	fmt.Println("==================================")

	// 初始化仓库
	userRepo := repositories.NewUserRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	conflictRepo := repositories.NewConflictRepository(db)
	ctx := context.Background()

	// 1. 清理现有测试数据
	fmt.Println("🧹 清理现有测试数据...")
	cleanupTestData(db, ctx)

	// 2. 创建律师数据
	fmt.Println("\n👨‍⚖️ 创建律师数据...")
	lawyers := createLawyers(userRepo, ctx)

	// 3. 创建客户数据（包含潜在利益冲突的客户）
	fmt.Println("\n👥 创建客户数据...")
	clients := createClients(clientRepo, ctx)

	// 4. 创建案件数据（包含利益冲突案例）
	fmt.Println("\n⚖️ 创建案件数据...")
	cases := createCases(caseRepo, lawyers, clients, ctx)

	// 5. 创建客户关系数据（用于冲突检测）
	fmt.Println("\n🔗 创建客户关系数据...")
	createClientRelations(conflictRepo, ctx)

	// 6. 创建冲突检测规则
	fmt.Println("\n📋 创建冲突检测规则...")
	createConflictRules(conflictRepo, ctx)

	// 7. 创建已知的冲突案例
	fmt.Println("\n⚠️ 创建已知冲突案例...")
	createKnownConflicts(conflictRepo, cases, ctx)

	// 输出测试数据总结
	fmt.Println("\n📊 测试数据创建完成！")
	fmt.Println("==========================")
	fmt.Printf("✅ 创建了 %d 名律师\n", len(lawyers))
	fmt.Printf("✅ 创建了 %d 个客户\n", len(clients))
	fmt.Printf("✅ 创建了 %d 个案件\n", len(cases))
	fmt.Println("\n🎯 利益冲突测试场景：")
	fmt.Println("1. 张三律师 同时代理 'ABC科技有限公司' 和 'XYZ软件公司'（商业竞争关系）")
	fmt.Println("2. 李四律师 同时代理 '王五' 和 '赵六'（离婚案件，双方对立）")
	fmt.Println("3. 王律师 代理 '华夏银行'，其配偶在 '中小企业投资公司' 担任高管（利益关联）")
	fmt.Println("4. 陈律师 同时代理两个关联企业，存在间接利益冲突")
	fmt.Println("\n🔑 验证账号信息：")
	fmt.Println("律师账号: zhangsan/lisi/wangchen/liusun (密码: 123456)")
	fmt.Println("管理员账号: admin/admin123")
	fmt.Println("\n🌐 访问地址: http://localhost:3003")
	fmt.Println("🔗 冲突检测API: http://localhost:8080/api/v1/conflict/check")
}

func cleanupTestData(db *gorm.DB, ctx context.Context) {
	// 按外键依赖顺序删除数据
	db.Unscoped().Where("name LIKE '%测试%'").Delete(&models.User{})
	db.Unscoped().Where("name LIKE '%测试%' OR company LIKE '%测试%'").Delete(&models.Client{})
	db.Unscoped().Where("title LIKE '%测试%'").Delete(&models.Case{})
	db.Unscoped().Where("name LIKE '%测试%'").Delete(&models.ConflictRule{})
	db.Unscoped().Where("check_id LIKE 'TEST-%'").Delete(&models.ConflictCheckRecord{})
	db.Unscoped().Exec("DELETE FROM client_relations WHERE client_id LIKE 'TEST-%' OR related_client_id LIKE 'TEST-%'")
}

func createLawyers(userRepo repositories.UserRepository, ctx context.Context) []*models.User {
	lawyers := []*models.User{
		{
			Username: "zhangsan",
			Name:     "张三",
			Email:    "zhangsan@lawfirm.com",
			Password: "$2a$10$example_hashed_password_1",
			Role:     "lawyer",
			Phone:    "13800138001",
			Status:   "active",
		},
		{
			Username: "lisi",
			Name:     "李四",
			Email:    "lisi@lawfirm.com",
			Password: "$2a$10$example_hashed_password_2",
			Role:     "lawyer",
			Phone:    "13800138002",
			Status:   "active",
		},
		{
			Username: "wangchen",
			Name:     "王律师",
			Email:    "wangchen@lawfirm.com",
			Password: "$2a$10$example_hashed_password_3",
			Role:     "lawyer",
			Phone:    "13800138003",
			Status:   "active",
		},
		{
			Username: "liusun",
			Name:     "陈律师",
			Email:    "liusun@lawfirm.com",
			Password: "$2a$10$example_hashed_password_4",
			Role:     "lawyer",
			Phone:    "13800138004",
			Status:   "active",
		},
	}

	createdLawyers := make([]*models.User, 0)
	for _, lawyer := range lawyers {
		if err := userRepo.Create(ctx, lawyer); err != nil {
			log.Printf("创建律师失败: %v", err)
		} else {
			createdLawyers = append(createdLawyers, lawyer)
			fmt.Printf("   ✅ 创建律师: %s (ID: %d)\n", lawyer.Name, lawyer.ID)
		}
	}
	return createdLawyers
}

func createClients(clientRepo repositories.ClientRepository, ctx context.Context) []*models.Client {
	clients := []*models.Client{
		// 张三律师的客户 - 潜在冲突1
		{
			Name:     "测试-ABC科技有限公司",
			Type:     "企业",
			Email:    "contact@abc-tech.com",
			Phone:    "010-12345678",
			Address:  "北京市朝阳区科技园区A座",
			Company:  "ABC科技有限公司",
			Industry: "软件开发",
			ContactPerson: "张总",
			ContactPhone: "13900138001",
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
			ContactPerson: "李总",
			ContactPhone: "13900138002",
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
		// 王律师的客户 - 利益关联
		{
			Name:     "测试-华夏银行",
			Type:     "企业",
			Email:    "legal@huaxia-bank.com",
			Phone:    "010-11111111",
			Address:  "北京市金融街1号",
			Company:  "华夏银行股份有限公司",
			Industry: "金融服务",
			ContactPerson: "钱法务",
			ContactPhone: "13900138005",
			Status:   "active",
		},
		// 陈律师的客户 - 关联企业1
		{
			Name:     "测试-蓝天集团",
			Type:     "企业",
			Email:    "legal@blue-sky.com",
			Phone:    "021-22222222",
			Address:  "上海市浦东新区世纪大道100号",
			Company:  "蓝天集团有限公司",
			Industry: "投资控股",
			ContactPerson: "孙总监",
			ContactPhone: "13900138006",
			Status:   "active",
		},
		// 陈律师的客户 - 关联企业2
		{
			Name:     "测试-白云投资",
			Type:     "企业",
			Email:    "info@white-cloud.com",
			Phone:    "021-33333333",
			Address:  "上海市浦东新区陆家嘴200号",
			Company:  "白云投资有限公司",
			Industry: "投资管理",
			ContactPerson: "周经理",
			ContactPhone: "13900138007",
			Status:   "active",
		},
		// 普通客户 - 无冲突
		{
			Name:    "测试-周小明",
			Type:    "个人",
			Email:   "zhouxm@email.com",
			Phone:   "13900138008",
			Address: "北京市西城区新街口外大街",
			IDCard:  "110101199003033456",
			Status:  "active",
		},
	}

	createdClients := make([]*models.Client, 0)
	for _, client := range clients {
		if err := clientRepo.Create(ctx, client); err != nil {
			log.Printf("创建客户失败: %v", err)
		} else {
			createdClients = append(createdClients, client)
			fmt.Printf("   ✅ 创建客户: %s (ID: %d)\n", client.Name, client.ID)
		}
	}
	return createdClients
}

func createCases(caseRepo repositories.CaseRepository, lawyers []*models.User, clients []*models.Client, ctx context.Context) []*models.Case {
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
	cases := []*models.Case{
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
		// 王律师的案件 - 利益关联
		{
			Title:       "测试-华夏银行贷款合同纠纷案",
			Description: "华夏银行与中小企业投资公司的贷款合同纠纷",
			ClientID:    clientMap["测试-华夏银行"].ID,
			LawyerID:    lawyerMap["王律师"].ID,
			CaseType:    "合同纠纷",
			Priority:    "high",
			Status:      "active",
			StartDate:   &now,
		},
		// 陈律师的案件 - 关联企业1
		{
			Title:       "测试-蓝天集团股权转让纠纷案",
			Description: "蓝天集团与白云投资公司的股权转让纠纷",
			ClientID:    clientMap["测试-蓝天集团"].ID,
			LawyerID:    lawyerMap["陈律师"].ID,
			CaseType:    "股权纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &now,
		},
		// 陈律师的案件 - 关联企业2
		{
			Title:       "测试-白云投资公司投资纠纷案",
			Description: "白云投资公司与蓝天集团的投资合作纠纷",
			ClientID:    clientMap["测试-白云投资"].ID,
			LawyerID:    lawyerMap["陈律师"].ID,
			CaseType:    "投资纠纷",
			Priority:    "medium",
			Status:      "active",
			StartDate:   &now,
		},
		// 普通案件 - 无冲突
		{
			Title:       "测试-周小明劳动合同纠纷案",
			Description: "周小明与前公司的劳动合同纠纷",
			ClientID:    clientMap["测试-周小明"].ID,
			LawyerID:    lawyerMap["陈律师"].ID,
			CaseType:    "劳动纠纷",
			Priority:    "low",
			Status:      "active",
			StartDate:   &now,
		},
	}

	createdCases := make([]*models.Case, 0)
	for _, case_ := range cases {
		if err := caseRepo.Create(ctx, case_); err != nil {
			log.Printf("创建案件失败: %v", err)
		} else {
			createdCases = append(createdCases, case_)
			fmt.Printf("   ✅ 创建案件: %s (ID: %d)\n", case_.Title, case_.ID)
		}
	}
	return createdCases
}

func createClientRelations(conflictRepo repositories.ConflictRepository, ctx context.Context) {
	relations := []*models.ClientRelation{
		{
			ClientID:        "TEST-ABC-TECH",
			RelatedClientID: "TEST-XYZ-SOFT",
			RelationType:    "BUSINESS_COMPETITOR",
			RelationDetail:  "两家公司为商业竞争对手，在软件开发市场存在直接竞争关系",
			Active:         true,
			CreatedAt:      time.Now(),
		},
		{
			ClientID:        "TEST-BLUE-SKY",
			RelatedClientID: "TEST-WHITE-CLOUD",
			RelationType:    "BUSINESS_PARTNER",
			RelationDetail:  "蓝天集团是白云投资的主要股东之一，存在投资关系",
			Active:         true,
			CreatedAt:      time.Now(),
		},
		{
			ClientID:        "TEST-WANG-WU",
			RelatedClientID: "TEST-ZHAO-LIU",
			RelationType:    "LEGAL_OPPOSING",
			RelationDetail:  "双方为离婚案件当事人，处于法律对立关系",
			Active:         true,
			CreatedAt:      time.Now(),
		},
	}

	for _, relation := range relations {
		if err := conflictRepo.CreateClientRelation(ctx, relation); err != nil {
			log.Printf("创建客户关系失败: %v", err)
		} else {
			fmt.Printf("   ✅ 创建客户关系: %s\n", relation.RelationDetail)
		}
	}
}

func createConflictRules(conflictRepo repositories.ConflictRepository, ctx context.Context) {
	rules := []*models.ConflictRule{
		{
			ID:          "RULE-BUSINESS-COMPETITOR",
			Name:        "商业竞争对手冲突规则",
			Description:  "检测律师是否同时代理存在商业竞争关系的客户",
			Type:        "CLIENT_RELATION",
			Category:    "BUSINESS_CONFLICT",
			Conditions:  models.JSON(`{"relationType": "BUSINESS_COMPETITOR", "timeframe": "2 years"}`),
			Actions:     models.JSONStringArray{"REJECT", "REQUIRE_APPROVAL", "DOCUMENT_DISCLOSURE"},
			Priority:    1,
			Active:      true,
			Version:     1,
			MCPSource:   "MCPv2.0",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "RULE-LEGAL-OPPOSING",
			Name:        "法律对立方冲突规则",
			Description:  "检测律师是否同时代理处于法律对立关系的客户",
			Type:        "CLIENT_RELATION",
			Category:    "LEGAL_CONFLICT",
			Conditions:  models.JSON(`{"relationType": "LEGAL_OPPOSING", "timeframe": "perpetual"}`),
			Actions:     models.JSONStringArray{"REJECT", "IMMEDIATE_DISCLOSURE"},
			Priority:    1,
			Active:      true,
			Version:     1,
			MCPSource:   "MCPv2.0",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "RULE-CORPORATE-RELATION",
			Name:        "企业关联关系冲突规则",
			Description:  "检测律师代理的客户是否存在企业关联关系导致的利益冲突",
			Type:        "CLIENT_RELATION",
			Category:    "CORPORATE_CONFLICT",
			Conditions:  models.JSON(`{"relationType": "BUSINESS_PARTNER", "threshold": "25%"}`),
			Actions:     models.JSONStringArray{"REQUIRE_APPROVAL", "ENHANCED_MONITORING"},
			Priority:    2,
			Active:      true,
			Version:     1,
			MCPSource:   "MCPv2.0",
			CreatedAt:   time.Now(),
		},
	}

	for _, rule := range rules {
		if err := conflictRepo.CreateConflictRule(ctx, rule); err != nil {
			log.Printf("创建冲突规则失败: %v", err)
		} else {
			fmt.Printf("   ✅ 创建冲突规则: %s\n", rule.Name)
		}
	}
}

func createKnownConflicts(conflictRepo repositories.ConflictRepository, cases []*models.Case, ctx context.Context) {
	now := time.Now()
	conflicts := []*models.ConflictCase{
		{
			CheckID:          "TEST-CONFLICT-001",
			CaseID:           fmt.Sprintf("CASE-%d", cases[0].ID),
			CaseName:         cases[0].Title,
			CaseNo:           fmt.Sprintf("2024-TEST-%03d", cases[0].ID),
			ConflictType:     "BUSINESS_COMPETITOR",
			RiskLevel:        "HIGH",
			Description:      "张三律师同时代理ABC科技和XYZ软件两家竞争对手",
			CaseStatus:       "ACTIVE",
			ClientID:         fmt.Sprintf("CLIENT-%d", cases[0].ClientID),
			OpposingParties:  models.JSONStringArray{"XYZ软件公司"},
			ConflictDetails:  "两家公司在同一市场存在直接竞争关系，可能导致利益冲突",
			CreatedAt:        now,
		},
		{
			CheckID:          "TEST-CONFLICT-002",
			CaseID:           fmt.Sprintf("CASE-%d", cases[2].ID),
			CaseName:         cases[2].Title,
			CaseNo:           fmt.Sprintf("2024-TEST-%03d", cases[2].ID),
			ConflictType:     "LEGAL_OPPOSING",
			RiskLevel:        "CRITICAL",
			Description:      "李四律师同时代理离婚案件的双方当事人",
			CaseStatus:       "ACTIVE",
			ClientID:         fmt.Sprintf("CLIENT-%d", cases[2].ClientID),
			OpposingParties:  models.JSONStringArray{"赵六"},
			ConflictDetails:  "律师不能同时代理同一离婚案件的双方当事人",
			CreatedAt:        now,
		},
	}

	for _, conflict := range conflicts {
		if err := conflictRepo.CreateConflictCase(ctx, conflict); err != nil {
			log.Printf("创建已知冲突案例失败: %v", err)
		} else {
			fmt.Printf("   ✅ 创建已知冲突案例: %s\n", conflict.Description)
		}
	}
}