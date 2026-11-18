-- P0 subject versioning and fail-closed action gate.
-- This migration deliberately has no foreign-key cascade and no delete path:
-- review, revision and compliance evidence must remain auditable.

ALTER TABLE cases
    ADD COLUMN IF NOT EXISTS subject_version INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS subject_state VARCHAR(50) NOT NULL DEFAULT 'EFFECTIVE',
    ADD COLUMN IF NOT EXISTS subject_snapshot TEXT,
    ADD COLUMN IF NOT EXISTS pending_subject_revision_id VARCHAR(36),
    ADD COLUMN IF NOT EXISTS conflict_check_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS conflict_coverage_status VARCHAR(30) NOT NULL DEFAULT 'COVERAGE_LIMITED';

CREATE TABLE IF NOT EXISTS conflict_search_scopes (
    id VARCHAR(100) PRIMARY KEY,
    scope_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    coverage_status VARCHAR(30) NOT NULL DEFAULT 'COVERAGE_LIMITED',
    source_version VARCHAR(120),
    covered_from TIMESTAMP NULL,
    covered_to TIMESTAMP NULL,
    missing_sources TEXT,
    approved_by BIGINT NULL,
    approved_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS case_subject_revisions (
    id VARCHAR(36) PRIMARY KEY,
    case_id BIGINT NOT NULL,
    base_subject_version INTEGER NOT NULL,
    change_type VARCHAR(80) NOT NULL,
    status VARCHAR(50) NOT NULL,
    payload TEXT NOT NULL,
    reason TEXT,
    conflict_check_id VARCHAR(100),
    requested_by BIGINT NOT NULL,
    reviewed_by BIGINT NULL,
    review_decision VARCHAR(40),
    review_notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    effective_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS compliance_audit_events (
    id VARCHAR(36) PRIMARY KEY,
    actor_id BIGINT NULL,
    actor_role VARCHAR(50),
    event_type VARCHAR(80) NOT NULL,
    object_type VARCHAR(80) NOT NULL,
    object_id VARCHAR(100) NOT NULL,
    request_id VARCHAR(100),
    from_state VARCHAR(50),
    to_state VARCHAR(50),
    subject_version INTEGER,
    payload TEXT,
    integrity_hash VARCHAR(64),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conflict_search_scopes_status_coverage
    ON conflict_search_scopes (status, coverage_status);
CREATE INDEX IF NOT EXISTS idx_case_subject_revisions_case_status
    ON case_subject_revisions (case_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_case_subject_revisions_check
    ON case_subject_revisions (conflict_check_id);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_events_object_created
    ON compliance_audit_events (object_type, object_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_compliance_audit_events_actor_created
    ON compliance_audit_events (actor_id, created_at DESC);
