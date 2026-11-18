-- P0: exact identity matching uses a keyed digest; the ciphertext is only
-- decrypted inside trusted server-side workflows. Existing identity_number is
-- retained temporarily for an explicit backfill audit and must be empty before
-- production readiness can pass.
ALTER TABLE entities ADD COLUMN IF NOT EXISTS identity_number_digest VARCHAR(64);
ALTER TABLE entities ADD COLUMN IF NOT EXISTS identity_number_ciphertext TEXT;
-- The legacy unique index is on the plaintext column. Once new rows keep that
-- column empty, it would incorrectly allow only one entity per identity type.
DROP INDEX IF EXISTS uq_entities_identity;
CREATE INDEX IF NOT EXISTS idx_entities_identity_number_digest ON entities(identity_number_digest);
