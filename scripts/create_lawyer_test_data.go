//go:build ignore

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

	// 律师测试数据
	lawyers := []struct {
		username string
		name     string
		email    string
		phone    string
		password string
	}{
		{"zhangwei", "张伟律师", "zhangwei@lawoa.com", "13800138001", "lawyer123"},
		{"liming", "李明律师", "liming@lawoa.com", "13800138002", "lawyer123"},
		{"wangfang", "王芳律师", "wangfang@lawoa.com", "13800138003", "lawyer123"},
		{"chenjie", "陈洁律师", "chenjie@lawoa.com", "13800138004", "lawyer123"},
		{"liuhua", "刘华律师", "liuhua@lawoa.com", "13800138005", "lawyer123"},
		{"zhaogang", "赵刚律师", "zhaogang@lawoa.com", "13800138006", "lawyer123"},
		{"sunmei", "孙美律师", "sunmei@lawoa.com", "13800138007", "lawyer123"},
		{"zhouqiang", "周强律师", "zhouqiang@lawoa.com", "13800138008", "lawyer123"},
		{"wuying", "吴颖律师", "wuying@lawoa.com", "13800138009", "lawyer123"},
		{"zhenglei", "郑磊律师", "zhenglei@lawoa.com", "13800138010", "lawyer123"},
	}

	fmt.Printf("开始创建律师测试数据...\n")

	createdCount := 0
	skippedCount := 0

	for _, lawyer := range lawyers {
		// 检查律师用户是否已存在
		var existingUser models.User
		err := db.Where("email = ?", lawyer.email).First(&existingUser).Error
		if err == nil {
			fmt.Printf("⚠️  律师用户已存在，跳过: %s (ID: %d)\n", existingUser.Name, existingUser.ID)
			skippedCount++
			continue
		}

		// 密码哈希
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(lawyer.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败 %s: %v", lawyer.name, err)
			continue
		}

		// 创建律师用户
		newLawyer := &models.User{
			Username: lawyer.username,
			Name:     lawyer.name,
			Email:    lawyer.email,
			Password: string(hashedPassword),
			Role:     "lawyer",
			Status:   "active",
			Phone:    lawyer.phone,
		}

		if err := db.Create(newLawyer).Error; err != nil {
			log.Printf("创建律师用户失败 %s: %v", lawyer.name, err)
			continue
		}

		fmt.Printf("✅ 成功创建律师用户:\n")
		fmt.Printf("   姓名: %s\n", newLawyer.Name)
		fmt.Printf("   邮箱: %s\n", newLawyer.Email)
		fmt.Printf("   电话: %s\n", newLawyer.Phone)
		fmt.Printf("   密码: %s\n", lawyer.password)
		fmt.Printf("   ID: %d\n", newLawyer.ID)
		fmt.Printf("\n")

		createdCount++
	}

	fmt.Printf("律师数据创建完成:\n")
	fmt.Printf("   新建数量: %d\n", createdCount)
	fmt.Printf("   跳过数量: %d\n", skippedCount)
	fmt.Printf("   总律师数: %d\n", createdCount+skippedCount)

	// 验证律师数据
	var lawyerCount int64
	db.Model(&models.User{}).Where("role = ? AND status = ?", "lawyer", "active").Count(&lawyerCount)
	fmt.Printf("   数据库中活跃律师总数: %d\n", lawyerCount)
}