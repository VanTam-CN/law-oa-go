-- Production release policy and measurable source-quality evidence for real-client conflict work.

ALTER TABLE conflict_search_scopes
    ADD COLUMN IF NOT EXISTS source_of_truth BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS sync_mode VARCHAR(30),
    ADD COLUMN IF NOT EXISTS max_sync_lag_minutes INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_successful_sync_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS minimum_field_coverage_bps INTEGER NOT NULL DEFAULT 10000,
    ADD COLUMN IF NOT EXISTS measured_field_coverage_bps INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS maximum_duplicate_rate_bps INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS measured_duplicate_rate_bps INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS quality_owner_id BIGINT NULL,
    ADD COLUMN IF NOT EXISTS quality_reviewed_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS max_quality_review_age_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS failure_alert_reference TEXT NULL,
    ADD COLUMN IF NOT EXISTS correction_procedure_reference TEXT NULL;

CREATE TABLE IF NOT EXISTS law_firm_compliance_policy_profiles (
    id VARCHAR(100) PRIMARY KEY,
    policy_version VARCHAR(80) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL,
    jurisdiction VARCHAR(120) NOT NULL,
    applicable_rule_name VARCHAR(255) NOT NULL,
    applicable_rule_version VARCHAR(120) NOT NULL,
    applicable_rule_authority VARCHAR(255) NOT NULL,
    applicable_rule_reference TEXT NOT NULL,
    data_source_policy_reference TEXT NOT NULL,
    privacy_basis_matrix_reference TEXT NOT NULL,
    retention_policy_reference TEXT NOT NULL,
    waiver_policy_reference TEXT NOT NULL,
    controlled_actions_reference TEXT NOT NULL,
    external_review_reference TEXT NOT NULL,
    management_approved_by BIGINT NOT NULL,
    compliance_approved_by BIGINT NOT NULL,
    approved_at TIMESTAMPTZ NULL,
    effective_at TIMESTAMPTZ NULL,
    next_review_at TIMESTAMPTZ NULL,
    expires_at TIMESTAMPTZ NULL,
    integrity_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_law_firm_compliance_policy_status
    ON law_firm_compliance_policy_profiles (status, effective_at, next_review_at);

DROP TRIGGER IF EXISTS trg_law_firm_policy_profiles_append_only
    ON law_firm_compliance_policy_profiles;
CREATE TRIGGER trg_law_firm_policy_profiles_append_only
    BEFORE UPDATE OR DELETE ON law_firm_compliance_policy_profiles
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();
