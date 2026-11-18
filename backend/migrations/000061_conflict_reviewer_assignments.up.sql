-- P0 independent-reviewer appointment and direct-management recusal data.
-- This migration is additive and deliberately has no foreign keys with
-- cascading deletes: conflict evidence and appointment history are auditable.

ALTER TABLE users ADD COLUMN IF NOT EXISTS department_id BIGINT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS manager_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_users_department_id ON users (department_id);
CREATE INDEX IF NOT EXISTS idx_users_manager_id ON users (manager_id);

CREATE TABLE IF NOT EXISTS conflict_reviewer_assignments (
    id VARCHAR(36) PRIMARY KEY,
    check_id VARCHAR(100) NOT NULL,
    case_id BIGINT NULL,
    reviewer_id BIGINT NOT NULL,
    delegate_for_id BIGINT NULL,
    assigned_by BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    recusal_declared BOOLEAN NOT NULL DEFAULT FALSE,
    independence_reason TEXT,
    sla_due_at TIMESTAMP NULL,
    effective_from TIMESTAMP NULL,
    effective_to TIMESTAMP NULL,
    revoked_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conflict_reviewer_assignments_check
    ON conflict_reviewer_assignments (check_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_conflict_reviewer_assignments_reviewer
    ON conflict_reviewer_assignments (reviewer_id, status, created_at DESC);
