-- 创建文档表
CREATE TABLE IF NOT EXISTS documents (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    name VARCHAR(200) NOT NULL COMMENT '文档名称',
    description TEXT DEFAULT NULL COMMENT '文档描述',
    filename VARCHAR(255) NOT NULL COMMENT '文件名',
    filepath VARCHAR(500) NOT NULL COMMENT '文件路径',
    filesize BIGINT UNSIGNED DEFAULT 0 COMMENT '文件大小（字节）',
    mime_type VARCHAR(100) DEFAULT NULL COMMENT 'MIME类型',
    category VARCHAR(100) DEFAULT NULL COMMENT '文档分类',
    tags VARCHAR(500) DEFAULT NULL COMMENT '标签（JSON格式）',
    entity_id BIGINT UNSIGNED DEFAULT NULL COMMENT '关联实体ID',
    entity_type VARCHAR(50) DEFAULT NULL COMMENT '关联实体类型',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '文档状态',
    
    INDEX idx_documents_deleted_at (deleted_at),
    INDEX idx_documents_category (category),
    INDEX idx_documents_entity (entity_type, entity_id),
    INDEX idx_documents_status (status),
    INDEX idx_documents_created_at (created_at),
    INDEX idx_documents_filesize (filesize),
    INDEX idx_documents_mime_type (mime_type),
    
    FULLTEXT INDEX ft_documents_name_description_tags (name, description, tags)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文档表';