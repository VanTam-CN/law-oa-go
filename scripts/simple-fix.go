//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🚨 PostgreSQL 简单修复工具")
	fmt.Println("==========================")

	// 获取连接参数
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "law_oa_go")

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ 数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("❌ 数据库ping失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println()

	// 1. 检查当前表状态
	fmt.Println("🔍 检查当前表状态...")
	tables := getCurrentTables(db)
	for _, table := range tables {
		fmt.Printf("📋 %s\n", table)
	}
	fmt.Println()

	// 2. 创建缺失的表
	fmt.Println("🛠️ 创建缺失的表...")
	createMissingTables(db)

	// 3. 添加缺失的字段
	fmt.Println("🛠️ 添加缺失的字段...")
	addMissingColumns(db)

	// 4. 插入示例数据
	fmt.Println("📝 插入示例数据...")
	insertSampleData(db)

	// 5. 最终验证
	fmt.Println("🔍 最终验证...")
	finalVerification(db)

	fmt.Println("\n🎉 PostgreSQL数据库修复完成！")
	fmt.Println("请刷新数据库管理工具查看新的表结构")
}

func getCurrentTables(db *sql.DB) []string {
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询表失败: %v", err)
		return nil
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("扫描表名失败: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}
	return tables
}

func createMissingTables(db *sql.DB) {
	tablesToCreate := []struct {
		name string
		sql  string
	}{
		{
			"lawyers",
			`CREATE TABLE IF NOT EXISTS lawyers (
				id BIGSERIAL PRIMARY KEY,
				lawyer_name VARCHAR(50) NOT NULL,
				phone VARCHAR(20),
				email VARCHAR(100),
				license_no VARCHAR(50) UNIQUE,
				position VARCHAR(50),
				department VARCHAR(100),
				specialty TEXT,
				status VARCHAR(20) DEFAULT 'active',
				remark TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP WITH TIME ZONE
			)`,
		},
		{
			"departments",
			`CREATE TABLE IF NOT EXISTS departments (
				id BIGSERIAL PRIMARY KEY,
				name VARCHAR(100) NOT NULL,
				code VARCHAR(50) NOT NULL UNIQUE,
				parent_id BIGINT DEFAULT 0,
				leader_id BIGINT,
				description TEXT,
				sort_order INTEGER DEFAULT 0,
				status INTEGER DEFAULT 1,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP WITH TIME ZONE
			)`,
		},
		{
			"documents",
			`CREATE TABLE IF NOT EXISTS documents (
				id BIGSERIAL PRIMARY KEY,
				document_no VARCHAR(50) UNIQUE NOT NULL,
				case_id BIGINT NOT NULL,
				client_id BIGINT,
				file_name VARCHAR(255) NOT NULL,
				original_name VARCHAR(255) NOT NULL,
				file_size BIGINT NOT NULL,
				file_type VARCHAR(100) NOT NULL,
				file_hash VARCHAR(64) NOT NULL,
				file_path VARCHAR(500) NOT NULL,
				document_type VARCHAR(50) NOT NULL,
				description TEXT,
				tags TEXT,
				is_public BOOLEAN DEFAULT FALSE,
				is_confidential BOOLEAN DEFAULT FALSE,
				uploader_id BIGINT NOT NULL,
				upload_time TIMESTAMP WITH TIME ZONE NOT NULL,
				download_count INTEGER DEFAULT 0,
				last_download_time TIMESTAMP WITH TIME ZONE,
				status VARCHAR(20) DEFAULT 'active',
				thumbnail_path VARCHAR(500),
				metadata TEXT,
				remark TEXT,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMP WITH TIME ZONE
			)`,
		},
	}

	for _, table := range tablesToCreate {
		if !tableExists(db, table.name) {
			fmt.Printf("📝 创建 %s 表...\n", table.name)
			if _, err := db.Exec(table.sql); err != nil {
				log.Printf("❌ 创建 %s 表失败: %v", table.name, err)
			} else {
				fmt.Printf("✅ %s 表创建成功\n", table.name)
				createIndexes(db, table.name)
			}
		} else {
			fmt.Printf("✅ %s 表已存在\n", table.name)
		}
	}
}

func createIndexes(db *sql.DB, tableName string) {
	var indexes []string

	switch tableName {
	case "lawyers":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS idx_lawyers_lawyer_name ON lawyers(lawyer_name)",
			"CREATE INDEX IF NOT EXISTS idx_lawyers_phone ON lawyers(phone)",
			"CREATE INDEX IF NOT EXISTS idx_lawyers_license_no ON lawyers(license_no)",
			"CREATE INDEX IF NOT EXISTS idx_lawyers_status ON lawyers(status)",
		}
	case "departments":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS idx_departments_code ON departments(code)",
			"CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id)",
			"CREATE INDEX IF NOT EXISTS idx_departments_leader_id ON departments(leader_id)",
		}
	case "documents":
		indexes = []string{
			"CREATE INDEX IF NOT EXISTS idx_documents_case_id ON documents(case_id)",
			"CREATE INDEX IF NOT EXISTS idx_documents_client_id ON documents(client_id)",
			"CREATE INDEX IF NOT EXISTS idx_documents_uploader_id ON documents(uploader_id)",
			"CREATE INDEX IF NOT EXISTS idx_documents_document_type ON documents(document_type)",
			"CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status)",
		}
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("❌ 创建索引失败: %v", err)
		}
	}
}

