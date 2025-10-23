package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"law-oa-go/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 从环境变量获取数据库配置
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := 5432
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "law_oa_user"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "law_oa_password"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "law_oa_db"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 创建测试用户
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	user := models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Name:      "测试用户",
		Role:      "lawyer",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 检查用户是否已存在
	var existingUser models.User
	result := db.Where("email = ?", user.Email).First(&existingUser)
	if result.Error == nil {
		fmt.Printf("用户 %s 已存在，ID: %d\n", user.Email, existingUser.ID)
		return
	}

	// 创建用户
	if err := db.Create(&user).Error; err != nil {
		log.Fatal("创建用户失败:", err)
	}

	fmt.Printf("测试用户创建成功:\n")
	fmt.Printf("邮箱: %s\n", user.Email)
	fmt.Printf("密码: password123\n")
	fmt.Printf("用户ID: %d\n", user.ID)

	// 创建一些测试案件
	for i := 1; i <= 5; i++ {
		case_ := models.Case{
			Title:       fmt.Sprintf("测试案件%d", i),
			Description: fmt.Sprintf("这是第%d个测试案件的描述", i),
			ClientID:    1,
			LawyerID:    user.ID,
			CaseType:    "civil",
			Priority:    "medium",
			Status:      "pending",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		db.Create(&case_)
		fmt.Printf("创建测试案件 %d，ID: %d\n", i, case_.ID)
	}

	fmt.Println("测试数据创建完成")
}