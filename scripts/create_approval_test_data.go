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

	// 创建测试数据
	if err := createApprovalTestData(db); err != nil {
		log.Fatal("创建测试数据失败:", err)
	}

	fmt.Println("✅ 审批测试数据创建完成！")
	fmt.Println("🎉 现在您可以重新访问审批管理界面查看数据了")
}

// createApprovalTestData 创建审批测试数据
func createApprovalTestData(db *sql.DB) error {
	// 首先检查是否有用户数据，如果没有则创建测试用户
	if err := ensureTestUsers(db); err != nil {
		return fmt.Errorf("创建测试用户失败: %v", err)
	}

	// 创建审批测试数据
	approvals := []struct {
		id                   string
		requestNumber        string
		title                string
		approvalType         string
		category             string
		content              string
		applicantID          string
		applicantName        string
		applicantTitle       string
		departmentID         string
		departmentName       string
		urgency              string
		priority             string
		status               string
		submissionDate       time.Time
		currentStage         string
		currentApproverID    string
		currentApproverName  string
		workflowType         string
		expectedEffectiveDate *time.Time
		expectedExpiryDate   *time.Time
		durationDays         int
		createdBy            string
		updatedBy            string
		metadata             string
		attachments          string
	}{
		// 待审批的申请
		{
			id:                "test-approval-001",
			requestNumber:     "AP-20241201001",
			title:             "张三的年假申请",
			approvalType:      "leave",
			category:          "人事行政",
			content:           "因家庭事务需要回老家处理，特申请年假5天，时间从2024年12月10日到2024年12月14日。",
			applicantID:       "1",
			applicantName:     "张三",
			applicantTitle:    "高级律师",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "submitted",
			submissionDate:    time.Now().Add(-2 * time.Hour),
			currentStage:      "department_head_review",
			currentApproverID: "2",
			currentApproverName: "李四",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      5,
			createdBy:         "1",
			updatedBy:         "1",
			metadata:          `{"leave_type": "年假", "reason": "家庭事务", "start_date": "2024-12-10", "end_date": "2024-12-14"}`,
			attachments:       `[]`,
		},
		// 已通过的申请
		{
			id:                "test-approval-002",
			requestNumber:     "AP-20241201002",
			title:             "李四的费用报销申请",
			approvalType:      "expense",
			category:          "财务",
			content:           "出差北京参与客户会议，报销交通费800元，住宿费1200元，餐费300元，合计2300元。相关发票已上传至系统。",
			applicantID:       "2",
			applicantName:     "李四",
			applicantTitle:    "部门主管",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "approved",
			submissionDate:    time.Now().Add(-48 * time.Hour),
			currentStage:      "completed",
			currentApproverID: "",
			currentApproverName: "",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      0,
			createdBy:         "2",
			updatedBy:         "3",
			metadata:          `{"expense_type": "差旅费", "amount": 2300, "expense_items": [{"name": "交通费", "amount": 800}, {"name": "住宿费", "amount": 1200}, {"name": "餐费", "amount": 300}]}`,
			attachments:       `["发票.pdf", "报销单.pdf"]`,
		},
		// 被拒绝的申请
		{
			id:                "test-approval-003",
			requestNumber:     "AP-20241201003",
			title:             "王五的立项申请",
			approvalType:      "project",
			category:          "业务",
			content:           "计划开展新的法律服务项目，需要启动资金50万元，项目周期6个月，预计收益100万元。",
			applicantID:       "3",
			applicantName:     "王五",
			applicantTitle:    "合伙人",
			departmentID:      "dept-002",
			departmentName:    "业务拓展部",
			urgency:           "urgent",
			priority:          "high",
			status:            "rejected",
			submissionDate:    time.Now().Add(-72 * time.Hour),
			currentStage:      "department_head_review",
			currentApproverID: "2",
			currentApproverName: "李四",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      180,
			createdBy:         "3",
			updatedBy:         "2",
			metadata:          `{"project_name": "新法律服务项目", "project_description": "开展新的法律服务业务", "budget": 500000, "duration": 180}`,
			attachments:       `["项目计划书.pdf", "预算表.xlsx"]`,
		},
		// 待我审批的申请（为当前用户ID=1创建）
		{
			id:                "test-approval-004",
			requestNumber:     "AP-20241201004",
			title:             "赵六的用章申请",
			approvalType:      "document",
			category:          "行政",
			content:           "需要加盖公司公章用于合同签署，合同已通过法务审核，请予批准。",
			applicantID:       "4",
			applicantName:     "赵六",
			applicantTitle:    "律师助理",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "submitted",
			submissionDate:    time.Now().Add(-1 * time.Hour),
			currentStage:      "supervisor_review",
			currentApproverID: "1", // 设置为当前用户ID
			currentApproverName: "张三",
			workflowType:      "QUICK_APPROVAL",
			durationDays:      1,
			createdBy:         "4",
			updatedBy:         "4",
			metadata:          `{"document_type": "合同", "document_name": "服务合同.pdf", "purpose": "合同签署"}`,
			attachments:       `["服务合同.pdf", "法务审核意见.pdf"]`,
		},
		// 另一个待我审批的申请
		{
			id:                "test-approval-005",
			requestNumber:     "AP-20241201005",
			title:             "钱七的紧急采购申请",
			approvalType:      "procurement",
			category:          "行政",
			content:           "办公室打印机故障，需要紧急采购新打印机一台，预算5000元，急需使用。",
			applicantID:       "5",
			applicantName:     "钱七",
			applicantTitle:    "行政主管",
			departmentID:      "dept-003",
			departmentName:    "行政部",
			urgency:           "very_urgent",
			priority:          "high",
			status:            "submitted",
			submissionDate:    time.Now().Add(-30 * time.Minute),
			currentStage:      "emergency_approval",
			currentApproverID: "1", // 设置为当前用户ID
			currentApproverName: "张三",
			workflowType:      "URGENT_APPROVAL",
			durationDays:      1,
			createdBy:         "5",
			updatedBy:         "5",
			metadata:          `{"item_name": "打印机", "item_type": "办公设备", "budget": 5000, "reason": "打印机故障急需更换"}`,
			attachments:       `["设备采购申请.pdf", "报价单.pdf"]`,
		},
	}

	// 清理现有的测试数据
	fmt.Println("🧹 清理现有测试数据...")
	_, err := db.Exec(`DELETE FROM approval_requests WHERE request_number LIKE 'AP-20241201%'`)
	if err != nil {
		return fmt.Errorf("清理测试数据失败: %v", err)
	}

	// 插入新的测试数据
	fmt.Println("📝 插入审批测试数据...")
	for _, approval := range approvals {
		query := `
		INSERT INTO approval_requests (
			id, request_number, title, type, category, content,
			applicant_id, applicant_name, applicant_title, department_id, department_name,
			urgency, priority, status, submission_date, current_stage,
			current_approver_id, current_approver_name, workflow_type,
			expected_effective_date, expected_expiry_date, duration_days,
			created_by, updated_by, created_at, updated_at,
			metadata, attachments
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27)
		`

		_, err := db.Exec(query,
			approval.id, approval.requestNumber, approval.title, approval.approvalType, approval.category, approval.content,
			approval.applicantID, approval.applicantName, approval.applicantTitle, approval.departmentID, approval.departmentName,
			approval.urgency, approval.priority, approval.status, approval.submissionDate, approval.currentStage,
			approval.currentApproverID, approval.currentApproverName, approval.workflowType,
			approval.expectedEffectiveDate, approval.expectedExpiryDate, approval.durationDays,
			approval.createdBy, approval.updatedBy, time.Now(), time.Now(),
			approval.metadata, approval.attachments,
		)

		if err != nil {
			return fmt.Errorf("插入审批记录 %s 失败: %v", approval.requestNumber, err)
		}

		fmt.Printf("✅ 创建审批记录: %s - %s\n", approval.requestNumber, approval.title)
	}

	// 创建一些审批记录
	fmt.Println("📝 创建审批记录...")
	if err := createApprovalRecords(db); err != nil {
		return fmt.Errorf("创建审批记录失败: %v", err)
	}

	return nil
}

