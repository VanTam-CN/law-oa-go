-- N4: server-owned workflow tasks; existing source_id and historical text stay intact.
ALTER TABLE inbox_items
    ADD COLUMN IF NOT EXISTS source_ref VARCHAR(100),
    ADD COLUMN IF NOT EXISTS completion_mode VARCHAR(20),
    ADD COLUMN IF NOT EXISTS workflow_task_key VARCHAR(200),
    ADD COLUMN IF NOT EXISTS workflow_group VARCHAR(150),
    ADD COLUMN IF NOT EXISTS workflow_purpose VARCHAR(40),
    ADD COLUMN IF NOT EXISTS workflow_version VARCHAR(100),
    ADD COLUMN IF NOT EXISTS closed_reason VARCHAR(30),
    ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- NULL legacy modes are interpreted as legacy_unresolved; no historical rewrite.
ALTER TABLE inbox_items ADD CONSTRAINT ck_inbox_workflow_contract CHECK (
    completion_mode IS DISTINCT FROM 'workflow' OR (
        NULLIF(BTRIM(source_ref), '') IS NOT NULL AND
        NULLIF(BTRIM(workflow_task_key), '') IS NOT NULL AND
        NULLIF(BTRIM(workflow_group), '') IS NOT NULL AND
        NULLIF(BTRIM(workflow_purpose), '') IS NOT NULL AND
        NULLIF(BTRIM(workflow_version), '') IS NOT NULL
    )
);
ALTER TABLE inbox_items ADD CONSTRAINT ck_inbox_workflow_key CHECK (
    workflow_task_key IS NULL OR LENGTH(BTRIM(workflow_task_key)) > 0
);
CREATE UNIQUE INDEX uq_inbox_workflow_task_key ON inbox_items(workflow_task_key);
CREATE INDEX idx_inbox_workflow_group ON inbox_items(workflow_group);
CREATE INDEX idx_inbox_owner_archive ON inbox_items(user_id, archived_at);
