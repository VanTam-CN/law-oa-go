package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Username  string    `gorm:"not null;uniqueIndex"`
	Password  string    `gorm:"not null"`
	Email     string    `gorm:"not null;uniqueIndex"`
	Role      string    `gorm:"not null;default:'user'"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func main() {
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 生成简单密码 "123456" 的哈希
	password := "123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("密码哈希失败:", err)
	}

	// 更新律师45的密码
	result := db.Exec("UPDATE users SET password = ? WHERE id = 45", hashedPassword)
	if result.Error != nil {
		log.Printf("更新密码失败: %v", result.Error)
	} else {
		fmt.Printf("✅ 成功为律师45设置密码: %s\n", password)
	}

	// 验证更新结果
	var user User
	if err := db.First(&user, 45).Error; err != nil {
		log.Printf("查询用户失败: %v", err)
	} else {
		fmt.Printf("✅ 用户信息: ID=%d, 姓名=%s, 角色=%s, 用户名=%s\n", 
			user.ID, user.Name, user.Role, user.Username)
	}
}
