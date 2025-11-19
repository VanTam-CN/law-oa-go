-- 审批系统测试数据插入脚本
-- 运行命令: psql -h localhost -U postgres -d law_oa -f insert_approval_test_data.sql

-- 清理现有测试数据
DELETE FROM approval_records WHERE approval_request_id LIKE 'test-approval-%';
DELETE FROM approval_requests WHERE request_number LIKE 'AP-20241201%';

-- 插入审批测试数据
INSERT INTO approval_requests (
    id, request_number, title, type, category, content,
    applicant_id, applicant_name, applicant_title, department_id, department_name,
    urgency, priority, status, submission_date, current_stage,
    current_approver_id, current_approver_name, workflow_type,
    duration_days, created_by, updated_by, created_at, updated_at,
    metadata, attachments
) VALUES
-- 待审批的申请（申请人张三）
('test-approval-001', 'AP-20241201001', '张三的年假申请', 'leave', '人事行政',
 '因家庭事务需要回老家处理，特申请年假5天，时间从2024年12月10日到2024年12月14日。',
 '1', '张三', '高级律师', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '2 hours', 'department_head_review',
 '2', '李四', 'STANDARD_APPROVAL',
 5, '1', '1', NOW(), NOW(),
 '{"leave_type": "年假", "reason": "家庭事务", "start_date": "2024-12-10", "end_date": "2024-12-14"}', '[]'),

-- 已通过的申请（申请人李四）
('test-approval-002', 'AP-20241201002', '李四的费用报销申请', 'expense', '财务',
 '出差北京参与客户会议，报销交通费800元，住宿费1200元，餐费300元，合计2300元。',
 '2', '李四', '部门主管', 'dept-001', '诉讼部',
 'normal', 'medium', 'approved', NOW() - INTERVAL '48 hours', 'completed',
 '', '', 'STANDARD_APPROVAL',
 0, '2', '3', NOW() - INTERVAL '48 hours', NOW() - INTERVAL '24 hours',
 '{"expense_type": "差旅费", "amount": 2300}', '["发票.pdf", "报销单.pdf"]'),

-- 被拒绝的申请（申请人王五）
('test-approval-003', 'AP-20241201003', '王五的立项申请', 'project', '业务',
 '计划开展新的法律服务项目，需要启动资金50万元，项目周期6个月。',
 '3', '王五', '合伙人', 'dept-002', '业务拓展部',
 'urgent', 'high', 'rejected', NOW() - INTERVAL '72 hours', 'department_head_review',
 '2', '李四', 'STANDARD_APPROVAL',
 180, '3', '2', NOW() - INTERVAL '72 hours', NOW() - INTERVAL '36 hours',
 '{"project_name": "新法律服务项目", "budget": 500000}', '["项目计划书.pdf", "预算表.xlsx"]'),

-- 待我审批的申请 1（申请人赵六，当前审批人为用户ID=1）
('test-approval-004', 'AP-20241201004', '赵六的用章申请', 'document', '行政',
 '需要加盖公司公章用于合同签署，合同已通过法务审核。',
 '4', '赵六', '律师助理', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '1 hour', 'supervisor_review',
 '1', '张三', 'QUICK_APPROVAL',
 1, '4', '4', NOW() - INTERVAL '1 hour', NOW() - INTERVAL '1 hour',
 '{"document_type": "合同", "document_name": "服务合同.pdf"}', '["服务合同.pdf"]'),

-- 待我审批的申请 2（申请人钱七，当前审批人为用户ID=1）
('test-approval-005', 'AP-20241201005', '钱七的紧急采购申请', 'procurement', '行政',
 '办公室打印机故障，需要紧急采购新打印机一台，预算5000元。',
 '5', '钱七', '行政主管', 'dept-003', '行政部',
 'very_urgent', 'high', 'submitted', NOW() - INTERVAL '30 minutes', 'emergency_approval',
 '1', '张三', 'URGENT_APPROVAL',
 1, '5', '5', NOW() - INTERVAL '30 minutes', NOW() - INTERVAL '30 minutes',
 '{"item_name": "打印机", "budget": 5000}', '["采购申请.pdf", "报价单.pdf"]'),

-- 我的另一个申请（申请人用户ID=1）
('test-approval-006', 'AP-20241201006', '张三的培训申请', 'training', '人事',
 '参加专业法律培训课程，提升专业技能。',
 '1', '张三', '高级律师', 'dept-001', '诉讼部',
 'normal', 'medium', 'submitted', NOW() - INTERVAL '4 hours', 'department_head_review',
 '2', '李四', 'STANDARD_APPROVAL',
 3, '1', '1', NOW() - INTERVAL '4 hours', NOW() - INTERVAL '4 hours',
 '{"course_name": "高级法律实务", "provider": "法律培训中心"}', '["培训课程介绍.pdf"]');

-- 插入审批记录
INSERT INTO approval_records (
    id, approval_request_id, stage, stage_order, approver_id, approver_name, approver_title,
    decision, decision_reason, decision_comments, approval_date, status, created_at, updated_at
) VALUES
-- 李四费用报销的审批记录
('test-record-001', 'test-approval-002', 'department_head_review', 1,
 '2', '李四', '部门主管',
 'approve', '费用合理，相关票据齐全，同意报销', '报销项目符合公司规定',
 NOW() - INTERVAL '24 hours', 'active', NOW() - INTERVAL '24 hours', NOW() - INTERVAL '24 hours'),

-- 王五立项申请的拒绝记录
('test-record-002', 'test-approval-003', 'department_head_review', 1,
 '2', '李四', '部门主管',
 'reject', '项目预算过高，风险评估不足', '需要提供更详细的市场分析',
 NOW() - INTERVAL '36 hours', 'active', NOW() - INTERVAL '36 hours', NOW() - INTERVAL '36 hours');

-- 显示插入结果
SELECT '审批测试数据插入完成！' as status;

-- 显示当前数据统计
SELECT
    'approval_requests' as table_name,
    COUNT(*) as record_count
FROM approval_requests
WHERE request_number LIKE 'AP-20241201%'

UNION ALL

SELECT
    'approval_records' as table_name,
    COUNT(*) as record_count
FROM approval_records
WHERE approval_request_id LIKE 'test-approval-%';

-- 显示按状态统计的审批申请
SELECT
    status as 状态,
    COUNT(*) as 数量,
    STRING_AGG(title, ', ') as 申请列表
FROM approval_requests
WHERE request_number LIKE 'AP-20241201%'
GROUP BY status
ORDER BY COUNT(*) DESC;