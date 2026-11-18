ALTER TABLE case_intakes
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(120);

CREATE UNIQUE INDEX IF NOT EXISTS idx_case_intakes_idempotency_key
    ON case_intakes (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
