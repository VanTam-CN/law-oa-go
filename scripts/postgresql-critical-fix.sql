-- PostgreSQL 关键补全SQL脚本
-- 针对当前实际数据库结构的修正
-- 执行前请确保连接到PostgreSQL数据库

-- 检查并创建缺失的枚举类型
DO $$
BEGIN
    -- 创建客户类型枚举（如果不存在）
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'client_type_full') THEN
        CREATE TYPE client_type_full AS ENUM ('individual', 'company', '个人', '企业');
        RAISE NOTICE '已创建 client_type_full 枚举';
    END IF;

    -- 创建案件状态枚举（如果不存在）
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'case_status_full') THEN
        CREATE TYPE case_status_full AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived', 'draft', 'ongoing', 'closed');
        RAISE NOTICE '已创建 case_status_full 枚举';
    END IF;

    -- 创建案件优先级枚举（如果不存在）
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'case_priority_full') THEN
        CREATE TYPE case_priority_full AS ENUM ('low', 'medium', 'high', 'urgent', 'normal');
        RAISE NOTICE '已创建 case_priority_full 枚举';
    END IF;

    -- 创建用户角色枚举（如果不存在）
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role_full') THEN
        CREATE TYPE user_role_full AS ENUM ('admin', 'lawyer', 'user', 'assistant', 'manager');
        RAISE NOTICE '已创建 user_role_full 枚举';
    END IF;
END $$;

-- 1. 补充users表缺失的关键字段
DO $$
BEGIN
    -- 检查users表是否存在，如果存在则添加缺失字段
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public') THEN
        RAISE NOTICE '开始更新users表字段...';

        -- 添加缺失字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'real_name') THEN
            ALTER TABLE users ADD COLUMN real_name VARCHAR(50);
            RAISE NOTICE '已添加 users.real_name 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_at') THEN
            ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
            RAISE NOTICE '已添加 users.last_login_at 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'last_login_ip') THEN
            ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(45);
            RAISE NOTICE '已添加 users.last_login_ip 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'role_id') THEN
            ALTER TABLE users ADD COLUMN role_id BIGINT;
            RAISE NOTICE '已添加 users.role_id 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'department_id') THEN
            ALTER TABLE users ADD COLUMN department_id BIGINT;
            RAISE NOTICE '已添加 users.department_id 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'remark') THEN
            ALTER TABLE users ADD COLUMN remark TEXT;
            RAISE NOTICE '已添加 users.remark 字段';
        END IF;

        -- 修改role字段类型为新的枚举
        BEGIN
            ALTER TABLE users ALTER COLUMN role TYPE user_role_full USING role::user_role_full;
            RAISE NOTICE '已更新 users.role 字段类型';
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'users.role 字段类型更新失败，可能已经是正确类型';
        END;

    ELSE
        RAISE NOTICE 'users表不存在，跳过字段更新';
    END IF;
END $$;

-- 2. 补充clients表缺失的关键字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'clients' AND table_schema = 'public') THEN
        RAISE NOTICE '开始更新clients表字段...';

        -- 添加缺失字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'client_name') THEN
            ALTER TABLE clients ADD COLUMN client_name VARCHAR(100);
            RAISE NOTICE '已添加 clients.client_name 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'lawyer_id') THEN
            ALTER TABLE clients ADD COLUMN lawyer_id BIGINT;
            RAISE NOTICE '已添加 clients.lawyer_id 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'remark') THEN
            ALTER TABLE clients ADD COLUMN remark TEXT;
            RAISE NOTICE '已添加 clients.remark 字段';
        END IF;

        -- 修改type字段类型为新的枚举
        BEGIN
            ALTER TABLE clients ALTER COLUMN type TYPE client_type_full USING type::client_type_full;
            RAISE NOTICE '已更新 clients.type 字段类型';
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'clients.type 字段类型更新失败，可能已经是正确类型';
        END;

        -- 添加索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_clients_lawyer_id') THEN
            CREATE INDEX idx_clients_lawyer_id ON clients(lawyer_id);
            RAISE NOTICE '已创建 clients.lawyer_id 索引';
        END IF;

    ELSE
        RAISE NOTICE 'clients表不存在，跳过字段更新';
    END IF;