func addMissingColumns(db *sql.DB) {
	columnsToAdd := []struct {
		table    string
		column   string
		sql      string
	}{
		{"users", "real_name", "ALTER TABLE users ADD COLUMN IF NOT EXISTS real_name VARCHAR(50)"},
		{"users", "last_login_at", "ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE"},
		{"users", "last_login_ip", "ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_ip VARCHAR(45)"},
		{"users", "role_id", "ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id BIGINT"},
		{"users", "department_id", "ALTER TABLE users ADD COLUMN IF NOT EXISTS department_id BIGINT"},
		{"users", "remark", "ALTER TABLE users ADD COLUMN IF NOT EXISTS remark TEXT"},
		{"clients", "client_name", "ALTER TABLE clients ADD COLUMN IF NOT EXISTS client_name VARCHAR(100)"},
		{"clients", "lawyer_id", "ALTER TABLE clients ADD COLUMN IF NOT EXISTS lawyer_id BIGINT"},
		{"clients", "remark", "ALTER TABLE clients ADD COLUMN IF NOT EXISTS remark TEXT"},
		{"cases", "case_no", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS case_no VARCHAR(50) UNIQUE"},
		{"cases", "case_name", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS case_name VARCHAR(200)"},
		{"cases", "assisting_lawyer_id", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS assisting_lawyer_id BIGINT"},
		{"cases", "contract_amount", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS contract_amount DECIMAL(12,2)"},
		{"cases", "opponent_info", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS opponent_info TEXT"},
		{"cases", "remark", "ALTER TABLE cases ADD COLUMN IF NOT EXISTS remark TEXT"},
	}

	for _, col := range columnsToAdd {
		if tableExists(db, col.table) && !columnExists(db, col.table, col.column) {
			fmt.Printf("📝 添加 %s.%s 字段...\n", col.table, col.column)
			if _, err := db.Exec(col.sql); err != nil {
				log.Printf("❌ 添加字段失败: %v", err)
			} else {
				fmt.Printf("✅ %s.%s 字段添加成功\n", col.table, col.column)
			}
		}
	}
}

func insertSampleData(db *sql.DB) {
	if tableExists(db, "lawyers") {
		fmt.Println("📝 插入律师示例数据...")
		insertSQL := `
			INSERT INTO lawyers (lawyer_name, phone, email, license_no, position, department, status) VALUES
				('张律师', '13800138001', 'zhang@lawfirm.com', 'LAW001', '高级合伙人', '民商事部', 'active'),
				('李律师', '13800138002', 'li@lawfirm.com', 'LAW002', '合伙人', '刑事部', 'active'),
				('王律师', '13800138003', 'wang@lawfirm.com', 'LAW003', '律师', '行政部', 'active')
			ON CONFLICT (license_no) DO NOTHING
		`
		if _, err := db.Exec(insertSQL); err != nil {
			log.Printf("❌ 插入律师数据失败: %v", err)
		} else {
			fmt.Println("✅ 律师示例数据插入成功")
		}
	}
}

func finalVerification(db *sql.DB) {
	fmt.Println("🔍 最终验证结果:")

	// 检查关键表
	criticalTables := []string{"users", "clients", "cases", "lawyers", "departments", "documents"}
	for _, tableName := range criticalTables {
		if tableExists(db, tableName) {
			count := getRecordCount(db, tableName)
			fmt.Printf("✅ %s 表存在 (%d 条记录)\n", tableName, count)
		} else {
			fmt.Printf("❌ %s 表不存在\n", tableName)
		}
	}

	// 检查关键字段
	keyFields := []struct {
		table string
		field string
	}{
		{"users", "real_name"},
		{"clients", "lawyer_id"},
		{"cases", "case_no"},
	}

	for _, field := range keyFields {
		if columnExists(db, field.table, field.field) {
			fmt.Printf("✅ %s.%s 字段存在\n", field.table, field.field)
		} else {
			fmt.Printf("❌ %s.%s 字段缺失\n", field.table, field.field)
		}
	}
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)"
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		log.Printf("检查表存在失败: %v", err)
		return false
	}
	return exists
}

func columnExists(db *sql.DB, tableName, columnName string) bool {
	var exists bool
	query := "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2)"
	err := db.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		log.Printf("检查字段存在失败: %v", err)
		return false
	}
	return exists
}

func getRecordCount(db *sql.DB, tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("获取记录数失败: %v", err)
		return 0
	}
	return count
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}