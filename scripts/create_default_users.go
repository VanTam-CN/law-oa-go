package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 使用配置中的数据库连接信息
	dsn := cfg.GetDatabaseDSN()

	// 连接数据库
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 检查用户表是否存在，不存在则创建
	if !db.Migrator().HasTable(&models.User{}) {
		if err := db.AutoMigrate(&models.User{}); err != nil {
			log.Fatalf("创建用户表失败: %v", err)
		}
	}

	// 密码哈希
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码哈希失败: %v", err)
	}

	// 检查管理员用户是否已存在
	var existingUser models.User
	err = db.Where("email = ?", "admin@lawoa.com").First(&existingUser).Error
	if err == nil {
		fmt.Printf("✅ 管理员用户已存在: %s (ID: %d)\n", existingUser.Email, existingUser.ID)
		return
	}

	// 创建管理员用户
	admin := &models.User{
		Username: "admin",
		Name:     "系统管理员",
		Email:    "admin@lawoa.com",
		Password: string(hashedPassword),
		Role:     "admin",
		Status:   "active",
	}

	// 创建用户
	if err := db.Create(admin).Error; err != nil {
		log.Fatalf("创建管理员用户失败: %v", err)
	}

	fmt.Printf("✅ 成功创建管理员用户:\n")
	fmt.Printf("   邮箱: admin@lawoa.com\n")
	fmt.Printf("   密码: %s\n", password)
	fmt.Printf("   用户名: %s\n", admin.Username)
	fmt.Printf("   角色: %s\n", admin.Role)
	fmt.Printf("   ID: %d\n", admin.ID)

	// 创建测试律师用户
	lawyerPassword := "lawyer123"
	lawyerHashedPassword, err := bcrypt.GenerateFromPassword([]byte(lawyerPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("律师密码哈希失败: %v", err)
	}

	var existingLawyer models.User
	err = db.Where("email = ?", "lawyer@lawoa.com").First(&existingLawyer).Error
	if err == nil {
		fmt.Printf("✅ 律师用户已存在: %s (ID: %d)\n", existingLawyer.Email, existingLawyer.ID)
	} else {
		lawyer := &models.User{
			Username: "lawyer",
			Name:     "测试律师",
			Email:    "lawyer@lawoa.com",
			Password: string(lawyerHashedPassword),
			Role:     "lawyer",
			Status:   "active",
		}

		if err := db.Create(lawyer).Error; err != nil {
			log.Printf("创建律师用户失败: %v", err)
		} else {
			fmt.Printf("✅ 成功创建律师用户:\n")
			fmt.Printf("   邮箱: lawyer@lawoa.com\n")
			fmt.Printf("   密码: %s\n", lawyerPassword)
			fmt.Printf("   用户名: %s\n", lawyer.Username)
			fmt.Printf("   角色: %s\n", lawyer.Role)
			fmt.Printf("   ID: %d\n", lawyer.ID)
		}
	}
}