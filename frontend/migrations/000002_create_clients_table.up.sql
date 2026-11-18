-- 创建客户表
CREATE TABLE IF NOT EXISTS clients (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name VARCHAR(100) NOT NULL COMMENT '客户姓名',
    email VARCHAR(100) DEFAULT NULL COMMENT '邮箱地址',
    phone VARCHAR(20) DEFAULT NULL COMMENT '电话号码',
    address VARCHAR(255) DEFAULT NULL COMMENT '地址',
    company VARCHAR(100) DEFAULT NULL COMMENT '公司名称',
    notes TEXT DEFAULT NULL COMMENT '备注信息',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '客户状态',
    
    INDEX idx_clients_deleted_at (deleted_at),
    INDEX idx_clients_email (email),
    INDEX idx_clients_phone (phone),
    INDEX idx_clients_status (status),
    INDEX idx_clients_created_at (created_at),
    INDEX idx_clients_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户表';