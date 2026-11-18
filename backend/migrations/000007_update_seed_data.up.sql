-- 更新数据：添加admin@example.com用户和更多测试数据

-- 添加admin@example.com用户（密码为123456的bcrypt哈希）
INSERT IGNORE INTO users (name, email, password, role, phone, avatar, status) VALUES 
('Admin User', 'admin@example.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'admin', '13800138000', 'https://ui-avatars.com/api/?name=Admin&background=0d8abc&color=fff', 'active');

-- 添加更多律师用户
INSERT IGNORE INTO users (name, email, password, role, phone, avatar, status) VALUES 
('王律师', 'wang@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer', '13800138004', 'https://ui-avatars.com/api/?name=%E7%8E%8B%E5%BE%8B%E5%B8%88&background=28a745&color=fff', 'active'),
('赵律师', 'zhao@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer', '13800138005', 'https://ui-avatars.com/api/?name=%E8%B5%B5%E5%BE%8B%E5%B8%88&background=17a2b8&color=fff', 'active'),
('孙律师', 'sun@law-oa.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'lawyer', '13800138006', 'https://ui-avatars.com/api/?name=%E5%AD%99%E5%BE%8B%E5%B8%88&background=6f42c1&color=fff', 'active');

-- 添加更多客户数据
INSERT IGNORE INTO clients (name, email, phone, address, company, notes, status) VALUES 
('周总', 'zhou@example.com', '13800138007', '广州市天河区xxx路', '广州科技公司', '重要客户，需要优先处理', 'active'),
('吴女士', 'wu@example.com', '13800138008', '杭州市西湖区xxx街道', '杭州电商公司', '合同续签咨询', 'active'),
('郑先生', 'zheng@example.com', '13800138009', '成都市高新区xxx大厦', '成都软件公司', '知识产权保护', 'active'),
('孙经理', 'sunmgr@example.com', '13800138010', '武汉市江汉区xxx路', '武汉制造企业', '劳动法咨询', 'prospect'),
('钱总', 'qian@example.com', '13800138011', '南京市鼓楼区xxx街', '南京投资公司', '并购法律服务', 'active'),
('冯女士', 'feng@example.com', '13800138012', '西安市雁塔区xxx道', '西安房地产公司', '房产纠纷', 'inactive');

-- 添加更多案件数据（注意：这里需要调整client_id以匹配实际的客户数据）
INSERT IGNORE INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date) VALUES 
('商业合同纠纷', '供应商违约导致的商业损失赔偿', 4, 1, 'commercial', 'high', 'active', '2024-01-15 10:00:00'),
('劳动仲裁', '员工劳动合同解除争议', 5, 2, 'civil', 'medium', 'pending', '2024-02-01 09:00:00'),
('商标注册', '公司商标申请和知识产权保护', 6, 3, 'administrative', 'medium', 'active', '2024-01-20 14:00:00'),
('股权转让', '公司股东股权转让协议制定', 7, 4, 'commercial', 'high', 'active', '2024-03-01 11:00:00'),
('房产买卖', '商业地产买卖合同纠纷', 8, 5, 'civil', 'urgent', 'pending', '2024-03-15 15:00:00'),
('公司设立', '新公司注册和法律架构设计', 1, 6, 'administrative', 'low', 'closed', '2024-01-10 08:00:00'),
('债务追讨', '客户欠款催收和法律诉讼', 2, 1, 'commercial', 'medium', 'active', '2024-02-10 16:00:00'),
('知识产权', '专利侵权诉讼和赔偿', 3, 2, 'commercial', 'high', 'suspended', '2024-01-25 13:00:00');

-- 添加文档示例数据
INSERT IGNORE INTO documents (name, description, filename, filepath, filesize, mime_type, category, tags, entity_id, entity_type, status) VALUES 
('合同模板', '标准商业合同模板', 'contract_template.docx', '/uploads/documents/contract_template.docx', 25600, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'template', '["合同", "模板", "商业"]', NULL, NULL, 'active'),
('身份证复印件', '客户身份证明文件', 'id_card_copy.jpg', '/uploads/documents/id_card_copy.jpg', 102400, 'image/jpeg', 'identification', '["身份证", "证明", "客户"]', 1, 'client', 'active'),
('案件委托书', '案件代理委托协议', 'power_of_attorney.pdf', '/uploads/documents/power_of_attorney.pdf', 51200, 'application/pdf', 'legal', '["委托", "协议", "案件"]', 1, 'case', 'active'),
('营业执照', '公司营业执照副本', 'business_license.jpg', '/uploads/documents/business_license.jpg', 204800, 'image/jpeg', 'identification', '["营业执照", "公司", "证件"]', 1, 'client', 'active'),
('起诉状', '法院起诉状模板', 'indictment.docx', '/uploads/documents/indictment.docx', 30720, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'legal', '["起诉", "法院", "模板"]', NULL, NULL, 'active'),
('证据清单', '案件证据材料清单', 'evidence_list.xlsx', '/uploads/documents/evidence_list.xlsx', 40960, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 'evidence', '["证据", "清单", "材料"]', 1, 'case', 'active');