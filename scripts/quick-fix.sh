#!/bin/bash

echo "🚨 PostgreSQL 快速修复脚本"
echo "=================================="

# 设置变量
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-law_oa_go}

echo "数据库连接信息:"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "User: $DB_USER"
echo "Database: $DB_NAME"
echo ""

# 检查psql命令
if ! command -v psql &> /dev/null; then
    echo "❌ psql命令未找到，请安装PostgreSQL客户端"
    echo "Ubuntu/Debian: sudo apt-get install postgresql-client"
    echo "macOS: brew install postgresql"
    exit 1
fi

echo "✅ 找到psql命令"

# 测试数据库连接
PGPASSWORD="$DB_PASSWORD"
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" &> /dev/null; then
    echo "❌ 无法连接到PostgreSQL数据库"
    echo "请检查："
    echo "1. PostgreSQL服务是否启动"
    echo "2. 连接参数是否正确"
    echo "3. 用户权限是否足够"
    exit 1
fi

echo "✅ 数据库连接成功"
echo ""

# 执行修复SQL
echo "🛠️ 执行数据库修复..."

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
-- 创建缺失的枚举类型
DO $$ BEGIN
    CREATE TYPE client_type_full AS ENUM ('individual', 'company', '个人', '企业');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE case_status_full AS ENUM ('pending', 'in_progress', 'completed', 'cancelled', 'archived', 'draft', 'ongoing', 'closed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE case_priority_full AS ENUM ('low', 'medium', 'high', 'urgent', 'normal');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE user_role_full AS ENUM ('admin', 'lawyer', 'user', 'assistant', 'manager');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- 1. 更新users表添加缺失字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users' AND table_schema = 'public') THEN
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
    END IF;
END $$;

-- 2. 更新clients表添加缺失字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'clients' AND table_schema = 'public') THEN
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

        -- 创建索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_clients_lawyer_id') THEN
            CREATE INDEX idx_clients_lawyer_id ON clients(lawyer_id);
            RAISE NOTICE '已创建 clients.lawyer_id 索引';
        END IF;
    END IF;
END $$;

-- 3. 更新cases表添加缺失字段
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'cases' AND table_schema = 'public') THEN
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

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'opponent_info') THEN
            ALTER TABLE cases ADD COLUMN opponent_info TEXT;
            RAISE NOTICE '已添加 cases.opponent_info 字段';
        END IF;

        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'cases' AND column_name = 'remark') THEN
            ALTER TABLE cases ADD COLUMN remark TEXT;
            RAISE NOTICE '已添加 cases.remark 字段';
        END IF;

        -- 创建索引
        IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_cases_assisting_lawyer_id') THEN
            CREATE INDEX idx_cases_assisting_lawyer_id ON cases(assisting_lawyer_id);
            RAISE NOTICE '已创建 cases.assisting_lawyer_id 索引';
        END IF;
    END IF;
END $$;

-- 4. 创建lawyers表
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
    END IF;
END $$;

-- 5. 创建departments表
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
    END IF;
END $$;

-- 6. 创建documents表
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
    END IF;
END $$;

-- 7. 插入示例数据
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

-- 8. 验证修复结果
DO $$
DECLARE
    table_count INTEGER := 0;
    lawyer_count INTEGER := 0;
BEGIN
    -- 统计表数量
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

    RAISE NOTICE '数据库表总数: %', table_count;

    -- 检查关键表
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'lawyers' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ lawyers 表已创建';

        SELECT COUNT(*) INTO lawyer_count FROM lawyers;
        RAISE NOTICE '律师数量: %', lawyer_count;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'departments' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ departments 表已创建';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'documents' AND table_schema = 'public') THEN
        RAISE NOTICE '✅ documents 表已创建';
    END IF;

    -- 检查字段
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

SELECT '✅ PostgreSQL数据库修复完成！' as result;
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 PostgreSQL数据库修复成功完成！"
    echo ""
    echo "📊 修复内容："
    echo "✅ 添加了缺失的枚举类型"
    echo "✅ 为users表添加了6个字段"
    echo "✅ 为clients表添加了3个字段"
    echo "✅ 为cases表添加了6个字段"
    echo "✅ 创建了lawyers表"
    echo "✅ 创建了departments表"
    echo "✅ 创建了documents表"
    echo "✅ 插入了示例律师数据"
    echo ""
    echo "🔍 下一步："
    echo "1. 重新刷新数据库管理工具查看结构"
    echo "2. 测试Go应用程序连接"
    echo "3. 验证API功能正常"
else
    echo ""
    echo "❌ 数据库修复失败，请检查错误信息"
    exit 1
fi