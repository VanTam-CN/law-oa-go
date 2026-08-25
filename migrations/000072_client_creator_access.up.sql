-- Preserve access to a newly created client before its first matter exists.
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS created_by BIGINT NULL;

CREATE INDEX IF NOT EXISTS idx_clients_created_by
    ON clients (created_by);
