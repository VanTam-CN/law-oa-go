DROP INDEX IF EXISTS idx_clients_id_card_digest;
ALTER TABLE clients DROP COLUMN IF EXISTS id_card_ciphertext;
ALTER TABLE clients DROP COLUMN IF EXISTS id_card_digest;
