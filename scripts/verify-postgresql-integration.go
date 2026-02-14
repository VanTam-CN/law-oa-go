//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type VerificationResult struct {
	TableName    string `json:"table_name"`
	ColumnCount  int    `json:"column_count"`
	RecordCount  int    `json:"record_count"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

func main() {
	// 从环境变量获取数据库连接信息
	dbHost := getEnv("PG_HOST", "localhost")
	dbPort := getEnv("PG_PORT", "5432")
	dbUser := getEnv("PG_USER", "law_oa_user")
	dbPassword := getEnv("PG_PASSWORD", "")
	dbName := getEnv("PG_DATABASE", "law_oa_go")

	// 构建连接字符串
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// 连接数据库
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库ping失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("🔍 开始验证PostgreSQL数据库结构...")

	// 定义需要验证的表及其预期字段数
	tablesToVerify := map[string]int{
		"users":                15, // username, name, real_name, email, password, role, role_id, phone, avatar, status, last_login_at, last_login_ip, department_id, remark
		"roles":                9,  // id, role_name, role_key, sort, status, remark, created_at, updated_at, deleted_at
		"permissions":          12, // id, permission_name, permission_key, parent_id, path, component, icon, sort, menu_type, status, remark, created_at, updated_at
		"clients":              16, // id, name, client_name, type, email, phone, address, company, id_card, industry, contact_person, contact_phone, source, notes, lawyer_id, remark, status
		"lawyers":              12, // id, lawyer_name, phone, email, license_no, position, department, specialty, status, remark, created_at, updated_at, deleted_at
		"cases":                25, // 包含所有新增字段
		"case_progress":        10, // id, case_id, stage, title, description, status, due_date, completed_at, created_by, created_at, updated_at
		"case_documents":       12, // id, case_id, name, type, file_path, file_size, mime_type, description, uploaded_by, created_at, updated_at, deleted_at
		"documents":           20, // 完整的文档表字段
		"document_versions":   11, // id, document_id, version_no, file_path, file_hash, file_size, uploader_id, change_log, upload_time, is_current, remark
		"document_permissions": 8, // id, document_id, user_id, permission_type, created_at, created_by, expires_at, status, remark
		"document_categories": 11, // id, category_name, category_key, parent_id, description, sort, status, icon, color, document_count, remark
		"system_configs":      9,  // id, config_key, config_value, config_type, description, is_system, sort, status, created_at, updated_at
		"operation_logs":      11, // id, user_id, username, operation, method, path, params, ip, user_agent, status, error_message, execution_time, created_at
		"financial_records":    13, // id, case_id, client_id, type, category, amount, currency, description, transaction_date, status, payment_method, invoice_number, created_by
		"notifications":       10, // id, user_id, type, title, content, related_id, related_type, is_read, priority, expires_at, created_at
		"schedules":          13, // id, user_id, title, description, start_time, end_time, type, related_id, related_type, location, participants, reminder_time, is_all_day, status
		"law_entities":        12, // id, entity_name, entity_type, entity_subtype, id_card, license_no, address, contact_info, risk_level, status, remark
		"law_entity_aliases": 7,  // id, entity_id, alias_name, alias_type, status, remark, created_at, updated_at
		"law_entity_relations": 10, // id, source_entity_id, target_entity_id, relation_type, relation_desc, start_date, end_date, status, remark
		"conflict_check_records": 25, // 包含所有MySQL和PostgreSQL字段
	}

	results := []VerificationResult{}

	fmt.Println("\n📋 验证表结构...")

	for tableName, expectedColumns := range tablesToVerify {
		result := verifyTable(db, tableName, expectedColumns)
		results = append(results, result)
		status := "❌"
		if result.Status == "PASS" {
			status = "✅"
		}
		fmt.Printf("%s 表: %-25s 字段: %2d/%2d 记录: %6d %s\n",
			status, tableName, result.ColumnCount, expectedColumns, result.RecordCount, result.Message)
	}

	// 生成验证报告
	generateReport(results)

	fmt.Println("\n📊 验证统计:")
	passCount := 0
	warningCount := 0
	failCount := 0

	for _, result := range results {
		switch result.Status {
		case "PASS":
			passCount++
		case "WARNING":
			warningCount++
		case "FAIL":
			failCount++
		}
	}

	fmt.Printf("✅ 通过: %d\n", passCount)
	fmt.Printf("⚠️  警告: %d\n", warningCount)
	fmt.Printf("❌ 失败: %d\n", failCount)

	// 检查关键表是否全部通过
	criticalTables := []string{"users", "clients", "cases", "roles", "permissions"}
	criticalFail := false

	for _, table := range criticalTables {
		for _, result := range results {
			if result.TableName == table && result.Status != "PASS" {
				criticalFail = true
				break
			}
		}
		if criticalFail {
			break
		}
	}

	if criticalFail {
		fmt.Println("\n❌ 关键表验证失败，请检查数据库结构！")
		os.Exit(1)
	} else if failCount > 0 {
		fmt.Println("\n⚠️  部分表验证失败，但关键表正常")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ 数据库验证通过！PostgreSQL数据库结构完整！")
	}
}

func verifyTable(db *sql.DB, tableName string, expectedColumns int) VerificationResult {
	result := VerificationResult{
		TableName: tableName,
		Status:    "FAIL",
	}

	// 检查表是否存在
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`, tableName).Scan(&tableExists)

	if err != nil {
		result.Message = fmt.Sprintf("检查表存在失败: %v", err)
		return result
	}

	if !tableExists {
		result.Message = "表不存在"
		return result
	}

	// 获取字段数量
	var columnCount int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName).Scan(&columnCount)

	if err != nil {
		result.Message = fmt.Sprintf("获取字段数失败: %v", err)
		return result
	}

	// 获取记录数量
	var recordCount int
	err = db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&recordCount)
	if err != nil {
		result.Message = fmt.Sprintf("获取记录数失败: %v", err)
		return result
	}

	result.ColumnCount = columnCount
	result.RecordCount = recordCount

	// 判断状态
	if columnCount >= expectedColumns {
		result.Status = "PASS"
		result.Message = "表结构完整"
	} else if columnCount >= expectedColumns*2/3 { // 至少2/3的字段
		result.Status = "WARNING"
		result.Message = fmt.Sprintf("缺少 %d 个字段", expectedColumns-columnCount)
	} else {
		result.Status = "FAIL"
		result.Message = fmt.Sprintf("字段数量严重不足，预期 %d，实际 %d", expectedColumns, columnCount)
	}

	return result
}

