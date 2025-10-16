package main

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	fmt.Println("🔍 调试客户ID=13的详细信息")
	fmt.Println("=====================================")

	// 数据库连接
	dsn := "root:@tcp(localhost:3306)/law_oa?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ 数据库连接失败: %v\n", err)
		return
	}
	defer db.Close()

	// 查询客户ID=13的所有字段
	fmt.Println("📋 查询客户ID=13的所有字段:")
	query := `
		SELECT
			id, name, type, email, phone, address,
			company, id_card, industry, contact_person, contact_phone,
			source, notes, status, created_at, updated_at,
			deleted_at
		FROM clients
		WHERE id = 13
	`

	var (
		id uint
		name, clientType, email, phone, address, company, status string
	)
	var idCard, industry, contactPerson, contactPhone, source, notes sql.NullString
	var deletedAt sql.NullTime
	var createdAt, updatedAt time.Time

	err = db.QueryRow(query).Scan(
		&id, &name, &clientType, &email, &phone, &address,
		&company, &idCard, &industry, &contactPerson, &contactPhone,
		&source, &notes, &status, &deletedAt, &createdAt, &updatedAt,
	)

	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}

	fmt.Printf("   ID: %d\n", id)
	fmt.Printf("   Name: '%s'\n", name)
	fmt.Printf("   Type: '%s'\n", clientType)
	fmt.Printf("   Email: '%s'\n", email)
	fmt.Printf("   Phone: '%s'\n", phone)
	fmt.Printf("   Address: '%s'\n", address)
	fmt.Printf("   Company: '%s'\n", company)
	if idCard.Valid {
		fmt.Printf("   ID Card: '%s'\n", idCard.String)
	} else {
		fmt.Printf("   ID Card: NULL\n")
	}
	fmt.Printf("   Status: '%s'\n", status)
	if notes.Valid {
		fmt.Printf("   Notes: '%s'\n", notes.String)
	} else {
		fmt.Printf("   Notes: NULL\n")
	}
	fmt.Printf("   CreatedAt: %v\n", createdAt)
	fmt.Printf("   UpdatedAt: %v\n", updatedAt)

	// 检查是否有NULL值
	fmt.Println("\n🔍 字段值检查:")
	checks := map[string]interface{}{
		"Name":   name,
		"Type":   clientType,
		"Email":  email,
		"Phone":  phone,
		"Address": address,
		"Company": company,
		"Status": status,
	}

	for field, value := range checks {
		status := "✅"
		if value == "" {
			status = "❌ NULL值"
		}
		fmt.Printf("   %s: %s '%s'\n", field, status, value)
	}

	// 检查可空字段
	fmt.Println("\n🔍 可空字段检查:")
	nullChecks := map[string]sql.NullString{
		"ID Card": idCard,
		"Industry": industry,
		"Contact Person": contactPerson,
		"Contact Phone": contactPhone,
		"Source": source,
		"Notes": notes,
	}

	for field, value := range nullChecks {
		status := "✅"
		display := "NULL"
		if value.Valid {
			display = value.String
		} else {
			status = "⚪ NULL"
		}
		fmt.Printf("   %s: %s '%s'\n", field, status, display)
	}

	// 检查是否被软删除
	if deletedAt.Valid {
		fmt.Printf("   软删除: ✅ 已删除 (删除时间: %v)\n", deletedAt.Time)
	} else {
		fmt.Printf("   软删除: ✅ 未删除\n")
	}

	// 比较API返回的值
	fmt.Println("\n📊 与API响应对比:")
	apiValues := map[string]string{
		"name":    "",  // 从API响应中看到name是空的
		"email":   "tamvanchina@gmail.com",
		"phone":   "13580425889",
		"status":  "active",
	}

	for field, apiValue := range apiValues {
		var dbValue string
		switch field {
		case "name":
			dbValue = name
		case "email":
			dbValue = email
		case "phone":
			dbValue = phone
		case "status":
			dbValue = status
		}

		status := "✅"
		if dbValue != apiValue {
			status = "❌ 不匹配"
		}
		fmt.Printf("   %s: 数据库='%s' vs API='%s' %s\n", field, dbValue, apiValue, status)
	}
}