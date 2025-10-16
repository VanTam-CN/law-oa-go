package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"law-oa-go/internal/models"
)

func main() {
	// 数据库连接配置
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 检查用户表
	var users []models.User
	result := db.Find(&users)
	if result.Error != nil {
		log.Fatal("查询用户失败:", result.Error)
	}

	fmt.Printf("📊 用户总数: %d\n", len(users))

	if len(users) == 0 {
		fmt.Println("⚠️ 没有找到用户，需要创建管理员用户")

		// 创建默认管理员用户
		admin := models.User{
			Username: "admin",
			Name:     "系统管理员",
			Email:    "admin@example.com",
			Password: "$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iKVjzieMwkOmANgNOgKQNNBDvAGK", // password123
			Role:     "admin",
			Phone:    "13800138000",
			Status:   "active",
		}

		if err := db.Create(&admin).Error; err != nil {
			log.Fatal("创建管理员失败:", err)
		}

		fmt.Println("✅ 已创建默认管理员用户")
		fmt.Println("   邮箱: admin@example.com")
		fmt.Println("   密码: password123")
	} else {
		fmt.Println("📋 现有用户列表:")
		for _, user := range users {
			fmt.Printf("- %s (%s) - %s - %s\n", user.Name, user.Username, user.Email, user.Role)
		}
	}
}