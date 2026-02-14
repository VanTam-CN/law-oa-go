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

	fmt.Println("🎯 创建字节跳动vs腾讯冲突检测测试数据")
	fmt.Println("=====================================")

	// 1. 清理现有的测试数据
	fmt.Println("🧹 清理现有测试数据...")
	cleanupTestData(db)

	// 2. 创建张伟律师（用户提到的律师）
	fmt.Println("\n👨‍⚖️ 创建张伟律师...")
	zhangwei := createLawyer(db, "zhangwei", "张伟", "zhangwei@lawfirm.com")

	// 3. 创建字节跳动和腾讯等公司客户
	fmt.Println("\n👥 创建公司客户...")
	bytedance := createCompanyClient(db, "字节跳动科技有限公司", "bytedance@company.com", "互联网科技")
	tencent := createCompanyClient(db, "深圳市腾讯计算机系统有限公司", "tencent@company.com", "互联网科技")
	alibaba := createCompanyClient(db, "阿里巴巴集团控股有限公司", "alibaba@company.com", "电子商务")

	// 4. 创建案件 - 张伟律师代理阿里巴巴的案件
	fmt.Println("\n⚖️ 创建案件数据...")
	createCase(db, zhangwei, alibaba, "阿里巴巴诉字节跳动不正当竞争纠纷案", "商业竞争纠纷", "字节跳动科技有限公司")

	// 5. 创建更多测试案件
	createCase(db, zhangwei, bytedance, "字节跳动诉腾讯游戏侵权纠纷案", "知识产权纠纷", "深圳市腾讯计算机系统有限公司")
	createCase(db, zhangwei, tencent, "腾讯诉阿里巴巴垄断纠纷案", "反垄断纠纷", "阿里巴巴集团控股有限公司")

	// 输出测试数据总结
	fmt.Println("\n📊 测试数据创建完成！")
	fmt.Println("========================")
	fmt.Printf("✅ 创建了律师: 张伟 (ID: %d)\n", zhangwei.ID)
	fmt.Printf("✅ 创建了客户: 字节跳动、腾讯、阿里巴巴\n")
	fmt.Printf("✅ 创建了 3 个测试案件\n")

	fmt.Println("\n🎯 测试场景：")
	fmt.Println("1. 张伟律师已代理阿里巴巴，现尝试为字节跳动代理案件（对手为腾讯）")
	fmt.Println("2. 检测行业竞争冲突：互联网科技行业内部竞争")
	fmt.Println("3. 检测直接客户冲突：同时代理对立案件的客户")

	fmt.Println("\n🔑 验证账号信息：")
	fmt.Println("律师账号: zhangwei / law123456")
	fmt.Println("管理员账号: admin / admin123")

	fmt.Println("\n🌐 访问地址: http://localhost:3003")
	fmt.Println("🔗 冲突检测API: http://localhost:8080/api/v1/conflict/check")
}

func cleanupTestData(db *gorm.DB) {
	// 按外键依赖顺序删除数据 - 先删除案件，再删除客户，最后删除用户
	_ = db.Unscoped().Where("title LIKE '测试%' OR title IN ?",
		[]string{"阿里巴巴诉字节跳动不正当竞争纠纷案", "字节跳动诉腾讯游戏侵权纠纷案", "腾讯诉阿里巴巴垄断纠纷案"}).Delete(&models.Case{})

	_ = db.Unscoped().Where("name LIKE '测试%' OR company LIKE '测试%' OR name IN ?",
		[]string{"字节跳动科技有限公司", "深圳市腾讯计算机系统有限公司", "阿里巴巴集团控股有限公司"}).Delete(&models.Client{})

	// 只删除测试律师，保留zhangwei如果已存在
	_ = db.Unscoped().Where("username IN ?", []string{"zhangwei_test", "lisi_test"}).Delete(&models.User{})
}

func createLawyer(db *gorm.DB, username, name, email string) *models.User {
	// 先查找是否已存在
	var lawyer models.User
	if err := db.Where("username = ?", username).First(&lawyer).Error; err == nil {
		fmt.Printf("   ✅ 找到现有律师: %s (ID: %d)\n", lawyer.Name, lawyer.ID)
		return &lawyer
	}

	// 不存在则创建
	lawyer = models.User{
		Username: username,
		Name:     name,
		Email:    email,
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMye.IY4J.y5j7OWg6K8WQV3cLyOv4Nj5.G", // 密码: law123456
		Role:     "lawyer",
		Phone:    "13800138888",
		Status:   "active",
	}

	if err := db.Create(&lawyer).Error; err != nil {
		log.Printf("创建律师失败: %v", err)
		return nil
	}

	fmt.Printf("   ✅ 创建律师: %s (ID: %d)\n", lawyer.Name, lawyer.ID)
	return &lawyer
}

func createCompanyClient(db *gorm.DB, name, email, industry string) *models.Client {
	// 先查找是否已存在
	var client models.Client
	if err := db.Where("name = ?", name).First(&client).Error; err == nil {
		fmt.Printf("   ✅ 找到现有客户: %s (ID: %d)\n", client.Name, client.ID)
		return &client
	}

	// 不存在则创建
	client = models.Client{
		Name:     name,
		Type:     "企业",
		Email:    email,
		Phone:    "010-88888888",
		Address:  "北京市朝阳区科技园区",
		Company:  name,
		Industry: industry,
		Status:   "active",
	}

	if err := db.Create(&client).Error; err != nil {
		log.Printf("创建客户失败: %v", err)
		return nil
	}

	fmt.Printf("   ✅ 创建客户: %s (ID: %d)\n", client.Name, client.ID)
	return &client
}

func createCase(db *gorm.DB, lawyer *models.User, client *models.Client, title, caseType, opposingParty string) *models.Case {
	now := time.Now()

	caseRecord := models.Case{
		Title:        title,
		Description:  fmt.Sprintf("%s与%s之间的%s案件", client.Name, opposingParty, caseType),
		ClientID:     client.ID,
		LawyerID:     lawyer.ID,
		CaseType:     caseType,
		Priority:     "high",
		Status:       "active",
		OpposingParty: opposingParty,
		StartDate:    &now,
		ClientName:   client.Name,
	}

	if err := db.Create(&caseRecord).Error; err != nil {
		log.Printf("创建案件失败: %v", err)
		return nil
	}

	fmt.Printf("   ✅ 创建案件: %s (ID: %d)\n", caseRecord.Title, caseRecord.ID)
	return &caseRecord
}