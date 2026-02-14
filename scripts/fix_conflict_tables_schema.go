//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔧 修复利益冲突检测表结构")
	fmt.Println("========================")

	dsn := "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("数据库连接测试失败: %v", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 修复 conflict_check_records 表结构
	fmt.Println("\n📋 修复 conflict_check_records 表结构...")
	fixConflictCheckRecordsTable(db)

	// 修复 conflict_rules 表结构
	fmt.Println("\n📋 修复 conflict_rules 表结构...")
	fixConflictRulesTable(db)

	// 修复 conflict_stats 表结构
	fmt.Println("\n📋 修复 conflict_stats 表结构...")
	fixConflictStatsTable(db)

	// 创建其他缺失的表
	fmt.Println("\n📋 创建其他冲突检测相关表...")
	createOtherConflictTables(db)

	fmt.Println("\n✅ 利益冲突检测表结构修复完成！")
	fmt.Println("==================================")
}

func fixConflictCheckRecordsTable(db *sql.DB) {
	// 检查表是否存在
	var tableExists bool
	checkTableSQL := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'conflict_check_records')"
	err := db.QueryRow(checkTableSQL).Scan(&tableExists)
	if err != nil {
		log.Printf("检查表存在性失败: %v", err)
		return
	}

	if !tableExists {
		fmt.Println("📋 创建 conflict_check_records 表...")
		createTableSQL := `
			CREATE TABLE conflict_check_records (
				check_id VARCHAR(255) PRIMARY KEY,
				client_id VARCHAR(255) NOT NULL,
				client_name VARCHAR(255) NOT NULL,
				case_name VARCHAR(500) NOT NULL,
				case_type VARCHAR(100) NOT NULL,
				check_status VARCHAR(50) DEFAULT 'PROCESSING',
				has_conflict BOOLEAN DEFAULT false,
				risk_level VARCHAR(50) DEFAULT 'LOW',
				search_parameters JSONB,
				check_result JSONB,
				user_id BIGINT NOT NULL,
				duration BIGINT DEFAULT 0,
				check_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
		`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			log.Printf("创建 conflict_check_records 表失败: %v", err)
			return
		}
	} else {
		fmt.Println("📋 检查并修改现有表结构...")

		// 检查是否有 check_id 列
		var checkIdExists bool
		checkColumnSQL := "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'conflict_check_records' AND column_name = 'check_id')"
		err := db.QueryRow(checkColumnSQL).Scan(&checkIdExists)
		if err != nil {
			log.Printf("检查 check_id 列失败: %v", err)
			return
		}

		if !checkIdExists {
			// 添加 check_id 列
			alterSQL := "ALTER TABLE conflict_check_records ADD COLUMN check_id VARCHAR(255) PRIMARY KEY"
			_, err := db.Exec(alterSQL)
			if err != nil {
				log.Printf("添加 check_id 列失败: %v", err)
				return
			}
			fmt.Println("✅ 添加 check_id 列成功")
		}

		// 检查并添加其他缺失的列
		addMissingColumns(db, "conflict_check_records")
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_client_id ON conflict_check_records(client_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_user_id ON conflict_check_records(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_has_conflict ON conflict_check_records(has_conflict)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_check_time ON conflict_check_records(check_time)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_risk_level ON conflict_check_records(risk_level)",
	}

	for _, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}

	fmt.Println("✅ conflict_check_records 表结构修复成功")
}

func fixConflictRulesTable(db *sql.DB) {
	// 检查表是否存在
	var tableExists bool
	checkTableSQL := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'conflict_rules')"
	err := db.QueryRow(checkTableSQL).Scan(&tableExists)
	if err != nil {
		log.Printf("检查表存在性失败: %v", err)
		return
	}

	if !tableExists {
		createTableSQL := `
			CREATE TABLE conflict_rules (
				id VARCHAR(255) PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				type VARCHAR(100) NOT NULL,
				description TEXT,
				category VARCHAR(100) NOT NULL,
				conditions JSONB NOT NULL,
				actions TEXT[],
				priority INTEGER DEFAULT 1,
				active BOOLEAN DEFAULT true,
				version INTEGER DEFAULT 1,
				mcp_source VARCHAR(255),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
		`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			log.Printf("创建 conflict_rules 表失败: %v", err)
			return
		}
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_conflict_rules_type ON conflict_rules(type)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_rules_category ON conflict_rules(category)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_rules_active ON conflict_rules(active)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_rules_priority ON conflict_rules(priority)",
	}

	for _, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}

	fmt.Println("✅ conflict_rules 表结构修复成功")
}

