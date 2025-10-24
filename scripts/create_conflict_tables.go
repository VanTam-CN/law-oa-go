package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔧 创建利益冲突检测相关表")
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

	// 创建冲突检测记录表
	fmt.Println("\n📋 创建冲突检测记录表...")
	createConflictCheckRecordsTable(db)

	// 创建冲突规则表
	fmt.Println("\n📋 创建冲突规则表...")
	createConflictRulesTable(db)

	// 创建冲突检测统计表
	fmt.Println("\n📋 创建冲突检测统计表...")
	createConflictStatsTable(db)

	fmt.Println("\n✅ 利益冲突检测相关表创建完成！")
	fmt.Println("=====================================")
}

func createConflictCheckRecordsTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conflict_check_records (
			id BIGSERIAL PRIMARY KEY,
			client_id BIGINT NOT NULL,
			client_name VARCHAR(255) NOT NULL,
			client_type VARCHAR(50) NOT NULL,
			lawyer_id BIGINT NOT NULL,
			case_name VARCHAR(500) NOT NULL,
			case_type VARCHAR(100) NOT NULL,
			other_parties TEXT,
			search_years INTEGER DEFAULT 3,
			include_corporate_relations BOOLEAN DEFAULT true,
			search_depth VARCHAR(20) DEFAULT 'STANDARD',
			has_conflict BOOLEAN DEFAULT false,
			conflict_cases TEXT,
			risk_assessment JSONB,
			check_statistics JSONB,
			request_time TIMESTAMP WITH TIME ZONE NOT NULL,
			check_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("创建冲突检测记录表失败: %v", err)
		return
	}

	// 创建索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_client_id ON conflict_check_records(client_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_lawyer_id ON conflict_check_records(lawyer_id)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_has_conflict ON conflict_check_records(has_conflict)",
		"CREATE INDEX IF NOT EXISTS idx_conflict_check_records_request_time ON conflict_check_records(request_time)",
	}

	for _, indexSQL := range indexes {
		_, err := db.Exec(indexSQL)
		if err != nil {
			log.Printf("创建索引失败: %v", err)
		}
	}

	fmt.Println("✅ 冲突检测记录表创建成功")
}

func createConflictRulesTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conflict_rules (
			id BIGSERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(100) NOT NULL,
			description TEXT,
			priority INTEGER DEFAULT 1,
			conditions JSONB NOT NULL,
			is_active BOOLEAN DEFAULT true,
			created_by BIGINT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Printf("创建冲突规则表失败: %v", err)
		return
	}

	// 创建默认冲突规则
	defaultRules := []struct {
		name        string
		typeStr     string
		description string
		priority    int
		conditions  string
	}{
		{
			name:        "直接冲突检测",
			typeStr:     "direct",
			description: "检测律师是否已经代理对方当事人",
			priority:    1,
			conditions:  `{"checkDirect": true, "severity": "critical"}`,
		},
		{
			name:        "关联公司检测",
			typeStr:     "corporate",
			description: "检测客户企业的关联公司",
			priority:    2,
			conditions:  `{"checkCorporate": true, "severity": "high"}`,
		},
		{
			name:        "行业竞争检测",
			typeStr:     "industry",
			description: "检测同行业主要竞争对手",
			priority:    3,
			conditions:  `{"checkIndustry": true, "severity": "medium"}`,
		},
	}

	for _, rule := range defaultRules {
		// 检查规则是否已存在
		var count int
		checkSQL := "SELECT COUNT(*) FROM conflict_rules WHERE name = $1"
		err := db.QueryRow(checkSQL, rule.name).Scan(&count)
		if err != nil {
			log.Printf("检查规则失败: %v", err)
			continue
		}

		if count > 0 {
			fmt.Printf("📋 规则 '%s' 已存在，跳过\n", rule.name)
			continue
		}

		// 执行插入
		insertSQL := `INSERT INTO conflict_rules (name, type, description, priority, conditions) VALUES ($1, $2, $3, $4, $5)`
		_, err = db.Exec(insertSQL, rule.name, rule.typeStr, rule.description, rule.priority, rule.conditions)
		if err != nil {
			log.Printf("创建默认规则 '%s' 失败: %v", rule.name, err)
		} else {
			fmt.Printf("✅ 创建规则: %s\n", rule.name)
		}
	}

	fmt.Println("✅ 冲突规则表创建成功")
}

func createConflictStatsTable(db *sql.DB) {
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS conflict_stats (
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
		log.Printf("创建冲突统计表失败: %v", err)
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

	fmt.Println("✅ 冲突统计表创建成功")
}
