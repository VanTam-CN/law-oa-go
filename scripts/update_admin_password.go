package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("🔧 更新admin用户密码")
	fmt.Println("===================")

	// 连接数据库
	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 生成密码哈希
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}

	fmt.Printf("🔐 新密码哈希: %s\n", string(hashedPassword))

	// 更新admin用户密码
	updateQuery := "UPDATE users SET password = $1 WHERE username = 'admin'"
	result, err := db.Exec(updateQuery, string(hashedPassword))
	if err != nil {
		log.Fatalf("更新密码失败: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ 更新了 %d 行\n", rowsAffected)

	// 验证更新
	var email string
	err = db.QueryRow("SELECT email FROM users WHERE username = 'admin'").Scan(&email)
	if err == nil {
		fmt.Printf("📧 admin用户邮箱: %s\n", email)
		fmt.Println("✅ admin用户密码更新成功")
		fmt.Println("🔑 登录信息:")
		fmt.Println("   邮箱: admin@lawoa.com")
		fmt.Println("   密码: admin123")
	}
}