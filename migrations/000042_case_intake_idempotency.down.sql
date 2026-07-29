DROP INDEX IF EXISTS idx_case_intakes_idempotency_key;

ALTER TABLE case_intakes
    DROP COLUMN IF EXISTS idempotency_key;
