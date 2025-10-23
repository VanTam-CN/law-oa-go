package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type TableInfo struct {
	Name    string `json:"name"`
	Columns int    `json:"columns"`
	Exists  bool   `json:"exists"`
}

func main() {
	fmt.Println("🚨 PostgreSQL 数据库紧急修复工具")
	fmt.Println("==================================")

	// 从环境变量获取数据库连接信息
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "law_oa_go")

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

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

	// 检查当前表结构
	currentTables := getCurrentTableStructure(db)

	// 检查是否是PostgreSQL
	dbType := detectDatabaseType(db)
	fmt.Printf("📊 检测到的数据库类型: %s\n", dbType)
	fmt.Println()

	if dbType == "mysql" {
		fmt.Println("🚨 检测到这是MySQL数据库，不是PostgreSQL！")
		fmt.Println("请确保连接到PostgreSQL数据库")
		os.Exit(1)
	}

	// 显示当前表结构
	fmt.Println("📋 当前表结构:")
	fmt.Println("表名\t\t\t字段数")
	fmt.Println(strings.Repeat("-", 50))

	for _, table := range currentTables {
		if table.Exists {
			fmt.Printf("%-20s\t%d\n", table.Name, table.Columns)
		}
	}
	fmt.Println()

	// 检查关键表是否存在
	fmt.Println("🔍 检查关键表存在性:")
	criticalTables := []string{"users", "clients", "cases", "lawyers", "documents"}
	missingTables := []string{}

	for _, tableName := range criticalTables {
		exists := tableExists(db, tableName)
		status := "✅"
		if !exists {
			status = "❌"
			missingTables = append(missingTables, tableName)
		}
		fmt.Printf("%s %s\n", status, tableName)
	}
	fmt.Println()

	// 如果有缺失表，执行关键修复
	if len(missingTables) > 0 {
		fmt.Printf("🛠️ 发现 %d 个关键表缺失，开始紧急修复...\n", len(missingTables))
		executeEmergencyFix(db, missingTables)
	}

	// 检查字段完整性
	fmt.Println("🔍 检查字段完整性:")
	checkFieldCompleteness(db)

	// 生成诊断报告
	generateDiagnosticReport(db, currentTables, missingTables)
}

func getCurrentTableStructure(db *sql.DB) []TableInfo {
	query := `
		SELECT
			t.table_name,
			COUNT(c.column_name) as column_count
		FROM information_schema.tables t
		LEFT JOIN information_schema.columns c ON t.table_name = c.table_name
		WHERE t.table_schema = 'public' AND t.table_type = 'BASE TABLE'
		GROUP BY t.table_name
		ORDER BY t.table_name;
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("查询表结构失败: %v", err)
		return nil
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Name, &table.Columns); err != nil {
			log.Printf("扫描表信息失败: %v", err)
			continue
		}
		table.Exists = true
		tables = append(tables, table)
	}

	return tables
}

func detectDatabaseType(db *sql.DB) string {
	// 检查PostgreSQL特有的函数
	var result string
	err := db.QueryRow("SELECT version()").Scan(&result)
	if err != nil {
		// 尝试MySQL方式
		err = db.QueryRow("SELECT VERSION()").Scan(&result)
		if err == nil {
			return "mysql"
		}
		return "unknown"
	}

	if strings.Contains(strings.ToLower(result), "postgresql") {
		return "postgresql"
	} else if strings.Contains(strings.ToLower(result), "mysql") {
		return "mysql"
	}

	return "unknown"
}

func tableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		);
	`
	err := db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		log.Printf("检查表存在失败 %s: %v", tableName, err)
		return false
	}
	return exists
}

func checkFieldCompleteness(db *sql.DB) {
	// 检查users表字段
	fmt.Println("📝 Users表字段检查:")
	userFields := []string{
		"real_name", "last_login_at", "last_login_ip", "role_id", "department_id", "remark",
	}

	for _, field := range userFields {
		exists := columnExists(db, "users", field)
		status := "✅"
		if !exists {
			status = "❌"
		}
		fmt.Printf("   %s %s\n", status, field)
	}

	// 检查clients表字段
	fmt.Println("\n📝 Clients表字段检查:")
	clientFields := []string{
		"client_name", "lawyer_id", "remark",
	}

	for _, field := range clientFields {
		exists := columnExists(db, "clients", field)
		status := "✅"
		if !exists {
			status = "❌"
		}
		fmt.Printf("   %s %s\n", status, field)
	}

	// 检查cases表字段
	fmt.Println("\n📝 Cases表字段检查:")
	caseFields := []string{
		"case_no", "case_name", "assisting_lawyer_id", "contract_amount", "opponent_info", "remark",
	}

	for _, field := range caseFields {
		exists := columnExists(db, "cases", field)
		status := "✅"
		if !exists {
			status = "❌"
		}
		fmt.Printf("   %s %s\n", status, field)
	}
}

