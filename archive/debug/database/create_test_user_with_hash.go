package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
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

	// 生成密码哈希
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("生成密码哈希失败:", err)
	}

	fmt.Printf("生成的密码哈希: %s\n", string(hash))

	// 清理现有测试用户（如果存在）
	_, err = db.Exec("DELETE FROM users WHERE email = 'testadmin@lawfirm.com'")
	if err != nil {
		log.Printf("清理现有用户失败: %v", err)
	}

	// 创建测试用户
	_, err = db.Exec(`
		INSERT INTO users (username, name, email, password, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'admin', 'active', ?, ?)
	`, "testadmin", "测试管理员", "testadmin@lawfirm.com", string(hash), time.Now(), time.Now())

	if err != nil {
		log.Fatal("创建用户失败:", err)
	}

	fmt.Println("✅ 测试用户创建成功")
	fmt.Printf("邮箱: testadmin@lawfirm.com\n")
	fmt.Printf("密码: %s\n", password)
	fmt.Printf("角色: admin\n")
	fmt.Printf("状态: active\n")

	// 验证用户创建
	var id int
	err = db.QueryRow("SELECT id FROM users WHERE email = 'testadmin@lawfirm.com'").Scan(&id)
	if err != nil {
		log.Fatal("验证用户创建失败:", err)
	}

	fmt.Printf("用户ID: %d\n", id)
}