package main

import (
	"context"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// 数据库连接
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	fmt.Println("🧪 测试冲突检测逻辑...")

	// 测试场景1: 张伟律师为字节跳动创建案件，对手方是腾讯
	// 应该检测到张伟已经代理了阿里巴巴，存在行业竞争冲突
	testConflictDetection(db, "57", 45, []string{"腾讯"}, "张伟律师 vs 字节跳动 vs 腾讯")

	// 测试场景2: 陈浩律师为朱丽倩创建案件
	// 应该检测到陈浩已经代理了刘德华，存在法律对立冲突
	testConflictDetection(db, "64", 48, []string{}, "陈浩律师 vs 朱丽倩")

	// 测试场景3: 张伟律师为宝能集团创建案件，对手方是万科
	// 应该检测到张伟已经代理了万科，存在股权纠纷冲突
	testConflictDetection(db, "61", 45, []string{"万科"}, "张伟律师 vs 宝能集团 vs 万科")
}

func testConflictDetection(db *gorm.DB, clientID string, lawyerID uint, otherParties []string, scenario string) {
	fmt.Printf("\n🔍 测试场景: %s\n", scenario)
	fmt.Printf("   客户ID: %s, 律师ID: %d, 对方当事人: %v\n", clientID, lawyerID, otherParties)

	ctx := context.Background()

	// 转换clientID为uint
	var clientIDUint uint
	if _, err := fmt.Sscanf(clientID, "%d", &clientIDUint); err != nil {
		fmt.Printf("❌ 客户ID格式错误: %s\n", clientID)
		return
	}

	// 查询同一律师代理的其他案件
	query := `
		SELECT
			c.id as case_id,
			c.title as case_name,
			c.case_type,
			c.description,
			cl.name as client_name,
			cl.type as client_type,
			u.name as lawyer_name,
			c.created_at,
			c.lawyer_id,
			c.client_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.lawyer_id = ? AND c.client_id != ?
		AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`

	fmt.Printf("📋 执行查询: 律师ID=%d, 排除客户ID=%d\n", lawyerID, clientIDUint)

	rows, err := db.WithContext(ctx).Raw(query, lawyerID, clientIDUint).Rows()
	if err != nil {
		fmt.Printf("❌ 查询失败: %v\n", err)
		return
	}
	defer rows.Close()

	conflictCount := 0
	for rows.Next() {
		var caseID, lawyerIDFound, clientIDFound uint
		var caseName, caseType, description, clientName, clientType, lawyerName string
		var createdAt string

		err := rows.Scan(
			&caseID,
			&caseName,
			&caseType,
			&description,
			&clientName,
			&clientType,
			&lawyerName,
			&createdAt,
			&lawyerIDFound,
			&clientIDFound,
		)
		if err != nil {
			fmt.Printf("⚠️ 扫描数据失败: %v\n", err)
			continue
		}

		conflictCount++
		fmt.Printf("   📁 冲突案件 %d:\n", conflictCount)
		fmt.Printf("      案件ID: %d\n", caseID)
		fmt.Printf("      案件名称: %s\n", caseName)
		fmt.Printf("      委托人: %s (ID: %d)\n", clientName, clientIDFound)
		fmt.Printf("      律师: %s (ID: %d)\n", lawyerName, lawyerIDFound)
		fmt.Printf("      案件类型: %s\n", caseType)
		fmt.Printf("      创建时间: %s\n", createdAt)

		// 分析冲突类型
		analyzeConflictType(clientName, otherParties, caseName)
		fmt.Println("      ---")
	}

	if conflictCount == 0 {
		fmt.Println("   ✅ 未发现冲突案件")
	} else {
		fmt.Printf("   ⚠️ 发现 %d 个潜在冲突案件\n", conflictCount)
	}

	// 如果有对方当事人，也查询相关案件
	if len(otherParties) > 0 {
		fmt.Printf("📋 检查对方当事人冲突: %v\n", otherParties)
		for _, party := range otherParties {
			checkOpponentConflicts(db, ctx, party, lawyerID)
		}
	}
}

func analyzeConflictType(clientName string, otherParties []string, caseName string) {
	fmt.Printf("      🔍 冲突分析:\n")

	// 检查是否为直接竞争对手
	competitors := map[string][]string{
		"阿里巴巴": {"腾讯", "字节跳动", "京东", "拼多多"},
		"腾讯":   {"阿里巴巴", "字节跳动", "快手"},
		"字节跳动": {"阿里巴巴", "腾讯", "快手"},
		"万科":   {"宝能"},
		"宝能":   {"万科"},
		"刘德华":  {"朱丽倩"},
		"朱丽倩":  {"刘德华"},
	}

	for _, party := range otherParties {
		if rivals, exists := competitors[clientName]; exists {
			for _, rival := range rivals {
				if party == rival {
					fmt.Printf("         🚨 检测到直接竞争冲突: %s vs %s\n", clientName, party)
					fmt.Printf("         🔍 风险等级: CRITICAL\n")
					return
				}
			}
		}
	}

	// 检查案件名称中的冲突
	if contains(caseName, "离婚") && (contains(clientName, "刘德华") || contains(clientName, "朱丽倩")) {
		fmt.Printf("         🚨 检测到法律对立冲突: 婚姻纠纷\n")
		fmt.Printf("         🔍 风险等级: CRITICAL\n")
	} else if contains(caseName, "股权") || contains(caseName, "收购") {
		fmt.Printf("         ⚠️ 检测到股权纠纷冲突\n")
		fmt.Printf("         🔍 风险等级: HIGH\n")
	} else {
		fmt.Printf("         ⚠️ 检测到一般代理冲突\n")
		fmt.Printf("         🔍 风险等级: MEDIUM\n")
	}
}

func checkOpponentConflicts(db *gorm.DB, ctx context.Context, party string, lawyerID uint) {
	// 查询包含对方当事人名称的案件
	partyQuery := `
		SELECT
			c.id as case_id,
			c.title as case_name,
			cl.name as client_name,
			u.name as lawyer_name,
			c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.deleted_at IS NULL
		AND (c.title ILIKE ? OR c.description ILIKE ? OR cl.name ILIKE ?)
		ORDER BY c.created_at DESC
		LIMIT 10
	`

	rows, err := db.WithContext(ctx).Raw(partyQuery,
		"%"+party+"%", "%"+party+"%", "%"+party+"%").Rows()
	if err != nil {
		fmt.Printf("   ❌ 查询对方当事人失败: %v\n", err)
		return
	}
	defer rows.Close()

	opponentCount := 0
	for rows.Next() {
		var caseID, foundLawyerID uint
		var caseName, clientName, lawyerName string

		err := rows.Scan(&caseID, &caseName, &clientName, &lawyerName, &foundLawyerID)
		if err != nil {
			continue
		}

		opponentCount++
		fmt.Printf("   📁 相关案件 %d: %s (委托人: %s, 律师: %s)\n",
			opponentCount, caseName, clientName, lawyerName)

		if foundLawyerID == lawyerID {
			fmt.Printf("      🚨 发现同一律师代理冲突！\n")
		}
	}

	if opponentCount == 0 {
		fmt.Printf("   ✅ 未发现与 '%s' 相关的案件\n", party)
	}
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					findInString(str, substr))))
}

func findInString(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
