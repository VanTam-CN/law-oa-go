package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	// 数据库连接字符串
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 检查数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库实例失败:", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatal("数据库ping失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 创建测试客户数据
	clients := []models.Client{
		{
			Name:    "张三",
			Type:    "个人",
			Phone:   "13800138001",
			Email:   "zhangsan@example.com",
			Address: "北京市朝阳区建国路88号",
			IDCard:  "110101199001011234",
			Status:  "active",
			Source:  "推荐",
			Notes:  "测试客户数据",
		},
		{
			Name:    "李四科技有限公司",
			Type:    "企业",
			Phone:   "010-88888888",
			Email:   "contact@lisi-tech.com",
			Address: "上海市浦东新区世纪大道100号",
			Company:       "李四科技有限公司",
			Industry:      "软件开发",
			ContactPerson: "李四",
			ContactPhone:  "13900139001",
			Status:        "active",
			Source:        "自主开发",
			Notes:        "企业客户测试数据",
		},
		{
			Name:    "王五",
			Type:    "个人",
			Phone:   "13700137002",
			Email:   "wangwu@example.com",
			Address: "广州市天河区天河路123号",
			IDCard:  "440101199002022345",
			Status:  "active",
			Source:  "网络推广",
			Notes:  "个人客户测试数据",
		},
		{
			Name:    "赵六集团",
			Type:    "企业",
			Phone:   "021-66666666",
			Email:   "info@zhaoliu-group.com",
			Address: "深圳市南山区科技园南区",
			Company:       "赵六集团",
			Industry:      "金融服务",
			ContactPerson: "赵六",
			ContactPhone:  "13600136003",
			Status:        "inactive",
			Source:        "合作机构",
			Notes:        "大型企业客户",
		},
		{
			Name:    "钱七",
			Type:    "个人",
			Phone:   "13500135004",
			Email:   "qianqi@example.com",
			Address: "成都市高新区天府大道456号",
			IDCard:  "510101199003033456",
			Status:  "active",
			Source:  "推荐",
			Notes:  "VIP客户",
		},
		{
			Name:    "孙八律师事务所",
			Type:    "企业",
			Phone:   "0755-88888888",
			Email:   "contact@sunba-law.com",
			Address: "杭州市西湖区文三路789号",
			Company:       "孙八律师事务所",
			Industry:      "法律服务",
			ContactPerson: "孙八",
			ContactPhone:  "13400134005",
			Status:        "active",
			Source:        "其他",
			Notes:        "同行合作客户",
		},
		{
			Name:    "周九",
			Type:    "个人",
			Phone:   "13300133006",
			Email:   "zhoujiu@example.com",
			Address: "武汉市江汉区解放大道321号",
			IDCard:  "420101199004044567",
			Status:  "inactive",
			Source:  "自主开发",
			Notes:  "普通个人客户",
		},
		{
			Name:    "吴十电商",
			Type:    "企业",
			Phone:   "023-77777777",
			Email:   "service@wushi-ecom.com",
			Address: "重庆市渝中区解放碑步行街168号",
			Company:       "吴十电商",
			Industry:      "电子商务",
			ContactPerson: "吴十",
			ContactPhone:  "13200132007",
			Status:        "active",
			Source:        "网络推广",
			Notes:        "电商企业客户",
		},
	}

	fmt.Printf("📝 准备插入 %d 条客户数据...\n", len(clients))

	// 批量插入客户数据
	for i, client := range clients {
		// 设置时间戳
		now := time.Now()
		client.CreatedAt = now
		client.UpdatedAt = now

		// 检查是否已存在（按手机号）
		var existingClient models.Client
		result := db.Where("phone = ?", client.Phone).First(&existingClient)

		if result.Error == nil {
			fmt.Printf("⚠️  客户 '%s' (手机: %s) 已存在，跳过插入\n", client.Name, client.Phone)
			continue
		}

		// 插入新客户
		if err := db.Create(&client).Error; err != nil {
			fmt.Printf("❌ 插入客户 '%s' 失败: %v\n", client.Name, err)
		} else {
			fmt.Printf("✅ 成功插入客户 '%s' (ID: %d)\n", client.Name, client.ID)
		}

		// 每插入一条数据后稍作停顿，避免过快
		if i%3 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Println("\n📊 统计客户数据...")

	// 统计客户总数
	var totalClients int64
	db.Model(&models.Client{}).Count(&totalClients)
	fmt.Printf("总客户数: %d\n", totalClients)

	// 按类型统计
	var personalClients, enterpriseClients int64
	db.Model(&models.Client{}).Where("type = ?", "个人").Count(&personalClients)
	db.Model(&models.Client{}).Where("type = ?", "企业").Count(&enterpriseClients)
	fmt.Printf("个人客户: %d\n", personalClients)
	fmt.Printf("企业客户: %d\n", enterpriseClients)

	// 按状态统计
	var activeClients, inactiveClients int64
	db.Model(&models.Client{}).Where("status = ?", "active").Count(&activeClients)
	db.Model(&models.Client{}).Where("status = ?", "inactive").Count(&inactiveClients)
	fmt.Printf("活跃客户: %d\n", activeClients)
	fmt.Printf("非活跃客户: %d\n", inactiveClients)

	// 显示最新插入的客户
	fmt.Println("\n📋 最新客户列表:")
	var latestClients []models.Client
	db.Order("created_at DESC").Limit(5).Find(&latestClients)

	for _, client := range latestClients {
		status := "活跃"
		if client.Status == "inactive" {
			status = "非活跃"
		}
		fmt.Printf("- %s (%s) - %s - %s\n", client.Name, client.Type, client.Phone, status)
	}

	fmt.Println("\n🎉 客户数据初始化完成！")
	fmt.Println("现在可以重启后端服务并测试客户管理界面了。")
}