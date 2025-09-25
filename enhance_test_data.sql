-- 完善数据库测试数据
-- 为法律办公自动化系统添加充分的测试数据

-- 1. 添加更多用户数据
INSERT OR IGNORE INTO users (email, password, name, role, phone, status, created_at, updated_at) VALUES 
('wang.lawyer@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '王律师', 'lawyer', '13900139001', 'active', datetime('now'), datetime('now')),
('chen.lawyer@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '陈律师', 'lawyer', '13900139002', 'active', datetime('now'), datetime('now')),
('liu.lawyer@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '刘律师', 'lawyer', '13900139003', 'active', datetime('now'), datetime('now')),
('zhao.lawyer@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '赵律师', 'lawyer', '13900139004', 'active', datetime('now'), datetime('now')),
('assistant@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '助理小王', 'assistant', '13900139005', 'active', datetime('now'), datetime('now')),
('assistant2@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '助理小李', 'assistant', '13900139006', 'active', datetime('now'), datetime('now')),
('accountant@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '会计小张', 'accountant', '13900139007', 'active', datetime('now'), datetime('now')),
('reception@example.com', '$2a$12$G.DTHR2xYdtpmqvjjNJGYOjLIRp2FGWI.sKZDWlD4BN7bDHWQy9eG', '前台小美', 'reception', '13900139008', 'active', datetime('now'), datetime('now'));

-- 2. 添加更多客户数据
INSERT OR IGNORE INTO clients (name, email, phone, address, company, status, notes, created_at, updated_at) VALUES 
('北京科创科技有限公司', 'tech@beijing.com', '010-88888888', '北京市海淀区中关村大街1号', '北京科创科技有限公司', 'active', '高新技术企业，主要涉及知识产权保护'),
('上海国际贸易公司', 'trade@shanghai.com', '021-66666666', '上海市浦东新区陆家嘴金融中心', '上海国际贸易公司', 'active', '进出口贸易企业，需要合同审查'),
('广州制造集团', 'manufacture@guangzhou.com', '020-77777777', '广州市天河区珠江新城', '广州制造集团', 'active', '大型制造企业，劳动纠纷较多'),
('深圳投资控股', 'investment@shenzhen.com', '0755-99999999', '深圳市南山区科技园', '深圳投资控股', 'active', '投资公司，涉及并购重组业务'),
('杭州电商企业', 'ecommerce@hangzhou.com', '0571-55555555', '杭州市余杭区未来科技城', '杭州电商企业', 'active', '电商平台，需要合规咨询'),
('成都地产公司', 'realestate@chengdu.com', '028-44444444', '成都市锦江区春熙路', '成都地产公司', 'active', '房地产开发商，涉及土地纠纷'),
('武汉医药集团', 'pharma@wuhan.com', '027-33333333', '武汉市洪山区光谷', '武汉医药集团', 'active', '医药企业，专利保护需求'),
('西安科技公司', 'xitech@xian.com', '029-22222222', '西安市高新区', '西安科技公司', 'active', '软件开发公司，涉及著作权保护'),
('天津物流公司', 'logistics@tianjin.com', '022-11111111', '天津市滨海新区', '天津物流公司', 'active', '物流企业，运输合同纠纷'),
('南京金融集团', 'finance@nanjing.com', '025-88887777', '南京市鼓楼区', '南京金融集团', 'active', '金融机构，需要合规审查'),
('重庆汽车制造', 'auto@chongqing.com', '023-66665555', '重庆市渝北区', '重庆汽车制造', 'active', '汽车制造商，供应链合同'),
('苏州电子企业', 'electronics@suzhou.com', '0512-77776666', '苏州市工业园区', '苏州电子企业', 'active', '电子制造，劳动密集型企业'),
('青岛外贸公司', 'qingdao@import.com', '0532-55554444', '青岛市市南区', '青岛外贸公司', 'active', '进出口贸易，海商纠纷'),
('大连造船厂', 'shipyard@dalian.com', '0411-44443333', '大连市甘井子区', '大连造船厂', 'active', '船舶制造，国际合同'),
('厦门旅游集团', 'tourism@xiamen.com', '0592-33332222', '厦门市思明区', '厦门旅游集团', 'active', '旅游企业，服务合同纠纷');

-- 3. 添加律师数据
INSERT OR IGNORE INTO lawyers (name, email, phone, specialty, status) VALUES 
('王大律师', 'wang.dalawyer@example.com', '13900139111', '公司法, 合同法, 知识产权', 'active'),
('李资深律师', 'li.zishen@example.com', '13900139222', '刑法, 民法, 诉讼代理', 'active'),
('张专业律师', 'zhang.zhuanye@example.com', '13900139333', '劳动法, 社会保障法', 'active'),
('刘高级律师', 'liu.gaoji@example.com', '13900139444', '金融法, 证券法', 'active'),
('陈顾问律师', 'chen.guwen@example.com', '13900139555', '房地产法, 建筑法', 'active'),
('赵合伙人', 'zhao.hehuoren@example.com', '13900139666', '知识产权法, 技术合同', 'active'),
('孙诉讼律师', 'sun.susong@example.com', '13900139777', '诉讼法, 仲裁法', 'active'),
('周合规律师', 'zhou.hegui@example.com', '13900139888', '合规管理, 风险控制', 'active'),
('吴国际律师', 'wu.guoji@example.com', '13900139999', '国际法, 海商法', 'active'),
('郑税法律师', 'zheng.shuilv@example.com', '税法, 财会法律', 'active');

-- 4. 添加更多案件数据
INSERT OR IGNORE INTO cases (title, description, case_no, case_type, priority, status, client_id, lawyer_id) VALUES 
('北京科创知识产权侵权案', '北京科创科技有限公司发现其专利技术被竞争对手侵权，需要提起诉讼保护知识产权。涉及技术专利、商业秘密保护等问题。', 'BJ2024001', 'intellectual_property', 'high', 'active', 4, 5),
('上海国际贸易合同纠纷', '上海国际贸易公司与海外供应商发生合同纠纷，涉及货物质量、交付时间等条款争议。需要国际贸易法和合同法专业支持。', 'SH2024001', 'commercial', 'high', 'in_progress', 5, 6),
('广州制造劳动争议案', '广州制造集团与员工发生劳动争议，涉及劳动合同解除、经济补偿金等问题。需要劳动法专家协助处理。', 'GZ2024001', 'labor', 'medium', 'pending', 6, 7),
('深圳投资并购重组案', '深圳投资控股计划收购一家科技公司，需要进行尽职调查、并购协议起草等法律服务。涉及公司法、证券法等专业领域。', 'SZ2024001', 'merger_acquisition', 'high', 'active', 7, 8),
('杭州电商合规咨询', '杭州电商企业需要建立完善的合规体系，包括数据保护、消费者权益保护、广告宣传合规等方面的法律咨询。', 'HZ2024001', 'compliance', 'medium', 'completed', 8, 9),
('成都地产土地纠纷', '成都地产公司涉及一起土地使用权纠纷，需要处理土地出让合同、规划许可等法律问题。', 'CD2024001', 'real_estate', 'high', 'active', 9, 10),
('武汉医药专利保护', '武汉医药集团的新药专利面临挑战，需要提起专利无效宣告诉讼，保护公司的知识产权。', 'WH2024001', 'intellectual_property', 'high', 'in_progress', 10, 11),
('西安科技软件著作权', '西安科技公司开发的软件被抄袭，需要通过法律手段保护软件著作权，追究侵权责任。', 'XA2024001', 'intellectual_property', 'medium', 'pending', 11, 12),
('天津物流运输合同', '天津物流公司与客户发生运输合同纠纷，涉及货物损坏、延迟交付等问题。需要物流法和合同法专业支持。', 'TJ2024001', 'commercial', 'medium', 'active', 12, 13),
('南京金融合规审查', '南京金融集团需要对其金融产品进行合规审查，确保符合最新的金融监管要求。', 'NJ2024001', 'financial', 'high', 'in_progress', 13, 14),
('重庆汽车供应链', '重庆汽车制造与供应商发生供应链合同纠纷，涉及零部件质量、交付时间等问题。', 'CQ2024001', 'commercial', 'medium', 'pending', 14, 15),
('苏州电子劳动仲裁', '苏州电子企业与员工发生集体劳动争议，需要进行劳动仲裁，处理劳动合同、工资待遇等问题。', 'SZ2024002', 'labor', 'high', 'active', 15, 5),
('青岛外贸海商纠纷', '青岛外贸公司的货物在运输过程中损坏，与船公司发生海商纠纷，需要国际货物运输法的支持。', 'QD2024001', 'maritime', 'high', 'in_progress', 16, 6),
('大连造船国际合同', '大连造船厂与外国买家发生造船合同纠纷，涉及技术标准、交付时间等条款争议。', 'DL2024001', 'international', 'high', 'pending', 17, 7),
('厦门旅游服务合同', '厦门旅游集团与在线旅游平台发生服务合同纠纷，涉及佣金结算、服务质量等问题。', 'XM2024001', 'service', 'medium', 'completed', 18, 8);

-- 5. 添加文档数据
INSERT OR IGNORE INTO documents (document_no, case_id, client_id, file_name, original_name, file_size, file_type, file_path, document_type, description, uploader_id) VALUES 
('DOC2024001', 1, 4, 'patent_application.pdf', '专利申请书.pdf', 2048576, 'application/pdf', '/documents/patent_application.pdf', 'patent', '专利侵权案的专利申请书和相关证据材料', 1),
('DOC2024002', 1, 4, 'infringement_evidence.pdf', '侵权证据材料.pdf', 1048576, 'application/pdf', '/documents/infringement_evidence.pdf', 'evidence', '收集的侵权证据和技术对比分析', 1),
('DOC2024003', 2, 5, 'contract_dispute.pdf', '合同纠纷相关文件.pdf', 3072000, 'application/pdf', '/documents/contract_dispute.pdf', 'contract', '国际贸易合同及相关往来邮件', 2),
('DOC2024004', 2, 5, 'trade_agreement.pdf', '贸易协议.pdf', 1536000, 'application/pdf', '/documents/trade_agreement.pdf', 'agreement', '争议贸易协议的完整文本', 2),
('DOC2024005', 3, 6, 'labor_contract.pdf', '劳动合同.pdf', 524288, 'application/pdf', '/documents/labor_contract.pdf', 'contract', '争议劳动合同的完整文本', 3),
('DOC2024006', 3, 6, 'termination_notice.pdf', '解除通知.pdf', 262144, 'application/pdf', '/documents/termination_notice.pdf', 'notice', '劳动合同解除通知及相关证明', 3),
('DOC2024007', 4, 7, 'due_diligence.pdf', '尽职调查报告.pdf', 4096000, 'application/pdf', '/documents/due_diligence.pdf', 'report', '目标公司尽职调查报告', 4),
('DOC2024008', 4, 7, 'ma_agreement.pdf', '并购协议草案.pdf', 2048000, 'application/pdf', '/documents/ma_agreement.pdf', 'agreement', '并购协议草案和修改记录', 4),
('DOC2024009', 5, 8, 'compliance_manual.pdf', '合规手册.pdf', 10485760, 'application/pdf', '/documents/compliance_manual.pdf', 'manual', '电商平台合规操作手册', 5),
('DOC2024010', 6, 9, 'land_contract.pdf', '土地出让合同.pdf', 2560000, 'application/pdf', '/documents/land_contract.pdf', 'contract', '争议土地使用权的出让合同', 6),
('DOC2024011', 6, 9, 'planning_permit.pdf', '规划许可证.pdf', 512000, 'application/pdf', '/documents/planning_permit.pdf', 'permit', '相关规划许可文件', 6),
('DOC2024012', 7, 10, 'patent_certificate.pdf', '专利证书.pdf', 786432, 'application/pdf', '/documents/patent_certificate.pdf', 'certificate', '争议专利的证书文件', 7),
('DOC2024013', 8, 11, 'software_copyright.pdf', '软件著作权证书.pdf', 655360, 'application/pdf', '/documents/software_copyright.pdf', 'certificate', '软件著作权登记证书', 8),
('DOC2024014', 9, 12, 'shipping_contract.pdf', '运输合同.pdf', 1280000, 'application/pdf', '/documents/shipping_contract.pdf', 'contract', '争议运输合同的完整文本', 9),
('DOC2024015', 10, 13, 'financial_report.pdf', '金融产品合规报告.pdf', 3145728, 'application/pdf', '/documents/financial_report.pdf', 'report', '金融产品合规审查报告', 10),
('DOC2024016', 11, 14, 'supply_contract.pdf', '供应链合同.pdf', 1792000, 'application/pdf', '/documents/supply_contract.pdf', 'contract', '汽车零部件供应链合同', 11),
('DOC2024017', 12, 15, 'collective_agreement.pdf', '集体劳动合同.pdf', 1024000, 'application/pdf', '/documents/collective_agreement.pdf', 'contract', '电子企业集体劳动合同', 12),
('DOC2024018', 13, 16, 'bill_of_lading.pdf', '提单文件.pdf', 819200, 'application/pdf', '/documents/bill_of_lading.pdf', 'shipping', '争议货物的提单和相关文件', 13),
('DOC2024019', 14, 17, 'shipbuilding_contract.pdf', '造船合同.pdf', 4096000, 'application/pdf', '/documents/shipbuilding_contract.pdf', 'contract', '国际造船合同的完整文本', 14),
('DOC2024020', 15, 18, 'service_agreement.pdf', '服务合作协议.pdf', 1536000, 'application/pdf', '/documents/service_agreement.pdf', 'agreement', '旅游平台服务合作协议', 15);

-- 更新sqlite_sequence以确保自增ID正确
INSERT OR REPLACE INTO sqlite_sequence (name, seq) VALUES 
('users', (SELECT COALESCE(MAX(id), 0) FROM users)),
('clients', (SELECT COALESCE(MAX(id), 0) FROM clients)),
('cases', (SELECT COALESCE(MAX(id), 0) FROM cases)),
('lawyers', (SELECT COALESCE(MAX(id), 0) FROM lawyers)),
('documents', (SELECT COALESCE(MAX(id), 0) FROM documents));