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

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("数据库连接成功")

	// 查看用户表数据
	fmt.Println("\n=== 用户表数据 ===")
	rows, err := db.Query("SELECT id, username, name, email, role, status FROM users LIMIT 10")
	if err != nil {
		log.Fatal("查询用户表失败:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var username, name, email, role, status string
		if err := rows.Scan(&id, &username, &name, &email, &role, &status); err != nil {
			log.Printf("扫描用户数据失败: %v", err)
			continue
		}
		fmt.Printf("ID: %d, 用户名: %s, 姓名: %s, 邮箱: %s, 角色: %s, 状态: %s\n",
			id, username, name, email, role, status)
	}

	// 查看案件表数据
	fmt.Println("\n=== 案件表数据 ===")
	rows, err = db.Query("SELECT id, case_number, title, case_type, status, client_id, lawyer_id FROM cases LIMIT 10")
	if err != nil {
		log.Fatal("查询案件表失败:", err)
	}
	defer rows.Close()

	caseCount := 0
	for rows.Next() {
		var id int
		var caseNumber, title, caseType, status string
		var clientID, lawyerID sql.NullInt64
		if err := rows.Scan(&id, &caseNumber, &title, &caseType, &status, &clientID, &lawyerID); err != nil {
			log.Printf("扫描案件数据失败: %v", err)
			continue
		}
		caseCount++
		fmt.Printf("ID: %d, 案件号: %s, 标题: %s, 类型: %s, 状态: %s, 客户ID: %v, 律师ID: %v\n",
			id, caseNumber, title, caseType, status, clientID, lawyerID)
	}

	if caseCount == 0 {
		fmt.Println("案件表中没有数据")
	}

	// 查看客户表数据
	fmt.Println("\n=== 客户表数据 ===")
	rows, err = db.Query("SELECT id, name, contact_person, phone, email FROM clients LIMIT 10")
	if err != nil {
		log.Fatal("查询客户表失败:", err)
	}
	defer rows.Close()

	clientCount := 0
	for rows.Next() {
		var id int
		var name, contactPerson, phone, email sql.NullString
		if err := rows.Scan(&id, &name, &contactPerson, &phone, &email); err != nil {
			log.Printf("扫描客户数据失败: %v", err)
			continue
		}
		clientCount++
		fmt.Printf("ID: %d, 名称: %s, 联系人: %s, 电话: %s, 邮箱: %s\n",
			id, name.String, contactPerson.String, phone.String, email.String)
	}

	if clientCount == 0 {
		fmt.Println("客户表中没有数据")
	}

	// 查看律师表数据
	fmt.Println("\n=== 律师表数据 ===")
	rows, err = db.Query("SELECT id, name, specialty, phone, email FROM lawyers LIMIT 10")
	if err != nil {
		log.Fatal("查询律师表失败:", err)
	}
	defer rows.Close()

	lawyerCount := 0
	for rows.Next() {
		var id int
		var name, specialty, phone, email sql.NullString
		if err := rows.Scan(&id, &name, &specialty, &phone, &email); err != nil {
			log.Printf("扫描律师数据失败: %v", err)
			continue
		}
		lawyerCount++
		fmt.Printf("ID: %d, 姓名: %s, 专业: %s, 电话: %s, 邮箱: %s\n",
			id, name.String, specialty.String, phone.String, email.String)
	}

	if lawyerCount == 0 {
		fmt.Println("律师表中没有数据")
	}

	fmt.Printf("\n总结: 用户表有数据, 案件表 %d 条记录, 客户表 %d 条记录, 律师表 %d 条记录\n",
		caseCount, clientCount, lawyerCount)
}