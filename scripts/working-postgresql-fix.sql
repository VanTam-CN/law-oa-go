-- PostgreSQL 修正版修复脚本
-- 解决语法错误，确保成功执行

-- 检查当前数据库版本
SELECT version() as current_version;

-- 1. 创建缺失的枚举类型
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'client_type_enum') THEN
        CREATE TYPE client_type_enum AS ENUM ('individual', 'company', '个人', '企业');
        RAISE NOTICE '✅ 创建 client_type_enum 枚举';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'case_status_enum') THEN
        CREATE TYPE case_status_enum AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived', 'draft', 'ongoing', 'closed');
        RAISE NOTICE '✅ 创建 case_status_enum 枚举';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'case_priority_enum') THEN
        CREATE TYPE case_priority_enum AS ENUM ('low', 'medium', 'high', 'urgent', 'normal');
        RAISE NOTICE '✅ 创建 case_priority_enum 枚举';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_enum') THEN
        CREATE TYPE user_role_enum AS ENUM ('admin', 'lawyer', 'user', 'assistant', 'manager');
        RAISE NOTICE '✅ 创建 user_role_enum 枚举';
    END IF;
END $$;

-- 2. 为现有表添加缺失的字段
-- 2.1 更新users表
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public') THEN
        -- 添加real_name字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'real_name') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN real_name VARCHAR(50)';
            RAISE NOTICE '✅ 添加 users.real_name 字段';
        END IF;

        -- 添加last_login_at字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_at') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE';
            RAISE NOTICE '✅ 添加 users.last_login_at 字段';
        END IF;

        -- 添加last_login_ip字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_ip') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(45)';
            RAISE NOTICE '✅ 添加 users.last_login_ip 字段';
        END IF;

        -- 添加role_id字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role_id') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN role_id BIGINT';
            RAISE NOTICE '✅ 添加 users.role_id 字段';
        END IF;

        -- 添加department_id字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'department_id') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN department_id BIGINT';
            RAISE NOTICE '✅ 添加 users.department_id 字段';
        END IF;

        -- 添加remark字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'remark') THEN
            EXECUTE 'ALTER TABLE users ADD COLUMN remark TEXT';
            RAISE NOTICE '✅ 添加 users.remark 字段';
        END IF;
    END IF;
END $$;

-- 2.2 更新clients表
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'clients' AND table_schema = 'public') THEN
        -- 添加client_name字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'client_name') THEN
            EXECUTE 'ALTER TABLE clients ADD COLUMN client_name VARCHAR(100)';
            RAISE NOTICE '✅ 添加 clients.client_name 字段';
        END IF;

        -- 添加lawyer_id字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'lawyer_id') THEN
            EXECUTE 'ALTER TABLE clients ADD COLUMN lawyer_id BIGINT';
            RAISE NOTICE '✅ 添加 clients.lawyer_id 字段';
        END IF;

        -- 添加remark字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'remark') THEN
            EXECUTE 'ALTER TABLE clients ADD COLUMN remark TEXT';
            RAISE NOTICE '✅ 添加 clients.remark 字段';
        END IF;

        -- 创建lawyer_id索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_clients_lawyer_id') THEN
            EXECUTE 'CREATE INDEX idx_clients_lawyer_id ON clients(lawyer_id)';
            RAISE NOTICE '✅ 创建 clients.lawyer_id 索引';
        END IF;
    END IF;
END $$;

-- 2.3 更新cases表
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cases' AND table_schema = 'public') THEN
        -- 添加case_no字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_no') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN case_no VARCHAR(50)';
            RAISE NOTICE '✅ 添加 cases.case_no 字段';
        END IF;

        -- 添加case_name字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_name') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN case_name VARCHAR(200)';
            RAISE NOTICE '✅ 添加 cases.case_name 字段';
        END IF;

        -- 添加assisting_lawyer_id字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'assisting_lawyer_id') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN assisting_lawyer_id BIGINT';
            RAISE NOTICE '✅ 添加 cases.assisting_lawyer_id 字段';
        END IF;

        -- 添加contract_amount字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'contract_amount') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN contract_amount DECIMAL(12,2)';
            RAISE NOTICE '✅ 添加 cases.contract_amount 字段';
        END IF;

        -- 添加opponent_info字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'opponent_info') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN opponent_info TEXT';
            RAISE NOTICE '✅ 添加 cases.opponent_info 字段';
        END IF;

        -- 添加remark字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'remark') THEN
            EXECUTE 'ALTER TABLE cases ADD COLUMN remark TEXT';
            RAISE NOTICE '✅ 添加 cases.remark 字段';
        END IF;

        -- 创建assisting_lawyer_id索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_cases_assisting_lawyer_id') THEN
            EXECUTE 'CREATE INDEX idx_cases_assisting_lawyer_id ON cases(assisting_lawyer_id)';
            RAISE NOTICE '✅ 创建 cases.assisting_lawyer_id 索引';
        END IF;
    END IF;
END $$;

