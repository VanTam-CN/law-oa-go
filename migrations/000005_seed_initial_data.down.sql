-- 清除初始数据

-- 删除示例案件
DELETE FROM cases WHERE id IN (1, 2, 3);

-- 删除示例客户
DELETE FROM clients WHERE id IN (1, 2, 3);

-- 删除初始用户
DELETE FROM users WHERE email IN ('admin@law-oa.com', 'zhang@law-oa.com', 'li@law-oa.com');