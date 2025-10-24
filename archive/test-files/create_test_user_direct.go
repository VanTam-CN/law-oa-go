package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint   `gorm:"primarykey"`
	Username  string `gorm:"not null"`    // 添加用户名字段
	Email     string `gorm:"not null"`
	Password  string `gorm:"not null"`
	Name      string `gorm:"not null"`
	Role      string `gorm:"not null;default:'user'"`
	Phone     string `gorm:"default:''"`
	Avatar    string `gorm:"default:''"`
	Status    string `gorm:"default:'active'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func main() {
	// 数据库连接字符串 - 使用正确的PostgreSQL容器配置
	dsn := "host=localhost user=law_oa_user password=law_oa_password dbname=law_oa_db port=5432 sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("数据库连接成功！")

	// 不进行自动迁移，直接使用现有表结构
	// err = db.AutoMigrate(&User{})
	// if err != nil {
	//	 log.Fatal("数据库迁移失败:", err)
	// }

	fmt.Println("跳过数据库迁移，直接创建用户！")

	// 创建测试用户
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("密码加密失败:", err)
	}

	testUser := User{
		Username: "test_lawyer",
		Email:    "test@law.com",
		Password: string(passwordHash),
		Name:     "测试律师",
		Role:     "lawyer",
		Phone:    "13800138000",
		Status:   "active",
	}

	// 检查用户是否已存在
	var existingUser User
	result := db.Where("email = ?", testUser.Email).First(&existingUser)
	if result.Error == nil {
		fmt.Printf("用户 %s 已存在\n", testUser.Email)
	} else {
		// 创建用户
		if err := db.Create(&testUser).Error; err != nil {
			log.Fatal("创建用户失败:", err)
		}
		fmt.Printf("测试用户创建成功！\n")
		fmt.Printf("邮箱: %s\n", testUser.Email)
		fmt.Printf("密码: 123456\n")
		fmt.Printf("角色: %s\n", testUser.Role)
	}

	// 创建管理员用户
	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("管理员密码加密失败:", err)
	}

	adminUser := User{
		Username: "admin",
		Email:    "admin@law.com",
		Password: string(adminPasswordHash),
		Name:     "系统管理员",
		Role:     "admin",
		Phone:    "13900139000",
		Status:   "active",
	}

	// 检查管理员是否已存在
	var existingAdmin User
	result = db.Where("email = ?", adminUser.Email).First(&existingAdmin)
	if result.Error == nil {
		fmt.Printf("管理员 %s 已存在\n", adminUser.Email)
	} else {
		// 创建管理员
		if err := db.Create(&adminUser).Error; err != nil {
			log.Fatal("创建管理员失败:", err)
		}
		fmt.Printf("管理员用户创建成功！\n")
		fmt.Printf("邮箱: %s\n", adminUser.Email)
		fmt.Printf("密码: 123456\n")
		fmt.Printf("角色: %s\n", adminUser.Role)
	}

	fmt.Println("测试数据创建完成！")
}