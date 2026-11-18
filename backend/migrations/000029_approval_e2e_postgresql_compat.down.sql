DROP TABLE IF EXISTS approval_case_creation_tracking;

ALTER TABLE approval_requests
    DROP COLUMN IF EXISTS case_creation_status,
    DROP COLUMN IF EXISTS case_creation_time,
    DROP COLUMN IF EXISTS created_case_id,
    DROP COLUMN IF EXISTS case_created,
    DROP COLUMN IF EXISTS conflict_result,
    DROP COLUMN IF EXISTS conflict_check_time,
    DROP COLUMN IF EXISTS conflict_risk_level,
    DROP COLUMN IF EXISTS conflict_check_id,
    DROP COLUMN IF EXISTS escalated_to,
    DROP COLUMN IF EXISTS escalated_at,
    DROP COLUMN IF EXISTS escalated,
    DROP COLUMN IF EXISTS timeout_at;

ALTER TABLE approval_records
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS evidence_references,
    DROP COLUMN IF EXISTS supporting_documents,
    DROP COLUMN IF EXISTS next_review_date,
    DROP COLUMN IF EXISTS effective_date,
    DROP COLUMN IF EXISTS original_approver_id,
    DROP COLUMN IF EXISTS is_delegation,
    DROP COLUMN IF EXISTS follow_up_actions,
    DROP COLUMN IF EXISTS imposed_requirements,
    DROP COLUMN IF EXISTS approved_conditions,
    DROP COLUMN IF EXISTS approver_role,
    DROP COLUMN IF EXISTS approver_title;

DROP INDEX IF EXISTS idx_cases_case_number;
ALTER TABLE cases DROP COLUMN IF EXISTS case_number;
