-- 创建测试用户并确保密码哈希正确
-- 删除现有测试用户
DELETE FROM users WHERE email IN ('test@example.com', 'test_admin@example.com', 'test_lawyer@example.com');

-- 插入新的测试用户，使用已知的密码哈希 (123456)
INSERT INTO users (username, password, email, phone, real_name, status) VALUES
('test_user', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'test@example.com', '13800138888', '测试用户', 'active'),
('test_admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'test_admin@example.com', '13800138889', '测试管理员', 'active'),
('test_lawyer', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/GBzUx/Fqe', 'test_lawyer@example.com', '13800138890', '测试律师', 'active');

-- 验证插入结果
SELECT id, username, email, real_name, status, created_at FROM users WHERE email LIKE 'test%@example.com';