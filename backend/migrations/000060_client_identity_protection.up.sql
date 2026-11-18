-- P0: protect client identity numbers at rest and keep exact lookup separate
-- from the legacy plaintext column until an explicit backfill is complete.
ALTER TABLE clients ADD COLUMN IF NOT EXISTS id_card_digest VARCHAR(64);
ALTER TABLE clients ADD COLUMN IF NOT EXISTS id_card_ciphertext TEXT;
CREATE INDEX IF NOT EXISTS idx_clients_id_card_digest ON clients(id_card_digest);
