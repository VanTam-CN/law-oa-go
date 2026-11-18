-- Reverse of 000084: drop additive objects only. Existing case_materials
-- rows are untouched; only the new columns/tables disappear.
ALTER TABLE case_tasks DROP COLUMN IF EXISTS completion_note;
DROP INDEX IF EXISTS uq_case_tasks_creator_idempotency_key;
DROP TABLE IF EXISTS case_tasks;
DROP TABLE IF EXISTS material_files;
ALTER TABLE case_materials
    DROP COLUMN IF EXISTS uploaded_at,
    DROP COLUMN IF EXISTS uploaded_by,
    DROP COLUMN IF EXISTS owner_id;
