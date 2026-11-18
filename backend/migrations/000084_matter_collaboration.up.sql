-- N7 material collaboration vertical: additive columns only. Existing
-- case_materials rows (000028) keep their meaning; new columns are
-- nullable with defaults so the intake save path keeps working untouched.
ALTER TABLE case_materials
    ADD COLUMN IF NOT EXISTS owner_id BIGINT,
    ADD COLUMN IF NOT EXISTS uploaded_by BIGINT,
    ADD COLUMN IF NOT EXISTS uploaded_at TIMESTAMPTZ;

COMMENT ON COLUMN case_materials.owner_id IS '材料责任人：案件团队内合法人员（案件律师/管理员或白名单成员）';
COMMENT ON COLUMN case_materials.uploaded_by IS '实际上传人；与 uploaded_at 只能由上传事务写入，客户端不能直接置 uploaded';

-- Physical attachments live in their own safe storage tree; they never
-- fake documents IDs. The digest lets the approval snapshot freeze a
-- verifiable reference of each uploaded file.
CREATE TABLE IF NOT EXISTS material_files (
    id BIGSERIAL PRIMARY KEY,
    material_id BIGINT NOT NULL,
    content_sha256 VARCHAR(64) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL,
    content_type VARCHAR(120),
    storage_path VARCHAR(500) NOT NULL,
    uploaded_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_material_files_material_id ON material_files (material_id);

-- Case tasks (general matter tasks, item 5). case_tasks may already exist
-- in environments where the inbox projection shipped early (000082-era
-- deploy); every statement is additive.
CREATE TABLE IF NOT EXISTS case_tasks (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    detail TEXT,
    assignee_id BIGINT NOT NULL,
    created_by BIGINT NOT NULL,
    due_date TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    completed_by BIGINT,
    completed_at TIMESTAMPTZ,
    idempotency_key VARCHAR(120),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_case_tasks_case_id ON case_tasks (case_id);
CREATE INDEX IF NOT EXISTS idx_case_tasks_assignee ON case_tasks (assignee_id);
CREATE INDEX IF NOT EXISTS idx_case_tasks_status ON case_tasks (status);

-- Idempotent creation per caller key scoped to its creator; NULL keys
-- (legacy/manual rows) stay outside the unique index. A second column
-- freezes the SHA-256 of the whole create payload so a same-key retry
-- with any differing field is a conflict instead of a silent reuse.
ALTER TABLE case_tasks ADD COLUMN IF NOT EXISTS idempotency_fingerprint VARCHAR(64) NOT NULL DEFAULT '';
-- Free-text completion record, written by the same transaction that flips
-- the task status; the API never reports a note as recorded unless this
-- column was updated in that commit.
ALTER TABLE case_tasks ADD COLUMN IF NOT EXISTS completion_note TEXT;
-- NULL keys (server-minted auto keys are always non-NULL) stay outside the
-- unique index on PostgreSQL; the service layer guarantees non-empty keys.
CREATE UNIQUE INDEX IF NOT EXISTS uq_case_tasks_creator_idempotency_key
    ON case_tasks (created_by, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
