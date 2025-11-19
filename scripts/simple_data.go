package main

import (
	"database/sql"
	"fmt"
	"log"

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

	// 1. 检查/创建users表
	fmt.Println("📋 检查users表...")
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
		log.Printf("创建users表失败: %v", err)
	}

	// 2. 创建测试用户
	fmt.Println("👥 创建测试用户...")
	_, err = db.Exec(`
		INSERT INTO users (id, username, name, email, password, role, phone, status)
		VALUES
		(1, 'zhangsan', '张三', 'zhangsan@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000002', 'active'),
		(2, 'lisi', '李四', 'lisi@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000003', 'active'),
		(3, 'wangwu', '王五', 'wangwu@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000004', 'active')
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		log.Printf("创建用户失败: %v", err)
	} else {
		fmt.Println("✅ 测试用户创建成功")
	}

	// 3. 创建审批申请
	fmt.Println("📋 创建审批申请...")
	_, err = db.Exec(`
		INSERT INTO approval_requests (
			id, request_number, title, type, category, content,
			applicant_id, applicant_name, applicant_title, department_id, department_name,
			urgency, priority, status, submission_date, current_stage,
			current_approver_id, current_approver_name, workflow_type,
			duration_days, created_by, updated_by, created_at, updated_at,
			metadata, attachments
		) VALUES
		('test-approval-001', 'AP-20241201001', '张三的年假申请', 'leave', '人事行政',
		 '因家庭事务需要回老家处理，特申请年假5天',
		 '1', '张三', '高级律师', 'dept-001', '诉讼部',
		 'normal', 'medium', 'submitted', NOW() - INTERVAL '2 hours', 'department_head_review',
		 '2', '李四', 'STANDARD_APPROVAL',
		 5, '1', '1', NOW() - INTERVAL '2 hours', NOW() - INTERVAL '2 hours',
		 '{"leave_type": "年假"}', '[]'),

		('test-approval-002', 'AP-20241201002', '李四的用章申请', 'document', '行政',
		 '需要加盖公司公章用于合同签署',
		 '2', '李四', '律师助理', 'dept-001', '诉讼部',
		 'normal', 'medium', 'submitted', NOW() - INTERVAL '1 hour', 'supervisor_review',
		 '1', '张三', 'QUICK_APPROVAL',
		 1, '2', '2', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour',
		 '{"document_type": "合同"}', '["合同.pdf"]')

		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		log.Printf("创建审批申请失败: %v", err)
	} else {
		fmt.Println("✅ 审批申请创建成功")
	}

	// 4. 显示统计信息
	var userCount, approvalCount int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE email IN ('zhangsan@law.com', 'lisi@law.com', 'wangwu@law.com')").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM approval_requests WHERE request_number LIKE 'AP-20241201%'").Scan(&approvalCount)

	fmt.Printf("📊 数据统计:\n")
	fmt.Printf("  👥 测试用户: %d 个\n", userCount)
	fmt.Printf("  📋 审批申请: %d 条\n", approvalCount)

	fmt.Println("\n🎉 测试数据创建完成！")
	fmt.Println("🔑 登录信息:")
	fmt.Println("  邮箱: zhangsan@law.com")
	fmt.Println("  密码: 123456")
	fmt.Println("  角色: 律师")
}