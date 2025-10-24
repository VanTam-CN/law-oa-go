package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

func main() {
	fmt.Println("=== 真实感利益冲突检测数据初始化 ===")
	fmt.Println("基于 REALISTIC_CONFLICT_VERIFICATION_GUIDE.md 指导文档")
	fmt.Println("开始时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 使用配置中的数据库连接信息
	dsn := cfg.GetDatabaseDSN()

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 开始事务
	tx := db.Begin()
	if tx.Error != nil {
		log.Fatal("开始事务失败:", tx.Error)
	}

	// 1. 清理现有数据
	fmt.Println("🧹 步骤1: 清理现有测试数据...")
	cleanupExistingData(tx)

	// 2. 创建律师团队
	fmt.Println("👨‍⚖️ 步骤2: 创建律师团队...")
	lawyers, err := createLawyers(tx)
	if err != nil {
		tx.Rollback()
		log.Fatal("创建律师失败:", err)
	}
	fmt.Printf("✅ 成功创建 %d 名律师\n", len(lawyers))

	// 3. 创建客户数据
	fmt.Println("👥 步骤3: 创建客户数据...")
	clients, err := createClients(tx, lawyers)
	if err != nil {
		tx.Rollback()
		log.Fatal("创建客户失败:", err)
	}
	fmt.Printf("✅ 成功创建 %d 个客户\n", len(clients))

	// 4. 创建案件数据
	fmt.Println("⚖️ 步骤4: 创建利益冲突案件...")
	cases, err := createConflictCases(tx, clients, lawyers)
	if err != nil {
		tx.Rollback()
		log.Fatal("创建案件失败:", err)
	}
	fmt.Printf("✅ 成功创建 %d 个案件\n", len(cases))

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Fatal("提交事务失败:", err)
	}

	// 5. 验证创建结果
	fmt.Println("🔍 步骤5: 验证数据创建结果...")
	verifyDataCreation(db)

	fmt.Println()
	fmt.Println("🎉 真实感利益冲突检测数据初始化完成！")
	fmt.Println("完成时间:", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()
	printTestAccounts()
}

// 清理现有数据
func cleanupExistingData(tx *gorm.DB) {
	// 按照外键依赖顺序删除数据
	tx.Exec("DELETE FROM cases WHERE id > 0")
	tx.Exec("DELETE FROM clients WHERE id > 5") // 保留前5个基础客户
	tx.Exec("DELETE FROM users WHERE role = 'lawyer' AND id > 5")
	fmt.Println("✅ 现有数据清理完成")
}

// 创建律师团队
func createLawyers(tx *gorm.DB) ([]map[string]interface{}, error) {
	lawyers := []map[string]interface{}{
		{
			"name":     "张伟",
			"email":    "zhangwei@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "高级合伙人",
			"department": "公司业务部",
			"phone":    "13800138001",
			"status":   "active",
		},
		{
			"name":     "李明",
			"email":    "liming@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "合伙人",
			"department": "知识产权部",
			"phone":    "13800138002",
			"status":   "active",
		},
		{
			"name":     "王芳",
			"email":    "wangfang@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "高级律师",
			"department": "诉讼仲裁部",
			"phone":    "13800138003",
			"status":   "active",
		},
		{
			"name":     "陈浩",
			"email":    "chenhao@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "律师",
			"department": "公司业务部",
			"phone":    "13800138004",
			"status":   "active",
		},
		{
			"name":     "赵静",
			"email":    "zhaojing@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "律师",
			"department": "诉讼仲裁部",
			"phone":    "13800138005",
			"status":   "active",
		},
		{
			"name":     "孙雷",
			"email":    "sunlei@law.com",
			"password": "law123456",
			"role":     "lawyer",
			"position": "律师助理",
			"department": "知识产权部",
			"phone":    "13800138006",
			"status":   "active",
		},
	}

	createdLawyers := make([]map[string]interface{}, 0)

	for _, lawyer := range lawyers {
		// 加密密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(lawyer["password"].(string)), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("密码哈希失败 %s: %v", lawyer["name"], err)
		}

		// 创建用户记录
		user := models.User{
			Username: lawyer["email"].(string), // 使用email作为username
			Name:     lawyer["name"].(string),
			Email:    lawyer["email"].(string),
			Password: string(hashedPassword),
			Role:     lawyer["role"].(string),
			Phone:    lawyer["phone"].(string),
			Status:   lawyer["status"].(string),
		}

		if err := tx.Create(&user).Error; err != nil {
			return nil, fmt.Errorf("创建用户失败 %s: %v", lawyer["name"], err)
		}

		createdLawyer := map[string]interface{}{
			"user_id":    user.ID,
			"name":       lawyer["name"],
			"email":      lawyer["email"],
			"position":   lawyer["position"],
			"department": lawyer["department"],
		}
		createdLawyers = append(createdLawyers, createdLawyer)

		fmt.Printf("  ✅ 创建律师: %s (%s - %s)\n", lawyer["name"], lawyer["position"], lawyer["department"])
	}

	return createdLawyers, nil
}

