//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

// DSN - 数据库连接信息
const DSN = "host=localhost port=5432 user=law_oa_user password=law_oa_password dbname=law_oa_db sslmode=disable"

func main() {
	// 连接数据库
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	// 测试连接
	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("🚀 开始创建审批测试数据...")

	// 创建简单的测试数据
	if err := createSimpleApprovalData(db); err != nil {
		log.Fatal("创建测试数据失败:", err)
	}

	fmt.Println("✅ 审批测试数据创建完成！")
}

// createSimpleApprovalData 创建简单的审批测试数据
func createSimpleApprovalData(db *sql.DB) error {
	// 清理现有测试数据
	fmt.Println("🧹 清理现有测试数据...")
	_, err := db.Exec(`DELETE FROM approval_requests WHERE id LIKE 'test-%'`)
	if err != nil {
		return fmt.Errorf("清理测试数据失败: %v", err)
	}

	// 插入新的测试数据
	fmt.Println("📝 插入审批测试数据...")

	// 测试数据
	testApprovals := []struct {
		id             string
		title          string
		typeName       string
		content        string
		applicantID    string
		applicantName  string
		departmentName string
		status         string
	}{
		{
			id:             "test-approval-001",
			title:          "测试年假申请",
			typeName:       "leave",
			content:        "这是一个测试的年假申请，用于验证审批功能。",
			applicantID:    "1",
			applicantName:  "测试用户",
			departmentName: "测试部门",
			status:         "submitted",
		},
		{
			id:             "test-approval-002",
			title:          "测试费用报销",
			typeName:       "expense",
			content:        "这是一个测试的费用报销申请，用于验证审批功能。",
			applicantID:    "1",
			applicantName:  "测试用户",
			departmentName: "测试部门",
			status:         "approved",
		},
		{
			id:             "test-approval-003",
			title:          "测试采购申请",
			typeName:       "procurement",
			content:        "这是一个测试的采购申请，用于验证审批功能。",
			applicantID:    "1",
			applicantName:  "测试用户",
			departmentName: "测试部门",
			status:         "pending",
		},
	}

	for _, approval := range testApprovals {
		query := `
		INSERT INTO approval_requests (
			id, request_number, title, type, content, applicant_id, applicant_name, department_name, status,
			submission_date, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		_, err := db.Exec(query,
			approval.id, fmt.Sprintf("AP-%s", approval.id), approval.title, approval.typeName, approval.content,
			approval.applicantID, approval.applicantName, approval.departmentName, approval.status,
			time.Now(), time.Now(), time.Now(),
		)

		if err != nil {
			return fmt.Errorf("插入审批记录 %s 失败: %v", approval.id, err)
		}

		fmt.Printf("✅ 创建审批记录: %s - %s\n", approval.id, approval.title)
	}

	return nil
}