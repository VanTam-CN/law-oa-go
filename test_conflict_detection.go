package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string
	Email string
	Role  string
}

type Client struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string
	Type     string
	LawyerID uint
}

type Case struct {
	ID          uint       `gorm:"primaryKey"`
	Title       string
	ClientID    uint
	LawyerID    uint
	CaseType    string
	Description string
	DeletedAt   *time.Time `gorm:"index"`
	CreatedAt   time.Time
}

func main() {
	// 数据库连接
	dsn := "host=localhost port=5432 user=law_oa_user password=1q2w#E$R dbname=law_oa_db sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 测试场景：张伟律师(ID:45)代理字节跳动(ID:57)，对手方是腾讯
	lawyerID := 45
	clientID := "57"
	otherParties := []string{"腾讯控股有限公司"}

	fmt.Printf("=== 测试冲突检测逻辑 ===\n")
	fmt.Printf("律师ID: %d (张伟)\n", lawyerID)
	fmt.Printf("当前客户ID: %s (字节跳动)\n", clientID)
	fmt.Printf("对方当事人: %v\n", otherParties)

	// 1. 查找同一律师代理的其他案件
	fmt.Println("\n=== 查找同一律师代理的其他案件 ===")

	// 转换clientID为数字
	clientIDUint, err := strconv.ParseUint(clientID, 10, 32)
	if err != nil {
		log.Fatal("转换客户ID失败:", err)
	}

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
			c.lawyer_id
		FROM cases c
		JOIN clients cl ON c.client_id = cl.id
		JOIN users u ON c.lawyer_id = u.id
		WHERE c.lawyer_id = ? AND c.client_id != ?
		AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT 50
	`

	rows, err := db.Raw(query, lawyerID, uint(clientIDUint)).Rows()
	if err != nil {
		log.Fatal("查询失败:", err)
	}
	defer rows.Close()

	fmt.Printf("找到同一律师代理的其他案件数量: ")
	count := 0
	for rows.Next() {
		var caseID uint
		var caseName, caseType, description, clientName, clientType, lawyerName string
		var foundLawyerID uint
		var createdAt time.Time

		if err := rows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &clientType, &lawyerName, &createdAt, &foundLawyerID); err != nil {
			continue
		}

		count++
		fmt.Printf("\n  %d. %s (客户: %s, 律师: %s)\n", count, caseName, clientName, lawyerName)

		// 检查是否与腾讯有冲突
		if containsCompany(caseName, "腾讯") || containsCompany(description, "腾讯") {
			fmt.Printf("     ⚠️ 发现与腾讯相关的案件!\n")
		}
	}

	if count == 0 {
		fmt.Println("0")
	}

	// 2. 查询包含对方当事人名称的案件
	fmt.Println("\n=== 查询包含腾讯的相关案件 ===")
	for _, party := range otherParties {
		fmt.Printf("搜索对方当事人: %s\n", party)

		partyQuery := `
			SELECT
				c.id as case_id,
				c.title as case_name,
				c.case_type,
				c.description,
				cl.name as client_name,
				cl.type as client_type,
				u.name as lawyer_name,
				c.created_at,
				c.lawyer_id
			FROM cases c
			JOIN clients cl ON c.client_id = cl.id
			JOIN users u ON c.lawyer_id = u.id
			WHERE c.deleted_at IS NULL
			AND (c.title ILIKE ? OR c.description ILIKE ?)
			ORDER BY c.created_at DESC
			LIMIT 20
		`

		partyRows, err := db.Raw(partyQuery, "%"+party+"%", "%"+party+"%").Rows()
		if err != nil {
			log.Printf("查询对方当事人案件失败: %v", err)
			continue
		}

		partyCount := 0
		for partyRows.Next() {
			var caseID uint
			var caseName, caseType, description, clientName, clientType, lawyerName string
			var foundLawyerID uint
			var createdAt time.Time

			if err := partyRows.Scan(&caseID, &caseName, &caseType, &description, &clientName, &clientType, &lawyerName, &createdAt, &foundLawyerID); err != nil {
				continue
			}

			partyCount++
			fmt.Printf("  %d. %s (客户: %s, 律师: %s)\n", partyCount, caseName, clientName, lawyerName)

			// 如果是同一个律师的案件，标记为潜在冲突
			if foundLawyerID == uint(lawyerID) {
				fmt.Printf("     ⚠️ 同一律师代理的对方案件!\n")
			}
		}
		partyRows.Close()

		if partyCount == 0 {
			fmt.Println("  未找到相关案件")
		}
	}

	// 3. 直接查询张伟律师代理的案件
	fmt.Println("\n=== 张伟律师代理的所有案件 ===")
	var lawyerCases []Case
	if err := db.Where("lawyer_id = ? AND deleted_at IS NULL", lawyerID).Find(&lawyerCases).Error; err != nil {
		log.Fatal("查询律师案件失败:", err)
	}

	fmt.Printf("张伟律师代理的案件数量: %d\n", len(lawyerCases))
	for i, case_ := range lawyerCases {
		// 获取客户名称
		var client Client
		if err := db.First(&client, case_.ClientID).Error; err != nil {
			continue
		}

		fmt.Printf("  %d. %s (客户: %s)\n", i+1, case_.Title, client.Name)

		// 检查是否与腾讯有直接冲突
		if containsCompany(case_.Title, "腾讯") || containsCompany(case_.Description, "腾讯") {
			fmt.Printf("     ⚠️ 直接冲突: 案件涉及腾讯!\n")
		}
	}

	// 4. 检查字节跳动和阿里巴巴的竞争关系
	fmt.Println("\n=== 检查行业竞争关系 ===")

	// 查找张伟代理的互联网公司案件
	var internetCases []Case
	if err := db.Where("lawyer_id = ? AND deleted_at IS NULL", lawyerID).Find(&internetCases).Error; err != nil {
		log.Fatal("查询互联网案件失败:", err)
	}

	internetCompanies := make(map[string]bool)
	for _, case_ := range internetCases {
		var client Client
		if err := db.First(&client, case_.ClientID).Error; err != nil {
			continue
		}

		if isInternetCompany(client.Name) {
			internetCompanies[client.Name] = true
			fmt.Printf("互联网客户: %s (案件: %s)\n", client.Name, case_.Title)
		}
	}

	fmt.Printf("\n发现的互联网公司数量: %d\n", len(internetCompanies))
	if len(internetCompanies) > 1 {
		fmt.Println("⚠️ 发现多个互联网公司客户，存在潜在竞争冲突!")

		// 检查是否有竞争关系
		companies := make([]string, 0, len(internetCompanies))
		for company := range internetCompanies {
			companies = append(companies, company)
		}

		for i, company1 := range companies {
			for j, company2 := range companies {
				if i < j {
					if hasCompetitionRelationship(company1, company2) {
						fmt.Printf("  🚨 竞争关系: %s vs %s\n", company1, company2)
					}
				}
			}
		}
	}
}

func containsCompany(text, company string) bool {
	return len(text) > 0 && (containsIgnoreCase(text, company))
}

func containsIgnoreCase(text, substr string) bool {
	text = strings.ToLower(text)
	substr = strings.ToLower(substr)
	return strings.Contains(text, substr)
}

func isInternetCompany(companyName string) bool {
	companyName = strings.ToLower(companyName)
	internetKeywords := []string{"科技", "网络", "软件", "信息", "互联网", "阿里", "腾讯", "字节", "百度", "京东", "美团"}

	for _, keyword := range internetKeywords {
		if strings.Contains(companyName, keyword) {
			return true
		}
	}
	return false
}

func hasCompetitionRelationship(company1, company2 string) bool {
	// 简化的竞争关系判断
	company1 = strings.ToLower(company1)
	company2 = strings.ToLower(company2)

	// 阿里巴巴 vs 腾讯
	if (strings.Contains(company1, "阿里") && strings.Contains(company2, "腾讯")) ||
		(strings.Contains(company1, "腾讯") && strings.Contains(company2, "阿里")) {
		return true
	}

	// 阿里巴巴 vs 字节跳动
	if (strings.Contains(company1, "阿里") && strings.Contains(company2, "字节")) ||
		(strings.Contains(company1, "字节") && strings.Contains(company2, "阿里")) {
		return true
	}

	// 腾讯 vs 字节跳动
	if (strings.Contains(company1, "腾讯") && strings.Contains(company2, "字节")) ||
		(strings.Contains(company1, "字节") && strings.Contains(company2, "腾讯")) {
		return true
	}

	return false
}