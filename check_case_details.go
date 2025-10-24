package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CaseDetail struct {
	ID          uint   `gorm:"column:id"`
	Title       string `gorm:"column:title"`
	Description string `gorm:"column:description"`
}

func main() {
	dsn := "host=localhost user=law_oa_user password=1q2w#E$R dbname=law_oa_db port=5432 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	// 检查被错误标记为冲突的案件
	caseIDs := []int{29, 28, 27, 26, 24, 23, 21, 35, 37}

	fmt.Println("🔍 检查被错误标记为腾讯冲突的案件内容:")
	fmt.Println("===========================================")

	for _, caseID := range caseIDs {
		var caseDetail CaseDetail
		if err := db.Table("cases").
			Select("id, title, description").
			Where("id = ? AND deleted_at IS NULL", caseID).
			First(&caseDetail).Error; err != nil {
			fmt.Printf("❌ 查询案件%d失败: %v\n", caseID, err)
			continue
		}

		fmt.Printf("\n案件ID: %d\n", caseDetail.ID)
		fmt.Printf("标题: %s\n", caseDetail.Title)
		fmt.Printf("描述: %s\n", caseDetail.Description)

		// 检查是否包含"腾讯"相关内容
		opponent := "腾讯\n垄断纠纷案"
		fmt.Printf("是否包含对方当事人 '%s': ", opponent)
		if contains(caseDetail.Title, opponent) || contains(caseDetail.Description, opponent) {
			fmt.Printf("❌ 错误匹配！\n")
		} else {
			fmt.Printf("✅ 无直接关联\n")
		}
	}
}

func contains(text, substr string) bool {
	return len(text) > 0 && len(substr) > 0 &&
		   (fmt.Sprintf("%s", text) == fmt.Sprintf("%s", substr) ||
		    findSubstring(text, substr))
}

func findSubstring(text, substr string) bool {
	if len(text) < len(substr) {
		return false
	}
	for i := 0; i <= len(text)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if text[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}