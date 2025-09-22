-- 插入测试客户数据
INSERT INTO clients (name, email, phone, address, company, type, id_card, industry, contact_person, contact_phone, source, notes, status) VALUES
('张三', 'zhangsan@example.com', '13800138001', '北京市朝阳区', '北京科技有限公司', 'company', '110101199001011234', '互联网', '李总', '13900139001', '朋友介绍', '重要客户，长期合作', 'active'),
('李四', 'lisi@example.com', '13800138002', '上海市浦东新区', '', 'personal', '310101199002022345', '金融', '', '', '网络推广', '个人客户，咨询离婚事宜', 'active'),
('王五', 'wangwu@example.com', '13800138003', '广州市天河区', '广州贸易有限公司', 'company', '440101199003033456', '贸易', '王经理', '13700137003', '展会', '企业客户，合同纠纷', 'active'),
('赵六', 'zhaoliu@example.com', '13800138004', '深圳市南山区', '', 'personal', '440101199004044567', '制造业', '', '', '广告', '个人客户，劳动仲裁', 'prospect'),
('孙七', 'sunqi@example.com', '13800138005', '杭州市西湖区', '杭州科技有限公司', 'company', '330101199005055678', '互联网', '孙总', '13600136005', '合作伙伴推荐', '新客户，知识产权案件', 'active'),
('周八', 'zhouba@example.com', '13800138006', '成都市武侯区', '', 'personal', '510101199006066789', '教育', '', '', '搜索引擎', '个人客户，房产纠纷', 'inactive'),
('吴九', 'wujiu@example.com', '13800138007', '武汉市洪山区', '武汉制造有限公司', 'company', '420101199007077890', '制造业', '吴总', '13500135007', '行业会议', '企业客户，产品质量纠纷', 'active'),
('郑十', 'zhengshi@example.com', '13800138008', '南京市鼓楼区', '', 'personal', '320101199008088901', '医疗', '', '', '社交媒体', '个人客户，医疗纠纷', 'active')
ON DUPLICATE KEY UPDATE name = name;

-- 插入测试案件数据
INSERT INTO cases (title, description, client_id, lawyer_id, case_type, priority, status, start_date, expected_end_date, case_amount, principal_info, opponent_info, case_number) VALUES
('张三与北京科技有限公司合同纠纷案', '客户与公司之间的劳动合同纠纷，涉及加班费和经济补偿金问题', 1, 1, 'civil', 'high', 'active', '2024-01-15', '2024-06-15', 500000.00, '张三，男，34岁，北京市朝阳区，原公司员工', '北京科技有限公司，法定代表人：李总，地址：北京市朝阳区', 'CASE202401150001'),
('李四离婚财产分割案', '客户与配偶离婚，涉及房产、车辆等财产分割问题', 2, 2, 'civil', 'medium', 'pending', '2024-02-01', '2024-08-01', 200000.00, '李四，女，32岁，上海市浦东新区，公司职员', '配偶：王五，男，35岁，上海市浦东新区，公司经理', 'CASE202402010002'),
('王五商标侵权案', '客户公司商标被侵权，需要维权', 3, 1, 'commercial', 'high', 'active', '2024-01-20', '2024-12-20', 800000.00, '广州贸易有限公司，法定代表人：王经理，地址：广州市天河区', '侵权方：某竞争对手公司，地址未知', 'CASE202401200003'),
('赵六劳动仲裁案', '客户与公司之间的劳动纠纷，涉及工资和解雇问题', 4, 3, 'civil', 'medium', 'closed', '2024-01-10', '2024-03-10', 100000.00, '赵六，男，35岁，深圳市南山区，原公司员工', '某科技公司，地址：深圳市南山区', 'CASE202401100004'),
('孙七专利申请案', '客户公司申请发明专利，需要法律支持', 5, 2, 'commercial', 'low', 'active', '2024-03-01', '2024-09-01', 300000.00, '杭州科技有限公司，法定代表人：孙总，地址：杭州市西湖区', '专利局，地址：杭州市西湖区', 'CASE202403010005'),
('周八房产买卖纠纷案', '客户购房过程中遇到的合同纠纷', 6, 3, 'civil', 'high', 'pending', '2024-02-15', '2024-10-15', 1500000.00, '周八，女，40岁，成都市武侯区，公司职员', '开发商：某房地产公司，地址：成都市武侯区', 'CASE202402150006'),
('吴九产品质量纠纷案', '客户公司产品质量问题引起的消费者投诉', 7, 1, 'commercial', 'urgent', 'active', '2024-01-25', '2024-05-25', 600000.00, '武汉制造有限公司，法定代表人：吴总，地址：武汉市洪山区', '消费者：多名客户，地址分散', 'CASE202401250007'),
('郑十医疗事故赔偿案', '客户在医院治疗过程中发生的医疗事故', 8, 2, 'criminal', 'high', 'active', '2024-02-20', '2024-11-20', 1200000.00, '郑十，女，28岁，南京市鼓楼区，公司职员', '某三甲医院，地址：南京市鼓楼区', 'CASE202402200008')
ON DUPLICATE KEY UPDATE title = title;