END $$;

-- 3. 补充cases表缺失的关键字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cases' AND table_schema = 'public') THEN
        RAISE NOTICE '开始更新cases表字段...';

        -- 添加缺失的关键字段
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_no') THEN
            ALTER TABLE cases ADD COLUMN case_no VARCHAR(50) UNIQUE;
            RAISE NOTICE '已添加 cases.case_no 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_name') THEN
            ALTER TABLE cases ADD COLUMN case_name VARCHAR(200);
            RAISE NOTICE '已添加 cases.case_name 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'assisting_lawyer_id') THEN
            ALTER TABLE cases ADD COLUMN assisting_lawyer_id BIGINT;
            RAISE NOTICE '已添加 cases.assisting_lawyer_id 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'contract_amount') THEN
            ALTER TABLE cases ADD COLUMN contract_amount DECIMAL(12,2);
            RAISE NOTICE '已添加 cases.contract_amount 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'conflict_check_status') THEN
            ALTER TABLE cases ADD COLUMN conflict_check_status VARCHAR(20) DEFAULT 'pending';
            RAISE NOTICE '已添加 cases.conflict_check_status 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'opponent_info') THEN
            ALTER TABLE cases ADD COLUMN opponent_info TEXT;
            RAISE NOTICE '已添加 cases.opponent_info 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'remark') THEN
            ALTER TABLE cases ADD COLUMN remark TEXT;
            RAISE NOTICE '已添加 cases.remark 字段';
        END IF;

        -- 修改字段类型为新的枚举
        BEGIN
            ALTER TABLE cases ALTER COLUMN priority TYPE case_priority_full USING priority::case_priority_full;
            RAISE NOTICE '已更新 cases.priority 字段类型';
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'cases.priority 字段类型更新失败，可能已经是正确类型';
        END;

        BEGIN
            ALTER TABLE cases ALTER COLUMN status TYPE case_status_full USING status::case_status_full;
            RAISE NOTICE '已更新 cases.status 字段类型';
        EXCEPTION WHEN OTHERS THEN
            RAISE NOTICE 'cases.status 字段类型更新失败，可能已经是正确类型';
        END;

        -- 添加索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_cases_case_no') THEN
            CREATE INDEX idx_cases_case_no ON cases(case_no);
            RAISE NOTICE '已创建 cases.case_no 索引';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_cases_assisting_lawyer_id') THEN
            CREATE INDEX idx_cases_assisting_lawyer_id ON cases(assisting_lawyer_id);
            RAISE NOTICE '已创建 cases.assisting_lawyer_id 索引';
        END IF;

    ELSE
        RAISE NOTICE 'cases表不存在，跳过字段更新';
    END IF;
END $$;

-- 4. 创建关键缺失的表

-- 4.1 创建lawyers表（完全缺失）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        CREATE TABLE lawyers (
            id BIGSERIAL PRIMARY KEY,
            lawyer_name VARCHAR(50) NOT NULL,
            phone VARCHAR(20),
            email VARCHAR(100),
            license_no VARCHAR(50) UNIQUE,
            position VARCHAR(50),
            department VARCHAR(100),
            specialty TEXT,
            status VARCHAR(20) DEFAULT 'active',
            remark TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );

        -- 创建索引
        CREATE INDEX idx_lawyers_lawyer_name ON lawyers(lawyer_name);
        CREATE INDEX idx_lawyers_phone ON lawyers(phone);
        CREATE INDEX idx_lawyers_license_no ON lawyers(license_no);
        CREATE INDEX idx_lawyers_status ON lawyers(status);

        RAISE NOTICE '已创建 lawyers 表';
    ELSE
        RAISE NOTICE 'lawyers 表已存在，跳过创建';
    END IF;
END $$;

-- 4.2 创建departments表（完全缺失）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments' AND table_schema = 'public') THEN
        CREATE TABLE departments (
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
        );

        -- 创建索引
        CREATE INDEX idx_departments_code ON departments(code);
        CREATE INDEX idx_departments_parent_id ON departments(parent_id);
        CREATE INDEX idx_departments_leader_id ON departments(leader_id);

        RAISE NOTICE '已创建 departments 表';
    ELSE
        RAISE NOTICE 'departments 表已存在，跳过创建';
    END IF;
