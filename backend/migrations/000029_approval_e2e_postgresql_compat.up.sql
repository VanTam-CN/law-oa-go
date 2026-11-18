-- PostgreSQL compatibility columns required by the approval -> case E2E path.

ALTER TABLE approval_requests
    ADD COLUMN IF NOT EXISTS timeout_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS escalated BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS escalated_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS escalated_to VARCHAR(36),
    ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS conflict_risk_level VARCHAR(30),
    ADD COLUMN IF NOT EXISTS conflict_check_time TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS conflict_result JSONB,
    ADD COLUMN IF NOT EXISTS case_created BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS created_case_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS case_creation_time TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS case_creation_status VARCHAR(30);

CREATE TABLE IF NOT EXISTS approval_case_creation_tracking (
    id VARCHAR(100) PRIMARY KEY,
    approval_request_id VARCHAR(36) NOT NULL,
    case_id VARCHAR(36),
    case_number VARCHAR(50),
    case_type VARCHAR(100),
    creation_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    creation_step VARCHAR(100),
    progress_percentage NUMERIC(5, 2) NOT NULL DEFAULT 0,
    error_code VARCHAR(50),
    error_message TEXT,
    error_details JSONB NOT NULL DEFAULT '{}'::jsonb,
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 3,
    data_mapping JSONB NOT NULL DEFAULT '{}'::jsonb,
    mapped_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    unmapped_fields JSONB NOT NULL DEFAULT '{}'::jsonb,
    applied_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    imposed_requirements JSONB NOT NULL DEFAULT '{}'::jsonb,
    workflow_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_by VARCHAR(36) NOT NULL,
    processed_by VARCHAR(36),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_approval_case_creation_approval_request_id
    ON approval_case_creation_tracking(approval_request_id);
CREATE INDEX IF NOT EXISTS idx_approval_case_creation_case_id
    ON approval_case_creation_tracking(case_id);
CREATE INDEX IF NOT EXISTS idx_approval_case_creation_creation_status
    ON approval_case_creation_tracking(creation_status);

ALTER TABLE approval_records
    ADD COLUMN IF NOT EXISTS approver_title VARCHAR(100),
    ADD COLUMN IF NOT EXISTS approver_role VARCHAR(100),
    ADD COLUMN IF NOT EXISTS approved_conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS imposed_requirements JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS follow_up_actions JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS is_delegation BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS original_approver_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS effective_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS next_review_date TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS supporting_documents JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS evidence_references JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE cases
    ADD COLUMN IF NOT EXISTS case_number VARCHAR(50);

UPDATE cases
SET case_number = 'CASE-' || id::text
WHERE case_number IS NULL OR case_number = '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_cases_case_number
    ON cases(case_number)
    WHERE case_number IS NOT NULL;