-- 3. 创建缺失的关键表
-- 3.1 创建lawyers表
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        EXECUTE 'CREATE TABLE lawyers (
            id BIGSERIAL PRIMARY KEY,
            lawyer_name VARCHAR(50) NOT NULL,
            phone VARCHAR(20),
            email VARCHAR(100),
            license_no VARCHAR(50) UNIQUE,
            position VARCHAR(50),
            department VARCHAR(100),
            specialty TEXT,
            status VARCHAR(20) DEFAULT ''active'',
            remark TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        )';

        -- 创建索引
        EXECUTE 'CREATE INDEX idx_lawyers_lawyer_name ON lawyers(lawyer_name)';
        EXECUTE 'CREATE INDEX idx_lawyers_phone ON lawyers(phone)';
        EXECUTE 'CREATE INDEX idx_lawyers_license_no ON lawyers(license_no)';
        EXECUTE 'CREATE INDEX idx_lawyers_status ON lawyers(status)';

        RAISE NOTICE '✅ 创建 lawyers 表';
    END IF;
END $$;

-- 3.2 创建departments表
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments' AND table_schema = 'public') THEN
        EXECUTE 'CREATE TABLE departments (
            id BIGSERIAL PRIMARY KEY,
            name VARCHAR(100) NOT NULL,
            code VARCHAR(50) NOT NULL UNIQUE,
            parent_id BIGINT DEFAULT 0,
            leader_id BIGINT,
            description TEXT,
            sort_order INTEGER DEFAULT 0,
            status INTEGER DEFAULT 1,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        )';

        -- 创建索引
        EXECUTE 'CREATE INDEX idx_departments_code ON departments(code)';
        EXECUTE 'CREATE INDEX idx_departments_parent_id ON departments(parent_id)';
        EXECUTE 'CREATE INDEX idx_departments_leader_id ON departments(leader_id)';

        RAISE NOTICE '✅ 创建 departments 表';
    END IF;
END $$;

-- 3.3 创建documents表
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'documents' AND table_schema = 'public') THEN
        EXECUTE 'CREATE TABLE documents (
            id BIGSERIAL PRIMARY KEY,
            document_no VARCHAR(50) UNIQUE NOT NULL,
            case_id BIGINT NOT NULL,
            client_id BIGINT,
            file_name VARCHAR(255) NOT NULL,
            original_name VARCHAR(255) NOT NULL,
            file_size BIGINT NOT NULL,
            file_type VARCHAR(100) NOT NULL,
            file_hash VARCHAR(64) NOT NULL,
            file_path VARCHAR(500) NOT NULL,
            document_type VARCHAR(50) NOT NULL,
            description TEXT,
            tags TEXT,
            is_public BOOLEAN DEFAULT FALSE,
            is_confidential BOOLEAN DEFAULT FALSE,
            uploader_id BIGINT NOT NULL,
            upload_time TIMESTAMP WITH TIME ZONE NOT NULL,
            download_count INTEGER DEFAULT 0,
            last_download_time TIMESTAMP WITH TIME ZONE,
            status VARCHAR(20) DEFAULT ''active'',
            thumbnail_path VARCHAR(500),
            metadata TEXT,
            remark TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        )';

        -- 创建索引
        EXECUTE 'CREATE INDEX idx_documents_case_id ON documents(case_id)';
        EXECUTE 'CREATE INDEX idx_documents_client_id ON documents(client_id)';
        EXECUTE 'CREATE INDEX idx_documents_uploader_id ON documents(uploader_id)';
        EXECUTE 'CREATE INDEX idx_documents_document_type ON documents(document_type)';
        EXECUTE 'CREATE INDEX idx_documents_status ON documents(status)';

        RAISE NOTICE '✅ 创建 documents 表';
    END IF;
END $$;

-- 4. 插入示例数据
DO $$
BEGIN
    -- 插入律师示例数据
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        EXECUTE 'INSERT INTO lawyers (lawyer_name, phone, email, license_no, position, department, status) VALUES
            (''张律师'', ''13800138001'', ''zhang@lawfirm.com'', ''LAW001'', ''高级合伙人'', ''民商事部'', ''active''),
            (''李律师'', ''13800138002'', ''li@lawfirm.com'', ''LAW002'', ''合伙人'', ''刑事部'', ''active''),
            (''王律师'', ''13800138003'', ''wang@lawfirm.com'', ''LAW003'', ''律师'', ''行政部'', ''active'')
            ON CONFLICT (license_no) DO NOTHING';

        RAISE NOTICE '✅ 插入律师示例数据';
    END IF;
END $$;

-- 5. 验证结果
DO $$
DECLARE
    total_tables INTEGER;
    lawyers_count INTEGER;
BEGIN
    -- 统计表数量
    SELECT COUNT(*) INTO total_tables
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

    RAISE NOTICE '📊 数据库表总数: %', total_tables;

    -- 检查关键表
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ lawyers 表创建成功';

        SELECT COUNT(*) INTO lawyers_count FROM lawyers;
        RAISE NOTICE '📊 律师数量: %', lawyers_count;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ departments 表创建成功';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'documents' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ documents 表创建成功';
    END IF;

    -- 检查字段
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'real_name') THEN
        RAISE NOTICE '✅ users.real_name 字段添加成功';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'lawyer_id') THEN
        RAISE NOTICE '✅ clients.lawyer_id 字段添加成功';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_no') THEN
        RAISE NOTICE '✅ cases.case_no 字段添加成功';
    END IF;

    RAISE NOTICE '🎉 PostgreSQL数据库修复完成！请刷新数据库管理工具查看结果。';
END $$;

SELECT 'PostgreSQL数据库修复完成' as status;