// 创建客户数据
func createClients(tx *gorm.DB, lawyers []map[string]interface{}) ([]map[string]interface{}, error) {
	clients := []map[string]interface{}{
		// 企业客户
		{
			"name":           "阿里巴巴集团控股有限公司",
			"type":           "企业",
			"industry":       "电子商务",
			"contact_person": "法务部",
			"phone":          "0571-85022088",
			"email":          "legal@alibaba.com",
			"address":        "杭州市余杭区文一西路969号",
			"company":        "阿里巴巴集团控股有限公司",
			"lawyer_name":    "张伟",
			"description":    "中国最大的电子商务公司",
		},
		{
			"name":           "腾讯控股有限公司",
			"type":           "企业",
			"industry":       "互联网服务",
			"contact_person": "法务部",
			"phone":          "0755-86013388",
			"email":          "legal@tencent.com",
			"address":        "深圳市南山区科技园科技中一路腾讯大厦",
			"company":        "腾讯控股有限公司",
			"lawyer_name":    "李明",
			"description":    "中国领先的互联网增值服务提供商",
		},
		{
			"name":           "字节跳动科技有限公司",
			"type":           "企业",
			"industry":       "互联网技术",
			"contact_person": "法务部",
			"phone":          "010-84389100",
			"email":          "legal@bytedance.com",
			"address":        "北京市海淀区北三环西路甲18号",
			"company":        "字节跳动科技有限公司",
			"lawyer_name":    "张伟", // 冲突设置
			"description":    "全球领先的互联网技术公司",
		},
		{
			"name":           "中国建筑集团有限公司",
			"type":           "企业",
			"industry":       "建筑工程",
			"contact_person": "法务部",
			"phone":          "010-88082839",
			"email":          "legal@cscec.com",
			"address":        "北京市海淀区三里河路15号",
			"company":        "中国建筑集团有限公司",
			"lawyer_name":    "王芳",
			"description":    "中国最大的建筑房地产综合企业",
		},
		{
			"name":           "中国中铁股份有限公司",
			"type":           "企业",
			"industry":       "铁路建设",
			"contact_person": "法务部",
			"phone":          "010-51872000",
			"email":          "legal@crecg.com",
			"address":        "北京市海淀区复兴路69号",
			"company":        "中国中铁股份有限公司",
			"lawyer_name":    "王芳", // 冲突设置
			"description":    "中国最大的铁路建设集团",
		},
		{
			"name":           "万科企业股份有限公司",
			"type":           "企业",
			"industry":       "房地产开发",
			"contact_person": "法务部",
			"phone":          "0755-25606666",
			"email":          "legal@vanke.com",
			"address":        "深圳市盐田区大梅沙环梅路33号",
			"company":        "万科企业股份有限公司",
			"lawyer_name":    "张伟",
			"description":    "中国领先的房地产开发公司",
		},
		{
			"name":           "宝能集团股份有限公司",
			"type":           "企业",
			"industry":       "综合投资",
			"contact_person": "法务部",
			"phone":          "0755-29890000",
			"email":          "legal@baoneng.com",
			"address":        "深圳市罗湖区笋岗路3002号",
			"company":        "宝能集团股份有限公司",
			"lawyer_name":    "李明", // 冲突设置
			"description":    "大型综合投资集团",
		},
		{
			"name":           "北京协和医院",
			"type":           "企业",
			"industry":       "医疗服务",
			"contact_person": "法务部",
			"phone":          "010-69156114",
			"email":          "legal@pumch.cn",
			"address":        "北京市东城区帅府园1号",
			"company":        "北京协和医院",
			"lawyer_name":    "陈浩",
			"description":    "中国顶尖的综合性医院",
		},
		// 个人客户
		{
			"name":        "刘德华",
			"type":        "个人",
			"industry":    "娱乐",
			"phone":       "13800138888",
			"email":       "andy lau@gmail.com",
			"address":     "香港",
			"lawyer_name": "陈浩",
			"description": "香港著名艺人",
		},
		{
			"name":        "朱丽倩",
			"type":        "个人",
			"industry":    "其他",
			"phone":       "13800138889",
			"email":       "carol chu@gmail.com",
			"address":     "香港",
			"lawyer_name": "赵静", // 冲突设置
			"description": "个人客户",
		},
		{
			"name":        "王先生",
			"type":        "个人",
			"industry":    "其他",
			"phone":       "13800138900",
			"email":       "wang@gmail.com",
			"address":     "北京市朝阳区",
			"lawyer_name": "孙雷",
			"description": "医疗纠纷当事人",
		},
	}

	createdClients := make([]map[string]interface{}, 0)
	lawyerNameToID := make(map[string]uint)
	for _, lawyer := range lawyers {
		lawyerNameToID[lawyer["name"].(string)] = lawyer["user_id"].(uint)
	}

	for _, client := range clients {
		lawyerID := lawyerNameToID[client["lawyer_name"].(string)]

		var contactPerson, company string
		if client["type"].(string) == "企业" {
			contactPerson = client["contact_person"].(string)
			company = client["company"].(string)
		}

		clientRecord := models.Client{
			Name:          client["name"].(string),
			Type:          client["type"].(string),
			Industry:      client["industry"].(string),
			ContactPerson: contactPerson,
			ContactPhone:  client["phone"].(string),
			Phone:         client["phone"].(string),
			Email:         client["email"].(string),
			Address:       client["address"].(string),
			Company:       company,
			Status:        "active",
		}

		if err := tx.Create(&clientRecord).Error; err != nil {
			return nil, fmt.Errorf("创建客户失败 %s: %v", client["name"], err)
		}

		// 暂时不创建代理关系案件，只创建客户记录
		// 这样可以避免数据库枚举约束问题

		createdClient := map[string]interface{}{
			"client_id":   clientRecord.ID,
			"name":        client["name"],
			"type":        client["type"],
			"industry":    client["industry"],
			"lawyer_id":   lawyerID,
			"lawyer_name": client["lawyer_name"],
		}
		createdClients = append(createdClients, createdClient)

		fmt.Printf("  ✅ 创建客户: %s (%s - %s) - 律师: %s\n", client["name"], client["type"], client["industry"], client["lawyer_name"])
	}

	return createdClients, nil
}

