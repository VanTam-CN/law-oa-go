DROP INDEX IF EXISTS idx_case_intake_parties_identity_digest;
DROP INDEX IF EXISTS idx_case_intake_parties_entity_id;

ALTER TABLE case_intake_parties
    DROP COLUMN IF EXISTS aliases,
    DROP COLUMN IF EXISTS identity_number_digest,
    DROP COLUMN IF EXISTS identity_number_ciphertext,
    DROP COLUMN IF EXISTS identity_type,
    DROP COLUMN IF EXISTS entity_id;
