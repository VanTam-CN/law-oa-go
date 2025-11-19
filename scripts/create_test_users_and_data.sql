-- 创建测试用户和审批数据的完整脚本
-- 运行命令: psql -h localhost -U postgres -d law_oa -f create_test_users_and_data.sql

-- 首先检查users表是否存在，如果不存在则创建简化版本
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') THEN
        RAISE NOTICE 'users表已存在';
    ELSE
        CREATE TABLE users (
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
        );
        RAISE NOTICE '创建了简化的users表';
    END IF;
END $$;

-- 清理现有测试数据
DELETE FROM approval_records WHERE approval_request_id LIKE 'test-approval-%';
DELETE FROM approval_requests WHERE request_number LIKE 'AP-20241201%';
DELETE FROM users WHERE email IN ('admin@law.com', 'zhangsan@law.com', 'lisi@law.com', 'wangwu@law.com', 'zhaoliu@law.com', 'qianqi@law.com');

-- 创建测试用户（密码都是123456，已bcrypt加密）
INSERT INTO users (id, username, name, email, password, role, phone, status) VALUES
(1, 'admin', '系统管理员', 'admin@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', '13800000001', 'active'),
(2, 'zhangsan', '张三', 'zhangsan@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000002', 'active'),
(3, 'lisi', '李四', 'lisi@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000003', 'active'),
(4, 'wangwu', '王五', 'wangwu@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', '13800000004', 'active'),
(5, 'zhaoliu', '赵六', 'zhaoliu@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'assistant', '13800000005', 'active'),
(6, 'qianqi', '钱七', 'qianqi@law.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'assistant', '13800000006', 'active');

-- 插入审批测试数据
INSERT INTO approval_requests (
    id, request_number, title, type, category, content,
    applicant_id, applicant_name, applicant_title, department_id, department_name,
    urgency, priority, status, submission_date, current_stage,
    current_approver_id, current_approver_name, workflow_type,
    duration_days, created_by, updated_by, created_at, updated_at,
    metadata, attachments
) VALUES
-- 待我审批的申请 1（申请人赵六，当前审批人为张三）
('test-approval-001', 'AP-20241201001', '赵六的用章申请', 'document', '行政',
 '需要加盖公司公章用于合同签署，合同已通过法务审核。',
 '5', '赵六', '律师助理', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '1 hour', 'supervisor_review',
 '2', '张三', 'QUICK_APPROVAL',
 1, '5', '5', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour',
 '{"document_type": "合同", "document_name": "服务合同.pdf"}', '["服务合同.pdf"]'),

-- 待我审批的申请 2（申请人钱七，当前审批人为张三）
('test-approval-002', 'AP-20241201002', '钱七的紧急采购申请', 'procurement', '行政',
 '办公室打印机故障，需要紧急采购新打印机一台，预算5000元。',
 '6', '钱七', '行政主管', 'dept-003', '行政部',
 'very_urgent', 'high', 'submitted', NOW() - INTERVAL '30 minutes', 'emergency_approval',
 '2', '张三', 'URGENT_APPROVAL',
 1, '6', '6', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes',
 '{"item_name": "打印机", "budget": 5000}', '["采购申请.pdf", "报价单.pdf"]'),

-- 我的申请 1（申请人张三）
('test-approval-003', 'AP-20241201003', '张三的年假申请', 'leave', '人事行政',
 '因家庭事务需要回老家处理，特申请年假5天，时间从2024年12月10日到2024年12月14日。',
 '2', '张三', '高级律师', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '4 hours', 'department_head_review',
 '3', '李四', 'STANDARD_APPROVAL',
 5, '2', '2', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours',
 '{"leave_type": "年假", "reason": "家庭事务", "start_date": "2024-12-10", "end_date": "2024-12-14"}', '[]'),

-- 我的申请 2（申请人张三）
('test-approval-004', 'AP-20241201004', '张三的培训申请', 'training', '人事',
 '参加专业法律培训课程，提升专业技能。',
 '2', '张三', '高级律师', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '2 days', 'department_head_review',
 '3', '李四', 'STANDARD_APPROVAL',
 3, '2', '2', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days',
 '{"course_name": "高级法律实务", "provider": "法律培训中心"}', '["培训课程介绍.pdf"]'),

-- 已通过的申请（李四的费用报销）
('test-approval-005', 'AP-20241201005', '李四的费用报销申请', 'expense', '财务',
 '出差北京参与客户会议，报销交通费800元，住宿费1200元，餐费300元，合计2300元。',
 '3', '李四', '部门主管', 'dept-001', '诉讼部',
 'normal', 'medium', 'approved', NOW() - INTERVAL '3 days', 'completed',
 '', '', 'STANDARD_APPROVAL',
 0, '3', '2', NOW() - INTERVAL '3 days', NOW() - INTERVAL '2 days',
 '{"expense_type": "差旅费", "amount": 2300}', '["发票.pdf", "报销单.pdf"]'),

-- 被拒绝的申请（王五的立项申请）
('test-approval-006', 'AP-20241201006', '王五的立项申请', 'project', '业务',
 '计划开展新的法律服务项目，需要启动资金50万元，项目周期6个月。',
 '4', '王五', '合伙人', 'dept-002', '业务拓展部',
 'urgent', 'high', 'rejected', NOW() - INTERVAL '5 days', 'department_head_review',
 '3', '李四', 'STANDARD_APPROVAL',
 180, '4', '3', NOW() - INTERVAL '5 days', NOW() - INTERVAL '4 days',
 '{"project_name": "新法律服务项目", "budget": 500000}', '["项目计划书.pdf", "预算表.xlsx"]');

-- 插入审批记录
INSERT INTO approval_records (
    id, approval_request_id, stage, stage_order, approver_id, approver_name, approver_title,
    decision, decision_reason, decision_comments, approval_date, status, created_at, updated_at
) VALUES
-- 李四费用报销的审批记录（已通过）
('test-record-001', 'test-approval-005', 'department_head_review', 1,
 '3', '李四', '部门主管',
 'approve', '费用合理，相关票据齐全，同意报销', '报销项目符合公司规定',
 NOW() - INTERVAL '2 days', 'active', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days'),

-- 王五立项申请的拒绝记录
('test-record-002', 'test-approval-006', 'department_head_review', 1,
 '3', '李四', '部门主管',
 'reject', '项目预算过高，风险评估不足', '需要提供更详细的市场分析',
 NOW() - INTERVAL '4 days', 'active', NOW() - INTERVAL '4 days', NOW() - INTERVAL '4 days');

-- 显示创建结果统计
SELECT '测试数据创建完成！' as status;

-- 显示用户数据统计
SELECT
    'users' as table_name,
    COUNT(*) as record_count
FROM users
WHERE email IN ('admin@law.com', 'zhangsan@law.com', 'lisi@law.com', 'wangwu@law.com', 'zhaoliu@law.com', 'qianqi@law.com')

UNION ALL

-- 显示审批申请数据统计
SELECT
    'approval_requests' as table_name,
    COUNT(*) as record_count
FROM approval_requests
WHERE request_number LIKE 'AP-20241201%'

UNION ALL

-- 显示审批记录数据统计
SELECT
    'approval_records' as table_name,
    COUNT(*) as record_count
FROM approval_records
WHERE approval_request_id LIKE 'test-approval-%';

-- 显示按申请人统计的审批申请（当前用户ID=2，张三）
SELECT
    '我申请的审批' as 类型,
    COUNT(*) as 数量,
    STRING_AGG(title, ', ') as 申请列表
FROM approval_requests
WHERE applicant_id = '2' AND request_number LIKE 'AP-20241201%'

UNION ALL

-- 显示待我审批的申请（当前用户ID=2，张三）
SELECT
    '待我审批' as 类型,
    COUNT(*) as 数量,
    STRING_AGG(title, ', ') as 申请列表
FROM approval_requests
WHERE current_approver_id = '2' AND request_number LIKE 'AP-20241201%';

-- 显示测试用户登录信息
SELECT
    '测试用户登录信息' as 信息,
    '用户名/邮箱' as 账号,
    '密码' as 密码,
    '角色' as 角色
FROM (VALUES
    ('admin@law.com', '123456', '管理员'),
    ('zhangsan@law.com', '123456', '律师'),
    ('lisi@law.com', '123456', '律师'),
    ('wangwu@law.com', '123456', '律师'),
    ('zhaoliu@law.com', '123456', '助理'),
    ('qianqi@law.com', '123456', '助理')
) AS test_users(账号, 密码, 角色);