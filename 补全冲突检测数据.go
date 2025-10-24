package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Case struct {
	ID          uint      `gorm:"primaryKey"`
	Title       string    `gorm:"size:255;not null"`
	ClientID    uint      `gorm:"not null"`
	LawyerID    uint      `gorm:"not null"`
	CaseType    string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	Status      string    `gorm:"size:50;default:'active'"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}

type Client struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:255;not null;uniqueIndex"`
	Type     string `gorm:"size:50;not null"`
	LawyerID uint   `gorm:"index"`
	Email    string `gorm:"size:255"`
	Phone    string `gorm:"size:50"`
	Address  string `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:100;not null"`
	Email    string `gorm:"size:255;not null;uniqueIndex"`
	Password string `gorm:"size:255;not null"`
	Role     string `gorm:"size:50;not null;default:'client'"`
	Status   string `gorm:"size:50;default:'active'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func main() {
	// 数据库连接
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("🔄 开始补全冲突检测测试数据...")

	// 1. 验证律师ID映射
	fmt.Println("\n📋 验证律师ID映射...")
	lawyerMap := map[string]uint{
		"张伟": 45,
		"李明": 46,
		"王芳": 47,
		"陈浩": 48,
		"赵静": 49,
		"孙雷": 50,
	}

	// 验证客户ID映射
	clientMap := map[string]uint{
		"阿里巴巴集团控股有限公司": 55,
		"腾讯控股有限公司":       56,
		"字节跳动科技有限公司":    57,
		"中国建筑集团有限公司":    58,
		"中国中铁股份有限公司":    59,
		"万科企业股份有限公司":    60,
		"宝能集团股份有限公司":    61,
		"北京协和医院":         62,
		"刘德华":             63,
		"朱丽倩":             64,
		"王先生":             65,
	}

	fmt.Printf("律师映射: %v\n", lawyerMap)
	fmt.Printf("客户映射: %v\n", clientMap)

	// 2. 补全关键冲突案件
	fmt.Println("\n📝 补全关键冲突案件...")

	// 字节跳动诉腾讯垄断纠纷案 (张伟代理字节跳动，对手方是腾讯)
	conflictCase1 := Case{
		Title:       "字节跳动诉腾讯垄断纠纷案",
		ClientID:    clientMap["字节跳动科技有限公司"],
		LawyerID:    lawyerMap["张伟"],
		CaseType:    "知识产权",
		Description: "字节跳动诉腾讯滥用市场支配地位，请求停止垄断行为并赔偿损失",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 阿里巴巴诉字节跳动不正当竞争纠纷案 (张伟代理阿里巴巴，字节跳动是对手方)
	conflictCase2 := Case{
		Title:       "阿里巴巴诉字节跳动不正当竞争纠纷案",
		ClientID:    clientMap["阿里巴巴集团控股有限公司"],
		LawyerID:    lawyerMap["张伟"],
		CaseType:    "知识产权",
		Description: "阿里巴巴指控字节跳动在其平台上实施不正当竞争行为，要求停止侵权并赔偿",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-24 * time.Hour), // 一天前
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}

	// 腾讯诉抖音短视频版权侵权案 (李明代理腾讯，抖音是字节跳动产品)
	conflictCase3 := Case{
		Title:       "腾讯诉抖音短视频版权侵权案",
		ClientID:    clientMap["腾讯控股有限公司"],
		LawyerID:    lawyerMap["李明"],
		CaseType:    "知识产权",
		Description: "腾讯指控抖音平台存在大量侵权内容，要求下架相关视频并赔偿损失",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-48 * time.Hour), // 两天前
		UpdatedAt:   time.Now().Add(-48 * time.Hour),
	}

	// 字节跳动诉腾讯技术秘密侵权案 (张伟代理字节跳动，涉及腾讯商业机密)
	conflictCase4 := Case{
		Title:       "字节跳动诉腾讯技术秘密侵权案",
		ClientID:    clientMap["字节跳动科技有限公司"],
		LawyerID:    lawyerMap["张伟"],
		CaseType:    "知识产权",
		Description: "字节跳动指控腾讯获取其商业机密，要求停止侵权并赔偿巨额损失",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-72 * time.Hour), // 三天前
		UpdatedAt:   time.Now().Add(-72 * time.Hour),
	}

	// 万科诉宝能恶意收购纠纷案 (张伟代理万科，宝能是收购方)
	conflictCase5 := Case{
		Title:       "万科企业股份有限公司诉宝能集团股份有限公司恶意收购纠纷案",
		ClientID:    clientMap["万科企业股份有限公司"],
		LawyerID:    lawyerMap["张伟"],
		CaseType:    "知识产权",
		Description: "万科指控宝能恶意收购扰乱市场秩序，要求停止收购行为",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-96 * time.Hour), // 四天前
		UpdatedAt:   time.Now().Add(-96 * time.Hour),
	}

	// 中国建筑与中国中铁项目竞标纠纷案 (王芳代理中建，中铁是竞争对手)
	conflictCase6 := Case{
		Title:       "中国建筑集团有限公司诉中国中铁股份有限公司项目竞标纠纷案",
		ClientID:    clientMap["中国建筑集团有限公司"],
		LawyerID:    lawyerMap["王芳"],
		CaseType:    "知识产权",
		Description: "中建指控中铁在项目竞标中存在不正当竞争行为",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-120 * time.Hour), // 五天前
		UpdatedAt:   time.Now().Add(-120 * time.Hour),
	}

	// 刘德华诉朱丽倩离婚纠纷案 (陈浩代理刘德华，朱丽倩是对方当事人)
	conflictCase7 := Case{
		Title:       "刘德华诉朱丽倩离婚纠纷案",
		ClientID:    clientMap["刘德华"],
		LawyerID:    lawyerMap["陈浩"], // 应该是48 (陈浩)
		CaseType:    "知识产权",
		Description: "刘德华提出离婚诉讼，涉及子女抚养和财产分割",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-168 * time.Hour), // 一周前
		UpdatedAt:   time.Now().Add(-168 * time.Hour),
	}

	// 朱丽倩诉刘德华离婚反诉案 (赵静代理朱丽倩，刘德华是对方当事人)
	conflictCase8 := Case{
		Title:       "朱丽倩诉刘德华离婚反诉案",
		ClientID:    clientMap["朱丽倩"],
		LawyerID:    lawyerMap["赵静"], // 应该是49 (赵静)
		CaseType:    "知识产权",
		Description: "朱丽倩对离婚诉讼提出反诉，要求重新分割财产",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-168 * time.Hour), // 一周前
		UpdatedAt:   time.Now().Add(-168 * time.Hour),
	}

	// 北京协和医院医疗纠纷案 (孙雷代理王先生，协和医院是对方当事人)
	conflictCase9 := Case{
		Title:       "王先生诉北京协和医院医疗事故纠纷案",
		ClientID:    clientMap["王先生"],
		LawyerID:    lawyerMap["孙雷"],
		CaseType:    "知识产权",
		Description: "王先生指控协和医院存在医疗过错，要求赔偿损失",
		Status:      "pending",
		CreatedAt:   time.Now().Add(-240 * time.Hour), // 十天前
		UpdatedAt:   time.Now().Add(-240 * time.Hour),
	}

	// 保存案件数据
	conflictCases := []Case{conflictCase1, conflictCase2, conflictCase3, conflictCase4, conflictCase5, conflictCase6, conflictCase7, conflictCase8, conflictCase9}

	for i, case_ := range conflictCases {
		// 检查是否已存在相同标题的案件
		var existingCase Case
		if err := db.Where("title = ?", case_.Title).First(&existingCase).Error; err == nil {
			fmt.Printf("⚠️ 案件已存在，跳过: %s\n", case_.Title)
			continue
		}

		if err := db.Create(&case_).Error; err != nil {
			log.Printf("❌ 创建案件 %d 失败: %v", i+1, err)
		} else {
			fmt.Printf("✅ 创建案件: %s (客户ID: %d, 律师ID: %d)\n", case_.Title, case_.ClientID, case_.LawyerID)
		}
	}

	// 3. 更新客户关联律师
	fmt.Println("\n🔄 更新客户律师关联...")
	updateClientRelations := []struct {
		clientName string
		lawyerName string
	}{
		{"字节跳动科技有限公司", "张伟"},
		{"阿里巴巴集团控股有限公司", "张伟"},
		{"腾讯控股有限公司", "李明"},
		{"中国建筑集团有限公司", "王芳"},
		{"中国中铁股份有限公司", "王芳"},
		{"万科企业股份有限公司", "张伟"},
		{"宝能集团股份有限公司", "李明"},
		{"北京协和医院", "陈浩"},
		{"刘德华", "陈浩"},
		{"朱丽倩", "赵静"},
		{"王先生", "孙雷"},
	}

	for _, update := range updateClientRelations {
		// 查找客户
		var client Client
		if err := db.Where("name = ?", update.clientName).First(&client).Error; err != nil {
			log.Printf("⚠️ 未找到客户: %s", update.clientName)
			continue
		}

		// 查找律师
		var user User
		if err := db.Where("name = ? AND role = ?", update.lawyerName, "lawyer").First(&user).Error; err != nil {
			log.Printf("⚠️ 未找到律师: %s", update.lawyerName)
			continue
		}

		// 更新客户的律师ID
		if err := db.Model(&client).Update("lawyer_id", user.ID).Error; err != nil {
			log.Printf("⚠️ 更新客户律师关联失败: %s (%s)", update.clientName, err)
		} else {
			fmt.Printf("✅ 更新客户 %s 的律师为 %s (ID: %d)\n", update.clientName, update.lawyerName, user.ID)
		}
	}

	// 4. 验证数据完整性
	fmt.Println("\n🔍 验证数据完整性...")

	// 验证案件数据
	var caseCount int64
	db.Model(&Case{}).Where("deleted_at IS NULL").Count(&caseCount)
	fmt.Printf("✅ 案件总数: %d\n", caseCount)

	// 验证关键冲突案件
	var keyConflictCases []Case
	db.Where("title LIKE ? OR title LIKE ? OR title LIKE ?",
		"%字节跳动%", "%阿里巴巴%", "%腾讯%").Find(&keyConflictCases)
	fmt.Printf("✅ 关键冲突案件数量: %d\n", len(keyConflictCases))

	for _, case_ := range keyConflictCases {
		fmt.Printf("  - %s (律师ID: %d, 客户ID: %d)\n", case_.Title, case_.LawyerID, case_.ClientID)
	}

	// 验证客户律师关联
	var clientCount int64
	db.Model(&Client{}).Where("deleted_at IS NULL").Count(&clientCount)
	var withLawyerCount int64
	db.Model(&Client{}).Where("deleted_at IS NULL AND lawyer_id > 0").Count(&withLawyerCount)
	fmt.Printf("✅ 客户总数: %d, 已关联律师的客户数: %d\n", clientCount, withLawyerCount)

	// 验证张伟律师的互联网公司客户
	var zhangweiCases []Case
	db.Where("lawyer_id = ? AND deleted_at IS NULL", lawyerMap["张伟"]).Find(&zhangweiCases)
	fmt.Printf("✅ 张伟律师案件数: %d\n", len(zhangweiCases))

	internetCompanies := 0
	seen := make(map[string]bool)
	for _, case_ := range zhangweiCases {
		var client Client
		if err := db.First(&client, case_.ClientID).Error; err != nil {
			continue
		}
		if isInternetCompany(client.Name) && !seen[client.Name] {
			internetCompanies++
			seen[client.Name] = true
			fmt.Printf("   互联网客户: %s\n", client.Name)
		}
	}
	fmt.Printf("✅ 张伟代理的互联网公司数量: %d\n", internetCompanies)

	// 验证李明律师的案件（腾讯）
	var limingCases []Case
	db.Where("lawyer_id = ? AND deleted_at IS NULL", lawyerMap["李明"]).Find(&limingCases)
	fmt.Printf("✅ 李明律师案件数: %d\n", len(limingCases))

	limingClients := make(map[string]bool)
	for _, case_ := range limingCases {
		var client Client
		if err := db.First(&client, case_.ClientID).Error; err != nil {
			continue
		}
		limingClients[client.Name] = true
		fmt.Printf("  李明代理客户: %s\n", client.Name)
	}

	fmt.Printf("✅ 数据补全完成！\n")
	fmt.Printf("✅ 总计创建 %d 个关键冲突案件\n", len(conflictCases))
	fmt.Printf("✅ 更新 %d 个客户的律师关联\n", len(updateClientRelations))
	fmt.Printf("✅ 现在可以进行完整的冲突检测验证\n")
}

func isInternetCompany(companyName string) bool {
	companyName = strings.ToLower(companyName)
	internetKeywords := []string{
		"字节跳动", "阿里", "腾讯", "百度", "京东", "美团", "滴滴", "快手", "小红书",
		"科技", "网络", "软件", "互联网", "电商", "社交媒体",
	}

	for _, keyword := range internetKeywords {
		if strings.Contains(companyName, keyword) {
			return true
		}
	}
	return false
}