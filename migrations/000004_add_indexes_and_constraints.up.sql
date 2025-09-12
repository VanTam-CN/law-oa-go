-- 添加额外的索引和约束

-- 用户表优化
ALTER TABLE users 
ADD CONSTRAINT chk_users_role CHECK (role IN ('admin', 'lawyer', 'user')),
ADD CONSTRAINT chk_users_status CHECK (status IN ('active', 'inactive', 'suspended'));

-- 客户表优化
ALTER TABLE clients 
ADD CONSTRAINT chk_clients_status CHECK (status IN ('active', 'inactive', 'prospect', 'lost', 'blacklisted'));

-- 案件表优化
ALTER TABLE cases 
ADD CONSTRAINT chk_cases_type CHECK (case_type IN ('civil', 'criminal', 'commercial', 'administrative')),
ADD CONSTRAINT chk_cases_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
ADD CONSTRAINT chk_cases_status CHECK (status IN ('pending', 'active', 'closed', 'suspended'));

-- 添加复合索引
CREATE INDEX idx_cases_client_status ON cases(client_id, status);
CREATE INDEX idx_cases_lawyer_status ON cases(lawyer_id, status);
CREATE INDEX idx_cases_type_priority ON cases(case_type, priority);
CREATE INDEX idx_users_role_status ON users(role, status);
CREATE INDEX idx_clients_status_created ON clients(status, created_at);