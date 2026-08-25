-- Immutable policy material packages and role-separated endorsements.

CREATE TABLE IF NOT EXISTS law_firm_compliance_policy_packages (
    id VARCHAR(100) PRIMARY KEY,
    policy_version VARCHAR(80) NOT NULL UNIQUE,
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
    effective_at TIMESTAMPTZ NOT NULL,
    next_review_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NULL,
    integrity_hash VARCHAR(64) NOT NULL,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS law_firm_compliance_policy_endorsements (
    id VARCHAR(36) PRIMARY KEY,
    policy_package_id VARCHAR(100) NOT NULL REFERENCES law_firm_compliance_policy_packages(id),
    endorsement_type VARCHAR(20) NOT NULL CHECK (endorsement_type IN ('MANAGEMENT', 'COMPLIANCE')),
    endorsed_by BIGINT NOT NULL,
    endorser_role VARCHAR(50) NOT NULL,
    package_integrity_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_policy_package_endorsement UNIQUE (policy_package_id, endorsement_type)
);

CREATE INDEX IF NOT EXISTS idx_policy_endorsement_actor
    ON law_firm_compliance_policy_endorsements (endorsed_by, created_at);

DROP TRIGGER IF EXISTS trg_law_firm_policy_packages_append_only
    ON law_firm_compliance_policy_packages;
CREATE TRIGGER trg_law_firm_policy_packages_append_only
    BEFORE UPDATE OR DELETE ON law_firm_compliance_policy_packages
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();

DROP TRIGGER IF EXISTS trg_law_firm_policy_endorsements_append_only
    ON law_firm_compliance_policy_endorsements;
CREATE TRIGGER trg_law_firm_policy_endorsements_append_only
    BEFORE UPDATE OR DELETE ON law_firm_compliance_policy_endorsements
    FOR EACH ROW EXECUTE FUNCTION law_oa_reject_append_only_mutation();
