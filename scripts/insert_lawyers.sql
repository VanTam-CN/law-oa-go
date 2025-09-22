-- 插入律师用户
INSERT INTO users (username, password, email, real_name, status, created_at) VALUES
('lawyer1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi.', 'lawyer1@example.com', '张律师', 'active', NOW()),
('lawyer2', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi.', 'lawyer2@example.com', '李律师', 'active', NOW()),
('lawyer3', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2uheWG/igi.', 'lawyer3@example.com', '王律师', 'active', NOW());