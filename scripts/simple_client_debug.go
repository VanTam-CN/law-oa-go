//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔍 调试客户ID=13的详细信息")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}
	defer db.Close()

	// 查询客户ID=13的基本信息
	fmt.Println("📋 查询客户ID=13的基本字段:")
	query := `SELECT id, name, type, email, phone, status FROM clients WHERE id = 13`

	var id int
	var name, clientType, email, phone, status string

	err = db.QueryRow(query).Scan(&id, &name, &clientType, &email, &phone, &status)
	if err != nil {
		log.Fatalf("❌ 查询失败: %v", err)
	}

	fmt.Printf("   ID: %d\n", id)
	fmt.Printf("   Name: '%s'\n", name)
	fmt.Printf("   Type: '%s'\n", clientType)
	fmt.Printf("   Email: '%s'\n", email)
	fmt.Printf("   Phone: '%s'\n", phone)
	fmt.Printf("   Status: '%s'\n", status)

	// 检查关键字段
	fmt.Println("\n🔍 字段值检查:")
	checks := map[string]string{
		"Name":   name,
		"Type":   clientType,
		"Email":  email,
		"Phone":  phone,
		"Status": status,
	}

	for field, value := range checks {
		status := "✅"
		if value == "" {
			status = "❌ 空值"
		}
		fmt.Printf("   %s: %s '%s'\n", field, status, value)
	}

	// 与API预期对比
	fmt.Println("\n📊 与API响应对比:")
	apiValues := map[string]string{
		"name":   name,
		"email":  email,
		"phone":  phone,
		"status": status,
	}

	fmt.Println("   数据库中的实际值:")
	for field, value := range apiValues {
		fmt.Printf("   %s: '%s'\n", field, value)
	}

	fmt.Println("\n🎯 结论:")
	if name != "" {
		fmt.Printf("   ✅ 数据库中name字段有值: '%s'\n", name)
		fmt.Println("   🔍 如果API返回空name，问题在于序列化或API层")
	} else {
		fmt.Println("   ❌ 数据库中name字段为空")
		fmt.Println("   🔍 问题在于数据插入或更新")
	}
}