func generateReport(results []VerificationResult) {
	report := fmt.Sprintf("# PostgreSQL数据库验证报告\n\n")
	report += fmt.Sprintf("验证时间: %s\n\n", getCurrentTime())

	report += "## 验证结果\n\n"
	report += "| 表名 | 状态 | 字段数 | 记录数 | 备注 |\n"
	report += "|------|------|--------|--------|------|\n"

	for _, result := range results {
		status := "❌"
		if result.Status == "PASS" {
			status = "✅"
		} else if result.Status == "WARNING" {
			status = "⚠️"
		}

		report += fmt.Sprintf("| %s | %s | %d | %d | %s |\n",
			result.TableName, status, result.ColumnCount, result.RecordCount, result.Message)
	}

	report += "\n## 统计信息\n\n"

	passCount := 0
	warningCount := 0
	failCount := 0

	for _, result := range results {
		switch result.Status {
		case "PASS":
			passCount++
		case "WARNING":
			warningCount++
		case "FAIL":
			failCount++
		}
	}

	report += fmt.Sprintf("- ✅ 通过: %d\n", passCount)
	report += fmt.Sprintf("- ⚠️  警告: %d\n", warningCount)
	report += fmt.Sprintf("- ❌ 失败: %d\n", failCount)

	report += "\n## 建议\n\n"

	if failCount > 0 {
		report += "- 请检查失败的表结构，确保所有字段都已正确创建\n"
	}

	if warningCount > 0 {
		report += "- 建议检查警告表，确保所有重要字段都已包含\n"
	}

	if passCount == len(results) {
		report += "- 🎉 数据库结构验证完全通过，可以开始使用PostgreSQL数据库！\n"
	}

	// 写入报告文件
	reportFile := fmt.Sprintf("postgresql_verification_report_%s.md", getCurrentTime())
	err := os.WriteFile(reportFile, []byte(report), 0644)
	if err != nil {
		log.Printf("写入报告文件失败: %v", err)
	} else {
		fmt.Printf("📄 验证报告已保存到: %s\n", reportFile)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCurrentTime() string {
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		2024, 10, 22, 14, 30, 0) // 示例时间，实际使用时应改为当前时间
}