-- 创建测试用户数据
-- 管理员用户
INSERT IGNORE INTO users (username, password, email, phone, real_name, avatar, status) VALUES
('admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'admin@example.com', '13800138000', '管理员', 'https://ui-avatars.com/api/?name=Admin&background=0d8abc&color=fff', 'active'),
('lawyer', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer@example.com', '13800138001', '张律师', 'https://ui-avatars.com/api/?name=张律师&background=28a745&color=fff', 'active'),
('assistant', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'assistant@example.com', '13800138002', '李助理', 'https://ui-avatars.com/api/?name=李助理&background=ffc107&color=000', 'active');

-- 添加测试客户
INSERT IGNORE INTO clients (name, email, phone, address, company, notes, status) VALUES
('腾讯科技', 'tencent@example.com', '13800138020', '深圳市南山区科技园', '腾讯科技有限公司', '重要客户，需要优先处理', 'active'),
('阿里巴巴', 'alibaba@example.com', '13800138021', '杭州市西湖区阿里巴巴西溪园区', '阿里巴巴集团', '重要客户，需要优先处理', 'active'),
('百度公司', 'baidu@example.com', '13800138022', '北京市海淀区上地十街', '百度在线网络技术有限公司', '重要客户，需要优先处理', 'active'),
('字节跳动', 'bytedance@example.com', '13800138023', '北京市海淀区知春路', '北京字节跳动科技有限公司', '重要客户，需要优先处理', 'active'),
('小米科技', 'xiaomi@example.com', '13800138024', '北京市海淀区清河中街', '小米科技有限责任公司', '重要客户，需要优先处理', 'active'),
('华为技术', 'huawei@example.com', '13800138025', '深圳市龙岗区坂田华为基地', '华为技术有限公司', '重要客户，需要优先处理', 'active'),
('美团点评', 'meituan@example.com', '13800138026', '北京市朝阳区望京东路', '北京三快在线科技有限公司', '重要客户，需要优先处理', 'active'),
('滴滴出行', 'didi@example.com', '13800138027', '北京市海淀区东北旺路', '滴滴出行科技有限公司', '重要客户，需要优先处理', 'active'),
('京东集团', 'jd@example.com', '13800138028', '北京市朝阳区北辰路', '京东集团股份有限公司', '重要客户，需要优先处理', 'active'),
('网易公司', 'netease@example.com', '13800138029', '杭州市滨江区网商路', '网易（杭州）网络有限公司', '重要客户，需要优先处理', 'active');

-- 添加测试律师
INSERT IGNORE INTO lawyers (name, email, phone, specialty, department, status, experience_years, bio) VALUES
('张三律师', 'zhangsan@example.com', '13800138010', '民商事诉讼', '诉讼部', 'active', 8, '资深民商事诉讼律师，擅长合同纠纷、债权债务等案件'),
('李四律师', 'lisi@example.com', '13800138011', '公司法务', '公司法务部', 'active', 10, '专业公司法务律师，在公司并购、股权转让等领域有丰富经验'),
('王五律师', 'wangwu@example.com', '13800138012', '知识产权', '知识产权部', 'active', 6, '知识产权专业律师，专注于商标、专利、著作权等法律事务'),
('赵六律师', 'zhaoliu@example.com', '13800138013', '刑事辩护', '刑事辩护部', 'active', 12, '资深刑事辩护律师，在重大刑事案件辩护方面经验丰富'),
('钱七律师', 'qianqi@example.com', '13800138014', '房地产', '房地产部', 'active', 9, '专业房地产律师，精通房地产开发、买卖、租赁等法律事务'),
('孙八律师', 'sunba@example.com', '13800138015', '劳动法', '劳动法部', 'active', 7, '劳动法专业律师，处理劳动合同纠纷、工伤赔偿等案件');

-- 添加测试案件
INSERT IGNORE INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date, expected_end_date) VALUES
('劳动合同纠纷案', '员工与公司之间的劳动合同纠纷，涉及加班费、经济补偿金等问题', 1, 1, 'civil', 'high', 'active', '2024-01-15 10:00:00', '2024-04-15'),
('房屋买卖合同纠纷', '买方与卖方因房屋质量问题产生的合同纠纷', 2, 2, 'civil', 'medium', 'pending', '2024-02-01 09:00:00', '2024-05-01'),
('商标侵权纠纷', '公司商标被他人侵权，需要进行法律维权', 3, 3, 'commercial', 'high', 'active', '2024-01-20 14:00:00', '2024-04-20'),
('股权转让纠纷', '公司股东之间因股权转让产生的纠纷', 4, 4, 'commercial', 'high', 'active', '2024-03-01 11:00:00', '2024-06-01'),
('建设工程合同纠纷', '施工单位与发包方因工程质量问题产生的纠纷', 5, 5, 'civil', 'urgent', 'pending', '2024-03-15 15:00:00', '2024-05-15'),
('知识产权侵权纠纷', '公司的专利技术被他人侵权使用', 6, 6, 'commercial', 'medium', 'closed', '2024-01-10 08:00:00', '2024-02-10'),
('投资合同纠纷', '投资者与公司之间的投资合同纠纷', 7, 1, 'commercial', 'medium', 'active', '2024-02-10 16:00:00', '2024-05-10'),
('技术秘密侵权纠纷', '公司的技术秘密被前员工泄露', 8, 2, 'commercial', 'high', 'suspended', '2024-01-25 13:00:00', '2024-04-25'),
('不正当竞争纠纷', '竞争对手采用不正当手段进行商业竞争', 9, 3, 'commercial', 'medium', 'active', '2024-02-20 10:00:00', '2024-05-20'),
('股东权益纠纷', '公司股东之间因权益分配产生的纠纷', 10, 4, 'commercial', 'high', 'pending', '2024-03-05 14:00:00', '2024-06-05');

-- 添加测试财务记录
INSERT IGNORE INTO financial_records (case_id, type, amount, description, status, payment_date, invoice_number) VALUES
(1, 'fee', 50000.00, '劳动合同纠纷案律师费', 'paid', '2024-01-15', 'INV-2024-001'),
(2, 'fee', 80000.00, '房屋买卖合同纠纷案律师费', 'unpaid', '2024-02-01', 'INV-2024-002'),
(3, 'fee', 120000.00, '商标侵权纠纷案律师费', 'partial', '2024-01-20', 'INV-2024-003'),
(4, 'fee', 150000.00, '股权转让纠纷案律师费', 'paid', '2024-03-01', 'INV-2024-004'),
(5, 'fee', 100000.00, '建设工程合同纠纷案律师费', 'unpaid', '2024-03-15', 'INV-2024-005'),
(6, 'fee', 90000.00, '知识产权侵权纠纷案律师费', 'paid', '2024-01-10', 'INV-2024-006'),
(7, 'fee', 110000.00, '投资合同纠纷案律师费', 'partial', '2024-02-10', 'INV-2024-007'),
(8, 'fee', 130000.00, '技术秘密侵权纠纷案律师费', 'unpaid', '2024-01-25', 'INV-2024-008'),
(9, 'fee', 70000.00, '不正当竞争纠纷案律师费', 'partial', '2024-02-20', 'INV-2024-009'),
(10, 'fee', 140000.00, '股东权益纠纷案律师费', 'unpaid', '2024-03-05', 'INV-2024-010');

-- 添加测试审批记录
INSERT IGNORE INTO approvals (title, description, type, status, applicant_id, approver_id, apply_date, approve_date) VALUES
('案件费用审批', '劳动合同纠纷案费用调整申请', 'fee', 'approved', 1, 1, '2024-01-16 09:00:00', '2024-01-17 10:00:00'),
('案件延期审批', '房屋买卖合同纠纷案延期申请', 'extension', 'pending', 2, 1, '2024-02-02 14:00:00', NULL),
('新案件审批', '商标侵权纠纷案新案件申请', 'new_case', 'approved', 3, 1, '2024-01-21 11:00:00', '2024-01-22 15:00:00'),
('律师费用审批', '股权转让纠纷案费用审批', 'fee', 'rejected', 4, 1, '2024-03-02 16:00:00', '2024-03-03 09:00:00'),
('案件暂停审批', '技术秘密侵权纠纷案暂停申请', 'suspension', 'approved', 5, 1, '2024-01-26 13:00:00', '2024-01-27 14:00:00');

-- 更新用户统计信息
UPDATE users SET remark = '系统管理员，拥有所有权限' WHERE username = 'admin';
UPDATE users SET remark = '专业律师，主要负责民商事案件' WHERE username = 'lawyer';
UPDATE users SET remark = '律师助理，协助律师处理案件' WHERE username = 'assistant';