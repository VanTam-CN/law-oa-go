-- Law OA Go 系统测试用户创建脚本
-- 使用真实的bcrypt密码哈希

-- 清理现有测试用户（如果存在）
DELETE FROM users WHERE email = 'admin@lawfirm.com';
DELETE FROM users WHERE email = 'testuser@example.com';

-- 创建管理员测试用户
INSERT INTO users (username, name, email, password, role, status, created_at, updated_at) VALUES
('admin', '管理员', 'admin@lawfirm.com', '$2a$10$hM786v5m4HMmtDIqD5Vpqe0txc5sXSOJ1pPD1y49QN08JqH8xXKli', 'admin', 'active', NOW(), NOW());

-- 创建普通用户测试用户
INSERT INTO users (username, name, email, password, role, status, created_at, updated_at) VALUES
('testuser', '测试用户', 'testuser@example.com', '$2a$10$hM786v5m4HMmtDIqD5Vpqe0txc5sXSOJ1pPD1y49QN08JqH8xXKli', 'user', 'active', NOW(), NOW());

-- 验证用户创建
SELECT id, username, name, email, role, status, created_at FROM users WHERE email IN ('admin@lawfirm.com', 'testuser@example.com');

-- 显示创建的用户数量
SELECT COUNT(*) as created_users FROM users WHERE email IN ('admin@lawfirm.com', 'testuser@example.com');