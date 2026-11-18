DROP INDEX IF EXISTS idx_clients_created_by;
ALTER TABLE clients DROP COLUMN IF EXISTS created_by;
