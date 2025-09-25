-- 为admin@example.com用户添加更多测试数据

-- 首先确保admin用户存在
INSERT IGNORE INTO users (name, email, password, role, phone, avatar, status) VALUES 
('Admin User', 'admin@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'admin', '13800138000', 'https://ui-avatars.com/api/?name=Admin&background=0d8abc&color=fff', 'active');

-- 添加更多测试客户
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

-- 添加更多测试案件
-- 注意：这里需要根据实际数据库中的client_id和lawyer_id进行调整
INSERT IGNORE INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date) VALUES 
('劳动合同纠纷案', '员工与公司之间的劳动合同纠纷，涉及加班费、经济补偿金等问题', 4, 1, 'civil', 'high', 'active', '2024-01-15 10:00:00'),
('房屋买卖合同纠纷', '买方与卖方因房屋质量问题产生的合同纠纷', 5, 2, 'civil', 'medium', 'pending', '2024-02-01 09:00:00'),
('商标侵权纠纷', '公司商标被他人侵权，需要进行法律维权', 6, 3, 'commercial', 'high', 'active', '2024-01-20 14:00:00'),
('股权转让纠纷', '公司股东之间因股权转让产生的纠纷', 7, 4, 'commercial', 'high', 'active', '2024-03-01 11:00:00'),
('建设工程合同纠纷', '施工单位与发包方因工程质量问题产生的纠纷', 8, 5, 'civil', 'urgent', 'pending', '2024-03-15 15:00:00'),
('知识产权侵权纠纷', '公司的专利技术被他人侵权使用', 1, 6, 'commercial', 'medium', 'closed', '2024-01-10 08:00:00'),
('投资合同纠纷', '投资者与公司之间的投资合同纠纷', 2, 1, 'commercial', 'medium', 'active', '2024-02-10 16:00:00'),
('技术秘密侵权纠纷', '公司的技术秘密被前员工泄露', 3, 2, 'commercial', 'high', 'suspended', '2024-01-25 13:00:00'),
('不正当竞争纠纷', '竞争对手采用不正当手段进行商业竞争', 4, 3, 'commercial', 'medium', 'active', '2024-02-20 10:00:00'),
('股东权益纠纷', '公司股东之间因权益分配产生的纠纷', 5, 4, 'commercial', 'high', 'pending', '2024-03-05 14:00:00'),
('企业借贷纠纷', '企业之间的借贷合同纠纷', 6, 5, 'commercial', 'medium', 'active', '2024-02-15 09:00:00'),
('房地产合作开发纠纷', '合作方因房地产项目开发产生的纠纷', 7, 6, 'civil', 'urgent', 'pending', '2024-03-20 16:00:00'),
('招标投标纠纷', '投标人因招标过程不公正产生的纠纷', 8, 1, 'administrative', 'medium', 'active', '2024-02-25 11:00:00'),
('特许经营合同纠纷', '特许经营双方因合同履行产生的纠纷', 1, 2, 'commercial', 'low', 'closed', '2024-01-05 15:00:00'),
('网络服务合同纠纷', '用户与网络服务商之间的合同纠纷', 2, 3, 'civil', 'medium', 'active', '2024-02-08 10:00:00'),
('广告合同纠纷', '广告主与广告公司之间的合同纠纷', 3, 4, 'commercial', 'low', 'pending', '2024-03-10 14:00:00'),
('运输合同纠纷', '货主与运输公司之间的运输合同纠纷', 4, 5, 'civil', 'medium', 'active', '2024-02-12 09:00:00'),
('保险合同纠纷', '投保人与保险公司之间的保险合同纠纷', 5, 6, 'civil', 'high', 'suspended', '2024-01-30 13:00:00'),
('委托合同纠纷', '委托人与受托人之间的委托合同纠纷', 6, 1, 'commercial', 'medium', 'active', '2024-02-18 15:00:00'),
('劳动合同纠纷案', '员工与公司之间的劳动合同纠纷，涉及加班费、经济补偿金等问题', 7, 2, 'civil', 'high', 'active', '2024-01-22 10:00:00'),
('房屋买卖合同纠纷', '买方与卖方因房屋质量问题产生的合同纠纷', 8, 3, 'civil', 'medium', 'pending', '2024-02-28 09:00:00');