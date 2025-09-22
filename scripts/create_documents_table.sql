-- 创建documents表的数据库迁移脚本
-- 运行方式: mysql -u root -p'1q2w#E$R' law_oa < create_documents_table.sql

USE law_oa;

-- 创建documents表
CREATE TABLE IF NOT EXISTS documents (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    filename VARCHAR(255) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    filesize BIGINT UNSIGNED DEFAULT 0,
    mime_type VARCHAR(100),
    category VARCHAR(100),
    tags VARCHAR(500),
    entity_id BIGINT UNSIGNED DEFAULT 0,
    entity_type VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active',
    
    -- 添加索引
    INDEX idx_entity (entity_id, entity_type),
    INDEX idx_category (category),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入一些测试数据
INSERT INTO documents (name, description, filename, filepath, filesize, mime_type, category, tags, entity_id, entity_type, status) VALUES
('合同模板-标准版', '标准合同模板，适用于一般业务场景', 'standard_contract.pdf', '/uploads/documents/standard_contract.pdf', 1024000, 'application/pdf', 'contract', '合同,模板,标准', 1, 'case', 'active'),
('客户身份证复印件', '客户身份证扫描件', 'customer_id_card.jpg', '/uploads/documents/customer_id_card.jpg', 512000, 'image/jpeg', 'identification', '身份证,客户', 1, 'client', 'active'),
('法律意见书', '关于XX案件的法律意见', 'legal_opinion.docx', '/uploads/documents/legal_opinion.docx', 2048000, 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'legal_opinion', '法律意见,案件', 2, 'case', 'active'),
('证据清单', '案件证据材料清单', 'evidence_list.xlsx', '/uploads/documents/evidence_list.xlsx', 307200, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', 'evidence', '证据,清单', 2, 'case', 'active'),
('委托协议', '与客户签订的委托协议', 'agency_agreement.pdf', '/uploads/documents/agency_agreement.pdf', 1536000, 'application/pdf', 'agreement', '委托,协议', 3, 'case', 'active'),
('案件相关照片', '案件现场照片', 'case_photos.zip', '/uploads/documents/case_photos.zip', 5120000, 'application/zip', 'media', '照片,现场', 3, 'case', 'active'),
('法院传票', '法院送达的传票文件', 'court_summons.pdf', '/uploads/documents/court_summons.pdf', 256000, 'application/pdf', 'legal_document', '传票,法院', 4, 'case', 'active');

-- 显示创建结果
SELECT 'Documents table created successfully' as message;
SELECT COUNT(*) as document_count FROM documents;