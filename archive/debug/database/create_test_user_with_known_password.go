package main

import (
	"database/sql"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔧 创建测试用户（已知密码）")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("✅ 数据库连接成功")

	// 生成密码哈希
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("❌ 密码哈希生成失败: %v\n", err)
		return
	}

	fmt.Printf("🔑 生成密码哈希成功 (密码: %s)\n", password)

	// 插入或更新测试用户
	queries := []struct {
		name     string
		email    string
		role     string
		status   string
		sql      string
	}{
		{
			name:   "测试管理员",
			email:  "testadmin@law-oa.com",
			role:   "admin",
			status: "active",
			sql:    `INSERT INTO users (name, email, password, role, status) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE password = ?, name = ?, role = ?, status = ?`,
		},
		{
			name:   "测试律师",
			email:  "testlawyer@law-oa.com",
			role:   "lawyer",
			status: "active",
			sql:    `INSERT INTO users (name, email, password, role, status) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE password = ?, name = ?, role = ?, status = ?`,
		},
	}

	for i, user := range queries {
		fmt.Printf("\n🔍 处理用户 %d: %s\n", i+1, user.name)

		// 执行插入或更新
		_, err = db.Exec(user.sql,
			user.name, user.email, string(hashedPassword), user.role, user.status,
			string(hashedPassword), user.name, user.role, user.status,
		)

		if err != nil {
			fmt.Printf("   ❌ 用户创建失败: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 用户创建成功\n")
		fmt.Printf("   📧 邮箱: %s\n", user.email)
		fmt.Printf("   🔑 密码: %s\n", password)
		fmt.Printf("   🎭 角色: %s\n", user.role)
	}

	// 验证用户是否创建成功
	fmt.Printf("\n🔍 验证用户创建结果:\n")
	rows, err := db.Query("SELECT id, name, email, role, status FROM users WHERE email IN ('testadmin@law-oa.com', 'testlawyer@law-oa.com')")
	if err != nil {
		fmt.Printf("   ❌ 查询用户失败: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name, email, role, status string
		err := rows.Scan(&id, &name, &email, &role, &status)
		if err != nil {
			fmt.Printf("   ❌ 读取用户数据失败: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ 用户: ID=%d, 姓名=%s, 邮箱=%s, 角色=%s, 状态=%s\n", id, name, email, role, status)
	}

	fmt.Printf("\n🚀 使用说明:\n")
	fmt.Printf("1. 用户已创建，可以使用以下凭据登录:\n")
	fmt.Printf("   管理员: testadmin@law-oa.com / %s\n", password)
	fmt.Printf("   律师:   testlawyer@law-oa.com / %s\n", password)
	fmt.Printf("2. 运行登录测试: go run test_default_users.go\n")
	fmt.Printf("3. 获取令牌后设置到前端localStorage\n")
}