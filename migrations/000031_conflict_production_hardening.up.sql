-- Production hardening for the active /api/v1/conflict module on PostgreSQL.
-- Keeps every conflict check auditable and persists individual conflict hits.

CREATE TABLE IF NOT EXISTS conflict_check_records (
    check_id VARCHAR(100) PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    case_name VARCHAR(255) NOT NULL,
    case_type VARCHAR(100) NOT NULL,
    check_status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    has_conflict BOOLEAN NOT NULL DEFAULT false,
    risk_level VARCHAR(20) NOT NULL DEFAULT 'LOW',
    search_parameters JSONB NOT NULL DEFAULT '{}'::jsonb,
    check_result JSONB NOT NULL DEFAULT '{}'::jsonb,
    user_id BIGINT,
    check_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    duration BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE conflict_check_records
    ALTER COLUMN check_status TYPE VARCHAR(20),
    ALTER COLUMN risk_level TYPE VARCHAR(20),
    ALTER COLUMN search_parameters TYPE JSONB USING COALESCE(search_parameters::jsonb, '{}'::jsonb),
    ALTER COLUMN search_parameters SET DEFAULT '{}'::jsonb,
    ALTER COLUMN check_result TYPE JSONB USING COALESCE(check_result::jsonb, '{}'::jsonb),
    ALTER COLUMN check_result SET DEFAULT '{}'::jsonb;

ALTER TABLE conflict_check_records
    DROP CONSTRAINT IF EXISTS chk_conflict_check_records_status,
    DROP CONSTRAINT IF EXISTS chk_conflict_check_records_risk_level;

ALTER TABLE conflict_check_records
    ADD CONSTRAINT chk_conflict_check_records_status
    CHECK (check_status IN ('QUEUED', 'RUNNING', 'PROCESSING', 'COMPLETED', 'FAILED')),
    ADD CONSTRAINT chk_conflict_check_records_risk_level
    CHECK (risk_level IN ('MINIMAL', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'));

CREATE INDEX IF NOT EXISTS idx_conflict_check_records_client_time
ON conflict_check_records(client_id, check_time DESC);

CREATE INDEX IF NOT EXISTS idx_conflict_check_records_status_updated
ON conflict_check_records(check_status, updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_conflict_check_records_risk
ON conflict_check_records(risk_level, has_conflict);

CREATE TABLE IF NOT EXISTS conflict_cases (
    id VARCHAR(160) PRIMARY KEY,
    check_id VARCHAR(100) NOT NULL,
    case_id VARCHAR(36),
    case_name VARCHAR(255) NOT NULL,
    case_no VARCHAR(50),
    case_type VARCHAR(100),
    conflict_type VARCHAR(100) NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    description TEXT,
    case_status VARCHAR(30),
    client_id VARCHAR(36) NOT NULL,
    opposing_parties JSONB NOT NULL DEFAULT '[]'::jsonb,
    conflict_details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE conflict_cases
    ALTER COLUMN opposing_parties TYPE JSONB USING COALESCE(opposing_parties::jsonb, '[]'::jsonb);

ALTER TABLE conflict_cases
    DROP CONSTRAINT IF EXISTS chk_conflict_cases_risk_level,
    DROP CONSTRAINT IF EXISTS fk_conflict_cases_check;

ALTER TABLE conflict_cases
    ADD CONSTRAINT chk_conflict_cases_risk_level
    CHECK (risk_level IN ('MINIMAL', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    ADD CONSTRAINT fk_conflict_cases_check
    FOREIGN KEY (check_id) REFERENCES conflict_check_records(check_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_conflict_cases_check_id
ON conflict_cases(check_id);

CREATE INDEX IF NOT EXISTS idx_conflict_cases_client_risk
ON conflict_cases(client_id, risk_level);

CREATE TABLE IF NOT EXISTS client_relations (
    id VARCHAR(36) PRIMARY KEY,
    client_id VARCHAR(36) NOT NULL,
    related_client_id VARCHAR(36) NOT NULL,
    relation_type VARCHAR(30) NOT NULL,
    relation_detail VARCHAR(500),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (client_id, related_client_id, relation_type)
);

CREATE INDEX IF NOT EXISTS idx_client_relations_client_active
ON client_relations(client_id, active);

CREATE INDEX IF NOT EXISTS idx_client_relations_related
ON client_relations(related_client_id);
