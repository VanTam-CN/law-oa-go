-- 修复客户表中的name字段
-- 将company字段的值复制到name字段（当name字段为空时）

UPDATE clients 
SET name = company 
WHERE (name IS NULL OR name = '') 
  AND company IS NOT NULL 
  AND company != '';

-- 验证修复结果
SELECT 
    id,
    name,
    company,
    email,
    CASE 
        WHEN name IS NULL OR name = '' THEN '需要修复'
        ELSE '正常'
    END as name_status
FROM clients 
ORDER BY id;