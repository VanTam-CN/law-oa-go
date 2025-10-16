package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Client struct {
	ID       uint   `gorm:"column:id"`
	Name     string `gorm:"column:name"`
	Type     string `gorm:"column:type"`
	Email    string `gorm:"column:email"`
	Phone    string `gorm:"column:phone"`
	Status   string `gorm:"column:status"`
}

func main() {
	fmt.Println("🔍 检查数据库中的客户数据")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}

	fmt.Println("✅ 数据库连接成功")

	// 1. 检查客户总数
	var totalClients int64
	db.Model(&Client{}).Count(&totalClients)
	fmt.Printf("\n📊 客户总数: %d\n", totalClients)

	// 2. 查找包含"张三"的客户
	var zhangsanClients []Client
	result := db.Where("name LIKE ?", "%张三%").Find(&zhangsanClients)
	if result.Error != nil {
		fmt.Printf("❌ 查询'张三'失败: %v\n", result.Error)
		return
	}

	fmt.Printf("\n🔍 名字包含'张三'的客户 (%d条):\n", len(zhangsanClients))
	for i, client := range zhangsanClients {
		fmt.Printf("   %d. ID:%d, 姓名:%s, 类型:%s, 状态:%s\n",
			i+1, client.ID, client.Name, client.Type, client.Status)
	}

	// 3. 查找名字等于"张三"的客户
	var exactZhangsanClients []Client
	result = db.Where("name = ?", "张三").Find(&exactZhangsanClients)
	if result.Error != nil {
		fmt.Printf("❌ 精确查询'张三'失败: %v\n", result.Error)
		return
	}

	fmt.Printf("\n🎯 名字等于'张三'的客户 (%d条):\n", len(exactZhangsanClients))
	for i, client := range exactZhangsanClients {
		fmt.Printf("   %d. ID:%d, 姓名:'%s', 类型:%s, 状态:%s\n",
			i+1, client.ID, client.Name, client.Type, client.Status)
	}

	// 4. 查找以"张"开头的客户
	var zhangClients []Client
	result = db.Where("name LIKE ?", "张%").Find(&zhangClients)
	if result.Error != nil {
		fmt.Printf("❌ 查询'张%'失败: %v\n", result.Error)
		return
	}

	fmt.Printf("\n📋 姓名以'张'开头的客户 (%d条):\n", len(zhangClients))
	for i, client := range zhangClients {
		if i >= 10 { // 只显示前10条
			fmt.Printf("   ... 还有 %d 条记录\n", len(zhangClients)-10)
			break
		}
		fmt.Printf("   %d. ID:%d, 姓名:'%s', 类型:%s, 状态:%s\n",
			i+1, client.ID, client.Name, client.Type, client.Status)
	}

	// 5. 测试不同的搜索条件
	fmt.Printf("\n🧪 测试不同搜索条件:\n")

	testSearchCondition(db, "LOWER(name) LIKE ?", "%张三%")
	testSearchCondition(db, "LOWER(name) LIKE ?", "张三%")
	testSearchCondition(db, "LOWER(name) LIKE ?", "%张三")
	testSearchCondition(db, "LOWER(name) = LOWER(?)", "张三")

	// 6. 检查表结构
	fmt.Printf("\n🏗️ 检查clients表结构:\n")
	var columns []struct {
		Field string `gorm:"column:Field"`
		Type  string `gorm:"column:Type"`
		Null  string `gorm:"column:Null"`
		Key   string `gorm:"column:Key"`
	}

	if err := db.Raw("SHOW COLUMNS FROM clients").Scan(&columns).Error; err != nil {
		fmt.Printf("❌ 获取表结构失败: %v\n", err)
		return
	}

	for _, col := range columns {
		fmt.Printf("   - %s: %s %s %s\n", col.Field, col.Type, col.Null, col.Key)
	}
}

func testSearchCondition(db *gorm.DB, condition string, args ...interface{}) {
	var count int64
	var clients []Client

	// 查询总数
	db.Model(&Client{}).Where(condition, args...).Count(&count)

	// 查询前3条记录
	db.Where(condition, args...).Limit(3).Find(&clients)

	fmt.Printf("   条件: %v, 结果: %d条, 示例: ", fmt.Sprintf(condition, args...))
	if len(clients) > 0 {
		fmt.Printf("%s, %s, %s\n", clients[0].Name, clients[1].Name, clients[2].Name)
	} else {
		fmt.Printf("无结果\n")
	}
}