DROP INDEX IF EXISTS idx_clients_version;
ALTER TABLE clients DROP COLUMN IF EXISTS version;
