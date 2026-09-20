ALTER TABLE case_intakes
    ADD COLUMN IF NOT EXISTS approval_claim_token VARCHAR(64);
