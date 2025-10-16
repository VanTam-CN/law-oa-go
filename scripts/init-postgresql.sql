-- PostgreSQL 初始化脚本
-- 创建Law OA Go项目所需的数据库结构

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 创建枚举类型
CREATE TYPE user_role AS ENUM ('admin', 'lawyer', 'user');
CREATE TYPE user_status AS ENUM ('active', 'inactive', 'pending');
CREATE TYPE client_type AS ENUM ('个人', '企业');
CREATE TYPE case_type AS ENUM ('民事', '刑事', '行政', '商事', '劳动', '知识产权', '其他');
CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'urgent');
CREATE TYPE case_status AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived');

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100),
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'user',
    phone VARCHAR(20),
    avatar VARCHAR(255),
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- 创建客户表
CREATE TABLE IF NOT EXISTS clients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    type client_type NOT NULL DEFAULT '个人',
    email VARCHAR(100),
    phone VARCHAR(20),
    address TEXT,
    company VARCHAR(100),
    id_card VARCHAR(18),
    industry VARCHAR(50),
    contact_person VARCHAR(50),
    contact_phone VARCHAR(20),
    source VARCHAR(50),
    notes TEXT,
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_clients_name ON clients(name);
CREATE INDEX idx_clients_type ON clients(type);
CREATE INDEX idx_clients_email ON clients(email);
CREATE INDEX idx_clients_status ON clients(status);
CREATE INDEX idx_clients_deleted_at ON clients(deleted_at);

-- 创建案件表
CREATE TABLE IF NOT EXISTS cases (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    lawyer_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    case_type case_type NOT NULL,
    priority case_priority NOT NULL DEFAULT 'medium',
    status case_status NOT NULL DEFAULT 'pending',
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 创建索引
CREATE INDEX idx_cases_title ON cases(title);
CREATE INDEX idx_cases_client_id ON cases(client_id);
CREATE INDEX idx_cases_lawyer_id ON cases(lawyer_id);
CREATE INDEX idx_cases_case_type ON cases(case_type);
CREATE INDEX idx_cases_priority ON cases(priority);
CREATE INDEX idx_cases_status ON cases(status);
CREATE INDEX idx_cases_deleted_at ON cases(deleted_at);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_clients_updated_at
    BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cases_updated_at
    BEFORE UPDATE ON cases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 插入默认管理员用户
INSERT INTO users (username, name, email, password, role, status)
VALUES
    ('admin', '系统管理员', 'admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'active')
ON CONFLICT (username) DO NOTHING;

-- 插入示例律师用户
INSERT INTO users (username, name, email, password, role, status)
VALUES
    ('lawyer1', '张律师', 'lawyer1@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', 'active'),
    ('lawyer2', '李律师', 'lawyer2@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'lawyer', 'active')
ON CONFLICT (username) DO NOTHING;

-- 插入示例客户数据
INSERT INTO clients (name, type, email, phone, address, status)
VALUES
    ('客户A', '个人', 'clienta@example.com', '13800138000', '北京市朝阳区', 'active'),
    ('客户B', '企业', 'clientb@example.com', '13800138001', '北京市海淀区', 'active')
ON CONFLICT (email) DO NOTHING;

-- 插入示例案件数据
INSERT INTO cases (title, description, client_id, lawyer_id, case_type, priority, status)
VALUES
    ('合同纠纷案', '涉及合同条款解释和违约责任认定', 1, 2, '商事', 'high', 'in_progress'),
    ('劳动争议案', '涉及劳动合同解除和经济补偿', 2, 3, '劳动', 'medium', 'pending')
ON CONFLICT DO NOTHING;

COMMENT ON TABLE users IS '用户表：存储系统用户信息，包括管理员、律师等';
COMMENT ON TABLE clients IS '客户表：存储客户基本信息，包括个人和企业客户';
COMMENT ON TABLE cases IS '案件表：存储案件相关信息';

-- 设置表权限
ALTER TABLE users OWNER TO law_oa_user;
ALTER TABLE clients OWNER TO law_oa_user;
ALTER TABLE cases OWNER TO law_oa_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO law_oa_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO law_oa_user;