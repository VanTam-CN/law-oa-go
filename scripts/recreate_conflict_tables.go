//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔧 重新创建利益冲突检测表")
	fmt.Println("==========================")

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

	// 删除旧的表并重新创建
	fmt.Println("\n🗑️ 删除旧的冲突检测表...")
	dropOldTables(db)

	// 重新创建所有表
	fmt.Println("\n📋 重新创建冲突检测表...")
	recreateAllTables(db)

	fmt.Println("\n✅ 利益冲突检测表重新创建完成！")
	fmt.Println("==================================")
}

func dropOldTables(db *sql.DB) {
	tables := []string{
		"conflict_check_records",
		"conflict_rules",
		"conflict_stats",
		"conflict_cases",
		"client_relations",
		"mcp_standards",
	}

	for _, table := range tables {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		_, err := db.Exec(dropSQL)
		if err != nil {
			log.Printf("删除表 %s 失败: %v", table, err)
		} else {
			fmt.Printf("✅ 删除表 %s 成功\n", table)
		}
	}
}

func recreateAllTables(db *sql.DB) {
	// 创建 conflict_check_records 表
	fmt.Println("📋 创建 conflict_check_records 表...")
	createConflictCheckRecordsTable(db)

	// 创建 conflict_rules 表
	fmt.Println("📋 创建 conflict_rules 表...")
	createConflictRulesTable(db)

	// 创建 conflict_stats 表
	fmt.Println("📋 创建 conflict_stats 表...")
	createConflictStatsTable(db)

	// 创建 conflict_cases 表
	fmt.Println("📋 创建 conflict_cases 表...")
	createConflictCasesTable(db)

	// 创建 client_relations 表
	fmt.Println("📋 创建 client_relations 表...")
	createClientRelationsTable(db)

	// 创建 mcp_standards 表
	fmt.Println("📋 创建 mcp_standards 表...")
	createMCPStandardsTable(db)

	// 创建默认冲突规则
	fmt.Println("📋 创建默认冲突规则...")
	createDefaultConflictRules(db)
}

func createConflictCheckRecordsTable(db *sql.DB) {
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

	fmt.Println("✅ conflict_check_records 表创建成功")
}

func createConflictRulesTable(db *sql.DB) {
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

	fmt.Println("✅ conflict_rules 表创建成功")
}

func createConflictStatsTable(db *sql.DB) {
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

	fmt.Println("✅ conflict_stats 表创建成功")
}

func createConflictCasesTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE conflict_cases (
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
		CREATE TABLE client_relations (
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
		CREATE TABLE mcp_standards (
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

func createDefaultConflictRules(db *sql.DB) {
	defaultRules := []struct {
		id          string
		name        string
		typeStr     string
		category    string
		description string
		priority    int
		conditions  string
	}{
		{
			id:          "rule_001",
			name:        "直接冲突检测",
			typeStr:     "direct",
			category:    "legal",
			description: "检测律师是否已经代理对方当事人",
			priority:    1,
			conditions:  `{"checkDirect": true, "severity": "critical"}`,
		},
		{
			id:          "rule_002",
			name:        "关联公司检测",
			typeStr:     "corporate",
			category:    "business",
			description: "检测客户企业的关联公司",
			priority:    2,
			conditions:  `{"checkCorporate": true, "severity": "high"}`,
		},
		{
			id:          "rule_003",
			name:        "行业竞争检测",
			typeStr:     "industry",
			category:    "business",
			description: "检测同行业主要竞争对手",
			priority:    3,
			conditions:  `{"checkIndustry": true, "severity": "medium"}`,
		},
	}

	for _, rule := range defaultRules {
		// 检查规则是否已存在
		var count int
		checkSQL := "SELECT COUNT(*) FROM conflict_rules WHERE id = $1"
		err := db.QueryRow(checkSQL, rule.id).Scan(&count)
		if err != nil {
			log.Printf("检查规则失败: %v", err)
			continue
		}

		if count > 0 {
			fmt.Printf("📋 规则 '%s' 已存在，跳过\n", rule.name)
			continue
		}

		// 执行插入
		insertSQL := `INSERT INTO conflict_rules (id, name, type, category, description, priority, conditions) VALUES ($1, $2, $3, $4, $5, $6, $7)`
		_, err = db.Exec(insertSQL, rule.id, rule.name, rule.typeStr, rule.category, rule.description, rule.priority, rule.conditions)
		if err != nil {
			log.Printf("创建默认规则 '%s' 失败: %v", rule.name, err)
		} else {
			fmt.Printf("✅ 创建规则: %s\n", rule.name)
		}
	}
}