package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// 构建数据库连接字符串
	dbUser := getEnv("DB_USERNAME", "law_oa")
	dbPassword := getEnv("DB_PASSWORD", "law_oa_password")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_DATABASE", "law_oa")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// 连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🔗 数据库连接成功")

	// 查看修复前的状态
	fmt.Println("📊 修复前的客户数据状态:")
	showClientStatus(db)

	// 执行修复
	fmt.Println("\n🔧 开始修复客户名称字段...")
	result, err := db.Exec(`
		UPDATE clients 
		SET name = company 
		WHERE (name IS NULL OR name = '') 
		  AND company IS NOT NULL 
		  AND company != ''
	`)
	if err != nil {
		log.Fatalf("Failed to update clients: %v", err)
	}

	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("✅ 修复完成，影响了 %d 行数据\n", rowsAffected)

	// 查看修复后的状态
	fmt.Println("\n📊 修复后的客户数据状态:")
	showClientStatus(db)
}

func showClientStatus(db *sql.DB) {
	rows, err := db.Query(`
		SELECT 
			id,
			COALESCE(name, '') as name,
			COALESCE(company, '') as company,
			COALESCE(email, '') as email,
			CASE 
				WHEN name IS NULL OR name = '' THEN '需要修复'
				ELSE '正常'
			END as name_status
		FROM clients 
		ORDER BY id
	`)
	if err != nil {
		log.Printf("Failed to query clients: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-4s %-20s %-20s %-30s %-10s\n", "ID", "Name", "Company", "Email", "Status")
	fmt.Println("--------------------------------------------------------------------------------")

	emptyNameCount := 0
	totalCount := 0

	for rows.Next() {
		var id int
		var name, company, email, status string

		if err := rows.Scan(&id, &name, &company, &email, &status); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		totalCount++
		if status == "需要修复" {
			emptyNameCount++
		}

		// 截断长字符串以适应显示
		if len(name) > 18 {
			name = name[:15] + "..."
		}
		if len(company) > 18 {
			company = company[:15] + "..."
		}
		if len(email) > 28 {
			email = email[:25] + "..."
		}

		fmt.Printf("%-4d %-20s %-20s %-30s %-10s\n", id, name, company, email, status)
	}

	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("总计: %d 个客户，其中 %d 个需要修复，%d 个正常\n",
		totalCount, emptyNameCount, totalCount-emptyNameCount)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
