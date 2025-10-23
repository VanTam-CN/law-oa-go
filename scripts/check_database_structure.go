package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// PostgreSQL数据库连接
	db, err := sql.Open("postgres", "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable")
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("=== 数据库表结构分析 ===")

	// 查询所有表
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Fatal("查询表列表失败:", err)
	}
	defer rows.Close()

	var tables []string
	fmt.Println("\n数据库中的表:")
	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Printf("扫描表名失败: %v", err)
			continue
		}
		tables = append(tables, tableName)
		fmt.Printf("- %s\n", tableName)
	}

	// 检查用户表结构
	fmt.Println("\n=== 用户表结构 ===")
	userColumns, err := db.Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = 'users' AND table_schema = 'public'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("查询用户表结构失败:", err)
	}
	defer userColumns.Close()

	for userColumns.Next() {
		var columnName, dataType, isNullable, columnDefault sql.NullString
		err := userColumns.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			log.Printf("扫描用户表列失败: %v", err)
			continue
		}
		fmt.Printf("  %s: %s (nullable: %s, default: %s)\n", 
			columnName.String, dataType.String, 
			isNullable.String, columnDefault.String)
	}

	// 查询现有用户数据
	fmt.Println("\n=== 现有用户数据 ===")
	userRows, err := db.Query(`
		SELECT id, username, name, email, role 
		FROM users 
		WHERE role IN ('lawyer', 'admin')
		ORDER BY id
		LIMIT 10
	`)
	if err != nil {
		log.Fatal("查询用户数据失败:", err)
	}
	defer userRows.Close()

	userCount := 0
	for userRows.Next() {
		var id int
		var username, name, email, role sql.NullString
		err := userRows.Scan(&id, &username, &name, &email, &role)
		if err != nil {
			log.Printf("扫描用户数据失败: %v", err)
			continue
		}
		userCount++
		fmt.Printf("用户%d: ID=%d, 用户名=%s, 姓名=%s, 角色=%s\n", 
			userCount, id, username.String, name.String, role.String)
	}

	// 检查客户表结构
	fmt.Println("\n=== 客户表结构 ===")
	clientColumns, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns 
		WHERE table_name = 'clients' AND table_schema = 'public'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("查询客户表结构失败:", err)
	}
	defer clientColumns.Close()

	for clientColumns.Next() {
		var columnName, dataType, isNullable sql.NullString
		err := clientColumns.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			log.Printf("扫描客户表列失败: %v", err)
			continue
		}
		fmt.Printf("  %s: %s (nullable: %s)\n", columnName.String, dataType.String, isNullable.String)
	}

	// 检查案件表结构
	fmt.Println("\n=== 案件表结构 ===")
	caseColumns, err := db.Query(`
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns 
		WHERE table_name = 'cases' AND table_schema = 'public'
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal("查询案件表结构失败:", err)
	}
	defer caseColumns.Close()

	for caseColumns.Next() {
		var columnName, dataType, isNullable sql.NullString
		err := caseColumns.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			log.Printf("扫描案件表列失败: %v", err)
			continue
		}
		fmt.Printf("  %s: %s (nullable: %s)\n", columnName.String, dataType.String, isNullable.String)
	}
}
