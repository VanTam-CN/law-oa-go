package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CaseTypeCount struct {
	CaseType string
	Count    int64
}

func main() {
	// 数据库连接
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 查询所有案件类型分布
	var caseTypeCounts []CaseTypeCount

	if err := db.WithContext(db.Statement.Context).
		Table("cases").
		Select("case_type, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Group("case_type").
		Order("count DESC").
		Scan(&caseTypeCounts).Error; err != nil {
		log.Fatal("查询案件类型失败:", err)
	}

	fmt.Println("📊 数据库中所有案件类型分布:")
	fmt.Println("==================================")
	totalCases := int64(0)
	for _, ct := range caseTypeCounts {
		fmt.Printf("类型: %s, 数量: %d\n", ct.CaseType, ct.Count)
		totalCases += ct.Count
	}
	fmt.Printf("总计: %d 个案件\n", totalCases)

	// 特别查看律师45的案件类型
	fmt.Println("\n🔍 律师45的案件类型分布:")
	var lawyerCaseTypes []CaseTypeCount
	if err := db.WithContext(db.Statement.Context).
		Table("cases").
		Select("case_type, COUNT(*) as count").
		Where("lawyer_id = ? AND deleted_at IS NULL", 45).
		Group("case_type").
		Order("count DESC").
		Scan(&lawyerCaseTypes).Error; err != nil {
		log.Fatal("查询律师45案件类型失败:", err)
	}

	for _, ct := range lawyerCaseTypes {
		fmt.Printf("类型: %s, 数量: %d\n", ct.CaseType, ct.Count)
	}

	// 查看律师45与阿里巴巴、字节跳动相关的案件
	fmt.Println("\n🎯 律师45与阿里巴巴、字节跳动相关的案件:")
	type ConflictCase struct {
		ID         uint   `gorm:"column:id"`
		Title      string `gorm:"column:title"`
		ClientName string `gorm:"column:clients>name"`
		CaseType   string `gorm:"column:case_type"`
	}

	var conflictCases []ConflictCase
	if err := db.WithContext(db.Statement.Context).
		Table("cases").
		Select("cases.id, cases.title, clients.name as client_name, cases.case_type").
		Joins("LEFT JOIN clients ON cases.client_id = clients.id").
		Where("(cases.client_id IN (?, ?) OR cases.title ILIKE ? OR cases.title ILIKE ?) AND cases.lawyer_id = ? AND cases.deleted_at IS NULL",
			55, 57, "%阿里巴巴%", "%字节跳动%", 45).
		Order("cases.id DESC").
		Scan(&conflictCases).Error; err != nil {
		log.Fatal("查询冲突案件失败:", err)
	}

	for _, c := range conflictCases {
		fmt.Printf("案件ID: %d, 标题: %s, 客户: %s, 类型: %s\n",
			c.ID, c.Title, c.ClientName, c.CaseType)
	}
}