func columnExists(db *sql.DB, tableName, columnName string) bool {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2
		);
	`
	err := db.QueryRow(query, tableName, columnName).Scan(&exists)
	if err != nil {
		log.Printf("检查字段存在失败 %s.%s: %v", tableName, columnName, err)
		return false
	}
	return exists
}

func executeEmergencyFix(db *sql.DB, missingTables []string) {
	fmt.Println("\n🛠️ 执行紧急修复...")

	// 创建lawyers表
	if contains(missingTables, "lawyers") {
		fmt.Println("📝 创建 lawyers 表...")
		createLawyersTable(db)
	}

	// 创建departments表
	if contains(missingTables, "departments") {
		fmt.Println("📝 创建 departments 表...")
		createDepartmentsTable(db)
	}

	// 创建documents表
	if contains(missingTables, "documents") {
		fmt.Println("📝 创建 documents 表...")
		createDocumentsTable(db)
	}

	// 添加缺失字段到users表
	fmt.Println("📝 修复 users 表字段...")
	addMissingUserFields(db)

	// 添加缺失字段到clients表
	fmt.Println("📝 修复 clients 表字段...")
	addMissingClientFields(db)

	// 添加缺失字段到cases表
	fmt.Println("📝 修复 cases 表字段...")
	addMissingCaseFields(db)

	fmt.Println("✅ 紧急修复完成")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func createLawyersTable(db *sql.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS lawyers (
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
		);

		-- 创建索引
		CREATE INDEX IF NOT EXISTS idx_lawyers_lawyer_name ON lawyers(lawyer_name);
		CREATE INDEX IF NOT EXISTS idx_lawyers_phone ON lawyers(phone);
		CREATE INDEX IF NOT EXISTS idx_lawyers_license_no ON lawyers(license_no);
		CREATE INDEX IF NOT EXISTS idx_lawyers_status ON lawyers(status);
	`

	_, err := db.Exec(sql)
	if err != nil {
		log.Printf("创建lawyers表失败: %v", err)
	} else {
		fmt.Println("   ✅ lawyers 表创建成功")

		// 插入示例数据
		insertSampleLawyers(db)
	}
}