-- 插入测试文件数据
INSERT INTO files (name, original_name, size, type, path, category, description, uploader_id) VALUES
('合同扫描件.pdf', '张三合同.pdf', 2048576, 'pdf', '/uploads/2024/01/contract_001.pdf', 'document', '张三与公司的劳动合同扫描件', 1),
('身份证复印件.jpg', '李四身份证.jpg', 512000, 'jpg', '/uploads/2024/01/id_card_001.jpg', 'image', '李四的身份证复印件', 1),
('房产证扫描件.pdf', '李四房产证.pdf', 1048576, 'pdf', '/uploads/2024/01/house_cert_001.pdf', 'document', '李四的房产证扫描件', 1),
('商标注册证.pdf', '王五商标证.pdf', 1536000, 'pdf', '/uploads/2024/01/trademark_001.pdf', 'document', '王五公司的商标注册证', 1),
('工资单.xlsx', '赵六工资单.xlsx', 256000, 'excel', '/uploads/2024/01/salary_001.xlsx', 'spreadsheet', '赵六的工资单记录', 1),
('专利申请书.docx', '孙七专利申请.docx', 786432, 'word', '/uploads/2024/01/patent_001.docx', 'document', '孙七公司的专利申请书', 1),
('购房合同.pdf', '周八购房合同.pdf', 3072000, 'pdf', '/uploads/2024/01/house_contract_001.pdf', 'document', '周八的购房合同', 1),
('医疗记录.pdf', '郑十医疗记录.pdf', 4096000, 'pdf', '/uploads/2024/01/medical_001.pdf', 'document', '郑十的医疗记录', 1)
ON DUPLICATE KEY UPDATE name = name;

-- 插入更多用户数据
INSERT INTO users (username, password, email, real_name, phone, role, status) VALUES
('lawyer1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer1@lawoa.com', '张律师', '13800138001', 'lawyer', 'active'),
('lawyer2', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer2@lawoa.com', '李律师', '13800138002', 'lawyer', 'active'),
('lawyer3', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer3@lawoa.com', '王律师', '13800138003', 'lawyer', 'active'),
('assistant1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'assistant1@lawoa.com', '助理小张', '13800138004', 'assistant', 'active'),
('manager1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'manager1@lawoa.com', '客户经理小李', '13800138005', 'manager', 'active')
ON DUPLICATE KEY UPDATE username = username;

-- 为新用户分配角色
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u CROSS JOIN roles r
WHERE u.username IN ('lawyer1', 'lawyer2', 'lawyer3') AND r.role_key = 'lawyer'
ON DUPLICATE KEY UPDATE user_id = user_id;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u CROSS JOIN roles r
WHERE u.username = 'assistant1' AND r.role_key = 'assistant'
ON DUPLICATE KEY UPDATE user_id = user_id;

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u CROSS JOIN roles r
WHERE u.username = 'manager1' AND r.role_key = 'manager'
ON DUPLICATE KEY UPDATE user_id = user_id;