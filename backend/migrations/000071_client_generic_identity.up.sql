-- Canonical protected identity fields for both individual and enterprise clients.
ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS identity_type VARCHAR(30) NULL,
    ADD COLUMN IF NOT EXISTS identity_number_ciphertext TEXT NULL,
    ADD COLUMN IF NOT EXISTS identity_number_digest VARCHAR(64) NULL,
    ADD COLUMN IF NOT EXISTS aliases TEXT NULL;

UPDATE clients
SET identity_type = CASE
        WHEN type IN ('企业', '公司', 'COMPANY') THEN 'SOCIAL_CREDIT_CODE'
        ELSE 'ID_CARD'
    END,
    identity_number_ciphertext = id_card_ciphertext,
    identity_number_digest = id_card_digest
WHERE COALESCE(identity_number_digest, '') = ''
  AND COALESCE(id_card_digest, '') <> '';

CREATE INDEX IF NOT EXISTS idx_clients_identity_type
    ON clients (identity_type);
CREATE INDEX IF NOT EXISTS idx_clients_identity_number_digest
    ON clients (identity_number_digest);
