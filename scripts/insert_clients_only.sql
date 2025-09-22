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