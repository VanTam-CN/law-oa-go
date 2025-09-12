-- 插入初始数据

-- 插入管理员用户
INSERT INTO users (name, email, password, role, status) VALUES 
('系统管理员', 'admin@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'admin', 'active'),
('张律师', 'zhang@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer', 'active'),
('李律师', 'li@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer', 'active');

-- 插入示例客户
INSERT INTO clients (name, email, phone, address, company, status) VALUES 
('王先生', 'wang@example.com', '13800138001', '北京市朝阳区xxx街道', '北京科技有限公司', 'active'),
('刘女士', 'liu@example.com', '13800138002', '上海市浦东新区xxx路', '上海贸易公司', 'active'),
('陈总', 'chen@example.com', '13800138003', '深圳市南山区xxx大厦', '深圳投资集团', 'active');

-- 插入示例案件
INSERT INTO cases (title, description, client_id, lawyer_id, case_type, priority, status) VALUES 
('合同纠纷案', '关于供货合同的争议处理', 1, 2, 'commercial', 'high', 'active'),
('劳动争议案', '员工离职补偿金纠纷', 2, 3, 'civil', 'medium', 'pending'),
('知识产权案', '商标侵权诉讼', 3, 2, 'commercial', 'urgent', 'active');