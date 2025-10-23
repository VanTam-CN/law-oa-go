-- Law OA Go 开发环境数据库初始化脚本
-- 修复数据库结构和字符编码问题

-- 设置数据库编码
SET client_encoding = 'UTF8';
SET client_min_messages = 'warning';

-- 创建角色表（如果不存在）
CREATE TABLE IF NOT EXISTS "roles" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(50) NOT NULL UNIQUE,
    "description" TEXT,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "deleted_at" TIMESTAMP WITH TIME ZONE
);

-- 创建用户角色关联表（如果不存在）
CREATE TABLE IF NOT EXISTS "user_roles" (
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL,
    "role_id" INTEGER NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE CASCADE,
    FOREIGN KEY ("role_id") REFERENCES "roles"("id") ON DELETE CASCADE,
    UNIQUE("user_id", "role_id")
);

-- 插入默认角色
INSERT INTO "roles" ("name", "description") VALUES
    ('admin', '系统管理员'),
    ('lawyer', '律师'),
    ('assistant', '助理'),
    ('user', '普通用户')
ON CONFLICT ("name") DO NOTHING;

-- 修复用户表约束
-- 检查约束是否存在并删除
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'uni_users_username'
        AND table_name = 'users'
    ) THEN
        ALTER TABLE "users" DROP CONSTRAINT "uni_users_username";
    END IF;
END $$;

-- 添加新的唯一约束
ALTER TABLE "users" ADD CONSTRAINT IF NOT EXISTS "uni_users_username" UNIQUE ("username");

-- 修复权限表（如果存在）
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions') THEN
        -- 修复权限表的字符编码问题
        ALTER TABLE "permissions"
        ALTER COLUMN "name" TYPE VARCHAR(255),
        ALTER COLUMN "resource" TYPE VARCHAR(255),
        ALTER COLUMN "action" TYPE VARCHAR(50);
    END IF;
END $$;

-- 修复法规表字符编码问题
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'legal_statutes') THEN
        -- 修复tags字段，确保其为JSON类型
        ALTER TABLE "legal_statutes"
        ALTER COLUMN "tags" TYPE JSON USING (CASE
            WHEN "tags" IS NULL THEN 'null'::JSON
            WHEN "tags" = '' THEN '[]'::JSON
            ELSE
                -- 尝试清理和修复现有数据
                REPLACE(
                    REPLACE(
                        REPLACE("tags"::TEXT, 'å', ''),
                        'é', ''
                    ),
                    'ç', ''
                )::JSON
        END);
    END IF;
END $$;

-- 创建必要的索引
CREATE INDEX IF NOT EXISTS "idx_users_username" ON "users" ("username");
CREATE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
CREATE INDEX IF NOT EXISTS "idx_users_status" ON "users" ("status");
CREATE INDEX IF NOT EXISTS "idx_users_deleted_at" ON "users" ("deleted_at");

CREATE INDEX IF NOT EXISTS "idx_user_roles_user_id" ON "user_roles" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_user_roles_role_id" ON "user_roles" ("role_id");

CREATE INDEX IF NOT EXISTS "idx_legal_statutes_title" ON "legal_statutes" USING gin(to_tsvector('chinese', "title"));
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_content" ON "legal_statutes" USING gin(to_tsvector('chinese', "content"));
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_category" ON "legal_statutes" ("category_id");
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_status" ON "legal_statutes" ("status");

-- 更新时间戳
UPDATE "users" SET "updated_at" = CURRENT_TIMESTAMP WHERE "updated_at" IS NULL;
UPDATE "legal_statutes" SET "updated_at" = CURRENT_TIMESTAMP WHERE "updated_at" IS NULL;

-- 修复Elasticsearch映射数据
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'elasticsearch_sync') THEN
        -- 清理同步数据，重新同步
        TRUNCATE TABLE "elasticsearch_sync";
    END IF;
END $$;

-- 输出修复结果
DO $$
BEGIN
    RAISE NOTICE '数据库初始化完成';
    RAISE NOTICE '角色数量: %', (SELECT COUNT(*) FROM "roles");
    RAISE NOTICE '用户数量: %', (SELECT COUNT(*) FROM "users" WHERE "deleted_at" IS NULL);
    RAISE NOTICE '法规数量: %', (SELECT COUNT(*) FROM "legal_statutes" WHERE "deleted_at" IS NULL);
END $$;