-- 删除索引和约束

-- 删除复合索引
DROP INDEX IF EXISTS idx_cases_client_status ON cases;
DROP INDEX IF EXISTS idx_cases_lawyer_status ON cases;
DROP INDEX IF EXISTS idx_cases_type_priority ON cases;
DROP INDEX IF EXISTS idx_users_role_status ON users;
DROP INDEX IF EXISTS idx_clients_status_created ON clients;

-- 删除约束
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS chk_users_role,
DROP CONSTRAINT IF EXISTS chk_users_status;

ALTER TABLE clients 
DROP CONSTRAINT IF EXISTS chk_clients_status;

ALTER TABLE cases 
DROP CONSTRAINT IF EXISTS chk_cases_type,
DROP CONSTRAINT IF EXISTS chk_cases_priority,
DROP CONSTRAINT IF EXISTS chk_cases_status;