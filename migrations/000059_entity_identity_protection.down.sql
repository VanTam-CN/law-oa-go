DROP INDEX IF EXISTS idx_entities_identity_number_digest;
CREATE UNIQUE INDEX IF NOT EXISTS uq_entities_identity ON entities(identity_type, identity_number);
ALTER TABLE entities DROP COLUMN IF EXISTS identity_number_ciphertext;
ALTER TABLE entities DROP COLUMN IF EXISTS identity_number_digest;