func fixConflictStatsTable(db *sql.DB) {
	// 检查表是否存在
	var tableExists bool
	checkTableSQL := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'conflict_stats')"
	err := db.QueryRow(checkTableSQL).Scan(&tableExists)
	if err != nil {
		log.Printf("检查表存在性失败: %v", err)
		return
	}

	if !tableExists {
		createTableSQL := `
			CREATE TABLE conflict_stats (
				id BIGSERIAL PRIMARY KEY,
				total_checks INTEGER DEFAULT 0,
				high_risk_checks INTEGER DEFAULT 0,
				critical_risk_checks INTEGER DEFAULT 0,
				medium_risk_checks INTEGER DEFAULT 0,
				low_risk_checks INTEGER DEFAULT 0,
				average_duration FLOAT DEFAULT 0,
				last_check_time TIMESTAMP WITH TIME ZONE,
				last_client_id VARCHAR(255),
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);
		`

		_, err := db.Exec(createTableSQL)
		if err != nil {
			log.Printf("创建 conflict_stats 表失败: %v", err)
			return
		}

		// 插入初始统计记录
		insertSQL := `
			INSERT INTO conflict_stats (total_checks, high_risk_checks, critical_risk_checks, medium_risk_checks, low_risk_checks, average_duration)
			VALUES (0, 0, 0, 0, 0, 0.0)
			ON CONFLICT (id) DO NOTHING;
		`

		_, err = db.Exec(insertSQL)
		if err != nil {
			log.Printf("插入初始统计数据失败: %v", err)
		}
	}

	fmt.Println("✅ conflict_stats 表结构修复成功")
}

func createOtherConflictTables(db *sql.DB) {
	// 创建 conflict_cases 表
	fmt.Println("📋 创建 conflict_cases 表...")
	createConflictCasesTable(db)

	// 创建 client_relations 表
	fmt.Println("📋 创建 client_relations 表...")
	createClientRelationsTable(db)

	// 创建 mcp_standards 表
	fmt.Println("📋 创建 mcp_standards 表...")
	createMCPStandardsTable(db)
}

func createConflictCasesTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conflict_cases (
			id VARCHAR(255) PRIMARY KEY,
			check_id VARCHAR(255) NOT NULL,
			case_id VARCHAR(255),
			case_name VARCHAR(500),
			case_no VARCHAR(255),
			conflict_type VARCHAR(100),
			risk_level VARCHAR(50),
			description TEXT,
			case_status VARCHAR(50),
			client_id VARCHAR(255),
			opposing_parties TEXT[],
			conflict_details TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("创建 conflict_cases 表失败: %v", err)
		return
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_conflict_cases_check_id ON conflict_cases(check_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_cases_client_id ON conflict_cases(client_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_cases_risk_level ON conflict_cases(risk_level)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_cases_conflict_type ON conflict_cases(conflict_type)",
	}

	for _, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}

	fmt.Println("✅ conflict_cases 表创建成功")
}

func createClientRelationsTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS client_relations (
			id VARCHAR(255) PRIMARY KEY,
			client_id VARCHAR(255) NOT NULL,
			related_client_id VARCHAR(255) NOT NULL,
			relation_type VARCHAR(100),
			relation_detail TEXT,
			active BOOLEAN DEFAULT true,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("创建 client_relations 表失败: %v", err)
		return
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_client_relations_client_id ON client_relations(client_id)",
		"CREATE INDEX IF NOT EXISTS idx_client_relations_related_client_id ON client_relations(related_client_id)",
		"CREATE INDEX IF NOT EXISTS idx_client_relations_active ON client_relations(active)",
	}

	for _, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}

	fmt.Println("✅ client_relations 表创建成功")
}

func createMCPStandardsTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS mcp_standards (
			version VARCHAR(50) PRIMARY KEY,
			last_updated TIMESTAMP WITH TIME ZONE,
			standards JSONB,
			best_practices TEXT[],
			compliance TEXT[],
			risk_thresholds JSONB,
			active BOOLEAN DEFAULT true
		);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("创建 mcp_standards 表失败: %v", err)
		return
	}

	fmt.Println("✅ mcp_standards 表创建成功")
}

func addMissingColumns(db *sql.DB, tableName string) {
	// 定义需要检查的列
	columns := map[string]string{
		"check_status":       "VARCHAR(50) DEFAULT 'PROCESSING'",
		"risk_level":         "VARCHAR(50) DEFAULT 'LOW'",
		"search_parameters":  "JSONB",
		"check_result":       "JSONB",
		"user_id":            "BIGINT NOT NULL DEFAULT 0",
		"duration":           "BIGINT DEFAULT 0",
		"case_name":          "VARCHAR(500) NOT NULL DEFAULT ''",
		"case_type":          "VARCHAR(100) NOT NULL DEFAULT ''",
		"client_name":        "VARCHAR(255) NOT NULL DEFAULT ''",
		"client_id":          "VARCHAR(255) NOT NULL DEFAULT ''",
	}

	for columnName, columnDef := range columns {
		var columnExists bool
		checkColumnSQL := fmt.Sprintf("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = '%s' AND column_name = '%s')", tableName, columnName)
		err := db.QueryRow(checkColumnSQL).Scan(&columnExists)
		if err != nil {
			log.Printf("检查列 %s 失败: %v", columnName, err)
			continue
		}

		if !columnExists {
			alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnDef)
			_, err := db.Exec(alterSQL)
			if err != nil {
				log.Printf("添加列 %s 失败: %v", columnName, err)
			} else {
				fmt.Printf("✅ 添加列 %s 成功\n", columnName)
			}
		}
	}
}