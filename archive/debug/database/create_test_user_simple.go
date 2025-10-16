package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 数据库连接
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	fmt.Println("✅ 数据库连接成功")

	// 检查是否存在admin用户
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = 'admin@example.com'").Scan(&count)
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}

	if count == 0 {
		// 创建测试用户
		_, err = db.Exec(`
			INSERT INTO users (name, email, phone, username, password, role, status, created_at, updated_at)
			VALUES ('管理员', 'admin@example.com', '13800138888', 'admin', 'admin123', 'admin', 'active', NOW(), NOW())
		`)
		if err != nil {
			log.Fatal("创建用户失败:", err)
		}
		fmt.Println("✅ 创建测试用户成功: admin@example.com / admin123")
	} else {
		fmt.Println("ℹ️  测试用户已存在: admin@example.com")
	}

	fmt.Println("\n📝 登录信息:")
	fmt.Println("用户名: admin@example.com")
	fmt.Println("密码: admin123")
	fmt.Println("\n可以使用这些凭据登录系统并测试利益冲突检测功能")
}