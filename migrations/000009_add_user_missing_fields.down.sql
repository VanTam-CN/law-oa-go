-- 删除用户表添加的字段
ALTER TABLE users DROP COLUMN username;
ALTER TABLE users DROP COLUMN real_name;
ALTER TABLE users DROP COLUMN last_login_at;
ALTER TABLE users DROP COLUMN last_login_ip;
ALTER TABLE users DROP COLUMN remark;