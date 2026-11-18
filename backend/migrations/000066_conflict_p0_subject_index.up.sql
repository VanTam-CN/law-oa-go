-- P0 subject snapshots, protected identifiers and normalized match evidence.
-- These tables deliberately avoid foreign keys and database-specific enums so
-- they can be created by both the PostgreSQL and MySQL migration paths.

CREATE TABLE IF NOT EXISTS conflict_subject_versions (
    id VARCHAR(100) PRIMARY KEY,
    subject_key VARCHAR(200) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id VARCHAR(100) NOT NULL,
    case_id VARCHAR(100),
    client_id VARCHAR(100),
    subject_role VARCHAR(50) NOT NULL,
    subject_type VARCHAR(50) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,
    alias_snapshot TEXT,
    source_version VARCHAR(120) NOT NULL,
    version_number INTEGER NOT NULL DEFAULT 1,
    verification_status VARCHAR(40) NOT NULL,
    snapshot TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conflict_subject_versions_key
    ON conflict_subject_versions (subject_key, created_at);
CREATE INDEX idx_conflict_subject_versions_name
    ON conflict_subject_versions (normalized_name);
CREATE INDEX idx_conflict_subject_versions_source
    ON conflict_subject_versions (source_type, source_id, created_at);

CREATE TABLE IF NOT EXISTS conflict_subject_identifiers (
    id VARCHAR(100) PRIMARY KEY,
    subject_version_id VARCHAR(100) NOT NULL,
    identifier_type VARCHAR(50) NOT NULL,
    digest VARCHAR(64),
    ciphertext TEXT,
    masked_value VARCHAR(80),
    verification_status VARCHAR(40) NOT NULL,
    source_reference TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conflict_subject_identifiers_version
    ON conflict_subject_identifiers (subject_version_id);
CREATE INDEX idx_conflict_subject_identifiers_digest
    ON conflict_subject_identifiers (identifier_type, digest);

CREATE TABLE IF NOT EXISTS conflict_match_evidence_v2 (
    id VARCHAR(100) PRIMARY KEY,
    check_id VARCHAR(100) NOT NULL,
    subject_version_id VARCHAR(100),
    match_type VARCHAR(40) NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_object_id VARCHAR(100),
    restricted BOOLEAN NOT NULL DEFAULT FALSE,
    evidence_snapshot TEXT NOT NULL,
    evidence_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_conflict_match_evidence_check
    ON conflict_match_evidence_v2 (check_id, created_at);
CREATE INDEX idx_conflict_match_evidence_source
    ON conflict_match_evidence_v2 (source_type, source_object_id);
