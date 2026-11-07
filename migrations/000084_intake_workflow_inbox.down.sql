DROP INDEX IF EXISTS idx_inbox_owner_archive;
DROP INDEX IF EXISTS idx_inbox_workflow_group;
DROP INDEX IF EXISTS uq_inbox_workflow_task_key;
ALTER TABLE inbox_items
    DROP CONSTRAINT IF EXISTS ck_inbox_workflow_contract,
    DROP CONSTRAINT IF EXISTS ck_inbox_workflow_key,
    DROP COLUMN IF EXISTS archived_at,
    DROP COLUMN IF EXISTS closed_reason,
    DROP COLUMN IF EXISTS workflow_version,
    DROP COLUMN IF EXISTS workflow_purpose,
    DROP COLUMN IF EXISTS workflow_group,
    DROP COLUMN IF EXISTS workflow_task_key,
    DROP COLUMN IF EXISTS completion_mode,
    DROP COLUMN IF EXISTS source_ref;
