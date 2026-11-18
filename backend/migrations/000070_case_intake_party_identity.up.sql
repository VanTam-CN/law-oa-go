-- Protected structured identities for initial intake conflict checks.
ALTER TABLE case_intake_parties
    ADD COLUMN IF NOT EXISTS entity_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS identity_type VARCHAR(30) NULL,
    ADD COLUMN IF NOT EXISTS identity_number_ciphertext TEXT NULL,
    ADD COLUMN IF NOT EXISTS identity_number_digest VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS aliases TEXT NULL;

CREATE INDEX IF NOT EXISTS idx_case_intake_parties_entity_id
    ON case_intake_parties (entity_id);
CREATE INDEX IF NOT EXISTS idx_case_intake_parties_identity_digest
    ON case_intake_parties (identity_number_digest);