// ensureTestUsers 确保测试用户存在
func ensureTestUsers(db *sql.DB) error {
	fmt.Println("👥 检查测试用户...")

	// 检查用户表是否存在
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'users'
		)
	`).Scan(&tableExists)

	if err != nil {
		return fmt.Errorf("检查用户表失败: %v", err)
	}

	if !tableExists {
		fmt.Println("⚠️ 用户表不存在，跳过用户数据创建")
		return nil
	}

	// 检查是否有测试用户，如果没有则创建
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id IN ($1, $2, $3, $4, $5)", 1, 2, 3, 4, 5).Scan(&userCount)
	if err != nil {
		// 如果用户表结构不同，我们跳过用户创建
		fmt.Println("⚠️ 无法查询用户表，可能表结构不同，跳过用户数据创建")
		return nil
	}

	if userCount < 5 {
		fmt.Println("👥 创建测试用户数据...")
		testUsers := []struct {
			id    int
			name  string
			email string
			role  string
		}{
			{1, "张三", "zhangsan@law.com", "律师"},
			{2, "李四", "lisi@law.com", "部门主管"},
			{3, "王五", "wangwu@law.com", "合伙人"},
			{4, "赵六", "zhaoliu@law.com", "律师助理"},
			{5, "钱七", "qianqi@law.com", "行政主管"},
		}

		for _, user := range testUsers {
			// 尝试不同的用户表结构
			_, err := db.Exec(`
				INSERT INTO users (id, name, email, role, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (id) DO NOTHING
			`, user.id, user.name, user.email, user.role, time.Now(), time.Now())

			if err != nil {
				fmt.Printf("⚠️ 创建用户 %s 失败（可能表结构不同）: %v\n", user.name, err)
			} else {
				fmt.Printf("✅ 创建测试用户: %s\n", user.name)
			}
		}
	} else {
		fmt.Println("✅ 测试用户已存在")
	}

	return nil
}

// createApprovalRecords 创建审批记录
func createApprovalRecords(db *sql.DB) error {
	records := []struct {
		id                string
		approvalRequestID string
		stage             string
		stageOrder        int
		approverID        string
		approverName      string
		approverTitle     string
		decision          string
		decisionReason    string
		decisionComments  string
		approvalDate      time.Time
		status            string
	}{
		// 李四的费用报销申请 - 已通过
		{
			id:                "test-record-001",
			approvalRequestID: "test-approval-002",
			stage:             "department_head_review",
			stageOrder:        1,
			approverID:        "2",
			approverName:      "李四",
			approverTitle:     "部门主管",
			decision:          "approve",
			decisionReason:    "费用合理，相关票据齐全，同意报销",
			decisionComments:  "报销项目符合公司规定，建议财务部门尽快处理",
			approvalDate:      time.Now().Add(-24 * time.Hour),
			status:            "active",
		},
		// 王五的立项申请 - 被拒绝
		{
			id:                "test-record-002",
			approvalRequestID: "test-approval-003",
			stage:             "department_head_review",
			stageOrder:        1,
			approverID:        "2",
			approverName:      "李四",
			approverTitle:     "部门主管",
			decision:          "reject",
			decisionReason:    "项目预算过高，风险评估不足，建议重新规划",
			decisionComments:  "需要提供更详细的市场分析和风险评估报告",
			approvalDate:      time.Now().Add(-36 * time.Hour),
			status:            "active",
		},
	}

	// 清理现有的审批记录
	fmt.Println("🧹 清理现有审批记录...")
	_, err := db.Exec(`DELETE FROM approval_records WHERE approval_request_id LIKE 'test-approval-%'`)
	if err != nil {
		return fmt.Errorf("清理审批记录失败: %v", err)
	}

	// 插入新的审批记录
	fmt.Println("📝 插入审批记录...")
	for _, record := range records {
		query := `
		INSERT INTO approval_records (
			id, approval_request_id, stage, stage_order, approver_id, approver_name, approver_title,
			decision, decision_reason, decision_comments, approval_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`

		_, err := db.Exec(query,
			record.id, record.approvalRequestID, record.stage, record.stageOrder,
			record.approverID, record.approverName, record.approverTitle,
			record.decision, record.decisionReason, record.decisionComments,
			record.approvalDate, record.status, time.Now(), time.Now(),
		)

		if err != nil {
			return fmt.Errorf("插入审批记录 %s 失败: %v", record.id, err)
		}

		fmt.Printf("✅ 创建审批记录: %s - %s\n", record.id, record.decision)
	}

	return nil
}