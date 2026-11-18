-- Operational contacts are independent records. Contact phone and email are
-- application-encrypted before persistence; no plaintext compatibility column
-- is introduced.
CREATE TABLE IF NOT EXISTS client_contacts (
    id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    position VARCHAR(100) NOT NULL DEFAULT '',
    phone_ciphertext TEXT NOT NULL DEFAULT '',
    email_ciphertext TEXT NOT NULL DEFAULT '',
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    version BIGINT NOT NULL DEFAULT 1,
    created_by BIGINT NOT NULL REFERENCES users(id),
    updated_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_client_contacts_client_id
    ON client_contacts(client_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_client_primary_contact
    ON client_contacts(client_id)
    WHERE is_primary = TRUE;

-- Legacy contact_person/contact_phone remain read-only compatibility data.
-- They are not copied because the old phone column is plaintext and the old
-- customer email field cannot reliably be classified as a contact email.
