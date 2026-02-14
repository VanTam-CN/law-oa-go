//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔍 检查PostgreSQL数据库")
	fmt.Println("===================")

	// 使用默认配置连接数据库
	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 查询用户
	fmt.Println("\n📋 查询用户数据:")
	query := "SELECT id, username, email, role, status FROM users LIMIT 10"
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询用户失败: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("ID | Username | Email | Role | Status")
		fmt.Println("---|----------|-------|------|-------")

		for rows.Next() {
			var id int
			var username, email, role, status string
			if err := rows.Scan(&id, &username, &email, &role, &status); err != nil {
				log.Printf("扫描行失败: %v", err)
				continue
			}
			fmt.Printf("%d | %s | %s | %s | %s\n", id, username, email, role, status)
		}
	}

	// 检查客户数据
	fmt.Println("\n📋 查询客户数据:")
	clientQuery := "SELECT id, name, type, status FROM clients LIMIT 10"
	rows, err = db.Query(clientQuery)
	if err != nil {
		log.Printf("查询客户失败: %v", err)
	} else {
		defer rows.Close()
		fmt.Println("ID | Name | Type | Status")
		fmt.Println("---|------|------|-------")

		for rows.Next() {
			var id int
			var name, clientType, status string
			if err := rows.Scan(&id, &name, &clientType, &status); err != nil {
				log.Printf("扫描客户行失败: %v", err)
				continue
			}
			fmt.Printf("%d | %s | %s | %s\n", id, name, clientType, status)
		}
	}

	// 创建一个测试用户（如果不存在）
	fmt.Println("\n📝 检查并创建测试用户...")
	checkAndCreateUser(db)
}

func checkAndCreateUser(db *sql.DB) {
	// 检查admin用户是否存在
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if err != nil {
		log.Printf("检查admin用户失败: %v", err)
		return
	}

	if count > 0 {
		fmt.Println("✅ admin用户已存在")

		// 获取admin用户的邮箱
		var email string
		err := db.QueryRow("SELECT email FROM users WHERE username = 'admin'").Scan(&email)
		if err == nil {
			fmt.Printf("📧 admin用户邮箱: %s\n", email)
		}
		return
	}

	// 创建admin用户
	fmt.Println("🔨 创建admin用户...")
	insertQuery := `
		INSERT INTO users (username, email, password, role, status, created_at, updated_at)
		VALUES ('admin', 'admin@lawoa.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active', NOW(), NOW())
	`

	_, err = db.Exec(insertQuery)
	if err != nil {
		log.Printf("创建admin用户失败: %v", err)
		return
	}

	fmt.Println("✅ admin用户创建成功")
	fmt.Println("📧 用户名: admin")
	fmt.Println("📧 邮箱: admin@lawoa.com")
	fmt.Println("🔑 密码: admin123")
}