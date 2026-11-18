-- Law OA Go 数据库迁移文件
-- 版本: 000001
-- 描述: 创建基础表结构

-- 创建用户角色枚举
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('admin', 'lawyer', 'user');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建用户状态枚举
DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'inactive', 'pending');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建客户类型枚举
DO $$ BEGIN
    CREATE TYPE client_type AS ENUM ('个人', '企业');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建案件类型枚举
DO $$ BEGIN
    CREATE TYPE case_type AS ENUM ('民事', '刑事', '行政', '商事', '劳动', '知识产权', '其他');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建案件优先级枚举
DO $$ BEGIN
    CREATE TYPE case_priority AS ENUM ('low', 'medium', 'high', 'urgent');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 创建案件状态枚举
DO $$ BEGIN
    CREATE TYPE case_status AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

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

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
CREATE TRIGGER update_clients_updated_at
    BEFORE UPDATE ON clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_cases_updated_at ON cases;
CREATE TRIGGER update_cases_updated_at
    BEFORE UPDATE ON cases
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();