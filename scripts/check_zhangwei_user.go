//go:build ignore

package main

import (
	"fmt"
	"log"

	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
	"law-oa-go/internal/models"

	"github.com/joho/godotenv"
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

	// 查找张伟用户
	var user models.User
	if err := db.Where("username = ?", "zhangwei").First(&user).Error; err != nil {
		fmt.Printf("❌ 找不到张伟用户: %v\n", err)
		return
	}

	fmt.Printf("✅ 找到张伟用户信息:\n")
	fmt.Printf("   ID: %d\n", user.ID)
	fmt.Printf("   用户名: %s\n", user.Username)
	fmt.Printf("   姓名: %s\n", user.Name)
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   角色: %s\n", user.Role)
	fmt.Printf("   状态: %s\n", user.Status)
	fmt.Printf("   创建时间: %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))

	// 测试密码验证
	fmt.Printf("\n🔑 登录信息:\n")
	fmt.Printf("   用户名: zhangwei\n")
	fmt.Printf("   邮箱: %s\n", user.Email)
	fmt.Printf("   密码: law123456 (应该对应哈希密码)\n")
	fmt.Printf("   登录端点: POST /api/auth/login\n")
	fmt.Printf("   冲突检测端点: POST /api/v1/conflict/check\n")
}