DROP INDEX IF EXISTS idx_clients_identity_number_digest;
DROP INDEX IF EXISTS idx_clients_identity_type;

ALTER TABLE clients
    DROP COLUMN IF EXISTS aliases,
    DROP COLUMN IF EXISTS identity_number_digest,
    DROP COLUMN IF EXISTS identity_number_ciphertext,
    DROP COLUMN IF EXISTS identity_type;
