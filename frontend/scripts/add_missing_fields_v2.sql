-- 添加缺失字段到clients表
ALTER TABLE clients
ADD COLUMN type VARCHAR(20) DEFAULT 'personal' COMMENT '客户类型：personal/company' AFTER contact_person,
ADD COLUMN id_card_number VARCHAR(20) DEFAULT NULL COMMENT '身份证号' AFTER type,
ADD COLUMN industry VARCHAR(50) DEFAULT NULL COMMENT '行业' AFTER id_card_number,
ADD COLUMN contact_phone VARCHAR(20) DEFAULT NULL COMMENT '联系电话' AFTER industry,
ADD COLUMN source VARCHAR(50) DEFAULT NULL COMMENT '客户来源' AFTER contact_phone;

-- 添加缺失字段到cases表
ALTER TABLE cases
ADD COLUMN case_number VARCHAR(50) DEFAULT NULL COMMENT '案件编号' AFTER end_date,
ADD COLUMN case_amount DECIMAL(12,2) DEFAULT NULL COMMENT '案件金额' AFTER case_number,
ADD COLUMN expected_end_date TIMESTAMP NULL DEFAULT NULL COMMENT '预计结束日期' AFTER case_amount,
ADD COLUMN principal_info TEXT DEFAULT NULL COMMENT '当事人信息' AFTER expected_end_date,
ADD COLUMN opponent_info TEXT DEFAULT NULL COMMENT '对方当事人信息' AFTER principal_info;