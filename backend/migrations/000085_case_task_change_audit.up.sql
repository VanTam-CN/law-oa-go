ALTER TABLE case_tasks
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS case_task_changes (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL,
    case_id BIGINT NOT NULL,
    change_type VARCHAR(20) NOT NULL,
    actor_id BIGINT NOT NULL,
    old_assignee_id BIGINT,
    new_assignee_id BIGINT,
    old_due_date TIMESTAMPTZ,
    new_due_date TIMESTAMPTZ,
    version BIGINT,
    old_version BIGINT NOT NULL DEFAULT 0,
    new_version BIGINT NOT NULL DEFAULT 0,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_case_task_changes_task_id ON case_task_changes (task_id);
CREATE INDEX IF NOT EXISTS idx_case_task_changes_case_id ON case_task_changes (case_id);
