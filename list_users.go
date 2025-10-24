package main

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}

	type User struct {
		ID    uint   `gorm:"id"`
		Email string `gorm:"email"`
		Name  string `gorm:"name"`
		Role  string `gorm:"role"`
	}

	var users []User
	if err := db.Find(&users).Error; err != nil {
		fmt.Printf("查询失败: %v\n", err)
		return
	}

	fmt.Printf("找到用户数: %d\n", len(users))
	for _, user := range users {
		fmt.Printf("ID: %d, Email: %s, Name: %s, Role: %s\n", user.ID, user.Email, user.Name, user.Role)
	}
}