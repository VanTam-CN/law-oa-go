package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
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

	// 检查是否已存在测试用户
	var existingUser models.User
	result := db.Where("email = ?", "test@client.com").First(&existingUser)
	if result.Error == nil {
		fmt.Println("⚠️ 测试用户已存在")
		fmt.Printf("邮箱: test@client.com\n")
		fmt.Printf("密码: test123\n")
		return
	}

	// 生成密码哈希
	password := "test123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("密码加密失败:", err)
	}

	// 创建测试用户
	testUser := models.User{
		Username: "testclient",
		Name:     "测试客户",
		Email:    "test@client.com",
		Password: string(hashedPassword),
		Role:     "user",
		Phone:    "13800138999",
		Status:   "active",
	}

	if err := db.Create(&testUser).Error; err != nil {
		log.Fatal("创建测试用户失败:", err)
	}

	fmt.Println("✅ 测试用户创建成功")
	fmt.Printf("邮箱: test@client.com\n")
	fmt.Printf("密码: test123\n")
	fmt.Printf("角色: user\n")
}