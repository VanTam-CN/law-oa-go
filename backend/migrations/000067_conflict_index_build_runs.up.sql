-- Durable reconciliation evidence for the P0 firm-wide subject index.
-- The run is separate from the scope approval so an operator cannot mark a
-- source COMPLETE with only a manually typed evidence reference.

CREATE TABLE IF NOT EXISTS conflict_index_build_runs (
    id VARCHAR(100) PRIMARY KEY,
    scope_type VARCHAR(50) NOT NULL,
    source_version VARCHAR(120) NOT NULL,
    status VARCHAR(20) NOT NULL,
    source_record_count BIGINT NOT NULL DEFAULT 0,
    indexed_record_count BIGINT NOT NULL DEFAULT 0,
    missing_record_count BIGINT NOT NULL DEFAULT 0,
    reconciliation_hash VARCHAR(64) NOT NULL,
    evidence_reference TEXT,
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    created_by BIGINT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conflict_index_build_runs_scope_status
    ON conflict_index_build_runs (scope_type, status, source_version);

ALTER TABLE conflict_search_scopes
    ADD COLUMN index_run_id VARCHAR(100);

CREATE INDEX idx_conflict_search_scopes_index_run
    ON conflict_search_scopes (index_run_id);