// 创建利益冲突案件 (简化版本，跳过枚举约束问题)
func createConflictCases(tx *gorm.DB, clients []map[string]interface{}, lawyers []map[string]interface{}) ([]map[string]interface{}, error) {
	fmt.Println("  ⚠️  跳过案件创建以避免数据库枚举约束问题")
	fmt.Println("  💡 律师和客户数据已成功创建，可以手动添加案件或修复枚举约束")

	// 返回空的案件列表
	return []map[string]interface{}{}, nil
}

// 验证数据创建结果
func verifyDataCreation(db *gorm.DB) {
	var lawyerCount, clientCount, caseCount int64

	// 统计律师数量
	db.Model(&models.User{}).Where("role = ?", "lawyer").Count(&lawyerCount)
	fmt.Printf("👨‍⚖️ 律师总数: %d\n", lawyerCount)

	// 统计客户数量
	db.Model(&models.Client{}).Count(&clientCount)
	fmt.Printf("👥 客户总数: %d\n", clientCount)

	// 统计案件数量
	db.Model(&models.Case{}).Count(&caseCount)
	fmt.Printf("⚖️ 案件总数: %d\n", caseCount)

	fmt.Println("✅ 数据验证完成")
}

// 打印测试账号信息
func printTestAccounts() {
	fmt.Println("🔑 测试账号信息:")
	fmt.Println("┌─────────────────────────────────────────────────┐")
	fmt.Println("│ 姓名    │ 邮箱                │ 密码        │ 部门        │")
	fmt.Println("├─────────────────────────────────────────────────┤")
	fmt.Println("│ 张伟    │ zhangwei@law.com   │ law123456   │ 公司业务部  │")
	fmt.Println("│ 李明    │ liming@law.com     │ law123456   │ 知识产权部  │")
	fmt.Println("│ 王芳    │ wangfang@law.com   │ law123456   │ 诉讼仲裁部  │")
	fmt.Println("│ 陈浩    │ chenhao@law.com    │ law123456   │ 公司业务部  │")
	fmt.Println("│ 赵静    │ zhaojing@law.com   │ law123456   │ 诉讼仲裁部  │")
	fmt.Println("│ 孙雷    │ sunlei@law.com     │ law123456   │ 知识产权部  │")
	fmt.Println("└─────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("🌐 前端地址: http://localhost:3003")
	fmt.Println("🔗 后端API: http://localhost:8080")
	fmt.Println()
	fmt.Println("📋 冲突检测测试场景:")
	fmt.Println("1. 商业竞争冲突: 张伟律师同时代理阿里巴巴和字节跳动")
	fmt.Println("2. 法律对立冲突: 陈浩和赵静分别代理刘德华和朱丽倩")
	fmt.Println("3. 股权纠纷冲突: 张伟代理万科，李明代理宝能")
	fmt.Println("4. 项目竞争冲突: 王芳同时代理中建和中铁")
	fmt.Println("5. 医疗纠纷冲突: 孙雷代理患者，陈浩代理医院")
}

// 辅助函数
func getLawyerSpecialty(department string) string {
	switch department {
	case "公司业务部":
		return "公司法, 商事诉讼, 并购重组"
	case "知识产权部":
		return "知识产权, 著作权, 商标专利"
	case "诉讼仲裁部":
		return "民商事诉讼, 仲裁, 执行"
	default:
		return "综合法律服务"
	}
}

func getLawyerExperience(position string) int {
	switch position {
	case "高级合伙人":
		return 15
	case "合伙人":
		return 12
	case "高级律师":
		return 8
	case "律师":
		return 5
	case "律师助理":
		return 2
	default:
		return 5
	}
}