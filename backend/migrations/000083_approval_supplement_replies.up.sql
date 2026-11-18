-- Supplement replies draft persistence (R4b-fix). Additive only; it does not change
-- business-fact fingerprints or intake versions.
ALTER TABLE approval_records
    ADD COLUMN IF NOT EXISTS supplement_replies JSONB NOT NULL DEFAULT '{}'::jsonb;

COMMENT ON COLUMN approval_records.supplement_replies IS '申请人逐项补充回复草稿';

-- Backfill idempotent: existing NULL rows become empty draft objects so the
-- Go string field never scans NULL.
UPDATE approval_records SET supplement_replies = '{}'::jsonb WHERE supplement_replies IS NULL;
