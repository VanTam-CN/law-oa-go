-- 添加客户表缺失字段
ALTER TABLE clients
ADD COLUMN type VARCHAR(20) DEFAULT 'personal' COMMENT '客户类型：personal/company' AFTER notes,
ADD COLUMN id_card VARCHAR(20) DEFAULT NULL COMMENT '身份证号' AFTER type,
ADD COLUMN industry VARCHAR(100) DEFAULT NULL COMMENT '所属行业' AFTER id_card,
ADD COLUMN contact_person VARCHAR(50) DEFAULT NULL COMMENT '联系人' AFTER industry,
ADD COLUMN contact_phone VARCHAR(20) DEFAULT NULL COMMENT '联系电话' AFTER contact_person,
ADD COLUMN source VARCHAR(100) DEFAULT NULL COMMENT '客户来源' AFTER contact_phone;

-- 添加案件表缺失字段
ALTER TABLE cases
ADD COLUMN case_number VARCHAR(50) DEFAULT NULL COMMENT '案件编号' AFTER id,
ADD COLUMN case_amount DECIMAL(12,2) DEFAULT NULL COMMENT '案件金额' AFTER status,
ADD COLUMN expected_end_date TIMESTAMP NULL DEFAULT NULL COMMENT '预计结束日期' AFTER end_date,
ADD COLUMN principal_info TEXT DEFAULT NULL COMMENT '当事人信息' AFTER expected_end_date,
ADD COLUMN opponent_info TEXT DEFAULT NULL COMMENT '对方信息' AFTER principal_info;

-- 添加文件表（如果不存在）
CREATE TABLE IF NOT EXISTS files (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name VARCHAR(255) NOT NULL COMMENT '文件名',
    original_name VARCHAR(255) NOT NULL COMMENT '原始文件名',
    size BIGINT UNSIGNED NOT NULL COMMENT '文件大小（字节）',
    type VARCHAR(100) NOT NULL COMMENT '文件类型',
    path VARCHAR(500) NOT NULL COMMENT '文件路径',
    category VARCHAR(50) DEFAULT 'other' COMMENT '文件分类',
    description TEXT DEFAULT NULL COMMENT '文件描述',
    uploader_id BIGINT UNSIGNED NOT NULL COMMENT '上传者ID',
    download_count INT UNSIGNED DEFAULT 0 COMMENT '下载次数',

    INDEX idx_files_deleted_at (deleted_at),
    INDEX idx_files_category (category),
    INDEX idx_files_type (type),
    INDEX idx_files_uploader_id (uploader_id),
    INDEX idx_files_created_at (created_at),

    FOREIGN KEY (uploader_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表';

-- 更新现有客户的type字段（基于company字段判断）
UPDATE clients
SET type = 'company'
WHERE company IS NOT NULL AND company != '';

-- 为现有案件生成案件编号
UPDATE cases
SET case_number = CONCAT('CASE', DATE_FORMAT(created_at, '%Y%m%d'), LPAD(id, 4, '0'))
WHERE case_number IS NULL OR case_number = '';