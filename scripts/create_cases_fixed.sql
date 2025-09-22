-- 创建案件表
CREATE TABLE IF NOT EXISTS cases (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    title VARCHAR(200) NOT NULL COMMENT '案件标题',
    description TEXT DEFAULT NULL COMMENT '案件描述',
    client_id INT UNSIGNED NOT NULL COMMENT '客户ID',
    lawyer_id INT UNSIGNED NOT NULL COMMENT '律师ID',
    case_type VARCHAR(50) NOT NULL COMMENT '案件类型',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT '优先级',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '案件状态',
    start_date TIMESTAMP NULL DEFAULT NULL COMMENT '开始日期',
    end_date TIMESTAMP NULL DEFAULT NULL COMMENT '结束日期',

    INDEX idx_cases_deleted_at (deleted_at),
    INDEX idx_cases_client_id (client_id),
    INDEX idx_cases_lawyer_id (lawyer_id),
    INDEX idx_cases_case_type (case_type),
    INDEX idx_cases_priority (priority),
    INDEX idx_cases_status (status),
    INDEX idx_cases_created_at (created_at),
    INDEX idx_cases_start_date (start_date),

    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,
    FOREIGN KEY (lawyer_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='案件表';