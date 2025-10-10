-- 创建基本测试用户数据
INSERT IGNORE INTO users (username, password, email, phone, real_name, status) VALUES
('admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'admin@example.com', '13800138000', '系统管理员', 'active'),
('lawyer', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer@example.com', '13800138001', '张律师', 'active'),
('assistant', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'assistant@example.com', '13800138002', '李助理', 'active');

-- 创建基本测试客户数据
INSERT IGNORE INTO clients (client_name, phone, email, company, status) VALUES
('腾讯科技', '13800138020', 'tencent@example.com', '腾讯科技有限公司', 'active'),
('阿里巴巴', '13800138021', 'alibaba@example.com', '阿里巴巴集团', 'active'),
('百度公司', '13800138022', 'baidu@example.com', '百度在线网络技术有限公司', 'active');

-- 创建基本测试案件数据
INSERT IGNORE INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date) VALUES
('劳动合同纠纷案', '员工与公司之间的劳动合同纠纷，涉及加班费、经济补偿金等问题', 1, 1, 'civil', 'high', 'active', '2024-01-15 10:00:00'),
('商标侵权纠纷', '公司商标被他人侵权，需要进行法律维权', 2, 1, 'commercial', 'high', 'active', '2024-01-20 14:00:00'),
('股权转让纠纷', '公司股东之间因股权转让产生的纠纷', 3, 1, 'commercial', 'high', 'active', '2024-03-01 11:00:00');