-- 客户档案版本元数据：用于前端并发编辑时的乐观锁校验（PostgreSQL）
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

COMMENT ON COLUMN clients.version IS '乐观锁版本号';

CREATE INDEX IF NOT EXISTS idx_clients_version ON clients(version);
