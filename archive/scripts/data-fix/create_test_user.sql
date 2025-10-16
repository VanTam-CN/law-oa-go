-- Law OA Go 系统测试用户创建脚本
-- 用于API测试和功能验证

-- 清理现有测试用户（如果存在）
DELETE FROM users WHERE email = 'admin@lawfirm.com';
DELETE FROM users WHERE email = 'test@lawfirm.com';

-- 创建管理员测试用户
INSERT INTO users (name, email, password, role, status, created_at, updated_at) VALUES
('管理员', 'admin@lawfirm.com', '$2a$10$rOzJqQjQjQjQjQjQjQjQjOzJqQjQjQjQjQjQjQjQjQjQjQjQjQjQjQjQ', 'admin', 'active', NOW(), NOW());

-- 创建普通用户测试用户
INSERT INTO users (name, email, password, role, status, created_at, updated_at) VALUES
('测试用户', 'test@lawfirm.com', '$2a$10$rOzJqQjQjQjQjQjQjQjQjOzJqQjQjQjQjQjQjQjQjQjQjQjQjQjQjQjQ', 'user', 'active', NOW(), NOW());

-- 注意：上面的密码哈希值是示例，实际使用时应该用真实的bcrypt哈希
-- 或者使用应用程序的注册接口来创建用户

-- 验证用户创建
SELECT id, name, email, role, status, created_at FROM users WHERE email IN ('admin@lawfirm.com', 'test@lawfirm.com');

-- 显示创建的用户数量
SELECT COUNT(*) as created_users FROM users WHERE email IN ('admin@lawfirm.com', 'test@lawfirm.com');