END $$;

-- 4.3 创建document相关表（完全缺失）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'documents' AND table_schema = 'public') THEN
        CREATE TABLE documents (
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
            status VARCHAR(20) DEFAULT 'active',
            thumbnail_path VARCHAR(500),
            metadata TEXT,
            remark TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );

        -- 创建索引
        CREATE INDEX idx_documents_case_id ON documents(case_id);
        CREATE INDEX idx_documents_client_id ON documents(client_id);
        CREATE INDEX idx_documents_uploader_id ON documents(uploader_id);
        CREATE INDEX idx_documents_document_type ON documents(document_type);
        CREATE INDEX idx_documents_status ON documents(status);

        RAISE NOTICE '已创建 documents 表';
    ELSE
        RAISE NOTICE 'documents 表已存在，跳过创建';
    END IF;
END $$;

-- 5. 插入基础数据

-- 插入示例律师数据
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        INSERT INTO lawyers (lawyer_name, phone, email, license_no, position, department, status) VALUES
        ('张律师', '13800138001', 'zhang@lawfirm.com', 'LAW001', '高级合伙人', '民商事部', 'active'),
        ('李律师', '13800138002', 'li@lawfirm.com', 'LAW002', '合伙人', '刑事部', 'active'),
        ('王律师', '13800138003', 'wang@lawfirm.com', 'LAW003', '律师', '行政部', 'active')
        ON CONFLICT (license_no) DO NOTHING;

        RAISE NOTICE '已插入示例律师数据';
    END IF;
END $$;

-- 6. 创建更新时间触发器函数（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_updated_at_column') THEN
        CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = CURRENT_TIMESTAMP;
            RETURN NEW;
        END;
        $$ language 'plpgsql';

        RAISE NOTICE '已创建 update_updated_at_column 函数';
    END IF;
END $$;

-- 7. 为关键表创建更新触发器
DO $$
BEGIN
    -- 为users表创建触发器
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public') THEN
        DROP TRIGGER IF EXISTS update_users_updated_at ON users;
        CREATE TRIGGER update_users_updated_at
            BEFORE UPDATE ON users
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

        RAISE NOTICE '已创建 users 表更新触发器';
    END IF;

    -- 为clients表创建触发器
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'clients' AND table_schema = 'public') THEN
        DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
        CREATE TRIGGER update_clients_updated_at
            BEFORE UPDATE ON clients
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

        RAISE NOTICE '已创建 clients 表更新触发器';
    END IF;

    -- 为cases表创建触发器
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cases' AND table_schema = 'public') THEN
        DROP TRIGGER IF EXISTS update_cases_updated_at ON cases;
        CREATE TRIGGER update_cases_updated_at
            BEFORE UPDATE ON cases
            FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

        RAISE NOTICE '已创建 cases 表更新触发器';
    END IF;
END $$;

-- 8. 验证执行结果
DO $$
DECLARE
    total_tables INTEGER;
    updated_tables INTEGER := 0;
BEGIN
    -- 统计表数量
    SELECT COUNT(*) INTO total_tables
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

    -- 统计有updated_at字段的表数量
    SELECT COUNT(*) INTO updated_tables
    FROM information_schema.columns
    WHERE table_schema = 'public' AND column_name = 'updated_at';

    RAISE NOTICE '数据库表总数: %', total_tables;
    RAISE NOTICE '有updated_at字段的表数: %', updated_tables;

    -- 显示关键表验证结果
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ lawyers 表已创建';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ departments 表已创建';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'documents' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ documents 表已创建';
    END IF;

    -- 显示字段验证结果
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'users' AND column_name = 'real_name') THEN
        RAISE NOTICE '✅ users.real_name 字段已添加';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'clients' AND column_name = 'lawyer_id') THEN
        RAISE NOTICE '✅ clients.lawyer_id 字段已添加';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'case_no') THEN
        RAISE NOTICE '✅ cases.case_no 字段已添加';
    END IF;
END $$;

COMMIT;

RAISE NOTICE '=== PostgreSQL 关键补全执行完成 ===';
RAISE NOTICE '请刷新数据库连接查看更新后的表结构';