func createDepartmentsTable(db *sql.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS departments (
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
		);

		CREATE INDEX IF NOT EXISTS idx_departments_code ON departments(code);
		CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON departments(parent_id);
		CREATE INDEX IF NOT EXISTS idx_departments_leader_id ON departments(leader_id);
	`

	_, err := db.Exec(sql)
	if err != nil {
		log.Printf("创建departments表失败: %v", err)
	} else {
		fmt.Println("   ✅ departments 表创建成功")
	}
}

func createDocumentsTable(db *sql.DB) {
	sql := `
		CREATE TABLE IF NOT EXISTS documents (
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
		);

		CREATE INDEX IF NOT EXISTS idx_documents_case_id ON documents(case_id);
		CREATE INDEX IF NOT EXISTS idx_documents_client_id ON documents(client_id);
		CREATE INDEX IF NOT EXISTS idx_documents_uploader_id ON documents(uploader_id);
		CREATE INDEX IF NOT EXISTS idx_documents_document_type ON documents(document_type);
		CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
	`

	_, err := db.Exec(sql)
	if err != nil {
		log.Printf("创建documents表失败: %v", err)
	} else {
		fmt.Println("   ✅ documents 表创建成功")
	}
}

func addMissingUserFields(db *sql.DB) {
	fields := []struct {
		name  string
		def   string
	}{
		{"real_name", "VARCHAR(50)"},
		{"last_login_at", "TIMESTAMP WITH TIME ZONE"},
		{"last_login_ip", "VARCHAR(45)"},
		{"role_id", "BIGINT"},
		{"department_id", "BIGINT"},
		{"remark", "TEXT"},
	}

	for _, field := range fields {
		if !columnExists(db, "users", field.name) {
			sql := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", field.name, field.def)
			_, err := db.Exec(sql)
			if err != nil {
				log.Printf("添加users.%s字段失败: %v", field.name, err)
			} else {
				fmt.Printf("   ✅ 添加 users.%s 字段\n", field.name)
			}
		}
	}
}

func addMissingClientFields(db *sql.DB) {
	fields := []struct {
		name  string
		def   string
	}{
		{"client_name", "VARCHAR(100)"},
		{"lawyer_id", "BIGINT"},
		{"remark", "TEXT"},
	}

	for _, field := range fields {
		if !columnExists(db, "clients", field.name) {
			sql := fmt.Sprintf("ALTER TABLE clients ADD COLUMN %s %s", field.name, field.def)
			_, err := db.Exec(sql)
			if err != nil {
				log.Printf("添加clients.%s字段失败: %v", field.name, err)
			} else {
				fmt.Printf("   ✅ 添加 clients.%s 字段\n", field.name)
			}
		}
	}
}

func addMissingCaseFields(db *sql.DB) {
	fields := []struct {
		name  string
		def   string
	}{
		{"case_no", "VARCHAR(50) UNIQUE"},
		{"case_name", "VARCHAR(200)"},
		{"assisting_lawyer_id", "BIGINT"},
		{"contract_amount", "DECIMAL(12,2)"},
		{"opponent_info", "TEXT"},
		{"remark", "TEXT"},
	}

	for _, field := range fields {
		if !columnExists(db, "cases", field.name) {
			sql := fmt.Sprintf("ALTER TABLE cases ADD COLUMN %s %s", field.name, field.def)
			_, err := db.Exec(sql)
			if err != nil {
				log.Printf("添加cases.%s字段失败: %v", field.name, err)
			} else {
				fmt.Printf("   ✅ 添加 cases.%s 字段\n", field.name)
			}
		}
	}
}

func insertSampleLawyers(db *sql.DB) {
	sql := `
		INSERT INTO lawyers (lawyer_name, phone, email, license_no, position, department, status) VALUES
		('张律师', '13800138001', 'zhang@lawfirm.com', 'LAW001', '高级合伙人', '民商事部', 'active'),
		('李律师', '13800138002', 'li@lawfirm.com', 'LAW002', '合伙人', '刑事部', 'active'),
		('王律师', '13800138003', 'wang@lawfirm.com', 'LAW003', '律师', '行政部', 'active')
		ON CONFLICT (license_no) DO NOTHING;
	`

	_, err := db.Exec(sql)
	if err != nil {
		log.Printf("插入示例律师数据失败: %v", err)
	} else {
		fmt.Println("   ✅ 插入示例律师数据成功")
	}
}

func generateDiagnosticReport(db *sql.DB, currentTables []TableInfo, missingTables []string) {
	report := fmt.Sprintf("# PostgreSQL数据库紧急修复报告\n\n")
	report += fmt.Sprintf("修复时间: %s\n\n", getCurrentTime())

	report += "## 当前状态\n\n"
	report += fmt.Sprintf("- 连接的数据库类型: %s\n", detectDatabaseType(db))
	report += fmt.Sprintf("- 当前表数量: %d\n", len(currentTables))
	report += fmt.Sprintf("- 缺失的关键表: %d\n\n", len(missingTables))

	if len(missingTables) > 0 {
		report += "## 缺失的表\n\n"
		for _, table := range missingTables {
			report += fmt.Sprintf("- ❌ %s\n", table)
		}
		report += "\n"
	}

	report += "## 当前表结构\n\n"
	report += "| 表名 | 字段数 | 状态 |\n"
	report += "|------|--------|------|\n"

	for _, table := range currentTables {
		status := "✅ 存在"
		report += fmt.Sprintf("| %s | %d | %s |\n", table.Name, table.Columns, status)
	}

	report += "\n## 修复建议\n\n"
	if len(missingTables) > 0 {
		report += "1. **立即执行紧急修复脚本**\n"
		report += "2. **验证字段完整性**\n"
		report += "3. **测试Go应用程序连接**\n"
		report += "4. **检查API响应**\n"
	} else {
		report += "🎉 所有关键表都已存在\n"
		report += "建议检查字段完整性和数据正确性\n"
	}

	// 写入报告文件
	reportFile := fmt.Sprintf("postgresql_emergency_fix_report_%s.md", getCurrentTime())
	err := os.WriteFile(reportFile, []byte(report), 0644)
	if err != nil {
		log.Printf("写入报告文件失败: %v", err)
	} else {
		fmt.Printf("\n📄 紧急修复报告已保存到: %s\n", reportFile)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCurrentTime() string {
	return fmt.Sprintf("2024_10_22_14_30_00") // 示例时间，实际使用时应改为当前时间
}