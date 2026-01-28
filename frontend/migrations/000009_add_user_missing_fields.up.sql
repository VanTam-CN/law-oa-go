-- 添加用户表缺失字段
ALTER TABLE users ADD COLUMN username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名' AFTER name;
ALTER TABLE users ADD COLUMN real_name VARCHAR(50) COMMENT '真实姓名' AFTER username;
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP NULL COMMENT '最后登录时间' AFTER status;
ALTER TABLE users ADD COLUMN last_login_ip VARCHAR(45) COMMENT '最后登录IP' AFTER last_login_at;
ALTER TABLE users ADD COLUMN remark VARCHAR(255) COMMENT '备注' AFTER last_login_ip;

-- 添加索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_real_name ON users(real_name);