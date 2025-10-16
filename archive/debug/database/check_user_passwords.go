package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接配置
	dsn := "law_oa:law_oa_password@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	fmt.Println("=== 检查用户密码哈希 ===")
	rows, err := db.Query("SELECT id, username, name, email, password, role, status FROM users WHERE email LIKE '%admin%' OR email LIKE '%test%' LIMIT 5")
	if err != nil {
		log.Fatal("查询用户表失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, name, email, password, role, status string
		if err := rows.Scan(&id, &username, &name, &email, &password, &role, &status); err != nil {
			log.Printf("扫描用户数据失败: %v", err)
			continue
		}
		fmt.Printf("ID: %d, 用户名: %s, 姓名: %s, 邮箱: %s, 密码哈希: %s, 角色: %s, 状态: %s\n",
			id, username, name, email, password, role, status)
	}

	// 尝试一些常见密码
	fmt.Println("\n=== 测试常见密码组合 ===")
	testPasswords := []string{"admin", "admin123", "123456", "password", "12345678", "test"}
	testEmails := []string{"admin@law-oa.com", "test@example.com", "admin@example.com"}

	for _, email := range testEmails {
		for _, password := range testPasswords {
			// 这里我们只是查看，实际测试需要通过API
			fmt.Printf("尝试: %s / %s\n", email, password)
		}
	}
}