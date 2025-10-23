-- 创建数据库
CREATE DATABASE IF NOT EXISTS document_service CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE document_service;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    tenant_id VARCHAR(64) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_users_tenant (tenant_id),
    INDEX idx_users_username (username),
    INDEX idx_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_roles_name_tenant (name, tenant_id),
    INDEX idx_roles_tenant (tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_roles_user_role (user_id, role_id),
    INDEX idx_user_roles_user (user_id),
    INDEX idx_user_roles_role (role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 文档表
CREATE TABLE IF NOT EXISTS documents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    tenant_id VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    original_name VARCHAR(255),
    mime_type VARCHAR(100) NOT NULL,
    size BIGINT NOT NULL,
    category VARCHAR(50),
    tags TEXT,
    entity_type VARCHAR(50),
    entity_id BIGINT,
    current_version INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    INDEX idx_documents_tenant (tenant_id),
    INDEX idx_documents_category (category),
    INDEX idx_documents_entity (entity_type, entity_id),
    INDEX idx_documents_status (status),
    INDEX idx_documents_created_by (created_by),
    INDEX idx_documents_deleted_at (deleted_at),
    INDEX idx_documents_uuid (uuid),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 文档版本表
CREATE TABLE IF NOT EXISTS document_versions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    document_id BIGINT NOT NULL,
    version INT NOT NULL,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    storage_path VARCHAR(512) NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    size BIGINT NOT NULL,
    description TEXT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_document_versions_document (document_id),
    INDEX idx_document_versions_hash (file_hash),
    INDEX idx_document_versions_uuid (uuid),
    UNIQUE KEY uk_document_versions_doc_version (document_id, version),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 文档权限表
CREATE TABLE IF NOT EXISTS document_permissions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    document_id BIGINT NOT NULL,
    user_id BIGINT,
    role_id BIGINT,
    tenant_id VARCHAR(64) NOT NULL,
    permission VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_document_permissions_document (document_id),
    INDEX idx_document_permissions_user (user_id),
    INDEX idx_document_permissions_role (role_id),
    INDEX idx_document_permissions_tenant (tenant_id),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT chk_document_permissions_user_or_role CHECK (
        (user_id IS NOT NULL AND role_id IS NULL) OR
        (user_id IS NULL AND role_id IS NOT NULL) OR
        (user_id IS NOT NULL AND role_id IS NOT NULL)
    )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 文档审计日志表
CREATE TABLE IF NOT EXISTS document_audits (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    document_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    tenant_id VARCHAR(64) NOT NULL,
    action VARCHAR(50) NOT NULL,
    details TEXT,
    ip_address VARCHAR(45),
    user_agent VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_document_audits_document (document_id),
    INDEX idx_document_audits_user (user_id),
    INDEX idx_document_audits_tenant (tenant_id),
    INDEX idx_document_audits_action (action),
    INDEX idx_document_audits_created_at (created_at),
    FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建默认角色
INSERT INTO roles (name, tenant_id, is_default) VALUES
('admin', 'default', TRUE),
('lawyer', 'default', FALSE),
('assistant', 'default', FALSE),
('viewer', 'default', FALSE)
ON DUPLICATE KEY UPDATE name = VALUES(name);

-- 创建默认用户 (开发环境)
INSERT INTO users (username, email, tenant_id, is_active) VALUES
('admin', 'admin@law-oa-go.com', 'default', TRUE),
('lawyer1', 'lawyer1@law-oa-go.com', 'default', TRUE),
('assistant1', 'assistant1@law-oa-go.com', 'default', TRUE)
ON DUPLICATE KEY UPDATE username = VALUES(username);

-- 分配默认用户角色
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin' AND u.tenant_id = r.tenant_id
ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'lawyer1' AND r.name = 'lawyer' AND u.tenant_id = r.tenant_id
ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'assistant1' AND r.name = 'assistant' AND u.tenant_id = r.tenant_id
ON DUPLICATE KEY UPDATE user_id = VALUES(user_id);