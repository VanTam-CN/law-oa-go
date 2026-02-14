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

    // 自动迁移
    err = db.AutoMigrate(&models.User{})
    if err != nil {
        log.Fatalf("数据库迁移失败: %v", err)
    }

    // 密码哈希
    password := "admin123"
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatalf("密码哈希失败: %v", err)
    }

    // 检查用户是否已存在
    var existingUser models.User
    err = db.Where("email = ?", "admin@lawoa.com").First(&existingUser).Error
    if err == nil {
        fmt.Printf("⚠️  管理员用户已存在: %s (ID: %d)\n", existingUser.Name, existingUser.ID)
        return
    }

    // 创建管理员用户
    user := &models.User{
        Name:     "系统管理员",
        Email:    "admin@lawoa.com",
        Password: string(hashedPassword),
        Role:     "admin",
        Status:   "active",
    }

    // 创建用户
    result := db.Create(user)
    if result.Error != nil {
        log.Fatalf("创建用户失败: %v", result.Error)
    }

    fmt.Printf("✅ 成功创建管理员用户:\n")
    fmt.Printf("   姓名: %s\n", user.Name)
    fmt.Printf("   邮箱: %s\n", user.Email)
    fmt.Printf("   密码: %s\n", password)
    fmt.Printf("   角色: %s\n", user.Role)
    fmt.Printf("   ID: %d\n", user.ID)
}
