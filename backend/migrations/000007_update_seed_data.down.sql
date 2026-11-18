-- 回滚数据更新：删除新增的数据

-- 删除新增的文档数据
DELETE FROM documents WHERE name IN ('合同模板', '身份证复印件', '案件委托书', '营业执照', '起诉状', '证据清单');

-- 删除新增的案件数据
DELETE FROM cases WHERE title IN ('商业合同纠纷', '劳动仲裁', '商标注册', '股权转让', '房产买卖', '公司设立', '债务追讨', '知识产权');

-- 删除新增的客户数据
DELETE FROM clients WHERE email IN ('zhou@example.com', 'wu@example.com', 'zheng@example.com', 'sunmgr@example.com', 'qian@example.com', 'feng@example.com');

-- 删除新增的用户数据
DELETE FROM users WHERE email IN ('admin@example.com', 'wang@law-oa.com', 'zhao@law-oa.com', 'sun@law-oa.com');