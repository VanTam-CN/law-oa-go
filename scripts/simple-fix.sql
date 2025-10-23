-- Law OA Go 简单数据库修复脚本

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
ON CONFLICT DO NOTHING;

-- 修复用户表约束
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
ALTER TABLE "users" ADD CONSTRAINT "uni_users_username" UNIQUE ("username");

-- 创建必要的索引
CREATE INDEX IF NOT EXISTS "idx_users_username" ON "users" ("username");
CREATE INDEX IF NOT EXISTS "idx_users_email" ON "users" ("email");
CREATE INDEX IF NOT EXISTS "idx_users_status" ON "users" ("status");

CREATE INDEX IF NOT EXISTS "idx_user_roles_user_id" ON "user_roles" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_user_roles_role_id" ON "user_roles" ("role_id");

-- 创建法律法规表的基本索引
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_title" ON "legal_statutes" ("title");
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_content" ON "legal_statutes" ("content");
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_category" ON "legal_statutes" ("category_id");
CREATE INDEX IF NOT EXISTS "idx_legal_statutes_status" ON "legal_statutes" ("status");

-- 输出修复结果
DO $$
BEGIN
    RAISE NOTICE '数据库基本修复完成';
    RAISE NOTICE '角色数量: %', (SELECT COUNT(*) FROM "roles");
    RAISE NOTICE '用户数量: %', (SELECT COUNT(*) FROM "users");
    RAISE NOTICE '法规数量: %', (SELECT COUNT(*) FROM "legal_statutes");
END $$;