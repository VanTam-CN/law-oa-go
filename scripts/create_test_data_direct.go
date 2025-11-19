package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const DSN = "host=localhost port=5432 user=postgres password=postgres dbname=law_oa sslmode=disable"

func main() {
	db, err := sql.Open("postgres", DSN)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")
	fmt.Println("🚀 开始创建测试用户和审批数据...")

	if err := createTestUsersAndData(db); err != nil {
		log.Fatal("创建测试数据失败:", err)
	}

	fmt.Println("✅ 测试数据创建完成！")
	fmt.Println("🎉 现在可以使用以下账号登录测试：")
	fmt.Println("  邮箱: zhangsan@law.com")
	fmt.Println("  密码: 123456")
	fmt.Println("  角色: 律师")
}

func createTestUsersAndData(db *sql.DB) error {
	// 1. 创建简化的users表（如果不存在）
	fmt.Println("📋 检查/创建users表...")
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100) NOT NULL UNIQUE,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		role VARCHAR(50) DEFAULT 'user',
		phone VARCHAR(20),
		avatar VARCHAR(500),
		status VARCHAR(20) DEFAULT 'active',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(createUsersTable); err != nil {
		return fmt.Errorf("创建users表失败: %v", err)
	}

	// 2. 清理现有测试数据
	fmt.Println("🧹 清理现有测试数据...")
	db.Exec("DELETE FROM approval_records WHERE approval_request_id LIKE 'test-approval-%'")
	db.Exec("DELETE FROM approval_requests WHERE request_number LIKE 'AP-20241201%'")
	db.Exec("DELETE FROM users WHERE email IN ('admin@law.com', 'zhangsan@law.com', 'lisi@law.com', 'wangwu@law.com', 'zhaoliu@law.com', 'qianqi@law.com')")

	// 3. 创建测试用户（密码都是123456的bcrypt哈希）
	fmt.Println("👥 创建测试用户...")
	users := []struct {
		id       int
		username string
		name     string
		email    string
		password string
		role     string
		phone    string
	}{
		{1, "admin", "系统管理员", "admin@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "admin", "13800000001"},
		{2, "zhangsan", "张三", "zhangsan@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "lawyer", "13800000002"},
		{3, "lisi", "李四", "lisi@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "lawyer", "13800000003"},
		{4, "wangwu", "王五", "wangwu@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "lawyer", "13800000004"},
		{5, "zhaoliu", "赵六", "zhaoliu@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "assistant", "13800000005"},
		{6, "qianqi", "钱七", "qianqi@law.com", "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", "assistant", "13800000006"},
	}

	for _, user := range users {
		_, err := db.Exec(`
			INSERT INTO users (id, username, name, email, password, role, phone, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
			ON CONFLICT (id) DO NOTHING`,
			user.id, user.username, user.name, user.email, user.password, user.role, user.phone)
		if err != nil {
			return fmt.Errorf("创建用户 %s 失败: %v", user.name, err)
		}
		fmt.Printf("✅ 创建用户: %s (%s)\n", user.name, user.email)
	}

	// 4. 创建审批测试数据
	fmt.Println("📋 创建审批测试数据...")
	type ApprovalData struct {
		id                   string
		requestNumber        string
		title                string
		type                 string
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
		durationDays         int
		createdBy            string
		updatedBy            string
		metadata             string
		attachments          string
	}

	approvals := []ApprovalData{
		// 待我审批的申请 1（申请人赵六，当前审批人为张三）
		{
			id:                "test-approval-001",
			requestNumber:     "AP-20241201001",
			title:             "赵六的用章申请",
			type:              "document",
			category:          "行政",
			content:           "需要加盖公司公章用于合同签署，合同已通过法务审核。",
			applicantID:       "5",
			applicantName:     "赵六",
			applicantTitle:    "律师助理",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "submitted",
			submissionDate:    time.Now().Add(-1 * time.Hour),
			currentStage:      "supervisor_review",
			currentApproverID: "2",
			currentApproverName: "张三",
			workflowType:      "QUICK_APPROVAL",
			durationDays:      1,
			createdBy:         "5",
			updatedBy:         "5",
			metadata:          `{"document_type": "合同", "document_name": "服务合同.pdf"}`,
			attachments:       `["服务合同.pdf"]`,
		},
		// 待我审批的申请 2（申请人钱七，当前审批人为张三）
		{
			id:                "test-approval-002",
			requestNumber:     "AP-20241201002",
			title:             "钱七的紧急采购申请",
			type:              "procurement",
			category:          "行政",
			content:           "办公室打印机故障，需要紧急采购新打印机一台，预算5000元。",
			applicantID:       "6",
			applicantName:     "钱七",
			applicantTitle:    "行政主管",
			departmentID:      "dept-003",
			departmentName:    "行政部",
			urgency:           "very_urgent",
			priority:          "high",
			status:            "submitted",
			submissionDate:    time.Now().Add(-30 * time.Minute),
			currentStage:      "emergency_approval",
			currentApproverID: "2",
			currentApproverName: "张三",
			workflowType:      "URGENT_APPROVAL",
			durationDays:      1,
			createdBy:         "6",
			updatedBy:         "6",
			metadata:          `{"item_name": "打印机", "budget": 5000}`,
			attachments:       `["采购申请.pdf", "报价单.pdf"]`,
		},
		// 我的申请 1（申请人张三）
		{
			id:                "test-approval-003",
			requestNumber:     "AP-20241201003",
			title:             "张三的年假申请",
			type:              "leave",
			category:          "人事行政",
			content:           "因家庭事务需要回老家处理，特申请年假5天，时间从2024年12月10日到2024年12月14日。",
			applicantID:       "2",
			applicantName:     "张三",
			applicantTitle:    "高级律师",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "submitted",
			submissionDate:    time.Now().Add(-4 * time.Hour),
			currentStage:      "department_head_review",
			currentApproverID: "3",
			currentApproverName: "李四",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      5,
			createdBy:         "2",
			updatedBy:         "2",
			metadata:          `{"leave_type": "年假", "reason": "家庭事务", "start_date": "2024-12-10", "end_date": "2024-12-14"}`,
			attachments:       `[]`,
		},
		// 我的申请 2（申请人张三）
		{
			id:                "test-approval-004",
			requestNumber:     "AP-20241201004",
			title:             "张三的培训申请",
			type:              "training",
			category:          "人事",
			content:           "参加专业法律培训课程，提升专业技能。",
			applicantID:       "2",
			applicantName:     "张三",
			applicantTitle:    "高级律师",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "submitted",
			submissionDate:    time.Now().Add(-2 * 24 * time.Hour),
			currentStage:      "department_head_review",
			currentApproverID: "3",
			currentApproverName: "李四",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      3,
			createdBy:         "2",
			updatedBy:         "2",
			metadata:          `{"course_name": "高级法律实务", "provider": "法律培训中心"}`,
			attachments:       `["培训课程介绍.pdf"]`,
		},
		// 已通过的申请（李四的费用报销）
		{
			id:                "test-approval-005",
			requestNumber:     "AP-20241201005",
			title:             "李四的费用报销申请",
			type:              "expense",
			category:          "财务",
			content:           "出差北京参与客户会议，报销交通费800元，住宿费1200元，餐费300元，合计2300元。",
			applicantID:       "3",
			applicantName:     "李四",
			applicantTitle:    "部门主管",
			departmentID:      "dept-001",
			departmentName:    "诉讼部",
			urgency:           "normal",
			priority:          "medium",
			status:            "approved",
			submissionDate:    time.Now().Add(-3 * 24 * time.Hour),
			currentStage:      "completed",
			currentApproverID: "",
			currentApproverName: "",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      0,
			createdBy:         "3",
			updatedBy:         "2",
			metadata:          `{"expense_type": "差旅费", "amount": 2300}`,
			attachments:       `["发票.pdf", "报销单.pdf"]`,
		},
		// 被拒绝的申请（王五的立项申请）
		{
			id:                "test-approval-006",
			requestNumber:     "AP-20241201006",
			title:             "王五的立项申请",
			type:              "project",
			category:          "业务",
			content:           "计划开展新的法律服务项目，需要启动资金50万元，项目周期6个月。",
			applicantID:       "4",
			applicantName:     "王五",
			applicantTitle:    "合伙人",
			departmentID:      "dept-002",
			departmentName:    "业务拓展部",
			urgency:           "urgent",
			priority:          "high",
			status:            "rejected",
			submissionDate:    time.Now().Add(-5 * 24 * time.Hour),
			currentStage:      "department_head_review",
			currentApproverID: "3",
			currentApproverName: "李四",
			workflowType:      "STANDARD_APPROVAL",
			durationDays:      180,
			createdBy:         "4",
			updatedBy:         "3",
			metadata:          `{"project_name": "新法律服务项目", "budget": 500000}`,
			attachments:       `["项目计划书.pdf", "预算表.xlsx"]`,
		},
	}

	for _, approval := range approvals {
		query := `
		INSERT INTO approval_requests (
			id, request_number, title, type, category, content,
			applicant_id, applicant_name, applicant_title, department_id, department_name,
			urgency, priority, status, submission_date, current_stage,
			current_approver_id, current_approver_name, workflow_type,
			duration_days, created_by, updated_by, created_at, updated_at,
			metadata, attachments
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25)
		ON CONFLICT (id) DO NOTHING`

		_, err := db.Exec(query,
			approval.id, approval.requestNumber, approval.title, approval.type, approval.category, approval.content,
			approval.applicantID, approval.applicantName, approval.applicantTitle, approval.departmentID, approval.departmentName,
			approval.urgency, approval.priority, approval.status, approval.submissionDate, approval.currentStage,
			approval.currentApproverID, approval.currentApproverName, approval.workflowType,
			approval.durationDays, approval.createdBy, approval.updatedBy, time.Now(), time.Now(),
			approval.metadata, approval.attachments,
		)

		if err != nil {
			return fmt.Errorf("插入审批记录 %s 失败: %v", approval.requestNumber, err)
		}
		fmt.Printf("✅ 创建审批: %s - %s\n", approval.requestNumber, approval.title)
	}

	// 5. 创建审批记录
	fmt.Println("📝 创建审批记录...")
	type RecordData struct {
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
	}

	records := []RecordData{
		{
			id:                "test-record-001",
			approvalRequestID: "test-approval-005",
			stage:             "department_head_review",
			stageOrder:        1,
			approverID:        "3",
			approverName:      "李四",
			approverTitle:     "部门主管",
			decision:          "approve",
			decisionReason:    "费用合理，相关票据齐全，同意报销",
			decisionComments:  "报销项目符合公司规定",
			approvalDate:      time.Now().Add(-2 * 24 * time.Hour),
			status:            "active",
		},
		{
			id:                "test-record-002",
			approvalRequestID: "test-approval-006",
			stage:             "department_head_review",
			stageOrder:        1,
			approverID:        "3",
			approverName:      "李四",
			approverTitle:     "部门主管",
			decision:          "reject",
			decisionReason:    "项目预算过高，风险评估不足",
			decisionComments:  "需要提供更详细的市场分析",
			approvalDate:      time.Now().Add(-4 * 24 * time.Hour),
			status:            "active",
		},
	}

	for _, record := range records {
		query := `
		INSERT INTO approval_records (
			id, approval_request_id, stage, stage_order, approver_id, approver_name, approver_title,
			decision, decision_reason, decision_comments, approval_date, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO NOTHING`

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

	// 6. 显示统计信息
	fmt.Println("\n📊 数据统计:")

	var userCount, approvalCount, recordCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE email IN ('admin@law.com', 'zhangsan@law.com', 'lisi@law.com', 'wangwu@law.com', 'zhaoliu@law.com', 'qianqi@law.com')").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM approval_requests WHERE request_number LIKE 'AP-20241201%'").Scan(&approvalCount)
	db.QueryRow("SELECT COUNT(*) FROM approval_records WHERE approval_request_id LIKE 'test-approval-%'").Scan(&recordCount)

	fmt.Printf("  👥 测试用户: %d 个\n", userCount)
	fmt.Printf("  📋 审批申请: %d 条\n", approvalCount)
	fmt.Printf("  📝 审批记录: %d 条\n", recordCount)

	return nil
}